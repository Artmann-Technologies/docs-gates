package docsfreshness

import (
	"fmt"
	"os"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// DumpItems prints the §f extraction decisions for one report
// (--dump-items <report>): whether a §f section was found, which lines
// ExtractItems accepted as open items, and the tokens each item carries.
// The point is inspectability WITHOUT test scaffolding — "why is this
// line not an item / why does this item not harvest" becomes one
// command. Informational only, exit 0 (a missing report path is usage).
func DumpItems(report string) int {
	if _, err := os.ReadFile(report); err != nil {
		gatecli.Errf("FATAL: reading %s: %v\n", report, err)

		return gatecli.ExitUsage
	}

	section := ExtractSectionF(report)
	if section == "" {
		gatecli.Errf(
			"FATAL: no §f section found in %s (the freshness gate reads the newest report's §f only)\n",
			report,
		)

		return gatecli.ExitUsage
	}

	items := ExtractItems(section)

	gatecli.Outf("docs-freshness --dump-items: %s — %d open item(s) extracted\n", report, len(items))

	for i, item := range items {
		gatecli.Outf("  item %d: %s\n", i+1, gatecli.Truncate(item, itemMax))
		gatecli.Outf("    tokens: %s\n", fmt.Sprint(Tokens(item)))
	}

	return gatecli.ExitClean
}
