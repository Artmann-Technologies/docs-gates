// Command docs-gates is the one binary behind the docs-health battery:
// every gate verb plus `run` (the ordered matrix). The five gate verbs
// keep the historical standalone behavior byte-for-byte — the legacy
// binary names (docs-freshness, docs-annotate, docs-routing,
// docs-cadence, docs-strike) ride flake wrappers that exec
// `docs-gates <verb>`, and each verb answers `--version` under its
// legacy name so scripts probing either shape read the same thing.
//
// Help is intercepted BEFORE any gate parser runs (--help/-h would die
// there as unknown flags, the same slot --version rides): the bare
// `docs-gates --help|-h|help` prints the root usage, `help <verb>` and
// `<verb> --help|-h` print the verb's Usage from its own package.
//
// Exit codes: 0 = clean, 1 = findings, 2 = usage/environment error.
package main

import (
	"fmt"
	"os"
	"strings"

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

// The verb vocabulary. Deliberately cmd-local: config.KnownGates names
// the BATTERY rows (the skip_gates vocabulary — it has audit-all and
// lacks run/strike). Overlapping values, distinct concepts.
const (
	verbRun       = "run"
	verbFreshness = "freshness"
	verbAnnotate  = "annotate"
	verbRouting   = "routing"
	verbCadence   = "cadence"
	verbStrike    = "strike"
)

// suggestMaxEditDistance is the typo tolerance of suggestVerb: any verb
// within this Levenshtein distance is offered as a did-you-mean.
const suggestMaxEditDistance = 2

// displayName maps each verb to the program name its --version line
// carries: the battery verb answers as docs-gates, the five gate verbs
// answer under their legacy standalone names (the wrappers exec this
// binary, so `docs-freshness --version` must still read
// "docs-freshness <version>").
//
//nolint:gochecknoglobals // the fixed verb-to-legacy-name map, same class as the contract tables
var displayName = map[string]string{
	verbRun:       "docs-gates",
	verbFreshness: "docs-freshness",
	verbAnnotate:  "docs-annotate",
	verbRouting:   "docs-routing",
	verbCadence:   "docs-cadence",
	verbStrike:    "docs-strike",
}

// verbHandlers is the dispatch table: every verb to its gate's Run. The
// table IS the vocabulary — unknownVerb answers for anything not in it.
//
//nolint:gochecknoglobals // the fixed dispatch table, same class as the contract tables
var verbHandlers = map[string]func([]string) int{
	verbRun:       battery.Run,
	verbFreshness: docsfreshness.Run,
	verbAnnotate:  docsannotate.Run,
	verbRouting:   docsrouting.Run,
	verbCadence:   docscadence.Run,
	verbStrike:    docsstrike.Run,
}

// verbOrder is the display order of the verbs in the root usage.
//
//nolint:gochecknoglobals // the fixed display order, same class as the contract tables
var verbOrder = []string{verbRun, verbFreshness, verbAnnotate, verbRouting, verbCadence, verbStrike}

// verbOverview is each verb's one-line purpose for the root usage; the
// detailed flag help lives in the gate packages' Usage.
//
//nolint:gochecknoglobals // the fixed verb one-liners, same class as the contract tables
var verbOverview = map[string]string{
	verbRun:       "all five gates as one ordered verdict matrix (the battery)",
	verbFreshness: "the newest report's §f items must be harvested into the living docs",
	verbAnnotate:  "struck items must carry honest, verifiable verdicts",
	verbRouting:   "open actionable rows outside §f must be routed to a living doc",
	verbCadence:   "TODO_LIST purge and report archive thresholds",
	verbStrike:    "strike one TODO_LIST row (dry-run by default)",
}

// verbUsage wires each verb to its package's Usage text — the same
// packages the dispatch table calls, so the help and the behavior
// cannot fork.
//
//nolint:gochecknoglobals // the fixed verb-to-usage map, same class as the contract tables
var verbUsage = map[string]string{
	verbRun:       battery.Usage,
	verbFreshness: docsfreshness.Usage,
	verbAnnotate:  docsannotate.Usage,
	verbRouting:   docsrouting.Usage,
	verbCadence:   docscadence.Usage,
	verbStrike:    docsstrike.Usage,
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

	// Help-shaped invocations: `docs-gates --help`, `-h`, or the `help`
	// verb (optionally naming one verb to explain).
	if len(argv) > 0 && isHelpWord(argv[0]) {
		return runHelp(argv[1:])
	}

	if len(argv) == 0 {
		printUsageError()

		return gatecli.ExitUsage
	}

	verb, rest := argv[0], argv[1:]

	handler, ok := verbHandlers[verb]
	if !ok {
		unknownVerb(verb)
		printUsageError()

		return gatecli.ExitUsage
	}

	// --version MUST run before any parser prologue: the gate parsers
	// reject unknown flags and would die on it.
	if gatecli.HandleVersion(rest, displayName[verb], version) {
		return gatecli.ExitClean
	}

	// --help/-h ride the same slot: intercepted before the gate parsers
	// (which reject unknown flags) and after --version (the probe
	// scripts' contract).
	if gatecli.WantsHelp(rest) {
		gatecli.Outf("%s", verbUsage[verb])

		return gatecli.ExitClean
	}

	return handler(rest)
}

// isHelpWord reports whether arg is one of the three root help shapes.
func isHelpWord(arg string) bool {
	return arg == "--help" || arg == "-h" || arg == "help"
}

// runHelp serves `docs-gates help [verb]` (and `--help`/`-h` with a
// verb after them): no verb named prints the root usage, one verb named
// prints that verb's Usage. A name that is not a verb is a usage error
// carrying the same suggestion the dispatcher gives.
func runHelp(rest []string) int {
	asked := ""

	for _, arg := range rest {
		if isHelpWord(arg) || arg == "--version" {
			continue
		}

		asked = arg

		break
	}

	if asked == "" {
		gatecli.Outf("%s", rootUsage())

		return gatecli.ExitClean
	}

	usage, ok := verbUsage[asked]
	if !ok {
		unknownVerb(asked)
		printUsageError()

		return gatecli.ExitUsage
	}

	gatecli.Outf("%s", usage)

	return gatecli.ExitClean
}

// unknownVerb prints the refusal for a name that is not a verb, with a
// did-you-mean suggestion when one is close enough.
func unknownVerb(typed string) {
	gatecli.Errf("unknown verb %q\n", typed)

	if hint, ok := suggestVerb(typed); ok {
		gatecli.Errf("did you mean %q?\n", hint)
	}
}

// printUsageError prints the usage footer every usage-class refusal
// ends with — the path to discovery.
func printUsageError() {
	gatecli.Errln("usage: docs-gates <verb> [args] — `docs-gates help` lists the verbs")
}

// rootUsage is the bare `docs-gates help` text: what the tool is, every
// verb with its one-liner, and where the knobs live.
func rootUsage() string {
	var b strings.Builder

	b.WriteString("docs-gates — the docs-health battery: five doc gates, one binary\n\n")
	b.WriteString("usage: docs-gates <verb> [args]\n\nverbs:\n")

	for _, verb := range verbOrder {
		fmt.Fprintf(&b, "  %-10s %s\n", verb, verbOverview[verb])
	}

	b.WriteString("\ndocs-gates help <verb> (or <verb> --help) explains one verb's flags.\n")
	b.WriteString("docs-gates --version prints the version. The knobs (corpus paths,\n")
	b.WriteString("thresholds, ratchets, skip_gates) live in the repo's .docs-gates.json.\n\n")
	b.WriteString("exit codes: 0 clean, 1 findings, 2 usage/environment error\n")

	return b.String()
}

// suggestVerb names the verb the user probably meant: a unique prefix
// wins first (fresh → freshness), else any verb within edit distance
// suggestMaxEditDistance (freshnness → freshness). Empty when nothing
// is close enough.
func suggestVerb(typed string) (string, bool) {
	if typed == "" {
		return "", false
	}

	want := strings.ToLower(typed)

	prefix := ""

	for _, verb := range verbOrder {
		if strings.HasPrefix(verb, want) {
			if prefix != "" {
				return "", false // ambiguous prefix — do not guess
			}

			prefix = verb
		}
	}

	if prefix != "" {
		return prefix, true
	}

	for _, verb := range verbOrder {
		if editDistance(want, verb) <= suggestMaxEditDistance {
			return verb, true
		}
	}

	return "", false
}

// editDistance is the Levenshtein distance between two lowercase words.
// The DP rows are append-built: length expressions the makezero lint
// cannot prove non-zero are exactly the indexing class it exists for.
func editDistance(a, b string) int {
	prev := make([]int, 0, len(b)+1)

	for j := 0; j <= len(b); j++ {
		prev = append(prev, j)
	}

	for i := 1; i <= len(a); i++ {
		curr := make([]int, 0, len(b)+1)
		curr = append(curr, i)

		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}

			curr = append(curr, min(prev[j]+1, curr[j-1]+1, prev[j-1]+cost))
		}

		prev = curr
	}

	return prev[len(b)]
}
