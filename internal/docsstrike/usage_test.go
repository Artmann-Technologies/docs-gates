package docsstrike

import (
	"strings"
	"testing"
)

// TestUsageDocumentsEveryAcceptedFlag pins the Usage text against
// ParseArgs, both directions: every flag the parser accepts must appear
// in the help, and the full documented set must parse cleanly together.
func TestUsageDocumentsEveryAcceptedFlag(t *testing.T) {
	t.Parallel()

	documented := []string{"--row", "--verdict", "--evidence", "--status", "--file", "--apply"}

	for _, flag := range documented {
		if !strings.Contains(Usage, flag) {
			t.Errorf("Usage does not document accepted flag %s", flag)
		}
	}

	argv := []string{
		".", "--row", "some task", "--verdict", "the evidence",
		"--evidence", "e2e log", "--status", "✅ `DONE`", "--file", "TODO_LIST.md", "--apply",
	}

	if _, err := ParseArgs(argv); err != nil {
		t.Errorf("ParseArgs rejects the documented flag set: %v", err)
	}

	if _, err := ParseArgs([]string{".", "--row", "x", "--verdict", "v", "--bogus"}); err == nil {
		t.Error("ParseArgs accepts an undocumented flag")
	}
}
