package battery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunJSONFace(t *testing.T) {
	root := writeRoot(t)

	out := gatecli.CaptureOutput(func() int {
		return Run([]string{root, flagJSON})
	})

	if out.Code != gatecli.ExitClean {
		t.Fatalf("green corpus must stay exit 0 under --json, got rc %d; stderr: %s", out.Code, out.Stderr)
	}

	var report batteryReport
	if err := json.Unmarshal([]byte(out.Stdout), &report); err != nil {
		t.Fatalf("stdout must be one JSON document: %v; stdout: %s", err, out.Stdout)
	}

	if !report.Pass || report.Root != root || len(report.Gates) != 8 {
		t.Fatalf("unexpected report: pass=%v root=%q gates=%d", report.Pass, report.Root, len(report.Gates))
	}

	for _, g := range report.Gates {
		if g.Verdict != verdictPass || g.Skipped || g.TimedOut || g.ExitCode != gatecli.ExitClean {
			t.Fatalf("green JSON row must be clean PASS: %+v", g)
		}
	}

	for _, human := range []string{"docs-gates: PASS"} {
		if strings.Contains(out.Stdout, human) {
			t.Fatalf("human line %q leaked into the JSON stdout: %s", human, out.Stdout)
		}

		if !strings.Contains(out.Stderr, human) {
			t.Fatalf("human line %q missing from stderr: %s", human, out.Stderr)
		}
	}

	if strings.Contains(out.Stderr, "verdict matrix") {
		t.Fatalf("the human matrix must not render in JSON mode: %s", out.Stderr)
	}
}

// The JSON face over a red corpus: the failing gate's row is FAIL with
// its captured findings embedded, the battery exits 1, and the green
// gates stay PASS — the matrix arithmetic, not the human formatting.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunJSONFaceRedGate(t *testing.T) {
	root := writeRoot(t)

	var struck strings.Builder
	for range 21 {
		struck.WriteString("| ~~task~~ done (evidence) | ✅ `DONE` | P1 | 5m | ev |\n")
	}

	body := "# TODO\n\n" + struck.String()
	if err := os.WriteFile(filepath.Join(root, "TODO_LIST.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	out := gatecli.CaptureOutput(func() int {
		return Run([]string{root, flagJSON})
	})

	if out.Code != gatecli.ExitFindings {
		t.Fatalf("red corpus must exit 1 under --json, got rc %d", out.Code)
	}

	var report batteryReport
	if err := json.Unmarshal([]byte(out.Stdout), &report); err != nil {
		t.Fatalf("stdout must be one JSON document: %v; stdout: %s", err, out.Stdout)
	}

	var cadence *gateReport

	for i := range report.Gates {
		if report.Gates[i].Name == "cadence" {
			cadence = &report.Gates[i]
		}
	}

	if cadence == nil || cadence.Verdict != verdictFail || cadence.TimedOut {
		t.Fatalf("cadence row must be plain FAIL: %+v", cadence)
	}

	if !strings.Contains(cadence.Stderr, "OVERDUE") {
		t.Fatalf("FAIL row must embed the gate's captured stderr: %+v", cadence)
	}
}

// The JSON face's timeout row: a wedged gate renders FAIL with
// timed_out=true and the battery still returns (bounded), the gate
// behind it still runs.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunJSONTimeoutRow(t *testing.T) {
	wedged := gate{name: "wedge", fn: func([]string) int { select {} }, args: nil}
	healthy := gate{name: "after-wedge", fn: func([]string) int { return gatecli.ExitClean }, args: nil}

	captured := gatecli.CaptureOutput(func() int {
		rows := runGates([]gate{wedged, healthy}, nil, 50*time.Millisecond, true)

		return renderJSON(rows, "root")
	})

	if captured.Code != gatecli.ExitFindings {
		t.Fatalf("a timed-out row must fail the JSON render, got rc %d", captured.Code)
	}

	var report batteryReport
	if err := json.Unmarshal([]byte(captured.Stdout), &report); err != nil {
		t.Fatalf("stdout must be one JSON document: %v; stdout: %s", err, captured.Stdout)
	}

	if len(report.Gates) != 2 || report.Pass {
		t.Fatalf("wedge matrix must fail with two rows: %+v", report)
	}

	if report.Gates[0].Name != "wedge" || !report.Gates[0].TimedOut || report.Gates[0].Verdict != verdictFail {
		t.Fatalf("wedge row must be FAIL timed_out: %+v", report.Gates[0])
	}

	if report.Gates[1].Verdict != verdictPass {
		t.Fatalf("the gate behind the wedge must stay PASS: %+v", report.Gates[1])
	}
}

// TestRunGateHalfBudgetWarnsLive pins the early warning: a gate past
// half its wall-clock budget names the cap on the REAL stdout while it
// still runs (Livef escapes the capture swap — the warning must be
// live, not parked in the captured buffer the battery replays only
// after the gate returns).
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunGateHalfBudgetWarnsLive(t *testing.T) {
	slow := gate{
		name: "slowpoke",
		fn: func([]string) int {
			time.Sleep(1 * time.Second)

			return gatecli.ExitClean
		},
		args: nil,
	}

	out := gatecli.CaptureOutput(func() int {
		rows := runGates([]gate{slow}, nil, 300*time.Millisecond, false)

		if renderMatrix(rows) {
			return gatecli.ExitFindings
		}

		return gatecli.ExitClean
	})

	if !strings.Contains(out.Stdout, "slowpoke over half its 300ms budget") {
		t.Fatalf("half-budget warning must reach the live stdout, got: %s", out.Stdout)
	}
}
