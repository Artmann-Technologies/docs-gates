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
	living := []string{"after the switch, build on the real cax evidence line"}
	livingSquashed := []string{Squash(living[0])}

	// Mirrors Run(): matching is squash-augmented, genericity counts stay
	// on raw lines (the tuned basis).
	tokenDF := func(token string) int {
		n := 0

		for _, line := range living {
			if strings.Contains(line, token) {
				n++
			}
		}

		return n
	}

	findings := CheckReport(report, living, tokenDF, livingSquashed, DefaultSettings())
	if len(findings) != 0 {
		t.Fatalf("hyphen/space mismatch must not drift, got %v", findings)
	}
}

func TestUnmatchedItemStillDrifts(t *testing.T) {
	t.Parallel()

	report := writeReport(t,
		"# Report\n\n## f) Next\n\n1. wire the zaphod-beeblebrox relay\n")
	living := []string{"something entirely different"}
	livingSquashed := []string{Squash(living[0])}

	tokenDF := func(token string) int {
		n := 0

		for _, line := range living {
			if strings.Contains(line, token) {
				n++
			}
		}

		return n
	}

	findings := CheckReport(report, living, tokenDF, livingSquashed, DefaultSettings())
	if len(findings) != 1 {
		t.Fatalf("want 1 finding, got %v", findings)
	}

	if !strings.Contains(findings[0][1], "missing tokens") {
		t.Fatalf("reason missing 'missing tokens': %q", findings[0][1])
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
