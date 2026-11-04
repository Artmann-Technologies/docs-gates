// Package docscadence pins the TODO_LIST purge cadence.
//
// The convention: completed items are DELETED from TODO_LIST.md and live
// in CHANGELOG.md only. Two data points (the 2026-10-08 purge of 111
// struck rows, the 2026-10-12 purge of 123) showed the class silently
// rebuilds — every struck-but-not-yet-purged DONE row is backlog the
// next session must re-triage, and the 123-row purge needed a CHANGELOG
// preservation entry that flipped 15 historical §f items (the biggest
// single docs-gates event of that month). This gate fails once the
// struck-row backlog crosses the threshold, so the purge happens at ~20
// rows of debt instead of ~120.
//
// The second arm pins the ARCHIVE cadence: the 2026-10-24 board held 9
// live reports in docs/status/ hours after a 31-report archive pass —
// reports accumulate as fast as struck rows, and an unarchived report is
// backlog the docs ritual must re-triage every pass. The arm fails once
// the unarchived count or the age of the oldest crosses its threshold.
package docscadence

import (
	"cmp"
	"os"
	"path/filepath"
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// Threshold is the maximum number of struck DONE rows tolerated before
// the purge ritual is owed. 20 keeps the manual strike-verify-purge loop
// comfortably inside one session while never letting the backlog compound.
const Threshold = 20

// ArchiveThreshold is the maximum number of unarchived reports tolerated
// in docs/status/ before the archive ritual is owed (archived/ itself is
// never counted — an archived report is closed inventory).
const ArchiveThreshold = 15

// ArchiveAgeThresholdDays is the maximum tolerated age of the OLDEST
// unarchived report. A report nobody archived for three weeks is drift
// nobody is watching; the banner ritual is owed long before that.
const ArchiveAgeThresholdDays = 21

// Settings is the gate's config-driven knobs. Every field falls back to
// the historical constant above when the repo's .docs-gates.json is
// absent or silent, so zero-value Settings never reach the gate logic
// (the same fallback shape the other ported gates use).
type Settings struct {
	StruckThreshold  int    // max struck DONE rows before the purge is owed
	ArchiveThreshold int    // max unarchived reports before the archive is owed
	ArchiveAgeDays   int    // max tolerated age (days) of the oldest unarchived report
	TodoFile         string // the TODO_LIST file classify() reads
	ReportDir        string // the unarchived status-report directory
}

// SettingsFromConfig projects the shared config onto the gate's knobs,
// keeping the historical constants wherever the config is silent.
func SettingsFromConfig(cfg config.Config) Settings {
	settings := Settings{
		StruckThreshold:  Threshold,
		ArchiveThreshold: ArchiveThreshold,
		ArchiveAgeDays:   ArchiveAgeThresholdDays,
		TodoFile:         cmp.Or(cfg.TodoFile, config.Defaults().TodoFile),
		ReportDir:        cmp.Or(cfg.ReportDir, config.Defaults().ReportDir),
	}

	if cfg.Cadence.StruckRows > 0 {
		settings.StruckThreshold = cfg.Cadence.StruckRows
	}

	if cfg.Cadence.MaxUnarchived > 0 {
		settings.ArchiveThreshold = cfg.Cadence.MaxUnarchived
	}

	if cfg.Cadence.OldestDays > 0 {
		settings.ArchiveAgeDays = cfg.Cadence.OldestDays
	}

	return settings
}

// Run is the CLI entry (repo-root positional).
func Run(args []string) int {
	flags, root, err := gatecli.Parse(args)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	_ = flags // no flags yet; the positional root is the only input

	cfg, err := gatecli.LoadConfig(root)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	settings := SettingsFromConfig(cfg)

	body, err := os.ReadFile(filepath.Join(root, settings.TodoFile))
	if err != nil {
		gatecli.Errf("FATAL: reading %s: %v\n", settings.TodoFile, err)

		return gatecli.ExitUsage
	}

	struck, bareDone := classify(string(body))

	gatecli.Outf(
		"docs-cadence: %d struck DONE row(s) in %s (threshold %d)\n",
		struck, settings.TodoFile, settings.StruckThreshold,
	)

	if bareDone > 0 {
		gatecli.Outf(
			"docs-cadence: %d row(s) carry a DONE status WITHOUT a ~~verdict~~ wrap — "+
				"docs-strike flips status and wraps together; an unwrapped DONE is a half-strike\n",
			bareDone,
		)
	}

	findings := 0

	if struck > settings.StruckThreshold {
		gatecli.Errf(
			"docs-cadence: OVERDUE — %d struck rows exceed the %d threshold; run the purge ritual now:\n",
			struck, settings.StruckThreshold,
		)
		gatecli.Errln("  1. verify every strike on-disk (the truth pass — a CHANGELOG claim is not state)")
		gatecli.Errln("  2. delete the struck rows, preserving substance in a CHANGELOG entry")
		gatecli.Errln(
			"  3. PRE-FLIGHT the CHANGELOG entry in ONE stdin batch BEFORE landing it " +
				"(the 2026-10-12 purge flipped 15 §f items)",
		)
		gatecli.Errln("  4. re-run docs-gates until green")

		findings++
	} else {
		gatecli.Outln("docs-cadence: PURGE-HEALTHY")
	}

	// The inventory lines display the dir with the historical trailing
	// slash (parity: the pre-config output spelled docs/status/ that way).
	archCount, archAge, archOldest, err := archiveInventory(root, time.Now(), settings.ReportDir)
	if err != nil {
		gatecli.Errf("FATAL: reading %s/: %v\n", settings.ReportDir, err)

		return gatecli.ExitUsage
	}

	if archCount > 0 {
		gatecli.Outf(
			"docs-cadence: %d unarchived report(s) in %s/ (threshold %d); oldest %s is %d day(s) old (threshold %d)\n",
			archCount,
			settings.ReportDir,
			settings.ArchiveThreshold,
			archOldest,
			archAge,
			settings.ArchiveAgeDays,
		)
	}

	if archCount > settings.ArchiveThreshold || archAge > settings.ArchiveAgeDays {
		gatecli.Errf(
			"docs-cadence: ARCHIVE-OVERDUE — %d unarchived report(s) (threshold %d), oldest %d day(s) old (threshold %d); run the archive ritual now:\n",
			archCount,
			settings.ArchiveThreshold,
			archAge,
			settings.ArchiveAgeDays,
		)
		gatecli.Errln("  1. verify every remaining item is struck, WONT-DO, or routed (check-rows BEFORE git mv)")
		gatecli.Errln(
			"  2. git mv to docs/status/archived/ with the `> **Live remainder:**` banner naming the owning sections",
		)
		gatecli.Errln("  3. re-run docs-gates until green")

		findings++
	} else {
		gatecli.Outln("docs-cadence: ARCHIVE-HEALTHY")
	}

	if findings > 0 {
		return gatecli.ExitFindings
	}

	return gatecli.ExitClean
}
