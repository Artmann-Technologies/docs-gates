package docsagents

import (
	"strings"
	"testing"
)

// TestUsageDocumentsNoFlags pins the flagless contract explicitly: the
// agents parser accepts no flags, so Usage must name none.
func TestUsageDocumentsNoFlags(t *testing.T) {
	t.Parallel()

	if !strings.Contains(Usage, "usage: docs-agents [repo-root]") {
		t.Errorf("Usage lost its usage line:\n%s", Usage)
	}

	for line := range strings.SplitSeq(Usage, "\n") {
		for tok := range strings.FieldsSeq(line) {
			if strings.HasPrefix(tok, "--") {
				t.Errorf("Usage documents flag %s, but the agents parser accepts none:\n%s", tok, Usage)
			}
		}
	}
}

// TestUsageNamesTheKnobs pins the config keys the Usage documents —
// the contract a consumer reads before opening .docs-gates.json.
func TestUsageNamesTheKnobs(t *testing.T) {
	t.Parallel()

	for _, knob := range []string{"agents_file", "agents_warn_days", "agents_max_days", "agents_enforce"} {
		if !strings.Contains(Usage, knob) {
			t.Errorf("Usage must document %s:\n%s", knob, Usage)
		}
	}
}
