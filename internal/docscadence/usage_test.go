package docscadence

import (
	"strings"
	"testing"
)

// TestUsageDocumentsNoFlags pins the flagless contract: the verb's Usage
// must not advertise flags the parser would reject as unknown.
func TestUsageDocumentsNoFlags(t *testing.T) {
	t.Parallel()

	if !strings.Contains(Usage, "usage: docs-cadence [repo-root]") {
		t.Errorf("Usage lost its usage line:\n%s", Usage)
	}

	if strings.Contains(Usage, "\n  --") {
		t.Errorf("Usage documents flags, but the cadence parser accepts none:\n%s", Usage)
	}
}
