// Package docsvocabulary fails a FEATURES.md whose table rows carry
// status tokens the file's own legend never defines.
//
// The 2026-11-22 go-verdict docs audit shipped exactly that gap: a
// "Deliberately absent" table used OUT_OF_SCOPE before any legend
// defined it — caught in self-review, with every gate green (#11). The
// battery verifies doc shape (freshness, annotate, routing, cadence)
// but not vocabulary; this gate closes that hole: every table's Status
// column must only contain tokens the file's own "## Status legend"
// section defines (legend table AND the prose around it — the
// absent-why note counts as a definition), and unused legend entries
// stay informational, never a failure.
//
// Rollout contract (the docsstub doctrine): findings WARN and exit 0
// until the repo flips vocabulary_enforce. A file without a legend
// section, or a missing file, is a SKIP note — repos whose FEATURES.md
// uses prose sections instead of status tables are not this gate's
// problem.
package docsvocabulary

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/docsrows"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// Settings is the gate's config-driven knobs. Every field falls back to
// the shipped default when the repo's .docs-gates.json is absent or
// silent, so zero-value Settings never reach the gate logic.
type Settings struct {
	Files   []string // repo-relative corpus to judge
	Enforce bool     // false = findings WARN and exit 0; true = FAIL
}

// SettingsFromConfig projects the shared config onto the gate's knobs,
// keeping the shipped defaults wherever the config is silent.
func SettingsFromConfig(cfg config.Config) Settings {
	return Settings{
		Files:   config.OrSlice(cfg.VocabularyFiles, config.Defaults().VocabularyFiles),
		Enforce: cfg.VocabularyEnforce,
	}
}

// Usage is the `docs-gates vocabulary --help` text — the verb's
// contract in one place, printed by the dispatcher before any gate
// parser runs.
const Usage = `docs-vocabulary — FEATURES status tokens must come from the file's own legend

usage: docs-vocabulary [repo-root]

No flags: the corpus (vocabulary_files, default FEATURES.md) and the
rollout switch (vocabulary_enforce) live in the repo's .docs-gates.json.
Every table carrying a Status column is judged cell-aware: each
backticked ALL-CAPS token in a status cell must be defined by the
file's own "## Status legend" section (the legend table and the prose
around it both define tokens). An undefined token is a finding — WARN +
exit 0 while vocabulary_enforce is false, FAIL + exit 1 once flipped.
Unused legend entries are informational. A file with no legend section
(or a missing file) is a SKIP note, not a finding.

Exit codes: 0 clean or warn-mode findings, 1 enforced findings, 2 usage error.
`

// legendHeadingRe matches the legend section heading; the level is
// unpinned so both `## Status legend` and deeper nestings resolve.
var legendHeadingRe = regexp.MustCompile(`(?i)^#{1,6}\s*status legend\b`)

// headingRe closes the legend section: any ATX heading ends it.
var headingRe = regexp.MustCompile(`^#{1,6}\s`)

// tokenRe captures one backticked ALL-CAPS token (the status
// vocabulary's shape: `FULLY_FUNCTIONAL`, `OUT_OF_SCOPE`).
var tokenRe = regexp.MustCompile("`(ERROR|[A-Z][A-Z0-9_]{2,})`")

// statusColumn names the header cell whose column carries statuses
// (case-insensitive; "Status/Note" style combined columns count).
var statusColumnRe = regexp.MustCompile(`(?i)^status\b`)

// Finding is one undefined status token: the file, the row's line, the
// token, and the legend section that should have defined it.
type Finding struct {
	Path  string
	Line  int
	Token string
}

// Info is one informational note (an unused legend entry).
type Info struct {
	Path  string
	Token string
}

