package docsfreshness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// nearMissBody carries one canon §f, both near-miss shapes the io-surface
// incident paid for, and a same-letter heading that must NOT match.
const nearMissBody = "# Report\n\n" +
	"## f) Canon next tasks\n\n1. one\n\n" +
	"## §f — Follow-ups\n\n2. two\n\n" +
	"## f. Another near miss\n\n" +
	"## Feedback loop\n"

func writeProbeReport(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "2026-01-01_00-00_probe.md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	return path
}

// TestFindSectionFNearMiss pins the detector matrix: canon headings stay
// silent, `## §f` and `## f.` are named, `## Feedback` (no word boundary
// after the f) never matches.
func TestFindSectionFNearMiss(t *testing.T) {
	t.Parallel()

	got := FindSectionFNearMiss(writeProbeReport(t, nearMissBody))
	if len(got) != 2 {
		t.Fatalf("near misses = %d, want 2: %+v", len(got), got)
	}

	if got[0].LineNo != 7 || !strings.Contains(got[0].Text, "§f") {
		t.Errorf("first near miss = %+v, want line 7 carrying §f", got[0])
	}

	if got[1].LineNo != 11 || !strings.Contains(got[1].Text, "f. Another") {
		t.Errorf("second near miss = %+v, want line 11 with `## f.`", got[1])
	}
}

// TestFindSectionFGaps pins the omission matrix: a report with no canon
// §f heading is one whole-report gap; a §f section with bullets names
// every bullet with its line; canon numbered items and explicitly empty
// §f prose stay silent (the 11-24 audit's d2: freshness passed reports
// it provably read nothing from).
func TestFindSectionFGaps(t *testing.T) {
	t.Parallel()

	t.Run("missing heading", func(t *testing.T) {
		t.Parallel()

		got := FindSectionFGaps(writeProbeReport(t, "# Report\n\n## d) Findings\n\nprose only\n"))
		if len(got) != 1 || got[0].Kind != GapMissingHeading || got[0].LineNo != 0 {
			t.Fatalf("gaps = %+v, want one whole-report GapMissingHeading", got)
		}
	})

	t.Run("bullet-only f section", func(t *testing.T) {
		t.Parallel()

		body := "# Report\n\n## f) Next\n\n- wire the framblatz calibration\n- refresh the zorblat inventory\n"

		got := FindSectionFGaps(writeProbeReport(t, body))
		if len(got) != 2 || got[0].Kind != GapBulletItem || got[0].LineNo != 5 {
			t.Fatalf("gaps = %+v, want two bullet gaps, first at line 5", got)
		}

		if !strings.Contains(got[0].Text, "framblatz") {
			t.Errorf("bullet gap must carry the line text: %+v", got[0])
		}
	})

	t.Run("canon and empty prose stay silent", func(t *testing.T) {
		t.Parallel()

		numbered := FindSectionFGaps(writeProbeReport(t, "# Report\n\n## f) Next\n\n1. real item\n"))
		if len(numbered) != 0 {
			t.Errorf("canon numbered §f must not gap: %+v", numbered)
		}

		empty := FindSectionFGaps(writeProbeReport(t, "# Report\n\n## f) Next\n\nNone — everything routed.\n"))
		if len(empty) != 0 {
			t.Errorf("explicitly empty §f prose must not gap: %+v", empty)
		}
	})
}

// runWarnProbe writes one report with the given body into a fresh repo
// root, runs the gate, and returns the captured output — the shared
// harness for the two omission-WARN tests (same shape, different WARN).
func runWarnProbe(t *testing.T, body string) gatecli.CapturedOutput {
	t.Helper()

	root := t.TempDir()

	reportDir := filepath.Join(root, "docs", "status")
	if err := os.MkdirAll(reportDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(reportDir, "2026-01-01_00-00_probe.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	return gatecli.CaptureOutput(func() int {
		return Run([]string{root})
	})
}

// TestFreshnessRunWarnsOnSectionFGaps proves the omission WARN rides the
// gate's stdout on the PASS path — the surface where a bullet-only §f
// scored FRESH for weeks.
//
//nolint:paralleltest // process-global capture swap, serialized on purpose
func TestFreshnessRunWarnsOnSectionFGaps(t *testing.T) {
	captured := runWarnProbe(t, "# Report\n\n## f) Next\n\n- a bullet-only next step\n")

	if !strings.Contains(captured.Stdout, "WARN 2026-01-01_00-00_probe.md:5") {
		t.Errorf("stdout must carry the bullet WARN with report and line:\n%s", captured.Stdout)
	}

	if !strings.Contains(captured.Stdout, "not a harvestable item") {
		t.Errorf("the warn must name why the bullet reads nothing:\n%s", captured.Stdout)
	}

	if captured.Code != gatecli.ExitClean {
		t.Errorf("an omission WARN never flips the verdict: code = %d", captured.Code)
	}
}

// TestFreshnessRunWarnsOnNearMissSectionF proves the WARN rides the
// gate's stdout on the PASS path (the battery swallows stderr there) —
// the exact surface where the io-surface incident stayed invisible.
//
//nolint:paralleltest // process-global capture swap, serialized on purpose
func TestFreshnessRunWarnsOnNearMissSectionF(t *testing.T) {
	captured := runWarnProbe(t, "# Report\n\n## §f — Follow-ups\n\n1. A wholly unharvested forward item\n")

	if !strings.Contains(captured.Stdout, "WARN 2026-01-01_00-00_probe.md:3") {
		t.Errorf("stdout must carry the near-miss WARN with report and line:\n%s", captured.Stdout)
	}

	if !strings.Contains(captured.Stdout, "INVISIBLE") {
		t.Errorf("the warn must name what is lost:\n%s", captured.Stdout)
	}

	if captured.Code != gatecli.ExitClean {
		t.Errorf("a near miss WARNS, never flips the verdict: code = %d", captured.Code)
	}
}
