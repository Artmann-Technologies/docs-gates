package docsarchive

import (
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// TestUsageDocumentsEveryAcceptedFlag pins the Usage text against the
// parser, both directions: every flag the verb accepts must appear in
// the help, and the full documented set must parse cleanly together.
func TestUsageDocumentsEveryAcceptedFlag(t *testing.T) {
	t.Parallel()

	documented := []string{"--apply"}

	for _, flag := range documented {
		if !strings.Contains(Usage, flag) {
			t.Errorf("Usage does not document accepted flag %s", flag)
		}
	}

	if _, _, err := gatecli.Parse([]string{".", "--apply"}, "--apply"); err != nil {
		t.Errorf("the documented flag set must parse: %v", err)
	}

	if _, _, err := gatecli.Parse([]string{".", "--bogus"}, "--apply"); err == nil {
		t.Error("undocumented flags must refuse")
	}
}

// TestUsageNamesEveryExitCode pins the exit-code contract to the help:
// a user reading only the Usage must learn when the verb answers 0, 1,
// or 2.
func TestUsageNamesEveryExitCode(t *testing.T) {
	t.Parallel()

	for _, line := range []string{
		"0 plan printed or applied",
		"1 unused",
		"2 usage or environment error",
	} {
		if !strings.Contains(Usage, line) {
			t.Errorf("Usage does not carry the exit-code line %q:\n%s", line, Usage)
		}
	}
}
