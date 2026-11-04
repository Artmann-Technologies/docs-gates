package docsfreshness

import (
	"path/filepath"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// itemKey identifies one open §f item by its report basename and raw text.
type itemKey struct{ report, item string }

// failingKeys indexes every open §f item the corpus FAILS to reflect, so a
// delta replay can tell "already failing" from "newly failing".
func failingKeys(
	reports, lines []string,
	tokenDF func(string) int,
	squashed []string,
	settings Settings,
) map[itemKey]bool {
	out := map[itemKey]bool{}

	for _, report := range reports {
		for _, f := range CheckReport(report, lines, tokenDF, squashed, settings) {
			out[itemKey{filepath.Base(report), f[0]}] = true
		}
	}

	return out
}

// newlyFailing returns the open §f items that fail only AFTER the
// candidate lines are added — the would-drift delta of a pre-flight replay.
func newlyFailing(
	reports, modLines []string,
	modDF func(string) int,
	modSquashed []string,
	baseFailing map[itemKey]bool,
	settings Settings,
) [][3]string {
	var newly [][3]string // report, item, reason

	for _, report := range reports {
		for _, f := range CheckReport(report, modLines, modDF, modSquashed, settings) {
			if !baseFailing[itemKey{filepath.Base(report), f[0]}] {
				newly = append(newly, [3]string{filepath.Base(report), f[0], f[1]})
			}
		}
	}

	return newly
}

// printNewly renders the would-drift findings of a pre-flight replay,
// with the optional per-item token table and the per-item culprit
// attribution (which candidate line carries which generic token).
func printNewly(
	newly [][3]string,
	explain bool,
	modLines, modSquashed []string,
	modDF func(string) int,
	candLower, candSquashed []string,
	genericDF int,
) {
	for _, f := range newly {
		gatecli.Outf("WOULD-DRIFT [%s]: %s\n", f[0], gatecli.Truncate(f[1], itemMax))
		gatecli.Outf("                  %s\n", f[2])

		for _, c := range culpritAttributions(f[1], candLower, candSquashed, modDF, genericDF) {
			gatecli.Outf("  culprit: %s\n", c)
		}

		if explain {
			for _, line := range TokenTable(f[1], modLines, modSquashed, modDF, genericDF) {
				gatecli.Outln(line)
			}
		}
	}
}

// bindHit is one open §f item the candidate lines would newly harvest.
type bindHit struct{ report, item string }

// printBindings is the BIND leg of a pre-flight replay (coin
// preflight-bind-check, born from the 2026-10-18 miss): it lists every
// currently-open §f item that STOPS failing once the candidate lines are
// added, each attributed to the candidate line that carries a matched
// token. A row believed to harvest an item but absent from this list
// binds NOTHING — e.g. an item whose token is the slash-joined
// "metric/counter" while the row spells it "metric and counter" — and
// the miss is now visible BEFORE the write, not at the next --audit-all.
// Informational only: generic rows legitimately bind nothing.
func printBindings(
	reports, modLines []string,
	modDF func(string) int,
	modSquashed []string,
	baseFailing map[itemKey]bool,
	candLower, candSquashed []string,
	genericDF int,
) {
	var hits []bindHit

	for _, report := range reports {
		section := ExtractSectionF(report)
		if section == "" {
			continue
		}

		for _, item := range ExtractItems(section) {
			if !baseFailing[itemKey{filepath.Base(report), item}] {
				continue
			}

			if itemVerdict(item, modLines, modSquashed, modDF, genericDF) != "" {
				continue
			}

			hits = append(hits, bindHit{report: filepath.Base(report), item: item})
		}
	}

	if len(hits) == 0 {
		gatecli.Outln("  binds: no open report item is harvested by these candidate lines " +
			"(fine for generic rows; when a row was MEANT to harvest one, run " +
			"--explain-item and reword until it binds)")

		return
	}

	for _, hit := range hits {
		gatecli.Outf("  binds: [%s] %s\n", hit.report, gatecli.Truncate(hit.item, itemMax))

		via := bindViaCandidate(hit.item, candLower, candSquashed)
		if via == "" {
			gatecli.Outln("         (no single candidate line carries a matched token)")
		} else {
			gatecli.Outf("         via candidate: %s\n", gatecli.Truncate(via, reasonMax))
		}
	}
}

// bindViaCandidate returns the first candidate line carrying one of the
// item's tokens, using the same raw-substring-then-squash match the
// harvest verdict itself applies, restricted to the candidate lines.
func bindViaCandidate(item string, candLower, candSquashed []string) string {
	for _, t := range Tokens(item) {
		for _, line := range candLower {
			if strings.Contains(line, t) {
				return line
			}
		}

		key := Squash(t)

		if len(key) < MinTokenLen {
			continue
		}

		for i, line := range candSquashed {
			if strings.Contains(line, key) {
				return candLower[i]
			}
		}
	}

	return ""
}
