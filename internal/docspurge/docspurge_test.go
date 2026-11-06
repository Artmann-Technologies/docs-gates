package docspurge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// fixture TODO content: one struck row (the docs-strike output shape)
// beside one open row, plus the CHANGELOG skeleton the entry inserts
// under.
const (
	fixtureTodo = `# TODO_LIST

| Task | Status | Priority |
| ---- | ------ | -------- |
| ~~Purge the zorphuser legacy gate~~ done at ` + "`abc1234`" + ` | ✅ ` + "`DONE`" + ` | P3 |
| Keep the open work moving | 🔴 ` + "`TODO`" + ` | P2 |
`
	fixtureChangelog = "# CHANGELOG\n\n## [Unreleased]\n\n### Changed\n\n- prior entry\n\n### Fixed\n\n- prior fix\n"
)

// writeFixtures materializes the fixture repo and returns its root.
func writeFixtures(t *testing.T) string {
	t.Helper()

	root := t.TempDir()

	if err := os.WriteFile(filepath.Join(root, "TODO_LIST.md"), []byte(fixtureTodo), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(root, "CHANGELOG.md"), []byte(fixtureChangelog), 0o644); err != nil {
		t.Fatal(err)
	}

	return root
}

func readFixtures(t *testing.T, root string) (string, string) {
	t.Helper()

	todo, err := os.ReadFile(filepath.Join(root, "TODO_LIST.md"))
	if err != nil {
		t.Fatal(err)
	}

	changelog, err := os.ReadFile(filepath.Join(root, "CHANGELOG.md"))
	if err != nil {
		t.Fatal(err)
	}

	return string(todo), string(changelog)
}

// TestRunDryRunPrintsPlanAndWritesNothing pins the default face: the
// old→new plan on stdout, both fixture files byte-identical.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunDryRunPrintsPlanAndWritesNothing(t *testing.T) {
	root := writeFixtures(t)
	beforeTodo, beforeChangelog := readFixtures(t, root)

	out := gatecli.CaptureOutput(func() int { return Run([]string{root}) })

	if out.Code != gatecli.ExitClean {
		t.Fatalf("Run = %d, want %d; stderr: %s", out.Code, gatecli.ExitClean, out.Stderr)
	}

	if !strings.Contains(out.Stdout, "--- TODO_LIST.md:5") || !strings.Contains(out.Stdout, "zorphuser legacy gate") {
		t.Fatalf("dry-run plan missing the struck-row pair:\n%s", out.Stdout)
	}

	if !strings.Contains(out.Stdout, "DRY-RUN") {
		t.Fatalf("dry-run notice missing:\n%s", out.Stdout)
	}

	afterTodo, afterChangelog := readFixtures(t, root)
	if afterTodo != beforeTodo || afterChangelog != beforeChangelog {
		t.Fatal("dry-run mutated the fixture files")
	}
}

// TestRunApplyRemovesRowsAndInsertsBullets pins the write face: the
// struck row leaves the TODO file, the bullet lands inside the
// [Unreleased] Changed section before ### Fixed, the open row survives.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunApplyRemovesRowsAndInsertsBullets(t *testing.T) {
	root := writeFixtures(t)

	out := gatecli.CaptureOutput(func() int { return Run([]string{root, "--apply"}) })

	if out.Code != gatecli.ExitClean {
		t.Fatalf("Run --apply = %d, want %d; stderr: %s", out.Code, gatecli.ExitClean, out.Stderr)
	}

	todo, changelog := readFixtures(t, root)

	if strings.Contains(todo, "~~") || strings.Contains(todo, "zorphuser legacy gate") {
		t.Fatalf("struck row survived the apply:\n%s", todo)
	}

	if !strings.Contains(todo, "Keep the open work moving") {
		t.Fatalf("open row was collateral damage:\n%s", todo)
	}

	changedIdx := strings.Index(changelog, "### Changed")
	fixedIdx := strings.Index(changelog, "### Fixed")
	bulletIdx := strings.Index(changelog, "- TODO_LIST purge (")

	if changedIdx < 0 || fixedIdx < 0 || bulletIdx < 0 || bulletIdx < changedIdx || bulletIdx > fixedIdx {
		t.Fatalf("purge bullet landed outside the [Unreleased] Changed section:\n%s", changelog)
	}

	if !strings.Contains(changelog, "done at `abc1234`") {
		t.Fatalf("verdict tail lost from the bullet:\n%s", changelog)
	}

	if !strings.Contains(out.Stdout, "purged 1 struck row(s)") {
		t.Fatalf("apply summary missing:\n%s", out.Stdout)
	}
}

