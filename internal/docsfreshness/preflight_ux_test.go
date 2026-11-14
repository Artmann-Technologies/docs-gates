package docsfreshness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// TestPreFlightRefusalEchoesTheDiskLine pins the WHERE beside the WHAT:
// an on-disk collision shows the candidate AND the disk line it matched.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestPreFlightRefusalEchoesTheDiskLine(t *testing.T) {
	root := preFlightRoot(t)

	captured := gatecli.CaptureOutput(func() int {
		return PreFlight(root, []string{"webphone note number 3"}, nil, false, false, DefaultSettings())
	})

	if captured.Code != gatecli.ExitUsage {
		t.Fatalf("an on-disk candidate must exit usage, got %d", captured.Code)
	}

	if !strings.Contains(captured.Stderr, "candidate ← disk") {
		t.Fatalf("the refusal must label the collision pairs, stderr:\n%s", captured.Stderr)
	}

	if !strings.Contains(captured.Stderr, "webphone note number 3") {
		t.Fatalf("the refusal must echo the candidate, stderr:\n%s", captured.Stderr)
	}

	if !strings.Contains(captured.Stderr, "- webphone note number 3") {
		t.Fatalf("the refusal must echo the full disk line beside the candidate, stderr:\n%s", captured.Stderr)
	}
}

// TestPreFlightNamesOutsideCorpusReplaceTarget pins the honest
// out-of-universe refusal: a --replace target living in a non-corpus
// file (AGENTS.md) is named with its FILE, not left as a bare
// not-found.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestPreFlightNamesOutsideCorpusReplaceTarget(t *testing.T) {
	root := preFlightRoot(t)

	const agentsLine = "the lattice reviewer owns the routing table"

	if err := os.WriteFile(
		filepath.Join(root, "AGENTS.md"),
		[]byte("# AGENTS\n\n"+agentsLine+"\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	captured := gatecli.CaptureOutput(func() int {
		return PreFlight(root, []string{"- the lattice reviewer calendar block"},
			[]string{agentsLine}, false, false, DefaultSettings())
	})

	if captured.Code != gatecli.ExitUsage {
		t.Fatalf("an out-of-corpus replace target must exit usage, got %d", captured.Code)
	}

	if !strings.Contains(captured.Stderr, "OUTSIDE the living-docs corpus") {
		t.Fatalf("the refusal must name the outside class, stderr:\n%s", captured.Stderr)
	}

	if !strings.Contains(captured.Stderr, "AGENTS.md") {
		t.Fatalf("the refusal must name the file the target lives in, stderr:\n%s", captured.Stderr)
	}

	if strings.Contains(captured.Stderr, "not found in the living docs") {
		t.Fatalf("a named outside target must not ALSO read as not-found, stderr:\n%s", captured.Stderr)
	}
}

// TestPreFlightMixedReplaceTargets: an out-of-corpus target AND a
// genuinely-missing one each land in their own refusal class.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestPreFlightMixedReplaceTargets(t *testing.T) {
	root := preFlightRoot(t)

	if err := os.WriteFile(
		filepath.Join(root, "NOTES.md"),
		[]byte("the qenor dispatch ledger is quarterly\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	captured := gatecli.CaptureOutput(func() int {
		return PreFlight(root, []string{"- new line"}, []string{
			"the qenor dispatch ledger is quarterly",
			"nowhere at all",
		}, false, false, DefaultSettings())
	})

	if captured.Code != gatecli.ExitUsage {
		t.Fatalf("mixed missing targets must exit usage, got %d", captured.Code)
	}

	if !strings.Contains(captured.Stderr, "NOTES.md") {
		t.Fatalf("the outside target must name its file, stderr:\n%s", captured.Stderr)
	}

	if !strings.Contains(captured.Stderr, "nowhere at all") {
		t.Fatalf("the nowhere target must stay in the not-found list, stderr:\n%s", captured.Stderr)
	}
}
