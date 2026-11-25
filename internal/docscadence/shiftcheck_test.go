package docscadence

import (
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// shiftFixtureHeader is the fleet's canonical 5-column TODO table header
// (Task/Status/Priority/Effort/Evidence) with its separator — the
// governing shape every fixture row below is judged against.
const shiftFixtureHeader = "| Task | Status | Priority | Effort | Evidence |\n" +
	"| ---- | ------ | -------- | ------ | -------- |\n"

// TestScanStrikeFaultsDetectsBareWrapShift proves the nine-row shape
// (2026-11-24 TODO_LIST §3) is named: the Task cell holds a bare `~~`
// while the actual task text sits unwrapped in the Status column. The
// old classify counted every such row as healthily struck.
func TestScanStrikeFaultsDetectsBareWrapShift(t *testing.T) {
	t.Parallel()

	body := shiftFixtureHeader +
		"| ~~ | Float SystemNix to the hardened flake | 🟡 PREPPED | P2 | M |\n" +
		"| ~~ | Fix the lockfile drift | 🟡 PARTIAL | P3 | S |\n"

	faults := scanStrikeFaults(body)
	if len(faults) != 2 {
		t.Fatalf("both bare-wrap shifts must fault, got %d: %v", len(faults), faults)
	}

	if faults[0].line != 3 || faults[1].line != 4 {
		t.Errorf("faults must name their 1-based body lines, got %d and %d", faults[0].line, faults[1].line)
	}

	for _, f := range faults {
		if !strings.Contains(f.reason, "wrap carries no text") {
			t.Errorf("the reason must name the empty wrap: %q", f.reason)
		}
	}
}

// TestScanStrikeFaultsDetectsStrikeOutsideTaskCell proves a strike wrap
// landing in any non-Task cell (here: Evidence) faults — the strike's
// text is not where the writer rituals put it.
func TestScanStrikeFaultsDetectsStrikeOutsideTaskCell(t *testing.T) {
	t.Parallel()

	body := shiftFixtureHeader +
		"| Land the provider pin bump | ✅ `DONE` | P3 | S | ~~done by reference: erraudit carries it~~ |\n"

	faults := scanStrikeFaults(body)
	if len(faults) != 1 {
		t.Fatalf("the misplaced wrap must fault, got %d: %v", len(faults), faults)
	}

	if !strings.Contains(faults[0].reason, "Evidence") {
		t.Errorf("the reason must name the misplaced cell's header: %q", faults[0].reason)
	}
}

// TestScanStrikeFaultsDetectsSurplusCell proves the wrap-crossed-pipe
// shape (the verdict appended as an EXTRA cell) faults on the cell
// count against the governing header.
func TestScanStrikeFaultsDetectsSurplusCell(t *testing.T) {
	t.Parallel()

	body := shiftFixtureHeader +
		"| ~~ | Write the note | 🔴 `TODO` | P3 | S | ~~ done by reference ~~ |\n"

	faults := scanStrikeFaults(body)
	if len(faults) != 1 {
		t.Fatalf("the surplus-cell row must fault, got %d: %v", len(faults), faults)
	}

	if !strings.Contains(faults[0].reason, "cells against the header's") {
		t.Errorf("the reason must name the count mismatch: %q", faults[0].reason)
	}
}

// TestScanStrikeFaultsDetectsStruckBeforePipe proves the shape whose
// wrap starts BEFORE the leading pipe faults: classify never sees the
// line at all (not |-prefixed), so the row is invisible to the struck
// count entirely — the worst blindness of the three.
func TestScanStrikeFaultsDetectsStruckBeforePipe(t *testing.T) {
	t.Parallel()

	body := shiftFixtureHeader + "~~| The wrap swallowed the leading pipe | 🟡 PARTIAL | P3 | S |\n"

	faults := scanStrikeFaults(body)
	if len(faults) != 1 {
		t.Fatalf("the struck-before-pipe row must fault, got %d: %v", len(faults), faults)
	}

	if !strings.Contains(faults[0].reason, "leading pipe") {
		t.Errorf("the reason must name the swallowed pipe: %q", faults[0].reason)
	}
}

// TestScanStrikeFaultsKeepsCanonicalStrikesSilent pins the positive
// shape: a wrap around the Task cell's text plus a DONE status cell is
// the docs-strike output and must never fault.
func TestScanStrikeFaultsKeepsCanonicalStrikesSilent(t *testing.T) {
	t.Parallel()

	body := shiftFixtureHeader +
		"| ~~Float SystemNix to the hardened flake~~ | ✅ `DONE` | P2 | M | tagged v0.4.0 |\n" +
		"| ~~Fix the lockfile drift~~ | ✅ DONE 2026-11-04 | P3 | S | flake.lock restored |\n"

	if faults := scanStrikeFaults(body); len(faults) != 0 {
		t.Fatalf("canonical strikes must stay silent, got %v", faults)
	}
}

// TestScanStrikeFaultsIgnoresOpenAndForeignTables pins the scope: rows
// without a strike, tables without a Task column, and prose lines stay
// out of the fault list.
func TestScanStrikeFaultsIgnoresOpenAndForeignTables(t *testing.T) {
	t.Parallel()

	body := shiftFixtureHeader +
		"| An open row with stray tildes in prose | 🔴 `TODO` | P3 | S | not a strike ~~yet~~ |\n" +
		"\n" +
		"| Name | Owner |\n| ---- | ----- |\n| ~~weird~~ | someone |\n"

	if faults := scanStrikeFaults(body); len(faults) != 0 {
		t.Fatalf("open rows and Task-less tables must stay silent, got %v", faults)
	}
}

// TestCadenceRunFailsOnShiftedStrikes wires the detector to the gate's
// face: the planted fixture turns the gate red, the findings name the
// offending file and line, and a clean ledger keeps the verdict green.
//
//nolint:paralleltest // process-global capture swap, serialized on purpose
func TestCadenceRunFailsOnShiftedStrikes(t *testing.T) {
	t.Run("shifted ledger fails naming file and line", func(t *testing.T) {
		root := writeTodo(t, shiftFixtureHeader+
			"| ~~ | Float SystemNix to the hardened flake | 🟡 PREPPED | P2 | M |\n")

		captured := gatecli.CaptureOutput(func() int {
			return Run([]string{root})
		})

		if captured.Code != gatecli.ExitFindings {
			t.Fatalf("a shifted strike must fail the gate, code = %d\n%s", captured.Code, captured.Stderr)
		}

		if !strings.Contains(captured.Stderr, "TODO_LIST.md:5") {
			t.Errorf("the finding must name the todo file and 1-based line:\n%s", captured.Stderr)
		}
	})

	t.Run("clean ledger stays green", func(t *testing.T) {
		root := writeTodo(t, shiftFixtureHeader+
			"| ~~Float SystemNix to the hardened flake~~ | ✅ `DONE` | P2 | M | tagged v0.4.0 |\n")

		captured := gatecli.CaptureOutput(func() int {
			return Run([]string{root})
		})

		if captured.Code != gatecli.ExitClean {
			t.Fatalf("a canonical strike must stay green, code = %d\n%s", captured.Code, captured.Stderr)
		}
	})
}
