package docsfreshness

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// Land-parity (L1-23, the 2026-11-11 retyped-bullet flip class): a
// --pre-flight SAFE verdict judged ONE batch of bytes, but nothing
// checked that the edit which LANDED was that batch — a retyped bullet
// sailed through on the strength of a verdict earned by different text.
// The fix closes the loop: every SAFE pre-flight records its candidate
// batch, and --land-parity verifies the recorded lines are on disk
// byte-exact. A divergence is a named refusal ("the landed edit is not
// the pre-flighted batch — re-run --pre-flight on the ACTUAL bytes"),
// never a silent pass.

const stampName = ".docs-gates-preflight.json"

// stampFileMode: the stamp is local editor state, never shared — owner-only
// reads/writes, and mnd wants the octal named.
const stampFileMode = 0o600

type preflightStamp struct {
	Recorded string   `json:"recorded"`
	Lines    []string `json:"lines"`
}

func stampPath(root string) string {
	return filepath.Join(root, stampName)
}

// writePreflightStamp records a SAFE candidate batch (one stamp file —
// the newest SAFE verdict owns it; a second pre-flight overwrites).
func writePreflightStamp(root string, lines []string) {
	stamp := preflightStamp{
		Recorded: time.Now().UTC().Format(time.RFC3339),
		Lines:    lines,
	}

	blob, err := json.MarshalIndent(stamp, "", "  ")
	if err != nil {
		// The stamp is a guard, never the gate: a marshal failure (empty
		// lines cannot fail, but defensive) only skips the parity ledger.
		return
	}

	blob = append(blob, '\n')

	_ = os.WriteFile(stampPath(root), blob, stampFileMode)
	warnIfStampShared(root)
}

// warnIfStampShared notes (stderr, best-effort) when the freshly written
// stamp is not git-ignored: it is local editor state, never shared, but
// a consumer repo without the .gitignore entry lets an auto-commit
// daemon sweep it into history (paid for in cmdguard 2026-11-14 — the
// entry existed in THIS repo's .gitignore and nowhere else). git
// check-ignore is the ground truth, and a TRACKED stamp still reports
// not-ignored — exactly the worst case, so the plain probe is the right
// one. Any probe failure (no git binary, not a checkout, sandbox) skips
// the note silently: the stamp is a guard, never a gate.
//
//nolint:gosec // git -C this repo; the subcommand and name are literals
func warnIfStampShared(root string) {
	probe := exec.CommandContext(
		context.Background(),
		"git",
		"-C",
		root,
		"check-ignore",
		"-q",
		stampName,
	)

	err := probe.Run()
	if err == nil {
		return
	}

	ee, ok := errors.AsType[*exec.ExitError](err)
	if !ok || ee.ExitCode() != 1 {
		return
	}

	gatecli.Errf(
		"docs-freshness: note: %s is not git-ignored — it is the land-parity ledger (local editor state, never shared); add it to .gitignore\n",
		stampName,
	)
}

// PreFlightParity verifies the last SAFE pre-flight batch landed
// verbatim: every recorded line must exist in the living docs through
// the SAME lens the verdict judged them — the corpus is lowercased
// (corpus.go), so the comparison lowercases and trims both sides. A
// raw byte-exact match would false-DRIFT every line carrying an
// uppercase letter (paid for 2026-11-14: cmdguard's first land-parity
// run refused a byte-identical landing). Exit 0 = parity (or no
// recorded batch — the guard is opt-in state); 1 = divergence, naming
// the lines that never landed as pre-flighted.
func PreFlightParity(root string, settings Settings) int {
	raw, err := os.ReadFile(stampPath(root))
	if err != nil {
		if os.IsNotExist(err) {
			gatecli.Outln("land-parity: no recorded pre-flight batch (run --pre-flight first); nothing to verify")

			return gatecli.ExitClean
		}

		gatecli.Errf("FATAL: reading %s: %v\n", stampName, err)

		return gatecli.ExitUsage
	}

	var stamp preflightStamp
	if err := json.Unmarshal(raw, &stamp); err != nil {
		gatecli.Errf("FATAL: %s is malformed (delete it and re-run --pre-flight): %v\n", stampName, err)

		return gatecli.ExitUsage
	}

	living := loadLivingDocs(root, settings)
	onDisk := make(map[string]bool, len(living.lines))

	for _, line := range living.lines {
		onDisk[strings.TrimSpace(line)] = true
	}

	var diverged []string

	for _, line := range stamp.Lines {
		if !onDisk[strings.ToLower(strings.TrimSpace(line))] {
			diverged = append(diverged, line)
		}
	}

	if len(diverged) > 0 {
		gatecli.Errln("land-parity: DRIFT — the landed edit diverged from the pre-flighted SAFE batch:")

		for _, line := range diverged {
			gatecli.Errf("    missing verbatim: %s\n", gatecli.Truncate(line, reasonMax))
		}

		gatecli.Errln("  the pre-flight verdict was earned by DIFFERENT bytes than the ones on disk")
		gatecli.Errln("  (the retyped-bullet class): re-run --pre-flight with the ACTUAL lines, or")
		gatecli.Errln("  strike the stamp file if the divergence is deliberate.")

		return gatecli.ExitFindings
	}

	gatecli.Outf(
		"land-parity: OK — %d recorded line(s) from the %s SAFE batch all landed verbatim\n",
		len(stamp.Lines), stamp.Recorded,
	)

	return gatecli.ExitClean
}
