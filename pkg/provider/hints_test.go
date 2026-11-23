package provider

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
		"freshness":  "docs-gates freshness D",
		"audit-all":  "docs-gates freshness D --audit-all",
		"annotate":   "docs-gates annotate D",
		"routing":    "docs-gates routing D",
		"cadence":    "docs-gates cadence D",
		"stub":       "docs-gates stub D",
		"citations":  "docs-gates citations D",
		"stamps":     "docs-gates stamps D",
		"vocabulary": "docs-gates vocabulary D",
		"agents":     "docs-gates agents D",
	}

	for gate, want := range cases {
		if got := rerunHint(gate, "D"); got != want {
			t.Errorf("rerunHint(%q) = %q, want %q", gate, got, want)
		}
	}
}

// TestRerunHintTableCoversTheBatteryMatrix runs the REAL binary's
// battery over a clean corpus and asserts every matrix row's replay
// line names a verb the dispatcher ACCEPTS — probed through the verb's
// --help face (intercepted before any gate parser: a known verb exits
// 0, an unknown one exits 2), so the guard stays binary-driven with no
// hand list to lag behind a new gate. audit-all is the one non-verb
// matrix name: it replays through freshness's --audit-all mode.
func TestRerunHintTableCoversTheBatteryMatrix(t *testing.T) {
	t.Parallel()

	bin := findBinaryOrSkip(t)
	dir := t.TempDir()

	code, out, errOut, err := runBattery(context.Background(), bin, dir)
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

	for _, row := range report.Gates {
		replayVerb := row.Name
		if replayVerb == gateAuditAll {
			replayVerb = "freshness"
		}

		if !verbHelpExitsZero(t, bin, replayVerb) {
			t.Errorf(
				"matrix gate %q replays via %q, which the dispatcher rejects — extend rerunHint's vocabulary before this hint misleads",
				row.Name,
				replayVerb,
			)
		}
	}
}

// verbHelpExitsZero probes the dispatcher's help interception for one
// verb: help is intercepted BEFORE any gate parser, so a known verb
// prints its usage and exits 0 while an unknown one refuses with exit
// 2 — the cheapest honest accept/reject oracle, with no corpus and no
// side effects.
func verbHelpExitsZero(t *testing.T, bin, verb string) bool {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		bin,
		verb,
		"--help",
	)

	err := cmd.Run()
	if err == nil {
		return true
	}

	if _, ok := errors.AsType[*exec.ExitError](err); ok {
		return false // refused: the dispatcher does not know this verb
	}

	t.Skipf("docs-gates binary became unavailable mid-probe: %v", err)

	return false
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
