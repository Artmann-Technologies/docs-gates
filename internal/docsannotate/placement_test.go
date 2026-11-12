package docsannotate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// placementFixture materializes one report file under the DEFAULT
// annotate corpus (config-less repo: DefaultSettings' report dir) and
// returns the repo root.
func placementFixture(t *testing.T, content string) string {
	t.Helper()

	root := t.TempDir()

	dir := filepath.Join(root, DefaultSettings().dir())

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "2026-11-12_10-00_fixture.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	return root
}

// runPlacement runs the gate over root with stdout/stderr captured.
func runPlacement(root string) gatecli.CapturedOutput {
	return gatecli.CaptureOutput(func() int { return CheckVerdictPlacement(root) })
}

func TestCheckVerdictPlacementClean(t *testing.T) {
	t.Parallel()

	root := placementFixture(t, strings.Join([]string{
		"# fixture",
		"",
		"3. ~~the stale claim~~ done (landed in the tree; see the CHANGELOG entry)",
		"- ~~the old lane~~ resolved by the follow-up train",
		"4. ~~wait on upstream~~ superseded by the shared engine",
		"",
	}, "\n"))

	out := runPlacement(root)

	if out.Code != gatecli.ExitClean {
		t.Fatalf("code = %d, want ExitClean (stdout: %s / stderr: %s)", out.Code, out.Stdout, out.Stderr)
	}

	if !strings.Contains(out.Stdout, "verdict placement CLEAN") {
		t.Errorf("CLEAN verdict line missing from stdout:\n%s", out.Stdout)
	}
}

func TestCheckVerdictPlacementVerdictInsideFires(t *testing.T) {
	t.Parallel()

	root := placementFixture(t, strings.Join([]string{
		"# fixture",
		"",
		"7. ~~resolved upstream in the 2026-10-01 release~~",
		"9. ~~the duplicate lane~~ superseded by the shared engine",
		"",
	}, "\n"))

	out := runPlacement(root)

	if out.Code != gatecli.ExitFindings {
		t.Fatalf("code = %d, want ExitFindings (stdout: %s)", out.Code, out.Stdout)
	}

	if strings.Count(out.Stdout, "P/verdict-inside") != 1 {
		t.Errorf("want exactly one P/verdict-inside finding:\n%s", out.Stdout)
	}

	if !strings.Contains(out.Stdout, "move it OUTSIDE the closing ~~") {
		t.Errorf("remedy text missing from the finding:\n%s", out.Stdout)
	}

	if !strings.Contains(out.Stdout, "fixture.md:3:") {
		t.Errorf("finding must cite line 3 (the trapped-verdict item):\n%s", out.Stdout)
	}
}

func TestCheckVerdictPlacementNumberInsideFires(t *testing.T) {
	t.Parallel()

	root := placementFixture(t, strings.Join([]string{
		"# fixture",
		"",
		"12. ~~first claim~~ done (row landed)",
		"",
		"~~14. the old text~~ done (the fix train)",
		"",
		"18. ~~fourth~~ done (x)",
		"~~19. fifth~~ done (y)",
		"",
	}, "\n"))

	out := runPlacement(root)

	if out.Code != gatecli.ExitFindings {
		t.Fatalf("code = %d, want ExitFindings (stdout: %s)", out.Code, out.Stdout)
	}

	// The stray shape (blank-line separated, line 5) AND the
	// continuation-join shape (no blank line, line 8 — ExtractItems
	// folds it into item 18's block, the line scan still catches it).
	for _, want := range []string{"fixture.md:5:", "fixture.md:8:"} {
		if !strings.Contains(out.Stdout, want) {
			t.Errorf("number-inside finding for %s missing:\n%s", want, out.Stdout)
		}
	}

	if strings.Count(out.Stdout, "P/number-inside") != 2 {
		t.Errorf("want exactly two P/number-inside findings:\n%s", out.Stdout)
	}
}

// The exempt shapes: a verdict word INSIDE a table row's strike is fine
// while its status cell carries the marker; table rows strike number
// cells legitimately (`~~3~~`); checkbox rows, prose quoting a number,
// indented continuations, inline code, and fenced content never match.
func TestCheckVerdictPlacementExemptionsStayClean(t *testing.T) {
	t.Parallel()

	root := placementFixture(t, strings.Join([]string{
		"# fixture",
		"",
		"| 3 | ~~old lane~~ | ✅ `DONE` | high | low | e |",
		"| 4 | ~~second lane~~ | ✅ `DONE` | low | low | e |",
		"",
		"- [x] ~~the checkbox lane~~ done (shipped)",
		"the ~~old 12. claim~~ stays as written",
		"  ~~13. indented continuation~~ done (x)",
		"",
		"`~~15. inline~~ code` never renders as an item",
		"",
		"```text",
		"~~16. fenced~~ fixed inside a fence",
		"```",
		"",
		"5. ~~real item~~ done (fine)",
		"",
	}, "\n"))

	out := runPlacement(root)

	if out.Code != gatecli.ExitClean {
		t.Fatalf("code = %d, want ExitClean (stdout: %s)", out.Code, out.Stdout)
	}

	if !strings.Contains(out.Stdout, "verdict placement CLEAN") {
		t.Errorf("CLEAN verdict line missing from stdout:\n%s", out.Stdout)
	}
}

// A malformed .docs-gates.json must FATAL with exit 2, never a verdict.
func TestCheckVerdictPlacementMalformedConfigFATALs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	if err := os.WriteFile(filepath.Join(root, ".docs-gates.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := runPlacement(root)

	if out.Code != gatecli.ExitUsage {
		t.Fatalf("code = %d, want ExitUsage (stdout: %s / stderr: %s)", out.Code, out.Stdout, out.Stderr)
	}

	if !strings.Contains(out.Stderr, "FATAL:") {
		t.Errorf("FATAL line missing from stderr:\n%s", out.Stderr)
	}
}
