package docsstrike

import (
	"fmt"
	"regexp"
	"strings"
)

// sepRowRe matches a markdown table separator row (`| --- | --- |`).
var sepRowRe = regexp.MustCompile(`^\|[\s:|-]+\|\s*$`)

// strikeError is the package's static error type: the parser and the
// strike planner report every refusal as *strikeError (user-facing CLI
// text; no errors.Is consumers by design, so no Unwrap).
type strikeError struct {
	msg string
}

func (e *strikeError) Error() string { return e.msg }

// Strike is one planned row replacement: the 1-based line number in the
// file, the original line, and its struck replacement.
type Strike struct {
	LineNo int
	Old    string
	New    string
}

// tableShape is one table's header-derived layout: the row cell count
// and the split-on-"|" indexes of the named columns (-1 = absent).
type tableShape struct {
	cells     int
	taskCol   int
	statusCol int
	evidence  int
}

// shapeOf derives the row contract from the table's OWN header row: the
// nearest `| --- |` separator at/above the matched row, whose line
// above is the header. Position constants cannot pin a fleet layout
// (#6): TODO_LISTs ship as `# Task Evidence Status`, `Task Status
// Priority Effort Evidence`, `# Task Status Impact Effort Evidence` and
// even header-only-Task shapes — only the header names the cells.
func shapeOf(lines []string, idx int) (tableShape, *strikeError) {
	for i := idx; i > 0; i-- {
		if !strings.HasPrefix(lines[i], "|") {
			break // left the table without crossing a separator
		}

		if !sepRowRe.MatchString(lines[i]) {
			continue
		}

		header := lines[i-1]
		if !strings.HasPrefix(header, "|") {
			break
		}

		return mapHeader(strings.Split(header, "|"))
	}

	return tableShape{}, &strikeError{msg: fmt.Sprintf(
		"row %d sits in no table with a header row (a | --- | separator above a header line) — the layout cannot be derived",
		idx+1,
	)}
}

// mapHeader names the columns by their header cells: Task is required,
// Status and Evidence are optional and stay -1 when the table has no
// such column.
func mapHeader(cells []string) (tableShape, *strikeError) {
	shape := tableShape{cells: len(cells), taskCol: -1, statusCol: -1, evidence: -1}

	for i, name := range cells {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "task":
			shape.taskCol = i
		case "status":
			shape.statusCol = i
		case "evidence":
			shape.evidence = i
		}
	}

	if shape.taskCol < 1 {
		return tableShape{}, &strikeError{
			msg: "the table's header carries no Task column — refusing to guess which cell holds the task",
		}
	}

	return shape, nil
}

// PlanStrike locates the single row matching opts.Row and builds its
// struck replacement without touching anything else in the content.
// Task, Status and Evidence cells are located BY HEADER NAME (#6), so
// every fleet TODO_LIST shape strikes correctly; a table without a
// Status column strikes only the task cell.
func PlanStrike(content string, opts Options) (Strike, *strikeError) {
	lines := strings.Split(content, "\n")

	idx, findErr := uniqueRowMatching(lines, opts.Row)
	if findErr != nil {
		return Strike{}, findErr
	}

	shape, shapeErr := shapeOf(lines, idx)
	if shapeErr != nil {
		return Strike{}, shapeErr
	}

	cells := strings.Split(lines[idx], "|")
	if len(cells) != shape.cells {
		return Strike{}, &strikeError{msg: fmt.Sprintf(
			"row %d has %d cells, the table's header has %d — refusing to guess the layout",
			idx+1,
			len(cells),
			shape.cells,
		)}
	}

	if strikeErr := applyStrike(cells, shape, idx, opts); strikeErr != nil {
		return Strike{}, strikeErr
	}

	return Strike{LineNo: idx + 1, Old: lines[idx], New: strings.Join(cells, "|")}, nil
}

// uniqueRowMatching returns the one table row containing the selector,
// refusing both the none and the ambiguous case.
func uniqueRowMatching(lines []string, selector string) (int, *strikeError) {
	var matches []int

	for i, line := range lines {
		if strings.HasPrefix(line, "|") && !sepRowRe.MatchString(line) && strings.Contains(line, selector) {
			matches = append(matches, i)
		}
	}

	switch len(matches) {
	case 0:
		return 0, &strikeError{msg: fmt.Sprintf("no table row contains the --row selector %q", selector)}
	case 1: // the only acceptable case
	default:
		return 0, &strikeError{msg: fmt.Sprintf(
			"--row selector %q is ambiguous: %d rows match — narrow it to one row",
			selector,
			len(matches),
		)}
	}

	return matches[0], nil
}

// applyStrike rewrites the located row's cells in place: the task cell
// carries the strike verdict, the Status cell flips when the table has
// one, and --evidence lands in the Evidence column or is refused.
func applyStrike(cells []string, shape tableShape, idx int, opts Options) *strikeError {
	task := strings.TrimSpace(cells[shape.taskCol])
	if task == "" {
		return &strikeError{msg: fmt.Sprintf("row %d has an empty task cell — not a strikeable row", idx+1)}
	}

	if strings.Contains(task, "~~") {
		return &strikeError{
			msg: fmt.Sprintf("row %d already carries a strike marker — never double-strike", idx+1),
		}
	}

	cells[shape.taskCol] = " ~~" + task + "~~ done (" + opts.Verdict + ") "

	if shape.statusCol >= 0 {
		cells[shape.statusCol] = " " + opts.Status + " "
	}

	if opts.Evidence != "" {
		if shape.evidence < 0 {
			return &strikeError{
				msg: "the table's header carries no Evidence column but --evidence was given — nowhere to write it",
			}
		}

		cells[shape.evidence] = " " + opts.Evidence + " "
	}

	return nil
}
