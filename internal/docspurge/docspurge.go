// Package docspurge implements the `docs-gates purge` verb: collect the
// struck TODO_LIST rows into CHANGELOG purge bullets as ONE gated
// command — the manual ritual was strike-collect → compose → pre-flight
// → append → delete → battery, eight error-prone steps that shipped two
// same-day purges on 2026-11-06. The tool always runs the freshness
// pre-flight over the drafted bullets (landing them is a living-docs
// edit like any other) and writes only on SAFE; without --apply it is a
// dry-run that prints the plan.
//
// Exit codes: 0 = purge planned/applied (0 struck rows is a clean
// no-op), 1 = the pre-flight refused (the entry would flip historical
// items — reword or resolve the drift first), 2 = usage/environment
// error.
package docspurge

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/docsfreshness"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// defaultChangelog is the CHANGELOG file name relative to the repo root;
// --changelog overrides it. The config schema carries no changelog key —
// the purge entry target has been CHANGELOG.md in every adopting repo.
const defaultChangelog = "CHANGELOG.md"

// unreleasedHeading and changedHeading are the CHANGELOG anchors the
// entry inserts under: the bullets join the tail of the [Unreleased]
// section's Changed list. A CHANGELOG without both anchors is refused —
// guessing an insertion point in an unknown shape corrupts history.
const (
	unreleasedHeading = "## [Unreleased]"
	changedHeading    = "### Changed"
)

// Usage is the verb's help text; usage_test.go pins it against the
// parser's accepted flag vocabulary, both directions.
const Usage = `docs-purge — collect struck TODO_LIST rows into a CHANGELOG purge entry (dry-run by default)

usage: docs-purge [repo-root] [--file TODO_LIST.md] [--changelog CHANGELOG.md] [--apply]

  --file       the TODO file to collect struck rows from
               (default: the config's todo_file)
  --changelog  the file the purge entry lands in
               (default: CHANGELOG.md at the root)
  --apply      write the purge: draft bullets → pre-flight gate →
               CHANGELOG entry + row removal

The drafted bullets ALWAYS pass the docs-freshness pre-flight first
(landing them is a living-docs edit like any other); nothing is written
unless the gate answers SAFE.

The collector never trusts a bare ~~~: a struck table row whose status
cell still reads OPEN (TODO/BLOCKED/IN_PROGRESS, emoji-wrapped or not)
is REFUSED — stray tilde noise on open work is edit noise, not done
work (#10: the io-surface purge deleted six open rows that way) — and
a struck row with no recognizable DONE cell is collected with a WARN.

Exit codes: 0 purge planned/applied (0 struck rows is a clean no-op),
1 the pre-flight gate refused — the entry would flip historical items to
DRIFT (reword or resolve the drift first), 2 usage or environment error
(a refused open-row strike is an environment error: fix the row).
`

// struckRow is one collected strike: the source line, its position, the
// task text between the first ~~ pair, and the verdict tail after it.
type struckRow struct {
	lineNo  int
	oldLine string
	task    string
	verdict string
}

// Run is the CLI entry: parse, then execute.
func Run(argv []string) int {
	opts, err := ParseArgs(argv)
	if err != nil {
		return gatecli.FatalUsage("purge", err)
	}

	return execute(opts)
}

// execute carries the purge flow: load config, collect the struck rows,
// draft the entry, gate it through the freshness pre-flight, and write
// only under --apply.
func execute(opts Options) int {
	cfg, err := gatecli.LoadConfig(opts.Root)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	if opts.File == "" {
		opts.File = cfg.TodoFile
	}

	if opts.Changelog == "" {
		opts.Changelog = defaultChangelog
	}

	todoPath := filepath.Join(opts.Root, opts.File)

	todoData, err := os.ReadFile(todoPath)
	if err != nil {
		gatecli.Errf("FATAL: reading %s: %v\n", todoPath, err)

		return gatecli.ExitUsage
	}

	rows, guards := collectStruck(string(todoData))

	refused := false

	for _, g := range guards {
		if g.refuse {
			refused = true

			gatecli.Errf(
				"docs-purge: REFUSED %s:%d carries a strike but the row status still reads OPEN (%q) — stray ~~ noise on open work is not done work; remove the strike or complete the row first\n",
				opts.File,
				g.lineNo,
				g.cell,
			)

			continue
		}

		gatecli.Outf(
			"WARN %s:%d is struck but carries no DONE status cell — collecting anyway; verify the row is really done\n",
			opts.File, g.lineNo,
		)
	}

	if refused {
		gatecli.Errln("docs-purge: nothing was written — fix the refused row(s) above and re-run")

		return gatecli.ExitUsage
	}

	if len(rows) == 0 {
		gatecli.Outf("docs-purge: no struck rows in %s — nothing to purge\n", opts.File)

		return gatecli.ExitClean
	}

	bullets := draftBullets(rows)
	printPlan(opts.File, opts.Changelog, rows, bullets)

	settings := docsfreshness.SettingsFromConfig(cfg)
	if code := docsfreshness.PreFlight(opts.Root, bullets, nil, false, false, settings); code != gatecli.ExitClean {
		return gateRefused(opts.Apply, code)
	}

	if !opts.Apply {
		gatecli.Outf("docs-purge: DRY-RUN — SAFE to land; re-run with --apply to purge %d row(s)\n", len(rows))

		return gatecli.ExitClean
	}

	return applyPurge(opts, todoPath, string(todoData), rows, bullets)
}

