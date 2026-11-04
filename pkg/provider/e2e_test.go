package provider

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
)

// TestDetectEndToEndOverRealBinary is the provider-path battery run: it
// builds the REAL cmd/docs-gates from the root module, points Detect at a
// fixture repo whose battery is red in exactly one gate (freshness alone
// runs, over a repo with no reports), and asserts Detect yields exactly
// one finding naming that gate — the unit tests only ever exercise canned
// matrix text.
//
// Skips when the root module source is not adjacent (the nix provider
// check sandbox contains only pkg/provider) or no go toolchain is on
// PATH. Not parallel: t.Setenv is process-global.
func TestDetectEndToEndOverRealBinary(t *testing.T) {
	rootCmd := "../../cmd/docs-gates"
	if _, err := os.Stat(rootCmd); err != nil {
		t.Skipf("root module source not adjacent (%s missing) — nix check sandbox", rootCmd)
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not on PATH")
	}

	binDir := t.TempDir()
	bin := filepath.Join(binDir, toolName)

	build := exec.Command("go", "build", "-o", bin, "./cmd/docs-gates") //nolint:gosec,noctx // fixed subcommand, test-scoped
	build.Dir = "../.."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/docs-gates: %v\n%s", err, out)
	}

	repo := t.TempDir()
	config := `{"report_dir": "docs/status", "living_docs": ["NOTES.md"], "skip_gates": ["audit-all", "annotate", "routing", "cadence"]}`
	if err := os.WriteFile(filepath.Join(repo, ".docs-gates.json"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "docs", "status"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Harvest red, date-independent: the §f item's distinctive token
	// (qxvorlag) appears in no living doc, so freshness fails regardless
	// of report age. Zero reports would be a vacuous PASS.
	report := "# red fixture\n\n## f) open items\n\n1. the qxvorlag spindle reconciliation needs a pass\n"
	if err := os.WriteFile(filepath.Join(repo, "docs", "status", "2026-11-04_00-00_red.md"), []byte(report), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "NOTES.md"), []byte("# notes\n\nnothing about qxvorlag here\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	ctx := finding.WithWorkingDir(context.Background(), repo)
	findings, err := Provider.Detect.Detect(ctx)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("findings = %d (%v); want exactly 1 for the single red gate", len(findings), findings)
	}
	if findings[0].Rule != "freshness" {
		t.Fatalf("rule = %q; want freshness", findings[0].Rule)
	}
}