// TestRunNoStruckRowsIsCleanNoop pins the empty board face: exit 0 with
// an explicit nothing-to-purge line, files untouched.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunNoStruckRowsIsCleanNoop(t *testing.T) {
	root := writeFixtures(t)

	unstruck := strings.ReplaceAll(
		fixtureTodo,
		"~~Purge the zorphuser legacy gate~~ done at `abc1234`",
		"Purge the zorphuser legacy gate (open)",
	)
	if err := os.WriteFile(filepath.Join(root, "TODO_LIST.md"), []byte(unstruck), 0o644); err != nil {
		t.Fatal(err)
	}

	beforeTodo, beforeChangelog := readFixtures(t, root)

	out := gatecli.CaptureOutput(func() int { return Run([]string{root, "--apply"}) })

	if out.Code != gatecli.ExitClean {
		t.Fatalf("Run = %d, want %d; stderr: %s", out.Code, gatecli.ExitClean, out.Stderr)
	}

	if !strings.Contains(out.Stdout, "no struck rows") {
		t.Fatalf("no-op notice missing:\n%s", out.Stdout)
	}

	afterTodo, afterChangelog := readFixtures(t, root)
	if afterTodo != beforeTodo || afterChangelog != beforeChangelog {
		t.Fatal("the no-op mutated the fixture files")
	}
}

// TestRunPreFlightRefusalWritesNothing pins the gate leg through its
// deterministic refusal shape: a bullet already on disk makes PreFlight
// exit 2, and the purge surfaces that code without touching either file
// (both dry-run and apply).
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunPreFlightRefusalWritesNothing(t *testing.T) {
	root := writeFixtures(t)

	poisoned := strings.Replace(
		fixtureChangelog,
		"- prior entry",
		fmt.Sprintf(
			"- prior entry\n- TODO_LIST purge (%s): Purge the zorphuser legacy gate — done at `abc1234`",
			time.Now().Format("2006-01-02"),
		),
		1,
	)
	if err := os.WriteFile(filepath.Join(root, "CHANGELOG.md"), []byte(poisoned), 0o644); err != nil {
		t.Fatal(err)
	}

	beforeTodo, beforeChangelog := readFixtures(t, root)

	out := gatecli.CaptureOutput(func() int { return Run([]string{root, "--apply"}) })

	if out.Code != gatecli.ExitUsage {
		t.Fatalf("Run over an on-disk bullet = %d, want %d (PreFlight's on-disk refusal); stderr: %s",
			out.Code, gatecli.ExitUsage, out.Stderr)
	}

	if !strings.Contains(out.Stderr, "ALREADY on disk") {
		t.Fatalf("refusal did not name the on-disk cause:\n%s", out.Stderr)
	}

	afterTodo, afterChangelog := readFixtures(t, root)
	if afterTodo != beforeTodo || afterChangelog != beforeChangelog {
		t.Fatal("the refused purge mutated the fixture files")
	}
}

