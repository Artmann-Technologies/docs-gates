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

// TestFreshnessRunWarnsOnNearMissSectionF proves the WARN rides the
// gate's stdout on the PASS path (the battery swallows stderr there) —
// the exact surface where the io-surface incident stayed invisible.
//
//nolint:paralleltest // process-global capture swap, serialized on purpose
func TestFreshnessRunWarnsOnNearMissSectionF(t *testing.T) {
	root := t.TempDir()

	reportDir := filepath.Join(root, "docs", "status")
	if err := os.MkdirAll(reportDir, 0o755); err != nil {
		t.Fatal(err)
	}

	body := "# Report\n\n## §f — Follow-ups\n\n1. A wholly unharvested forward item\n"

	if err := os.WriteFile(filepath.Join(reportDir, "2026-01-01_00-00_probe.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	captured := gatecli.CaptureOutput(func() int {
		return Run([]string{root})
	})

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
