// Package docsrows is the read-only half of the docs-health table-row
// tooling: the completeness gate (formerly `annotate.go check-rows`) and the
// bare-row scan (formerly `annotate.go bare-rows`).
//
// The writer verbs (rows, prose, status-items, repair, …) stay in the
// docs-health skill until the mutator migration is designed and the callers
// are converted; this package deliberately contains NO file-writing code.
// The split is by read/write, not by repository: a public binary may judge
// tables, but it may not rewrite a corpus.
//
// Exit codes: 0 clean, 1 findings, 2 usage/environment error.
package docsrows

import (
	"fmt"
	"os"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// Usage is the verb contract, printed by the dispatcher's help surface.
const Usage = `docs-gates rows — table-row completeness and bare-row gates

usage: docs-gates rows [--strict] [--bare] <file>...

  (default)  every table must be complete: no PARTIAL rows and no mixed
             struck tables (the annotate.go check-rows contract)
  --strict   additionally flag a wholly-unstruck UNTOUCHED table (the
             archive-decision gate: a table with no struck or routed row
             was never inventoried)
  --bare     list every BARE data row outside ledger-marked tables instead
             (the archive-time row-level gate)

A table directly below <!-- check-rows:ledger --> is exempt: a ledger
intentionally mixes resolved and open rows.

exit codes: 0 clean, 1 findings, 2 usage/environment error
`

// snippetMax is the finding-line truncation shared with the writer tool.
const snippetMax = 100

// occurrence is one finding line about a file: the line number supplies the
// location, detail the rendered message. Notes (ledger exemptions) are
// informational and never fail the gate.
type occurrence struct {
	line   int
	detail string
}

// fileReport is one file's verdict used for the summary line and exit code.
type fileReport struct {
	name     string
	notes    []occurrence
	findings []occurrence
}

func (r fileReport) dirty() bool { return len(r.findings) > 0 }

// Run is the CLI entry: `rows [--strict] [--bare] <file>...`. Explicit file
// paths, not a repo root: the writer's corpus scoping belongs to the caller,
// and a gate that silently guessed a corpus would be a different contract.
func Run(argv []string) int {
	opts, files, err := parseArgs(argv)
	if err != nil {
		return gatecli.FatalUsage("rows", err)
	}

	reports, readErr := scanFiles(files, opts)
	if readErr != nil {
		gatecli.Errf("FATAL: %v\n", readErr)

		return gatecli.ExitUsage
	}

	printReports(reports)

	return verdict(reports, opts)
}

type options struct {
	strict bool
	bare   bool
}

type usageError struct{ msg string }

func (e usageError) Error() string { return e.msg }

// parseArgs splits the two boolean modes off argv. A flag or a missing file
// list is a usage error (exit 2); everything else is a file path.
func parseArgs(argv []string) (options, []string, error) {
	var (
		opts  options
		files []string
	)

	for _, arg := range argv {
		switch arg {
		case "--strict":
			opts.strict = true
		case "--bare":
			opts.bare = true
		default:
			if strings.HasPrefix(arg, "-") {
				return options{}, nil, usageError{"unknown flag: " + arg}
			}

			files = append(files, arg)
		}
	}

	if len(files) == 0 {
		return options{}, nil, usageError{"rows needs at least one file path"}
	}

	return opts, files, nil
}

// scanFiles reads every path and returns one report each. A read failure is
// an environment error, not a finding: an unreadable file was not judged.
func scanFiles(files []string, opts options) ([]fileReport, error) {
	reports := make([]fileReport, 0, len(files))

	for _, path := range files {
		tables, err := scanTables(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}

		reports = append(reports, scanReport(path, tables, opts))
	}

	return reports, nil
}

// scanReport turns parsed tables into the mode's notes and findings.
func scanReport(path string, tables []table, opts options) fileReport {
	name := baseName(path)
	if opts.bare {
		return fileReport{name: name, findings: bareFindings(name, tables)}
	}

	return fileReport{
		name:     name,
		notes:    ledgerNotes(tables),
		findings: completenessFindings(name, tables, opts.strict),
	}
}

// printReports renders notes first, then findings, matching the writer
// tool's ordering so operators (and parity diffs) see the same lines.
func printReports(reports []fileReport) {
	for _, r := range reports {
		for _, n := range r.notes {
			gatecli.Outf("%s: table at line %d %s\n", r.name, n.line, n.detail)
		}

		for _, f := range r.findings {
			gatecli.Outf("%s\n", f.detail)
		}
	}
}

// verdict prints the success summary and returns the exit code. A reporting
// failure on any file makes the run dirty; findings already named the lines.
func verdict(reports []fileReport, opts options) int {
	dirty := 0

	for _, r := range reports {
		if r.dirty() {
			dirty++
		}
	}

	mode := "rows"

	if opts.bare {
		mode = "rows --bare"
	}

	if dirty == 0 {
		gatecli.Outf("%s: %d file(s) clean\n", mode, len(reports))

		return gatecli.ExitClean
	}

	gatecli.Errf("%s: %d/%d file(s) incomplete\n", mode, dirty, len(reports))

	return gatecli.ExitFindings
}

// baseName is the display name of a corpus file (the writer's convention:
// findings name the file, not the path).
func baseName(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}

	return path
}

// readLines returns a file's lines without their terminators.
func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSuffix(line, "\r")
	}

	return lines, nil
}

// truncate is the writer's 100-byte finding snippet.
func truncate(s string) string {
	if len(s) > snippetMax {
		return s[:snippetMax]
	}

	return s
}
