package docsannotate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// markerFixture is the executable half of the marker contract between the
// docs-health skill's writer (assets/annotate.go markerFor, pinned by that
// tool's own self-test) and this gate's reader (markerRe). The gate never
// sees the kind letter — only the rendered phrase — so these strings are
// the only artifact binding the two repositories.
//
// Every string below is byte-for-byte what markerFor renders for that
// kind; changing one here without changing the writer (or the reverse) is
// exactly the drift this test exists to catch.
var markerFixture = []struct {
	kind   string
	marker string
	strike bool // the writer strikes the item for this kind
}{
	{"h", "done at `a1b2c3d`", true},
	{"v", "done — verified live 2026-11-15", true},
	{"p", "done (docs-health pass 2026-11-15)", true},
	{"w", "**Won't implement — gated upstream.**", true},
	{"n", "**NOT-DO — owner decision 2026-11-15.**", true},
	{"r", "**→ open — owner lane (TODO_LIST row)**", false},
	{"s", "**Superseded — docs-health successor report.**", true},
}

// TestMarkerVocabularyMatchesTheGate pins which rendered markers count as
// a verdict for rule A (struck items) and which do not.
//
// The routed kind (`r`) is deliberately NOT a rule-A marker: the writer
// never strikes the item for it, so rule A cannot see it. Its resolution
// is owned by the rows gate, which classifies the arrow form as ROUTED
// (see internal/docsrows). Pinning the negative here prevents a future
// widening of markerRe from silently absorbing the routed class.
//
// The routed WORD still counts (2026-11-22, #3): a STRUCK routed row —
// `~~item~~ **→ routed to TODO_LIST #32**`, the natural harvest closure —
// must pass rule A, because docsrouting already treats that row as
// closed. TestStruckRoutedRowPassesRuleA pins that leg; the fixture
// phrase above contains no "routed" word, so both contracts hold at once.
func TestMarkerVocabularyMatchesTheGate(t *testing.T) {
	t.Parallel()

	for _, f := range markerFixture {
		got := markerRe.MatchString(f.marker)
		if got != f.strike {
			t.Errorf(
				"kind %q renders %q: markerRe match = %v, want %v (the writer's strike-ness is the contract)",
				f.kind, f.marker, got, f.strike,
			)
		}
	}
}

// TestStruckItemWithEveryStrikeVerdictIsClean runs the gate over one
// struck item per strike-verdict kind. This is the round-trip leg: if the
// writer renames a phrase and the gate is not updated, this fails.
func TestStruckItemWithEveryStrikeVerdictIsClean(t *testing.T) {
	t.Parallel()

	var b strings.Builder

	b.WriteString("# Report\n\n## f) Next\n\n")

	n := 0

	for _, f := range markerFixture {
		if !f.strike {
			continue
		}

		n++
		b.WriteString(itemLine(n, "an open claim", f.marker))
	}

	root := t.TempDir()
	path := filepath.Join(root, "report.md")

	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	var infos []Finding

	findings, struck, err := checkFile(root, path, gitRunner{root: root, missing: true}, false, &infos)
	if err != nil {
		t.Fatal(err)
	}

	if struck != n {
		t.Fatalf("struck items = %d, want %d", struck, n)
	}

	if len(findings) != 0 {
		t.Fatalf("every strike verdict must pass rule A, got findings: %+v", findings)
	}
}

// TestSabotageRenamedMarkerIsFlagged proves the test has teeth: a phrase
// that reads like a verdict to a human but carries none of the writer's
// keywords must be refused. Without this leg, loosening markerRe to
// `.*` would still leave the round-trip test green.
func TestSabotageRenamedMarkerIsFlagged(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "report.md")

	body := "# Report\n\n## f) Next\n\n" +
		itemLine(1, "an open claim", "**Discarded — replaced by the battery.**") + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	var infos []Finding

	findings, _, err := checkFile(root, path, gitRunner{root: root, missing: true}, false, &infos)
	if err != nil {
		t.Fatal(err)
	}

	if len(findings) != 1 {
		t.Fatalf("sabotage marker must yield exactly one finding, got %+v", findings)
	}

	if findings[0].Rule != "A/evidence" {
		t.Fatalf("sabotage finding rule = %q, want A/evidence", findings[0].Rule)
	}
}

// TestRoutedMarkerOnAnUnstruckItemIsSilent pins the routed kind's
// intended invisibility at THIS gate: it is unstruck by construction, so
// rule A has nothing to judge. The rows gate owns it. T11's decision was
// "intended", not "hole to close here".
func TestRoutedMarkerOnAnUnstruckItemIsSilent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "report.md")

	body := "# Report\n\n## f) Next\n\n1. An open claim " + markerFixture[5].marker + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	var infos []Finding

	findings, struck, err := checkFile(root, path, gitRunner{root: root, missing: true}, false, &infos)
	if err != nil {
		t.Fatal(err)
	}

	if struck != 0 {
		t.Fatalf("routed verdict must not strike the item, struck = %d", struck)
	}

	if len(findings) != 0 {
		t.Fatalf("routed verdict is not rule A's jurisdiction, got %+v", findings)
	}
}

// TestStruckRoutedRowPassesRuleA is issue #3's minimal repro: the routing
// gate accepts `**→ routed to TODO_LIST #32**` as a completed route, so
// striking that closed row must not fail rule A. Before the fix the two
// gates disagreed — routing green, annotate red on the same row.
func TestStruckRoutedRowPassesRuleA(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "repro-report.md")

	body := "# Repro Report\n\n## f) Forward items\n\n" +
		"1. Item moved to the living backlog\n" +
		"   ~~1. Item moved to the living backlog~~ **→ routed to TODO_LIST #32 (harvested 2026-01-01)**\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	var infos []Finding

	findings, struck, err := checkFile(root, path, gitRunner{root: root, missing: true}, false, &infos)
	if err != nil {
		t.Fatal(err)
	}

	if struck != 1 {
		t.Fatalf("struck routed row must be seen as struck, struck = %d", struck)
	}

	if len(findings) != 0 {
		t.Fatalf("struck routed row must pass rule A (routing already closed it), got %+v", findings)
	}
}

// itemLine renders a numbered prose item so the fixture reads like a real
// report section rather than a table row (rule A applies to both).
func itemLine(n int, claim, verdict string) string {
	return fmt.Sprintf("%d. ~~%s~~ %s\n", n, claim, verdict)
}
