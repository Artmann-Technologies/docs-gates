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
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

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

	settings, err := loadSettings(root)
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
		return runExplainItem(root, positional, settings)
	}

	if flags["--pre-flight"] {
		candidate, err := readCandidates()
		if err != nil {
			gatecli.Errf("FATAL: %v\n", err)

			return gatecli.ExitUsage
		}

		return PreFlight(root, candidate, flags["--explain-token"], flags["--duplicates"], settings)
	}

	livingLines, livingSquashed, tokenDF := loadLivingDocs(root, settings)

	if flags["--audit-all"] {
		return auditAll(root, livingLines, livingSquashed, tokenDF, flags["--explain-token"], auditMax, settings)
	}

	return checkNewest(root, livingLines, livingSquashed, tokenDF, flags["--explain-token"], settings)
}
