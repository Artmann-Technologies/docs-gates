// Package battery runs the five doc gates as one ordered matrix — the
// extracted form of the shell battery's Go segment (scripts/docs-gates.sh
// steps 2-4 and 7-8 in pbx-artmann). Every gate runs INDEPENDENTLY and
// the run prints one verdict matrix at the end: one foreign red at step 3
// must never mask steps 4-8 (the 2026-10-30 failure class the shell
// matrix was built for). Exit contract matches the gate CLIs: 0 = all
// green, 1 = any gate red, 2 = usage/environment error.
//
// The gate set and their argv mirror the proven battery invocations
// byte-for-byte; skip_gates and audit_max come from the repo's
// .docs-gates.json (a repo config can drop a gate or pin the audit
// ratchet without touching this package).
package battery

import (
	"encoding/json"
	"slices"
	"strconv"
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/docsannotate"
	"github.com/Artmann-Technologies/docs-gates/internal/docscadence"
	"github.com/Artmann-Technologies/docs-gates/internal/docsfreshness"
	"github.com/Artmann-Technologies/docs-gates/internal/docsrouting"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// Verdicts the matrix prints per gate (the JSON face's `verdict` field
// carries the same words).
const (
	verdictPass = "PASS"
	verdictFail = "FAIL"
	verdictSkip = "SKIP"
)

// flagJSON is the battery's one CLI flag: the machine face. Everything
// else (thresholds, ratchets, skip list) lives in the repo config.
const flagJSON = "--json"

// gate is one battery entry: the matrix name, the gate's Run, and the
// exact argv the proven battery passes it (root first).
type gate struct {
	name string
	fn   func([]string) int
	args []string
}

// gates builds the ordered gate list for a root under cfg. The audit-all
// argv carries the config's audit_max only when it is a ratchet pin
// (>= 0); the strict default (-1) omits the flag, matching the gate's
// own CLI semantics.
func gates(root string, cfg config.Config) []gate {
	auditArgs := []string{root, "--audit-all"}
	if cfg.AuditMax >= 0 {
		auditArgs = append(auditArgs, "--audit-max="+strconv.Itoa(cfg.AuditMax))
	}

	annotateArgs := []string{root}
	if cfg.AnnotateMax >= 0 {
		annotateArgs = append(annotateArgs, "--annotate-max="+strconv.Itoa(cfg.AnnotateMax))
	}

	return []gate{
		{name: config.GateFreshness, fn: docsfreshness.Run, args: []string{root}},
		{name: config.GateAuditAll, fn: docsfreshness.Run, args: auditArgs},
		{name: config.GateAnnotate, fn: docsannotate.Run, args: annotateArgs},
		{name: config.GateRouting, fn: docsrouting.Run, args: []string{root, "--strict"}},
		{name: config.GateCadence, fn: docscadence.Run, args: []string{root}},
	}
}

// Usage is the `docs-gates run --help` text — the verb's contract in one
// place, printed by the dispatcher before any gate parser runs.
const Usage = `docs-gates run — all five doc gates as one ordered verdict matrix

usage: docs-gates run [repo-root] [--json]

Runs freshness, audit-all, annotate, routing and cadence INDEPENDENTLY:
one red gate never masks the gates behind it. Gates named in skip_gates
(config) print SKIP and stay exit-clean. Every gate runs under the
gate_timeout_sec wall-clock cap (0 disables it).

  --json  machine face: stdout carries ONLY the verdict matrix as one
          JSON document; every human line (per-gate heartbeat, red
          gate output, the final PASS/FAIL) rides stderr

The remaining knobs (skip_gates, audit_max, annotate_max,
gate_timeout_sec, thresholds) live in the repo's .docs-gates.json — a
repo without one gets the fleet defaults, zero config needed.

Exit codes: 0 all green, 1 any gate FAIL, 2 usage/config error.
`

// Run is the CLI entry for `docs-gates run` (repo-root positional plus
// the --json machine face; every other knob lives in the repo config).
func Run(args []string) int {
	flags, root, err := gatecli.Parse(args, flagJSON)
	if err != nil {
		return gatecli.FatalUsage("run", err)
	}

	cfg, err := gatecli.LoadConfig(root)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	// result rows flow through the package-level runGates/renderMatrix
	// pair so the timeout discrimination test can plant a wedged gate
	// without going through the real gate list.
	results := runGates(gates(root, cfg), cfg.SkipGates, time.Duration(cfg.GateTimeoutSec)*time.Second, flags[flagJSON])

	if flags[flagJSON] {
		return renderJSON(results, root)
	}

	if renderMatrix(results) {
		gatecli.Errln("docs-gates: FAIL — the matrix above names the red gates")

		return gatecli.ExitFindings
	}

	gatecli.Outln("docs-gates: PASS — all doc gates green")

	return gatecli.ExitClean
}

// result is one gate's outcome for the verdict matrix; a timedOut row
// renders as FAIL (rc=timeout) — an expired wedge counts as a finding,
// never a hang of the whole battery. The captured streams ride the row
// so the --json face can embed them verbatim.
type result struct {
	name     string
	code     int
	skip     bool
	timedOut bool
	duration time.Duration
	stdout   string
	stderr   string
}

// runGates executes the gate list INDEPENDENTLY under one per-gate
// wall-clock cap: a gate that exceeds it is abandoned (the gate cannot
// be killed — see gatecli.CaptureOutputWithTimeout) and its row fails
// while the battery moves on, so one wedged gate can never mask the
// gates behind it (the 2026-10-30 failure class, now also bounded in
// time). A timeout <= 0 never fires — the config opt-out.
//
// jsonMode is the machine face: stdout stays reserved for the final JSON
// document, so every progress line (heartbeat, red-gate output, matrix)
// detours to stderr and the captured streams ride the result rows.
func runGates(list []gate, skipGates []string, timeout time.Duration, jsonMode bool) []result {
	results := make([]result, 0, len(list))

	for _, g := range list {
		if slices.Contains(skipGates, g.name) {
			results = append(results, result{name: g.name, code: gatecli.ExitClean, skip: true})

			continue
		}

		progress := gatecli.Outf
		if jsonMode {
			progress = gatecli.Errf
		}

		progress("docs-gates: gate %s ...\n", g.name)

		stopWatchdog := startHalfBudgetWatchdog(g.name, timeout)

		start := time.Now()

		out, timedOut := gatecli.CaptureOutputWithTimeout(func() int {
			return g.fn(g.args)
		}, timeout)

		elapsed := time.Since(start)

		stopWatchdog()

		row := result{name: g.name, code: out.Code, duration: elapsed, stdout: out.Stdout, stderr: out.Stderr}

		switch {
		case timedOut:
			row.code, row.timedOut = gatecli.ExitFindings, true

			gatecli.Errf("docs-gates: gate %s TIMED OUT after %s (cap %s) — row fails, battery continues\n",
				g.name, gatecli.FormatDuration(elapsed), gatecli.FormatDuration(timeout))
			progress("docs-gates: gate %s abandoned after %s (timeout)\n",
				g.name, gatecli.FormatDuration(elapsed))
		case out.Code == gatecli.ExitClean:
			if !jsonMode {
				gatecli.Outf("%s", out.Stdout)
			}

			progress("docs-gates: gate %s done in %s\n", g.name, gatecli.FormatDuration(elapsed))
		default:
			gatecli.Errf("docs-gates: gate %s FAILED (rc=%d):\n", g.name, out.Code)
			gatecli.Errf("%s", out.Stdout)
			gatecli.Errf("%s", out.Stderr)
			progress("docs-gates: gate %s done in %s\n", g.name, gatecli.FormatDuration(elapsed))
		}

		results = append(results, row)
	}

	return results
}

// startHalfBudgetWatchdog arms the ONE live mid-gate line: once a gate
// burns past half its wall-clock budget, the watcher names the cap on
// the REAL stdout (gatecli.Livef — during a capture the swapped streams
// would park the warning inside the gate's own buffer, killing exactly
// the early-warning value it exists for). The CV wedge sat invisible in
// band for hours; a watcher would have named the budget it was burning.
// timeout <= 0 (the config opt-out) arms nothing; the returned func is
// always safe to call.
func startHalfBudgetWatchdog(gateName string, timeout time.Duration) func() {
	if timeout <= 0 {
		return func() {}
	}

	stop := make(chan struct{})

	go func() {
		timer := time.NewTimer(timeout / 2)
		defer timer.Stop()

		select {
		case <-timer.C:
			gatecli.Livef(
				"docs-gates: gate %s over half its %s budget — abandoned at the cap\n",
				gateName,
				gatecli.FormatDuration(timeout),
			)
		case <-stop:
		}
	}()

	return func() { close(stop) }
}

// batteryReport is the --json machine face: one document describing the
// whole matrix, shaped for the consumers that exec this binary (the
// BuildFlow provider parses it instead of scanning human matrix rows).
// Snake_case json tags are pinned by the tagliatelle lint rule.
type batteryReport struct {
	Root  string       `json:"root"`
	Pass  bool         `json:"pass"`
	Gates []gateReport `json:"gates"`
}

// gateReport is one matrix row's machine face. Verdict carries the same
// words the human matrix prints; Stdout/Stderr embed the gate's captured
// streams verbatim so a red gate's findings travel with the document.
type gateReport struct {
	Name       string `json:"name"`
	Verdict    string `json:"verdict"`
	ExitCode   int    `json:"exit_code"`
	Skipped    bool   `json:"skipped"`
	TimedOut   bool   `json:"timed_out"`
	DurationMS int64  `json:"duration_ms"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
}

// renderJSON prints the machine face: stdout carries ONLY the JSON
// document (the human matrix and the final PASS/FAIL line ride stderr),
// and the exit contract is unchanged. The verdict arithmetic is the
// human matrix's, read from the same rows.
func renderJSON(results []result, root string) int {
	report := batteryReport{Root: root, Gates: make([]gateReport, 0, len(results))}

	failed := false

	for _, r := range results {
		verdict := verdictFail

		switch {
		case r.skip:
			verdict = verdictSkip
		case !r.timedOut && r.code == gatecli.ExitClean:
			verdict = verdictPass
		default:
			failed = true
		}

		report.Gates = append(report.Gates, gateReport{
			Name:       r.name,
			Verdict:    verdict,
			ExitCode:   r.code,
			Skipped:    r.skip,
			TimedOut:   r.timedOut,
			DurationMS: r.duration.Milliseconds(),
			Stdout:     r.stdout,
			Stderr:     r.stderr,
		})
	}

	report.Pass = !failed

	doc, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		gatecli.Errf("FATAL: rendering the JSON matrix: %v\n", err)

		return gatecli.ExitUsage
	}

	gatecli.Outln(string(doc))

	if failed {
		gatecli.Errln("docs-gates: FAIL — the JSON matrix names the red gates")

		return gatecli.ExitFindings
	}

	gatecli.Errln("docs-gates: PASS — all doc gates green")

	return gatecli.ExitClean
}

// renderMatrix prints the verdict rows and reports whether any gate
// failed (skip rows never count).
func renderMatrix(results []result) bool {
	gatecli.Outln("docs-gates: verdict matrix")

	failed := false

	for _, r := range results {
		switch {
		case r.skip:
			gatecli.Outf("  %s  %s (skip_gates)\n", verdictSkip, r.name)
		case r.timedOut:
			gatecli.Outf("  %s  %s (in %s, rc=timeout)\n", verdictFail, r.name, gatecli.FormatDuration(r.duration))

			failed = true
		case r.code == gatecli.ExitClean:
			gatecli.Outf("  %s  %s (in %s)\n", verdictPass, r.name, gatecli.FormatDuration(r.duration))
		default:
			gatecli.Outf("  %s  %s (in %s, rc=%d)\n", verdictFail, r.name, gatecli.FormatDuration(r.duration), r.code)

			failed = true
		}
	}

	return failed
}
