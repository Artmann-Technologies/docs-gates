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

// minTableRows is the smallest run of pipe lines that is a real table: a
// header plus at least one data row. Separator rows are dropped before the
// count, so a header+separator+row table arrives here as two rows.
const minTableRows = 2

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
	cells := splitCells(line)
	struckCells := 0

	for _, c := range cells {
		if strings.Contains(outsideCodeSpansCheck(c), "~~") {
			struckCells++
		}
	}

	switch {
	case struckCells == len(cells):
		return stateStruck
	case isRouted(line):
		return stateRouted
	case struckCells == 0:
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

	s := &tableScanner{}

	for i, line := range lines {
		s.consume(line, i+1)
	}

	s.flush()

	return s.tables, nil
}

// tableScanner carries the block/ledger state machine across lines. It
// exists so scanTables stays a loop and the boundary rules stay testable.
type tableScanner struct {
	tables        []table
	pendingLedger bool
	blockLedger   bool
	block         []row
}

// consume feeds one source line: a pipe line joins the open block, anything
// else closes the block and may arm the ledger marker.
func (s *tableScanner) consume(line string, no int) {
	if isTableLine(line) {
		if len(s.block) == 0 && s.pendingLedger {
			s.blockLedger = true
			s.pendingLedger = false
		}

		s.block = append(s.block, row{no: no, text: strings.TrimRight(line, "\r")})

		return
	}

	s.flush()
	s.marker(strings.TrimSpace(line))
}

// marker arms or cancels the NEXT table's ledger exemption.
func (s *tableScanner) marker(text string) {
	switch text {
	case ledgerMarker:
		s.pendingLedger = true
	case "":
		// A blank line between the marker and its table does not cancel it.
	default:
		s.pendingLedger = false
	}
}

// flush closes the open block into a table when it has a header and at
// least one data row after separator rows are dropped.
func (s *tableScanner) flush() {
	if len(s.block) == 0 {
		return
	}

	wasLedger := s.blockLedger
	s.blockLedger = false

	data := make([]row, 0, len(s.block))

	for _, r := range s.block {
		if !isSeparator(r.text) {
			data = append(data, r)
		}
	}

	s.block = nil

	if len(data) < minTableRows {
		return
	}

	t := table{headerLine: data[0].no, ledger: wasLedger}

	for _, r := range data[1:] {
		r.state = classifyRow(r.text)
		t.rows = append(t.rows, r)
	}

	s.tables = append(s.tables, t)
}

// isTableLine reports whether a line starts a (or continues a) table block.
func isTableLine(line string) bool {
	return strings.HasPrefix(strings.TrimLeft(line, " \t"), "|")
}
