package docsannotate

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// numberInsideLineRe matches a COLUMN-ZERO struck span that opens with
// the list-item number inside the strike (`~~14. text~~ done`) — the
// 2026-09-26 incident shape: the line stops parsing as a list item, so
// dprint absorbs it into the PREVIOUS item and renumbers the whole tail
// (one malformed verdict renumbered 22 lines). Deliberately anchored
// with NO leading-whitespace tolerance: an INDENTED struck line is an
// item continuation (no renumber hazard), a column-zero one is the
// malformed item itself. Table rows (`| …`), checkbox rows (`- [x] …`),
// and prose quoting a number (`the ~~old 12. claim~~ stays`) cannot
// match the anchor by construction; inline code is masked before the
// match (the AGENTS rule text itself carries “ `~~14. …~~ done` “).
var numberInsideLineRe = regexp.MustCompile(`^~~\s*\d+[.)]\s`)

// CheckVerdictPlacement is the verdict-PLACEMENT gate over the annotate
// corpus (the e.4 lane, 2026-11-12): a struck item must keep both its
// verdict word and its list-item number OUTSIDE the strike spans,
// because a renderer strikes `~~…~~` through — a verdict inside the
// span renders deleted, and a number inside stops the line parsing as
// a list item (dprint renumbers the tail, format-safety's 2026-09-26
// find). Rule A already answers "struck item with no verdict marker"
// for the whole class; this lens names the two REPAIRABLE placements:
//
//	P/verdict-inside - a struck ITEM whose tail (rule A's VerdictText)
//	    carries no marker but some `~~…~~` span does: the verdict word
//	    exists and is trapped inside the strike. Remedy: move it
//	    OUTSIDE the closing `~~`.
//	P/number-inside  - a column-zero LINE opening with `~~<digits>.` :
//	    the list-item number is inside the strike (ExtractItems rightly
//	    refuses it as an item — the anchor being swallowed IS the
//	    hazard — so this class scans raw lines, fence-aware, inline
//	    code masked). Remedy: strike the TEXT only
//	    (`14. ~~text~~ done (…)`).
//
// Table rows, checkbox rows, prose strikes, indented continuations, and
// fenced content are exempt (per-class, pinned in placement_test.go).
// Returns the exit code: 0 clean, 1 findings, 2 usage (malformed repo
// config).
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

	gatecli.Outf("verdict placement: %d file(s), %d struck block(s), %d finding(s)\n",
		len(files), struck, len(findings))

	if len(findings) == 0 {
		gatecli.Outln("docs-strike: verdict placement CLEAN")

		return gatecli.ExitClean
	}

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

		lines := strings.Split(string(data), "\n")

		for _, b := range ExtractItems(lines) {
			if !strings.Contains(b.text, "~~") {
				continue
			}

			struck++

			if finding, inside := verdictInside(rel, b); inside {
				findings = append(findings, finding)
			}
		}

		findings = append(findings, numbersInside(rel, lines)...)
	}

	return findings, struck
}

// verdictInside reports the P/verdict-inside finding for one struck
// block: an ITEM whose verdict tail carries no marker while some strike
// span does — the verdict word exists but renders deleted.
func verdictInside(rel string, b block) (Finding, bool) {
	if !b.isItem || markerRe.MatchString(VerdictText(b.text)) {
		return Finding{}, false
	}

	for _, span := range strikeRe.FindAllString(b.text, -1) {
		if markerRe.MatchString(span) {
			return Finding{
				Rule: "P/verdict-inside", Path: rel, Line: b.start,
				Detail: "verdict word inside the strike - move it OUTSIDE the closing ~~",
				Text:   gatecli.Truncate(b.text, snippetMax),
			}, true
		}
	}

	return Finding{}, false
}

// numbersInside scans raw lines for the P/number-inside shape. The
// fence toggle mirrors ExtractItems' own boundary protocol, so fenced
// content (code, never renumbered) stays exempt.
func numbersInside(rel string, lines []string) []Finding {
	var findings []Finding

	fenced := false

	for idx, raw := range lines {
		if fenceRe.MatchString(raw) {
			fenced = !fenced

			continue
		}

		if fenced || !numberInsideLineRe.MatchString(MaskCode(raw)) {
			continue
		}

		findings = append(findings, Finding{
			Rule: "P/number-inside", Path: rel, Line: idx + 1,
			Detail: "the list-item NUMBER is inside the strike - strike the TEXT only " +
				"(`14. ~~text~~ done (…)`); dprint renumbers the tail",
			Text: gatecli.Truncate(strings.TrimSpace(raw), snippetMax),
		})
	}

	return findings
}
