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
	"slices"
	"strconv"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/docsannotate"
	"github.com/Artmann-Technologies/docs-gates/internal/docscadence"
	"github.com/Artmann-Technologies/docs-gates/internal/docsfreshness"
	"github.com/Artmann-Technologies/docs-gates/internal/docsrouting"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// Verdicts the matrix prints per gate.
const (
	verdictPass = "PASS"
	verdictFail = "FAIL"
	verdictSkip = "SKIP"
)

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
		{name: "freshness", fn: docsfreshness.Run, args: []string{root}},
		{name: "audit-all", fn: docsfreshness.Run, args: auditArgs},
		{name: "annotate", fn: docsannotate.Run, args: annotateArgs},
		{name: "routing", fn: docsrouting.Run, args: []string{root, "--strict"}},
		{name: "cadence", fn: docscadence.Run, args: []string{root}},
	}
}

// skipped reports whether cfg.SkipGates names the gate (the battery verbs
// share the gate names).
func skipped(cfg config.Config, name string) bool {
	return slices.Contains(cfg.SkipGates, name)
}

// Run is the CLI entry for `docs-gates run` (repo-root positional, no
// flags: the knobs live in the repo config).
func Run(args []string) int {
	_, root, err := gatecli.Parse(args)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	cfg, warnings, err := config.Load(root)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	for _, warning := range warnings {
		gatecli.Errf("WARN: %s\n", warning)
	}

	type result struct {
		name string
		code int
		skip bool
	}

	list := gates(root, cfg)
	results := make([]result, 0, len(list))

	for _, g := range list {
		if skipped(cfg, g.name) {
			results = append(results, result{name: g.name, code: gatecli.ExitClean, skip: true})

			continue
		}

		gatecli.Outf("docs-gates: gate %s ...\n", g.name)

		out := gatecli.CaptureOutput(func() int {
			return g.fn(g.args)
		})

		if out.Code == gatecli.ExitClean {
			gatecli.Outf("%s", out.Stdout)

			results = append(results, result{name: g.name, code: out.Code})

			continue
		}

		gatecli.Errf("docs-gates: gate %s FAILED (rc=%d):\n", g.name, out.Code)
		gatecli.Errf("%s", out.Stdout)
		gatecli.Errf("%s", out.Stderr)

		results = append(results, result{name: g.name, code: out.Code})
	}

	gatecli.Outln("docs-gates: verdict matrix")

	failed := false

	for _, r := range results {
		switch {
		case r.skip:
			gatecli.Outf("  %s  %s (skip_gates)\n", verdictSkip, r.name)
		case r.code == gatecli.ExitClean:
			gatecli.Outf("  %s  %s\n", verdictPass, r.name)
		default:
			gatecli.Outf("  %s  %s (rc=%d)\n", verdictFail, r.name, r.code)

			failed = true
		}
	}

	if failed {
		gatecli.Errln("docs-gates: FAIL — the matrix above names the red gates")

		return gatecli.ExitFindings
	}

	gatecli.Outln("docs-gates: PASS — all doc gates green")

	return gatecli.ExitClean
}
