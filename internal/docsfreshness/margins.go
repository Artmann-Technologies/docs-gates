package docsfreshness

import (
	"sort"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// marginNearBar is how close to the generic bar a token's document
// frequency must sit to make the proactive margins report: a token this
// close is one or two careless lines away from genericizing every item
// whose only distinctive token it is.
const marginNearBar = 2

// marginToken is one near-bar token with its current living-docs DF.
type marginToken struct {
	tok string
	df  int
}

// TokenMargins prints the proactive token-margin report (--margins):
// every living-docs token within marginNearBar DF of the generic bar,
// so living-doc drafts avoid those words FIRST-TRY instead of
// discovering the boundary through a failed gate run. Informational
// only — the report never flips a verdict.
func TokenMargins(root string, settings Settings) int {
	lines, _, tokenDF := loadLivingDocs(root, settings)

	bar := settings.df()

	seen := map[string]bool{}

	var nears []marginToken

	for _, line := range lines {
		for _, tok := range Tokens(line) {
			if seen[tok] {
				continue
			}

			seen[tok] = true

			d := tokenDF(tok)
			if d >= bar-marginNearBar && d < bar {
				nears = append(nears, marginToken{tok, d})
			}
		}
	}

	sort.Slice(nears, func(i, j int) bool {
		if nears[i].df != nears[j].df {
			return nears[i].df > nears[j].df
		}

		return nears[i].tok < nears[j].tok
	})

	gatecli.Outf(
		"docs-freshness --margins: %d token(s) within %d DF of the generic bar (%d) — a new line mentioning them costs budget:\n",
		len(nears),
		marginNearBar,
		bar,
	)

	for _, m := range nears {
		gatecli.Outf("  %-34s DF %d — %d more line(s) genericize it\n",
			m.tok, m.df, bar-m.df)
	}

	return gatecli.ExitClean
}