// Scan judges the corpus and returns the undefined-token findings, the
// unused-legend infos, and the human-readable skip notes.
func Scan(root string, settings Settings) ([]Finding, []Info, []string, error) {
	var findings []Finding

	var infos []Info

	var skips []string

	for _, rel := range settings.Files {
		path := filepath.Join(root, rel)

		body, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				skips = append(skips, rel+" (missing)")

				continue
			}

			return nil, nil, nil, fmt.Errorf("reading %s: %w", rel, err)
		}

		section, spanStart, spanEnd := legendSection(strings.Split(string(body), "\n"))
		if section == "" {
			skips = append(skips, rel+" (no status legend section — not judged)")

			continue
		}

		legend := legendTokens(section)
		used := map[string]bool{}

		tables, err := docsrows.Tables(path)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("scanning %s: %w", rel, err)
		}

		for _, table := range tables {
			// The legend table itself (its Status column IS the definition)
			// never counts as usage — otherwise every legend entry would
			// mark itself used and the informational leg could never fire.
			if table.HeaderLine >= spanStart && table.HeaderLine < spanEnd {
				continue
			}

			cells := docsrows.SplitCells(table.Header)
			statusIdx := -1

			for i, name := range cells {
				if statusColumnRe.MatchString(strings.TrimSpace(name)) {
					statusIdx = i

					break
				}
			}

			if statusIdx < 0 {
				continue
			}

		for _, row := range table.Rows {
			rowCells := docsrows.SplitCells(row.Text)
			if statusIdx >= len(rowCells) {
				continue
			}

			for _, token := range cellTokens(rowCells[statusIdx]) {
				used[token] = true

				if !legend[token] {
					findings = append(findings, Finding{Path: rel, Line: row.Line, Token: token})
				}
			}
			}
		}

		for _, token := range sortedKeys(legend) {
			if !used[token] {
				infos = append(infos, Info{Path: rel, Token: token})
			}
		}
	}

	return findings, infos, skips, nil
}

// legendSection returns the "## Status legend" section: its body and
// its 1-based line span [start, end) — from the heading to the next ATX
// heading. Prose inside the section (the absent-why note) is part of
// the legend — definitions live in tables AND in sentences — and the
// span lets the usage scan skip the definition table itself.
func legendSection(lines []string) (string, int, int) {
	start := -1

	for i, line := range lines {
		if legendHeadingRe.MatchString(line) {
			start = i

			break
		}
	}

	if start < 0 {
		return "", 0, 0
	}

	var section []string

	for i, line := range lines[start+1:] {
		if headingRe.MatchString(line) {
			return strings.Join(section, "\n"), start + 1, start + 1 + i
		}

		section = append(section, line)
	}

	return strings.Join(section, "\n"), start + 1, len(lines) + 1
}

// legendTokens collects every backticked ALL-CAPS token of the legend
// section — the table's status column and the prose around it both
// define tokens.
func legendTokens(section string) map[string]bool {
	tokens := map[string]bool{}

	for _, m := range tokenRe.FindAllStringSubmatch(section, -1) {
		tokens[m[1]] = true
	}

	return tokens
}

// cellTokens collects the backticked ALL-CAPS tokens of one status
// cell (an emoji prefix or a trailing note never hides the token).
func cellTokens(cell string) []string {
	var tokens []string

	for _, m := range tokenRe.FindAllStringSubmatch(cell, -1) {
		tokens = append(tokens, m[1])
	}

	return tokens
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

// Run is the CLI entry (repo-root positional).
func Run(args []string) int {
	flags, root, err := gatecli.Parse(args)
	if err != nil {
		return gatecli.FatalUsage("vocabulary", err)
	}

	_ = flags // no flags; the positional root is the only input

	cfg, err := gatecli.LoadConfig(root)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	settings := SettingsFromConfig(cfg)

	findings, infos, skips, err := Scan(root, settings)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	for _, skip := range skips {
		gatecli.Outf("SKIP %s\n", skip)
	}

	for _, info := range infos {
		gatecli.Outf("INFO %s defines %s but no status column uses it\n", info.Path, info.Token)
	}

	remedy := "define it under \"## Status legend\" or use a defined status"

	for _, finding := range findings {
		if settings.Enforce {
			gatecli.Outf(
				"FAIL %s:%d: status %s is not defined by the file's own legend — %s\n",
				finding.Path, finding.Line, finding.Token, remedy,
			)
		} else {
			gatecli.Outf(
				"WARN %s:%d: status %s is not defined by the file's own legend — %s (vocabulary_enforce=false, warn mode)\n",
				finding.Path, finding.Line, finding.Token, remedy,
			)
		}
	}

	if len(findings) == 0 {
		gatecli.Outf("OK every status token comes from the legend in %d file(s)\n", len(settings.Files))

		return gatecli.ExitClean
	}

	if settings.Enforce {
		return gatecli.ExitFindings
	}

	return gatecli.ExitClean
}
