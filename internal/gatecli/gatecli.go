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
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
)

// The gate-CLI exit-code contract (see the package comment).
const (
	ExitClean    = 0
	ExitFindings = 1
	ExitUsage    = 2
)

// Duration-rendering precisions: milliseconds under a second, tenths
// under a minute, whole seconds past it.
const (
	subSecondPrecision = 10 * time.Millisecond
	subMinutePrecision = 100 * time.Millisecond
)

// FormatDuration renders a wall time for progress lines and verdict
// matrices (the battery heartbeat, the audit-all replay estimate). The
// parity harness normalizes the emitted values to a literal token, so
// goldens pin the SHAPE of a timing-bearing line, never the value.
func FormatDuration(d time.Duration) string {
	switch {
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

// Truncate returns s unchanged when it fits within n bytes, else its first
// n bytes (finding snippets are display-only). The cut never splits a
// UTF-8 sequence: when byte n would land mid-rune, the cut backs off to
// the rune boundary, so the truncated snippet stays valid UTF-8.
func Truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}

	cut := n
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}

	return s[:cut]
}

// Outf prints a formatted verdict line to stdout; a stdout write failure (a
// closed pipe, a full disk) is deliberately not a gate finding — the exit
// code carries the verdict. One home for that policy so the gates cannot
// drift into unchecked fmt.Fprint* calls (errcheck) or forbidden fmt.Print*
// (forbidigo). Outln, Errf and Errln are the newline and stderr variants.
func Outf(format string, args ...any) {
	_, _ = fmt.Fprintf(
		os.Stdout,
		format,
		args...) //nolint:erraudit // policy above: a stdout write failure is not a gate finding
}

func Outln(args ...any) {
	_, _ = fmt.Fprintln(os.Stdout, args...) //nolint:erraudit // policy above
}

func Errf(format string, args ...any) {
	_, _ = fmt.Fprintf(
		os.Stderr,
		format,
		args...) //nolint:erraudit // policy above: a stderr write failure is not a gate finding
}

func Errln(args ...any) {
	_, _ = fmt.Fprintln(os.Stderr, args...) //nolint:erraudit // policy above
}

// CapturedOutput is what a Run-shaped function emitted plus its exit code.
type CapturedOutput struct {
	Stdout string
	Stderr string
	Code   int
}

// CaptureOutput redirects os.Stdout and os.Stderr around fn and returns
// everything it emitted plus its exit code (the gates write via the
// process-level streams; tests capture them the way the original Python
// harness captured subprocess output).
//
// The swap is a PROCESS-GLOBAL mutation: every caller must stay out of
// t.Parallel() — mixing this helper with parallel stdout writers races
// the whole suite's output. The serialization rule lives here so every
// consuming suite inherits it from one place (TODO stdout-capture,
// 2026-10-16: one shared capture instead of one hand-rolled per suite).
func CaptureOutput(fn func() int) CapturedOutput {
	outR, outW, _ := os.Pipe()
	errR, errW, _ := os.Pipe()
	origOut, origErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outW, errW
	code := fn()
	os.Stdout, os.Stderr = origOut, origErr
	_ = outW.Close()
	_ = errW.Close()
	outData, _ := io.ReadAll(outR)
	errData, _ := io.ReadAll(errR)
	_ = outR.Close()
	_ = errR.Close()

	return CapturedOutput{Stdout: string(outData), Stderr: string(errData), Code: code}
}
