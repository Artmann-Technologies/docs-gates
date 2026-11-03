// Package docsfreshness is the docs-freshness gate (docs-health loop
// closure, 2026-09-21).
//
// Asserts that the newest `docs/status/*.md` report's forward-looking §f
// items are reflected in the living docs (TODO_LIST.md, ROADMAP.md,
// CHANGELOG.md) — the harvest step that hand-maintenance always forgets
// becomes a mechanical check.
//
// Rules:
//   - The newest report is the lexically last `docs/status/2*.md` file
//     (filenames start with `YYYY-MM-DD_HH-MM`); `archived/` is ignored.
//   - Living docs are TODO_LIST.md + ROADMAP.md + CHANGELOG.md: completed
//     items are DELETED from the TODO list by convention and live in the
//     CHANGELOG only (discovered live 2026-09-23: harvesting a just-done
//     item flipped the gate to DRIFT until the CHANGELOG joined).
//   - Only OPEN items count: lines carrying a strikethrough marker (`~~`)
//     are resolved and skipped. Add `<!-- fresh:skip -->` (or the suffixed
//     `<!-- fresh:skip: reason -->`) to exempt an item
//     (e.g. sibling-repo-scoped work tracked in the telephony TODO_LIST).
//   - An item passes when at least 2 of its distinctive tokens (>= 4 chars,
//     minus stopwords) appear in the living docs; items with fewer
//     than 2 tokens pass on their single token.
//   - Genericity threshold (adversarial audit 2026-09-22): at least one
//     MATCHED token must be distinctive — present on fewer than
//     GenericDF separate lines of the living docs. Tokens like
//     "backup"/"switch"/"webphone" ride dozens of lines, so an unharvested
//     item can match two of them by accident. A fail on this clause alone
//     is reported as "only generic tokens".
//   - Hyphen/space-insensitive matching (TODO 8.4, 2026-09-25): a report's
//     `real-cax` must match a living-docs line saying "real CAX". Tokens
//     are matched raw FIRST, then with hyphens/spaces collapsed — the
//     GENERICITY count stays on RAW lines (the basis the 2026-09-22 tuning
//     was measured against).
//
// Usage: docs-freshness [--audit-all] [repo-root]   (default: cwd)
//
// The report corpus, the living-docs list and the distinctive-token bar
// come from the repo's .docs-gates.json when present (internal/config
// defaults otherwise — the historical constants).
//
// Exit 0 = fresh, 1 = drift (unharvested items listed), 2 = usage error.
package docsfreshness

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

//nolint:gochecknoglobals // static lookup table; rebuilding it per token would be the real waste
var stopwords = map[string]bool{
	"the": true, "and": true, "for": true, "with": true, "from": true,
	"that": true, "this": true, "then": true, "when": true, "into": true,
	"after": true, "before": true, "once": true, "over": true, "under": true,
	"must": true, "they": true, "them": true, "their": true,
	"should": true, "would": true, "could": true, "also": true,
	"more": true, "most": true, "than": true, "very": true,
	"items": true, "item": true, "task": true, "tasks": true,
	"open": true, "still": true, "pending": true, "until": true,
	"about": true, "again": true, "each": true, "every": true,
	"some": true, "what": true, "which": true, "where": true,
	"while": true, "yours": true, "ours": true, "have": true,
	"been": true, "being": true, "does": true, "doing": true,
	"done": true, "todo": true, "list": true, "roadmap": true,
	"report": true, "reports": true, "session": true, "sessions": true,
	"next": true, "things": true, "work": true, "owner": true,
	"user": true, "owners": true, "users": true, "here": true,
	"there": true, "only": true, "just": true, "like": true,
	"make": true, "made": true, "take": true, "taken": true,
	"keep": true, "kept": true, "both": true, "same": true,
	"other": true, "another": true, "such": true, "much": true,
	"many": true, "less": true, "least": true, "first": true,
	"last": true, "second": true, "third": true, "half": true,
	"part": true, "full": true, "wide": true, "long": true,
	"short": true, "small": true, "large": true, "real": true,
	"really": true,
}

