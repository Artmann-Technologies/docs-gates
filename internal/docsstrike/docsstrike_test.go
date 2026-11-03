package docsstrike

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sample = `# TODO List

| Task                            | Status        | Impact | Effort | Evidence  |
| ------------------------------- | ------------- | ------ | ------ | --------- |
| write the envelope audit now    | ` + "`TODO`" + `  | Med    | 30min  | plan F104 |
| pin the IPv6 range from panel   | ` + "`BLOCKED`" + ` | High   | 15min  | r3        |
| ~~old done row~~ done (earlier) | ` + "`DONE`" + `   | Low    | 5min   | log       |
`

func writeSample(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "TODO_LIST.md"), []byte(sample), 0o644); err != nil {
		t.Fatal(err)
	}

	return dir
}

func baseOpts() Options {
	return Options{
		Root:    ".",
		File:    "TODO_LIST.md",
		Row:     "envelope audit",
		Verdict: "2026-10-08: audit recorded in the plan",
		Status:  DefaultStatus,
	}
}

func TestPlanStrikeWrapsTaskFlipsStatusPreservesColumns(t *testing.T) {
	t.Parallel()

	strike, err := PlanStrike(sample, baseOpts())
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(strike.New, "| ~~write the envelope audit now") {
		t.Fatalf("task cell not wrapped: %q", strike.New)
	}

	if !strings.Contains(strike.New, "~~ done (2026-10-08: audit recorded in the plan)") {
		t.Fatalf("verdict clause missing: %q", strike.New)
	}

	if !strings.Contains(strike.New, "| ✅ `DONE` |") {
		t.Fatalf("status cell not flipped: %q", strike.New)
	}

	if !strings.Contains(strike.New, "| Med    | 30min  |") {
		t.Fatalf("Impact/Effort must be preserved byte-identically: %q", strike.New)
	}

	if strike.Old == strike.New {
		t.Fatal("strike must change the row")
	}
}

func TestPlanStrikeRefusesAmbiguousMissingStruckAndAlienShapes(t *testing.T) {
	t.Parallel()

	opts := baseOpts()

	opts.Row = "does-not-exist-anywhere"
	if _, err := PlanStrike(sample, opts); err == nil {
		t.Fatal("a selector matching zero rows must be an error")
	}

	opts = baseOpts()

	opts.Row = "the" // matches BOTH content rows ("the installer", "from the Hetzner")
	if _, err := PlanStrike(sample, opts); err == nil {
		t.Fatal("a selector matching multiple rows must be an error")
	}

	opts = baseOpts()

	opts.Row = "old done row"
	if _, err := PlanStrike(sample, opts); err == nil {
		t.Fatal("an already-struck row must be refused")
	}

	alien := "| ID | Task | Owner |\n| -- | ---- | ----- |\n| T1 | push | OWNER |\n"
	opts = baseOpts()

	opts.Row = "push"
	if _, err := PlanStrike(alien, opts); err == nil {
		t.Fatal("a 3-column row (not the TODO_LIST shape) must be refused")
	}
}

func TestRunDryRunLeavesFileIntactApplyWrites(t *testing.T) {
	t.Parallel()

	dir := writeSample(t)
	path := filepath.Join(dir, "TODO_LIST.md")

	rc := Run(
		[]string{dir, "--row", "envelope audit", "--verdict", "2026-10-08: audit done", "--evidence", "plan §T21"},
	)
	if rc != 0 {
		t.Fatalf("dry-run exit = %d, want 0", rc)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(after) != sample {
		t.Fatal("dry-run must not write the file")
	}

	rc = Run([]string{
		dir, "--row", "envelope audit", "--verdict", "2026-10-08: audit done",
		"--evidence", "plan T21", "--apply",
	})
	if rc != 0 {
		t.Fatalf("apply exit = %d, want 0", rc)
	}

	after, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	got := string(after)
	if !strings.Contains(got, "~~ done (2026-10-08: audit done)") {
		t.Fatalf("applied file misses the strike: %q", got)
	}

	if !strings.Contains(got, "| plan T21 |") {
		t.Fatalf("applied file misses the replaced evidence cell: %q", got)
	}

	if strings.Contains(got, "`TODO`  | Med") {
		t.Fatalf("applied file still carries the unstruck status: %q", got)
	}

	// A second apply of the same selector must refuse, not double-strike.
	rc = Run([]string{dir, "--row", "envelope audit", "--verdict", "again", "--apply"})
	if rc == 0 {
		t.Fatal("striking an already-struck row must fail")
	}
}

func TestParseArgsRequiresRowAndVerdictAndKnowsValueFlags(t *testing.T) {
	t.Parallel()

	if _, err := ParseArgs([]string{"--row", "x"}); err == nil {
		t.Fatal("--verdict must be required")
	}

	if _, err := ParseArgs([]string{"--verdict", "y"}); err == nil {
		t.Fatal("--row must be required")
	}

	opts, err := ParseArgs([]string{"--row", "sel", "--verdict", "v", "--file", "OTHER.md", "--apply"})
	if err != nil {
		t.Fatal(err)
	}

	if opts.Row != "sel" || opts.Verdict != "v" || opts.File != "OTHER.md" || !opts.Apply {
		t.Fatalf("value flags misparsed: %+v", opts)
	}

	if opts.Status != DefaultStatus {
		t.Fatalf("status must default to %q, got %q", DefaultStatus, opts.Status)
	}

	if _, err := ParseArgs([]string{"--row", "x", "--verdict", "v", "--nonsense"}); err == nil {
		t.Fatal("unknown flags must be rejected")
	}

	if _, err := ParseArgs([]string{"--row"}); err == nil {
		t.Fatal("a value flag without its value must be rejected")
	}
}

// The nesting guard (TODO strike-nesting-guard, the 2026-10-12
// double-wrapped wart): a --verdict that already OPENS with a verdict
// word must be refused at parse time — the tool prepends the done (…)
// wrap itself, so a prefixed verdict lands as done (done (…)).
func TestParseArgsRefusesAlreadyWrappedVerdict(t *testing.T) {
	t.Parallel()

	for _, verdict := range []string{
		"done (2026-10-14: landed)",
		"DONE at `abc1234`",
		"resolved: upstream fix",
		"superseded by the public flip",
		"decided 2026-10-12",
		"won't implement — reason",
		"`done (code-ticked)`",
		"**done (bold)**",
	} {
		if _, err := ParseArgs([]string{"--row", "envelope audit", "--verdict", verdict}); err == nil {
			t.Fatalf("verdict %q must be refused (the tool prepends the wrap itself)", verdict)
		} else if !strings.Contains(err.Error(), "verdict word") {
			t.Fatalf("refusal must name the verdict-word class, got: %v", err)
		}
	}

	for _, verdict := range []string{
		"2026-10-14: landed with the ledger block",
		"at `abc1234` the hash resolves",
		"rotation re-minted the same shape",
		"notarized in the ops journal",
	} {
		if _, err := ParseArgs([]string{"--row", "envelope audit", "--verdict", verdict}); err != nil {
			t.Fatalf("evidence verdict %q must be accepted, got: %v", verdict, err)
		}
	}
}
