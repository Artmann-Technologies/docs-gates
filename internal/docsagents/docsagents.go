// Package docsagents dates AGENTS.md — the AI-collaboration contract
// every fleet repo carries at its root.
//
// AGENTS.md is a claim with a shelf life: it pins conventions, budgets,
// and non-obvious truths, and code moves while the file stays. The
// structure linter guards existence and the line budget; FRESHNESS is a
// docs-health concern, so this gate owns it. It stats the file's mtime
// and flags it against two bars: warn past agents_warn_days (default 7,
// "revisit soon"), fail past agents_max_days (default 14 — the proven
// bar go-structure-linter's agent-config rule carried before it retired
// its freshness check on 2026-11-23). A missing AGENTS.md is a
// fail-severity finding: the agent contract is fleet-mandatory.
//
// Rollout contract (the docsstub doctrine): findings WARN and exit 0
// until the repo flips agents_enforce; once flipped, only
// fail-severity findings exit 1 — the warn bar stays informational at
// every enforcement level, because "revisit soon" and "the contract has
// rotted" are different obligations. mtime is an honest-enough signal
// for drift (a fresh clone resets it, but content rot is the stamps
// gate's job, not this one's).
package docsagents

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// hoursPerDay converts an age in hours to whole days.
const hoursPerDay = 24

// Finding severities: the warn bar is informational at every
// enforcement level; the fail bar is what enforcement gates on.
const (
	severityWarn = "warn"
	severityFail = "fail"
)

// Settings is the gate's config-driven knobs. Every field falls back to
// the shipped default when the repo's .docs-gates.json is absent or
// silent, so zero-value Settings never reach the gate logic.
type Settings struct {
	File     string // repo-relative AGENTS.md path to date
	WarnDays int    // age (days) past this WARN
	MaxDays  int    // age (days) past this FAIL
	Enforce  bool   // false = findings WARN and exit 0; true = fail-severity findings FAIL
	Now      time.Time
}

// SettingsFromConfig projects the shared config onto the gate's knobs,
// keeping the shipped defaults wherever the config is silent. Now is
// injectable for tests; zero means the wall clock.
func SettingsFromConfig(cfg config.Config) Settings {
	return Settings{
		File:     config.Or(cfg.AgentsFile, config.DefaultAgentsFile),
		WarnDays: cfg.AgentsWarnDays,
		MaxDays:  cfg.AgentsMaxDays,
		Enforce:  cfg.AgentsEnforce,
	}
}

// Usage is the `docs-gates agents --help` text — the verb's contract in
// one place, printed by the dispatcher before the parser runs.
const Usage = `docs-agents — the AGENTS.md freshness gate

usage: docs-agents [repo-root]

No flags: the file (agents_file, default AGENTS.md), the age bars
(agents_warn_days, agents_max_days) and the rollout switch
(agents_enforce) live in the repo's .docs-gates.json. The gate stats
the file's mtime and flags it: WARN past the warn bar, FAIL past the
max bar; a missing file is a fail-severity finding — the agent
contract is fleet-mandatory. While agents_enforce is false every
finding prints WARN and exits 0; once flipped, only fail-severity
findings exit 1. The warn bar stays informational at every
enforcement level.

Exit codes: 0 clean or warn-mode findings, 1 enforced fail findings, 2 usage error.
`

// Run is the CLI entry (repo-root positional).
func Run(args []string) int {
	flags, root, err := gatecli.Parse(args)
	if err != nil {
		return gatecli.FatalUsage("agents", err)
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

	findings, err := Scan(root, settings)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	printFindings(findings, settings)

	if len(findings) == 0 {
		gatecli.Outf(
			"OK %s within the freshness bars (warn %dd, max %dd)\n",
			settings.File,
			settings.WarnDays,
			settings.MaxDays,
		)

		return gatecli.ExitClean
	}

	if anyEnforcedFailure(findings, settings) {
		return gatecli.ExitFindings
	}

	return gatecli.ExitClean
}

// printFindings renders one line per finding, naming the bar it crossed
// and whether enforcement is active.
func printFindings(findings []Finding, settings Settings) {
	for _, finding := range findings {
		enforced := settings.Enforce && finding.Severity == severityFail
		verdict := map[bool]string{true: "FAIL", false: "WARN"}[enforced]

		mode := ""
		if !enforced && finding.Severity == severityFail {
			mode = " (agents_enforce=false, warn mode)"
		}

		if finding.Missing {
			gatecli.Outf("%s %s missing — the agent contract is fleet-mandatory%s\n", verdict, finding.Path, mode)

			continue
		}

		gatecli.Outf(
			"%s %s is %dd old (warn %dd, max %dd)%s\n",
			verdict,
			finding.Path,
			finding.AgeDays,
			settings.WarnDays,
			settings.MaxDays,
			mode,
		)
	}
}

// anyEnforcedFailure reports whether an enforced fail-severity finding exists.
func anyEnforcedFailure(findings []Finding, settings Settings) bool {
	for _, finding := range findings {
		if settings.Enforce && finding.Severity == severityFail {
			return true
		}
	}

	return false
}

// Finding is one agents-doc verdict: either the file is missing
// (Missing, fail severity) or it aged past a bar ("warn" or "fail").
type Finding struct {
	Path     string
	Missing  bool
	AgeDays  int
	Severity string
}

// Scan dates settings.File under root against the bars and returns the
// findings. A missing file is a fail-severity finding; an unreadable
// one is an error.
func Scan(root string, settings Settings) ([]Finding, error) {
	info, err := os.Stat(filepath.Join(root, settings.File))
	if err != nil {
		if os.IsNotExist(err) {
			return []Finding{{Path: settings.File, Missing: true, Severity: severityFail}}, nil
		}

		return nil, fmt.Errorf("stating %s: %w", settings.File, err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("%s is a directory, not the agent contract file", settings.File)
	}

	ageDays := int(settings.Now.Sub(info.ModTime()).Hours() / hoursPerDay)
	if ageDays <= settings.WarnDays {
		return nil, nil
	}

	severity := severityWarn
	if ageDays > settings.MaxDays {
		severity = severityFail
	}

	return []Finding{{Path: settings.File, AgeDays: ageDays, Severity: severity}}, nil
}
