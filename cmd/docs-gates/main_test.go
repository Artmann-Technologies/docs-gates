package main

import (
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// The dispatch tests assert exit codes and output shapes for the help
// and usage-error surfaces. Serial on purpose: gatecli.CaptureOutput
// swaps process-global streams.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestHelpSurfaces(t *testing.T) {
	cases := []struct {
		name string
		argv []string
		code int
		want string // must appear in stdout
	}{
		{"root long", []string{"--help"}, gatecli.ExitClean, "usage: docs-gates <verb>"},
		{"root short", []string{"-h"}, gatecli.ExitClean, "exit codes:"},
		{"help verb", []string{"help"}, gatecli.ExitClean, "verbs:"},
		{"help named verb", []string{"help", "strike"}, gatecli.ExitClean, "usage: docs-strike"},
		{"help flag then verb", []string{"--help", "cadence"}, gatecli.ExitClean, "usage: docs-cadence"},
		{"verb long help", []string{"freshness", "--help"}, gatecli.ExitClean, "usage: docs-freshness"},
		{"verb short help", []string{"annotate", "-h"}, gatecli.ExitClean, "usage: docs-annotate"},
		{"run help", []string{"run", "--help"}, gatecli.ExitClean, "usage: docs-gates run"},
		{"routing help", []string{"routing", "--help"}, gatecli.ExitClean, "usage: docs-routing"},
	}

	for _, tc := range cases {
		out := gatecli.CaptureOutput(func() int { return run(tc.argv) })

		if out.Code != tc.code {
			t.Errorf("%s: code = %d, want %d (stderr %q)", tc.name, out.Code, tc.code, out.Stderr)
		}

		if !strings.Contains(out.Stdout, tc.want) {
			t.Errorf("%s: stdout missing %q:\n%s", tc.name, tc.want, out.Stdout)
		}
	}
}

// Usage-class refusals: exit 2 on stderr, with a did-you-mean
// suggestion when one verb is close enough and silence when none is.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestUsageErrors(t *testing.T) {
	cases := []struct {
		name      string
		argv      []string
		errText   string
		suggest   string // empty = must NOT suggest
		suggested bool
	}{
		{name: "no args", argv: nil, errText: "usage: docs-gates <verb> [args]"},
		{
			name:      "unknown verb with suggestion",
			argv:      []string{"fresh"},
			errText:   `unknown verb "fresh"`,
			suggest:   `"freshness"`,
			suggested: true,
		},
		{
			name:      "unknown verb levenshtein",
			argv:      []string{"freshnness"},
			errText:   `unknown verb "freshnness"`,
			suggest:   `"freshness"`,
			suggested: true,
		},
		{name: "unknown verb without suggestion", argv: []string{"zzzzzz"}, errText: `unknown verb "zzzzzz"`},
		{name: "help unknown verb", argv: []string{"help", "zzzzzz"}, errText: `unknown verb "zzzzzz"`},
		{
			name:      "help typo suggests",
			argv:      []string{"help", "ann"},
			errText:   `unknown verb "ann"`,
			suggest:   `"annotate"`,
			suggested: true,
		},
	}

	for _, tc := range cases {
		out := gatecli.CaptureOutput(func() int { return run(tc.argv) })

		if out.Code != gatecli.ExitUsage {
			t.Errorf("%s: code = %d, want %d", tc.name, out.Code, gatecli.ExitUsage)
		}

		if !strings.Contains(out.Stderr, tc.errText) {
			t.Errorf("%s: stderr missing %q:\n%s", tc.name, tc.errText, out.Stderr)
		}

		if tc.suggested && !strings.Contains(out.Stderr, "did you mean "+tc.suggest) {
			t.Errorf("%s: stderr missing suggestion %s:\n%s", tc.name, tc.suggest, out.Stderr)
		}

		if !tc.suggested && strings.Contains(out.Stderr, "did you mean") {
			t.Errorf("%s: stderr carries an unwanted suggestion:\n%s", tc.name, out.Stderr)
		}
	}
}

