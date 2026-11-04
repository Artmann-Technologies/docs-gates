package gatecli

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

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

func TestTruncateKeepsShortAndCutsLong(t *testing.T) {
	t.Parallel()

	if got := Truncate("short", 10); got != "short" {
		t.Fatalf("Truncate(short, 10) = %q", got)
	}

	if got, want := Truncate("abcdefgh", 3), "abc"; got != want {
		t.Fatalf("Truncate(abcdefgh, 3) = %q, want %q", got, want)
	}

	if got, want := Truncate("abc", 3), "abc"; got != want {
		t.Fatalf("exact-fit boundary: got %q, want %q", got, want)
	}
}

func TestTruncateNeverSplitsUTF8Runes(t *testing.T) {
	t.Parallel()

	// "aé€b": é is 2 bytes (offsets 1-2), € is 3 bytes (offsets 3-5).
	tests := []struct {
		n    int
		want string
	}{
		{0, ""},
		{1, "a"},
		{2, "a"},   // would split é; backs off to "a"
		{3, "aé"},  // é fits whole
		{4, "aé"},  // would split €; backs off to "aé"
		{5, "aé"},  // still mid-€
		{6, "aé€"}, // € fits whole
		{7, "aé€b"},
	}

	for _, tt := range tests {
		if got := Truncate("aé€b", tt.n); got != tt.want {
			t.Fatalf("Truncate(multi-byte, %d) = %q, want %q", tt.n, got, tt.want)
		}
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

// HandleVersion: the shared --version prologue every cmd main runs before
// its own parsing (the gate parsers would reject --version as unknown).
func TestHandleVersion(t *testing.T) {
	t.Parallel()

	if !HandleVersion([]string{".", "--version"}, "docs-freshness", "test") {
		t.Fatal("--version must be intercepted")
	}

	if HandleVersion([]string{".", "--audit-all"}, "docs-freshness", "test") {
		t.Fatal("plain args must not be intercepted")
	}
}
