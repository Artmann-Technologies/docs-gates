package docsfreshness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// Land-parity (L1-23): the SAFE batch is recorded and the landing is
// verified byte-exact — the retyped-bullet flip class dies.
func TestLandParityPassesWhenBatchLanded(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	lines := []string{"- [ ] some-new-row: the pre-flighted wording (coin: x)"}
	writePreflightStamp(root, lines)

	settings := Settings{}
	settings.LivingDocs = []string{"TODO_LIST.md"}

	// The living docs carry the batch verbatim.
	if err := os.WriteFile(filepath.Join(root, "TODO_LIST.md"), []byte(lines[0]+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if rc := PreFlightParity(root, settings); rc != gatecli.ExitClean {
		t.Fatalf("expected parity OK, got rc=%d", rc)
	}
}

func TestLandParityFailsOnRetypedLanding(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writePreflightStamp(root, []string{"- [ ] some-new-row: the pre-flighted wording (coin: x)"})

	settings := Settings{}
	settings.LivingDocs = []string{"TODO_LIST.md"}

	// The landing RETYPED the row — different bytes than the SAFE batch.
	if err := os.WriteFile(
		filepath.Join(root, "TODO_LIST.md"),
		[]byte("- [ ] some-new-row: the retyped wording that actually landed\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	captured := gatecli.CaptureOutput(func() int { return PreFlightParity(root, settings) })
	rc, out := captured.Code, captured.Stdout+captured.Stderr
	if rc != gatecli.ExitFindings {
		t.Fatalf("expected parity FAIL, got rc=%d", rc)
	}

	for _, want := range []string{"land-parity: DRIFT", "missing verbatim", "pre-flighted wording"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestLandParityNoStampIsOK(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if rc := PreFlightParity(root, Settings{}); rc != gatecli.ExitClean {
		t.Fatalf("expected OK with no stamp, got rc=%d", rc)
	}
}

func TestLandParityMalformedStampIsUsage(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(stampPath(root), []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	if rc := PreFlightParity(root, Settings{}); rc != gatecli.ExitUsage {
		t.Fatalf("expected usage exit on malformed stamp, got rc=%d", rc)
	}
}
