// Package docscitations keeps glossary "where it lives" pointers true.
//
// The fleet's mature DOMAIN_LANGUAGE glossaries cite code in two shapes:
// a backticked repo-relative path (`internal/models/schedule.go:27`) or
// the same path bare inside a table cell. Those citations rot silently —
// a refactor moves a file and the glossary points at nothing, with no
// gate the wiser. This gate resolves every backticked code-path citation
// in the corpus against the repo root and fails on missing targets.
// Line numbers are deliberately NOT checked (they drift by design);
// file existence is the contract.
//
// Rollout contract: findings WARN and exit 0 until the repo flips
// citation_enforce (fleet-wide flip rides a TODO row one week after
// ship). Extension list (citation_exts) scopes what counts as a code
// citation; anything else in backticks is prose, not a pointer.
package docscitations

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// Settings is the gate's config-driven knobs. Every field falls back to
// the shipped default when the repo's .docs-gates.json is absent or
// silent, so zero-value Settings never reach the gate logic.
type Settings struct {
	Files   []string // repo-relative corpus to scan
	Ext     []string // dot-prefixed extensions that make a backticked token a citation
	Enforce bool     // false = findings WARN and exit 0; true = FAIL
}

// SettingsFromConfig projects the shared config onto the gate's knobs,
// keeping the shipped defaults wherever the config is silent.
func SettingsFromConfig(cfg config.Config) Settings {
	return Settings{
		Files:   config.OrSlice(cfg.CitationFiles, config.Defaults().CitationFiles),
		Ext:     config.OrSlice(cfg.CitationExts, config.Defaults().CitationExts),
		Enforce: cfg.CitationEnforce,
	}
}

// Usage is the `docs-gates citations --help` text — the verb's contract
// in one place, printed by the dispatcher before the parser runs.
const Usage = `docs-citations — fails a glossary whose code citations point at nothing

usage: docs-citations [repo-root]

No flags: the corpus (citation_files), the code-extension list
(citation_exts) and the rollout switch (citation_enforce) live in the
repo's .docs-gates.json. Every backticked repo-relative path with a
listed extension (optionally :line or :line-line, which are ignored) is
resolved against the repo root; a missing target is a finding — WARN +
exit 0 while citation_enforce is false, FAIL + exit 1 once flipped.
URLs, absolute paths, globs and env-style tokens are prose, not
citations. A missing corpus file is a SKIP note, not a finding.

Exit codes: 0 clean or warn-mode findings, 1 enforced citation findings, 2 usage error.
`

// Run is the CLI entry (repo-root positional).
func Run(args []string) int {
	flags, root, err := gatecli.Parse(args)
	if err != nil {
		return gatecli.FatalUsage("citations", err)
	}

	_ = flags // no flags; the positional root is the only input

	cfg, err := gatecli.LoadConfig(root)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	settings := SettingsFromConfig(cfg)

	findings, skips, cited, err := Scan(root, settings)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	for _, skip := range skips {
		gatecli.Outf("SKIP %s\n", skip)
	}

	for _, finding := range findings {
		if settings.Enforce {
			gatecli.Outf(
				"FAIL %s cites %s — target does not exist under the repo root\n",
				finding.File,
				finding.Citation,
			)
		} else {
			gatecli.Outf(
				"WARN %s cites %s — target does not exist under the repo root (citation_enforce=false, warn mode)\n",
				finding.File,
				finding.Citation,
			)
		}
	}

	if len(findings) == 0 {
		gatecli.Outf("OK %d citation(s) across %d file(s) all resolve\n", cited, len(settings.Files))

		return gatecli.ExitClean
	}

	if settings.Enforce {
		return gatecli.ExitFindings
	}

	return gatecli.ExitClean
}

// Finding is one dangling citation: the citing file and the citation
// token as written.
type Finding struct {
	File     string
	Citation string
}

// backtickRe pulls every backticked token; classification happens after.
var backtickRe = regexp.MustCompile("`([^`\n]+)`")

// Citation extracts the citations from one corpus file's content:
// every backticked token that is a repo-relative path with a listed
// extension, optionally carrying a :line or :line-line suffix (dropped
// for resolution, kept in the token for the verdict line).
func Citation(content string, exts []string) []string {
	var citations []string

	for _, token := range backtickRe.FindAllStringSubmatch(content, -1) {
		path, ok := classify(token[1], exts)
		if !ok {
			continue
		}

		if !slices.Contains(citations, path) {
			citations = append(citations, path)
		}
	}

	return citations
}

// classify decides whether a backticked token is a code citation and
// returns its path form (suffix stripped). Prose shapes — URLs, leading
// slash or dot, globs, spaces, env-dollar tokens, missing listed
// extension — are not citations.
func classify(token string, exts []string) (string, bool) {
	if strings.ContainsAny(token, " $*?<>|\"'") || strings.Contains(token, "://") {
		return "", false
	}

	path := token

	if i := strings.Index(path, ":"); i >= 0 {
		path = path[:i]
	}

	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, ".") {
		return "", false
	}

	if !slices.Contains(exts, filepath.Ext(path)) {
		return "", false
	}

	return path, true
}

// Scan resolves the corpus's citations against root: findings, skip
// notes, total unique citations cited.
func Scan(root string, settings Settings) ([]Finding, []string, int, error) {
	var findings []Finding

	var skips []string

	cited := 0

	for _, rel := range settings.Files {
		body, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			if os.IsNotExist(err) {
				skips = append(skips, rel+" (missing)")

				continue
			}

			return nil, nil, 0, fmt.Errorf("reading %s: %w", rel, err)
		}

		for _, citation := range Citation(string(body), settings.Ext) {
			cited++

			if _, err := os.Stat(filepath.Join(root, citation)); err != nil {
				findings = append(findings, Finding{File: rel, Citation: citation})
			}
		}
	}

	return findings, skips, cited, nil
}
