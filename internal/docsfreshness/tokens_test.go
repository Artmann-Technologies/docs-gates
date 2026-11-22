package docsfreshness

// These tests cover the token matching (ex tests/test_docs_freshness.py,
// item 8.4 of the docs-health list, 2026-09-25): a report's `real-cax` token
// must match a living-docs line saying "real CAX" — hyphen/space-insensitive.
// Before the fix, two evidence cells had to be hand-phrased to carry the
// hyphenated spelling, a manual workaround the gate was supposed to remove.

import (
	"strings"
	"testing"
)

func TestSquashCollapsesHyphensAndSpaces(t *testing.T) {
	t.Parallel()

	if got := Squash("real-cax"); got != "realcax" {
		t.Fatalf("Squash(real-cax) = %q", got)
	}

	if got := Squash("real  CAX"); got != "realcax" {
		t.Fatalf("Squash(real  CAX) = %q", got)
	}

	if got := Squash("a - b -- c"); got != "abc" {
		t.Fatalf("Squash(a - b -- c) = %q", got)
	}
}

// --explain-token's table builder: the boundary-token class needs the DF
// number and the distinctive/generic verdict per token, visible at a
// glance (the 2026-10-07 flip: "batch" DF 7→8 silently genericized an
// item's last distinctive token; the diagnosis needed a hand-rolled grep
// probe before this existed).
func TestTokenTableShowsDFAndDistinctiveness(t *testing.T) {
	t.Parallel()

	living := make([]string, 0, 8)
	for range 7 {
		living = append(living, "some line mentioning batch")
	}

	living = append(living, "a line with the zephyr-unique token")

	table := TokenTable("Re-run the ritual for the batch", NewCorpus(living), DefaultSettings().df())
	joined := strings.Join(table, "\n")

	for _, want := range []string{
		"batch", "DF=7", "DISTINCTIVE",
		"ritual", "missing",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("token table missing %q:\n%s", want, joined)
		}
	}

	if strings.Contains(joined, "zephyr") {
		t.Errorf("table must only carry the item's own tokens:\n%s", joined)
	}
}

// TestTokensSplitSlashJoinedKeywords (#5): a slash-joined keyword run
// must yield one token per segment — the timesheets false-DRIFT, where
// "grouping/valid-args/positional" stayed one unmatchable token while
// the living docs carried every word. Path-like tokens now match as
// segments too.
func TestTokensSplitSlashJoinedKeywords(t *testing.T) {
	t.Parallel()

	got := Tokens("F8: cmdguard P4 help grouping/valid-args/positional IDs (#64).")
	joined := strings.Join(got, " ")

	for _, want := range []string{"grouping", "valid-args", "positional", "cmdguard"} {
		if !strings.Contains(joined, want) {
			t.Errorf("token %q missing from %v", want, got)
		}
	}

	for _, whole := range []string{"grouping/valid-args/positional"} {
		if strings.Contains(joined, whole) {
			t.Errorf("slash-joined run must not survive as one token: %v", got)
		}
	}

	paths := Tokens("see docs/handoff/report-kit.md for the shape")
	for _, want := range []string{"handoff", "report-kit.md"} {
		if !strings.Contains(strings.Join(paths, " "), want) {
			t.Errorf("path segment %q missing from %v", want, paths)
		}
	}
}
