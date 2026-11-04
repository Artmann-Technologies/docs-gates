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
// Detect is read-only by contract (the battery never mutates the tree it
// judges); Repair stays nil — the docs rituals (harvest, archive, purge)
// are human edits with their own tooling (docs-strike, the annotate
// helper), not mechanical fixers.
package provider

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/larsartmann/go-finding"
	toolsdk "github.com/larsartmann/go-finding/toolsdk"
)

// toolName is the unique identifier BuildFlow uses for this provider.
const toolName = "docs-gates"

// matrixLinePrefix is the verdict-matrix row shape the battery prints
// (`  PASS  freshness` / `  FAIL  cadence (rc=1)` / `  SKIP  gate`). The
// parser matches rows by this exact indentation so prose lines can never
// parse as verdicts.
const matrixLinePrefix = "  "

// minMatrixFields is the shortest verdict row the parser accepts: the
// verdict word plus the gate name.
const minMatrixFields = 2

// installHint is the failure text HealthCheck reports when the binary is
// not on PATH.
const installHint = "docs-gates binary not on PATH — install via `nix run github:Artmann-Technologies/docs-gates` or `go install github.com/Artmann-Technologies/docs-gates/cmd/docs-gates@latest`"

// workingDir resolves the project directory from the context, falling back
// to the process working directory. BuildFlow's runner.WithWorkingDir sets
// the canonical key via finding.WithWorkingDir during module fan-out.
func workingDir(ctx context.Context) string {
	if dir := finding.WorkingDirFromContext(ctx); dir != "" {
		return dir
	}

	return "."
}

// runBattery execs the docs-gates binary's battery verb over dir and
// returns its exit code plus both output streams. The binary name is a
// package constant (gosec G204), and the context bounds the run.
func runBattery(ctx context.Context, bin, dir string) (int, string, string, error) {
	cmd := exec.CommandContext( //nolint:gosec,noctx // constant binary, not user input
		ctx,
		bin,
		"run",
		dir,
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

// parseMatrix extracts the failed gate names from the battery's verdict
// matrix, preserving the printed order.
func parseMatrix(stdout string) []string {
	var failed []string

	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, matrixLinePrefix) {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < minMatrixFields {
			continue
		}

		if fields[0] == "FAIL" {
			failed = append(failed, fields[1])
		}
	}

	return failed
}

// Provider is the BuildFlow tool-sdk registration. The var assignment
// executes toolsdk.Register at package init time, adding this Spec to the
// process-global registry that BuildFlow walks via ToolsFromSDK().
//
// Detect: runs the five-gate battery (freshness, audit-all, annotate,
// routing --strict, cadence — thresholds from the repo's .docs-gates.json)
// and reports one finding per red gate. Clean runs produce zero findings;
// a usage-environment failure (exit 2) becomes one error finding naming
// the FATAL line.
//
//nolint:gochecknoglobals // BuildFlow plugin SDK requires package-level Provider registration
var Provider = toolsdk.Register(toolsdk.Spec{
	Name: toolName,
	Description: "docs-health battery: newest-report harvest freshness, whole-corpus audit replay, " +
		"struck-item verdicts + inline-code spacing, non-f routing gaps, TODO/report cadence. " +
		"Reads the repo's .docs-gates.json (defaults are the fleet-proven constants).",
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
		case 0:
			return nil, nil // clean corpus, zero findings
		case 1:
			failed := parseMatrix(out)
			if len(failed) == 0 {
				// A red exit whose matrix carries no parseable FAIL row is a
				// verdict-shape contract break — report it as its own finding
				// instead of silently passing.
				failed = []string{"battery-exit-without-matrix"}
			}

			findings := make([]finding.Finding, 0, len(failed))
			for _, gate := range failed {
				findings = append(findings, finding.NewFinding(
					finding.RuleName(gate),
					toolName,
					"docs gate "+gate+" is red over "+dir+" — run `docs-gates "+gate+" "+dir+"` for the full output",
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
