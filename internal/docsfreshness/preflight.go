package docsfreshness

import (
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
// item flips yet. Only ADDITIONS are modeled (new rows, new bullets);
// rewording an EXISTING line needs the old line removed too, which this
// mode does not see. Exit 0 = safe, 1 = the edit would introduce drift,
// 2 = usage (no candidate lines).
func PreFlight(root string, candidate []string, explain, allowDuplicates bool, settings Settings) int {
	if len(candidate) == 0 {
		gatecli.Errln("FATAL: --pre-flight needs the candidate line(s) on stdin — " +
			"the exact new lines about to land in TODO_LIST/ROADMAP/CHANGELOG")

		return gatecli.ExitUsage
	}

	baseLines, baseSquashed, baseDF := loadLivingDocs(root, settings)

	if allowDuplicates {
		// The operator-side dedup workaround (grep -Fxvf) silently DROPS the
		// +1 DF push of a line arriving as a second copy — exactly the push a
		// tail-ledger block makes. --duplicates is the explicit assertion "I
		// know this line is on disk; another copy is about to land", so the
		// delta models the copy instead of refusing the question.
		for _, d := range onDiskLines(candidate, baseLines) {
			gatecli.Outf("  second copy: %-52s stays modeled — every token on it gains +1 DF\n",
				gatecli.Truncate(d, reasonMax))
		}
	} else if refuseAlreadyOnDisk(candidate, baseLines) {
		return gatecli.ExitUsage
	}

	modLines := append(slices.Clone(baseLines), lowerAll(candidate)...)
	modSquashed := squashLines(modLines)
	modDF := tokenDFOver(modLines)
	candLower := modLines[len(baseLines):]
	candSquashed := squashLines(candLower)

	reports := statusReports(root, settings)
	baseFailing := failingKeys(reports, baseLines, baseDF, baseSquashed, settings)
	newly := newlyFailing(reports, modLines, modDF, modSquashed, baseFailing, settings)

	gatecli.Outf("docs-freshness --pre-flight: %d candidate line(s), replayed over %d report(s)\n",
		len(candidate), len(reports))

	for _, c := range BoundaryCrossings(candidate, baseLines, modLines, settings.df()) {
		gatecli.Outf("  boundary token: %-28s DF %d→%d (genericizes every item whose only "+
			"distinctive token this was)\n", c.Token, c.Before, c.After)
	}

	printNewly(newly, explain, modLines, modSquashed, modDF, candLower, candSquashed, settings.df())
	printBindings(reports, modLines, modDF, modSquashed, baseFailing, candLower, candSquashed, settings.df())

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
// is refused instead of answered.
func refuseAlreadyOnDisk(candidate, baseLines []string) bool {
	onDisk := onDiskLines(candidate, baseLines)

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
