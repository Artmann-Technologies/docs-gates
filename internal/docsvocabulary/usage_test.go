package docsvocabulary

import (
	"strings"
	"testing"
)

// TestUsageDocumentsNoFlags pins the flagless contract explicitly: the
// vocabulary parser accepts no flags, so Usage must name none.
func TestUsageDocumentsNoFlags(t *testing.T) {
	t.Parallel()

	if !strings.Contains(Usage, "usage: docs-vocabulary [repo-root]") {
		t.Errorf("Usage lost its usage line:\n%s", Usage)
	}

	for line := range strings.SplitSeq(Usage, "\n") {
		for tok := range strings.FieldsSeq(line) {
			if strings.HasPrefix(tok, "--") {
				t.Errorf("Usage documents flag %s, but the vocabulary parser accepts none:\n%s", tok, Usage)
			}
		}
	}
}

// TestUsageNamesTheKnobs pins the two config keys the Usage documents —
// the contract a consumer reads before opening .docs-gates.json.
func TestUsageNamesTheKnobs(t *testing.T) {
	t.Parallel()

	for _, knob := range []string{"vocabulary_files", "vocabulary_enforce", "Status legend"} {
		if !strings.Contains(Usage, knob) {
			t.Errorf("Usage must document %s:\n%s", knob, Usage)
		}
	}
}
