// Package docsrouting closes the non-§f blind spot of the docs-health
// ritual (ROADMAP "tooling answers", standing from the 2026-10-04 coverage
// decision): docs-freshness harvests ONLY §f sections, so live actionable
// rows sitting in next-tasks tables, checklists or other sections of
// UNARCHIVED reports are invisible to it — the SIP-gateway password
// rotation sat unrouted for three weeks despite five green freshness
// passes. This lint scans the non-§f body of every unarchived
// docs/status report for OPEN actionable rows (status markers or
// checkboxes, not struck, not skip-marked) and reports those whose tokens
// appear NOWHERE in the living docs (TODO_LIST/ROADMAP/CHANGELOG) under
// the same matched/distinctive bar the freshness gate uses — inverted:
// a row failing the bar is UNROUTED.
//
// Report-only by default (the triage-on-touch doctrine): run it in the
// docs-gates ritual, triage what it names (harvest into a living doc, or
// annotate the row resolved, or mark the line <!-- routing:skip -->).
// --strict exits 1 once the corpus is clean and the gate can be promoted.
package docsrouting

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/docsfreshness"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// SkipMarker exempts a line from routing triage — the fresh:skip sibling.
// Use ONLY for rows that look actionable but are deliberately historical.
const SkipMarker = "<!-- routing:skip -->"

var (
	// Candidate = an OPEN actionable row anywhere in the report body:
	// the 🔴 TODO / 🔵 BLOCKED status markers the reports and TODO_LIST
	// share, or an unticked checkbox. Struck (~~) rows are resolved.
	markerRe   = regexp.MustCompile(`🔴|🔵`)
	checkboxRe = regexp.MustCompile(`^\s*[-*]\s+\[ \]`)
	openRowRe  = regexp.MustCompile(`\|\s*(TODO|BLOCKED)\s*\|`)
	// pointerRe is the corpus's own routing convention: a row citing an
	// explicit living-doc section pointer ("TODO_LIST §15 row", "ROADMAP §2") is routed
	// BY POINTER — the token bar cannot see pointers, and treating such
	// rows as unrouted flags every pointer-style verdict (the 2026-10-03
	// republish false positive, found on the first corpus run).
	pointerRe = regexp.MustCompile(`(?i)\b(?:todo(?:_list)?|roadmap|changelog)\s*§\s*\d+`)
)

// hintTokens caps the token evidence printed per gap.
const hintTokens = 3

// gapCap is the initial gaps capacity (a floor: most reports carry few
// open rows; the slice grows past it without harm).
const gapCap = 16

// rowPreviewRunes bounds the gap row preview printed by Run.
const rowPreviewRunes = 100

// Gap is one unrouted actionable row.
type Gap struct {
	Report string
	LineNo int
	Row    string
	Hint   string
}

// RoutingGaps scans every unarchived report under the configured report
// dir and returns the open non-§f actionable rows the living docs do not
// reflect.
func RoutingGaps(root string, s Settings) []Gap {
	living := docsfreshness.LivingDocsLines(root, s.livingDocs())

	gaps := make([]Gap, 0, gapCap)

	reports, _ := filepath.Glob(
		filepath.Join(root, s.reportDir(), s.reportGlob()),
	) //nolint:erraudit // constant pattern cannot be malformed
	sort.Strings(reports)

	for _, report := range reports {
		gaps = append(gaps, reportGaps(report, living, s.df())...)
	}

	return gaps
}

func reportGaps(report string, living []string, genericDF int) []Gap {
	data, err := os.ReadFile(report)
	if err != nil {
		return nil
	}

	var (
		gaps []Gap
		inF  bool
	)

	for i, line := range strings.Split(string(data), "\n") {
		switch {
		case docsfreshness.SectionFRe.MatchString(line):
			inF = true

			continue
		case docsfreshness.HeadingRe.MatchString(line):
			inF = false

			continue
		}

		if inF || !isOpenActionable(line) {
			continue
		}

		if !routed(line, living, genericDF) {
			gaps = append(gaps, Gap{
				Report: filepath.Base(report),
				LineNo: i + 1,
				Row:    strings.TrimSpace(line),
				Hint:   tokenHint(line, living),
			})
		}
	}

	return gaps
}

