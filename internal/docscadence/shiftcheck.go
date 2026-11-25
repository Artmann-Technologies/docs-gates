package docscadence

import (
	"fmt"
	"regexp"
	"strings"
)

// sepRowRe matches a markdown table separator row (`| --- | --- |`),
// the same shape docsstrike's table walk keys on.
var sepRowRe = regexp.MustCompile(`^\|[\s:|-]+\|\s*$`)

// strikeFault names one malformed strike row in the TODO file: the
// 1-based body line and a reason that says what shape broke.
type strikeFault struct {
	line   int
	reason string
}

// todoTable is the governing header shape of the table a row sits in:
// the split-cell count (with the leading/trailing phantom empties a
// strings.Split on "|" produces), the Task column's index, and the
// trimmed header cell names for the fault reasons — mirrored from
// docsstrike's shapeOf/mapHeader (#6: only the header names the cells;
// fleet TODOs ship 4-, 5- and 6-column layouts).
type todoTable struct {
	cells   int
	taskCol int
	names   []string
}

// splitPhantoms is the two phantom cells a strings.Split on "|"
// produces around a markdown table row (before the leading and after
// the trailing pipe).
const splitPhantoms = 2

// scanStrikeFaults walks the TODO body and names every strike whose
// SHAPE is broken. classify (the struck-row counter) reads any
// ~~…~~-plus-DONE-cell line as healthily struck, which let ten shifted
// strikes ride the cadence count for weeks: nine rows with a bare `~~`
// in the Task cell while the actual task text sat unwrapped in the
// Status column, and one row whose verdict was appended as a surplus
// cell. Four shapes fault:
//
//  1. bare wrap — the Task cell's ~~ carries no text (the nine-shifted
//     shape; the task text must sit INSIDE the wrap)
//  2. misplaced wrap — the ~~ sits in a non-Task cell
//  3. surplus/missing cells — the row's cell count disagrees with the
//     governing header (a wrap that crossed a pipe shifts everything)
//  4. struck before the pipe — the wrap swallowed the leading pipe, so
//     classify never even sees the line
//
// A canonical strike (~~task~~ in the Task cell) stays silent, as does
// any row without a ~~ and any table without a Task column (its wrap
// placement is not judgeable).
func scanStrikeFaults(body string) []strikeFault {
	lines := strings.Split(body, "\n")

	var faults []strikeFault

	for i, line := range lines {
		if fault, hit := pipeFault(line, i); hit {
			faults = append(faults, fault)

			continue
		}

		if !strings.HasPrefix(line, "|") || !strings.Contains(line, "~~") {
			continue
		}

		if fault, hit := tableFault(lines, i); hit {
			faults = append(faults, fault)
		}
	}

	return faults
}

// pipeFault names the struck-before-the-pipe shape: the wrap swallowed
// the leading pipe, so classify never even sees the line.
func pipeFault(line string, i int) (strikeFault, bool) {
	trimmed := strings.TrimSpace(line)

	if strings.HasPrefix(trimmed, "~~") && strings.Contains(trimmed, "|") {
		return strikeFault{
			line:   i + 1,
			reason: "strike swallows the leading pipe (struck before its pipe) — the row is invisible to the struck count",
		}, true
	}

	return strikeFault{}, false
}

// tableFault judges one struck-looking table row against its governing
// header: cell-count faults fire regardless of status (a shifted row is
// broken wherever its cells land), wrap-placement faults need strike
// intent (see hasStrikeIntent).
func tableFault(lines []string, i int) (strikeFault, bool) {
	table, ok := governingTable(lines, i)
	if !ok {
		return strikeFault{}, false
	}

	cells := strings.Split(lines[i], "|")

	if len(cells) != table.cells {
		return strikeFault{
			line: i + 1,
			reason: fmt.Sprintf(
				"shifted strike: %d cells against the header's %d (a wrap that crossed a pipe shifts the row)",
				len(cells)-splitPhantoms, table.cells-splitPhantoms,
			),
		}, true
	}

	if table.taskCol < 0 || !hasStrikeIntent(table, cells) {
		return strikeFault{}, false
	}

	if reason, bad := wrapFault(table, cells); bad {
		return strikeFault{line: i + 1, reason: reason}, true
	}

	return strikeFault{}, false
}

// hasStrikeIntent reports whether the row reads as a strike at all: a
// DONE status cell (the struck shape classify counts) or a strike
// attempt in the Task cell itself. A tilde pair on an OPEN row is edit
// noise or emphasis, not a misplaced strike — the purge collector owns
// that class (#10).
func hasStrikeIntent(table todoTable, cells []string) bool {
	if strings.Contains(cells[table.taskCol], "~~") {
		return true
	}

	for _, cell := range cells {
		if IsDoneCell(strings.TrimSpace(cell)) {
			return true
		}
	}

	return false
}

// governingTable derives the table shape ruling row i: the nearest
// separator row at/above i inside the same table, whose line above is
// the header (the docsstrike shapeOf walk, read-only).
func governingTable(lines []string, row int) (todoTable, bool) {
	for i := row; i > 0; i-- {
		if !strings.HasPrefix(lines[i], "|") {
			break
		}

		if !sepRowRe.MatchString(lines[i]) {
			continue
		}

		header := lines[i-1]
		if !strings.HasPrefix(header, "|") {
			break
		}

		return mapTodoHeader(strings.Split(header, "|")), true
	}

	return todoTable{}, false
}

// mapTodoHeader pins the cell count and the Task column from the
// header's own cell names (Task matched case-insensitively like
// docsstrike's mapHeader; no Task column keeps taskCol at -1).
func mapTodoHeader(cells []string) todoTable {
	table := todoTable{cells: len(cells), taskCol: -1, names: make([]string, len(cells))}

	for i, name := range cells {
		table.names[i] = strings.TrimSpace(name)

		if strings.EqualFold(table.names[i], "task") {
			table.taskCol = i
		}
	}

	return table
}

// wrapFault names the wrap-placement fault on a row whose cell count
// matches its header: the Task cell must carry a text-bearing wrap and
// no other cell may carry a strike.
func wrapFault(table todoTable, cells []string) (string, bool) {
	for ci, cell := range cells {
		if ci == table.taskCol || !strings.Contains(cell, "~~") {
			continue
		}

		where := "a non-Task cell"
		if ci < len(table.names) && table.names[ci] != "" {
			where = fmt.Sprintf("the %s cell", table.names[ci])
		}

		return fmt.Sprintf(
			"strike wrap sits in %s, not the Task cell",
			where,
		), true
	}

	taskCell := strings.TrimSpace(cells[table.taskCol])

	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(taskCell, "~~"), "~~"))
	if inner != "" {
		return "", false
	}

	return "the Task cell's strike wrap carries no text (shifted strike: the task text must sit inside the wrap)", true
}
