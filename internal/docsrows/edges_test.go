package docsrows

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// TestFencedTablesAreIgnored pins that documentation examples inside fenced
// code blocks are never corpus: an unresolved example table must not be
// judged in any mode, and a fenced ledger marker must not exempt the next
// real table (the 2026-11-16 report §f.47 class).
//
//nolint:paralleltest // gatecli.CaptureOutput swaps process-global streams
func TestFencedTablesAreIgnored(t *testing.T) {
	for _, argv := range [][]string{
		{"testdata/fenced.md"},
		{"testdata/fenced.md", "--strict"},
	} {
		out := gatecli.CaptureOutput(func() int { return Run(argv) })

		if out.Code != gatecli.ExitFindings {
			t.Fatalf("%v: code = %d, want %d (stdout %q, stderr %q)",
				argv, out.Code, gatecli.ExitFindings, out.Stdout, out.Stderr)
		}

		if got := strings.Count(out.Stdout, "table at line"); got != 1 {
			t.Errorf("%v: %d table(s) judged, want exactly the one real partial table:\n%s",
				argv, got, out.Stdout)
		}

		if !strings.Contains(out.Stdout, "fenced.md: table at line 28") {
			t.Errorf("%v: the one judged table must be the real table at line 28:\n%s", argv, out.Stdout)
		}

		if strings.Contains(out.Stdout, "LEDGER-EXEMPT") {
			t.Errorf("%v: the fenced ledger marker exempted a real table:\n%s", argv, out.Stdout)
		}
	}
}

// TestCRLFClassifiesLikeLF pins that a report saved with Windows line endings
// classifies byte-identically to its LF twin: readLines strips the carriage
// returns before any cell or verdict matching sees them (report §f.45).
//
//nolint:paralleltest // gatecli.CaptureOutput swaps process-global streams
func TestCRLFClassifiesLikeLF(t *testing.T) {
	want := gatecli.CaptureOutput(func() int { return Run([]string{"testdata/partial.md"}) })
	got := gatecli.CaptureOutput(func() int { return Run([]string{"testdata/crlf.md"}) })

	if got.Code != want.Code {
		t.Fatalf("crlf code = %d, lf code = %d", got.Code, want.Code)
	}

	normalized := strings.ReplaceAll(got.Stdout, "crlf.md", "partial.md")

	if normalized != want.Stdout {
		t.Errorf("crlf findings differ from lf:\n--- lf ---\n%s\n--- crlf ---\n%s",
			want.Stdout, normalized)
	}
}

// TestEmptyFileIsClean pins that a zero-byte report is a clean file, never a
// crash or a usage error (report §f.44).
//
//nolint:paralleltest // gatecli.CaptureOutput swaps process-global streams
func TestEmptyFileIsClean(t *testing.T) {
	out := gatecli.CaptureOutput(func() int { return Run([]string{"testdata/empty.md"}) })

	if out.Code != gatecli.ExitClean {
		t.Fatalf("code = %d, want %d (stdout %q, stderr %q)",
			out.Code, gatecli.ExitClean, out.Stdout, out.Stderr)
	}

	if !strings.Contains(out.Stdout, "1 file(s) clean") {
		t.Errorf("stdout missing the clean summary:\n%s", out.Stdout)
	}
}

// TestPerCellStrikeShapesStayClean pins the archived-report repair shapes
// end-to-end: per-cell strikes with the verdict word outside the strike, the
// superseded (s-kind) verdict, the Won't-implement (w-kind) wording, and a
// routed (r-kind) row resolving without any strike (reports §f.28/29/30).
//
//nolint:paralleltest // gatecli.CaptureOutput swaps process-global streams
func TestPerCellStrikeShapesStayClean(t *testing.T) {
	out := gatecli.CaptureOutput(func() int { return Run([]string{"testdata/percell.md"}) })

	if out.Code != gatecli.ExitClean {
		t.Fatalf("code = %d, want %d (stdout %q, stderr %q)",
			out.Code, gatecli.ExitClean, out.Stdout, out.Stderr)
	}
}

// TestEscapedPipeInsideStruckCell pins that an escaped pipe inside a struck
// cell keeps the row one row: the strike must survive the split (§f.46).
func TestEscapedPipeInsideStruckCell(t *testing.T) {
	t.Parallel()

	line := `| ~~4~~ | ~~cmd \| flag~~ | ~~x.md~~ |`

	if got := classifyRow(line); got != stateStruck {
		t.Errorf("classifyRow = %q, want %q", got, stateStruck)
	}
}

// TestRowsIgnoresRatchetConfig pins the exit-code contract: rows owns 0/1/2
// and never inherits the annotate gate's ratchet — no .docs-gates.json key
// can silence a findings exit (reports §f.25/26).
//
//nolint:paralleltest // gatecli.CaptureOutput swaps process-global streams
func TestRowsIgnoresRatchetConfig(t *testing.T) {
	dir := t.TempDir()

	config := `{"annotate_max": 0, "audit_max": 0}`

	if err := os.WriteFile(filepath.Join(dir, ".docs-gates.json"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile("testdata/partial.md")
	if err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(dir, "partial.md")

	if err := os.WriteFile(target, data, 0o644); err != nil {
		t.Fatal(err)
	}

	out := gatecli.CaptureOutput(func() int { return Run([]string{target}) })

	if out.Code != gatecli.ExitFindings {
		t.Errorf("code = %d, want %d even with annotate_max/audit_max = 0 (stdout %q, stderr %q)",
			out.Code, gatecli.ExitFindings, out.Stdout, out.Stderr)
	}
}
