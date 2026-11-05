package docsfreshness

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// runExplainItem validates the needle positional and delegates to the
// diagnostic ExplainItem mode.
func runExplainItem(root string, positional []string, settings Settings) int {
	if len(positional) < 2 || strings.TrimSpace(positional[1]) == "" {
		gatecli.Errln("FATAL: --explain-item needs a substring as the second positional —")
		gatecli.Errln("  e.g. docs-freshness . --explain-item rewrite — matched against every")
		gatecli.Errln("  open §f item of every report (docs/status + archived), harvested or not")

		return gatecli.ExitUsage
	}

	living := loadLivingDocs(root, settings)

	return ExplainItem(root, positional[1], living, settings)
}

// HtmlBlindSpot describes the docs/status HTML report snapshots this gate
// cannot parse (the 2026-10-07 self-review finding: the gate called a
// 13-10 .md "newest" while the newer 16-25 .html sat beside it — every
// verdict, §f item and routing obligation written into an HTML report is
// invisible to freshness/annotate/routing alike).
type HtmlBlindSpot struct {
	Count    int
	Newest   string // base name, "" when none
	Outranks bool   // newest .html is lexically newer than the newest .md
}

// FindHtmlBlindSpot scans docs/status for *.html report snapshots and
// records whether the newest one outranks the newest parsed .md (both
// filename families share the YYYY-MM-DD_HH-MM prefix, so lexical order
// is chronological).
func FindHtmlBlindSpot(root string, settings Settings) HtmlBlindSpot {
	htmls, _ := filepath.Glob(
		filepath.Join(root, settings.dir(), "2*.html"),
	) //nolint:erraudit // constant pattern cannot be malformed
	sort.Strings(htmls)

	spot := HtmlBlindSpot{Count: len(htmls)}
	if len(htmls) == 0 {
		return spot
	}

	spot.Newest = filepath.Base(htmls[len(htmls)-1])

	md := NewestReport(root, settings)
	mdBase := strings.TrimSuffix(filepath.Base(md), ".md")
	htmlBase := strings.TrimSuffix(spot.Newest, ".html")
	spot.Outranks = md == "" || htmlBase > mdBase

	return spot
}

// warnHtmlBlindSpot prints the count+WARN the 2026-10-07 review asked for:
// the gate stays .md-parsed (parsing rendered HTML would couple it to the
// report kit's markup), but it must never again silently SKIP a newer
// HTML report as if it did not exist.
func warnHtmlBlindSpot(spot HtmlBlindSpot) {
	if spot.Count == 0 {
		return
	}

	scope := "all older than the newest parsed .md"
	if spot.Outranks {
		scope = "NEWER than the newest parsed .md — its items are INVISIBLE to this gate"
	}

	gatecli.Errf("WARN: %d HTML report(s) in docs/status are outside the .md scan (%s); "+
		"newest: %s — route any obligations they carry into the living docs in the same pass\n",
		spot.Count, scope, spot.Newest)
}

// TokenTable builds the per-token DF table for one item: state, raw-line
// document frequency, and the distinctive/generic verdict — the
// boundary-token flip diagnosed in one command. The 2026-10-07 flip (a
// fresh CHANGELOG line's "decision batch" pushed the `batch` token from
// DF=7 to 8, silently genericizing a 2026-10-03 item's last distinctive
// token) needed a hand-rolled grep probe to root-cause; `--explain-token`
// prints these lines under every UNHARVESTED item.
func TokenTable(item string, living Corpus, genericDF int) []string {
	toks := Tokens(item)

	out := make([]string, 0, 1+len(toks))
	out = append(out, fmt.Sprintf("    token table (%d tokens):", len(toks)))

	for _, t := range toks {
		state := "missing"
		if living.Contains(t) {
			state = "matched"
		}

		kind := "generic"
		if living.Count(t) < genericDF {
			kind = "DISTINCTIVE"
		}

		out = append(out, fmt.Sprintf("      %-28s DF=%-4d %-8s %s", t, living.Count(t), state, kind))
	}

	return out
}

// ExplainItem prints the token table for EVERY open §f item containing the
// needle (case-insensitive substring), across the full report corpus,
// regardless of harvest state — the proactive-margin companion to
// --explain-token (whose tables ride UNHARVESTED findings only, so a FRESH
// corpus prints none). Always exits clean: it is a diagnostic, not a gate.
func ExplainItem(
	root, needle string,
	living Corpus,
	settings Settings,
) int {
	lowerNeedle := strings.ToLower(needle)
	matches := 0

	for _, report := range statusReports(root, settings) {
		section := ExtractSectionF(report)
		if section == "" {
			continue
		}

		for _, item := range ExtractItems(section) {
			if strings.Contains(item, "~~") || strings.Contains(item, SkipMarker) {
				continue // resolved or exempt items carry no harvest margin
			}

			if !strings.Contains(strings.ToLower(item), lowerNeedle) {
				continue
			}

			matches++

			verdict := "harvested"
			if reason := itemVerdict(item, living, settings.df()); reason != "" {
				verdict = "UNHARVESTED (" + reason + ")"
			}

			gatecli.Outf("--explain-item [%s]: %s\n", filepath.Base(report), gatecli.Truncate(item, itemMax))
			gatecli.Outf("             verdict: %s\n", verdict)

			for _, line := range TokenTable(item, living, settings.df()) {
				gatecli.Outln(line)
			}
		}
	}

	if matches == 0 {
		gatecli.Outf("--explain-item: no open §f item in any report matches %q\n", needle)
	}

	return gatecli.ExitClean
}
