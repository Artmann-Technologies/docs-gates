// Package gatecli is the CLI scaffolding shared by the gate commands
// (the docs-gates battery verbs — freshness, annotate, routing, cadence,
// strike — and the standalone legacy binaries that exec them): boolean
// flag parsing, repo-root resolution, output helpers, and finding-snippet
// truncation. One home so the gate CLIs cannot drift apart on argument
// semantics or error wording.
//
// Exit-code contract of the consumers: 0 = clean, 1 = findings, 2 =
// usage/environment error — the values this package's failures feed.
package gatecli

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
)

// The gate-CLI exit-code contract (see the package comment).
const (
	ExitClean    = 0
	ExitFindings = 1
	ExitUsage    = 2
)

// Duration-rendering precisions: exact microseconds and 100µs steps
// under 10ms, milliseconds under a second, tenths under a minute, whole
// seconds past it.
const (
	microPrecision     = 1 * time.Microsecond
	subMilliPrecision  = 100 * time.Microsecond
	subSecondPrecision = 10 * time.Millisecond
	subMinutePrecision = 100 * time.Millisecond
)

// FormatDuration renders a wall time for progress lines and verdict
// matrices (the battery heartbeat, the audit-all replay estimate). The
// parity harness normalizes the emitted values to a literal token, so
// goldens pin the SHAPE of a timing-bearing line, never the value.
// Sub-second durations render at honest precision instead of rounding
// to a lying "0s" — a gate that ran in 3µs did not run in no time, and
// a sub-10ms estimate must stay legible as a cost signal (the <10ms
// bucket was honest but useless there).
func FormatDuration(d time.Duration) string {
	switch {
	case d < microPrecision:
		return "<1µs"
	case d < time.Millisecond:
		return d.Round(microPrecision).String()
	case d < subSecondPrecision:
		return d.Round(subMilliPrecision).String()
	case d < time.Second:
		return d.Round(subSecondPrecision).String()
	case d < time.Minute:
		return d.Round(subMinutePrecision).String()
	default:
		return d.Round(time.Second).String()
	}
}

// notDirectoryError reports a repo-root argument that does not resolve to
// an existing directory (a usage error: the gate CLIs exit 2 on it).
type notDirectoryError struct{ path string }

func (e notDirectoryError) Error() string { return e.path + " is not a directory" }

// unknownFlagsError reports flags outside the command's allowed set.
type unknownFlagsError struct{ flags []string }

func (e unknownFlagsError) Error() string { return "unknown flag(s): " + strings.Join(e.flags, " ") }

// NewUnknownFlagsError builds the unknown-flag error for consumers that
// call ParseFlags directly (e.g. to read positionals beyond the repo root)
// and still want the exact Parse-style FATAL message.
func NewUnknownFlagsError(flags []string) error { return unknownFlagsError{flags} }

// WantsHelp reports whether argv asks for help (--help or -h anywhere).
// The gate parsers reject both as unknown flags, so the dispatcher must
// intercept them BEFORE the gate prologue — the same slot --version rides
// (HandleVersion).
func WantsHelp(args []string) bool {
	return slices.Contains(args, "--help") || slices.Contains(args, "-h")
}

// FatalUsage prints the FATAL line for a usage-class failure plus the
// pointer to the verb's help, and returns ExitUsage — one shape so every
// gate's parse refusal ends with a path to discovery instead of a dead
// end. Config/environment failures (LoadConfig, unreadable files) stay
// on the bare `FATAL: %v` shape: their fix is not in the argv.
func FatalUsage(verb string, err error) int {
	Errf("FATAL: %v\n", err)
	Errf("usage: docs-gates %s --help\n", verb)

	return ExitUsage
}

// HandleVersion prints "<name> <version>" and reports true when --version
// is among the args; the caller returns immediately in that case. It MUST
// run before any flag/gatecli.Parse prologue — the gate parsers reject
// unknown flags and --version would die as one. The version value is the
// caller's main-package var, stamped by the flake build (ldflags
// -X main.version); "dev" in bare go build/run.
func HandleVersion(args []string, name, version string) bool {
	if !slices.Contains(args, "--version") {
		return false
	}

	_, _ = fmt.Fprintf(os.Stdout, "%s %s\n", name, version)

	return true
}

