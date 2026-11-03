package docsannotate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
)

// Rule F's corruption signature: a letter or digit touching a backtick.
// The *.nix widening (2026-10-04) and the non-git recursive walk
// (2026-10-05, replacing one-level-deep `**` globs) are pinned here.
func TestSpacingReMatchesAdjacency(t *testing.T) {
	t.Parallel()

	for _, bad := range []string{
		"a`b", "`b`a", "1`x", "x`2", "see`flake.nix`here",
	} {
		if !spacingRe.MatchString(bad) {
			t.Errorf("spacingRe should match %q", bad)
		}
	}

	for _, good := range []string{"a `b` c", "nix fmt", "backtick", "a` `b"} {
		if spacingRe.MatchString(good) {
			t.Errorf("spacingRe should NOT match %q", good)
		}
	}
}

// The non-git (store-copy) branch must reach DEEP .nix files — the old
// `filepath.Glob("**/*.nix")` matched only one directory level and missed
// hosts/pbx/*.nix entirely.
func TestCheckSpacingNonGitWalkCoversDeepNixFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	files := map[string]string{
		filepath.Join(root, "flake.nix"):                     "# clean comment, no adjacency\n",
		filepath.Join(root, "hosts", "pbx", "telephony.nix"): "# wiring lives in`flake/nixos.nix`(post-split)\n",
		filepath.Join(root, "docs", "status", "x.md"):        "fine `AGENTS.md` spacing\nbroken`here`now\n",
		filepath.Join(root, "scripts", "clean.sh"):           "# bad`adjacency`but `# format:allow` opts out\n",
	}

	for path, content := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	findings := CheckSpacing(root, gitRunner{root: root, missing: true}, false, config.Defaults().RuleFExtensions)

	want := map[string]int{
		"hosts/pbx/telephony.nix": 1,
		"docs/status/x.md":        1,
		"scripts/clean.sh":        0,
		"flake.nix":               0,
	}

	got := map[string]int{}

	for _, f := range findings {
		got[f.Path]++
	}

	for path, n := range want {
		if got[path] != n {
			t.Errorf("%s: got %d findings, want %d (all findings: %v)", path, got[path], n, findings)
		}
	}
}

// F56 (TODO §16 row): a failing walk must surface as a finding, never a
// silent clean. The old closure answered nil for every walk error, and
// filepath.WalkDir only reports errors the closure RETURNS — so the
// F/walk branch was unreachable dead code. A nonexistent root is the
// deterministic trigger (works under any user, no chmod tricks).
func TestDocsByWalkFailureIsAFinding(t *testing.T) {
	t.Parallel()

	findings, docs := docsByWalk(filepath.Join(t.TempDir(), "does-not-exist"), config.Defaults().RuleFExtensions)

	if len(findings) != 1 {
		t.Fatalf("walk failure must produce exactly one F/walk finding, got %+v", findings)
	}

	f := findings[0]

	if f.Rule != "F/walk" || f.Detail == "" {
		t.Fatalf("unexpected finding shape: %+v", f)
	}

	if f.Text == "" {
		t.Error("F/walk must carry remedy Text so --verbose renders a line, not a bare indent")
	}

	if len(docs) != 0 {
		t.Errorf("failed walk must yield no docs, got %v", docs)
	}
}

// End-to-end leg: CheckSpacing (non-git branch) must propagate the walk
// failure into its findings — the gate verdict has to be FAIL, not CLEAN.
func TestCheckSpacingWalkFailureSurfaces(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "gone")

	for _, f := range CheckSpacing(root, gitRunner{root: root, missing: true}, false, config.Defaults().RuleFExtensions) {
		if f.Rule == "F/walk" {
			return
		}
	}

	t.Fatal("CheckSpacing must surface the walk failure (silent clean regression)")
}

// Rule G (TODO §44, 2026-10-13): the 2026-09-30-era ID-cell-only strikes
// (a strike opening in the ID cell and closing inside the task cell)
// passed every gate — rendered as a strike over the ID column. Cell
// parity is the detector; per-cell wraps are the uniform shape.
func TestRuleGCrossCellStrikeFails(t *testing.T) {
	t.Parallel()

	bad := "| ~~C18   | Assistant          | Task text (§3.5)~~ done (evidence) | 30 | rest | — |"
	good := "| ~~C18~~ | ~~Assistant~~ | ~~Task text (§3.5)~~ done (evidence) | 30 | rest | — |"

	if f := crossCellFindings("x.md", []string{bad}); len(f) != 1 || f[0].Rule != "G/uniform" {
		t.Fatalf("cross-cell strike must yield one G/uniform finding, got %+v", f)
	}

	if f := crossCellFindings("x.md", []string{good}); len(f) != 0 {
		t.Fatalf("per-cell strike must stay clean, got %+v", f)
	}

	// An escaped ~~ inside inline code is a literal, not a strike.
	if f := crossCellFindings("x.md", []string{"| grep `~~` marker | cell two |"}); len(f) != 0 {
		t.Fatalf("inline-code ~~ literal must not count, got %+v", f)
	}
}
