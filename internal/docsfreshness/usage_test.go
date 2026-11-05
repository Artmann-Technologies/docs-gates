package docsfreshness

import (
	"slices"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// TestUsageDocumentsEveryAcceptedFlag pins the Usage text against the
// parser's flag vocabulary, both directions: a flag the parser accepts
// must appear in the help, and a flag the help names must parse. A help
// that lies is worse than none.
func TestUsageDocumentsEveryAcceptedFlag(t *testing.T) {
	t.Parallel()

	for _, flag := range slices.Concat(allowedFlags, valuedFlags) {
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

	if _, rest, err := gatecli.CutIntFlag([]string{"--audit-max=3"}, "audit-max"); err != nil || len(rest) != 0 {
		t.Errorf("valued flag --audit-max does not parse: err=%v rest=%v", err, rest)
	}

	if replaces, _, rc := parseReplaceFlags([]string{"--replace=old line"}); rc != 0 || len(replaces) != 1 {
		t.Errorf("valued flag --replace does not parse: replaces=%v rc=%d", replaces, rc)
	}
}
