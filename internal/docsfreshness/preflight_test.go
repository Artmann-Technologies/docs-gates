package docsfreshness

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// --pre-flight's delta contract: the exact burn class of 2026-10-07
// (five boundary flips in one session, all from the author's own fresh
// lines) must be caught BEFORE the line lands, not by the next ritual
// run. The scenario pins the mechanics: an item whose only distinctive
// token sits at DF 7, plus a candidate line that pushes it to 8.
func preFlightRoot(t *testing.T) string {
	t.Helper()

	root := t.TempDir()

	status := filepath.Join(root, "docs", "status")
	if err := os.MkdirAll(status, 0o755); err != nil {
		t.Fatal(err)
	}

	report := "# Report\n\n## f) Next\n\n1. webphone alpha-probe calibration recheck\n"
	if err := os.WriteFile(filepath.Join(status, "2026-09-01_00-00_BASE.md"), []byte(report), 0o644); err != nil {
		t.Fatal(err)
	}

	// "webphone" rides 8 lines (generic by definition), "alpha-probe"
	// rides 7 (the item's only distinctive token).
	todo := make([]string, 0, 8)
	for i := range 8 {
		todo = append(todo, fmt.Sprintf("- webphone note number %d", i))
	}

	for i := range 7 {
		todo = append(todo, fmt.Sprintf("- alpha-probe evidence number %d", i))
	}

	if err := os.WriteFile(
		filepath.Join(root, "TODO_LIST.md"),
		[]byte("# TODO\n\n"+strings.Join(todo, "\n")+"\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	return root
}

func TestPreFlightFlagsWouldBeDrift(t *testing.T) {
	t.Parallel()

	root := preFlightRoot(t)

	if rc := PreFlight(
		root,
		[]string{"changelog: alpha-probe summary row"},
		nil,
		false,
		false,
		DefaultSettings(),
	); rc != gatecli.ExitFindings {
		t.Fatalf("a line genericizing the last distinctive token must exit findings, got %d", rc)
	}

	unrelated := []string{"changelog: a wholly unrelated qwertyzz row"}
	if rc := PreFlight(root, unrelated, nil, false, false, DefaultSettings()); rc != gatecli.ExitClean {
		t.Fatalf("an unrelated line must exit clean, got %d", rc)
	}
}

// The culprit attribution (TODO preflight-culprit, 2026-10-16): a
// WOULD-DRIFT must name the candidate LINE carrying the generic token —
// the 2026-10-13 filename chase burned two guess rounds because the
// boundary-token report names tokens, never lines, and a hyphenated
// spelling matches without moving the raw DF count at all.
//
// Deliberately NOT parallel: the output-capturing tests swap the global
// os.Stdout, which would race every other parallel test's stdout writes.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestPreFlightNamesCulpritCandidateToken(t *testing.T) {
	root := preFlightRoot(t)

	captured := gatecli.CaptureOutput(func() int {
		return PreFlight(root, []string{"changelog: alpha-probe summary row"}, nil, false, false, DefaultSettings())
	})

	if captured.Code != gatecli.ExitFindings {
		t.Fatalf("the genericizing line must exit findings, got %d", captured.Code)
	}

	got := captured.Stdout

	if !strings.Contains(got, `culprit: candidate #1 token "alpha-probe"`) {
		t.Fatalf("the WOULD-DRIFT must attribute the carrying candidate line, got output:\n%s", got)
	}
}

// The silent class the attribution exists for: a hyphen/space spelling of
// the token matches via the squash key WITHOUT raising the raw DF count —
// the boundary report never fires for it, so the WOULD-DRIFT must mark it
// explicitly.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestPreFlightMarksSquashSpellingCulprit(t *testing.T) {
	root := preFlightRoot(t)

	captured := gatecli.CaptureOutput(func() int {
		return PreFlight(
			root,
			[]string{"changelog: the alpha probe summary row", "changelog: alpha-probe landing row"},
			nil,
			false,
			false,
			DefaultSettings(),
		)
	})

	if captured.Code != gatecli.ExitFindings {
		t.Fatalf("the batch must exit findings, got %d", captured.Code)
	}

	got := captured.Stdout

	if !strings.Contains(got, `"alpha-probe" (hyphen/space spelling`) {
		t.Fatalf("the squash-only spelling must be marked as the silent culprit, got output:\n%s", got)
	}
}

func TestBoundaryCrossingsReportsOnlyThresholdFlips(t *testing.T) {
	t.Parallel()

	base := make([]string, 0, 15)
	for i := range 8 {
		base = append(base, fmt.Sprintf("webphone line %d", i))
	}

	for i := range 7 {
		base = append(base, fmt.Sprintf("alpha-probe line %d", i))
	}

	mod := append(slices.Clone(base), "alpha-probe mention")

	crossings := BoundaryCrossings([]string{"alpha-probe mention"}, base, mod, DefaultSettings().df())
	if len(crossings) != 1 || crossings[0].Token != "alpha-probe" || crossings[0].Before != 7 ||
		crossings[0].After != 8 {
		t.Fatalf("want exactly alpha-probe 7→8, got %+v", crossings)
	}

	// Already-generic tokens do not cross (8→9 stays above the bar) and
	// fresh tokens do not either (0→1 stays below it).
	crossings = BoundaryCrossings([]string{"webphone again and brandnew-zz"}, base, base, DefaultSettings().df())
	if len(crossings) != 0 {
		t.Fatalf("no crossing expected for already-generic or fresh tokens, got %+v", crossings)
	}
}

func TestPreFlightEmptyCandidateIsUsage(t *testing.T) {
	t.Parallel()

	root := preFlightRoot(t)

	if rc := PreFlight(root, nil, nil, false, false, DefaultSettings()); rc != gatecli.ExitUsage {
		t.Fatalf("empty candidate must exit usage, got %d", rc)
	}
}

// The 2026-10-10 report feared pre-flight skips docs/status/archived/
// while --audit-all replays it (m144). The corpus lister has included
// archived reports since 2026-09-30 (63af9cf), so this pins the
// behavior end-to-end: a candidate line genericizing the last
// distinctive token of an ARCHIVED report's §f item must flip the
// verdict to findings — the archived corpus is inside the delta.
func TestPreFlightReplaysArchivedCorpus(t *testing.T) {
	t.Parallel()

	root := preFlightRoot(t)

	status := filepath.Join(root, "docs", "status", "archived")
	if err := os.MkdirAll(status, 0o755); err != nil {
		t.Fatal(err)
	}

	report := "# Report\n\n## f) Next\n\n1. webphone alpha-probe calibration recheck\n"
	if err := os.WriteFile(
		filepath.Join(status, "2026-08-01_00-00_ARCHIVED.md"),
		[]byte(report),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	if rc := PreFlight(
		root,
		[]string{"changelog: alpha-probe summary row"},
		nil,
		false,
		false,
		DefaultSettings(),
	); rc != gatecli.ExitFindings {
		t.Fatalf("an archived item's last distinctive token must still flip the verdict, got %d", rc)
	}

	unrelated := []string{"changelog: a wholly unrelated qwertyzz row"}
	if rc := PreFlight(root, unrelated, nil, false, false, DefaultSettings()); rc != gatecli.ExitClean {
		t.Fatalf("an unrelated line must exit clean, got %d", rc)
	}
}
