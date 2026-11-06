package gatecli

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
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

// LoadConfig: a repo without a config is silent (no stderr noise), a
// mis-keyed config announces its warnings BEFORE the error surfaces, and
// the error message is exactly what the caller's FATAL line prints.
// Serial: the stderr capture is a process-global swap (CaptureOutput doc).
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestLoadConfigWarningsOrderingAndSilence(t *testing.T) {
	root := t.TempDir()

	out := CaptureOutput(func() int {
		cfg, err := LoadConfig(root)
		if err != nil {
			t.Fatalf("clean LoadConfig: %v", err)
		}

		if cfg.AuditMax != -1 {
			t.Fatalf("clean cfg AuditMax = %d, want the strict default", cfg.AuditMax)
		}

		return 0
	})

	if out.Stderr != "" || out.Stdout != "" {
		t.Fatalf("clean LoadConfig emitted %q / %q, want silence", out.Stderr, out.Stdout)
	}

	bad := t.TempDir()
	body := []byte(`{"generic_df_bar": 1, "audit_max": -2}`)

	if err := os.WriteFile(filepath.Join(bad, ".docs-gates.json"), body, 0o600); err != nil {
		t.Fatal(err)
	}

	var cfgErr error

	out = CaptureOutput(func() int {
		_, cfgErr = LoadConfig(bad)

		return 0
	})

	wantWarning := `config: unknown config key "generic_df_bar" ignored` + "\n"
	if out.Stderr != wantWarning {
		t.Fatalf("stderr = %q, want warning printed before the error is returned: %q", out.Stderr, wantWarning)
	}

	wantErr := filepath.Join(bad, ".docs-gates.json") + ": audit_max must be >= -1 (-1 = strict), got -2"
	if cfgErr == nil || !strings.HasPrefix(cfgErr.Error(), "invalid ") || !strings.HasSuffix(cfgErr.Error(), wantErr) {
		t.Fatalf("error = %v, want `invalid ` prefix and suffix %q", cfgErr, wantErr)
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

// WantsHelp: the --help/-h detection the dispatcher runs in the same
// slot as HandleVersion (the gate parsers reject both as unknown).
func TestWantsHelp(t *testing.T) {
	t.Parallel()

	for _, argv := range [][]string{
		{"--help"},
		{"-h"},
		{".", "--help"},
		{".", "report", "-h"},
	} {
		if !WantsHelp(argv) {
			t.Errorf("WantsHelp(%v) = false, want true", argv)
		}
	}

	for _, argv := range [][]string{
		nil,
		{},
		{"."},
		{".", "--audit-all"},
		{"--version"},
	} {
		if WantsHelp(argv) {
			t.Errorf("WantsHelp(%v) = true, want false", argv)
		}
	}
}

// FatalUsage: the FATAL line plus the verb's help pointer, one shape for
// every gate's parse refusal.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestFatalUsagePrintsFATALAndHelpPointer(t *testing.T) {
	out := CaptureOutput(func() int {
		code := FatalUsage("freshness", NewUnknownFlagsError([]string{"--bogus"}))
		if code != ExitUsage {
			t.Fatalf("code = %d, want ExitUsage", code)
		}

		return code
	})

	want := "FATAL: unknown flag(s): --bogus\nusage: docs-gates freshness --help\n"
	if out.Stderr != want {
		t.Fatalf("stderr = %q, want %q", out.Stderr, want)
	}

	if out.Stdout != "" {
		t.Fatalf("stdout = %q, want empty", out.Stdout)
	}
}

// TestCaptureOutputSurvivesLargeOutput is the 2026-11-05 fleet-wedge
// regression: a gate writing past the ~64KB pipe buffer into its own
// capture used to deadlock (the write parked on the full pipe while
// io.ReadAll waited for fn to return first — netpoll-parked at 0% CPU
// forever; crm's 179-finding attribution dump was the live trigger).
// With concurrent drainers the capture must complete and lose nothing.
//
//nolint:paralleltest // process-global capture swap, serialized on purpose
func TestCaptureOutputSurvivesLargeOutput(t *testing.T) {
	const lines = 40000 // ~4MB per stream, far past the pipe buffer

	start := time.Now()

	out := CaptureOutput(func() int {
		for range lines {
			fmt.Fprintln(os.Stdout, strings.Repeat("x", 99))
			fmt.Fprintln(os.Stderr, strings.Repeat("y", 99))
		}

		return ExitClean
	})

	if elapsed := time.Since(start); elapsed > 30*time.Second {
		t.Fatalf("capture deadlocked on large output: %v", elapsed)
	}

	if out.Code != ExitClean {
		t.Fatalf("code = %d, want clean", out.Code)
	}

	if want := lines * 100; len(out.Stdout) != want || len(out.Stderr) != want {
		t.Fatalf("capture lost bytes: stdout %d/%d, stderr %d/%d", len(out.Stdout), want, len(out.Stderr), want)
	}
}

func TestFormatDurationHonestPrecisionTiers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   time.Duration
		want string
	}{
		{"sub-microsecond rounds to a floor, not a lie", 400, "<1µs"},
		{"exact microsecond tier renders µs", 800 * time.Microsecond, "800µs"},
		{"sub-millisecond keeps one decimal", 1500 * time.Microsecond, "1.5ms"},
		{"sub-10ms rounds to 100µs steps", 3140 * time.Microsecond, "3.1ms"},
		{"the old <10ms bucket is gone", 9 * time.Millisecond, "9ms"},
		{"sub-second keeps 10ms steps", 123 * time.Millisecond, "120ms"},
		{"sub-minute keeps 100ms steps", 4*time.Second + 567*time.Millisecond, "4.6s"},
		{"minutes round whole seconds", 65 * time.Second, "1m5s"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := FormatDuration(tc.in); got != tc.want {
				t.Fatalf("FormatDuration(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
