package docspurge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// TestCollectStruckRefusesOpenRowStrikes is the io-surface incident
// (#10): a concurrent session left stray ~~ fragments on OPEN rows and
// the collector swept them into CHANGELOG as done work, deleting open
// tasks. A struck table row whose status cell still reads OPEN must be
// refused, never collected.
func TestCollectStruckRefusesOpenRowStrikes(t *testing.T) {
	t.Parallel()

	content := fixtureTodo +
		"| ~~172~~ Extract the mail client seam | 🔴 `TODO` |\n"

	rows, guards := collectStruck(content)

	for _, row := range rows {
		if strings.Contains(row.oldLine, "mail client seam") {
			t.Fatalf("an OPEN row's strike must never collect: %+v", row)
		}
	}

	if len(guards) != 1 || !guards[0].refuse {
		t.Fatalf("the open-row strike must yield exactly one refusal, got %+v", guards)
	}

	if !strings.Contains(guards[0].cell, "TODO") {
		t.Errorf("the refusal must name the open status cell, got %q", guards[0].cell)
	}
}

// TestRunRefusedOpenRowStrikeWritesNothing: the refusal is an
// environment error (exit 2) and nothing is written in either mode.
//
//nolint:paralleltest // process-global capture swap, serialized on purpose
func TestRunRefusedOpenRowStrikeWritesNothing(t *testing.T) {
	root := writeFixtures(t)

	noisy := fixtureTodo + "| ~~172~~ Extract the mail client seam | 🟡 `IN_PROGRESS` |\n"
	if err := os.WriteFile(filepath.Join(root, "TODO_LIST.md"), []byte(noisy), 0o644); err != nil {
		t.Fatal(err)
	}

	before, _ := readFixtures(t, root)

	captured := gatecli.CaptureOutput(func() int {
		return Run([]string{root, "--apply"})
	})

	if captured.Code != gatecli.ExitUsage {
		t.Fatalf("a refused open-row strike is an environment error, code = %d", captured.Code)
	}

	if !strings.Contains(captured.Stderr, "REFUSED") {
		t.Errorf("the refusal must be loud on stderr:\n%s", captured.Stderr)
	}

	after, _ := readFixtures(t, root)
	if after != before {
		t.Fatal("a refused purge must not touch the TODO file")
	}
}

// TestCollectStruckWarnsOnUnrecognizedStatus: a struck row whose status
// cell is neither DONE nor OPEN (a `Won't implement` verdict, an
// unrecognized spelling) collects under a warning — advisory, mirroring
// how absorb drift stays advisory.
func TestCollectStruckWarnsOnUnrecognizedStatus(t *testing.T) {
	t.Parallel()

	content := "| ~~Retire the legacy parser~~ done (superseded) | 🚫 `Won't implement` |\n"

	rows, guards := collectStruck(content)

	if len(rows) != 1 {
		t.Fatalf("an unrecognized-status strike still collects, rows = %d", len(rows))
	}

	if len(guards) != 1 || guards[0].refuse {
		t.Fatalf("an unrecognized status cell warns, never refuses, got %+v", guards)
	}
}
