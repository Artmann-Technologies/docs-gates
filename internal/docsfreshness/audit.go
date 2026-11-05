package docsfreshness

import (
	"path/filepath"
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// estPerReportLine anchors the pre-replay cost estimate to the worst
// REAL corpus measured (CV, 2026-11-05: 1,640 reports × ~2,000 living
// lines replayed in 49s, findings-heavy — harvest attributions for
// 8,788 unharvested items dominate that run). A findings-free corpus
// of the same shape replays ~700x faster (the committed benchmark:
// ~50ms), so the estimate deliberately overstates — a size warning
// that fires early beats one that fires late. This is a visibility
// line, not a performance contract; TestReplaySpeedGuard holds the
// real one.
const estPerReportLine = 15 * time.Microsecond

// auditAllEstimate scales the worst-case replay cost linearly with the
// reports × living-lines product.
func auditAllEstimate(reports, livingLines int) time.Duration {
	return time.Duration(reports*max(livingLines, 1)) * estPerReportLine
}

func auditAll(
	root string,
	living Corpus,
	explain bool,
	auditMax int,
	settings Settings,
) int {
	reports := statusReports(root, settings)

	gatecli.Outf(
		"docs-freshness --audit-all: replaying %d report(s) against %d living line(s) — est ~%s (worst case, findings-heavy)\n",
		len(reports),
		len(living.lines),
		gatecli.FormatDuration(auditAllEstimate(len(reports), len(living.lines))),
	)

	checked, failed := 0, 0

	for _, report := range reports {
		for _, f := range CheckReport(report, living, settings) {
			failed++

			gatecli.Outf("UNHARVESTED [%s]: %s\n", filepath.Base(report), gatecli.Truncate(f[0], itemMax))
			gatecli.Outf("             %s\n", f[1])

			for _, line := range harvestAttributions(f[0], living, settings.df()) {
				gatecli.Outln(line)
			}

			if explain {
				for _, line := range TokenTable(f[0], living, settings.df()) {
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
	living Corpus,
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

	findings := CheckReport(report, living, settings)
	for _, f := range findings {
		gatecli.Outf("UNHARVESTED: %s\n", gatecli.Truncate(f[0], itemMax))
		gatecli.Outf("             %s\n", f[1])

		for _, line := range harvestAttributions(f[0], living, settings.df()) {
			gatecli.Outln(line)
		}

		if explain {
			for _, line := range TokenTable(f[0], living, settings.df()) {
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
