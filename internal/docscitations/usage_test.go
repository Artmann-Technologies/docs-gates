package docscitations

import (
	"strings"
	"testing"
)

// TestUsageDocumentsNoFlags pins the flagless contract explicitly: the
// citations parser accepts no flags, so Usage must name none.
func TestUsageDocumentsNoFlags(t *testing.T) {
	t.Parallel()

	if !strings.Contains(Usage, "usage: docs-citations [repo-root]") {
		t.Errorf("Usage lost its usage line:\n%s", Usage)
	}

	for line := range strings.SplitSeq(Usage, "\n") {
		for tok := range strings.FieldsSeq(line) {
			if strings.HasPrefix(tok, "--") {
				t.Errorf("Usage documents flag %s, but the citations parser accepts none:\n%s", tok, Usage)
			}
		}
	}
}
