// Package docsarchive is the cadence ritual's mechanical core as one
// gated command: banner + git mv for fully-resolved status reports.
//
// The ritual (AGENTS.md, paid for twice): audit-all proof, per-report
// inspection, a Live-remainder banner prepended above the title, git mv
// into the archived subdir, battery re-run. This verb owns the two
// mechanical steps; the proof legs stay gates — the banner's claim is
// only honest if `freshness --audit-all`, `routing` and `annotate` ran
// green over the corpus first, and the dry-run face says so.
//
// A report is a candidate iff the same extraction dump-items uses finds
// ZERO open §f items — the harvest obligation is the hard bar. Reports
// with open items print as skips and never move; a report that already
// carries a banner is refused (double-archiving is the corruption
// class safe-strike exists to prevent). Dry-run is the default; only
// --apply writes, and every write rides `git mv` so history keeps the
// rename.
package docsarchive

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/docsfreshness"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// bannerPrefix identifies an already-archived report; the house
// convention opens with the bolded Live-remainder claim.
const bannerPrefix = "> **Live remainder:**"

// Usage is the `docs-gates archive --help` text — the verb's contract
// in one place, printed by the dispatcher before the parser runs.
const Usage = `docs-archive — banner + git-mv fully-resolved reports into the archive (dry-run by default)

usage: docs-archive [repo-root] [--apply]

Scans the report dir (report_dir, default docs/status) for reports
whose §f section carries ZERO open items (the same extraction
dump-items uses). Each candidate gets the Live-remainder banner
prepended (the house convention: a blockquote above the title claiming
the proof legs) and is git-mv'd into the archived subdir
(archived_subdir, default archived).

The banner CLAIMS the proof; the proof is the battery's job: run
freshness --audit-all, routing and annotate green over the corpus
first (the ritual's order), then archive. Reports with open §f items
print as skips and never move; a report that already carries a banner
is refused.

  --apply      write the banners and run the git mv operations

Exit codes: 0 plan printed or applied, 1 unused, 2 usage or environment error.
`

// Run is the CLI entry (repo-root positional, --apply boolean).
func Run(argv []string) int {
	flags, root, err := gatecli.Parse(argv, "--apply")
	if err != nil {
		return gatecli.FatalUsage("archive", err)
	}

	cfg, err := gatecli.LoadConfig(root)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	return execute(root, cfg, flags["--apply"], time.Now())
}

// execute carries the verb flow: collect candidates, print the plan,
// and only under apply write banners + run the renames.
func execute(root string, cfg config.Config, apply bool, now time.Time) int {
	reports, skips := collectCandidates(root, cfg)

	if len(reports) == 0 {
		gatecli.Outf("docs-archive: nothing to archive (%d report(s) skipped: open §f items)\n", len(skips))

		return gatecli.ExitClean
	}

	archived := cfg.ArchivedSubdir
	if archived == "" {
		archived = config.Defaults().ArchivedSubdir
	}

	printPlan(archived, reports, skips)

	if !apply {
		gatecli.Outf("docs-archive: DRY-RUN — %d candidate(s); re-run with --apply to banner + git mv\n", len(reports))

		return gatecli.ExitClean
	}

	return applyArchive(root, archived, reports, now)
}

// candidate is one fully-resolved report: path relative to the repo
// root, plus the open-item count that proved it (always zero here —
// carried for the plan face).
type candidate struct {
	rel  string
	open int
}

// collectCandidates walks the unarchived report corpus and splits it
// into fully-resolved candidates and open-item skips.
func collectCandidates(root string, cfg config.Config) ([]candidate, []candidate) {
	glob := filepath.Join(orDefault(cfg.ReportDir, config.Defaults().ReportDir), orDefault(cfg.ReportGlob, config.Defaults().ReportGlob))

	entries, err := filepath.Glob(filepath.Join(root, glob))
	if err != nil {
		return nil, nil
	}

	sort.Strings(entries)

	var (
		reports []candidate
		skips   []candidate
	)

	for _, entry := range entries {
		info, err := os.Stat(entry)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}

		rel, err := filepath.Rel(root, entry)
		if err != nil {
			continue
		}

		open := len(docsfreshness.ExtractItems(docsfreshness.ExtractSectionF(entry)))

		if open > 0 {
			skips = append(skips, candidate{rel: rel, open: open})

			continue
		}

		reports = append(reports, candidate{rel: rel})
	}

	return reports, skips
}

