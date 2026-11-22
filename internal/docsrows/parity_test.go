package docsrows

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// writerSource is the docs-health skill's single-file writer tool. It is a
// local fan-out file, so this test SKIPS on a machine without the skill
// (CI); where it exists, it is the parity oracle for the read-only port.
const writerSource = "/home/lars/.config/crush/skills/docs-health/assets/annotate.go"

// parityFixtures are the shapes whose rendering must stay byte-identical:
// clean, partial, mixed/untouched, ledger-exempt, and bare rows.
var parityFixtures = []string{
	"testdata/clean.md",
	"testdata/partial.md",
	"testdata/mixed.md",
	"testdata/ledger.md",
	"testdata/bare.md",
}

// TestWriterToolParity runs the original writer tool and this port over the
// same fixtures and requires identical finding lines and exit codes. The
// summary lines are excluded on purpose: the port names its own verb, which
// is the one intended difference.
//
//nolint:paralleltest // gatecli.CaptureOutput swaps process-global streams
func TestWriterToolParity(t *testing.T) {
	if _, err := os.Stat(writerSource); err != nil {
		t.Skipf("docs-health writer tool not present at %s (local fan-out only)", writerSource)
	}

	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not on PATH: cannot run the writer oracle")
	}

	cases := []struct {
		name   string
		ours   []string
		theirs []string
	}{
		{
			name:   "check-rows",
			ours:   parityFixtures,
			theirs: append([]string{"check-rows"}, parityFixtures...),
		},
		{
			name:   "check-rows --strict",
			ours:   append([]string{"--strict"}, parityFixtures...),
			theirs: append([]string{"check-rows", "--strict"}, parityFixtures...),
		},
		{
			name:   "bare-rows",
			ours:   append([]string{"--bare"}, parityFixtures...),
			theirs: append([]string{"bare-rows"}, parityFixtures...),
		},
	}

	for _, tc := range cases {
		out := gatecli.CaptureOutput(func() int { return Run(tc.ours) })

		got, code := runWriter(t, tc.theirs)

		if out.Code != code {
			t.Errorf("%s: exit = %d, writer = %d", tc.name, out.Code, code)
		}

		if normalizeFindings(out.Stdout) != got {
			t.Errorf(
				"%s: findings differ from the writer tool\n--- port ---\n%s\n--- writer ---\n%s",
				tc.name, normalizeFindings(out.Stdout), got,
			)
		}
	}
}

// runWriter executes the writer tool and returns its normalised stdout and
// exit code. An unexpected launch failure is a test failure, not a skip.
func runWriter(t *testing.T, argv []string) (string, int) {
	t.Helper()

	cmd := exec.CommandContext(context.Background(), "go", append([]string{"run", writerSource}, argv...)...)

	raw, err := cmd.Output()
	if err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			return normalizeFindings(string(raw)), exitErr.ExitCode()
		}

		t.Fatalf("running writer oracle: %v", err)
	}

	return normalizeFindings(string(raw)), 0
}

// normalizeFindings keeps the verdict lines and drops each tool's summary
// line, the only place the two contracts intentionally differ.
func normalizeFindings(stdout string) string {
	kept := make([]string, 0, strings.Count(stdout, "\n")+1)

	for line := range strings.SplitSeq(stdout, "\n") {
		switch {
		case strings.HasPrefix(line, "check-rows:"),
			strings.HasPrefix(line, "bare-rows:"),
			strings.HasPrefix(line, "rows:"),
			strings.HasPrefix(line, "rows --bare:"):
			continue
		}

		kept = append(kept, line)
	}

	return strings.TrimSpace(strings.Join(kept, "\n"))
}
