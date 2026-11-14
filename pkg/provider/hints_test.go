package provider

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestRerunHintTable pins the gate-name → replay-line vocabulary: every
// battery matrix name must rerun through a command the dispatcher
// ACCEPTS. audit-all is the known non-verb (the freshness gate's
// --audit-all replay mode); the rest replay as bare verbs. A NEW battery
// gate name must extend this table — the naive fallback would advise
// `docs-gates <new-name>`, an unknown-verb exit 2.
func TestRerunHintTable(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"freshness": "docs-gates freshness D",
		"audit-all": "docs-gates freshness D --audit-all",
		"annotate":  "docs-gates annotate D",
		"routing":   "docs-gates routing D",
		"cadence":   "docs-gates cadence D",
	}

	for gate, want := range cases {
		if got := rerunHint(gate, "D"); got != want {
			t.Errorf("rerunHint(%q) = %q, want %q", gate, got, want)
		}
	}
}

// TestRerunHintTableCoversTheBatteryMatrix runs the REAL binary's
// battery over a clean corpus and asserts every matrix row's name is
// known to rerunHint's special-case vocabulary (or a plain verb) — the
// silent-mis-hint guard driven from the binary side, not a hand list.
func TestRerunHintTableCoversTheBatteryMatrix(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	code, out, errOut, err := runBattery(context.Background(), findBinaryOrSkip(t), dir)
	if err != nil {
		t.Skipf("docs-gates binary unavailable: %v", err)
	}

	if code != 0 && code != 1 {
		t.Skipf("battery did not produce a verdict document over %s: %s", dir, errOut)
	}

	report, decodeErr := decodeReport(out)
	if decodeErr != nil {
		t.Skipf("binary does not emit the pinned --json document (version skew): %v", decodeErr)
	}

	known := map[string]bool{
		"freshness": true,
		"audit-all": true,
		"annotate":  true,
		"routing":   true,
		"cadence":   true,
	}

	for _, row := range report.Gates {
		if !known[row.Name] {
			t.Errorf(
				"battery matrix names unknown gate %q — extend rerunHint's vocabulary before this hint misleads",
				row.Name,
			)
		}
	}
}

// findBinaryOrSkip resolves the real docs-gates binary for the
// binary-driven guards: PATH first, then the session's usual build
// output. Tests skip (not fail) where no binary exists — the CI class.
func findBinaryOrSkip(t *testing.T) string {
	t.Helper()

	if path, err := exec.LookPath(toolName); err == nil {
		return path
	}

	candidates := []string{
		filepath.Join("..", "..", "docs-gates"),
		"/tmp/dg-bin",
	}

	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && !info.IsDir() && info.Mode()&0o111 != 0 {
			return c
		}
	}

	t.Skip("no docs-gates binary on PATH or in the known build outputs")

	return ""
}

// TestInstallHintPinsTheReleaseTag guards the §UX rule: the install
// hint must name a PINNED version, never @latest or a floating nix ref
// — a hint that floats can install a battery different from the one
// that produced the verdict (the version-skew class).
func TestInstallHintPinsTheReleaseTag(t *testing.T) {
	t.Parallel()

	if strings.Contains(installHint, "@latest") {
		t.Fatal("install hint must pin a version, got @latest")
	}

	if !strings.Contains(installHint, "@v0.") || !strings.Contains(installHint, "/v0.") {
		t.Fatalf("install hint must carry pinned @vX.Y.Z on both install lines, got %q", installHint)
	}
}
