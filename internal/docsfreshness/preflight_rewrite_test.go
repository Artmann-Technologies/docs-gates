package docsfreshness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// rewriteRoot builds a corpus whose historical §f item is anchored by
// exactly ONE living-doc line — the rewrite target. Removing that line
// must flip the item (the additions-only model cannot see this class:
// with the old line still present the item keeps harvesting).
func rewriteRoot(t *testing.T) string {
	t.Helper()

	root := t.TempDir()

	status := filepath.Join(root, "docs", "status")
	if err := os.MkdirAll(status, 0o755); err != nil {
		t.Fatal(err)
	}

	report := "# Report\n\n## f) Next\n\n1. the qxvorlag spindle reconciliation needs a pass\n"
	if err := os.WriteFile(filepath.Join(status, "2026-09-01_00-00_BASE.md"), []byte(report), 0o644); err != nil {
		t.Fatal(err)
	}

	// The anchor rides TODO_LIST.md (a default living doc — the tests
	// pass DefaultSettings, which does not read the repo config).
	todo := "# TODO_LIST\n\n- the qxvorlag spindle reconciliation needs a pass\n"
	if err := os.WriteFile(filepath.Join(root, "TODO_LIST.md"), []byte(todo), 0o644); err != nil {
		t.Fatal(err)
	}

	return root
}

// TestPreFlightRewriteModeJudgesNewWordsOnly is the rewrite-mode
// discrimination: the SAME candidate line is SAFE under the
// additions-only model (the old line still anchors the item) but UNSAFE
// once --replace removes that line — an item anchored only by the
// replaced line loses its harvest anchor, which the additions-only
// model cannot see.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestPreFlightRewriteModeJudgesNewWordsOnly(t *testing.T) {
	root := rewriteRoot(t)

	candidate := []string{"- the qzorp fleet dashboard wears a cloak"}
	oldLine := "- the qxvorlag spindle reconciliation needs a pass"

	if rc := PreFlight(root, candidate, nil, false, false, DefaultSettings()); rc != gatecli.ExitClean {
		t.Fatalf("additions-only model must judge the candidate alone SAFE (old line still present), got %d", rc)
	}

	captured := gatecli.CaptureOutput(func() int {
		return PreFlight(root, candidate, []string{oldLine}, false, false, DefaultSettings())
	})

	if captured.Code != gatecli.ExitFindings {
		t.Fatalf("rewrite-mode must flip the item anchored only by the replaced line, got %d (output:\n%s)",
			captured.Code, captured.Stdout)
	}
}

// TestPreFlightRewriteToSameLineIsExempt: a candidate equal to the line
// it replaces is not an on-disk duplicate (the model removes that line
// first), so the refusal must not fire and the no-op rewrite is SAFE.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestPreFlightRewriteToSameLineIsExempt(t *testing.T) {
	root := rewriteRoot(t)

	sameLine := []string{"- the qxvorlag spindle reconciliation needs a pass"}

	captured := gatecli.CaptureOutput(func() int {
		return PreFlight(root, sameLine, sameLine, false, false, DefaultSettings())
	})

	if captured.Code != gatecli.ExitClean {
		t.Fatalf("a no-op rewrite (candidate equals its replace target) must stay clean, got %d (stderr:\n%s)",
			captured.Code, captured.Stderr)
	}
}

// TestPreFlightReplaceTargetNotFound: a --replace naming a line that is
// not on disk is a usage refusal naming the target — silently ignoring
// it would model a removal that will not happen.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestPreFlightReplaceTargetNotFound(t *testing.T) {
	root := rewriteRoot(t)

	captured := gatecli.CaptureOutput(func() int {
		return PreFlight(root, []string{"- the qzorp fleet dashboard wears a cloak"},
			[]string{"- no such line anywhere"}, false, false, DefaultSettings())
	})

	if captured.Code != gatecli.ExitUsage {
		t.Fatalf("a missing replace target must exit usage, got %d", captured.Code)
	}

	if !strings.Contains(captured.Stderr, "no such line anywhere") {
		t.Fatalf("the refusal must name the missing target, stderr:\n%s", captured.Stderr)
	}
}

// TestPreFlightHeadingAdvisory: heading candidates get the split-landing
// advisory (a heading pre-flight'd separately from its batch models only
// each part's DF push) while the verdict stays the normal replay.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestPreFlightHeadingAdvisory(t *testing.T) {
	root := rewriteRoot(t)

	candidate := []string{"## qzorp batch", "- the qzorp fleet dashboard wears a cloak"}

	captured := gatecli.CaptureOutput(func() int {
		return PreFlight(root, candidate, nil, false, false, DefaultSettings())
	})

	if captured.Code != gatecli.ExitClean {
		t.Fatalf("the advisory must not change the verdict, got %d (stderr:\n%s)", captured.Code, captured.Stderr)
	}

	if !strings.Contains(captured.Stdout, "ONE") {
		t.Fatalf("heading candidates must carry the split-landing advisory, stdout:\n%s", captured.Stdout)
	}
}
