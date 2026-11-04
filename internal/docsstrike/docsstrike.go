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
	"os"
	"path/filepath"
	"strings"

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

// Run is the CLI entry: parse, plan, print the old→new pair, and write
// only under --apply.
func Run(argv []string) int {
	opts, err := ParseArgs(argv)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	cfg, err := gatecli.LoadConfig(opts.Root)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
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
