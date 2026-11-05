package battery

import (
	"slices"
	"strings"
	"testing"
)

// TestUsageDocumentsOnlyKnownFlags pins the flag contract exactly, both
// directions: the run verb accepts --json and nothing else, so Usage
// must advertise --json and must not name ANY flag outside the
// documented set (the old "\n  --" substring heuristic false-positived
// on any indented prose; a token scan over the accepted set is the
// explicit form). TestRunUsageErrors holds the parser-side red control.
func TestUsageDocumentsOnlyKnownFlags(t *testing.T) {
	t.Parallel()

	if !strings.Contains(Usage, "usage: docs-gates run [repo-root]") {
		t.Errorf("Usage lost its usage line:\n%s", Usage)
	}

	if !strings.Contains(Usage, flagJSON) {
		t.Errorf("Usage does not document the accepted flag %s:\n%s", flagJSON, Usage)
	}

	documented := []string{flagJSON}

	for line := range strings.SplitSeq(Usage, "\n") {
		for tok := range strings.FieldsSeq(line) {
			if strings.HasPrefix(tok, "--") && !slices.Contains(documented, tok) {
				t.Errorf("Usage documents flag %s outside the accepted set %v:\n%s", tok, documented, Usage)
			}
		}
	}
}
