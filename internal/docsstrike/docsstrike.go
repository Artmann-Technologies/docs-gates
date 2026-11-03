// Package docsstrike is the TODO_LIST row-strike helper (2026-10-08).
//
// The 2026-10-07 full-execution sessions struck ~20 rows with ad-hoc
// perl/python cell-split scripts kept in /tmp — every strike needed the
// same surgery (wrap the task cell in ~~…~~ + a `done (…)` verdict, flip
// the status cell, PRESERVE Impact/Effort, optionally refresh Evidence)
// and every hand-rolled variant risked the documented failure modes: a
// loose regex eating a neighboring row, a strike that renumbered the
// list tail, or a lost column. This package is that pattern as a repo
// tool, so a strike is one reviewed command instead of a fresh script.
//
// Contract (the TODO_LIST 5-column row shape only):
//
//		| Task | Status | Impact | Effort | Evidence |
//
//	  - The selector (--row) must match EXACTLY ONE table row in the file;
//	    zero or multiple matches are errors, never a guess.
//	  - A row that already carries ~~ is refused (never double-strike).
//	  - Task cell becomes `~~<task>~~ done (<verdict>)`; the `done (`
//	    prefix is the docs-annotate verdict vocabulary, so the gate stays
//	    green by construction.
//	  - Status cell becomes ✅ `DONE` unless --status overrides (e.g. a
//	    `Won't implement` style verdict).
//	  - Impact/Effort cells are preserved BYTE-identically; Evidence is
//	    preserved unless --evidence replaces it.
//	  - Default is DRY-RUN: the old→new row pair is printed and nothing
//	    is written until --apply.
//
// Usage: docs-strike [repo-root] --row <selector> --verdict <text>
//
//	[--evidence <text>] [--status <text>] [--file TODO_LIST.md] [--apply]
//
// Exit 0 = strike planned/applied, 2 = usage or selector error.
package docsstrike

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// DefaultStatus is the status cell a completed row flips to; the ✅ emoji
// + backticked DONE matches every existing DONE row in TODO_LIST.md.
const DefaultStatus = "✅ `DONE`"

// TODOColumns is the TODO_LIST row shape: 5 columns means strings.Split
// on "|" yields exactly 7 cells (two empty edges). A row with any other
// cell count is not a TODO_LIST row and is refused — the helper never
// guesses which cell is the task on a differently-shaped table.
const TODOColumns = 7

// defaultFileMode is the fallback permission when the target's own mode
// cannot be observed (fresh-file creation).
const defaultFileMode = 0o644

// Options carries the parsed CLI contract of Run.
type Options struct {
	Root     string
	File     string
	Row      string
	Verdict  string
	Evidence string
	Status   string
	Apply    bool
}

// isValueFlag reports whether the flag consumes the next argument as its
// value; --apply is the lone boolean. gatecli.Parse is boolean-only, so
// this parser keeps its value-carrying semantics local and hands only
// the root resolution back to gatecli.
func isValueFlag(arg string) bool {
	switch arg {
	case "--row", "--verdict", "--evidence", "--status", "--file":
		return true
	}

	return false
}

// strikeError is the package's static error type: the parser and the
// strike planner report every refusal as *strikeError (user-facing CLI
// text; no errors.Is consumers by design, so no Unwrap).
type strikeError struct {
	msg string
}

func (e *strikeError) Error() string { return e.msg }

// consumeValueFlag consumes one value-carrying flag at argv[i], wiring the
// value into opts, and returns the index of the last consumed argument.
func consumeValueFlag(argv []string, i int, arg string, opts *Options) (int, *strikeError) {
	if i+1 >= len(argv) {
		return i, &strikeError{msg: arg + " needs a value"}
	}

	value := argv[i+1]

	switch arg {
	case "--row":
		opts.Row = value
	case "--verdict":
		opts.Verdict = value
	case "--evidence":
		opts.Evidence = value
	case "--status":
		opts.Status = value
	case "--file":
		opts.File = value
	}

	return i + 1, nil
}

