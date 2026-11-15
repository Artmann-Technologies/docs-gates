package docsrows

import (
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// TestClassifyRowMatchesWriterVocabulary pins the row-state machine against
// the cases the writer tool's own self-test asserts (annotate.go selfTest):
// escaped pipes stay inside one cell, a routed verdict resolves a row
// without a strike, and an open row stays CLEAN.
func TestClassifyRowMatchesWriterVocabulary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		line string
		want rowState
	}{
		{"escaped pipe keeps one cell", `| x | example: ` + "`" + `\| - \| --- \|` + "`" + ` inline |`, stateClean},
		{"routed arrow", `| ~~2~~ | **→ routed** to TODO | ` + "`x.md`" + ` |`, stateRouted},
		{"plain routed arrow", `| 3 | item — → routed | ` + "`x.md`" + ` |`, stateRouted},
		{"open row", `| 4 | still open | ` + "`x.md`" + ` |`, stateClean},
		{"fully struck", `| ~~4~~ | ~~still open~~ ~~done~~ | ~~x.md~~ |`, stateStruck},
		{"half struck", `| ~~4~~ | still open | ` + "`x.md`" + ` |`, statePartial},
	}

	for _, tc := range cases {
		if got := classifyRow(tc.line); got != tc.want {
			t.Errorf("%s: classifyRow = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestIsSeparator(t *testing.T) {
	t.Parallel()

	if !isSeparator("| --- | :---: |") {
		t.Error("a markdown separator must be detected")
	}

	if isSeparator("| --- | value |") {
		t.Error("a data row must not be a separator")
	}
}

// TestCompletenessModes covers the three shapes the gate exists for: a
// partial row, a mixed table (some struck, some open), and the --strict
// wholly-unstruck class the default mode deliberately passes.
//
//nolint:paralleltest // gatecli.CaptureOutput swaps process-global streams
func TestCompletenessModes(t *testing.T) {
	cases := []struct {
		name    string
		argv    []string
		code    int
		want    []string
		errWant []string
	}{
		{
			name: "clean corpus",
			argv: []string{"testdata/clean.md"},
			code: gatecli.ExitClean,
			want: []string{"rows: 1 file(s) clean"},
		},
		{
			name: "partial row",
			argv: []string{"testdata/partial.md"},
			code: gatecli.ExitFindings,
			want: []string{
				"partial.md: table at line 3 INCOMPLETE (0/2 rows struck)",
				"  line 5: PARTIAL row:",
			},
			errWant: []string{"rows: 1/1 file(s) incomplete"},
		},
		{
			name: "mixed table",
			argv: []string{"testdata/mixed.md"},
			code: gatecli.ExitFindings,
			want: []string{
				"mixed.md: table at line 3 INCOMPLETE (1/2 rows struck)",
				"  line 6: CLEAN row in a struck table",
			},
		},
		{
			name: "strict untouched",
			argv: []string{"testdata/mixed.md", "--strict"},
			code: gatecli.ExitFindings,
			want: []string{
				"mixed.md: table at line 8 UNTOUCHED (0/2 rows struck)",
			},
		},
		{
			name: "ledger exempt",
			argv: []string{"testdata/ledger.md"},
			code: gatecli.ExitClean,
			want: []string{"ledger.md: table at line 3 LEDGER-EXEMPT (check-rows:ledger marker)"},
		},
		{
			name: "bare rows",
			argv: []string{"testdata/bare.md", "--bare"},
			code: gatecli.ExitFindings,
			want: []string{"bare.md:5: BARE row:"},
			errWant: []string{"rows --bare: 1/1 file(s) incomplete"},
		},
	}

	for _, tc := range cases {
		out := gatecli.CaptureOutput(func() int { return Run(tc.argv) })

		if out.Code != tc.code {
			t.Errorf("%s: code = %d, want %d (stdout %q, stderr %q)", tc.name, out.Code, tc.code, out.Stdout, out.Stderr)
		}

		for _, want := range tc.want {
			if !strings.Contains(out.Stdout, want) {
				t.Errorf("%s: stdout missing %q:\n%s", tc.name, want, out.Stdout)
			}
		}

		for _, want := range tc.errWant {
			if !strings.Contains(out.Stderr, want) {
				t.Errorf("%s: stderr missing %q:\n%s", tc.name, want, out.Stderr)
			}
		}
	}
}

// TestUsageRefusals proves a malformed invocation and an unreadable file are
// usage/environment errors (2), never quiet clean exits.
//
//nolint:paralleltest // gatecli.CaptureOutput swaps process-global streams
func TestUsageRefusals(t *testing.T) {
	cases := []struct {
		name string
		argv []string
		want string
	}{
		{"no files", nil, "rows needs at least one file path"},
		{"unknown flag", []string{"--nope", "testdata/clean.md"}, "unknown flag: --nope"},
		{"missing file", []string{"testdata/does-not-exist.md"}, "FATAL: reading"},
	}

	for _, tc := range cases {
		out := gatecli.CaptureOutput(func() int { return Run(tc.argv) })

		if out.Code != gatecli.ExitUsage {
			t.Errorf("%s: code = %d, want %d", tc.name, out.Code, gatecli.ExitUsage)
		}

		if !strings.Contains(out.Stderr, tc.want) {
			t.Errorf("%s: stderr missing %q:\n%s", tc.name, tc.want, out.Stderr)
		}
	}
}
