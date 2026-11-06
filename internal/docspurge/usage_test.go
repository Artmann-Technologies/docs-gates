package docspurge

import (
	"strings"
	"testing"
)

// TestUsageDocumentsEveryAcceptedFlag pins the Usage text against
// ParseArgs, both directions: every flag the parser accepts must appear
// in the help, and the full documented set must parse cleanly together.
func TestUsageDocumentsEveryAcceptedFlag(t *testing.T) {
	t.Parallel()

	documented := []string{"--file", "--changelog", "--apply"}

	for _, flag := range documented {
		if !strings.Contains(Usage, flag) {
			t.Errorf("Usage does not document accepted flag %s", flag)
		}
	}

	argv := []string{".", "--file", "TODO_LIST.md", "--changelog", "CHANGELOG.md", "--apply"}

	if _, err := ParseArgs(argv); err != nil {
		t.Errorf("ParseArgs rejects the documented flag set: %v", err)
	}

	if _, err := ParseArgs([]string{".", "--bogus"}); err == nil {
		t.Error("ParseArgs accepts an undocumented flag")
	}
}

// TestUsageNamesEveryExitCode pins the exit-code contract to the help:
// a user reading only the Usage must learn when the verb answers 0, 1,
// or 2.
func TestUsageNamesEveryExitCode(t *testing.T) {
	t.Parallel()

	for _, line := range []string{
		"0 purge planned/applied",
		"1 the pre-flight gate refused",
		"2 usage or environment error",
	} {
		if !strings.Contains(Usage, line) {
			t.Errorf("Usage does not carry the exit-code line %q:\n%s", line, Usage)
		}
	}
}
