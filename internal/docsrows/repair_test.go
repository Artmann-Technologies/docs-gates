package docsrows

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// TestRepairMangleCompatibility proves the split's boundary: the writer's
// repair verb still owns mutation, and this read-only port agrees with what
// it produces. A repaired fixture must be clean under `rows`.
//
//nolint:paralleltest // gatecli.CaptureOutput swaps process-global streams
func TestRepairMangleCompatibility(t *testing.T) {
	requireWriter(t)

	path := copyFixture(t, "testdata/mangle-fixable.md")

	// The pre-repair fixture is already strike-bearing; the mangle is in
	// the cell SHAPE, not the strike absence. The writer must report work.
	out, code := runWriter(t, []string{"repair-mangle", path})
	if code != 0 {
		t.Fatalf("repair-mangle exited %d: %s", code, out)
	}

	if !strings.Contains(out, "repaired") {
		t.Fatalf("repair-mangle did not report a repair: %s", out)
	}

	repaired, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	// Idempotence: a second run must change nothing and still exit clean.
	again, code := runWriter(t, []string{"repair-mangle", path})
	if code != 0 || !strings.Contains(again, "repaired 0 row(s)") {
		t.Fatalf("second repair-mangle must be a no-op, got rc=%d: %s", code, again)
	}

	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(repaired) != string(second) {
		t.Fatal("repair-mangle is not idempotent: bytes changed on the second run")
	}

	// The read-only gate must call the writer's canonical output clean.
	if res := gatecli.CaptureOutput(func() int { return Run([]string{path}) }); res.Code != gatecli.ExitClean {
		t.Fatalf("rows after repair = %d, want clean\n%s", res.Code, res.Stdout)
	}
}

// TestContentLossStaysWithTheWriter pins WHY the mutator is not ported: a
// tilde quad at a cell boundary swallowed text that no mechanical pass can
// recover, so the writer refuses (exit 1) and demands a human. The reader
// sees only the surviving strike, which is exactly why this check cannot
// move into a gate.
//
//nolint:paralleltest // gatecli.CaptureOutput swaps process-global streams
func TestContentLossStaysWithTheWriter(t *testing.T) {
	requireWriter(t)

	path := copyFixture(t, "testdata/mangle-lossy.md")

	out, code := runWriter(t, []string{"repair-mangle", path})
	if code != 1 {
		t.Fatalf("content-loss mangle must exit 1, got %d: %s", code, out)
	}

	if !strings.Contains(out, "content-loss") {
		t.Fatalf("refusal must name the content-loss class: %s", out)
	}

	// The read-only classifier cannot detect fabricated/lost text; it only
	// sees the mangled cell shape.
	res := gatecli.CaptureOutput(func() int { return Run([]string{path}) })
	if res.Code != gatecli.ExitFindings {
		t.Fatalf("rows on the lossy fixture = %d, want findings (PARTIAL row)", res.Code)
	}
}

// requireWriter skips the test where the local docs-health fan-out is absent.
func requireWriter(t *testing.T) {
	t.Helper()

	if _, err := os.Stat(writerSource); err != nil {
		t.Skipf("docs-health writer tool not present at %s (local fan-out only)", writerSource)
	}

	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not on PATH: cannot run the writer oracle")
	}
}

// copyFixture copies an embedded test fixture into a temp dir so the writer
// can mutate the copy without touching the repo.
func copyFixture(t *testing.T, rel string) string {
	t.Helper()

	data, err := os.ReadFile(rel)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), filepath.Base(rel))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	return path
}
