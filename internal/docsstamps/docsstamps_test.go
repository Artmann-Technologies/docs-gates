package docsstamps

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

func stampDate(daysAgo int) string {
	return time.Now().AddDate(0, 0, -daysAgo).Format("2006-01-02")
}

func writeStampedGlossary(t *testing.T, stamp string) string {
	t.Helper()

	root := t.TempDir()

	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}

	body := "# Domain Language\n\n> Verified against code on " + stamp + ".\n"

	if err := os.WriteFile(filepath.Join(root, "docs", "DOMAIN_LANGUAGE.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	return root
}

func writeConfig(t *testing.T, root, json string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(root, config.FileName), []byte(json), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFreshStampIsClean(t *testing.T) {
	t.Parallel()

	if rc := Run([]string{writeStampedGlossary(t, stampDate(1))}); rc != gatecli.ExitClean {
		t.Fatalf("a one-day-old stamp must be clean, got rc %d", rc)
	}
}

func TestWarnBarStampStaysCleanEvenEnforced(t *testing.T) {
	t.Parallel()

	root := writeStampedGlossary(t, stampDate(config.DefaultStampWarnDays+10))
	writeConfig(t, root, `{"stamp_enforce": true}`)

	if rc := Run([]string{root}); rc != gatecli.ExitClean {
		t.Fatalf("a warn-bar stamp is informational at every enforcement level, got rc %d", rc)
	}
}

func TestFailBarStampWarnsByDefault(t *testing.T) {
	t.Parallel()

	if rc := Run(
		[]string{writeStampedGlossary(t, stampDate(config.DefaultStampFailDays+20))},
	); rc != gatecli.ExitClean {
		t.Fatalf("a fail-bar stamp must WARN-exit-clean in default warn mode, got rc %d", rc)
	}
}

func TestFailBarStampFailsWhenEnforced(t *testing.T) {
	t.Parallel()

	root := writeStampedGlossary(t, stampDate(config.DefaultStampFailDays+20))
	writeConfig(t, root, `{"stamp_enforce": true}`)

	if rc := Run([]string{root}); rc != gatecli.ExitFindings {
		t.Fatalf("an enforced fail-bar stamp must fail, got rc %d", rc)
	}
}

func TestNoStampsIsSkipNotFinding(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(root, "docs", "DOMAIN_LANGUAGE.md"),
		[]byte("# Domain Language\n\nglossary without any verification claim\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	if rc := Run([]string{root}); rc != gatecli.ExitClean {
		t.Fatalf("a corpus with no stamps owes nothing, got rc %d", rc)
	}
}

// The fleet grammar, from the census sample: mixed-case verify words,
// "verified live", dates inside prose segments, and table cells whose
// pipes END the segment (a date in the next cell must not attach to a
// verify word in the previous one).
func TestDatesReadsTheFleetGrammar(t *testing.T) {
	t.Parallel()

	content := "re-verified 2026-01-02 ok\n" +
		"Verified empirically 2026-03-04\n" +
		"verified live, 2026-05-06\n" +
		"refreshed 2026-07-08 by docs-health\n" +
		"Rebuilt 2026-09-10 by tooling\n" +
		"| verified thing | 2026-11-12 |\n" +
		"not a stamp: 2026-02-02 alone\n"

	dates := Dates(content)
	if len(dates) != 5 {
		t.Fatalf("grammar: got %d dates, want 5 (pipe boundary must cut the table row): %v", len(dates), dates)
	}
}

func TestScanOldestWins(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}

	body := "> verified " + stampDate(30) + "\n> re-verified " + stampDate(100) + "\n"

	if err := os.WriteFile(filepath.Join(root, "docs", "DOMAIN_LANGUAGE.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	findings, skips, err := Scan(root, Settings{
		Files:    []string{config.DefaultDomainLanguageFile},
		WarnDays: config.DefaultStampWarnDays,
		FailDays: config.DefaultStampFailDays,
		Now:      time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(skips) != 0 || len(findings) != 1 {
		t.Fatalf("oldest stamp must be the one finding, got findings=%v skips=%v", findings, skips)
	}

	if findings[0].AgeDays < 99 || findings[0].Severity != "warn" {
		t.Fatalf("finding must carry the 100d-old warn stamp, got %+v", findings[0])
	}
}