func isOpenActionable(line string) bool {
	if strings.Contains(line, "~~") || strings.Contains(line, SkipMarker) {
		return false
	}

	return markerRe.MatchString(line) || checkboxRe.MatchString(line) || openRowRe.MatchString(line)
}

// routed applies the freshness bar inverted: the row must share at least
// the quorum of tokens with the living corpus, one of them distinctive
// (on fewer than genericDF living lines) — the same semantics under
// which a harvested §f item counts as represented.
func routed(row string, living []string, genericDF int) bool {
	if pointerRe.MatchString(row) {
		return true // routed by explicit living-doc pointer (pointerRe)
	}

	toks := docsfreshness.Tokens(row)
	if len(toks) == 0 {
		return true
	}

	need := 2
	if len(toks) < need {
		need = 1
	}

	found, distinctive := 0, false

	for _, tok := range toks {
		if !tokenIn(tok, living) {
			continue
		}

		found++

		if livingLineCount(tok, living) < genericDF {
			distinctive = true
		}
	}

	return found >= need && distinctive
}

func tokenIn(token string, living []string) bool {
	for _, line := range living {
		if strings.Contains(line, token) {
			return true
		}
	}

	return len(token) >= docsfreshness.MinTokenLen &&
		squashContains(docsfreshness.Squash(token), living)
}

func squashContains(key string, living []string) bool {
	if key == "" {
		return false
	}

	for _, line := range living {
		if strings.Contains(docsfreshness.Squash(line), key) {
			return true
		}
	}

	return false
}

func livingLineCount(token string, living []string) int {
	n := 0

	for _, line := range living {
		if strings.Contains(line, token) {
			n++
		}
	}

	return n
}

// tokenHint names the row's most useful search tokens (rare-in-living
// first) so triage is a copy-paste, not a re-read.
func tokenHint(row string, living []string) string {
	toks := docsfreshness.Tokens(row)

	type counted struct {
		tok string
		n   int
	}

	counts := make([]counted, 0, len(toks))

	for _, tok := range toks {
		counts = append(counts, counted{tok, livingLineCount(tok, living)})
	}

	sort.SliceStable(counts, func(i, j int) bool { return counts[i].n < counts[j].n })

	parts := make([]string, 0, hintTokens)

	for _, c := range counts {
		if len(parts) == hintTokens {
			break
		}

		parts = append(parts, fmt.Sprintf("%s(%d lines)", c.tok, c.n))
	}

	return strings.Join(parts, " ")
}

// Run is the CLI: report-only by default, --strict fails on any gap.
// The report corpus and the living-docs list come from the repo's
// .docs-gates.json when present.
func Run(args []string) int {
	flags, root, err := gatecli.Parse(args, "--strict")
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	cfg, err := gatecli.LoadConfig(root)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	s := settingsFromConfig(cfg)

	gaps := RoutingGaps(root, s)

	for _, gap := range gaps {
		gatecli.Outf("UNROUTED [%s:%d] %s\n", gap.Report, gap.LineNo, gatecli.Truncate(gap.Row, rowPreviewRunes))
		gatecli.Outf("          tokens: %s\n", gap.Hint)
	}

	verb := "report-only"
	if flags["--strict"] {
		verb = "STRICT"
	}

	gatecli.Outf("docs-routing (%s): %d unarchived report(s) scanned, %d unrouted open row(s)\n",
		verb, countUnarchived(root, s), len(gaps))

	if flags["--strict"] && len(gaps) > 0 {
		gatecli.Errln("docs-routing --strict: DRIFT — open actionable rows outside §f sections " +
			"are not represented in TODO_LIST/ROADMAP/CHANGELOG. Harvest them, annotate them " +
			"resolved, or mark the line with <!-- routing:skip -->.")

		return 1
	}

	return 0
}

func countUnarchived(root string, s Settings) int {
	reports, _ := filepath.Glob(
		filepath.Join(root, s.reportDir(), s.reportGlob()),
	) //nolint:erraudit // constant pattern cannot be malformed

	return len(reports)
}