// --version keeps its precedence over --help: probe scripts ask
// `docs-freshness --version` and must never see help text.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestVersionPrecedenceOverHelp(t *testing.T) {
	out := gatecli.CaptureOutput(func() int { return run([]string{"freshness", "--version", "--help"}) })

	if out.Code != gatecli.ExitClean {
		t.Fatalf("code = %d, want clean", out.Code)
	}

	if !strings.HasPrefix(out.Stdout, "docs-freshness ") {
		t.Fatalf("stdout = %q, want the version line", out.Stdout)
	}
}

// Every verb in the dispatch tables must reach its gate through the
// switch: a verb the maps know but the switch misses would fall to the
// unreachable return and exit 2 with EMPTY stderr, silently. The gate's
// own FATAL on a bogus root proves the switch dispatched.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestEveryVerbReachesItsGate(t *testing.T) {
	for _, verb := range verbOrder {
		argv := []string{verb, "/nonexistent-docs-gates-root"}

		out := gatecli.CaptureOutput(func() int { return run(argv) })

		if out.Code != gatecli.ExitUsage {
			t.Errorf("%s: code = %d, want usage error over a bogus root", verb, out.Code)
		}

		if !strings.Contains(out.Stderr, "FATAL:") {
			t.Errorf("%s: stderr = %q, want the gate's own FATAL (proves switch dispatch)", verb, out.Stderr)
		}
	}
}

// The verb tables (order, legacy names, one-liners, usage texts) must
// cover exactly the same verbs — one source of truth, no orphan rows.
func TestVerbTablesAgree(t *testing.T) {
	t.Parallel()

	if len(verbOrder) != len(displayName) || len(verbOrder) != len(verbOverview) || len(verbOrder) != len(verbUsage) {
		t.Fatalf("table sizes disagree: order=%d names=%d overview=%d usage=%d",
			len(verbOrder), len(displayName), len(verbOverview), len(verbUsage))
	}

	for _, verb := range verbOrder {
		if _, ok := displayName[verb]; !ok {
			t.Errorf("verb %q missing from displayName", verb)
		}

		if _, ok := verbOverview[verb]; !ok {
			t.Errorf("verb %q missing from verbOverview", verb)
		}

		if usage, ok := verbUsage[verb]; !ok || usage == "" {
			t.Errorf("verb %q missing from verbUsage", verb)
		}
	}
}

// suggestVerb: unique prefix wins, ambiguity refuses to guess, edit
// distance <= 2 catches typos, everything else stays silent.
func TestSuggestVerb(t *testing.T) {
	t.Parallel()

	cases := []struct {
		typed string
		want  string
	}{
		{"fresh", "freshness"},
		{"ru", "run"},
		{"st", "strike"},
		{"freshnness", "freshness"},
		{"annnotate", "annotate"},
		{"stike", "strike"},
		{"ROUTING", "routing"},
		{"r", ""},    // prefix of run AND routing: ambiguous
		{"", ""},     // nothing typed
		{"xyz", ""},  // too far from every verb
		{"zzzz", ""}, // even-ish length but no match
	}

	for _, tc := range cases {
		got, ok := suggestVerb(tc.typed)

		if tc.want == "" {
			if ok {
				t.Errorf("suggestVerb(%q) = %q, want no suggestion", tc.typed, got)
			}

			continue
		}

		if !ok || got != tc.want {
			t.Errorf("suggestVerb(%q) = %q,%v want %q", tc.typed, got, ok, tc.want)
		}
	}
}

// help takes ONE verb: arguments beyond it are never silently eaten —
// they are named on stderr while the asked verb's usage still prints,
// and the exit stays clean (a typo in the tail must not kill the help).
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestHelpExtraArgsWarn(t *testing.T) {
	out := gatecli.CaptureOutput(func() int { return run([]string{"help", "strike", "cadence"}) })

	if out.Code != gatecli.ExitClean {
		t.Fatalf("extra help args warn but stay clean, got rc %d", out.Code)
	}

	if !strings.Contains(out.Stdout, "usage: docs-strike") {
		t.Fatalf("asked verb's usage must still print, got:\n%s", out.Stdout)
	}

	if !strings.Contains(out.Stderr, `ignoring extra argument(s): "cadence"`) {
		t.Fatalf("the ignored argument must be named on stderr, got: %s", out.Stderr)
	}
}
