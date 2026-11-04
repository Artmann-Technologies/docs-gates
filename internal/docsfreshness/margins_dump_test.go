package docsfreshness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// TestTokenMarginsNamesNearBarTokens: the preFlightRoot corpus rides
// alpha-probe at DF 7 (bar 8) — the margins report must name it as one
// line away from genericizing, and stay exit-clean.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestTokenMarginsNamesNearBarTokens(t *testing.T) {
	root := preFlightRoot(t)

	captured := gatecli.CaptureOutput(func() int {
		return TokenMargins(root, DefaultSettings())
	})

	if captured.Code != gatecli.ExitClean {
		t.Fatalf("the margins report is informational and must stay clean, got %d", captured.Code)
	}

	if !strings.Contains(captured.Stdout, "alpha-probe") ||
		!strings.Contains(captured.Stdout, "1 more line(s) genericize it") {
		t.Fatalf("margins must name the DF-7 token and its remaining budget, stdout:\n%s", captured.Stdout)
	}
}

// TestDumpItemsExtractsAndTokens: the report's §f items print with their
// tokens; an unreadable report and a report without §f are usage
// refusals naming the cause.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestDumpItemsExtractsAndTokens(t *testing.T) {
	dir := t.TempDir()

	report := filepath.Join(dir, "2026-11-01_00-00_X.md")
	body := "# X\n\n## f) open items\n\n1. the qxvorlag spindle reconciliation needs a pass\n"

	if err := os.WriteFile(report, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	captured := gatecli.CaptureOutput(func() int {
		return DumpItems(report)
	})

	if captured.Code != gatecli.ExitClean {
		t.Fatalf("dump-items over a §f report must stay clean, got %d", captured.Code)
	}

	for _, want := range []string{"1 open item(s)", "qxvorlag spindle reconciliation", "tokens: [qxvorlag"} {
		if !strings.Contains(captured.Stdout, want) {
			t.Fatalf("dump-items output must contain %q, stdout:\n%s", want, captured.Stdout)
		}
	}

	if rc := DumpItems(filepath.Join(dir, "nope.md")); rc != gatecli.ExitUsage {
		t.Fatalf("an unreadable report must exit usage, got %d", rc)
	}

	empty := filepath.Join(dir, "2026-11-02_00-00_empty.md")
	if err := os.WriteFile(empty, []byte("# no section here\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	captured = gatecli.CaptureOutput(func() int {
		return DumpItems(empty)
	})

	if captured.Code != gatecli.ExitUsage {
		t.Fatalf("a report without §f must exit usage, got %d", captured.Code)
	}

	if !strings.Contains(captured.Stderr, "no §f section") {
		t.Fatalf("the no-section refusal must say so, stderr:\n%s", captured.Stderr)
	}
}

// TestHarvestAttributionNamesMatchedLines: a failing item's attributions
// carry matched tokens with DF and example living lines — the preflight
// culprit idea extended to the harvest findings. The alpha-probe corpus
// rides webphone at DF 8 (generic at the bar) and alpha-probe at DF 7
// (still distinctive): both must appear, only the first tagged generic.
//
//nolint:paralleltest // shares the helper corpus with the margins test style
func TestHarvestAttributionNamesMatchedLines(t *testing.T) {
	root := preFlightRoot(t)

	lines := loadLivingLines(t, root)

	out := harvestAttributions(
		"1. webphone alpha-probe calibration recheck",
		lines,
		squashLines(lines),
		tokenDFOver(lines),
		DefaultSettings().df(),
	)

	joined := strings.Join(out, "\n")

	if !strings.Contains(joined, "matched: webphone (DF 8, generic)") {
		t.Fatalf("the attribution must name the generic match with its DF, got:\n%s", joined)
	}

	if !strings.Contains(joined, "matched: alpha-probe (DF 7)") {
		t.Fatalf("the distinctive match must be named without the generic tag, got:\n%s", joined)
	}

	if !strings.Contains(joined, "line: ") {
		t.Fatalf("the attribution must show an example living line, got:\n%s", joined)
	}
}

// loadLivingLines is the raw half of the corpus for direct-helper tests.
func loadLivingLines(t *testing.T, root string) []string {
	t.Helper()

	return LivingDocsLines(root, DefaultSettings().livingDocs())
}
