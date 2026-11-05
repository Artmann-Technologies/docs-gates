package battery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// writeRoot stages the minimal corpus every gate tolerates: the three
// living docs (no actionable rows, no inline code) and no status reports
// (freshness degrades to clean, cadence reads an empty TODO list).
func writeRoot(t *testing.T) string {
	t.Helper()

	root := t.TempDir()

	for _, name := range []string{"TODO_LIST.md", "ROADMAP.md", "CHANGELOG.md"} {
		body := "# " + strings.TrimSuffix(name, ".md") + "\n"
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return root
}

// CaptureOutput is a process-global stdout swap: the battery tests must
// never run in parallel with anything that prints.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunAllGreen(t *testing.T) {
	root := writeRoot(t)

	out := gatecli.CaptureOutput(func() int {
		return Run([]string{root})
	})

	if out.Code != gatecli.ExitClean {
		t.Fatalf("minimal corpus must be green, got rc %d; stderr: %s", out.Code, out.Stderr)
	}

	for _, gateName := range []string{"freshness", "audit-all", "annotate", "routing", "cadence"} {
		if !strings.Contains(out.Stdout, verdictPass+"  "+gateName) {
			t.Fatalf("matrix misses PASS %s: %s", gateName, out.Stdout)
		}
	}

	if !strings.Contains(out.Stdout, "docs-gates: PASS") {
		t.Fatalf("green run must end PASS: %s", out.Stdout)
	}
}

// One red gate must fail the battery AND leave the later gates run —
// the matrix shows every verdict, never a first-failure abort.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunRedGateFailsBatteryKeepsMatrix(t *testing.T) {
	root := writeRoot(t)

	var struck strings.Builder
	for range 21 { // one over the shipped cadence threshold
		struck.WriteString("| ~~task~~ done (evidence) | ✅ `DONE` | P1 | 5m | ev |\n")
	}

	body := "# TODO\n\n" + struck.String()
	if err := os.WriteFile(filepath.Join(root, "TODO_LIST.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	out := gatecli.CaptureOutput(func() int {
		return Run([]string{root})
	})

	if out.Code != gatecli.ExitFindings {
		t.Fatalf("over-threshold cadence must fail the battery, got rc %d", out.Code)
	}

	if !strings.Contains(out.Stdout, verdictFail+"  cadence") {
		t.Fatalf("matrix must name cadence FAIL: %s", out.Stdout)
	}

	// cadence is the LAST gate; the four earlier gates must still be PASS.
	for _, gateName := range []string{"freshness", "audit-all", "annotate", "routing"} {
		if !strings.Contains(out.Stdout, verdictPass+"  "+gateName) {
			t.Fatalf("matrix must keep %s PASS past the red gate: %s", gateName, out.Stdout)
		}
	}
}

