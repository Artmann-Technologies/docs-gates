package docsannotate

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// numberInsideRe matches a struck span that OPENS an item's block text
// with the list-item number inside the strike (`~~14. text~~ done`) —
// the shape that stops the line parsing as a list item and lets dprint
// absorb it into the PREVIOUS item and renumber the whole tail (burned
// 2026-09-26: one malformed verdict renumbered 22 lines). Anchored at
// block-text START (block text is trimmed, so the span must open the
// item): struck prose merely quoting a number, and checkbox rows
// (`- [x] ~~…~~`), can never match. Numbered TABLE rows are exempt via
// the `|` prefix guard — their number cell is rowStartRe's own
// `~~\d+~~` shape, a deliberate strike of the number ALONE.
var numberInsideRe = regexp.MustCompile(`^~~\s*\d+[.)]\s`)

// CheckVerdictPlacement is the verdict-PLACEMENT gate over the annotate
// corpus (the e.4 lane, 2026-11-12): a struck item must keep both its
// verdict word and its list-item number OUTSIDE the strike spans,
// because a renderer strikes `~~…~~` through — a verdict inside the
// span renders deleted, and a number inside stops the line parsing as
// a list item (dprint renumbers the tail, format-safety's 2026-09-26
// find). Rule A already answers "struck item with no verdict marker"
// for the whole class; this lens names the two REPAIRABLE placements:
//
//	P/verdict-inside - the tail (rule A's VerdictText) carries no
//	    marker but some `~~…~~` span does: the verdict word exists and
//	    is trapped inside the strike. Remedy: move it OUTSIDE the
//	    closing `~~`.
//	P/number-inside  - the block text OPENS with `~~<digits>.` : the
//	    list-item number is inside the strike. Remedy: strike the TEXT
//	    only (`14. ~~text~~ done (…)`).
//
// Table rows, checkbox rows, prose strikes, and fenced content are
// exempt (per-class, pinned in placement_test.go). Returns the exit
// code: 0 clean, 1 findings, 2 usage (malformed repo config).
func CheckVerdictPlacement(root string) int {
	settings, err := loadSettings(root)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	docGlobs := append([]string{
		filepath.Join(settings.dir(), "*.md"),
		filepath.Join(settings.dir(), settings.archivedSub(), "*.md"),
	}, settings.extraGlobs()...)

	files := collectDocFiles(root, docGlobs)

	findings, struck := placementScan(root, files)

	for _, finding := range findings {
		gatecli.Outf("%s: %s:%d: %s\n", finding.Rule, finding.Path, finding.Line, finding.Detail)
	}

	if len(findings) == 0 {
		gatecli.Outf("verdict placement: %d file(s), %d struck block(s), 0 finding(s)\n", len(files), struck)
		gatecli.Outln("docs-strike: verdict placement CLEAN")

		return gatecli.ExitClean
	}

	gatecli.Outf("verdict placement: %d file(s), %d struck block(s), %d finding(s)\n",
		len(files), struck, len(findings))
	gatecli.Errln("docs-strike: verdict placement FAIL - keep the verdict word AND the item " +
		"number OUTSIDE the ~~ spans (14. ~~text~~ done (…)); inside, the renderer deletes " +
		"the verdict and dprint renumbers the list tail.")

	return gatecli.ExitFindings
}

// placementScan walks the corpus files and returns the placement
// findings plus the struck-block count (the summary's denominator).
func placementScan(root string, files []string) ([]Finding, int) {
	var findings []Finding

	struck := 0

	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			continue //nolint:erraudit // an unreadable report is skipped; the rest of the corpus still gets gated
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			continue //nolint:erraudit // Rel against a collected descendant cannot fail; skip, never crash the gate
		}

		for _, b := range ExtractItems(strings.Split(string(data), "\n")) {
			if !strings.Contains(b.text, "~~") {
				continue
			}

			struck++

			findings = append(findings, placementFindings(rel, b)...)
		}
	}

	return findings, struck
}

// placementFindings returns the P-rule findings for one struck block.
func placementFindings(rel string, b block) []Finding {
	var findings []Finding

	if b.isItem && !markerRe.MatchString(VerdictText(b.text)) {
		for _, span := range strikeRe.FindAllString(b.text, -1) {
			if markerRe.MatchString(span) {
				findings = append(findings, Finding{
					Rule: "P/verdict-inside", Path: rel, Line: b.start,
					Detail: "verdict word inside the strike - move it OUTSIDE the closing ~~",
					Text:   gatecli.Truncate(b.text, snippetMax),
				})

				break
			}
		}
	}

	if b.isItem && !strings.HasPrefix(b.text, "|") && numberInsideRe.MatchString(b.text) {
		findings = append(findings, Finding{
			Rule: "P/number-inside", Path: rel, Line: b.start,
			Detail: "the list-item NUMBER is inside the strike - strike the TEXT only " +
				"(`14. ~~text~~ done (…)`); dprint renumbers the tail",
			Text: gatecli.Truncate(b.text, snippetMax),
		})
	}

	return findings
}
