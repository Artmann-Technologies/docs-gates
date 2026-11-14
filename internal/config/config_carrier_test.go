package config

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadErrorCarrierRoundTrip pins the loadError carrier contract:
// errors.AsType reaches it from the returned error and reads the
// refusing step, the path, and the unwrapped cause. A directory sitting
// at the config path forces the read step to fail.
func TestLoadErrorCarrierRoundTrip(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	if err := os.MkdirAll(filepath.Join(root, FileName), 0o755); err != nil {
		t.Fatal(err)
	}

	_, _, err := Load(root)
	if err == nil {
		t.Fatal("Load succeeded reading a directory, want a read-step loadError")
	}

	le, ok := errors.AsType[*loadError](err)
	if !ok {
		t.Fatalf("errors.AsType[*loadError] failed on: %v", err)
	}

	if le.step != loadRead {
		t.Errorf("step = %v, want loadRead", le.step)
	}

	if le.path != Path(root) {
		t.Errorf("path = %q, want %q", le.path, Path(root))
	}

	if !errors.Is(err, io.EOF) && !strings.Contains(err.Error(), "directory") &&
		!strings.Contains(err.Error(), "is a") {
		t.Errorf("the read cause must stay reachable through Unwrap, got: %v", err)
	}
}

// TestLoadErrorMessageBytes pins the Error() spelling — it is the
// caller's FATAL line verbatim — plus the loud default for an unset
// step.
func TestLoadErrorMessageBytes(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		err  *loadError
		want string
	}{
		"read step": {
			err:  newLoadError(loadRead, "/r/.docs-gates.json", io.ErrUnexpectedEOF),
			want: "reading /r/.docs-gates.json: unexpected EOF",
		},
		"validate step": {
			err:  newLoadError(loadValidate, "/r/.docs-gates.json", invalidf("generic_df must be >= 1, got 0")),
			want: "invalid /r/.docs-gates.json: generic_df must be >= 1, got 0",
		},
		"loud default step": {
			err:  newLoadError(loadStep(99), "/r/.docs-gates.json", io.ErrUnexpectedEOF),
			want: "loadStep(99) /r/.docs-gates.json: unexpected EOF",
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

// TestLoadParseRefusalNamesLineColumn pins the UX refusal: a malformed
// config surfaces the decoder's byte offset as a human line:column, and
// the *json.SyntaxError stays reachable through both carriers for
// classification.
func TestLoadParseRefusalNamesLineColumn(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeConfig(t, root, "{\n  \"report_dir\": \"docs/status\",\n  \"generic_df\": tru\n}\n")

	_, _, err := Load(root)
	if err == nil {
		t.Fatal("Load succeeded on malformed JSON, want an error")
	}

	const wantSuffix = " (line 3, column 20)"

	if !strings.HasSuffix(err.Error(), wantSuffix) {
		t.Fatalf("Error() must end with %q, got: %s", wantSuffix, err.Error())
	}

	le, ok := errors.AsType[*loadError](err)
	if !ok || le.step != loadParse {
		t.Fatalf("the parse refusal must carry a loadParse loadError, got: %v", err)
	}

	syn, ok := errors.AsType[*json.SyntaxError](err)
	if !ok || syn.Offset == 0 {
		t.Fatalf("*json.SyntaxError must stay reachable through the carriers, got: %v", err)
	}
}

// TestByteLineColumnTable pins the offset→line:column arithmetic: Go's
// SyntaxError.Offset counts the bytes READ, so the named byte is the one
// before it; columns count bytes since the last newline.
func TestByteLineColumnTable(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		data      string
		offset    int64
		line, col int
	}{
		"newline terminator names the literal's end": {data: "{\n  \"a\": tru\n}\n", offset: 13, line: 2, col: 11},
		"error char after the last newline":          {data: "{\n  \"a\": 1,\n}", offset: 13, line: 3, col: 1},
		"first byte":                                 {data: "}", offset: 1, line: 1, col: 1},
		"offset past the end clamps to EOF":          {data: "x", offset: 99, line: 1, col: 1},
		"empty document":                             {data: "", offset: 0, line: 1, col: 1},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			line, col := byteLineColumn([]byte(tc.data), tc.offset)
			if line != tc.line || col != tc.col {
				t.Fatalf("byteLineColumn(%d) = %d:%d, want %d:%d", tc.offset, line, col, tc.line, tc.col)
			}
		})
	}
}