// ParseArgs splits argv into the strike options: value flags
// (--row/--verdict/--evidence/--status/--file), the --apply boolean, and
// an optional positional repo root (resolved by gatecli.ResolveRoot).
func ParseArgs(argv []string) (Options, error) {
	opts := Options{Status: DefaultStatus}

	var positional []string

	for i := 0; i < len(argv); i++ {
		arg := argv[i]

		switch {
		case arg == "--apply":
			opts.Apply = true
		case isValueFlag(arg):
			last, err := consumeValueFlag(argv, i, arg, &opts)
			if err != nil {
				return opts, err
			}

			i = last
		case strings.HasPrefix(arg, "--"):
			return opts, fmt.Errorf( //nolint:err113 // user-facing CLI refusal; no errors.Is consumers by design
				"unknown flag: %s",
				arg,
			)
		default:
			positional = append(positional, arg)
		}
	}

	if opts.Row == "" {
		return opts, errors.New( //nolint:err113 // user-facing CLI refusal; no errors.Is consumers by design
			"--row <selector> is required (a substring of the row to strike)",
		)
	}

	if opts.Verdict == "" {
		return opts, errors.New( //nolint:err113 // user-facing CLI refusal; no errors.Is consumers by design
			"--verdict <text> is required (the done (…) evidence payload)",
		)
	}

	if opener, prefixed := prefixedVerdictOpener(opts.Verdict); prefixed {
		return opts, fmt.Errorf( //nolint:err113 // user-facing CLI refusal; no errors.Is consumers by design
			"the --verdict already starts with the verdict word %q — the tool prepends "+
				"the done (…) wrap itself, so pass the evidence WITHOUT a leading verdict "+
				"marker (a prefixed verdict double-wraps into done (done (…)))",
			opener,
		)
	}

	root, err := gatecli.ResolveRoot(positional)
	if err != nil {
		return opts, fmt.Errorf("resolve repo root: %w", err)
	}

	opts.Root = root

	return opts, nil
}

// Strike is one planned row replacement: the 1-based line number in the
// file, the original line, and its struck replacement.
type Strike struct {
	LineNo int
	Old    string
	New    string
}

// verdictOpeners are the leading words of the docs-health verdict
// vocabulary the tool already wraps the evidence in. A --verdict that
// OPENS with one of them would land as done (done (…)) — the
// double-wrapped wart class (burned 2026-10-12, caught in dry-run).
//
//nolint:gochecknoglobals // the fixed docs-health verdict vocabulary, same class as the contract tables
var verdictOpeners = []string{
	"done",
	"resolved",
	"superseded",
	"decided",
	"won't implement",
	"wont implement",
	"not-do",
	"not do",
}

// prefixedVerdictOpener reports whether the verdict already opens with a
// verdict word (case-insensitive, markdown emphasis tolerated) and
// returns the matched opener for the refusal message.
func prefixedVerdictOpener(verdict string) (string, bool) {
	v := strings.ToLower(strings.TrimSpace(verdict))
	v = strings.TrimLeft(v, "*~`_ ")

	for _, opener := range verdictOpeners {
		if strings.HasPrefix(v, opener) {
			return opener, true
		}
	}

	return "", false
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

// Run is the CLI entry: parse, plan, print the old→new pair, and write
// only under --apply.
func Run(argv []string) int {
	opts, err := ParseArgs(argv)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	cfg, warnings, err := config.Load(opts.Root)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	for _, warning := range warnings {
		gatecli.Errf("WARN: %s\n", warning)
	}

	if opts.File == "" {
		opts.File = cfg.TodoFile
	}

	path := filepath.Join(opts.Root, opts.File)

	data, err := os.ReadFile(
		path,
	)
	if err != nil {
		gatecli.Errf("FATAL: reading %s: %v\n", path, err)

		return gatecli.ExitUsage
	}

	strike, strikeErr := PlanStrike(string(data), opts)
	if strikeErr != nil {
		gatecli.Errf("FATAL: %v\n", strikeErr)

		return gatecli.ExitUsage
	}

	gatecli.Outf("--- %s:%d\n%s\n", opts.File, strike.LineNo, strike.Old)
	gatecli.Outf("+++ %s:%d\n%s\n", opts.File, strike.LineNo, strike.New)

	if !opts.Apply {
		gatecli.Outf("docs-strike: DRY-RUN — re-run with --apply to write %s\n", opts.File)

		return gatecli.ExitClean
	}

	lines := strings.Split(string(data), "\n")
	lines[strike.LineNo-1] = strike.New

	info, err := os.Stat(path)

	mode := os.FileMode(defaultFileMode)
	if err == nil {
		mode = info.Mode().Perm()
	}

	if err := os.WriteFile( //nolint:gosec // G703: the file argument IS the tool's input contract
		path,
		[]byte(strings.Join(lines, "\n")),
		mode,
	); err != nil {
		gatecli.Errf("FATAL: writing %s: %v\n", path, err)

		return gatecli.ExitUsage
	}

	gatecli.Outf("docs-strike: struck %s:%d\n", opts.File, strike.LineNo)

	return gatecli.ExitClean
}