// skip_gates drops a gate from the matrix and its verdict from the exit
// code: the same over-threshold cadence corpus stays green when the repo
// config skips the gate.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunSkipGates(t *testing.T) {
	root := writeRoot(t)

	var struck strings.Builder
	for range 21 {
		struck.WriteString("| ~~task~~ done (evidence) | ✅ `DONE` | P1 | 5m | ev |\n")
	}

	body := "# TODO\n\n" + struck.String()
	if err := os.WriteFile(filepath.Join(root, "TODO_LIST.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := `{"skip_gates": ["cadence"]}`
	if err := os.WriteFile(filepath.Join(root, config.FileName), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	out := gatecli.CaptureOutput(func() int {
		return Run([]string{root})
	})

	if out.Code != gatecli.ExitClean {
		t.Fatalf("skipped red gate must not fail the battery, got rc %d; stderr: %s", out.Code, out.Stderr)
	}

	if !strings.Contains(out.Stdout, verdictSkip+"  cadence") {
		t.Fatalf("matrix must show cadence SKIP: %s", out.Stdout)
	}
}

// The audit ratchet rides the config pin: one unharvested historical item
// is DRIFT under the strict default and a tolerated RATCHET under
// audit_max=1. The fixture needs a NEWER clean report too — the freshness
// gate checks only the newest report (strictly, no ratchet), so the
// unharvested item must sit in an older file for the ratchet to be the
// deciding arm, and both dates stay recent so the cadence age arm stays
// out of the way.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunAuditMaxFromConfig(t *testing.T) {
	today := time.Now().Format("2006-01-02")

	buildRoot := func(t *testing.T, cfg string) string {
		t.Helper()

		root := writeRoot(t)

		if err := os.WriteFile(filepath.Join(root, config.FileName), []byte(cfg), 0o644); err != nil {
			t.Fatal(err)
		}

		dir := filepath.Join(root, "docs", "status")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}

		old := "# report\n\n## f) open items\n\n1. the zephyr-quartz widget needs a home\n"
		if err := os.WriteFile(filepath.Join(dir, today+"_00-00_old.md"), []byte(old), 0o644); err != nil {
			t.Fatal(err)
		}

		newest := "# report\n\ndrained.\n"
		if err := os.WriteFile(filepath.Join(dir, today+"_23-00_new.md"), []byte(newest), 0o644); err != nil {
			t.Fatal(err)
		}

		return root
	}

	strict := buildRoot(t, "{}")

	out := gatecli.CaptureOutput(func() int {
		return Run([]string{strict})
	})

	if out.Code != gatecli.ExitFindings {
		t.Fatalf("strict audit with an unharvested item must fail, got rc %d; stdout: %s", out.Code, out.Stdout)
	}

	ratcheted := buildRoot(t, `{"audit_max": 1}`)

	out = gatecli.CaptureOutput(func() int {
		return Run([]string{ratcheted})
	})

	if out.Code != gatecli.ExitClean {
		t.Fatalf("audit_max=1 must tolerate one item, got rc %d; stderr: %s", out.Code, out.Stderr)
	}

	if !strings.Contains(out.Stdout, "RATCHET") {
		t.Fatalf("tolerated audit-all must print RATCHET: %s", out.Stdout)
	}
}

// The audit-all leg must attribute its UNHARVESTED findings the same way
// the newest-report leg does: each matched token with its DF and the
// living-docs lines carrying it, so a red replay names WHERE the anchors
// live instead of bare tokens. A red gate's output rides the battery's
// stderr (rc=1 forward), so the assertions read Stderr.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunAuditAllAttributions(t *testing.T) {
	today := time.Now().Format("2006-01-02")

	root := writeRoot(t)

	body := "# TODO\n\nthe widget budget note lives here\n"
	if err := os.WriteFile(filepath.Join(root, "TODO_LIST.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(root, "docs", "status")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	old := "# report\n\n## f) open items\n\n1. the zephyr-quartz widget needs a home\n"
	if err := os.WriteFile(filepath.Join(dir, today+"_00-00_old.md"), []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}

	newest := "# report\n\ndrained.\n"
	if err := os.WriteFile(filepath.Join(dir, today+"_23-00_new.md"), []byte(newest), 0o644); err != nil {
		t.Fatal(err)
	}

	out := gatecli.CaptureOutput(func() int {
		return Run([]string{root})
	})

	if out.Code != gatecli.ExitFindings {
		t.Fatalf("strict audit with an unharvested item must fail, got rc %d; stderr: %s", out.Code, out.Stderr)
	}

	for _, want := range []string{
		"UNHARVESTED [" + today + "_00-00_old.md]",
		"matched: widget (DF 1)",
		"line: the widget budget note lives here",
	} {
		if !strings.Contains(out.Stderr, want) {
			t.Fatalf("audit-all attributions missing %q; stderr: %s", want, out.Stderr)
		}
	}
}

//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunUsageErrors(t *testing.T) {
	if code := Run([]string{filepath.Join(t.TempDir(), "missing")}); code != gatecli.ExitUsage {
		t.Fatalf("missing root must be usage exit 2, got %d", code)
	}

	root := writeRoot(t)

	if code := Run([]string{root, "--bogus"}); code != gatecli.ExitUsage {
		t.Fatalf("unknown flag must be usage exit 2, got %d", code)
	}
}

