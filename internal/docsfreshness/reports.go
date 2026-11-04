package docsfreshness

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

var (
	// SectionFRe matches a status report's §f heading (case-insensitive).
	// Routing shares this definition so both gates agree on where §f
	// begins — freshness harvests ONLY §f, routing skips ONLY §f; two
	// spellings of the boundary would let a report be harvested AND
	// triaged at once.
	SectionFRe = regexp.MustCompile(`(?i)^#{2,}\s*f\)`)
	// HeadingRe matches the 1-2-hash heading that closes a §f section
	// (see SectionFRe for why the definition is shared with routing).
	HeadingRe = regexp.MustCompile(`^#{1,2}\s`)
	proseRe   = regexp.MustCompile(`^\s*(\d+)[.)]\s+(.*)$`)
	rowRe     = regexp.MustCompile(`^\|\s*(~~\d+~~|\d+)\s*\|`)
	sepRowRe  = regexp.MustCompile(`^\|[-: ]+\|?\s*$`)
)

func NewestReport(root string, settings Settings) string {
	entries, err := filepath.Glob(settings.reportGlobPath(root))
	if err != nil || len(entries) == 0 {
		return ""
	}

	sort.Strings(entries)

	for _, entry := range slices.Backward(entries) {
		if info, err := os.Stat(entry); err == nil && info.Mode().IsRegular() {
			return entry
		}
	}

	return ""
}

func ExtractSectionF(report string) string {
	data, err := os.ReadFile(report)
	if err != nil {
		return ""
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
		return ""
	}

	var section []string

	for _, line := range lines[start:] {
		if HeadingRe.MatchString(line) {
			break
		}

		section = append(section, line)
	}

	return strings.Join(section, "\n")
}

func ExtractItems(section string) []string {
	var (
		items   []string
		current []string
	)

	flush := func() {
		if len(current) > 0 {
			text := strings.TrimSpace(strings.Join(current, " "))
			if text != "" {
				items = append(items, text)
			}

			current = nil
		}
	}

	for line := range strings.SplitSeq(section, "\n") {
		switch {
		case proseRe.MatchString(line):
			flush()

			current = append(current, strings.TrimSpace(line))
		case rowRe.MatchString(line):
			flush()
			// Keep the whole row: a struck number cell (~~5~~) must leave
			// its ~~ marker in the item text so it is skipped as resolved.
			current = append(current, strings.TrimSpace(line))
		case sepRowRe.MatchString(line) || strings.HasPrefix(line, "| #"):
			continue
		case strings.TrimSpace(line) != "":
			if len(current) > 0 {
				current = append(current, strings.TrimSpace(line))
			}
		default:
			flush()
		}
	}

	flush()

	return items
}
