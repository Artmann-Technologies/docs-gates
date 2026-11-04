package docsfreshness

import (
	"fmt"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

const (
	// SkipMarker is a PREFIX, not the full tag: both the bare
	// `<!-- fresh:skip -->` and the documented suffixed form
	// `<!-- fresh:skip: reason -->` must exempt an item. The 2026-10-25
	// audit drift burned on exactly this: reports carried suffixed
	// markers (e.g. "already routed in TODO_LIST") while the byte-exact
	// substring match silently ignored them, resurfacing routed items
	// as unharvested.
	SkipMarker = "<!-- fresh:skip"
	// GenericDF is the genericity threshold: a token appearing on >= this many
	// separate lines of TODO_LIST+ROADMAP is "generic" — matched by accident,
	// not by harvest. Tuned 2026-09-22 by replaying all 29 historical §f
	// sections — 8 catches every known-unharvested weak pass while the newest
	// report stays FRESH.
	GenericDF = 8

	// An item passes when at least this many of its tokens match the living
	// docs; items with fewer tokens pass on their single token.
	matchQuorum = 2

	// Display truncations (drift reasons, item text).
	reasonMax = 100
	itemMax   = 120
)

// CheckReport returns the open §f items of one report that the living docs
// do not reflect, as (item, reason) pairs. The settings carry the
// distinctive-token bar.
func CheckReport(
	report string,
	livingLines []string,
	tokenDF func(string) int,
	livingSquashed []string,
	settings Settings,
) [][2]string {
	if livingSquashed == nil {
		livingSquashed = squashLines(livingLines)
	}

	section := ExtractSectionF(report)
	if section == "" {
		return nil
	}

	var findings [][2]string

	for _, item := range ExtractItems(section) {
		if strings.Contains(item, "~~") || strings.Contains(item, SkipMarker) {
			continue
		}

		if reason := itemVerdict(item, livingLines, livingSquashed, tokenDF, settings.df()); reason != "" {
			findings = append(findings, [2]string{item, reason})
		}
	}

	return findings
}

// itemVerdict returns the drift reason for one open item, "" when the living
// docs reflect it: enough tokens matched, at least one of them distinctive.
func itemVerdict(item string, livingLines, livingSquashed []string, tokenDF func(string) int, genericDF int) string {
	toks := Tokens(item)
	if len(toks) == 0 {
		return ""
	}

	need := matchQuorum
	if len(toks) < matchQuorum {
		need = 1
	}

	var found []string

	for _, t := range toks {
		if tokenPresent(t, livingLines, livingSquashed) {
			found = append(found, t)
		}
	}

	if len(found) < need {
		return "missing tokens: " + gatecli.Truncate(strings.Join(missingTokens(toks, found), " "), reasonMax)
	}

	if !hasDistinctive(found, tokenDF, genericDF) {
		return fmt.Sprintf(
			"only generic tokens matched: %s (each appears on >= %d lines — accidental overlap, "+
				"not a harvest)",
			gatecli.Truncate(strings.Join(found, " "), reasonMax),
			genericDF,
		)
	}

	return ""
}

// culpritLimit caps the per-item attribution list (a 7-candidate batch can
// match several generic tokens on several lines — the first entries name
// the reword targets, the rest is noise).
const culpritLimit = 6

// culpritAttributions names the candidate lines that carry a match of one
// of the item's now-generic tokens — TODO preflight-culprit (the
// 2026-10-13 filename chase burned two guess rounds because the
// boundary-token report names TOKENS, never the LINE carrying them, and
// a hyphen/space spelling matches without moving the raw DF count at
// all). Squash-only matches are marked: they are the silent class.
func culpritAttributions(
	item string,
	candLower, candSquashed []string,
	modDF func(string) int,
	genericDF int,
) []string {
	var out []string

	seen := map[string]bool{}

	for _, t := range Tokens(item) {
		if modDF(t) < genericDF {
			continue
		}

		key := Squash(t)

		for i := range candLower {
			raw := strings.Contains(candLower[i], t)
			sq := len(key) >= MinTokenLen && strings.Contains(candSquashed[i], key)

			if !raw && !sq {
				continue
			}

			label := fmt.Sprintf("candidate #%d token %q", i+1, t)

			if sq && !raw {
				label += " (hyphen/space spelling — invisible to the raw DF count)"
			}

			if !seen[label] {
				seen[label] = true
				out = append(out, label)
			}
		}
	}

	if len(out) > culpritLimit {
		out = append(out[:culpritLimit], fmt.Sprintf("… %d more", len(out)-culpritLimit))
	}

	return out
}
