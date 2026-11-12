package docsstrike

import (
	"fmt"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// Options carries the parsed CLI contract of Run.
type Options struct {
	Root                string
	File                string
	Row                 string
	Verdict             string
	Evidence            string
	Status              string
	Apply               bool
	CheckVerdictOutside bool
}

// isValueFlag reports whether the flag consumes the next argument as its
// value; --apply and --check-verdict-outside are the booleans. gatecli.Parse
// is boolean-only, so this parser keeps its value-carrying semantics local
// and hands only the root resolution back to gatecli.
func isValueFlag(arg string) bool {
	switch arg {
	case "--row", "--verdict", "--evidence", "--status", "--file":
		return true
	}

	return false
}

// consumeValueFlag consumes one value-carrying flag at argv[i], wiring the
// value into opts, and returns the index of the last consumed argument.
func consumeValueFlag(argv []string, i int, arg string, opts *Options) (int, *strikeError) {
	if i+1 >= len(argv) {
		return i, &strikeError{msg: arg + " needs a value"}
	}

	value := argv[i+1]

	switch arg {
	case "--row":
		opts.Row = value
	case "--verdict":
		opts.Verdict = value
	case "--evidence":
		opts.Evidence = value
	case "--status":
		opts.Status = value
	case "--file":
		opts.File = value
	}

	return i + 1, nil
}

// ParseArgs splits argv into the strike options: value flags
// (--row/--verdict/--evidence/--status/--file), the --apply and
// --check-verdict-outside booleans, and an optional positional repo root
// (resolved by gatecli.ResolveRoot). In CHECK MODE (--check-verdict-outside)
// neither --row nor --verdict is required — the mode scans the annotate
// corpus and takes no strike inputs at all.
func ParseArgs(argv []string) (Options, error) {
	opts := Options{Status: DefaultStatus}

	var positional []string

	for i := 0; i < len(argv); i++ {
		arg := argv[i]

		switch {
		case arg == "--apply":
			opts.Apply = true
		case arg == "--check-verdict-outside":
			opts.CheckVerdictOutside = true
		case isValueFlag(arg):
			last, err := consumeValueFlag(argv, i, arg, &opts)
			if err != nil {
				return opts, err
			}

			i = last
		case strings.HasPrefix(arg, "--"):
			return opts, fmt.Errorf( //nolint:err113 // user-facing CLI refusal; no errors.Is consumers by design
				"unknown flag: %s",
				arg,
			)
		default:
			positional = append(positional, arg)
		}
	}

	// The strike inputs are required only for a STRIKE: check mode
	// (--check-verdict-outside) scans the annotate corpus and passes none
	// of them.
	if !opts.CheckVerdictOutside {
		if err := validateStrikeInputs(opts); err != nil {
			return opts, err
		}
	}

	root, err := gatecli.ResolveRoot(positional)
	if err != nil {
		return opts, fmt.Errorf("resolve repo root: %w", err)
	}

	opts.Root = root

	return opts, nil
}

// validateStrikeInputs refuses a strike invocation missing its required
// --row/--verdict pair or carrying an already-wrapped verdict — the
// strike-mode contract, split out of ParseArgs to keep its branching
// under the cyclop bar once check mode joined the flag set.
func validateStrikeInputs(opts Options) error {
	if opts.Row == "" {
		return newStrikeInputsError(refusalMissingRow, "")
	}

	if opts.Verdict == "" {
		return newStrikeInputsError(refusalMissingVerdict, "")
	}

	if opener, prefixed := prefixedVerdictOpener(opts.Verdict); prefixed {
		return newStrikeInputsError(refusalPrefixedVerdict, opener)
	}

	return nil
}

// strikeRefusal is the strike-mode input-contract violation a
// strikeInputsError classifies: a missing --row, a missing --verdict, or
// a verdict already opening with a verdict word. Classification only;
// the message bytes live in Error, not here.
type strikeRefusal int

const (
	refusalMissingRow strikeRefusal = iota + 1
	refusalMissingVerdict
	refusalPrefixedVerdict
)

// strikeInputsError is the package's static refusal type: the strike
// invocation misses its required --row/--verdict pair or carries an
// already-wrapped verdict.
type strikeInputsError struct {
	kind   strikeRefusal
	opener string
}

func (e *strikeInputsError) Error() string {
	switch e.kind {
	case refusalMissingRow:
		return "--row <selector> is required (a substring of the row to strike)"
	case refusalMissingVerdict:
		return "--verdict <text> is required (the done (…) evidence payload)"
	case refusalPrefixedVerdict:
		return fmt.Sprintf(
			"the --verdict already starts with the verdict word %q — the tool prepends "+
				"the done (…) wrap itself, so pass the evidence WITHOUT a leading verdict "+
				"marker (a prefixed verdict double-wraps into done (done (…)))",
			e.opener,
		)
	default:
		return fmt.Sprintf("strikeRefusal(%d)", int(e.kind))
	}
}

// newStrikeInputsError builds one strikeInputsError with the refusal
// message; opener carries the matched verdict word for
// refusalPrefixedVerdict and stays empty otherwise.
func newStrikeInputsError(kind strikeRefusal, opener string) *strikeInputsError {
	return &strikeInputsError{kind: kind, opener: opener}
}

// verdictOpeners are the leading words of the docs-health verdict
// vocabulary the tool already wraps the evidence in. A --verdict that
// OPENS with one of them would land as done (done (…)) — the
// double-wrapped wart class (burned 2026-10-12, caught in dry-run).
//
//nolint:gochecknoglobals // the fixed docs-health verdict vocabulary, same class as the contract tables
var verdictOpeners = []string{
	"done",
	"resolved",
	"superseded",
	"decided",
	"won't implement",
	"wont implement",
	"not-do",
	"not do",
}

// prefixedVerdictOpener reports whether the verdict already opens with a
// verdict word (case-insensitive, markdown emphasis tolerated) and
// returns the matched opener for the refusal message.
func prefixedVerdictOpener(verdict string) (string, bool) {
	v := strings.ToLower(strings.TrimSpace(verdict))
	v = strings.TrimLeft(v, "*~`_ ")

	for _, opener := range verdictOpeners {
		if strings.HasPrefix(v, opener) {
			return opener, true
		}
	}

	return "", false
}
