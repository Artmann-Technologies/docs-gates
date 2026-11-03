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
