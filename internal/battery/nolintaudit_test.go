package battery

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The nolint-audit flake check must be able to FAIL: a staleness gate
// that cannot go red is a no-op pass (the source fleet shipped three
// stale directives green for days before one session ran the audit).
// This plants a finding-free //nolint:erraudit directive in a temp
// module and asserts erraudit reports it STALE with a non-zero rc —
// the discrimination proof of 2026-11-04, now permanent and gated by
// the flake's nolint-audit check. Skips outside the devShell/flake
// (erraudit is a private-input binary, not on bare PATH).
func TestNolintAuditDiscrimination(t *testing.T) {
	t.Parallel()

	bin, err := exec.LookPath("erraudit")
	if err != nil {
		t.Skip("erraudit not on PATH — run inside `nix develop` (the flake checks provide it)")
	}

	dir := t.TempDir()

	mod := "module discrimination\n\ngo 1.24\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o600); err != nil {
		t.Fatal(err)
	}

	// helper() returns no error, so the directive suppresses nothing —
	// exactly the stale-directive class the audit exists to catch.
	prog := "package main\n\nfunc helper() {}\n\nfunc main() {\n\thelper() //nolint:erraudit // planted stale directive: no finding behind it\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(prog), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(bin, "nolint-audit", ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local")

	out, execErr := cmd.CombinedOutput()
	if execErr == nil {
		t.Fatalf("planted stale directive PASSED the audit — the check cannot go red\noutput: %s", out)
	}

	if !strings.Contains(string(out), "STALE") || !strings.Contains(string(out), "1 stale") {
		t.Fatalf("expected a STALE verdict for the planted directive, got:\n%s", out)
	}
}
