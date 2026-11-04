package docsfreshness

import (
	"fmt"
	"os"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// DumpItems prints the §f extraction decisions for one report
// (--dump-items <report>): whether a §f section was found, which lines
// ExtractItems accepted as open items, the tokens each item carries, and
// which items the strike marker exempts — the extraction decision is
// fully explainable, including the "why is this ~~row~~ not harvested"
// side. Informational only, exit 0 (a missing report path is usage).
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

	var open, struck []string

	for _, item := range ExtractItems(section) {
		if strings.Contains(item, "~~") {
			struck = append(struck, item)

			continue
		}

		open = append(open, item)
	}

	gatecli.Outf("docs-freshness --dump-items: %s — %d open item(s) extracted\n", report, len(open))

	for i, item := range open {
		gatecli.Outf("  item %d: %s\n", i+1, gatecli.Truncate(item, itemMax))
		gatecli.Outf("    tokens: %s\n", fmt.Sprint(Tokens(item)))
	}

	for i, item := range struck {
		gatecli.Outf("  struck %d (exempt: the ~~ strike marker carries the resolution): %s\n",
			i+1, gatecli.Truncate(item, itemMax))
	}

	return gatecli.ExitClean
}
