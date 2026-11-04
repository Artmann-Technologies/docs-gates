package docsfreshness

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// --duplicates outside pre-flight mode is a usage error: the assertion it
// encodes ("this on-disk line lands as another copy") only has meaning
// for the additions-only delta.
func TestDuplicatesFlagNeedsPreFlightMode(t *testing.T) {
	t.Parallel()

	if rc := Run([]string{".", "--duplicates"}); rc != gatecli.ExitUsage {
		t.Fatalf("--duplicates without --pre-flight must exit usage, got %d", rc)
	}
}

// The audit ratchet (TODO §91 coin audit-ratchet): --audit-max=N turns the
// audit red only on a REGRESSION past the pin — a legacy backlog nobody is
// draining in one sitting stays green but named, zero stays green, and a
// count past the pin (or a non-integer pin) still fails. The gate was
// permanently red 2026-10-08..23 while the backlog sat at 78 because any
// nonzero count failed.
func TestAuditMaxRatchet(t *testing.T) {
	t.Parallel()

	// One unharvested item: every token missing from the living docs.
	root := preFlightRoot(t)

	report := "# Report\n\n## f) Next\n\n1. wire the zaphod-beeblebrox relay\n"
	if err := os.WriteFile(
		filepath.Join(root, "docs", "status", "2026-09-02_00-00_GAP.md"),
		[]byte(report), 0o644,
	); err != nil {
		t.Fatal(err)
	}

	if rc := Run([]string{root, "--audit-all"}); rc != gatecli.ExitFindings {
		t.Fatalf("strict audit (no pin) must fail on 1 unharvested, got %d", rc)
	}

	if rc := Run([]string{root, "--audit-all", "--audit-max=1"}); rc != gatecli.ExitClean {
		t.Fatalf("1 unharvested under a --audit-max=1 pin must stay green, got %d", rc)
	}

	if rc := Run([]string{root, "--audit-all", "--audit-max=0"}); rc != gatecli.ExitFindings {
		t.Fatalf("1 unharvested past a --audit-max=0 pin is a REGRESSION, got %d", rc)
	}

	if rc := Run([]string{root, "--audit-all", "--audit-max=bogus"}); rc != gatecli.ExitUsage {
		t.Fatalf("a non-integer pin must exit usage, got %d", rc)
	}

	// Legacy-zero stays green under any pin.
	fresh := preFlightRoot(t)
	if rc := Run([]string{fresh, "--audit-all", "--audit-max=0"}); rc != gatecli.ExitClean {
		t.Fatalf("a fresh corpus stays green under --audit-max=0, got %d", rc)
	}
}

// TestSkipMarkerAcceptsSuffixedForm pins the 2026-10-25 parser gap: reports
// carry the documented suffixed markers (e.g. "<!-- fresh:skip: already
// routed in TODO_LIST -->") while the byte-exact substring match silently
// ignored them, resurfacing routed items as unharvested on --audit-all.
func TestSkipMarkerAcceptsSuffixedForm(t *testing.T) {
	t.Parallel()

	root := preFlightRoot(t)

	report := "# Report\n\n## f) Next\n\n1. beta-probe routed elsewhere <!-- fresh:skip: already routed in TODO_LIST -->\n2. beta-probe next-session <!-- fresh:skip: next-session new work -->\n"
	if err := os.WriteFile(
		filepath.Join(root, "docs", "status", "2026-09-02_00-00_SUFFIXED.md"),
		[]byte(report), 0o644,
	); err != nil {
		t.Fatal(err)
	}

	if rc := Run([]string{root, "--audit-all", "--audit-max=0"}); rc != gatecli.ExitClean {
		t.Fatalf("suffixed fresh:skip markers must exempt their items (audit-max=0), got %d", rc)
	}
}
