package docsrows

import "fmt"

// ledgerNotes names every ledger-exempt table. Informational: an exemption
// is a documented decision, not a finding.
func ledgerNotes(tables []table) []occurrence {
	var notes []occurrence

	for _, t := range tables {
		if t.ledger {
			notes = append(notes, occurrence{
				line:   t.headerLine,
				detail: "LEDGER-EXEMPT (check-rows:ledger marker)",
			})
		}
	}

	return notes
}

// rowCounts is the per-table tally the completeness rules are written
// against. Partial rows carry their text so the finding can quote the line.
type rowCounts struct {
	struck  int
	routed  int
	partial []row
}

// countRows tallies one table's rows. The explicit CLEAN arm is deliberate:
// an exhaustive switch keeps a future fifth state from silently counting as
// open.
func countRows(rows []row) rowCounts {
	var c rowCounts

	for _, r := range rows {
		switch r.state {
		case stateStruck:
			c.struck++
		case stateRouted:
			c.routed++
		case statePartial:
			c.partial = append(c.partial, r)
		case stateClean:
			// CLEAN rows are named only when the table is mixed.
		}
	}

	return c
}

// mixed reports the half-annotated shape: something is struck, but some rows
// are neither struck nor routed. a mixed table is incomplete even when no
// single row is PARTIAL.
func (c rowCounts) mixed(total int) bool {
	return c.struck > 0 && c.struck+c.routed < total
}

// completenessFindings is the check-rows contract: a table is incomplete
// when some rows are struck and others are not, or when any row is partially
// struck. Under --strict a table with NOT ONE struck or routed row is
// UNTOUCHED — it was never inventoried.
func completenessFindings(name string, tables []table, strict bool) []occurrence {
	var findings []occurrence

	for _, t := range tables {
		if t.ledger {
			continue
		}

		findings = append(findings, tableFindings(name, t, strict)...)
	}

	return findings
}

// tableFindings is one table's verdict: the UNTOUCHED arm (strict only),
// the INCOMPLETE arm, and the per-row evidence lines.
func tableFindings(name string, t table, strict bool) []occurrence {
	counts := countRows(t.rows)

	var findings []occurrence

	if strict && counts.struck == 0 && counts.routed == 0 {
		findings = append(findings, occurrence{
			line: t.headerLine,
			detail: fmt.Sprintf(
				"%s: table at line %d UNTOUCHED (0/%d rows struck)",
				name, t.headerLine, len(t.rows),
			),
		})
	}

	mixed := counts.mixed(len(t.rows))
	if len(counts.partial) == 0 && !mixed {
		return findings
	}

	findings = append(findings, occurrence{
		line: t.headerLine,
		detail: fmt.Sprintf(
			"%s: table at line %d INCOMPLETE (%d/%d rows struck)",
			name, t.headerLine, counts.struck, len(t.rows),
		),
	})

	for _, r := range counts.partial {
		findings = append(findings, occurrence{
			line:   r.no,
			detail: fmt.Sprintf("  line %d: PARTIAL row: %s", r.no, truncate(r.text)),
		})
	}

	if mixed && len(counts.partial) == 0 {
		findings = append(findings, mixedRowFindings(t)...)
	}

	return findings
}

// mixedRowFindings names the non-resolved rows of a mixed table. Kept apart
// so tableFindings stays under the cyclop budget.
func mixedRowFindings(t table) []occurrence {
	var findings []occurrence

	for _, r := range t.rows {
		if r.state == stateStruck || r.state == stateRouted {
			continue
		}

		findings = append(findings, occurrence{
			line:   r.no,
			detail: fmt.Sprintf("  line %d: %s row in a struck table", r.no, r.state),
		})
	}

	return findings
}

// bareFindings lists every CLEAN data row outside ledger-marked tables. A
// ledger's open rows are the point, so it is skipped entirely; PARTIAL rows
// are the completeness gate's jurisdiction and are not repeated here.
func bareFindings(name string, tables []table) []occurrence {
	var findings []occurrence

	for _, t := range tables {
		if t.ledger {
			continue
		}

		for _, r := range t.rows {
			if r.state != stateClean {
				continue
			}

			findings = append(findings, occurrence{
				line:   r.no,
				detail: fmt.Sprintf("%s:%d: BARE row: %s", name, r.no, truncate(r.text)),
			})
		}
	}

	return findings
}