// TestRunUnsafePreFlightRefusesWithFindings pins the exit-1 leg end to
// end: the report item's only living-present distinctive token is
// zorphuser at DF 7 (the struck row + 6 filler lines); the drafted
// bullet pushes it to DF 8 — the GenericDF bar — and the pre-flight
// answers UNSAFE while --apply writes nothing.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunUnsafePreFlightRefusesWithFindings(t *testing.T) {
	root := writeFixtures(t)

	var filler strings.Builder

	filler.WriteString(fixtureChangelog)

	for range 8 {
		filler.WriteString("- audit ledger rows land weekly\n")
	}

	for range 6 {
		filler.WriteString("- the zorphuser archive keeps its evidence alive\n")
	}

	if err := os.WriteFile(filepath.Join(root, "CHANGELOG.md"), []byte(filler.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	reportDir := filepath.Join(root, "docs", "status")
	if err := os.MkdirAll(reportDir, 0o755); err != nil {
		t.Fatal(err)
	}

	report := "# R\n\n## f) Next\n\n1. audit the zorphuser ledger weekly\n"
	if err := os.WriteFile(filepath.Join(reportDir, "2026-01-01_00-00_r.md"), []byte(report), 0o644); err != nil {
		t.Fatal(err)
	}

	beforeTodo, _ := readFixtures(t, root)

	out := gatecli.CaptureOutput(func() int { return Run([]string{root, "--apply"}) })

	if out.Code != gatecli.ExitFindings {
		t.Fatalf("Run over a flip-bound entry = %d, want %d; stdout:\n%s\nstderr:\n%s",
			out.Code, gatecli.ExitFindings, out.Stdout, out.Stderr)
	}

	if !strings.Contains(out.Stderr, "UNSAFE") {
		t.Fatalf("the UNSAFE verdict is missing:\n%s", out.Stderr)
	}

	afterTodo, _ := readFixtures(t, root)
	if afterTodo != beforeTodo {
		t.Fatal("the refused purge mutated the TODO file")
	}
}

// TestRunMissingChangelogAnchorsIsUsageError pins the anchor refusal:
// a CHANGELOG without [Unreleased]/### Changed refuses with exit 2 and
// leaves the TODO file intact.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunMissingChangelogAnchorsIsUsageError(t *testing.T) {
	root := writeFixtures(t)

	if err := os.WriteFile(
		filepath.Join(root, "CHANGELOG.md"),
		[]byte("# CHANGELOG\n\n## [v0.1.0]\n\n### Changed\n\n- x\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	beforeTodo, _ := readFixtures(t, root)

	out := gatecli.CaptureOutput(func() int { return Run([]string{root, "--apply"}) })

	if out.Code != gatecli.ExitUsage {
		t.Fatalf("Run without anchors = %d, want %d; stderr: %s", out.Code, gatecli.ExitUsage, out.Stderr)
	}

	if !strings.Contains(out.Stderr, "[Unreleased]") {
		t.Fatalf("the anchor refusal did not name the missing anchor:\n%s", out.Stderr)
	}

	afterTodo, _ := readFixtures(t, root)
	if afterTodo != beforeTodo {
		t.Fatal("the refused purge mutated the TODO file")
	}
}

// TestParseArgsKnowsTheContract covers the parse face: unknown flags,
// value flags needing a value, --apply, and root resolution.
func TestParseArgsKnowsTheContract(t *testing.T) {
	t.Parallel()

	if _, err := ParseArgs([]string{".", "--bogus"}); err == nil {
		t.Fatal("ParseArgs accepts an unknown flag")
	}

	if _, err := ParseArgs([]string{".", "--file"}); err == nil {
		t.Fatal("ParseArgs accepts --file without a value")
	}

	opts, err := ParseArgs([]string{"/tmp", "--file", "BOARD.md", "--changelog", "NEWS.md", "--apply"})
	if err != nil {
		t.Fatalf("ParseArgs rejects the documented set: %v", err)
	}

	if !opts.Apply || opts.File != "BOARD.md" || opts.Changelog != "NEWS.md" || opts.Root != "/tmp" {
		t.Fatalf("options misparsed: %+v", opts)
	}
}
