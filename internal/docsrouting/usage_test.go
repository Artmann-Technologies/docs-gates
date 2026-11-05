package docsrouting

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

	for _, flag := range allowedFlags {
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
}
