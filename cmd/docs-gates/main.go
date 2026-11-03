// Command docs-gates is the one binary behind the docs-health battery:
// every gate verb plus `run` (the ordered matrix). The five gate verbs
// keep the historical standalone behavior byte-for-byte — the legacy
// binary names (docs-freshness, docs-annotate, docs-routing,
// docs-cadence, docs-strike) ride flake wrappers that exec
// `docs-gates <verb>`, and each verb answers `--version` under its
// legacy name so scripts probing either shape read the same thing.
//
// Exit codes: 0 = clean, 1 = findings, 2 = usage/environment error.
package main

import (
	"os"

	"github.com/Artmann-Technologies/docs-gates/internal/battery"
	"github.com/Artmann-Technologies/docs-gates/internal/docsannotate"
	"github.com/Artmann-Technologies/docs-gates/internal/docscadence"
	"github.com/Artmann-Technologies/docs-gates/internal/docsfreshness"
	"github.com/Artmann-Technologies/docs-gates/internal/docsrouting"
	"github.com/Artmann-Technologies/docs-gates/internal/docsstrike"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// version is stamped by the flake build (ldflags -X main.version);
// "dev" in bare go build/run.
var version = "dev"

// displayName maps each verb to the program name its --version line
// carries: the battery verb answers as docs-gates, the five gate verbs
// answer under their legacy standalone names (the wrappers exec this
// binary, so `docs-freshness --version` must still read
// "docs-freshness <version>").
var displayName = map[string]string{ //nolint:gochecknoglobals // the fixed verb-to-legacy-name map, same class as the contract tables
	"run":       "docs-gates",
	"freshness": "docs-freshness",
	"annotate":  "docs-annotate",
	"routing":   "docs-routing",
	"cadence":   "docs-cadence",
	"strike":    "docs-strike",
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(argv []string) int {
	// A bare `docs-gates --version` (no verb) answers as the binary
	// itself; per-verb invocations fall through to their legacy names.
	if len(argv) == 1 && argv[0] == "--version" {
		gatecli.HandleVersion(argv, "docs-gates", version)

		return gatecli.ExitClean
	}

	if len(argv) == 0 {
		gatecli.Errln("usage: docs-gates <verb> [args]")
		gatecli.Errln("verbs: run | freshness | annotate | routing | cadence | strike")

		return gatecli.ExitUsage
	}

	verb, rest := argv[0], argv[1:]

	name, ok := displayName[verb]
	if !ok {
		gatecli.Errf("unknown verb %q — verbs: run | freshness | annotate | routing | cadence | strike\n", verb)

		return gatecli.ExitUsage
	}

	// --version MUST run before any parser prologue: the gate parsers
	// reject unknown flags and would die on it.
	if gatecli.HandleVersion(rest, name, version) {
		return gatecli.ExitClean
	}

	switch verb {
	case "run":
		return battery.Run(rest)
	case "freshness":
		return docsfreshness.Run(rest)
	case "annotate":
		return docsannotate.Run(rest)
	case "routing":
		return docsrouting.Run(rest)
	case "cadence":
		return docscadence.Run(rest)
	case "strike":
		return docsstrike.Run(rest)
	}

	return gatecli.ExitUsage // unreachable: the map and the switch cover the same verbs
}
