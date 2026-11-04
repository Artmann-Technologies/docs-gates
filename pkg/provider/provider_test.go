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
// canned output and exits with the canned code, then prepends its dir to
// PATH so the provider's PATH-based exec resolves the fake. Not parallel:
// t.Setenv is process-global.
//

func writeFakeBinary(t *testing.T, script string, exitCode int) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, toolName)
	body := "#!/usr/bin/env bash\ncat <<'MATRIX'\n" + script + "\nMATRIX\nexit " + strconv.Itoa(exitCode) + "\n"

	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestParseMatrixKeepsFailOrder(t *testing.T) {
	t.Parallel()

	out := "docs-gates: verdict matrix\n" +
		"  PASS  freshness\n" +
		"  FAIL  cadence (rc=1)\n" +
		"  SKIP  annotate (skip_gates)\n" +
		"  FAIL  annotate (rc=1)\n" +
		"docs-gates: FAIL — the matrix above names the red gates\n"

	got := parseMatrix(out)
	want := []string{"cadence", "annotate"}

	if len(got) != len(want) {
		t.Fatalf("parseMatrix = %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("parseMatrix[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestParseMatrixIgnoresProse(t *testing.T) {
	t.Parallel()

	// Prose carrying FAIL names (a red gate's own stderr echo) must not
	// parse as a verdict row: only the two-space-indented matrix rows count.
	out := "docs-gates: gate cadence FAILED (rc=1):\nFAIL cadence in prose\n  PASS  freshness\n"

	if got := parseMatrix(out); len(got) != 0 {
		t.Fatalf("prose must not parse as verdicts, got %v", got)
	}
}

//nolint:paralleltest // process-global PATH swap, serialized on purpose
func TestDetectRedGatesBecomeFindings(t *testing.T) {
	writeFakeBinary(t, "  PASS  freshness\n  FAIL  cadence (rc=1)\n  FAIL  routing (rc=1)\n", 1)

	findings, err := Provider.Detect.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if len(findings) != 2 {
		t.Fatalf("two red gates must yield two findings, got %d: %+v", len(findings), findings)
	}

	if findings[0].Rule != "cadence" || findings[1].Rule != "routing" {
		t.Fatalf("rules = %q, %q; want cadence, routing", findings[0].Rule, findings[1].Rule)
	}

	if findings[0].Severity != finding.SeverityError {
		t.Fatalf("red gates are errors, got %q", findings[0].Severity)
	}

	if findings[0].ToolName != toolName {
		t.Fatalf("tool name = %q, want %q", findings[0].ToolName, toolName)
	}
}

//nolint:paralleltest // process-global PATH swap, serialized on purpose
func TestDetectCleanCorpusYieldsZeroFindings(t *testing.T) {
	writeFakeBinary(t, "  PASS  freshness\n  PASS  cadence\n", 0)

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

//nolint:paralleltest // process-global PATH swap, serialized on purpose
func TestDetectRedExitWithoutMatrixIsItsOwnFinding(t *testing.T) {
	writeFakeBinary(t, "", 1)

	findings, err := Provider.Detect.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if len(findings) != 1 || findings[0].Rule != "battery-exit-without-matrix" {
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
