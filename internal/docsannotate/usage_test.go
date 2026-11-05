package docsannotate

import (
	"slices"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// TestUsageDocumentsEveryAcceptedFlag pins the Usage text against the
// parser's flag vocabulary, both directions: a flag the parser accepts
// must appear in the help, and a flag the help names must parse.
func TestUsageDocumentsEveryAcceptedFlag(t *testing.T) {
	t.Parallel()

	documented := slices.Concat(allowedFlags, []string{"--annotate-max"})

	for _, flag := range documented {
		if !strings.Contains(Usage, flag) {
			t.Errorf("Usage does not document accepted flag %s", flag)
		}
	}

	for _, flag := range allowedFlags {
		_, _, unknown := gatecli.ParseFlags([]string{flag}, allowedFlags...)
		if slices.Contains(unknown, flag) {
			t.Errorf("parser rejects documented flag %s", flag)
		}
	}

	if _, rest, err := gatecli.CutIntFlag([]string{"--annotate-max=3"}, "annotate-max"); err != nil || len(rest) != 0 {
		t.Errorf("valued flag --annotate-max does not parse: err=%v rest=%v", err, rest)
	}
}