// ParseFlags splits argv into boolean `--flags` and positional arguments.
// Order is irrelevant: flags may appear anywhere (before, between, after
// positionals) and all land in the same map. Flags not in allowed are
// reported in the returned unknown slice, sorted so the caller's rejection
// message (and any test) is deterministic.
func ParseFlags(argv []string, allowed ...string) (map[string]bool, []string, []string) {
	allowedSet := make(map[string]bool, len(allowed))
	for _, a := range allowed {
		allowedSet[a] = true
	}

	flags := map[string]bool{}

	var (
		positional []string
		unknown    []string
	)

	for _, a := range argv {
		if strings.HasPrefix(a, "--") {
			flags[a] = true

			continue
		}

		positional = append(positional, a)
	}

	for f := range flags {
		if !allowedSet[f] {
			unknown = append(unknown, f)
		}
	}

	sort.Strings(unknown)

	return flags, positional, unknown
}

// Parse is the gate-CLI prologue: split argv into known boolean flags and
// positionals, then resolve the repo root from the first positional. Any
// caller-input failure — an unknown flag, or a root that is not an existing
// directory — comes back as one error whose message is ready for the
// consumer's `FATAL: %v` line (exit 2).
func Parse(argv []string, allowed ...string) (map[string]bool, string, error) {
	flags, positional, unknown := ParseFlags(argv, allowed...)
	if len(unknown) > 0 {
		return nil, "", unknownFlagsError{unknown}
	}

	root, err := ResolveRoot(positional)
	if err != nil {
		return nil, "", err
	}

	return flags, root, nil
}

// valuedFlagError reports a valued `--flag=N` whose value is not a
// non-negative integer (a usage error: the gate CLIs exit 2 on it).
type valuedFlagError struct {
	flag  string
	value string
}

func (e valuedFlagError) Error() string {
	return "--" + e.flag + " wants a non-negative integer, got " + strconv.Quote(e.value)
}

// CutIntFlag splits a valued `--flag=N` off argv BEFORE ParseFlags runs (a
// valued flag would die there as an unknown boolean). Every occurrence is
// split off and the last value wins. The value must be a non-negative
// integer — anything else is the returned error, whose message is ready
// for the caller's `FATAL: %v` line. rest is non-nil on success and nil
// on error, so a nil rest can double as the caller's failure sentinel.
func CutIntFlag(args []string, flag string) (int, []string, error) {
	value := -1

	rest := make([]string, 0, len(args))

	prefix := "--" + flag + "="

	for _, a := range args {
		v, ok := strings.CutPrefix(a, prefix)
		if !ok {
			rest = append(rest, a)

			continue
		}

		n, parseErr := strconv.Atoi(v)
		if parseErr != nil || n < 0 {
			return -1, nil, valuedFlagError{flag, v}
		}

		value = n
	}

	return value, rest, nil
}

// LoadConfig is the gate-verb config prologue: load the repo's
// .docs-gates.json, print non-fatal unknown-key warnings to stderr as
// `config: ...` lines, and return the error for the caller's `FATAL: %v`
// line (exit 2). Warnings print before the error is surfaced, so a config
// that is BOTH unreadable and mis-keyed still announces what it announced.
// One home for the shape so the gates cannot drift apart on warning
// wording or drop warnings on the FATAL path.
func LoadConfig(root string) (config.Config, error) {
	cfg, warnings, err := config.Load(root)
	for _, w := range warnings {
		Errf("config: %s\n", w)
	}

	return cfg, err //nolint:wrapcheck // config.Load's error is pre-formatted for the caller's FATAL line (config.Load doc); wrapping would change the historical usage text
}

// ResolveRoot resolves the repo root: the FIRST positional argument when
// given (later positionals are ignored), else the working directory, made
// absolute and verified to be an existing directory.
func ResolveRoot(positional []string) (string, error) {
	root := "."
	if len(positional) > 0 {
		root = positional[0]
	}

	abs, err := filepath.Abs(root)
	if err != nil {
		return "", notDirectoryError{root}
	}

	if info, err := os.Stat(abs); err != nil || !info.IsDir() {
		return "", notDirectoryError{abs}
	}

	return abs, nil
}
