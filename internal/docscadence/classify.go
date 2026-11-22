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
			if c := strings.TrimSpace(cell); IsDoneCell(c) {
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

// IsDoneCell reports whether a table cell is a DONE status cell. The
// canonical docs-strike shape is `✅ \`DONE\“, but strike rituals also
// land dated verdicts (`✅ DONE 2026-11-04`, `DONE — evidence …`): a
// cell that STARTS with a DONE marker counts, while task text merely
// containing the word ("the switch IS DONE") never occupies a whole
// cell and stays excluded.
//
// Exported since 2026-11-22 (#10): the purge collector consults the
// same predicate when guarding struck rows — one definition of "done
// status cell" keeps cadence counting and purge collection aligned.
func IsDoneCell(cell string) bool {
	return cell == "✅ `DONE`" || cell == "DONE" ||
		strings.HasPrefix(cell, "✅ DONE") || strings.HasPrefix(cell, "DONE ")
}

// IsOpenCell reports whether a table cell is an OPEN status cell: the
// TODO/BLOCKED/IN_PROGRESS markers the fleet's TODO_LISTs and status
// reports share, bare or emoji/backtick-wrapped (`🔴 \`TODO\“,
// `🟡 \`IN_PROGRESS\“), optionally with a trailing note ("TODO — blocked
// upstream"). Cell-scoped for the same reason as IsDoneCell: task text
// merely CONTAINING "todo" never occupies a whole cell. The purge
// collector refuses to collect a strike sitting on a row whose cell
// reads open (#10: the io-surface purge deleted six open rows that
// carried stray ~~ edit noise).
func IsOpenCell(cell string) bool {
	c := strings.TrimSpace(cell)

	for _, marker := range []string{"🔴", "🟡", "🔵"} {
		c = strings.TrimPrefix(c, marker)
	}

	c = strings.TrimSpace(strings.Trim(c, "`"))

	return c == "TODO" || c == "BLOCKED" || c == "IN_PROGRESS" ||
		strings.HasPrefix(c, "TODO ") || strings.HasPrefix(c, "BLOCKED ") ||
		strings.HasPrefix(c, "IN_PROGRESS ")
}
