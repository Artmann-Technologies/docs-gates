package docspurge

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// collectStruck returns the struck rows of a TODO file in order: every
// line carrying a ~~…~~ wrap with non-empty task text. Empty struck
// cells (`~~ ~~`, the formatter-stable padding shape) are decoration,
// not strikes, and never collect.
func collectStruck(content string) []struckRow {
	var rows []struckRow

	for i, line := range strings.Split(content, "\n") {
		row, ok := splitStrike(line, i+1)
		if !ok {
			continue
		}

		rows = append(rows, row)
	}

	return rows
}

// splitStrike extracts the strike at a line: the task text between the
// first ~~ pair and the verdict tail after it, cut at the next table
// pipe. A line without a wrap, or one whose wrap is empty on both sides,
// is not a strike.
func splitStrike(line string, lineNo int) (struckRow, bool) {
	first := strings.Index(line, "~~")
	if first < 0 {
		return struckRow{}, false
	}

	second := strings.Index(line[first+2:], "~~")
	if second < 0 {
		return struckRow{}, false
	}

	task := line[first+2 : first+2+second]

	verdictTail := line[first+2+second+2:]
	if cut := strings.Index(verdictTail, "|"); cut >= 0 {
		verdictTail = verdictTail[:cut]
	}

	verdict := strings.TrimSpace(verdictTail)

	if strings.TrimSpace(task) == "" && verdict == "" {
		return struckRow{}, false
	}

	return struckRow{lineNo: lineNo, oldLine: line, task: task, verdict: verdict}, true
}

// draftBullets renders one CHANGELOG bullet per struck row: the unwrapped
// task text and the verdict, dated, in the prose shape the manual purges
// established.
func draftBullets(rows []struckRow) []string {
	date := time.Now().Format("2006-01-02")

	bullets := make([]string, 0, len(rows))

	for _, row := range rows {
		entry := fmt.Sprintf("- TODO_LIST purge (%s): %s", date, strings.TrimSpace(row.task))
		if row.verdict != "" {
			entry += " — " + row.verdict
		}

		bullets = append(bullets, entry)
	}

	return bullets
}

// anchorScope is the CHANGELOG anchor whose absence refuses a purge
// draft: the [Unreleased] heading anywhere in the file, or the Changed
// section constrained to the [Unreleased] block. Its String spelling is
// the refusal message's anchor phrase, so the two refusal shapes stay
// byte-for-byte with the historical fmt.Errorf texts.
type anchorScope int

const (
	anchorUnreleasedHeading anchorScope = iota + 1
	anchorChangedSection
)

func (s anchorScope) String() string {
	switch s {
	case anchorUnreleasedHeading:
		return unreleasedHeading + " heading"
	case anchorChangedSection:
		return changedHeading + " section under " + strings.TrimPrefix(unreleasedHeading, "## ")
	default:
		return fmt.Sprintf("anchorScope(%d)", int(s))
	}
}

// level returns the heading marker of the anchor's own level — the
// class a "nearest heading" hint searches (## for the [Unreleased]
// anchor, ### for the Changed section).
func (s anchorScope) level() string {
	if s == anchorChangedSection {
		return "###"
	}

	return "##"
}

// missingAnchorError is the package's static refusal type: the changelog
// lacks an anchor the draft needs, so no insertion point is guessed.
// nearestLine names the file's nearest heading of the anchor's own
// level (1-based; 0 = the file has none) — the fix is editing near a
// heading the reader can actually find, not at a byte offset.
type missingAnchorError struct {
	changelog   string
	scope       anchorScope
	nearestLine int
}

func (e *missingAnchorError) Error() string {
	msg := fmt.Sprintf("%s has no %s — refusing to guess an insertion point", e.changelog, e.scope)
	if e.nearestLine > 0 {
		msg += fmt.Sprintf(" (nearest %s heading at line %d)", e.scope.level(), e.nearestLine)
	}

	return msg
}

// newMissingAnchorError builds one missingAnchorError with the refusal
// message; nearestLine is the file's nearest same-level heading line
// (0 when the file has none).
func newMissingAnchorError(scope anchorScope, changelog string, nearestLine int) *missingAnchorError {
	return &missingAnchorError{changelog: changelog, scope: scope, nearestLine: nearestLine}
}

// writePreservingMode writes content with the file's existing permission
// bits (a fresh write gets the default file mode).
func writePreservingMode(path, content string) error {
	mode := os.FileMode(defaultFileMode)

	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}

	return os.WriteFile( //nolint:gosec // G306,G703: the mode IS the source file's; the path argument IS the tool's input contract
		path,
		[]byte(content),
		mode,
	)
}
