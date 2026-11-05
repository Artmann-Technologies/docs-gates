package docsfreshness

import (
	"regexp"
	"slices"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// PreFlight answers "would these NEW living-doc lines flip a historical
// §f item to DRIFT?" BEFORE the edit lands: the full --audit-all replay
// runs against the corpus with and without the candidate lines, and the
// DELTA is the verdict — items that would newly fail are listed with
// their reasons (plus token tables under explain). Boundary-crossing
// tokens are always listed: they are the leading indicator even when no
// item flips yet. The default model is additions-only (new rows, new
// bullets); with --replace=OLD (repeatable), the old lines are removed
// first so a REWRITE is judged by its new words only — an item anchored
// only by the replaced line still flips, which the additions-only model
// cannot see. Exit 0 = safe, 1 = the edit would introduce drift, 2 =
// usage (no candidate lines, or a replace target not on disk).
func PreFlight(
	root string,
	candidate []string,
	replaces []string,
	explain, allowDuplicates bool,
	settings Settings,
) int {
	if len(candidate) == 0 {
		gatecli.Errln("FATAL: --pre-flight needs the candidate line(s) on stdin — " +
			"the exact new lines about to land in TODO_LIST/ROADMAP/CHANGELOG")

		return gatecli.ExitUsage
	}

	base := loadLivingDocs(root, settings)

	// Rewrite-mode removes the replaced lines FIRST: the delta then judges
	// the candidate words only — the old words' DF presence and any
	// harvest anchor they carried vanish with the line being replaced.
	// The DELTA BASELINE stays the pre-edit corpus: an item that passes
	// before and fails after the edit is the rewrite's damage, while one
	// already failing before is not blamed on it.
	remaining, missing := applyReplacements(base.lines, replaces)
	if len(missing) > 0 {
		gatecli.Errln("docs-freshness --pre-flight: --replace target(s) not found in the living docs:")

		for _, m := range missing {
			gatecli.Errf("    %s\n", gatecli.Truncate(m, reasonMax))
		}

		return gatecli.ExitUsage
	}

	if allowDuplicates {
		// The operator-side dedup workaround (grep -Fxvf) silently DROPS the
		// +1 DF push of a line arriving as a second copy — exactly the push a
		// tail-ledger block makes. --duplicates is the explicit assertion "I
		// know this line is on disk; another copy is about to land", so the
		// delta models the copy instead of refusing the question.
		for _, d := range onDiskLines(candidate, remaining) {
			gatecli.Outf("  second copy: %-52s stays modeled — every token on it gains +1 DF\n",
				gatecli.Truncate(d, reasonMax))
		}
	} else if refuseAlreadyOnDisk(candidate, remaining) {
		return gatecli.ExitUsage
	}

	if hasHeadingLine(candidate) {
		gatecli.Outln("  note: candidate carries heading line(s) — land the heading and its batch through ONE")
		gatecli.Outln("        pre-flight run; split runs model only each part's DF push, and a boundary the")
		gatecli.Outln("        combined push crosses is invisible to both halves")
	}

	modLines := append(slices.Clone(remaining), lowerAll(candidate)...)
	mod := NewCorpus(modLines)
	candLower := modLines[len(remaining):]
	candSquashed := squashLines(candLower)

	reports := statusReports(root, settings)
	baseFailing := failingKeys(reports, base, settings)
	newly := newlyFailing(reports, mod, baseFailing, settings)

	gatecli.Outf("docs-freshness --pre-flight: %d candidate line(s), replayed over %d report(s)\n",
		len(candidate), len(reports))

	for _, c := range BoundaryCrossings(candidate, base.lines, mod.lines, settings.df()) {
		gatecli.Outf("  boundary token: %-28s DF %d→%d (genericizes every item whose only "+
			"distinctive token this was)\n", c.Token, c.Before, c.After)
	}

	printNewly(newly, explain, mod, candLower, candSquashed, settings.df())
	printBindings(reports, mod, baseFailing, candLower, candSquashed, settings.df())

	if len(newly) > 0 {
		gatecli.Errf("docs-freshness --pre-flight: UNSAFE — %d historical item(s) would flip to "+
			"DRIFT; reword the candidate line(s) before landing them\n", len(newly))

		return gatecli.ExitFindings
	}

	gatecli.Outln("docs-freshness --pre-flight: SAFE — no historical item flips to DRIFT")

	return gatecli.ExitClean
}

// Crossing is one candidate-line token whose living-docs document
// frequency crosses the GenericDF threshold when the candidate lines
// are added: before it kept some historical item's last distinctive
// token distinctive, after it the token is generic.
type Crossing struct {
	Token  string
	Before int
	After  int
}

// BoundaryCrossings returns the tokens of the candidate lines whose DF
// crosses the GenericDF boundary between the baseline corpus and the
// corpus-with-candidates — the exact mechanics of the 2026-10-07 burn
// class (a fresh CHANGELOG line's "decision batch" pushed `batch` from
// DF 7 to 8 and silently genericized a 2026-10-03 item). Tokens that
// were already generic, or stay below the threshold, do not cross.
func BoundaryCrossings(candidate, baseLines, modLines []string, genericDF int) []Crossing {
	baseDF := tokenDFOver(baseLines)
	modDF := tokenDFOver(modLines)

	var crossings []Crossing

	seen := map[string]bool{}

	for _, line := range candidate {
		for _, tok := range Tokens(line) {
			if seen[tok] {
				continue
			}

			seen[tok] = true

			before, after := baseDF(tok), modDF(tok)
			if before < genericDF && after >= genericDF {
				crossings = append(crossings, Crossing{Token: tok, Before: before, After: after})
			}
		}
	}

	return crossings
}

// applyReplacements removes each replace target from the corpus
// (lowercased trimmed exact-LINE match — a rewrite removes the whole old
// line, not a substring) and returns the targets not found on disk for
// the usage refusal.
func applyReplacements(baseLines, replaces []string) ([]string, []string) {
	if len(replaces) == 0 {
		return baseLines, nil
	}

	onDisk := make(map[string]bool, len(baseLines))

	for _, l := range baseLines {
		onDisk[strings.TrimSpace(l)] = true
	}

	var missing []string

	drop := make(map[string]bool, len(replaces))

	for _, r := range replaces {
		key := strings.ToLower(strings.TrimSpace(r))
		if !onDisk[key] {
			missing = append(missing, r)

			continue
		}

		drop[key] = true
	}

	if len(missing) > 0 {
		return baseLines, missing
	}

	out := make([]string, 0, len(baseLines))

	for _, l := range baseLines {
		if !drop[strings.TrimSpace(l)] {
			out = append(out, l)
		}
	}

	return out, nil
}

// headingLineRe matches a markdown ATX heading candidate line
// (`# Example` .. `###### Example`).
var headingLineRe = regexp.MustCompile(`^#{1,6}\s`)

// hasHeadingLine reports whether any candidate line is a markdown
// heading — the split-landing hazard: pre-flight the heading and its
// batch in one run or the combined DF push goes unmodeled.
func hasHeadingLine(candidate []string) bool {
	return slices.ContainsFunc(candidate, headingLineRe.MatchString)
}

// onDiskLines returns the candidate lines already present verbatim in
// the living-docs corpus (lowercased substring match, blank lines never
// count). Shared by the refusal and the --duplicates accounting.
func onDiskLines(candidate, baseLines []string) []string {
	joined := "\n" + strings.Join(baseLines, "\n") + "\n"

	var onDisk []string

	for _, line := range candidate {
		trimmed := strings.ToLower(strings.TrimSpace(line))
		if trimmed != "" && strings.Contains(joined, trimmed) {
			onDisk = append(onDisk, line)
		}
	}

	return onDisk
}

// refuseAlreadyOnDisk reports (and explains) candidate lines already
// present verbatim in the living-docs corpus: the additions-only delta
// model needs the PRE-edit tree, and a post-write run (the 2026-10-10
// burn) replays its own additions against a baseline that already
// contains them — a would-be flip then hides as SAFE, so the question
// is refused instead of answered. Rewrite-mode candidates that REPLACE
// an on-disk line are exempt: the model removes that line first.
func refuseAlreadyOnDisk(candidate, remaining []string) bool {
	onDisk := onDiskLines(candidate, remaining)

	if len(onDisk) == 0 {
		return false
	}

	gatecli.Errln("docs-freshness --pre-flight: candidate line(s) ALREADY on disk — the additions-only model")
	gatecli.Errln("  needs the PRE-edit tree; once landed, baseline and modified corpus coincide and a would-be")
	gatecli.Errln("  flip hides as SAFE. Run pre-flight BEFORE writing (or against a stashed/HEAD tree).")
	gatecli.Errln("  If the line is about to land as a deliberate SECOND copy (a tail-ledger block), rerun with")
	gatecli.Errln("  --duplicates to model its +1 DF push instead of refusing. Present:")

	for _, d := range onDisk {
		gatecli.Errf("    %s\n", gatecli.Truncate(d, reasonMax))
	}

	return true
}

// lowerAll lowercases the candidate lines to match the living-docs
// corpus basis (all comparisons are case-insensitive).
func lowerAll(lines []string) []string {
	out := make([]string, 0, len(lines))

	for _, l := range lines {
		out = append(out, strings.ToLower(l))
	}

	return out
}
