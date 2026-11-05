// Package provider wires docs-gates as a BuildFlow DAG provider via the
// go-finding/toolsdk Spec contract. BuildFlow discovers this Provider
// automatically through toolsdk.All() when a consumer blank-imports this
// package:
//
//	import _ "github.com/Artmann-Technologies/docs-gates/pkg/provider"
//
// This is the thin-adapter shape the fleet's other toolsdk providers use
// (go-structure-linter precedent): the Spec wraps the existing battery CLI
// behind toolsdk types so BuildFlow gets detect capability with zero
// BuildFlow-side glue. The root docs-gates module stays stdlib-only — this
// is a SEPARATE go module precisely so the toolsdk dependency never lands
// in the root go.mod (the hermetic vendorHash-free nix build contract,
// AGENTS.md).
//
// The gate binary is exec'd from PATH, not imported: the gates live in the
// root module's internal/ packages, unimportable across module boundaries
// by design. HealthCheck verifies the binary resolves, and the failure text
// carries the install line.
//
// Detect runs the battery's `run --json` machine face and parses the JSON
// verdict document — findings-as-data, no human-text row scanning (the
// byte-shape contract is enforced on the battery side by the parity
// golden, and HERE by the strict decoder: any unknown field fails the
// parse and surfaces as its own finding instead of silently reading zero
// values). Detect is read-only by contract (the battery never mutates the
// tree it judges); Repair stays nil — the docs rituals (harvest, archive,
// purge) are human edits with their own tooling (docs-strike, the
// annotate helper), not mechanical fixers.
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/larsartmann/go-finding"
	toolsdk "github.com/larsartmann/go-finding/toolsdk"
)

// toolName is the unique identifier BuildFlow uses for this provider.
const toolName = "docs-gates"

// verdictFail is the JSON verdict word that becomes a finding (PASS and
// SKIP rows are exit-clean by the battery's contract).
const verdictFail = "FAIL"

// installHint is the failure text HealthCheck reports when the binary is
// not on PATH.
const installHint = "docs-gates binary not on PATH — install via `nix run github:Artmann-Technologies/docs-gates` or `go install github.com/Artmann-Technologies/docs-gates/cmd/docs-gates@latest`"

// batteryReport is the provider-side mirror of the battery's --json
// document (pkg/provider cannot import internal/battery, so the shape is
// duplicated by contract and pinned by decodeReport's strict decoder plus
// the battery's battery-json parity golden).
type batteryReport struct {
	Root  string    `json:"root"`
	Pass  bool      `json:"pass"`
	Gates []gateRow `json:"gates"`
}

