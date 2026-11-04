package docscadence

import "strings"

// classify counts the completed-row signatures in a TODO_LIST body:
// struck = a table row carrying BOTH a ~~…~~ wrap and a DONE status cell
// (the docs-strike output shape); bareDone = a DONE status cell with no
// wrap on the same line (a half-strike, reported but not failing). The
// DONE check is CELL-scoped: task text legitimately contains the word
// ("the switch IS DONE"), and a line-based substring would miscount.
func classify(body string) (int, int) {
	struck := 0
	bareDone := 0

	for line := range strings.SplitSeq(body, "\n") {
		if !strings.HasPrefix(line, "|") {
			continue // only table rows carry the status cells
		}

		hasStrike := strings.Contains(line, "~~")
		hasDone := false

		for cell := range strings.SplitSeq(line, "|") {
			if c := strings.TrimSpace(cell); isDoneCell(c) {
				hasDone = true

				break
			}
		}

		switch {
		case hasStrike && hasDone:
			struck++
		case !hasStrike && hasDone:
			bareDone++
		}
	}

	return struck, bareDone
}

// isDoneCell reports whether a table cell is a DONE status cell. The
// canonical docs-strike shape is `✅ \`DONE\“, but strike rituals also
// land dated verdicts (`✅ DONE 2026-11-04`, `DONE — evidence …`): a
// cell that STARTS with a DONE marker counts, while task text merely
// containing the word ("the switch IS DONE") never occupies a whole
// cell and stays excluded.
func isDoneCell(cell string) bool {
	return cell == "✅ `DONE`" || cell == "DONE" ||
		strings.HasPrefix(cell, "✅ DONE") || strings.HasPrefix(cell, "DONE ")
}
