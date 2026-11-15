package docsrows

import (
	"regexp"
	"strings"
)

// ledgerMarker exempts exactly the NEXT table from the completeness gate.
// A ledger table intentionally mixes struck-resolved and open rows (per-row
// inline verdicts are the point of a ledger). The marker line sits above the
// table; blank lines in between are allowed; any other non-blank non-table
// line cancels it.
const ledgerMarker = "<!-- check-rows:ledger -->"

// rowState is a data row's annotation state. The four values are the writer
// tool's vocabulary, kept byte-identical so a parity diff needs no mapping.
type rowState string

const (
	stateClean   rowState = "CLEAN"
	stateStruck  rowState = "STRUCK"
	statePartial rowState = "PARTIAL"
	stateRouted  rowState = "ROUTED"
)

// row is one data row of a scanned table.
type row struct {
	no    int
	text  string
	state rowState
}

// table is one markdown table: the header line's number, whether a
// check-rows:ledger marker exempted it, and every data row with its
// classification. Separator rows are excluded.
type table struct {
	headerLine int
	ledger     bool
	rows       []row
}

//nolint:gochecknoglobals // compiled-once static patterns, immutable data tables
var (
	reTableSepCell     = regexp.MustCompile(`^:?-+:?$`)
	reDoubledSpanCheck = regexp.MustCompile("``[^`]+``")
	reSingleSpan       = regexp.MustCompile("`[^`]*`")
	// reRoutedArrow matches the routed-verdict convention: the writer's
	// r-marker renders `**→ verdict**`, and hand-routed rows carry a plain
	// `→ routed` arrow. A routed row is INTENTIONALLY unstruck — the
	// completeness gate must count it as resolved.
	reRoutedArrow = regexp.MustCompile(`\*\*→[^*]+\*\*`)
)

// outsideCodeSpansCheck removes inline code spans so a `~~` inside them is
// not mistaken for a strike (doubled-backtick spans first, then singles).
func outsideCodeSpansCheck(text string) string {
	return reSingleSpan.ReplaceAllString(reDoubledSpanCheck.ReplaceAllString(text, ""), "")
}

// splitCells splits a table row into cells on UNESCAPED pipes: `\|` inside a
// cell (escaped-pipe code spans showing table syntax) must not become a cell
// boundary. The sentinel is restored after the split so cell text stays
// byte-faithful.
func splitCells(line string) []string {
	const sentinel = "\x00"

	masked := strings.ReplaceAll(strings.TrimSpace(line), `\|`, sentinel)
	parts := strings.Split(strings.Trim(masked, "|"), "|")

	for i, p := range parts {
		parts[i] = strings.TrimSpace(strings.ReplaceAll(p, sentinel, `\|`))
	}

	return parts
}

// isSeparator reports whether a row is a markdown table separator
// (`| --- | :---: |`).
func isSeparator(line string) bool {
	for _, c := range splitCells(line) {
		if !reTableSepCell.MatchString(c) {
			return false
		}
	}

	return true
}

// isRouted reports whether a row carries a routed verdict, which resolves an
// item without striking it.
func isRouted(line string) bool {
	stripped := outsideCodeSpansCheck(strings.TrimSpace(line))

	return strings.Contains(stripped, "→ routed") || reRoutedArrow.MatchString(stripped)
}

// classifyRow is the row-state machine: every cell struck means the row is
// resolved; a routed verdict also resolves it; no strike anywhere leaves it
// CLEAN; anything between is PARTIAL (the formatter-damage shape).
func classifyRow(line string) rowState {
	states := make([]bool, 0, 4)

	for _, c := range splitCells(line) {
		states = append(states, strings.Contains(outsideCodeSpansCheck(c), "~~"))
	}

	all, any := true, false

	for _, s := range states {
		all = all && s
		any = any || s
	}

	switch {
	case all:
		return stateStruck
	case isRouted(line):
		return stateRouted
	case !any:
		return stateClean
	default:
		return statePartial
	}
}

// scanTables walks a file's markdown tables with the exact block/ledger
// adjacency semantics the writer has always used: a
// <!-- check-rows:ledger --> line exempts the NEXT table only, blank lines
// between marker and table are allowed, and any other non-blank line resets
// the pending exemption. Both modes (completeness, bare) consume this so
// their ledger semantics cannot drift apart.
func scanTables(path string) ([]table, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	var (
		tables        []table
		pendingLedger bool
		blockLedger   bool
		block         []row
	)

	flush := func() {
		if len(block) == 0 {
			return
		}

		wasLedger := blockLedger
		blockLedger = false

		data := make([]row, 0, len(block))

		for _, r := range block {
			if !isSeparator(r.text) {
				data = append(data, r)
			}
		}

		block = nil

		if len(data) < 2 {
			return
		}

		t := table{headerLine: data[0].no, ledger: wasLedger}

		for _, r := range data[1:] {
			r.state = classifyRow(r.text)
			t.rows = append(t.rows, r)
		}

		tables = append(tables, t)
	}

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimLeft(line, " \t"), "|") {
			if len(block) == 0 && pendingLedger {
				blockLedger = true
				pendingLedger = false
			}

			block = append(block, row{no: i + 1, text: strings.TrimRight(line, "\r")})

			continue
		}

		flush()

		switch t := strings.TrimSpace(line); {
		case t == ledgerMarker:
			pendingLedger = true
		case t != "":
			pendingLedger = false
		}
	}

	flush()

	return tables, nil
}
