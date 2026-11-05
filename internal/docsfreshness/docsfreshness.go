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
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// refuseOutsidePreFlight rejects the pre-flight-only flags used outside
// pre-flight mode: --duplicates asserts a deliberate second copy (its +1
// DF push is modeled), --replace names the living-doc line a rewrite
// removes so the candidate words are judged alone. rc 0 = allowed.
func refuseOutsidePreFlight(flags map[string]bool, replacesUsed bool) int {
	if flags["--pre-flight"] {
		return 0
	}

	if flags["--duplicates"] {
		gatecli.Errln("FATAL: --duplicates only applies to --pre-flight mode — it declares that an")
		gatecli.Errln("  on-disk candidate line is about to land as a SECOND copy (ledger-block class)")
		gatecli.Errln("  so the +1 token-DF push of that copy is modeled instead of refused")

		return gatecli.ExitUsage
	}

	if replacesUsed {
		gatecli.Errln("FATAL: --replace only applies to --pre-flight (rewrite) mode — it removes the named")
		gatecli.Errln("  living-doc line before the replay so the candidate words are judged alone")

		return gatecli.ExitUsage
	}

	return 0
}

// dumpReportArg is the positional index of the --dump-items report path
// (positional[0] is the repo root).
const dumpReportArg = 1

// runDumpItems validates the report positional and prints the extraction
// decisions for it.
func runDumpItems(positional []string) int {
	if len(positional) <= dumpReportArg {
		gatecli.Errln("FATAL: --dump-items needs the report path after the root: " +
			"docs-freshness . --dump-items docs/status/2026-11-04_16-45_report.md")

		return gatecli.ExitUsage
	}

	return DumpItems(positional[dumpReportArg])
}

// runPreFlight reads the candidate lines from stdin and runs the
// additions-only or rewrite-mode replay.
func runPreFlight(root string, flags map[string]bool, replaces []string, settings Settings) int {
	candidate, err := readCandidates()
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	return PreFlight(root, candidate, replaces, flags["--explain-token"], flags["--duplicates"], settings)
}

// parseAuditMax splits --audit-max=N off the argument list: the parsed pin
// (-1 = strict), the remaining args, and a nonzero exit code on a malformed
// value (0 = ok).
func parseAuditMax(args []string) (int, []string, int) {
	auditMax, rest, err := gatecli.CutIntFlag(args, "audit-max")
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return -1, nil, gatecli.ExitUsage
	}

	return auditMax, rest, 0
}

// parseReplaceFlags splits the repeated --replace=OLD flags off the
// argument list (= form only — the old line text rides one argv element,
// spaces included). A bare --replace is refused: without = the target
// would silently be empty.
func parseReplaceFlags(args []string) ([]string, []string, int) {
	var replaces []string

	rest := make([]string, 0, len(args))

	for _, arg := range args {
		if value, ok := strings.CutPrefix(arg, "--replace="); ok {
			replaces = append(replaces, value)

			continue
		}

		if arg == "--replace" {
			gatecli.Errln("FATAL: --replace needs the old line in = form: --replace=the old line text")

			return nil, nil, gatecli.ExitUsage
		}

		rest = append(rest, arg)
	}

	return replaces, rest, 0
}

// Run is the CLI entry (repo-root positional + --audit-all flag semantics).
// The repo's .docs-gates.json (when present) drives the corpus paths, the
// living-docs list and the distinctive bar; a malformed config is a usage
// error.
func Run(args []string) int {
	// The audit ratchet (coin audit-ratchet, minted in the source fleet): --audit-max=N pins
	// the tolerated unharvested count so the audit goes red only on a
	// REGRESSION past N — draining a legacy backlog in one sitting was
	// never this gate's job (it was permanently red 2026-10-08..23 while
	// the backlog sat at 78). Valued form --audit-max=N; -1 = strict.
	auditMax, rest, badExit := parseAuditMax(args)
	if badExit != 0 {
		return badExit
	}

	replaces, rest, badExit := parseReplaceFlags(rest)
	if badExit != 0 {
		return badExit
	}

	flags, positional, unknown := gatecli.ParseFlags(
		rest,
		"--audit-all",
		"--explain-token",
		"--pre-flight",
		"--duplicates",
		"--explain-item",
		"--margins",
		"--dump-items",
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

	settings, err := loadSettings(root)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	if rc := refuseOutsidePreFlight(flags, len(replaces) > 0); rc != 0 {
		return rc
	}

	if flags["--explain-item"] {
		// A FRESH corpus prints no --explain-token tables (they ride
		// UNHARVESTED findings only); this mode answers the margin
		// question for ANY item — including harvested ones sitting at
		// DF just below the bar. The needle is the SECOND positional:
		// `docs-freshness . --explain-item rewrite`.
		return runExplainItem(root, positional, settings)
	}

	if flags["--margins"] {
		return TokenMargins(root, settings)
	}

	if flags["--dump-items"] {
		return runDumpItems(positional)
	}

	if flags["--pre-flight"] {
		return runPreFlight(root, flags, replaces, settings)
	}

	living := loadLivingDocs(root, settings)

	if flags["--audit-all"] {
		return auditAll(root, living, flags["--explain-token"], auditMax, settings)
	}

	return checkNewest(root, living, flags["--explain-token"], settings)
}
