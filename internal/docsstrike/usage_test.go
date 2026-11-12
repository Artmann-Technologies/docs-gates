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

	documented := []string{
		"--row", "--verdict", "--evidence", "--status", "--file", "--apply",
		"--check-verdict-outside",
	}

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

	// CHECK MODE takes no strike inputs: neither --row nor --verdict is
	// required when the mode flag rides along.
	if opts, err := ParseArgs([]string{".", "--check-verdict-outside"}); err != nil || !opts.CheckVerdictOutside {
		t.Errorf("ParseArgs rejects the documented check mode: opts=%+v err=%v", opts, err)
	}

	if _, err := ParseArgs([]string{".", "--row", "x", "--verdict", "v", "--bogus"}); err == nil {
		t.Error("ParseArgs accepts an undocumented flag")
	}

	if _, err := ParseArgs([]string{".", "--check-verdict-outside", "--bogus"}); err == nil {
		t.Error("ParseArgs accepts an undocumented flag in check mode")
	}
}

// TestUsageDefaultStatusIsTheConst pins the documented --status default
// to the DefaultStatus constant: the Usage text renders it via a
// backtick splice, so a constant change would otherwise leave the help
// describing a status the tool no longer writes.
func TestUsageDefaultStatusIsTheConst(t *testing.T) {
	t.Parallel()

	if !strings.Contains(Usage, DefaultStatus) {
		t.Errorf("Usage default-status text drifted from DefaultStatus %q:\n%s", DefaultStatus, Usage)
	}
}
