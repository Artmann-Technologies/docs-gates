package docsrows

import "fmt"

// ledgerNotes names every ledger-exempt table. Informational: an exemption
// is a documented decision, not a finding.
func ledgerNotes(name string, tables []table) []occurrence {
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

		struck, routed := 0, 0

		var partial []row

		for _, r := range t.rows {
			switch r.state {
			case stateStruck:
				struck++
			case stateRouted:
				routed++
			case statePartial:
				partial = append(partial, r)
			}
		}

		if strict && struck == 0 && routed == 0 {
			findings = append(findings, occurrence{
				line: t.headerLine,
				detail: fmt.Sprintf(
					"%s: table at line %d UNTOUCHED (0/%d rows struck)",
					name, t.headerLine, len(t.rows),
				),
			})
		}

		mixed := struck > 0 && struck+routed < len(t.rows)
		if len(partial) == 0 && !mixed {
			continue
		}

		findings = append(findings, occurrence{
			line: t.headerLine,
			detail: fmt.Sprintf(
				"%s: table at line %d INCOMPLETE (%d/%d rows struck)",
				name, t.headerLine, struck, len(t.rows),
			),
		})

		for _, r := range partial {
			findings = append(findings, occurrence{
				line:   r.no,
				detail: fmt.Sprintf("  line %d: PARTIAL row: %s", r.no, truncate(r.text)),
			})
		}

		if mixed && len(partial) == 0 {
			findings = append(findings, mixedRowFindings(t)...)
		}
	}

	return findings
}

// mixedRowFindings names the non-resolved rows of a mixed table. Kept apart
// from completenessFindings so the loop body stays under the cyclop budget.
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
