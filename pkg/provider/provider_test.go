package provider

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/larsartmann/go-finding"
)

// writeFakeBinary stages an executable named docs-gates that prints the
// canned stdout and exits with the canned code, then prepends its dir to
// PATH so the provider's PATH-based exec resolves the fake. Not parallel:
// t.Setenv is process-global.
func writeFakeBinary(t *testing.T, stdout string, exitCode int) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, toolName)
	body := "#!/bin/sh\ncat <<'REPORT'\n" + stdout + "\nREPORT\nexit " + strconv.Itoa(exitCode) + "\n"

	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// matrixJSON is the canned verdict document the row-shape tests feed:
// one green row, one plain-red row, one timed-out row, one skip row —
// every field the battery's battery-json parity golden pins, spelled the
// way Detect must read it.
const matrixJSON = `{
  "root": "/repo",
  "pass": false,
  "gates": [
    {
      "name": "freshness",
      "verdict": "PASS",
      "exit_code": 0,
      "skipped": false,
      "timed_out": false,
      "duration_ms": 830,
      "stdout": "docs-freshness: FRESH\n",
      "stderr": ""
    },
    {
      "name": "cadence",
      "verdict": "FAIL",
      "exit_code": 1,
      "skipped": false,
      "timed_out": false,
      "duration_ms": 41,
      "stdout": "",
      "stderr": "docs-cadence: OVERDUE\n"
    },
    {
      "name": "audit-all",
      "verdict": "FAIL",
      "exit_code": 1,
      "skipped": false,
      "timed_out": true,
      "duration_ms": 300000,
      "stdout": "",
      "stderr": ""
    },
    {
      "name": "annotate",
      "verdict": "SKIP",
      "exit_code": 0,
      "skipped": true,
      "timed_out": false,
      "duration_ms": 0,
      "stdout": "",
      "stderr": ""
    }
  ]
}`

func TestDecodeReportKeepsFailOrder(t *testing.T) {
	t.Parallel()

	report, err := decodeReport(matrixJSON)
	if err != nil {
		t.Fatalf("decodeReport: %v", err)
	}

	got := failedGates(report)

	if len(got) != 2 {
		t.Fatalf("failedGates = %d rows, want 2: %+v", len(got), got)
	}

	if got[0].Name != "cadence" || got[1].Name != "audit-all" {
		t.Fatalf("order = %q, %q; want cadence, audit-all", got[0].Name, got[1].Name)
	}
}

// The shape-contract red control: a renamed or removed field must FAIL
// the decode (DisallowUnknownFields), never silently land as a zero
// value — this is what makes a future battery row-shape change loud on
// the provider side.
func TestDecodeReportRejectsShapeDrift(t *testing.T) {
	t.Parallel()

	drifted := strings.Replace(matrixJSON, `"timed_out": true,`, `"timedout": true,`, 1)

	if _, err := decodeReport(drifted); err == nil {
		t.Fatal("a renamed row field must fail the strict decode")
	}

	renamed := strings.Replace(matrixJSON, `"duration_ms": 830,`, `"durationMs": 830,`, 1)

	if _, err := decodeReport(renamed); err == nil {
		t.Fatal("a renamed leaf field must fail the strict decode")
	}
}

// Detect over a red matrix: one finding per FAIL row in document order,
// the timeout row distinguished in its message.
//
//nolint:paralleltest // process-global PATH swap, serialized on purpose
func TestDetectRedRowsBecomeFindings(t *testing.T) {
	writeFakeBinary(t, matrixJSON, 1)

	findings, err := Provider.Detect.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if len(findings) != 2 {
		t.Fatalf("two red rows must yield two findings, got %d: %+v", len(findings), findings)
	}

	if findings[0].Rule != "cadence" || findings[1].Rule != "audit-all" {
		t.Fatalf("rules = %q, %q; want cadence, audit-all", findings[0].Rule, findings[1].Rule)
	}

	if findings[0].Severity != finding.SeverityError {
		t.Fatalf("red rows are errors, got %q", findings[0].Severity)
	}

	if findings[0].ToolName != toolName {
		t.Fatalf("tool name = %q, want %q", findings[0].ToolName, toolName)
	}

	if !strings.Contains(findings[1].Message, "TIMED OUT") || !strings.Contains(findings[1].Message, "300000ms") {
		t.Fatalf("the timeout row's message must name the cap burn, got %q", findings[1].Message)
	}
}

//nolint:paralleltest // process-global PATH swap, serialized on purpose
func TestDetectCleanCorpusYieldsZeroFindings(t *testing.T) {
	clean := strings.Replace(matrixJSON, `"pass": false`, `"pass": true`, 1)

	writeFakeBinary(t, clean, 0)

	findings, err := Provider.Detect.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if len(findings) != 0 {
		t.Fatalf("clean battery must yield zero findings, got %+v", findings)
	}
}

//nolint:paralleltest // process-global PATH swap, serialized on purpose
func TestDetectUsageFailureBecomesOneFinding(t *testing.T) {
	writeFakeBinary(t, "FATAL: /tmp/nowhere is not a directory\n", 2)

	findings, err := Provider.Detect.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if len(findings) != 1 || findings[0].Rule != "usage" {
		t.Fatalf("usage failure must yield one usage finding, got %+v", findings)
	}

	if !strings.Contains(findings[0].Message, "is not a directory") {
		t.Fatalf("usage finding must carry the FATAL text, got %q", findings[0].Message)
	}
}

// The migration contract: a battery WITHOUT the --json face (the old
// human matrix on stdout) is a shape break, not a silent pass.
//
//nolint:paralleltest // process-global PATH swap, serialized on purpose
func TestDetectHumanMatrixIsAShapeBreak(t *testing.T) {
	writeFakeBinary(t, "docs-gates: verdict matrix\n  FAIL  cadence (rc=1)\n", 1)

	findings, err := Provider.Detect.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if len(findings) != 1 || findings[0].Rule != "battery-exit-without-report" {
		t.Fatalf("human-matrix stdout must surface as a shape break, got %+v", findings)
	}
}

//nolint:paralleltest // process-global PATH swap, serialized on purpose
func TestDetectRedExitWithoutReportIsItsOwnFinding(t *testing.T) {
	writeFakeBinary(t, "", 1)

	findings, err := Provider.Detect.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if len(findings) != 1 || findings[0].Rule != "battery-exit-without-report" {
		t.Fatalf("shape break must surface, got %+v", findings)
	}
}

func TestHealthCheckReportsMissingBinary(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // no docs-gates anywhere

	err := Provider.HealthCheck(context.Background())
	if err == nil {
		t.Fatal("missing binary must fail the health check")
	}

	if !strings.Contains(err.Error(), "docs-gates binary not on PATH") {
		t.Fatalf("health failure must carry the install hint, got %q", err.Error())
	}
}
