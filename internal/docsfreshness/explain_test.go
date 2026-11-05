package docsfreshness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

func TestFindHtmlBlindSpot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	status := filepath.Join(root, "docs", "status")
	if err := os.MkdirAll(status, 0o755); err != nil {
		t.Fatal(err)
	}

	if s := FindHtmlBlindSpot(root, DefaultSettings()); s.Count != 0 || s.Newest != "" || s.Outranks {
		t.Fatalf("empty status dir must yield a zero spot, got %+v", s)
	}

	write := func(name string) {
		t.Helper()

		if err := os.WriteFile(filepath.Join(status, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// The exact production shape (2026-10-07): an .md report plus an
	// OLDER .html companion, then a NEWER .html the .md scan skipped.
	write("2026-10-07_13-10_md-report.md")

	if s := FindHtmlBlindSpot(root, DefaultSettings()); s.Count != 0 {
		t.Fatalf("no html yet, got %+v", s)
	}

	write("2026-10-07_10-00_older.html")

	if s := FindHtmlBlindSpot(root, DefaultSettings()); s.Count != 1 || s.Outranks {
		t.Fatalf("older html must count but not outrank, got %+v", s)
	}

	write("2026-10-07_16-25_mail-stack-status.html")

	s := FindHtmlBlindSpot(root, DefaultSettings())
	if s.Count != 2 || !s.Outranks || s.Newest != "2026-10-07_16-25_mail-stack-status.html" {
		t.Fatalf("newer html must outrank the newest .md, got %+v", s)
	}
}

// --explain-item prints the token table for ANY matching open item —
// harvested ones included — because --explain-token tables ride
// UNHARVESTED findings only and a FRESH corpus therefore prints none
// (the §43 row: AGENTS claimed a table that never came).
func TestExplainItemTablesHarvestedItems(t *testing.T) {
	t.Parallel()

	root := preFlightRoot(t)
	living := loadLivingDocs(root, DefaultSettings())

	rc := ExplainItem(root, "alpha-probe", living, DefaultSettings())
	if rc != gatecli.ExitClean {
		t.Fatalf("diagnostic mode always exits clean, got %d", rc)
	}
}

func TestExplainItemRunEndToEnd(t *testing.T) {
	t.Parallel()

	root := preFlightRoot(t)

	if rc := Run([]string{root, "--explain-item", "alpha-probe"}); rc != gatecli.ExitClean {
		t.Fatalf("matching needle must exit clean, got %d", rc)
	}

	if rc := Run([]string{root, "--explain-item", "no-such-needle"}); rc != gatecli.ExitClean {
		t.Fatalf("unmatched needle must also exit clean (report-only), got %d", rc)
	}

	if rc := Run([]string{root, "--explain-item"}); rc != gatecli.ExitUsage {
		t.Fatalf("missing needle must exit usage, got %d", rc)
	}
}

// Struck (~~) and skip-marked items carry no harvest margin and must not
// match in --explain-item — same open-item filter as CheckReport.
// Deliberately NOT parallel: the only output-capturing test in the suite,
// it swaps the global os.Stdout, which would race every other parallel
// test's stdout writes under test-race.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel(); all other tests only assert exit codes
func TestExplainItemSkipsStruckAndExemptItems(t *testing.T) {
	root := preFlightRoot(t)
	struck := filepath.Join(root, "docs", "status", "2026-09-02_00-00_STRUCK.md")

	body := "# Report\n\n## f) Next\n\n1. ~~alpha-probe thing~~ done\n2. alpha-probe exempt <!-- fresh:skip -->\n"
	if err := os.WriteFile(struck, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	living := loadLivingDocs(root, DefaultSettings())

	captured := gatecli.CaptureOutput(func() int {
		return ExplainItem(root, "exempt", living, DefaultSettings())
	})

	if captured.Code != gatecli.ExitClean {
		t.Fatalf("diagnostic mode always exits clean, got %d", captured.Code)
	}

	if got := captured.Stdout; strings.Contains(got, "exempt") && strings.Contains(got, "verdict:") {
		t.Fatalf("struck/skip-marked items must not be tabled, got output:\n%s", got)
	}
}
