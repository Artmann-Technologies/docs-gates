package docsannotate

import (
	"regexp"
	"sort"
	"strings"
)

var (
	itemStartRe = regexp.MustCompile(`^\s*(?:\d+[.)]|[-*+])\s+\S`)
	rowStartRe  = regexp.MustCompile(`^\|\s*(?:~~\d+~~|\d+)\s*\|`)
	sepRowRe    = regexp.MustCompile(`^\|[\s:|-]+\|\s*$`)
	headingRe   = regexp.MustCompile(`^#{1,6}\s`)
	fenceRe     = regexp.MustCompile(`^\s*` + "```")
	strikeRe    = regexp.MustCompile(`(?s)~~.*?~~`)
	// A `~~` inside inline code is a literal (e.g. the completeness gate's
	// own grep over archived reports), not a strike. Neutralise the tildes
	// inside code spans before strike detection.
	inlineCodeRe = regexp.MustCompile("`[^`]*`")
	markerRe     = regexp.MustCompile(
		`(?i)\b(?:done|resolved|superseded|fixed|adopted|decided|verified|` +
			`won'?t implement|not-do|duplicate|struck|moot|no longer)\b`,
	)
	openPhrases = regexp.MustCompile(
		`(?i)(?:\bstill (?:open|stale|owed|outstanding)|\bremains open|\bnot yet\b|\bowes\b|\bowed\b|\bnot deployed yet\b)`,
	)
)

// MaskCode neutralises `~~` inside inline code spans (literal, not a strike).
func MaskCode(text string) string {
	return inlineCodeRe.ReplaceAllStringFunc(text, func(m string) string {
		return strings.ReplaceAll(m, "~~", "--")
	})
}

// VerdictText is the item text with every struck span removed — the
// verdict's own words.
func VerdictText(item string) string {
	return strikeRe.ReplaceAllString(item, " ")
}

type block struct {
	start  int
	text   string
	isItem bool
}

// ExtractItems returns (1-based start line, text, is_item) for every
// checkable block. Numbered/bulleted items and numbered table rows take
// their continuation lines with them (multi-line strikes are the norm
// here). A line holding a strike that no item owns becomes a standalone
// block, so prose corrections are checked too.
func ExtractItems(lines []string) []block {
	var items []block

	covered := map[int]bool{}

	var current []string

	currentStart := 0
	currentIsItem := true
	fenced := false

	flush := func() {
		if len(current) > 0 {
			items = append(items, block{currentStart, strings.TrimSpace(strings.Join(current, " ")), currentIsItem})
			current = nil
		}
	}

	for idx, raw := range lines {
		num := idx + 1

		if fenceRe.MatchString(raw) {
			fenced = !fenced

			flush()

			continue
		}

		if fenced {
			continue
		}

		if headingRe.MatchString(raw) {
			flush()

			continue
		}

		if strings.TrimSpace(raw) == "" {
			flush()

			continue
		}

		if itemStartRe.MatchString(raw) || rowStartRe.MatchString(raw) {
			flush()

			currentStart = num
			currentIsItem = true
			current = []string{MaskCode(strings.TrimSpace(raw))}
			covered[num] = true

			continue
		}

		if sepRowRe.MatchString(raw) {
			flush()

			continue
		}

		if len(current) > 0 {
			current = append(current, MaskCode(strings.TrimSpace(raw)))
			covered[num] = true

			continue
		}
		// Unattached continuation or a bare prose line.
		if !strings.HasPrefix(raw, " ") {
			flush()
		}
	}

	flush()

	items = append(items, strayStrikes(lines, covered)...)

	sort.SliceStable(items, func(i, j int) bool { return items[i].start < items[j].start })

	return items
}

// strayStrikes owns the lines no block claimed that still carry a strike, so
// bare prose corrections are checked too.
func strayStrikes(lines []string, covered map[int]bool) []block {
	var strikes []block

	for idx, raw := range lines {
		num := idx + 1
		if covered[num] || !strings.Contains(MaskCode(raw), "~~") {
			continue
		}

		if headingRe.MatchString(raw) || fenceRe.MatchString(raw) {
			continue
		}

		strikes = append(strikes, block{num, MaskCode(strings.TrimSpace(raw)), false})
	}

	return strikes
}
