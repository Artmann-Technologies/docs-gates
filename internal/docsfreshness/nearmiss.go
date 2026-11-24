package docsfreshness

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// nearMissWidth bounds the quoted heading text in the WARN line so one
// warn stays on one screen row.
const nearMissWidth = 80

// sectionFNearMissRe matches a heading that READS like a §f section in a
// shape the canon SectionFRe rejects (`## §f — Follow-ups`, `## f.`,
// `## §f)`). The canon stays the one spelling — see SectionFRe for why
// the boundary definition is shared — this probe only makes the blind
// spot loud instead of silent (#9: two io-surface reports sat
// machine-unharvestable for 6 days behind `## §f —` while the battery
// stayed 5/5 PASS the whole time).
var sectionFNearMissRe = regexp.MustCompile(`(?i)^#{2,}\s*§?\s*f\b`)

// NearMissHeading is one heading that reads like a §f section while the
// canon SectionFRe rejects it: the section LOOKS harvested-ready to a
// human but freshness/audit-all/routing cannot see any of its items.
type NearMissHeading struct {
	LineNo int
	Text   string
}

// FindSectionFNearMiss returns the canon-rejected §f-shaped headings of
// one report (read from disk). Canon headings (`## f)`) never match —
// a report whose every §f heading is canon returns nil — and
// `## Feedback`-style headings carry no word boundary after the f and
// never match either.
func FindSectionFNearMiss(report string) []NearMissHeading {
	data, err := os.ReadFile(report)
	if err != nil {
		return nil
	}

	var found []NearMissHeading

	for i, line := range strings.Split(string(data), "\n") {
		if SectionFRe.MatchString(line) {
			continue
		}

		if sectionFNearMissRe.MatchString(line) {
			found = append(found, NearMissHeading{LineNo: i + 1, Text: strings.TrimSpace(line)})
		}
	}

	return found
}

// WarnSectionFNearMiss prints one WARN per near-miss heading. STDOUT on
// purpose: a stderr note is swallowed by the battery's PASS path, and
// this warn must stay visible exactly there (the annotate no-git skip
// marker precedent, 15-52 report §f.30). The canon stays strict — the
// warn never flips a verdict, it names the blind spot.
func WarnSectionFNearMiss(report string) {
	for _, m := range FindSectionFNearMiss(report) {
		gatecli.Outf(
			"WARN %s:%d: %q reads like a §f heading but misses the canon shape — its items are INVISIBLE to freshness/audit-all/routing; rename it to start \"## f)\"\n",
			filepath.Base(report),
			m.LineNo,
			gatecli.Truncate(m.Text, nearMissWidth),
		)
	}
}

// SectionFGapKind names the two ways a report reads nothing at the
// harvest while looking like it should: no canon §f heading at all, or a
// §f section whose item-shaped lines are bullets the extractor ignores.
type SectionFGapKind string

const (
	// GapMissingHeading: the report has NO canon `## f)` heading — every
	// next-step line in it is invisible to the harvest (freshness passes
	// by omission, the 11-24 audit's d2).
	GapMissingHeading SectionFGapKind = "missing-heading"
	// GapBulletItem: a §f bullet line — item-shaped, but ExtractItems
	// harvests only numbered lines and table rows, so it reads nothing.
	GapBulletItem SectionFGapKind = "bullet-item"
)

// SectionFGap is one harvest-blind spot: the report's next steps exist
// but the extractor provably reads none of them.
type SectionFGap struct {
	Kind   SectionFGapKind
	LineNo int    // 1-based; 0 for GapMissingHeading (whole-report)
	Text   string // trimmed offending line; "" for GapMissingHeading
}

var bulletItemRe = regexp.MustCompile(`^\s*[-*]\s+\S`)

// FindSectionFGaps returns the ways one report's next steps are invisible
// to the harvest WITHOUT any near-miss heading being involved: a report
// with no canon §f heading at all, or a §f section carrying bullet items
// the extractor ignores. A report whose §f is canon and item-shaped (or
// explicitly empty prose) returns nil.
func FindSectionFGaps(report string) []SectionFGap {
	data, err := os.ReadFile(report)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(data), "\n")

	start := -1

	for i, line := range lines {
		if SectionFRe.MatchString(line) {
			start = i + 1

			break
		}
	}

	if start < 0 {
		return []SectionFGap{{Kind: GapMissingHeading}}
	}

	var gaps []SectionFGap

	for i := start; i < len(lines); i++ {
		if HeadingRe.MatchString(lines[i]) {
			break
		}

		if bulletItemRe.MatchString(lines[i]) {
			gaps = append(gaps, SectionFGap{
				Kind:   GapBulletItem,
				LineNo: i + 1,
				Text:   strings.TrimSpace(lines[i]),
			})
		}
	}

	return gaps
}

// WarnSectionFGaps prints one WARN per harvest blind spot — STDOUT for
// the same reason as WarnSectionFNearMiss: the battery swallows a PASS
// row's stderr, and passing-by-omission must be loud exactly there.
func WarnSectionFGaps(report string) {
	base := filepath.Base(report)

	for _, g := range FindSectionFGaps(report) {
		switch g.Kind {
		case GapMissingHeading:
			gatecli.Outf(
				"WARN %s: no §f (\"## f)\") heading — the report's next steps are INVISIBLE to the harvest; freshness passes by omission\n",
				base,
			)
		case GapBulletItem:
			gatecli.Outf(
				"WARN %s:%d: bullet next step %q is not a harvestable item (only numbered lines and table rows are) — it reads as done while it is not\n",
				base,
				g.LineNo,
				gatecli.Truncate(g.Text, nearMissWidth),
			)
		}
	}
}
