package gatecli

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// The argv-parsing pair: ParseFlags/Parse/ResolveRoot ordering and
// unknown-flag semantics, plus the CutIntFlag valued-flag splitter.
// Split out of gatecli_test.go at 339 lines (the 350 fleet ceiling)
// before the next parser test lands.

func TestParseFlagsSplitsPositionalsAndSortsUnknown(t *testing.T) {
	t.Parallel()

	flags, positional, unknown := ParseFlags(
		[]string{"--verbose", "repo", "--zulu", "--audit-all", "--alpha"},
		"--verbose", "--audit-all",
	)

	if !flags["--verbose"] || !flags["--audit-all"] || !flags["--zulu"] || !flags["--alpha"] {
		t.Fatalf("flags not all recorded: %v", flags)
	}

	if !slices.Equal(positional, []string{"repo"}) {
		t.Fatalf("positional = %v, want [repo]", positional)
	}

	want := []string{"--alpha", "--zulu"}
	if !slices.Equal(unknown, want) {
		t.Fatalf("unknown = %v, want %v (sorted, deterministic)", unknown, want)
	}
}

func TestResolveRootAcceptsDirectoryRejectsFileAndMissing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	got, err := ResolveRoot([]string{dir})
	if err != nil {
		t.Fatalf("directory rejected: %v", err)
	}

	if got != dir {
		t.Fatalf("root = %q, want %q", got, dir)
	}

	file := filepath.Join(dir, "plain-file")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, bad := range []string{file, filepath.Join(dir, "missing")} {
		if _, err := ResolveRoot([]string{bad}); err == nil {
			t.Fatalf("%s accepted, want %q error", bad, "is not a directory")
		}
	}
}

func TestParsePrologueRejectsUnknownFlagsAndBadRoot(t *testing.T) {
	t.Parallel()

	if _, _, err := Parse([]string{"--zulu", "--alpha"}, "--alpha"); err == nil ||
		err.Error() != "unknown flag(s): --zulu" {
		t.Fatalf("unknown-flag error = %v, want sorted single-flag message", err)
	}

	missing := filepath.Join(t.TempDir(), "missing")
	if _, _, err := Parse([]string{missing}); err == nil || err.Error() != missing+" is not a directory" {
		t.Fatalf("bad-root error = %v, want %q", err, missing+" is not a directory")
	}

	dir := t.TempDir()

	flags, root, err := Parse([]string{dir, "--verbose"}, "--verbose")
	if err != nil {
		t.Fatalf("valid prologue rejected: %v", err)
	}

	if !flags["--verbose"] || root != dir {
		t.Fatalf("Parse returned flags=%v root=%q", flags, root)
	}
}

func TestOrderingSemantics(t *testing.T) {
	t.Parallel()

	flags, positional, unknown := ParseFlags(
		[]string{"first", "--verbose", "second", "--audit-all", "third"},
		"--verbose", "--audit-all",
	)

	if !flags["--verbose"] || !flags["--audit-all"] || len(unknown) != 0 {
		t.Fatalf("interleaved flags must all register regardless of position: %v (unknown %v)", flags, unknown)
	}

	if !slices.Equal(positional, []string{"first", "second", "third"}) {
		t.Fatalf("positional = %v, want all three in order", positional)
	}

	dir := t.TempDir()
	other := t.TempDir()

	root, err := ResolveRoot([]string{dir, other})
	if err != nil {
		t.Fatal(err)
	}

	if root != dir {
		t.Fatalf("ResolveRoot = %q, want the FIRST positional %q (later ones ignored)", root, dir)
	}
}

// CutIntFlag: every occurrence splits off and the LAST value wins.
func TestCutIntFlagLastValueWins(t *testing.T) {
	t.Parallel()

	value, rest, err := CutIntFlag([]string{".", "--audit-max=3", "--strict", "--audit-max=7"}, "audit-max")
	if err != nil {
		t.Fatalf("CutIntFlag: %v", err)
	}

	if value != 7 {
		t.Fatalf("value = %d, want 7 (last occurrence wins)", value)
	}

	if !slices.Equal(rest, []string{".", "--strict"}) {
		t.Fatalf("rest = %v, want every occurrence removed", rest)
	}
}

// CutIntFlag error path: the message bytes are pinned (they ride the
// caller's FATAL line, exit 2), value is -1 and rest nil.
func TestCutIntFlagErrorBytesAndNilRest(t *testing.T) {
	t.Parallel()

	cases := []struct {
		args []string
		want string
	}{
		{[]string{"--audit-max=x"}, `--audit-max wants a non-negative integer, got "x"`},
		{[]string{"--audit-max=-1"}, `--audit-max wants a non-negative integer, got "-1"`},
		{[]string{"--audit-max="}, `--audit-max wants a non-negative integer, got ""`},
	}

	for _, tt := range cases {
		value, rest, err := CutIntFlag(tt.args, "audit-max")
		if err == nil || err.Error() != tt.want {
			t.Fatalf("CutIntFlag(%v) error = %v, want %q", tt.args, err, tt.want)
		}

		if value != -1 {
			t.Fatalf("error-path value = %d, want -1", value)
		}

		if rest != nil {
			t.Fatalf("error-path rest = %v, want nil (the caller's failure sentinel)", rest)
		}
	}
}

// CutIntFlag success sentinels: rest is non-nil even when every arg was a
// valued flag, and an absent flag returns -1 with the args untouched.
func TestCutIntFlagRestSentinels(t *testing.T) {
	t.Parallel()

	value, rest, err := CutIntFlag([]string{"--annotate-max=2"}, "annotate-max")
	if err != nil || value != 2 {
		t.Fatalf("CutIntFlag: value=%d err=%v, want 2/nil", value, err)
	}

	if rest == nil || len(rest) != 0 {
		t.Fatalf("rest = %v, want non-nil empty on success", rest)
	}

	value, rest, err = CutIntFlag([]string{"--strict"}, "annotate-max")
	if err != nil || value != -1 || !slices.Equal(rest, []string{"--strict"}) {
		t.Fatalf("absent flag: value=%d rest=%v err=%v, want -1, the args, nil", value, rest, err)
	}
}
