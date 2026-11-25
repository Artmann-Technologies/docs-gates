package docsfreshness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeReport(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()

	report := filepath.Join(dir, "2026-09-25_00-00_TEST.md")
	if err := os.WriteFile(report, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	return report
}

func TestHyphenTokenMatchesSpacedLine(t *testing.T) {
	t.Parallel()

	report := writeReport(t,
		"# Report\n\n## f) Next\n\n1. close the real-cax gap after the switch\n")
	living := NewCorpus([]string{"after the switch, build on the real cax evidence line"})

	findings := CheckReport(report, living, DefaultSettings())
	if len(findings) != 0 {
		t.Fatalf("hyphen/space mismatch must not drift, got %v", findings)
	}
}

func TestUnmatchedItemStillDrifts(t *testing.T) {
	t.Parallel()

	report := writeReport(t,
		"# Report\n\n## f) Next\n\n1. wire the zaphod-beeblebrox relay\n")
	living := NewCorpus([]string{"something entirely different"})

	findings := CheckReport(report, living, DefaultSettings())
	if len(findings) != 1 {
		t.Fatalf("want 1 finding, got %v", findings)
	}

	if !strings.Contains(findings[0][1], "missing tokens") {
		t.Fatalf("reason missing 'missing tokens': %q", findings[0][1])
	}
}

// TestTokensOnSeparateLinesAreCoincidence pins the 2026-11-24 fail-closed
// bar: an item whose distinctive tokens each match a DIFFERENT living
// line is coincidence, not a harvest — the shape that let a 50-item
// report score FRESH on incidental tokens (11-24 audit d2).
func TestTokensOnSeparateLinesAreCoincidence(t *testing.T) {
	t.Parallel()

	report := writeReport(t,
		"# Report\n\n## f) Next\n\n1. refresh the framblatz calibration\n")
	living := NewCorpus([]string{
		"the framblatz ledger is canonical",
		"calibration rituals run weekly here",
	})

	findings := CheckReport(report, living, DefaultSettings())
	if len(findings) != 1 {
		t.Fatalf("separate-line overlap must fail, got %v", findings)
	}

	if !strings.Contains(findings[0][1], "never share one living line") {
		t.Fatalf("reason must name the same-line rule: %q", findings[0][1])
	}
}

// TestSameLineQuorumPasses asserts the healthy path: one living line
// carrying both tokens (one distinctive) still counts as harvested.
func TestSameLineQuorumPasses(t *testing.T) {
	t.Parallel()

	report := writeReport(t,
		"# Report\n\n## f) Next\n\n1. refresh the framblatz calibration\n")
	living := NewCorpus([]string{
		"the framblatz calibration ledger is canonical",
		"an unrelated second line",
	})

	findings := CheckReport(report, living, DefaultSettings())
	if len(findings) != 0 {
		t.Fatalf("same-line quorum must pass, got %v", findings)
	}
}

// TestSameLineQuorumFindsTheSharedLine pins the corpus helper itself: the
// first line carrying the quorum is returned for attribution, and a
// corpus whose lines never co-carry the tokens reports !ok.
func TestSameLineQuorumFindsTheSharedLine(t *testing.T) {
	t.Parallel()

	living := NewCorpus([]string{
		"alpha only here",
		"beta alone there",
		"alpha and beta together",
	})

	matched, line, ok := living.SameLineQuorum([]string{"alpha", "beta"}, 2, GenericDF)
	if !ok {
		t.Fatal("the co-carrying line must satisfy the quorum")
	}

	if line != "alpha and beta together" {
		t.Fatalf("line = %q, want the co-carrying third line", line)
	}

	if len(matched) != 2 {
		t.Fatalf("matched = %v, want both tokens", matched)
	}

	if _, _, ok := living.SameLineQuorum([]string{"gamma", "delta"}, 2, GenericDF); ok {
		t.Fatal("absent tokens must never form a quorum")
	}
}

// TestSingleTokenItemStillUsesDistinctiveBar pins that the same-line rule
// tightened nothing for single-token items: a distinctive match passes, a
// generic-only corpus still fails with the generic reason.
func TestSingleTokenItemStillUsesDistinctiveBar(t *testing.T) {
	t.Parallel()

	report := writeReport(t,
		"# Report\n\n## f) Next\n\n1. framblatz\n")

	distinctive := NewCorpus([]string{"the framblatz ledger"})
	if findings := CheckReport(report, distinctive, DefaultSettings()); len(findings) != 0 {
		t.Fatalf("distinctive single token must pass, got %v", findings)
	}

	generic := make([]string, 0, GenericDF)
	for i := range GenericDF {
		generic = append(generic, "line "+string(rune('a'+i%26))+string(rune('a'+i%7))+" mentions framblatz somewhere")
	}

	findings := CheckReport(report, NewCorpus(generic), DefaultSettings())
	if len(findings) != 1 || !strings.Contains(findings[0][1], "only generic tokens") {
		t.Fatalf("generic single token must keep the generic reason, got %v", findings)
	}
}

func TestStatusReportsIncludesArchived(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	archived := filepath.Join(root, "docs", "status", "archived")
	if err := os.MkdirAll(archived, 0o755); err != nil {
		t.Fatal(err)
	}

	live := filepath.Join(root, "docs", "status", "2026-09-01_00-00_LIVE.md")
	old := filepath.Join(archived, "2026-08-01_00-00_OLD.md")

	for _, p := range []string{live, old} {
		if err := os.WriteFile(p, []byte("# Report\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	reports := statusReports(root, DefaultSettings())
	if len(reports) != 2 {
		t.Fatalf("want the live and the archived report, got %d: %v", len(reports), reports)
	}

	if !strings.HasSuffix(reports[0], filepath.Join("docs", "status", "2026-09-01_00-00_LIVE.md")) {
		t.Fatalf("live report must come first, got %s", reports[0])
	}

	if !strings.Contains(reports[1], "archived") {
		t.Fatalf("the archived report must be part of the replay corpus, got %s", reports[1])
	}
}
