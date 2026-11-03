package docsrouting

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/docsfreshness"
)

// fixtureTree builds a minimal repo tree: one unarchived report + one
// archived report (must be ignored) + the living docs.
func fixtureTree(t *testing.T, report, living string) string {
	t.Helper()

	root := t.TempDir()

	if err := os.MkdirAll(filepath.Join(root, "docs", "status", "archived"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(root, "docs", "status", "2026-10-06_x.md"),
		[]byte(report),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	archived := "## a) done section\n| 1 | ~~resolved historical row~~ | 🔴 `TODO` |\n"

	if err := os.WriteFile(
		filepath.Join(root, "docs", "status", "archived", "2026-01-01_old.md"),
		[]byte(archived),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(root, "TODO_LIST.md"), []byte(living), 0o644); err != nil {
		t.Fatal(err)
	}

	return root
}

// TestRoutingGaps pins the blind-spot class the lint exists for: an open
// actionable row OUTSIDE §f whose tokens appear nowhere in the living docs
// must be reported, while routed, struck, §f-internal, skip-marked and
// archived rows must not.
func TestRoutingGaps(t *testing.T) {
	t.Parallel()

	report := `# report

## a) next tasks

| 1 | rotate the zeptophone credential quarterly | 🔴 ` + "`TODO`" + ` |
| 2 | already tracked flerb maintenance work | 🔴 ` + "`TODO`" + ` |
| 3 | ~~struck resolved row about zeptophone~~ | 🔴 ` + "`TODO`" + ` |
| 4 | skip-marked zeptophone row | 🔴 ` + "`TODO`" + ` <!-- routing:skip --> |

## f) next steps

| 9 | zeptophone §f item is the harvester's domain | 🔴 ` + "`TODO`" + ` |

## b) open ends

- [ ] unticked zeptophone checklist item
`
	living := "## 9) whatever\n- flerb maintenance work is tracked here\n"

	root := fixtureTree(t, report, living)

	gaps := RoutingGaps(root, Settings{})

	wantRows := map[string]bool{
		"rotate the zeptophone credential": false,
		"unticked zeptophone":              false,
	}

	for _, gap := range gaps {
		for prefix := range wantRows {
			if strings.Contains(gap.Row, prefix) {
				wantRows[prefix] = true
			}
		}

		if strings.Contains(gap.Row, "already tracked") {
			t.Errorf("routed row must not be a gap: %s", gap.Row)
		}
	}

	for prefix, seen := range wantRows {
		if !seen {
			t.Errorf("expected row %q to be reported UNROUTED; gaps were %+v", prefix, gaps)
		}
	}

	if len(gaps) != len(wantRows) {
		t.Errorf("expected exactly %d gaps (struck/skip/§f/archived excluded), got %d: %+v",
			len(wantRows), len(gaps), gaps)
	}
}

// TestRoutedBar: a row sharing only GENERIC tokens with the living docs is
// still unrouted (accidental overlap is not a route — the freshness
// doctrine inverted). "backup" sits on >= GenericDF living lines, so the
// quorum matches but no token is distinctive. (NB "work"/"todo" are
// docsfreshness STOPWORDS — they never tokenize.)
func TestRoutedBar(t *testing.T) {
	t.Parallel()

	report := "## a) next tasks\n\n| 1 | the backup cleanup | 🔴 `TODO` |\n"

	var living strings.Builder
	living.WriteString("## 1) section\n")

	for range docsfreshness.GenericDF {
		living.WriteString("- filler line about the backup item\n")
	}

	root := fixtureTree(t, report, living.String())

	if gaps := RoutingGaps(root, Settings{}); len(gaps) != 1 {
		t.Errorf("generic-only overlap must stay UNROUTED, got %d gaps: %+v", len(gaps), gaps)
	}
}

// TestCleanCorpusExitsClean: every open row routed means zero gaps — the
// state after triage that would let the gate promote to --strict.
func TestCleanCorpusExitsClean(t *testing.T) {
	t.Parallel()

	report := "## a) next tasks\n\n| 1 | flerb maintenance work row | 🔴 `TODO` |\n"
	living := "## 1) section\n- tracked: flerb maintenance work\n"

	root := fixtureTree(t, report, living)

	if gaps := RoutingGaps(root, Settings{}); len(gaps) != 0 {
		t.Errorf("routed corpus must produce zero gaps, got %+v", gaps)
	}
}

// TestPointerRowIsRouted: the corpus's pointer convention ("TODO §15 row
// now BLOCKED") is routing — the 2026-10-03 republish false positive that
// surfaced on the first real corpus run.
func TestPointerRowIsRouted(t *testing.T) {
	t.Parallel()

	report := "## a) findings\n\n1. **verdict recorded: STALE.** TODO §15 row now 🔵 BLOCKED on the owner republish leg.\n"
	living := "## 1) section\n- anything\n"

	root := fixtureTree(t, report, living)

	if gaps := RoutingGaps(root, Settings{}); len(gaps) != 0 {
		t.Errorf("pointer-citing rows are routed by convention, got %+v", gaps)
	}
}
