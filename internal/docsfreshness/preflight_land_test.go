package docsfreshness

import (
	"os"
	"os/exec"
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

// The corpus lens is lowercased (corpus.go), so parity must compare
// through it: the stamp records the line as TYPED, and a raw byte-exact
// match false-DRIFTed every mixed-case landing (cmdguard 2026-11-14 —
// the disk line was byte-identical and parity still refused it).
func TestLandParityPassesCaseInsensitiveLanding(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	stampLine := "- [ ] Some-New-Row: The Pre-Flighted Wording (coin: X)"
	writePreflightStamp(root, []string{stampLine})

	settings := Settings{}
	settings.LivingDocs = []string{"TODO_LIST.md"}

	if err := os.WriteFile(filepath.Join(root, "TODO_LIST.md"), []byte(stampLine+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if rc := PreFlightParity(root, settings); rc != gatecli.ExitClean {
		t.Fatalf("expected parity OK across the corpus lowercase lens, got rc=%d", rc)
	}
}

// Trailing whitespace is formatter territory, not content: the landing
// comparison rides the same trim+lower lens as the --replace matching.
func TestLandParityIgnoresTrailingWhitespace(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	stampLine := "- [ ] some-new-row: the pre-flighted wording (coin: x)"
	writePreflightStamp(root, []string{stampLine + "   "})

	settings := Settings{}
	settings.LivingDocs = []string{"TODO_LIST.md"}

	if err := os.WriteFile(filepath.Join(root, "TODO_LIST.md"), []byte(stampLine+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if rc := PreFlightParity(root, settings); rc != gatecli.ExitClean {
		t.Fatalf("expected parity OK modulo trailing whitespace, got rc=%d", rc)
	}
}

// The un-ignored-stamp note (cmdguard 2026-11-14: an auto-commit daemon
// swept the stamp into history because only THIS repo's .gitignore knew
// the name). git fixtures prove both arms: no .gitignore → the note
// fires; the .gitignore entry → silent.
//
//nolint:paralleltest // process-global capture, serialized on purpose
func TestStampWriteWarnsWhenStampNotIgnored(t *testing.T) {
	root := t.TempDir()
	gitInit(t, root)

	stampLine := "- [ ] some-new-row: the pre-flighted wording (coin: x)"

	// CaptureOutput swaps process-global stdout: this test stays serial.
	captured := gatecli.CaptureOutput(func() int {
		writePreflightStamp(root, []string{stampLine})

		return 0
	})

	if !strings.Contains(captured.Stderr, "not git-ignored") {
		t.Fatalf("expected the not-git-ignored note in an un-ignored repo:\n%s", captured.Stderr)
	}

	gitignore := ".docs-gates-preflight.json\n"
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(gitignore), 0o644); err != nil {
		t.Fatal(err)
	}

	captured = gatecli.CaptureOutput(func() int {
		writePreflightStamp(root, []string{stampLine})

		return 0
	})

	if strings.Contains(captured.Stderr, "not git-ignored") {
		t.Fatalf("expected silence once the stamp is ignored:\n%s", captured.Stderr)
	}
}

func gitInit(t *testing.T, dir string) {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), "git", "init", "-q", dir)

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init %s: %v: %s", dir, err, out)
	}
}
