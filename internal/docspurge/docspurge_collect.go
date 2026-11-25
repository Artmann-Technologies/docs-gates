package docspurge

import (
	"fmt"
	"strings"
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/docscadence"
)

// guardNote is one collected line the status-cell guard flagged: a
// refusal (a struck TABLE row whose status cell still reads OPEN — edit
// noise, never done work) or a warning (a struck table row with no
// recognizable DONE cell — collected, but named).
type guardNote struct {
	lineNo int
	cell   string
	refuse bool
}

// collectStruck returns the struck rows of a TODO file in order, plus
// the status-cell guard notes: every line carrying a ~~…~~ wrap with
// non-empty task text. Empty struck cells (`~~ ~~`, the
// formatter-stable padding shape) are decoration, not strikes, and
// never collect. A struck TABLE row whose status cell still reads OPEN
// (cadence.IsOpenCell) is refused — not collected, named — and one
// with no DONE cell (cadence.IsDoneCell) collects under a warning.
func collectStruck(content string) ([]struckRow, []guardNote) {
	var rows []struckRow

	var guards []guardNote

	for i, line := range strings.Split(content, "\n") {
		row, ok := splitStrike(line, i+1)
		if !ok {
			continue
		}

		if strings.HasPrefix(line, "|") {
			if cell, open := statusCell(line, docscadence.IsOpenCell); open {
				guards = append(guards, guardNote{lineNo: i + 1, cell: cell, refuse: true})

				continue
			}

			if _, done := statusCell(line, docscadence.IsDoneCell); !done {
				guards = append(guards, guardNote{lineNo: i + 1, refuse: false})
			}
		}

		rows = append(rows, row)
	}

	return rows, guards
}

// statusCell finds the first cell of a table row matching the predicate,
// returning its trimmed text. Cell-scoped on purpose: a bare substring
// scan would read task text that merely CONTAINS a status word.
func statusCell(line string, predicate func(string) bool) (string, bool) {
	for cell := range strings.SplitSeq(line, "|") {
		if c := strings.TrimSpace(cell); predicate(c) {
			return c, true
		}
	}

	return "", false
}

// splitStrike extracts the strike at a line: the task text between the
// first ~~ pair and the verdict tail after it, cut at the next table
// pipe. A line without a wrap, or one whose wrap is empty on both sides,
// is not a strike.
//
// The wrap that swallowed the leading pipe (the shift-strike shape the
// cadence detector faults) is normalized first: the opening ~~ moves
// inside the first cell, and a task span that still spans cells keeps
// only its first cell — the draft renders one bullet per row, and a
// pipe-laden task renders as a broken table row in the CHANGELOG.
func splitStrike(line string, lineNo int) (struckRow, bool) {
	raw := line

	if normalized, swallowed := normalizeSwallowedPipe(line); swallowed {
		raw = normalized
	}

	first := strings.Index(raw, "~~")
	if first < 0 {
		return struckRow{}, false
	}

	second := strings.Index(raw[first+2:], "~~")
	if second < 0 {
		return struckRow{}, false
	}

	task := raw[first+2 : first+2+second]

	verdictTail := raw[first+2+second+2:]
	if cut := strings.Index(verdictTail, "|"); cut >= 0 {
		verdictTail = verdictTail[:cut]
	}

	verdict := strings.TrimSpace(verdictTail)

	if i := strings.Index(task, "|"); i >= 0 {
		task = task[:i]
	}

	// A canonical one-cell strike keeps its verdict in the STATUS cell
	// (`~~task~~ | ✅ DONE 2026-11-14 | …`), not in the wrap tail — the
	// cell parse, not the marker split, is where the verdict lives.
	if verdict == "" && strings.HasPrefix(raw, "|") {
		if cell, done := statusCell(raw, docscadence.IsDoneCell); done {
			verdict = cell
		}
	}

	if strings.TrimSpace(task) == "" && verdict == "" {
		return struckRow{}, false
	}

	return struckRow{lineNo: lineNo, oldLine: line, task: task, verdict: verdict}, true
}

// normalizeSwallowedPipe repairs the struck-before-the-pipe shape: a
// line whose wrap opened before the leading pipe reads as `~~| cells
// |~~ verdict`, which would otherwise put the whole pipe-laden row into
// the drafted bullet. The repair rewrites to the canonical leading pipe
// with the wrap inside the first cell; ok reports whether the shape
// matched.
func normalizeSwallowedPipe(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)

	if !strings.HasPrefix(trimmed, "~~|") {
		return line, false
	}

	return "| ~~" + trimmed[3:], true
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