const (
	MinTokenLen = 4
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

var (
	sectionFRe = regexp.MustCompile(`(?i)^#{2,}\s*f\)`)
	headingRe  = regexp.MustCompile(`^#{1,2}\s`)
	proseRe    = regexp.MustCompile(`^\s*(\d+)[.)]\s+(.*)$`)
	rowRe      = regexp.MustCompile(`^\|\s*(~~\d+~~|\d+)\s*\|`)
	sepRowRe   = regexp.MustCompile(`^\|[-: ]+\|?\s*$`)
	tokenRe    = regexp.MustCompile(`[a-z0-9][a-z0-9._/-]*`)
	hyphenRe   = regexp.MustCompile(`[-\s]+`)
)

func NewestReport(root string, s Settings) string {
	entries, err := filepath.Glob(s.reportGlobPath(root))
	if err != nil || len(entries) == 0 {
		return ""
	}

	sort.Strings(entries)

	for _, entry := range slices.Backward(entries) {
		if info, err := os.Stat(entry); err == nil && info.Mode().IsRegular() {
			return entry
		}
	}

	return ""
}

func ExtractSectionF(report string) string {
	data, err := os.ReadFile(report)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	start := -1

	for i, line := range lines {
		if sectionFRe.MatchString(line) {
			start = i + 1

			break
		}
	}

	if start < 0 {
		return ""
	}

	var section []string

	for _, line := range lines[start:] {
		if headingRe.MatchString(line) {
			break
		}

		section = append(section, line)
	}

	return strings.Join(section, "\n")
}

func ExtractItems(section string) []string {
	var (
		items   []string
		current []string
	)

	flush := func() {
		if len(current) > 0 {
			text := strings.TrimSpace(strings.Join(current, " "))
			if text != "" {
				items = append(items, text)
			}

			current = nil
		}
	}

	for line := range strings.SplitSeq(section, "\n") {
		switch {
		case proseRe.MatchString(line):
			flush()

			current = append(current, strings.TrimSpace(line))
		case rowRe.MatchString(line):
			flush()
			// Keep the whole row: a struck number cell (~~5~~) must leave
			// its ~~ marker in the item text so it is skipped as resolved.
			current = append(current, strings.TrimSpace(line))
		case sepRowRe.MatchString(line) || strings.HasPrefix(line, "| #"):
			continue
		case strings.TrimSpace(line) != "":
			if len(current) > 0 {
				current = append(current, strings.TrimSpace(line))
			}
		default:
			flush()
		}
	}

	flush()

	return items
}

// Tokens: lowercased raw tokens minus stopwords, trailing punctuation
// stripped, minimum length enforced. A filename's INTERNAL dot survives
// ("backup-restore.md" stays one token).
func Tokens(text string) []string {
	raw := tokenRe.FindAllString(strings.ToLower(text), -1)

	var out []string

	for _, t := range raw {
		t = strings.TrimRight(t, ".,;:!?)](\"'")
		if len(t) >= MinTokenLen && !stopwords[t] {
			out = append(out, t)
		}
	}

	return out
}

// Squash is the hyphen/space-collapsed, lowercased comparison key (TODO 8.4).
func Squash(text string) string {
	return strings.ToLower(hyphenRe.ReplaceAllString(text, ""))
}

// CheckReport returns the open §f items of one report that the living docs
// do not reflect, as (item, reason) pairs. The settings carry the
// distinctive-token bar.
func CheckReport(
	report string,
	livingLines []string,
	tokenDF func(string) int,
	livingSquashed []string,
	s Settings,
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

		if reason := itemVerdict(item, livingLines, livingSquashed, tokenDF, s.df()); reason != "" {
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

// tokenPresent: raw substring match first, then the hyphen/space-collapsed
// key (TODO 8.4) when the token is long enough to carry it.
func tokenPresent(token string, livingLines, livingSquashed []string) bool {
	for _, line := range livingLines {
		if strings.Contains(line, token) {
			return true
		}
	}

	key := Squash(token)
	if len(key) >= MinTokenLen {
		for _, line := range livingSquashed {
			if strings.Contains(line, key) {
				return true
			}
		}
	}

	return false
}

func missingTokens(toks, found []string) []string {
	var missing []string

	for _, t := range toks {
		if !slices.Contains(found, t) {
			missing = append(missing, t)
		}
	}

	return missing
}

func hasDistinctive(found []string, tokenDF func(string) int, genericDF int) bool {
	for _, t := range found {
		if tokenDF(t) < genericDF {
			return true
		}
	}

	return false
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

// squashLines is the batch form of Squash used for the living-docs corpus.
func squashLines(lines []string) []string {
	squashed := make([]string, 0, len(lines))

	for _, l := range lines {
		squashed = append(squashed, Squash(l))
	}

	return squashed
}

// parseAuditMax splits --audit-max=N off the argument list: the parsed pin
// (-1 = strict), the remaining args, and a nonzero exit code on a malformed
// value (0 = ok).
func parseAuditMax(args []string) (int, []string, int) {
	auditMax := -1

	var rest []string

	for _, a := range args {
		if v, ok := strings.CutPrefix(a, "--audit-max="); ok {
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				gatecli.Errf("FATAL: --audit-max wants a non-negative integer, got %q\n", v)

				return -1, nil, gatecli.ExitUsage
			}

			auditMax = n

			continue
		}

		rest = append(rest, a)
	}

	return auditMax, rest, 0
}

// Run is the CLI entry (repo-root positional + --audit-all flag semantics).
func Run(args []string) int {
	// The audit ratchet (TODO §91 coin audit-ratchet): --audit-max=N pins
	// the tolerated unharvested count so the audit goes red only on a
	// REGRESSION past N — draining a legacy backlog in one sitting was
	// never this gate's job (it was permanently red 2026-10-08..23 while
	// the backlog sat at 78). Valued form --audit-max=N; -1 = strict.
	auditMax, rest, badExit := parseAuditMax(args)
	if badExit != 0 {
		return badExit
	}

	flags, positional, unknown := gatecli.ParseFlags(
		rest, "--audit-all", "--explain-token", "--pre-flight", "--duplicates", "--explain-item",
	)
	if len(unknown) > 0 {
		gatecli.Errf("FATAL: %v\n", gatecli.NewUnknownFlagsError(unknown))

		return gatecli.ExitUsage
	}

	root, err := gatecli.ResolveRoot(positional)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	if flags["--duplicates"] && !flags["--pre-flight"] {
		gatecli.Errln("FATAL: --duplicates only applies to --pre-flight mode — it declares that an")
		gatecli.Errln("  on-disk candidate line is about to land as a SECOND copy (ledger-block class)")
		gatecli.Errln("  so the +1 token-DF push of that copy is modeled instead of refused")

		return gatecli.ExitUsage
	}

	if flags["--explain-item"] {
		// A FRESH corpus prints no --explain-token tables (they ride
		// UNHARVESTED findings only); this mode answers the margin
		// question for ANY item — including harvested ones sitting at
		// DF just below the bar. The needle is the SECOND positional:
		// `docs-freshness . --explain-item rewrite`.
		return runExplainItem(root, positional)
	}

	if flags["--pre-flight"] {
		candidate, err := readCandidates()
		if err != nil {
			gatecli.Errf("FATAL: %v\n", err)

			return gatecli.ExitUsage
		}

		return PreFlight(root, candidate, flags["--explain-token"], flags["--duplicates"])
	}

	livingLines, livingSquashed, tokenDF := loadLivingDocs(root)

	if flags["--audit-all"] {
		return auditAll(root, livingLines, livingSquashed, tokenDF, flags["--explain-token"], auditMax)
	}

	return checkNewest(root, livingLines, livingSquashed, tokenDF, flags["--explain-token"])
}

// candidatesError is the package's static error type: readCandidates
// reports stdin failures as *candidatesError with the cause reachable
// through Unwrap.
type candidatesError struct {
	msg string
	err error
}

func (e *candidatesError) Error() string { return e.msg + ": " + e.err.Error() }

func (e *candidatesError) Unwrap() error { return e.err }

// readCandidates reads the --pre-flight candidate lines from stdin,
// dropping blank lines.
func readCandidates() ([]string, *candidatesError) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, &candidatesError{msg: "reading candidate lines from stdin", err: err}
	}

	var candidate []string

	for line := range strings.SplitSeq(string(data), "\n") {
		if strings.TrimSpace(line) != "" {
			candidate = append(candidate, line)
		}
	}

	return candidate, nil
}

// runExplainItem validates the needle positional and delegates to the
// diagnostic ExplainItem mode.
func runExplainItem(root string, positional []string) int {
	if len(positional) < 2 || strings.TrimSpace(positional[1]) == "" {
		gatecli.Errln("FATAL: --explain-item needs a substring as the second positional —")
		gatecli.Errln("  e.g. docs-freshness . --explain-item rewrite — matched against every")
		gatecli.Errln("  open §f item of every report (docs/status + archived), harvested or not")

		return gatecli.ExitUsage
	}

	livingLines, livingSquashed, tokenDF := loadLivingDocs(root)

	return ExplainItem(root, positional[1], livingLines, livingSquashed, tokenDF)
}

// loadLivingDocs reads TODO_LIST + ROADMAP + CHANGELOG into the comparison
// corpus: lowercased raw lines, their squashed keys, and the raw-line token
// document frequency.
func loadLivingDocs(root string) ([]string, []string, func(string) int) {
	var livingLines []string

	for _, name := range []string{"TODO_LIST.md", "ROADMAP.md", "CHANGELOG.md"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			continue //nolint:erraudit // a not-yet-existing living doc simply contributes no lines
		}

		for line := range strings.SplitSeq(string(data), "\n") {
			livingLines = append(livingLines, strings.ToLower(line))
		}
	}

	return livingLines, squashLines(livingLines), tokenDFOver(livingLines)
}

