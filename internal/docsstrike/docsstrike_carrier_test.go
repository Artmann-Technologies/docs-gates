package docsstrike

import (
	"errors"
	"strings"
	"testing"
)

// TestStrikeInputsErrorRoundTrip pins the carrier contract: errors.AsType
// reads the refusal kind and the matched verdict opener.
func TestStrikeInputsErrorRoundTrip(t *testing.T) {
	t.Parallel()

	err := newStrikeInputsError(refusalPrefixedVerdict, "done")

	sie, ok := errors.AsType[*strikeInputsError](err)
	if !ok {
		t.Fatalf("errors.AsType[*strikeInputsError] failed on: %v", err)
	}

	if sie.kind != refusalPrefixedVerdict {
		t.Errorf("kind = %v, want refusalPrefixedVerdict", sie.kind)
	}

	if sie.opener != "done" {
		t.Errorf("opener = %q, want done", sie.opener)
	}
}

// TestStrikeInputsErrorMessageBytes pins the refusal spellings — each
// missing-input refusal shows a CORRECT example invocation (the
// docs-strike shape from Usage), and the example's verdict payload is
// itself parseable (it must not trip the prefixed-verdict refusal).
func TestStrikeInputsErrorMessageBytes(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		err  *strikeInputsError
		want string
	}{
		"missing row": {
			err: newStrikeInputsError(refusalMissingRow, ""),
			want: "--row <selector> is required (a substring of the row to strike) — e.g. " +
				`docs-strike . --row "<unique task text>" --verdict "landed in <sha>"` +
				" (the evidence carries NO leading verdict word — the tool prepends the done (…) wrap itself)",
		},
		"missing verdict": {
			err: newStrikeInputsError(refusalMissingVerdict, ""),
			want: "--verdict <text> is required (the done (…) evidence payload) — e.g. " +
				`docs-strike . --row "<selector>" --verdict "landed in <sha>"`,
		},
		"prefixed verdict": {
			err: newStrikeInputsError(refusalPrefixedVerdict, "resolved"),
			want: `the --verdict already starts with the verdict word "resolved" — the tool prepends ` +
				`the done (…) wrap itself, so pass the evidence WITHOUT a leading verdict ` +
				`marker (a prefixed verdict double-wraps into done (done (…)))`,
		},
		"loud default kind": {
			err:  newStrikeInputsError(strikeRefusal(99), ""),
			want: "strikeRefusal(99)",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tc.err.Error(); got != tc.want {
				t.Fatalf("Error() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestRefusalExamplesAreSelfConsistent proves the printed examples
// actually parse: the example verdict payloads carry no leading verdict
// word, so a user retyping the example verbatim is not refused again.
func TestRefusalExamplesAreSelfConsistent(t *testing.T) {
	t.Parallel()

	for _, example := range []string{"landed in <sha>"} {
		if opener, prefixed := prefixedVerdictOpener(example); prefixed {
			t.Fatalf(
				"example verdict %q opens with verdict word %q — the printed example must be retypable",
				example,
				opener,
			)
		}
	}

	if _, err := ParseArgs([]string{".", "--row", "<unique task text>", "--verdict", "landed in <sha>"}); err != nil {
		t.Fatalf("the printed example invocation must parse, got: %v", err)
	}
}

// TestRefusalExampleMentionedInHelpVocabulary guards drift between the
// refusal examples and the Usage contract: both must speak of the same
// flags.
func TestRefusalExampleMentionedInHelpVocabulary(t *testing.T) {
	t.Parallel()

	for _, err := range []error{
		newStrikeInputsError(refusalMissingRow, ""),
		newStrikeInputsError(refusalMissingVerdict, ""),
	} {
		for _, flag := range []string{"--row", "--verdict"} {
			if !strings.Contains(err.Error(), flag) {
				t.Fatalf("refusal %q must name %s like Usage does", err.Error(), flag)
			}
		}
	}

	if !strings.Contains(Usage, "--row <selector>") || !strings.Contains(Usage, "--verdict <text>") {
		t.Fatal("Usage must document the --row/--verdict contract the refusals exemplify")
	}
}