// gateRow is one matrix row's machine face.
type gateRow struct {
	Name       string `json:"name"`
	Verdict    string `json:"verdict"`
	ExitCode   int    `json:"exit_code"`
	Skipped    bool   `json:"skipped"`
	TimedOut   bool   `json:"timed_out"`
	DurationMS int64  `json:"duration_ms"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
}

// decodeReport parses the battery's --json stdout with unknown fields
// REFUSED: a renamed or removed row field fails the decode (and surfaces
// as a shape-contract finding) instead of silently decoding to a zero
// value — the provider-side half of the row-shape contract.
func decodeReport(stdout string) (batteryReport, error) {
	var report batteryReport

	dec := json.NewDecoder(strings.NewReader(stdout))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&report); err != nil {
		return batteryReport{}, err
	}

	return report, nil
}

// failedGates returns the FAIL rows in document order.
func failedGates(report batteryReport) []gateRow {
	var failed []gateRow

	for _, g := range report.Gates {
		if g.Verdict == verdictFail {
			failed = append(failed, g)
		}
	}

	return failed
}

// workingDir resolves the project directory from the context, falling back
// to the process working directory. BuildFlow's runner.WithWorkingDir sets
// the canonical key via finding.WithWorkingDir during module fan-out.
func workingDir(ctx context.Context) string {
	if dir := finding.WorkingDirFromContext(ctx); dir != "" {
		return dir
	}

	return "."
}

// runBattery execs the docs-gates binary's battery verb (the --json
// machine face) over dir and returns its exit code plus both output
// streams. The binary name is a package constant (gosec G204), and the
// context bounds the run.
func runBattery(ctx context.Context, bin, dir string) (int, string, string, error) {
	cmd := exec.CommandContext( //nolint:gosec // constant binary, not user input
		ctx,
		bin,
		"run",
		dir,
		"--json",
	)

	var out, errOut strings.Builder

	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err := cmd.Run()

	code := 0
	if exitErr, ok := err.(*exec.ExitError); ok { //nolint:errorlint // exec.ExitError is the ONLY error carrying an exit code; everything else is a start failure below
		code = exitErr.ExitCode()
		err = nil
	}

	return code, out.String(), errOut.String(), err
}

// Provider is the BuildFlow tool-sdk registration. The var assignment
// executes toolsdk.Register at package init time, adding this Spec to the
// process-global registry that BuildFlow walks via ToolsFromSDK().
//
// Detect: runs the five-gate battery (freshness, audit-all, annotate,
// routing --strict, cadence — thresholds from the repo's .docs-gates.json)
// and reports one finding per FAIL row. Clean runs produce zero findings;
// a usage-environment failure (exit 2) becomes one error finding naming
// the FATAL line; a red exit whose stdout does not decode as the JSON
// verdict document is a shape-contract break and becomes its own finding
// instead of silently passing.
//
//nolint:gochecknoglobals // BuildFlow plugin SDK requires package-level Provider registration
var Provider = toolsdk.Register(toolsdk.Spec{
	Name: toolName,
	Description: "docs-health battery: newest-report harvest freshness, whole-corpus audit replay, " +
		"struck-item verdicts + inline-code spacing, non-f routing gaps, TODO/report cadence. " +
		"Reads the repo's .docs-gates.json (defaults are the fleet-proven constants). " +
		"Parses `docs-gates run --json` (requires a battery with the JSON face, v0.2.0+).",
	// The docs-doctrine file set that activates the tool: a repo without
	// any of these has nothing for the gates to judge. TODO_LIST.md alone
	// is enough (the cadence gate needs it; the other gates degrade to
	// clean on a missing corpus).
	Trigger: toolsdk.Trigger{
		Files: []string{
			"TODO_LIST.md",
			"ROADMAP.md",
			"CHANGELOG.md",
			"docs/status/*.md",
			"docs/status/**/*.md",
			"docs/planning/*.md",
			".docs-gates.json",
		},
		Language: "", // language-agnostic: the gates judge markdown living docs
	},
	Inputs: []string{
		"TODO_LIST.md",
		"ROADMAP.md",
		"CHANGELOG.md",
		"FEATURES.md",
		"README.md",
		"AGENTS.md",
		"docs/status/**/*.md",
		"docs/planning/**/*.md",
		".docs-gates.json",
	},
	Outputs: nil, // read-only: the battery never writes the tree it judges
	Detect: finding.NamedDetectorFunc(toolName, func(ctx context.Context) ([]finding.Finding, error) {
		dir := workingDir(ctx)

		code, out, errOut, runErr := runBattery(ctx, toolName, dir)
		if runErr != nil {
			return nil, fmt.Errorf("%s run %s: %w", toolName, dir, runErr)
		}

		switch code {
		case 0, 1:
			report, decodeErr := decodeReport(out)
			if decodeErr != nil {
				// A battery exit whose stdout is not the pinned JSON
				// document is a verdict-shape contract break (version skew,
				// a renamed field) — report it as its own finding instead
				// of silently passing.
				return []finding.Finding{finding.NewFinding(
					"battery-exit-without-report",
					toolName,
					"docs-gates battery output over "+dir+" is not the pinned --json document (exit "+
						strconv.Itoa(code)+"): "+decodeErr.Error(),
					finding.SeverityError,
					finding.Position{File: finding.FilePath(dir)},
					finding.ConfidenceHigh,
				)}, nil
			}

			failed := failedGates(report)

			findings := make([]finding.Finding, 0, len(failed))
			for _, gate := range failed {
				message := "docs gate " + gate.Name + " is red over " + dir + " — run `docs-gates " + gate.Name + " " + dir + "` for the full output"
				if gate.TimedOut {
					message = "docs gate " + gate.Name + " TIMED OUT over " + dir + " (" +
						strconv.FormatInt(gate.DurationMS, 10) + "ms) — run `docs-gates " + gate.Name + " " + dir + "` for the full output"
				}

				findings = append(findings, finding.NewFinding(
					finding.RuleName(gate.Name),
					toolName,
					message,
					finding.SeverityError,
					finding.Position{File: finding.FilePath(dir)},
					finding.ConfidenceHigh,
				))
			}

			return findings, nil
		default: // 2 = usage/environment, or an unexpected code
			message := strings.TrimSpace(errOut)
			if message == "" {
				message = strings.TrimSpace(out)
			}

			return []finding.Finding{finding.NewFinding(
				"usage",
				toolName,
				"docs-gates battery could not run over "+dir+" (exit "+strconv.Itoa(code)+"): "+message,
				finding.SeverityError,
				finding.Position{File: finding.FilePath(dir)},
				finding.ConfidenceHigh,
			)}, nil
		}
	}),
	HealthCheck: func(_ context.Context) error {
		if _, err := exec.LookPath(toolName); err != nil {
			return fmt.Errorf("%s", installHint) //nolint:err113 // user-facing HealthCheck text; no errors.Is consumers
		}

		return nil
	},
})
