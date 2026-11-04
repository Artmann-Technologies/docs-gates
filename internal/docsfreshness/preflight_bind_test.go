package docsfreshness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// bindFixture builds a root whose ONE report item fails in base (its
// tokens appear nowhere in the living docs) so the pre-flight bind leg
// has a real unharvested item to harvest or miss.
func bindFixture(t *testing.T, item string) string {
	t.Helper()

	root := t.TempDir()

	status := filepath.Join(root, "docs", "status")
	if err := os.MkdirAll(status, 0o755); err != nil {
		t.Fatal(err)
	}

	report := "# Report\n\n## f) Next\n\n1. " + item + "\n"
	if err := os.WriteFile(filepath.Join(status, "2026-09-02_00-00_BIND.md"), []byte(report), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(root, "TODO_LIST.md"),
		[]byte("# TODO\n\n- unrelated filler row for the living docs\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	return root
}

// The bind leg (TODO §77 preflight-bind-check): a candidate row that
// harvests the open item must SAY so, attributed to the exact line.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestPreFlightReportsBoundItemWithCandidate(t *testing.T) {
	root := bindFixture(t, "quota-xyz replenish window followup")

	captured := gatecli.CaptureOutput(func() int {
		return PreFlight(
			root,
			[]string{"todo: quota-xyz replenish window owner row"},
			nil,
			false,
			false,
			DefaultSettings(),
		)
	})

	if captured.Code != gatecli.ExitClean {
		t.Fatalf("the harvesting line must exit clean, got %d (output:\n%s)", captured.Code, captured.Stdout)
	}

	got := captured.Stdout

	if !strings.Contains(got, "binds: [2026-09-02_00-00_BIND.md]") ||
		!strings.Contains(got, "quota-xyz replenish window followup") {
		t.Fatalf("the newly-harvested item must be listed with its report, got:\n%s", got)
	}

	if !strings.Contains(got, "via candidate: todo: quota-xyz replenish window owner row") {
		t.Fatalf("the binding must attribute the candidate line, got:\n%s", got)
	}
}

// The 2026-10-18 miss class made visible BEFORE the write: the item's
// distinctive token is the slash-joined "metric/counter"; a row that
// spells it "metric and counter" binds NOTHING, and the bind leg must
// say so instead of letting a SAFE verdict imply a harvest.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestPreFlightBindsNothingWhenSlashTokenIsSplit(t *testing.T) {
	// The item's only harvestable tokens are the slash-joined
	// "metric/counter" plus a pair of fresh words the candidate must NOT
	// carry — otherwise THOSE bind and the slash miss hides again (the
	// first draft of this test failed exactly that way).
	root := bindFixture(t, "the metric/counter qvadrant alarm needs a drain row")

	captured := gatecli.CaptureOutput(func() int {
		return PreFlight(
			root,
			[]string{"todo: one unrelated row spelling metric and counter separately"},
			nil,
			false,
			false,
			DefaultSettings(),
		)
	})

	if captured.Code != gatecli.ExitClean {
		t.Fatalf("the non-binding line must still exit clean, got %d (output:\n%s)",
			captured.Code, captured.Stdout)
	}

	got := captured.Stdout

	if strings.Contains(got, "via candidate:") {
		t.Fatalf("a split spelling of a slash-joined token must not bind, got:\n%s", got)
	}

	if !strings.Contains(got, "binds: no open report item is harvested") {
		t.Fatalf("the empty-bind summary must be present, got:\n%s", got)
	}
}

func TestPreFlightRefusesCandidatesAlreadyOnDisk(t *testing.T) {
	t.Parallel()

	root := preFlightRoot(t)

	// A candidate that is already a TODO_LIST row: a post-write
	// pre-flight carries its own additions inside the baseline, so the
	// delta cannot see a flip — refuse instead of printing a
	// misleading SAFE (the 2026-10-10 burn).
	if rc := PreFlight(
		root,
		[]string{"- alpha-probe evidence number 3"},
		nil,
		false,
		false,
		DefaultSettings(),
	); rc != gatecli.ExitUsage {
		t.Fatalf("a candidate already on disk must exit usage, got %d", rc)
	}

	// Whitespace-only candidates are vacuous, not "on disk".
	if rc := PreFlight(root, []string{"   "}, nil, false, false, DefaultSettings()); rc != gatecli.ExitClean {
		t.Fatalf("a whitespace-only candidate must stay clean, got %d", rc)
	}
}

// The 2026-10-12 ledger-block class: a candidate line that is ALREADY on
// disk is normally refused (the post-write-run guard), which forced
// operator-side dedup (grep -Fxvf) — silently DROPPING the +1 DF push of
// the second copy that was about to land. --duplicates is the explicit
// "another copy is intentional" assertion: the copy stays modeled and its
// push must surface as findings when it genericizes an item.
func TestPreFlightDuplicatesModelsSecondCopyPush(t *testing.T) {
	t.Parallel()

	root := preFlightRoot(t)

	// Without the flag the second copy is still refused, not answered.
	if rc := PreFlight(
		root,
		[]string{"- alpha-probe evidence number 3"},
		nil,
		false,
		false,
		DefaultSettings(),
	); rc != gatecli.ExitUsage {
		t.Fatalf("second copy without --duplicates must stay refused, got %d", rc)
	}

	// With it, the +1 push genericizes the item's only distinctive token.
	if rc := PreFlight(
		root,
		[]string{"- alpha-probe evidence number 3"},
		nil,
		false,
		true,
		DefaultSettings(),
	); rc != gatecli.ExitFindings {
		t.Fatalf("second copy with --duplicates must model the DF push and exit findings, got %d", rc)
	}

	// A second copy of an already-generic line pushes nothing that matters.
	if rc := PreFlight(
		root,
		[]string{"- webphone note number 3"},
		nil,
		false,
		true,
		DefaultSettings(),
	); rc != gatecli.ExitClean {
		t.Fatalf("second copy of a generic line must stay clean, got %d", rc)
	}
}

// The on-disk refusal must NAME the offending line (the message contract
// the 2026-10-11 report asked for): an operator who cannot see WHICH
// candidate is already present cannot tell a stale re-run from a
// deliberate ledger second copy. Not parallel: CaptureOutput swaps the
// process-global streams.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestPreFlightRefusalNamesOffendingLine(t *testing.T) {
	root := preFlightRoot(t)

	captured := gatecli.CaptureOutput(func() int {
		return PreFlight(root, []string{"- alpha-probe evidence number 3"}, nil, false, false, DefaultSettings())
	})

	if captured.Code != gatecli.ExitUsage {
		t.Fatalf("on-disk candidate must exit usage, got %d", captured.Code)
	}

	if !strings.Contains(captured.Stderr, "alpha-probe evidence number 3") {
		t.Fatalf("refusal must name the offending line in its Present list, stderr:\n%s", captured.Stderr)
	}
}
