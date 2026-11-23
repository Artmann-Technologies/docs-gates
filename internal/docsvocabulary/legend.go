package docsvocabulary

// The legend extraction half of the vocabulary gate: locating a file's
// `## Status legend` section and collecting its token vocabulary. Split
// from the scan core (docsvocabulary.go) at the 350-line file ceiling —
// same package, no API change.

import (
	"regexp"
	"sort"
	"strings"
)

// legendHeadingRe matches the legend section heading; the level is
// unpinned so both `## Status legend` and deeper nestings resolve.
var legendHeadingRe = regexp.MustCompile(`(?i)^#{1,6}\s*status legend\b`)

// headingRe closes the legend section: any ATX heading ends it.
var headingRe = regexp.MustCompile(`^#{1,6}\s`)

// tokenRe captures one backticked ALL-CAPS token (the status
// vocabulary's shape: `FULLY_FUNCTIONAL`, `OUT_OF_SCOPE`).
var tokenRe = regexp.MustCompile("`([A-Z][A-Z0-9_]{2,})`")

// legendSection returns the text between the `## Status legend`
// heading and the next heading, plus its 1-based line span. Definitions
// live in tables AND in sentences, and the span lets the usage scan
// skip the definition table itself.
func legendSection(lines []string) (string, int, int) {
	start := -1

	for i, line := range lines {
		if legendHeadingRe.MatchString(line) {
			start = i

			break
		}
	}

	if start < 0 {
		return "", 0, 0
	}

	var section []string

	for i, line := range lines[start+1:] {
		if headingRe.MatchString(line) {
			return strings.Join(section, "\n"), start + 1, start + 1 + i
		}

		section = append(section, line)
	}

	return strings.Join(section, "\n"), start + 1, len(lines) + 1
}

// legendTokens collects every backticked ALL-CAPS token of the legend
// section — the table's status column and the prose around it both
// define tokens.
func legendTokens(section string) map[string]bool {
	tokens := map[string]bool{}

	for _, m := range tokenRe.FindAllStringSubmatch(section, -1) {
		tokens[m[1]] = true
	}

	return tokens
}

// cellTokens collects the backticked ALL-CAPS tokens of one status
// cell (an emoji prefix or a trailing note never hides the token).
func cellTokens(cell string) []string {
	matches := tokenRe.FindAllStringSubmatch(cell, -1)

	tokens := make([]string, 0, len(matches))

	for _, m := range matches {
		tokens = append(tokens, m[1])
	}

	return tokens
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}
