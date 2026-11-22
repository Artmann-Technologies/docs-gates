package docsfreshness

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

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
			filepath.Base(report), m.LineNo, gatecli.Truncate(m.Text, 80),
		)
	}
}
