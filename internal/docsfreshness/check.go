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
func CheckReport(report string, living Corpus, settings Settings) [][2]string {
	section := ExtractSectionF(report)
	if section == "" {
		return nil
	}

	var findings [][2]string

	for _, item := range ExtractItems(section) {
		if strings.Contains(item, "~~") || strings.Contains(item, SkipMarker) {
			continue
		}

		if reason := itemVerdict(item, living, settings.df()); reason != "" {
			findings = append(findings, [2]string{item, reason})
		}
	}

	return findings
}

// itemVerdict returns the drift reason for one open item, "" when the living
// docs reflect it: enough tokens matched, at least one of them distinctive.
func itemVerdict(item string, living Corpus, genericDF int) string {
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
		if living.Contains(t) {
			found = append(found, t)
		}
	}

	if len(found) < need {
		return "missing tokens: " + gatecli.Truncate(strings.Join(missingTokens(toks, found), " "), reasonMax)
	}

	if !hasDistinctive(found, living.Count, genericDF) {
		return fmt.Sprintf(
			"only generic tokens matched: %s (each appears on >= %d lines — accidental overlap, "+
				"not a harvest)",
			gatecli.Truncate(strings.Join(found, " "), reasonMax),
			genericDF,
		)
	}

	return ""
}

// attributionExamples is how many example living lines a matched token
// shows in the harvest attributions (two suffice to see the pattern).
const attributionExamples = 2

// culpritLimit caps the per-item attribution list (a 7-candidate batch can
// match several generic tokens on several lines — the first entries name
// the reword targets, the rest is noise).
const culpritLimit = 6

// harvestAttributions names, for a FAILING item, each matched token with
// its living-docs DF and up to two example lines — the preflight's
// culprit attribution (preflight-culprit) extended to the harvest
// findings themselves, so a red item shows WHERE its anchors live
// instead of naming bare tokens. Bounded by culpritLimit.
func harvestAttributions(
	item string,
	living Corpus,
	genericDF int,
) []string {
	var out []string

	seen := map[string]bool{}

	for _, t := range Tokens(item) {
		if seen[t] || !living.Contains(t) {
			continue
		}

		seen[t] = true

		d := living.Count(t)

		label := fmt.Sprintf("  matched: %s (DF %d", t, d)
		if d >= genericDF {
			label += ", generic"
		}

		label += ")"

		out = append(out, label)

		shown := 0

		for _, line := range living.lines {
			if !strings.Contains(line, t) {
				continue
			}

			out = append(out, "    line: "+gatecli.Truncate(line, reasonMax))

			shown++

			if shown == attributionExamples {
				break
			}
		}

		if len(out) >= culpritLimit {
			break
		}
	}

	return out
}

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