// gateRefused reports the pre-flight refusal on both faces and surfaces
// the gate's own exit code (1 UNSAFE, 2 the on-disk refusal class).
func gateRefused(applied bool, code int) int {
	if !applied {
		gatecli.Errln("docs-purge: DRY-RUN — nothing was written; the gate above refuses this entry")
	} else {
		gatecli.Errln("docs-purge: REFUSED — the pre-flight gate rejects this entry; nothing was written")
	}

	return code
}

// applyPurge writes both faces: the struck rows leave the TODO file, the
// bullets join the [Unreleased] Changed section.
func applyPurge(opts Options, todoPath, todoData string, rows []struckRow, bullets []string) int {
	return applyPurgeWith(func(path, content string) error {
		return gatecli.WritePreservingMode(path, content, defaultFileMode)
	}, opts, todoPath, todoData, rows, bullets)
}

// applyPurgeWith is the write-order core behind applyPurge, taking the
// writer so tests can watch the order and fail a leg. The CHANGELOG is
// written FIRST and the TODO file SECOND: a crash between the two
// writes must leave the board purged but the verdicts already
// preserved, never the reverse — deleted rows with the evidence lost
// (the io-surface incident's recovery existed only because the loss was
// noticed).
func applyPurgeWith(
	write func(path, content string) error,
	opts Options,
	todoPath, todoData string,
	rows []struckRow,
	bullets []string,
) int {
	changelogPath := filepath.Join(opts.Root, opts.Changelog)

	changelogData, err := os.ReadFile(changelogPath)
	if err != nil {
		gatecli.Errf("FATAL: reading %s: %v\n", changelogPath, err)

		return gatecli.ExitUsage
	}

	updatedChangelog, err := insertBullets(opts.Changelog, string(changelogData), bullets)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	if err := write(changelogPath, updatedChangelog); err != nil {
		gatecli.Errf("FATAL: writing %s: %v\n", changelogPath, err)

		return gatecli.ExitUsage
	}

	if err := write(todoPath, removeRows(todoData, rows)); err != nil {
		gatecli.Errf("FATAL: writing %s: %v\n", todoPath, err)

		return gatecli.ExitUsage
	}

	gatecli.Outf(
		"docs-purge: purged %d struck row(s) from %s into %s — run `docs-gates cadence` to verify the backlog\n",
		len(rows),
		opts.File,
		opts.Changelog,
	)

	return gatecli.ExitClean
}

// printPlan writes the dry-run face: one old→new pair per row.
func printPlan(todoFile, changelogFile string, rows []struckRow, bullets []string) {
	for i, row := range rows {
		gatecli.Outf("--- %s:%d\n%s\n", todoFile, row.lineNo, row.oldLine)
		gatecli.Outf("+++ %s\n%s\n", changelogFile, bullets[i])
	}
}

// removeRows drops the struck lines from the TODO content.
func removeRows(content string, rows []struckRow) string {
	struck := make(map[int]bool, len(rows))

	for _, row := range rows {
		struck[row.lineNo-1] = true
	}

	lines := strings.Split(content, "\n")

	kept := make([]string, 0, len(lines))

	for i, line := range lines {
		if !struck[i] {
			kept = append(kept, line)
		}
	}

	return strings.Join(kept, "\n")
}

// insertBullets appends the bullets to the tail of the [Unreleased]
// Changed section, right before its next heading. Missing anchors are a
// refusal, not a guess — the message names changelogName (the actual
// target, honoring --changelog) and the nearest same-level heading.
func insertBullets(changelogName, changelog string, bullets []string) (string, error) {
	lines := strings.Split(changelog, "\n")

	unreleased, nearestH2, nearestH3 := locateAnchors(lines)
	if unreleased < 0 {
		return "", newMissingAnchorError(anchorUnreleasedHeading, changelogName, nearestH2)
	}

	changed := locateChanged(lines, unreleased)
	if changed < 0 {
		return "", newMissingAnchorError(anchorChangedSection, changelogName, nearestH3)
	}

	insert := len(lines)

	for i := changed + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "### ") || strings.HasPrefix(lines[i], "## ") {
			insert = i

			break
		}
	}

	block := append([]string{}, bullets...)
	block = append(block, "")

	updated := append([]string{}, lines[:insert]...)
	updated = append(updated, block...)
	updated = append(updated, lines[insert:]...)

	return strings.Join(updated, "\n"), nil
}

// locateAnchors scans the whole file for the [Unreleased] heading and
// the nearest ## / ### heading lines (1-based; 0 when a class is
// absent) — the material for the missing-anchor hints.
func locateAnchors(lines []string) (int, int, int) {
	unreleased, nearestH2, nearestH3 := -1, 0, 0

	for i, line := range lines {
		if strings.HasPrefix(line, "## ") {
			nearestH2 = i + 1
		}

		if strings.HasPrefix(line, "### ") {
			nearestH3 = i + 1
		}

		if strings.TrimSpace(line) == unreleasedHeading {
			unreleased = i
		}
	}

	return unreleased, nearestH2, nearestH3
}

// locateChanged finds the ### Changed heading constrained to the
// [Unreleased] block (before the block's next ## heading); -1 when the
// section is absent.
func locateChanged(lines []string, unreleased int) int {
	for i := unreleased + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			break
		}

		if strings.TrimSpace(lines[i]) == changedHeading {
			return i
		}
	}

	return -1
}