// applyArchive banners each candidate and git-mvs it into the archived
// subdir. Any failure aborts with exit 2 — a half-archived corpus is
// fine (each report is its own boundary), but the operator must see it.
func applyArchive(root, archived string, reports []candidate, now time.Time) int {
	for _, report := range reports {
		path := filepath.Join(root, report.rel)

		data, err := os.ReadFile(path)
		if err != nil {
			gatecli.Errf("FATAL: reading %s: %v\n", report.rel, err)

			return gatecli.ExitUsage
		}

		if strings.HasPrefix(string(data), bannerPrefix) {
			gatecli.Errf("FATAL: %s already carries the archive banner — refusing to double-archive\n", report.rel)

			return gatecli.ExitUsage
		}

		if err := writePreservingMode(path, banner(now)+string(data)); err != nil {
			gatecli.Errf("FATAL: writing banner into %s: %v\n", report.rel, err)

			return gatecli.ExitUsage
		}

		destRel := filepath.Join(filepath.Dir(report.rel), archived, filepath.Base(report.rel))

		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, destRel)), 0o750); err != nil {
			gatecli.Errf("FATAL: creating the archive dir: %v\n", err)

			return gatecli.ExitUsage
		}

		//nolint:gosec // git -C this repo; the subcommand is a literal and both paths are repo-relative
		cmd := exec.CommandContext(context.Background(), "git", "-C", root, "mv", "--", report.rel, destRel)

		if out, err := cmd.CombinedOutput(); err != nil {
			gatecli.Errf("FATAL: git mv %s: %v\n%s", report.rel, err, out)

			return gatecli.ExitUsage
		}

		gatecli.Outf("ARCHIVED %s -> %s\n", report.rel, destRel)
	}

	gatecli.Outf("docs-archive: archived %d report(s) — re-run the battery over the repo\n", len(reports))

	return gatecli.ExitClean
}

// writePreservingMode writes content keeping the file's own permission
// bits (fresh files get the house default) — a banner must never change
// a report's mode.
func writePreservingMode(path, content string) error {
	mode := os.FileMode(0o600)

	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}

	return os.WriteFile( //nolint:gosec // G306,G703: the mode IS the source file's; the path argument IS the tool's input contract
		path,
		[]byte(content),
		mode,
	)
}

// banner renders the Live-remainder blockquote: the house convention,
// genericized (the observed 2026-11-23 banners name the owning sections
// of that repo; the claim and proof legs are the invariant).
func banner(now time.Time) string {
	date := now.Format("2006-01-02")

	return fmt.Sprintf(
		"%s archived %s by the cadence ritual. All open rows were verified\n"+
			"> harvested/routed at archive time (freshness --audit-all + routing +\n"+
			"> annotate green over the corpus); owning sections live in the living\n"+
			"> docs (TODO_LIST.md, ROADMAP.md).\n\n",
		bannerPrefix, date,
	)
}

// orDefault is the settings-projection fallback: config value, else
// the shipped default (the zero-value-never-reaches-logic contract).
func orDefault(value, fallback string) string {
	if value != "" {
		return value
	}

	return fallback
}

// printPlan writes the dry-run face: one ARCHIVE line per candidate,
// one SKIP line per open report.
func printPlan(archived string, reports, skips []candidate) {
	for _, report := range reports {
		gatecli.Outf(
			"ARCHIVE %s -> %s\n",
			report.rel,
			filepath.Join(filepath.Dir(report.rel), archived, filepath.Base(report.rel)),
		)
	}

	for _, skip := range skips {
		gatecli.Outf("SKIP    %s (%d open §f item(s))\n", skip.rel, skip.open)
	}
}
