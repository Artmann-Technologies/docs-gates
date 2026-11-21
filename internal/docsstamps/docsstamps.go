// Package docsstamps dates a glossary's verification stamps.
//
// The fleet's honesty doctrine pins glossaries with "verified against
// code on 2026-11-17" stamps (~18 repos on census day 2026-11-21). A
// stamp is a claim with a shelf life: code moves, the stamp stays, and
// six months later the doc asserts a verification nobody could repeat.
// This gate parses the verify-word + ISO-date grammar and flags the
// OLDEST stamp per corpus file against two bars: warn past
// stamp_warn_days (default 90), fail past stamp_fail_days (default
// 180). Archived corpora are out of scope by construction — the gate
// scans only the configured files.
//
// Rollout contract: findings WARN and exit 0 until the repo flips
// stamp_enforce (fleet-wide flip rides a TODO row one week after
// ship); once enforced, only the fail-bar findings exit 1 — the warn
// bar stays informational at every enforcement level, because "revisit
// soon" and "the claim has rotted" are different obligations.
package docsstamps

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// Settings is the gate's config-driven knobs. Every field falls back to
// the shipped default when the repo's .docs-gates.json is absent or
// silent, so zero-value Settings never reach the gate logic.
type Settings struct {
	Files    []string // repo-relative corpus to date
	WarnDays int      // stamps older than this WARN
	FailDays int      // stamps older than this FAIL
	Enforce  bool     // false = findings WARN and exit 0; true = fail-bar findings FAIL
	Now      time.Time
}

// SettingsFromConfig projects the shared config onto the gate's knobs,
// keeping the shipped defaults wherever the config is silent. Now is
// injectable for tests; zero means the wall clock.
func SettingsFromConfig(cfg config.Config) Settings {
	return Settings{
		Files:    config.OrSlice(cfg.StampFiles, config.Defaults().StampFiles),
		WarnDays: cfg.StampWarnDays,
		FailDays: cfg.StampFailDays,
		Enforce:  cfg.StampEnforce,
	}
}

// Usage is the `docs-gates stamps --help` text — the verb's contract in
// one place, printed by the dispatcher before the parser runs.
const Usage = `docs-stamps — dates a glossary's verification stamps

usage: docs-stamps [repo-root]

No flags: the corpus (stamp_files), the age bars (stamp_warn_days,
stamp_fail_days) and the rollout switch (stamp_enforce) live in the
repo's .docs-gates.json. The gate reads the verify-word grammar
(verified / re-verified / refreshed / rebuilt followed by an ISO date
within the same line segment), takes the OLDEST stamp per file, and
flags it: WARN past the warn bar, FAIL past the fail bar. While
stamp_enforce is false every finding prints WARN and exits 0; once
flipped, only fail-bar findings exit 1. A missing corpus file is a
SKIP note, not a finding; a corpus file with no stamps at all is a
SKIP note too (nothing claimed, nothing owed).

Exit codes: 0 clean or warn-mode findings, 1 enforced fail-bar findings, 2 usage error.
`

// Run is the CLI entry (repo-root positional).
func Run(args []string) int {
	flags, root, err := gatecli.Parse(args)
	if err != nil {
		return gatecli.FatalUsage("stamps", err)
	}

	_ = flags // no flags; the positional root is the only input

	cfg, err := gatecli.LoadConfig(root)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	settings := SettingsFromConfig(cfg)
	if settings.Now.IsZero() {
		settings.Now = time.Now()
	}

	findings, skips, err := Scan(root, settings)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	for _, skip := range skips {
		gatecli.Outf("SKIP %s\n", skip)
	}

	for _, finding := range findings {
		enforced := settings.Enforce && finding.Severity == "fail"
		verdict := map[bool]string{true: "FAIL", false: "WARN"}[enforced]

		mode := ""
		if !enforced && finding.Severity == "fail" {
			mode = " (stamp_enforce=false, warn mode)"
		}

		gatecli.Outf(
			"%s %s oldest verification stamp %s is %dd old (warn %dd, fail %dd)%s\n",
			verdict,
			finding.Path,
			finding.Date.Format("2006-01-02"),
			finding.AgeDays,
			settings.WarnDays,
			settings.FailDays,
			mode,
		)
	}

	if len(findings) == 0 {
		gatecli.Outf("OK no verification stamp past %d day(s) in %d file(s)\n", settings.WarnDays, len(settings.Files))

		return gatecli.ExitClean
	}

	for _, finding := range findings {
		if settings.Enforce && finding.Severity == "fail" {
			return gatecli.ExitFindings
		}
	}

	return gatecli.ExitClean
}

// Finding is one aged-out stamp: the file, the oldest stamp date, its
// age in days, and the bar it crossed ("warn" or "fail").
type Finding struct {
	Path     string
	Date     time.Time
	AgeDays  int
	Severity string
}

// stampRe reads the verify-word grammar: a verify word, then an ISO
// date within the same line segment (pipes end the segment — table
// cells are segments). Capital-insensitive: the fleet writes both
// "Verified against code" and "verified live".
var stampRe = regexp.MustCompile(
	`(?i)\b(re-?verified|verified|refreshed|rebuilt)\b[^|\n]{0,120}?(20[0-9]{2}-[0-9]{2}-[0-9]{2})`,
)

// Dates extracts every verification-stamp date from content.
func Dates(content string) []time.Time {
	var dates []time.Time

	for _, match := range stampRe.FindAllStringSubmatch(content, -1) {
		date, err := time.Parse("2006-01-02", match[2])
		if err != nil {
			continue
		}

		dates = append(dates, date)
	}

	return dates
}

// Scan collects the oldest stamp per corpus file and flags it against
// the bars. Files that are missing or carry no stamps are skips.
func Scan(root string, settings Settings) ([]Finding, []string, error) {
	var findings []Finding

	var skips []string

	for _, rel := range settings.Files {
		body, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			if os.IsNotExist(err) {
				skips = append(skips, rel+" (missing)")

				continue
			}

			return nil, nil, fmt.Errorf("reading %s: %w", rel, err)
		}

		dates := Dates(string(body))
		if len(dates) == 0 {
			skips = append(skips, rel+" (no stamps)")

			continue
		}

		oldest := dates[0]
		for _, date := range dates[1:] {
			if date.Before(oldest) {
				oldest = date
			}
		}

		ageDays := int(settings.Now.Sub(oldest).Hours() / 24)
		if ageDays <= settings.WarnDays {
			continue
		}

		severity := "warn"
		if ageDays > settings.FailDays {
			severity = "fail"
		}

		findings = append(findings, Finding{
			Path:     rel,
			Date:     oldest,
			AgeDays:  ageDays,
			Severity: severity,
		})
	}

	return findings, skips, nil
}
