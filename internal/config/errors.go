package config

import (
	"encoding/json"
	"errors"
	"fmt"
)

// syntaxError decorates a malformed-JSON refusal with the 1-based line
// and column the decoder's byte Offset names — a raw offset is useless
// to a human editing the file. Unwrap keeps the underlying
// *json.SyntaxError reachable for classification.
type syntaxError struct {
	cause  *json.SyntaxError
	line   int
	column int
}

func (e *syntaxError) Error() string {
	return fmt.Sprintf("%v (line %d, column %d)", e.cause, e.line, e.column)
}

func (e *syntaxError) Unwrap() error { return e.cause }

// withSyntaxLocation passes a decode error through unchanged, unless it
// is a *json.SyntaxError — the one decode failure carrying a byte
// offset a human can act on — decorated with its line/column.
func withSyntaxLocation(data []byte, err error) error {
	syn, ok := errors.AsType[*json.SyntaxError](err)
	if !ok {
		return err
	}

	line, column := byteLineColumn(data, syn.Offset)

	return &syntaxError{cause: syn, line: line, column: column}
}

// byteLineColumn names the position of a decode error over the raw
// file bytes: Go's SyntaxError.Offset counts the bytes READ, so the
// offending byte is the one before it — a newline there means the
// failing literal ended at the previous line's last column. Columns
// count bytes since the last newline (a tab is one column, the JSON
// decoder's own convention).
func byteLineColumn(data []byte, offset int64) (int, int) {
	if offset > int64(len(data)) {
		offset = int64(len(data))
	}

	if offset > 0 {
		offset--
	}

	line, column := 1, 1

	for _, b := range data[:offset] {
		if b == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}

	return line, column
}

// loadStep is the Load stage whose refusal produced a loadError: reading
// the file, parsing its JSON, or validating the decoded values. The zero
// value is no stage — the default String spelling makes an unset step
// loud instead of silently reading as the first stage.
type loadStep int

const (
	loadRead loadStep = iota + 1
	loadParse
	loadValidate
)

func (s loadStep) String() string {
	switch s {
	case loadRead:
		return "reading"
	case loadParse:
		return "parsing"
	case loadValidate:
		return "invalid"
	default:
		return fmt.Sprintf("loadStep(%d)", int(s))
	}
}

// loadError is the package's static error type for Load failures: the
// refusing step, the file it refused on, and the underlying cause. The
// message is the caller's FATAL line verbatim (historical
// "reading/parsing/invalid <path>: <cause>" spelling); Unwrap keeps
// errors.Is/errors.AsType classification of the cause reachable.
type loadError struct {
	step  loadStep
	path  string
	cause error
}

func (e *loadError) Error() string {
	return fmt.Sprintf("%s %s: %v", e.step, e.path, e.cause)
}

func (e *loadError) Unwrap() error { return e.cause }

// newLoadError builds one loadError with the FATAL-line message.
func newLoadError(step loadStep, path string, cause error) *loadError {
	return &loadError{step: step, path: path, cause: cause}
}

// invalidError is the package's static error type: every validate
// refusal is the same concrete type (user-facing CLI text; no errors.Is
// consumers by design, so no Unwrap).
type invalidError struct {
	msg string
}

func (e *invalidError) Error() string { return e.msg }

// invalidf builds one invalidError with the refusal message.
func invalidf(format string, args ...any) *invalidError {
	return &invalidError{msg: fmt.Sprintf(format, args...)}
}