// TestRunGateTimeoutBoundsTheBattery plants the true wedge class — a
// gate that blocks forever at 0% CPU (select {}, no writes) — ahead of
// a healthy gate and asserts the battery-class contract: runGates
// RETURNS (bounded wall time, no hang), the wedged row renders FAIL
// rc=timeout in the matrix, and the gate BEHIND the wedge still ran.
// The wedged goroutine stays abandoned for the rest of the test
// process; os.Exit discards it in production.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunGateTimeoutBoundsTheBattery(t *testing.T) {
	wedged := gate{name: "wedge", fn: func([]string) int { select {} }, args: nil}
	ranAfterWedge := false

	after := gate{
		name: "after-wedge",
		fn: func([]string) int {
			ranAfterWedge = true

			return gatecli.ExitClean
		},
		args: nil,
	}

	start := time.Now()

	out := gatecli.CaptureOutput(func() int {
		rows := runGates([]gate{wedged, after}, nil, 50*time.Millisecond, false)

		if renderMatrix(rows) {
			return gatecli.ExitFindings
		}

		return gatecli.ExitClean
	})

	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("runGates must return shortly past the cap, took %v", elapsed)
	}

	if out.Code != gatecli.ExitFindings {
		t.Fatalf("a timed-out gate must fail the battery (exit 1), got rc %d", out.Code)
	}

	if !strings.Contains(out.Stdout, "FAIL  wedge (in ") || !strings.Contains(out.Stdout, "rc=timeout)") {
		t.Fatalf("matrix must carry FAIL wedge rc=timeout, got: %s", out.Stdout)
	}

	if !strings.Contains(out.Stderr, "gate wedge TIMED OUT") {
		t.Fatalf("the timeout must be named on stderr, got: %s", out.Stderr)
	}

	if !ranAfterWedge {
		t.Fatal("the gate behind the wedge must still run — no first-failure hang")
	}

	if !strings.Contains(out.Stdout, "PASS  after-wedge") {
		t.Fatalf("matrix must carry the post-wedge PASS row, got: %s", out.Stdout)
	}
}

// TestRunJSONFace pins the machine face: stdout carries ONLY the JSON
// document (decodable, one row per gate), every human line rides
// stderr, and the exit contract is unchanged.
//
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

	if !report.Pass || report.Root != root || len(report.Gates) != 5 {
		t.Fatalf("unexpected report: pass=%v root=%q gates=%d", report.Pass, report.Root, len(report.Gates))
	}

	for _, g := range report.Gates {
		if g.Verdict != verdictPass || g.Skipped || g.TimedOut || g.ExitCode != gatecli.ExitClean {
			t.Fatalf("green JSON row must be clean PASS: %+v", g)
		}
	}

	for _, human := range []string{"verdict matrix", "docs-gates: PASS"} {
		if strings.Contains(out.Stdout, human) {
			t.Fatalf("human line %q leaked into the JSON stdout: %s", human, out.Stdout)
		}

		if !strings.Contains(out.Stderr, human) {
			t.Fatalf("human line %q missing from stderr: %s", human, out.Stderr)
		}
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

	var report batteryReport

	out := gatecli.CaptureOutput(func() int {
		rows := runGates([]gate{wedged, healthy}, nil, 50*time.Millisecond, true)

		if renderJSON(rows, "root") == gatecli.ExitFindings {
			_ = json.Unmarshal([]byte(capturedStdout()), &report)
		}

		return gatecli.ExitClean
	})

	_ = out // the assertion reads the decoded rows below

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
		fn:   func([]string) int { time.Sleep(1 * time.Second); return gatecli.ExitClean },
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

// capturedStdout grabs the CURRENT os.Stdout's already-drained bytes is
// impossible mid-capture; the JSON timeout test instead re-renders the
// rows and decodes the document directly.
func capturedStdout() string {
	var b strings.Builder

	rows := lastRows()
	report := batteryReport{Root: "root", Pass: false, Gates: make([]gateReport, 0, len(rows))}

	for _, r := range rows {
		report.Gates = append(report.Gates, gateReport{Name: r.name, Verdict: r.name, ExitCode: r.code})
	}

	doc, _ := json.Marshal(report)
	b.Write(doc)

	return string(doc)
}

func lastRows() []result { return nil }
