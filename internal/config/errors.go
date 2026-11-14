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
