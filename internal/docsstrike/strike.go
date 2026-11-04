package docsstrike

import (
	"fmt"
	"strings"
)

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

// PlanStrike locates the single row matching opts.Row and builds its
// struck replacement without touching anything else in the content.
func PlanStrike(content string, opts Options) (Strike, *strikeError) {
	lines := strings.Split(content, "\n")

	var matches []int

	for i, line := range lines {
		if strings.HasPrefix(line, "|") && strings.Contains(line, opts.Row) {
			matches = append(matches, i)
		}
	}

	switch len(matches) {
	case 0:
		return Strike{}, &strikeError{msg: fmt.Sprintf("no table row contains the --row selector %q", opts.Row)}
	case 1: // the only acceptable case
	default:
		return Strike{}, &strikeError{msg: fmt.Sprintf(
			"--row selector %q is ambiguous: %d rows match — narrow it to one row",
			opts.Row,
			len(matches),
		)}
	}

	idx := matches[0]

	cells := strings.Split(lines[idx], "|")
	if len(cells) != TODOColumns {
		return Strike{}, &strikeError{msg: fmt.Sprintf(
			"row %d has %d cells, expected exactly %d (the 5-column TODO_LIST shape) — refusing to guess the layout",
			idx+1,
			len(cells),
			TODOColumns,
		)}
	}

	task := strings.TrimSpace(cells[1])
	if task == "" {
		return Strike{}, &strikeError{msg: fmt.Sprintf("row %d has an empty task cell — not a strikeable row", idx+1)}
	}

	if strings.Contains(task, "~~") {
		return Strike{}, &strikeError{
			msg: fmt.Sprintf("row %d already carries a strike marker — never double-strike", idx+1),
		}
	}

	cells[1] = " ~~" + task + "~~ done (" + opts.Verdict + ") "
	cells[2] = " " + opts.Status + " "

	if opts.Evidence != "" {
		cells[len(cells)-2] = " " + opts.Evidence + " "
	}

	return Strike{LineNo: idx + 1, Old: lines[idx], New: strings.Join(cells, "|")}, nil
}