// tokenDFOver is the raw-line document-frequency counter: how many
// living-doc lines contain the token as a substring. Genericity counts
// stay on RAW lines: GenericDF = 8 was tuned 2026-09-22 against raw
// substring counts, and squashed counting inflates them (words join
// across boundaries) — measured as a false DRIFT on a resolved item the
// same day this landed.
func tokenDFOver(lines []string) func(string) int {
	return func(token string) int {
		n := 0

		for _, line := range lines {
			if strings.Contains(line, token) {
				n++
			}
		}

		return n
	}
}

// statusReports lists every replayable report: docs/status/2*.md plus
// docs/status/archived/2*.md, oldest first — the corpus --audit-all and
// --pre-flight replay.
func statusReports(root string) []string {
	reports, _ := filepath.Glob(
		filepath.Join(root, "docs", "status", "2*.md"),
	) //nolint:erraudit // constant pattern cannot be malformed
	sort.Strings(reports)

	archived, _ := filepath.Glob(
		filepath.Join(root, "docs", "status", "archived", "2*.md"),
	) //nolint:erraudit // constant pattern cannot be malformed
	sort.Strings(archived)

	return append(reports, archived...)
}

func auditAll(
	root string,
	livingLines, livingSquashed []string,
	tokenDF func(string) int,
	explain bool,
	auditMax int,
) int {
	reports := statusReports(root)

	checked, failed := 0, 0

	for _, report := range reports {
		for _, f := range CheckReport(report, livingLines, tokenDF, livingSquashed) {
			failed++

			gatecli.Outf("UNHARVESTED [%s]: %s\n", filepath.Base(report), gatecli.Truncate(f[0], itemMax))
			gatecli.Outf("             %s\n", f[1])

			if explain {
				for _, line := range TokenTable(f[0], livingLines, livingSquashed, tokenDF) {
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
func FindHtmlBlindSpot(root string) HtmlBlindSpot {
	htmls, _ := filepath.Glob(
		filepath.Join(root, "docs", "status", "2*.html"),
	) //nolint:erraudit // constant pattern cannot be malformed
	sort.Strings(htmls)

	spot := HtmlBlindSpot{Count: len(htmls)}
	if len(htmls) == 0 {
		return spot
	}

	spot.Newest = filepath.Base(htmls[len(htmls)-1])

	md := NewestReport(root)
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
func TokenTable(item string, livingLines, livingSquashed []string, tokenDF func(string) int) []string {
	toks := Tokens(item)

	out := make([]string, 0, 1+len(toks))
	out = append(out, fmt.Sprintf("    token table (%d tokens):", len(toks)))

	for _, t := range toks {
		state := "missing"
		if tokenPresent(t, livingLines, livingSquashed) {
			state = "matched"
		}

		kind := "generic"
		if tokenDF(t) < GenericDF {
			kind = "DISTINCTIVE"
		}

		out = append(out, fmt.Sprintf("      %-28s DF=%-4d %-8s %s", t, tokenDF(t), state, kind))
	}

	return out
}

// ExplainItem prints the token table for EVERY open §f item containing the
// needle (case-insensitive substring), across the full report corpus,
// regardless of harvest state — the proactive-margin companion to
// --explain-token (whose tables ride UNHARVESTED findings only, so a FRESH
// corpus prints none). Always exits clean: it is a diagnostic, not a gate.
func ExplainItem(root, needle string, livingLines, livingSquashed []string, tokenDF func(string) int) int {
	lowerNeedle := strings.ToLower(needle)
	matches := 0

	for _, report := range statusReports(root) {
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
			if reason := itemVerdict(item, livingLines, livingSquashed, tokenDF); reason != "" {
				verdict = "UNHARVESTED (" + reason + ")"
			}

			gatecli.Outf("--explain-item [%s]: %s\n", filepath.Base(report), gatecli.Truncate(item, itemMax))
			gatecli.Outf("             verdict: %s\n", verdict)

			for _, line := range TokenTable(item, livingLines, livingSquashed, tokenDF) {
				gatecli.Outln(line)
			}
		}
	}

	if matches == 0 {
		gatecli.Outf("--explain-item: no open §f item in any report matches %q\n", needle)
	}

	return gatecli.ExitClean
}

func checkNewest(root string, livingLines, livingSquashed []string, tokenDF func(string) int, explain bool) int {
	report := NewestReport(root)
	warnHtmlBlindSpot(FindHtmlBlindSpot(root))

	if report == "" {
		gatecli.Outln("docs-freshness: no docs/status reports found — nothing to check")

		return gatecli.ExitClean
	}

	if ExtractSectionF(report) == "" {
		gatecli.Outf("docs-freshness: %s has no §f section — nothing to check\n", filepath.Base(report))

		return gatecli.ExitClean
	}

	findings := CheckReport(report, livingLines, tokenDF, livingSquashed)
	for _, f := range findings {
		gatecli.Outf("UNHARVESTED: %s\n", gatecli.Truncate(f[0], itemMax))
		gatecli.Outf("             %s\n", f[1])

		if explain {
			for _, line := range TokenTable(f[0], livingLines, livingSquashed, tokenDF) {
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
func BoundaryCrossings(candidate, baseLines, modLines []string) []Crossing {
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
			if before < GenericDF && after >= GenericDF {
				crossings = append(crossings, Crossing{Token: tok, Before: before, After: after})
			}
		}
	}

	return crossings
}

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
func PreFlight(root string, candidate []string, explain, allowDuplicates bool) int {
	if len(candidate) == 0 {
		gatecli.Errln("FATAL: --pre-flight needs the candidate line(s) on stdin — " +
			"the exact new lines about to land in TODO_LIST/ROADMAP/CHANGELOG")

		return gatecli.ExitUsage
	}

	baseLines, baseSquashed, baseDF := loadLivingDocs(root)

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

	reports := statusReports(root)
	baseFailing := failingKeys(reports, baseLines, baseDF, baseSquashed)
	newly := newlyFailing(reports, modLines, modDF, modSquashed, baseFailing)

	gatecli.Outf("docs-freshness --pre-flight: %d candidate line(s), replayed over %d report(s)\n",
		len(candidate), len(reports))

	for _, c := range BoundaryCrossings(candidate, baseLines, modLines) {
		gatecli.Outf("  boundary token: %-28s DF %d→%d (genericizes every item whose only "+
			"distinctive token this was)\n", c.Token, c.Before, c.After)
	}

	printNewly(newly, explain, modLines, modSquashed, modDF, candLower, candSquashed)
	printBindings(reports, modLines, modDF, modSquashed, baseFailing, candLower, candSquashed)

	if len(newly) > 0 {
		gatecli.Errf("docs-freshness --pre-flight: UNSAFE — %d historical item(s) would flip to "+
			"DRIFT; reword the candidate line(s) before landing them\n", len(newly))

		return gatecli.ExitFindings
	}

	gatecli.Outln("docs-freshness --pre-flight: SAFE — no historical item flips to DRIFT")

	return gatecli.ExitClean
}

// bindHit is one open §f item the candidate lines would newly harvest.
type bindHit struct{ report, item string }

// printBindings is the BIND leg of a pre-flight replay (TODO §77 coin
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

			if itemVerdict(item, modLines, modSquashed, modDF) != "" {
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

// itemKey identifies one open §f item by its report basename and raw text.
type itemKey struct{ report, item string }

// newlyFailing returns the open §f items that fail only AFTER the
// candidate lines are added — the would-drift delta of a pre-flight replay.
func newlyFailing(
	reports, modLines []string,
	modDF func(string) int,
	modSquashed []string,
	baseFailing map[itemKey]bool,
) [][3]string {
	var newly [][3]string // report, item, reason

	for _, report := range reports {
		for _, f := range CheckReport(report, modLines, modDF, modSquashed) {
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
) {
	for _, f := range newly {
		gatecli.Outf("WOULD-DRIFT [%s]: %s\n", f[0], gatecli.Truncate(f[1], itemMax))
		gatecli.Outf("                  %s\n", f[2])

		for _, c := range culpritAttributions(f[1], candLower, candSquashed, modDF) {
			gatecli.Outf("  culprit: %s\n", c)
		}

		if explain {
			for _, line := range TokenTable(f[1], modLines, modSquashed, modDF) {
				gatecli.Outln(line)
			}
		}
	}
}

// failingKeys indexes every open §f item the corpus FAILS to reflect, so a
// delta replay can tell "already failing" from "newly failing".
func failingKeys(reports, lines []string, tokenDF func(string) int, squashed []string) map[itemKey]bool {
	out := map[itemKey]bool{}

	for _, report := range reports {
		for _, f := range CheckReport(report, lines, tokenDF, squashed) {
			out[itemKey{filepath.Base(report), f[0]}] = true
		}
	}

	return out
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
