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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

Exit codes: 0 purge planned/applied (0 struck rows is a clean no-op),
1 the pre-flight gate refused — the entry would flip historical items to
DRIFT (reword or resolve the drift first), 2 usage or environment error.
`

// struckRow is one collected strike: the source line, its position, the
// task text between the first ~~ pair, and the verdict tail after it.
type struckRow struct {
	lineNo  int
	oldLine string
	task    string
	verdict string
}

// Run is the CLI entry: parse, collect, draft, pre-flight, and write
// only under --apply.
func Run(argv []string) int {
	opts, err := ParseArgs(argv)
	if err != nil {
		return gatecli.FatalUsage("purge", err)
	}

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
	changelogPath := filepath.Join(opts.Root, opts.Changelog)

	todoData, err := os.ReadFile(todoPath) //nolint:gosec // G304: the file argument IS the tool's input contract
	if err != nil {
		gatecli.Errf("FATAL: reading %s: %v\n", todoPath, err)

		return gatecli.ExitUsage
	}

	rows := collectStruck(string(todoData))
	if len(rows) == 0 {
		gatecli.Outf("docs-purge: no struck rows in %s — nothing to purge\n", opts.File)

		return gatecli.ExitClean
	}

	bullets := draftBullets(rows)
	printPlan(opts.File, rows, bullets)

	settings := docsfreshness.SettingsFromConfig(cfg)
	if code := docsfreshness.PreFlight(opts.Root, bullets, nil, false, false, settings); code != gatecli.ExitClean {
		if !opts.Apply {
			gatecli.Errln("docs-purge: DRY-RUN — nothing was written; the gate above refuses this entry")
		} else {
			gatecli.Errln("docs-purge: REFUSED — the pre-flight gate rejects this entry; nothing was written")
		}

		return code
	}

	if !opts.Apply {
		gatecli.Outf("docs-purge: DRY-RUN — SAFE to land; re-run with --apply to purge %d row(s)\n", len(rows))

		return gatecli.ExitClean
	}

	changelogData, err := os.ReadFile(changelogPath) //nolint:gosec // G304: the file argument IS the tool's input contract
	if err != nil {
		gatecli.Errf("FATAL: reading %s: %v\n", changelogPath, err)

		return gatecli.ExitUsage
	}

	updatedChangelog, err := insertBullets(string(changelogData), bullets)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	if err := writePreservingMode(todoPath, removeRows(string(todoData), rows)); err != nil {
		gatecli.Errf("FATAL: writing %s: %v\n", todoPath, err)

		return gatecli.ExitUsage
	}

	if err := writePreservingMode(changelogPath, updatedChangelog); err != nil {
		gatecli.Errf("FATAL: writing %s: %v\n", changelogPath, err)

		return gatecli.ExitUsage
	}

	gatecli.Outf("docs-purge: purged %d struck row(s) from %s into %s — run `docs-gates cadence` to verify the backlog\n",
		len(rows), opts.File, opts.Changelog)

	return gatecli.ExitClean
}

// collectStruck returns the struck rows of a TODO file in order: every
// line carrying a ~~…~~ wrap with non-empty task text. Empty struck
// cells (`~~ ~~`, the formatter-stable padding shape) are decoration,
// not strikes, and never collect.
func collectStruck(content string) []struckRow {
	var rows []struckRow

	for i, line := range strings.Split(content, "\n") {
		task, verdict, ok := splitStrike(line)
		if !ok {
			continue
		}

		rows = append(rows, struckRow{lineNo: i + 1, oldLine: line, task: task, verdict: verdict})
	}

	return rows
}

// splitStrike extracts the task text between the first ~~ pair and the
// verdict tail after it. A line without a wrap, or one whose wrap is
// empty on both sides, is not a strike.
func splitStrike(line string) (task, verdict string, ok bool) {
	first := strings.Index(line, "~~")
	if first < 0 {
		return "", "", false
	}

	second := strings.Index(line[first+2:], "~~")
	if second < 0 {
		return "", "", false
	}

	task = line[first+2 : first+2+second]
	verdict = strings.TrimSpace(line[first+2+second+2:])

	if strings.TrimSpace(task) == "" && verdict == "" {
		return "", "", false
	}

	return task, verdict, true
}

// draftBullets renders one CHANGELOG bullet per struck row: the unwrapped
// task text and the verdict, dated, in the prose shape the manual purges
// established.
func draftBullets(rows []struckRow) []string {
	date := time.Now().Format("2006-01-02")
	bullets := make([]string, 0, len(rows))

	for _, row := range rows {
		entry := fmt.Sprintf("- TODO_LIST purge (%s): %s", date, strings.TrimSpace(row.task))
		if row.verdict != "" {
			entry += " — " + row.verdict
		}

		bullets = append(bullets, entry)
	}

	return bullets
}

// printPlan writes the dry-run face: one old→new pair per row.
func printPlan(file string, rows []struckRow, bullets []string) {
	for i, row := range rows {
		gatecli.Outf("--- %s:%d\n%s\n", file, row.lineNo, row.oldLine)
		gatecli.Outf("+++ %s\n%s\n", defaultChangelog, bullets[i])
	}
}

// removeRows drops the struck lines from the TODO content.
func removeRows(content string, rows []struckRow) string {
	struck := make(map[int]bool, len(rows))

	lines := strings.Split(content, "\n")
	for _, row := range rows {
		struck[row.lineNo-1] = true
	}

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
// refusal, not a guess.
func insertBullets(changelog string, bullets []string) (string, error) {
	lines := strings.Split(changelog, "\n")

	unreleased := -1
	changed := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == unreleasedHeading {
			unreleased = i
		}
	}

	if unreleased < 0 {
		return "", fmt.Errorf("%s has no %s heading — refusing to guess an insertion point", //nolint:err113 // user-facing CLI refusal; no errors.Is consumers by design
			defaultChangelog, unreleasedHeading)
	}

	for i := unreleased + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			break
		}

		if strings.TrimSpace(lines[i]) == changedHeading {
			changed = i

			break
		}
	}

	if changed < 0 {
		return "", fmt.Errorf("%s has no %s section under [Unreleased] — refusing to guess an insertion point", //nolint:err113 // user-facing CLI refusal; no errors.Is consumers by design
			defaultChangelog, changedHeading)
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

// writePreservingMode writes content with the file's existing permission
// bits (a fresh write gets the default file mode).
func writePreservingMode(path, content string) error {
	mode := os.FileMode(defaultFileMode)

	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}

	return os.WriteFile(path, []byte(content), mode) //nolint:gosec // G306, G703: the mode IS the source file's; the path argument IS the tool's input contract
}
