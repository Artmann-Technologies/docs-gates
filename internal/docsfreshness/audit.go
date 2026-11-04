package docsfreshness

import (
	"path/filepath"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

func auditAll(
	root string,
	livingLines, livingSquashed []string,
	tokenDF func(string) int,
	explain bool,
	auditMax int,
	settings Settings,
) int {
	reports := statusReports(root, settings)

	checked, failed := 0, 0

	for _, report := range reports {
		for _, f := range CheckReport(report, livingLines, tokenDF, livingSquashed, settings) {
			failed++

			gatecli.Outf("UNHARVESTED [%s]: %s\n", filepath.Base(report), gatecli.Truncate(f[0], itemMax))
			gatecli.Outf("             %s\n", f[1])

			if explain {
				for _, line := range TokenTable(f[0], livingLines, livingSquashed, tokenDF, settings.df()) {
					gatecli.Outln(line)
				}
			}
		}

		checked++
	}

	gatecli.Outf("docs-freshness --audit-all: %d report(s) replayed, %d unharvested open item(s)\n",
		checked, failed)

	if failed > 0 && auditMax >= 0 && failed <= auditMax {
		gatecli.Outf(
			"docs-freshness --audit-all: RATCHET — %d unharvested tolerated (pinned --audit-max=%d); red only past %d\n",
			failed,
			auditMax,
			auditMax,
		)
		gatecli.Outln("drain them with the docs-health harvest ritual, then lower the pin")

		return gatecli.ExitClean
	}

	if failed > 0 {
		if auditMax >= 0 {
			gatecli.Errf(
				"docs-freshness --audit-all: REGRESSION — %d unharvested past the pinned --audit-max=%d. ",
				failed,
				auditMax,
			)
			gatecli.Errln("Harvest them (docs-health) or annotate them resolved, then re-pin.")
		} else {
			gatecli.Errln("docs-freshness --audit-all: DRIFT — historical open items are " +
				"not represented in TODO_LIST/ROADMAP/CHANGELOG. Harvest them " +
				"(docs-health) or annotate them resolved.")
		}

		return gatecli.ExitFindings
	}

	gatecli.Outln("docs-freshness --audit-all: FRESH")

	return gatecli.ExitClean
}

func checkNewest(
	root string,
	livingLines, livingSquashed []string,
	tokenDF func(string) int,
	explain bool,
	settings Settings,
) int {
	report := NewestReport(root, settings)
	warnHtmlBlindSpot(FindHtmlBlindSpot(root, settings))

	if report == "" {
		gatecli.Outln("docs-freshness: no docs/status reports found — nothing to check")

		return gatecli.ExitClean
	}

	if ExtractSectionF(report) == "" {
		gatecli.Outf("docs-freshness: %s has no §f section — nothing to check\n", filepath.Base(report))

		return gatecli.ExitClean
	}

	findings := CheckReport(report, livingLines, tokenDF, livingSquashed, settings)
	for _, f := range findings {
		gatecli.Outf("UNHARVESTED: %s\n", gatecli.Truncate(f[0], itemMax))
		gatecli.Outf("             %s\n", f[1])

		for _, line := range harvestAttributions(f[0], livingLines, livingSquashed, tokenDF, settings.df()) {
			gatecli.Outln(line)
		}

		if explain {
			for _, line := range TokenTable(f[0], livingLines, livingSquashed, tokenDF, settings.df()) {
				gatecli.Outln(line)
			}
		}
	}

	gatecli.Outf("docs-freshness: %s: %d unharvested open §f item(s)\n", filepath.Base(report), len(findings))

	if len(findings) > 0 {
		gatecli.Errf("docs-freshness: DRIFT — %d item(s) not reflected in "+
			"TODO_LIST.md/ROADMAP.md. Harvest them (docs-health) or mark "+
			"resolved; exempt sibling-scoped items with %s -->.\n", len(findings), SkipMarker)

		return gatecli.ExitFindings
	}

	gatecli.Outln("docs-freshness: FRESH")

	return gatecli.ExitClean
}
