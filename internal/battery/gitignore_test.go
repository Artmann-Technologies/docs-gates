package battery

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// minCorpusFiles is the floor under which the walk below would be a
// vacuous pass (the corpus currently carries 19 files; if the fixtures
// themselves got swallowed, this arm must fail loudly, not green on an
// empty walk).
const minCorpusFiles = 10

// TestCorpusFixturesStayTracked is the regression arm for the
// `.gitignore` `reports/` trap (paid for 2026-11-04): BuildFlow's managed
// ignore block carries a broad `reports/` rule, and corpus fixtures under
// `testdata/parity/*/reports/` silently matched it — the files existed in
// the working tree, tests passed locally, but `git add testdata/...`
// staged nothing and CI failed with golden drift only. Scoped negations
// outside the managed block fixed it once; this arm makes the trap
// mechanically impossible again whenever this repo is a git checkout
// (which CI always is — the arm rides the existing `go test` job, no
// separate CI step). Two legs:
//
//  1. TRACKED — every file under testdata/ is in the index (`git
//     ls-files`), the exact property the trap broke (present but never
//     staged).
//  2. UNIGNORED — no testdata path matches an ignore rule (`git
//     check-ignore --no-index`, because check-ignore refuses to report
//     tracked files at all — the red control proved the plain form blind
//     here), so a NEW fixture cannot silently fall into the swallow.
//
// Not parallel: exec's git over the shared working tree.
//
//nolint:paralleltest // git working-tree probe, serialized on purpose
func TestCorpusFixturesStayTracked(t *testing.T) {
	topOut, err := exec.CommandContext(t.Context(), "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Skipf("not inside a git checkout — nothing to assert (%v)", err)
	}

	toplevel := strings.TrimSpace(string(topOut))

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	if toplevel != repoRoot {
		t.Skipf("testdata belongs to another checkout (%s)", toplevel)
	}

	var files []string

	corpus := filepath.Join(repoRoot, "testdata")

	walkErr := filepath.WalkDir(corpus, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			files = append(files, path)
		}

		return nil
	})
	if walkErr != nil {
		t.Fatal(walkErr)
	}

	if len(files) < minCorpusFiles {
		t.Fatalf("corpus walk found only %d file(s) (< %d) — the walk itself is broken", len(files), minCorpusFiles)
	}

	sort.Strings(files)

	trackedOut, err := exec.CommandContext(t.Context(), "git", "ls-files", "--", corpus).Output()
	if err != nil {
		t.Fatalf("git ls-files failed: %v", err)
	}

	tracked := map[string]bool{}

	for line := range strings.SplitSeq(strings.TrimSpace(string(trackedOut)), "\n") {
		if line == "" {
			continue
		}

		// ls-files prints paths relative to the INVOKING cwd (the test
		// binary's package dir inside the repo), not repo-relative —
		// resolve against cwd or the join silently leaves the repo.
		abs, absErr := filepath.Abs(line)
		if absErr != nil {
			t.Fatal(absErr)
		}

		tracked[abs] = true
	}

	var untracked []string

	for _, f := range files {
		if !tracked[f] {
			untracked = append(untracked, f)
		}
	}

	if len(untracked) > 0 {
		t.Fatalf(
			"gitignore trap: %d corpus fixture(s) exist on disk but are NOT tracked (the 2026-11-04 failure shape — "+
				"`git add` stages nothing and CI sees golden drift only):\n  %s",
			len(untracked),
			strings.Join(untracked, "\n  "),
		)
	}

	// check-ignore exit codes: 0 = at least one path is ignored (the
	// trap), 1 = none ignored (clean), anything else = git broke.
	// --no-index bypasses the tracked-files blind spot; go1.26+
	// compile-rejects mixing literals with a slice spread in one variadic
	// call, so the argv is built with append first.
	args := append([]string{"check-ignore", "--no-index", "--"}, files...)

	out, err := exec.CommandContext(t.Context(), "git", args...).CombinedOutput()

	listed := strings.TrimSpace(string(out))

	if err == nil {
		t.Fatalf("gitignore trap: %d corpus fixture(s) are IGNORED and any re-add would stage nothing:\n%s",
			strings.Count(listed, "\n")+1, string(out))
	}

	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok && exitErr.ExitCode() == 1 {
		return // clean: check-ignore matched nothing
	}

	t.Fatalf("git check-ignore failed: %v\n%s", err, string(out))
}
