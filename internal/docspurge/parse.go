package docspurge

import (
	"fmt"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// defaultFileMode is the permission bits for freshly created files;
// existing files keep their own mode (writePreservingMode).
const defaultFileMode = 0o644

// purgeError is the parse-layer refusal; FatalUsage renders it.
type purgeError struct {
	msg string
}

func (e *purgeError) Error() string { return e.msg }

// Options carries the parsed CLI contract of Run.
type Options struct {
	Root      string
	File      string
	Changelog string
	Apply     bool
}

// isValueFlag reports whether the flag consumes the next argument as its
// value; --apply is the lone boolean. gatecli.Parse is boolean-only, so
// this parser keeps its value-carrying semantics local and hands only
// the root resolution back to gatecli.
func isValueFlag(arg string) bool {
	switch arg {
	case "--file", "--changelog":
		return true
	}

	return false
}

// consumeValueFlag consumes one value-carrying flag at argv[i], wiring
// the value into opts, and returns the index of the last consumed
// argument.
func consumeValueFlag(argv []string, i int, arg string, opts *Options) (int, *purgeError) {
	if i+1 >= len(argv) {
		return i, &purgeError{msg: arg + " needs a value"}
	}

	value := argv[i+1]

	switch arg {
	case "--file":
		opts.File = value
	case "--changelog":
		opts.Changelog = value
	}

	return i + 1, nil
}

// ParseArgs splits argv into the purge options: the value flags
// (--file/--changelog), the --apply boolean, and an optional positional
// repo root (resolved by gatecli.ResolveRoot).
func ParseArgs(argv []string) (Options, error) {
	var opts Options

	var positional []string

	for i := 0; i < len(argv); i++ {
		arg := argv[i]

		switch {
		case arg == "--apply":
			opts.Apply = true
		case isValueFlag(arg):
			last, err := consumeValueFlag(argv, i, arg, &opts)
			if err != nil {
				return opts, err
			}

			i = last
		case strings.HasPrefix(arg, "--"):
			return opts, fmt.Errorf(
				"unknown flag: %s",
				arg,
			) //nolint:err113 // user-facing CLI refusal; no errors.Is consumers by design
		default:
			positional = append(positional, arg)
		}
	}

	root, err := gatecli.ResolveRoot(positional)
	if err != nil {
		return opts, fmt.Errorf("resolve repo root: %w", err)
	}

	opts.Root = root

	return opts, nil
}
