# CLI I/O surface polish — session status

> **Date:** 2026-11-06 00:14 CET · **Session:** `io-superb` (docs-gates CLI help/error/suggestion surface)
> **Trigger:** "MAKE SURE OUR I/O is superb!" + a mid-session question applying lessons from
> dnsblockd's io-surface fleet brainstorm.

## Scope

**Verified this session, with commands:** full root-module suite `go test ./...`
(9/9 ok), provider module `(cd pkg/provider && go test ./...)` ok, parity goldens
forced re-run `go test ./internal/battery/ -run TestParityGolden -count=1` ok
(byte-identical, no re-pin), `nix develop -c golangci-lint run ./...` 0 findings,
`erraudit lint --type-aware` rc=0, `erraudit nolint-audit .` 13 directives / 0 stale,
battery over this repo `docs-gates run .` PASS ×3 (start, mid, end), real-binary
exit codes spot-checked with a built `/tmp` binary (0/1/2 as contracted).

**NOT verified:** `nix flake check` and the pre-commit hook leg were not run this
session (flake untouched; the daemon, not a reviewed commit, landed the train —
the standalone nolint-audit equivalent WAS run clean). The daemon-wrapped commits
were never re-tested as a SHA baseline — verification was over the working tree,
which `git status` (clean) then matched to HEAD. `buildflow` full loop stays RED
on treefmt/nix-fmt — the documented BuildFlow roundtrip bug (AGENTS.md), worked
around by running lint/fmt/test steps directly.

---

## a) FULLY DONE

1. **Root help surface**: `docs-gates --help`, `-h`, and `help` print a root
   usage (purpose line, all six verbs with one-liners, help/config/version
   pointers, the zero-config guarantee, the exit-code contract) to STDOUT,
   exit 0. No-args keeps the usage-error shape (stderr, exit 2).
2. **Per-verb help**: `docs-gates help <verb>` and `<verb> --help|-h` print the
   gate package's exported `Usage` const — six texts (battery, freshness,
   annotate, routing, cadence, strike), each documenting positionals, boolean
   and valued flags with when/why prose, exit codes, and config knobs.
   Intercepted in the dispatcher in the same slot as `--version`.
3. **Did-you-mean suggestions** for unknown verbs: unique-prefix first
   (`fresh` → `freshness`), edit distance ≤ 2 second (`freshnness` →
   `freshness`), ambiguous prefixes (`r`) refuse to guess; case-folded.
4. **`gatecli.FatalUsage(verb, err)`**: every argv-shaped refusal now ends with
   `usage: docs-gates <verb> --help` — wired in all six verbs, including the
   valued-flag malformed paths (`--audit-max=abc`, `--annotate-max=x`, bare
   `--replace`) and strike selector errors (zero/multi/double-strike refusals).
   Config/environment failures (LoadConfig, unreadable files) deliberately keep
   the bare `FATAL: %v` shape — the fix is not in the argv. Golden-pinned
   refusal texts (dump-items-refusal) untouched.
5. **Unknown-verb output reworked**: `unknown verb "fresh"` + suggestion +
   one usage footer line pointing at `docs-gates help`.
6. **Dispatch restructure**: verb constants, `verbHandlers` map replaces the
   switch (vocabulary = the map), `displayName`/`verbOverview`/`verbUsage`/
   `verbOrder` tables with a table-agreement test; `--version` precedence over
   `--help` pinned by test (probe-script contract).
7. **Drift guards (help cannot lie)**: per-gate `usage_test.go` pins `Usage`
   against the parser's real flag vocabulary BOTH directions (accepted →
   documented; documented → parses), including valued-flag parses through
   `CutIntFlag`/`parseReplaceFlags`/strike `ParseArgs`. Allowed-flag lists
   hoisted to package vars in freshness/annotate/routing for exactly this.
8. **New tests**: gatecli `WantsHelp` + `FatalUsage` byte-pinning; main_test
   help shapes (9 cases), usage errors (6), version precedence,
   every-verb-reaches-its-gate (proves the map↔dispatch, not just the maps),
   suggestVerb table (11 cases incl. ambiguity refusals).
9. **Consistency fix**: docsrouting's bare `return 1`/`return 0` →
   `gatecli.ExitFindings`/`ExitClean` (no output change).
10. **parseAnnotateMax normalized** to parseAuditMax's `(int, []string, int)`
    badExit shape — see (d) for how the first attempt went.
11. **Docs**: AGENTS.md (dispatch table row + the interception/FatalUsage
    bullet, incl. "goldens pin none of the help/FATAL-refusal paths"),
    CHANGELOG [Unreleased] Added entry, README help sentence, ROADMAP --json
    row now names the provider's text-parsing debt (io-surface brainstorm
    cross-ref).
12. **io-surface brainstorm lessons applied** (the user's mid-session
    question): precision-first suggestions (noise kills adoption), usage↔parser
    parity tests (two-sources-of-truth fence), judgment-bearing help texts
    (facts alone are worthless), zero-config adoption line, and one real gap
    it exposed — strike selector errors are argv-class and now carry the hint.
13. **Dogfood rituals for THIS report**: §f harvested into TODO_LIST (new
    section), oldest report (2026-11-04_02-24) archived under the banner
    convention — battery re-run green after both (this was forced: the new
    report is #16, crossing max_unarchived=15).

## b) PARTIALLY DONE

1. **Commit hygiene**: all work landed, but daemon-wrapped into three heuristic
   `chore: auto-commit` entries (fbc8ad3, 0bcbd62, 1aacc5f) — no curated train,
   no attribution line, and no commit-time hook leg ran over it (standalone
   equivalents were run clean before the daemon picked files up).
2. **Flagless-verb usage guards are heuristic**: battery/cadence
   `usage_test.go` asserts no `\n  --` bullet lines rather than an explicit
   no-flags contract; a prose flag mention would false-fail, a non-bullet
   mention would false-pass.
3. **strike `--status` default text in Usage is un-pinned** to the
   `DefaultStatus` const — they can drift apart silently.
4. **AGENTS.md says "all 15 erraudit directives"**; the live count is 13
   (0 stale). Pre-session drift, noticed here, not reconciled.

## c) NOT STARTED (all session-scoped observations)

- `--json` machine face for the verdict matrix (ROADMAP'd for a while; this
  session added the provider-frees-from-text-parsing rationale).
- `docs-gates doctor` verb (ROADMAP'd).
- `--version` at argv[0] with trailing args still reads `unknown verb
  "--version"` (pre-existing; noted, deliberately skipped).
- `docs-gates help freshness extra` silently ignores `extra` (noted in design,
  accepted for now).
- Full-output-byte snapshot tests for the six Usage texts (contains-based
  assertions today; a golden per help text would be cheap insurance).

## d) TOTALLY FUCKED UP (caught in-session, fixed before any commit)

1. **parseAnnotateMax first edit was semantically broken**: I dropped the FATAL
   print, wrote a wrong "unreachable" comment, and left the caller's
   nil-argv sentinel reading a path that no longer existed — a malformed
   `--annotate-max` would have silently run the gate over the cwd instead of
   refusing. Caught on self-review within minutes; fixed properly by changing
   the signature to the badExit shape. The ~2-minute window in which the
   auto-commit daemon could have immortalized the broken state: real, unhit.
2. **editDistance first draft**: `curr := append(prev[:0:0], prev...)` built a
   row of length len(b)+2 — an off-by-one DP corruption. Caught immediately
   on re-read, before any test run.

Both failures share one root cause: signature/shape churn executed faster than
it was thought through. The nets (self-review, then tests) caught both — but
two sloppy-then-fixed cycles in one session is a pattern to correct, not luck
to celebrate.

## e) WHAT WE SHOULD IMPROVE

- **Sketch before edit on signature-changing refactors** — both (d) items were
  multi-value signature churn; a 10-second shape sketch would have prevented
  each.
- **Daemon-commit labeling**: deliberate trains should be committed by the
  session (attribution form) BEFORE the ~3-minute daemon window; this session
  lost that race three times. (Owner question g2.)
- **Help-text byte snapshots** beyond contains-assertions.
- **Reconcile stale AGENTS.md numbers** (erraudit 15→13) at the next
  docs-health pass.
- **TODO_LIST half-strike debt**: cadence reports 18 rows carrying DONE
  WITHOUT a ~~verdict~~ wrap (printed every battery run, never acted on this
  session).
- **The b.N → b.Loop() modernization** in docsfreshness/bench_test.go surfaced
  by gopls twice this session and left alone.
- **Report-authoring footgun, caught by this repo's own gate (live this
  session)**: the `## f) ...` heading IS the freshness §f section — a `## §f)`
  heading does NOT match the gate's `^#{2,}\s*f\)` regex and is inert; every
  deliberately-not-harvested §f row needs its OWN `<!-- fresh:skip: ... -->`
  marker (a shared "rows 11-15 stay local" note exempts nothing), and rows
  that pass on accidental generic-token overlap are passing on sand — mark
  them, do not trust the match.

## f) Next (harvested where actionable; tags: [todo] now in TODO_LIST, [roadmap] already in ROADMAP, [skip] brainstorm-only)

1. [todo] `--json` verdict-matrix output face + provider migration off
   human-text parsing (the io-surface findings-as-data direction).
2. [todo] Wrap-or-strike the 18 bare-DONE TODO_LIST rows (cadence's recurring
   half-strike finding).
3. [todo] Pin strike Usage's default-status text to the `DefaultStatus` const.
4. [todo] Replace the flagless-usage `\n  --` heuristic with an explicit
   no-flags assertion.
5. [todo] Reconcile the AGENTS.md erraudit directive count (13 vs 15).
6. [todo] runHelp: warn when extra args after `help <verb>` are ignored.
7. [todo] bench_test.go: adopt `b.Loop()` (gopls modernization).
8. [todo] Archive-cadence stays tight: this session archived exactly one
   report; the corpus sits AT the 15 threshold again — the next report owes
   the next archive.
9. [roadmap] `docs-gates doctor` verb (PATH, config, flake state; provider
   HealthCheck hint is the seed).
10. [roadmap] Generic counts gate, link/reference checker, report-age gate,
    false item-boundary detector (all ROADMAP'd candidates).
11. [skip] <!-- fresh:skip: brainstorm-only, deliberately report-local --> `--version` handling when it appears at argv[0] with trailing args.
12. [skip] <!-- fresh:skip: brainstorm-only, deliberately report-local --> Usage-text byte goldens per verb (beyond contains-assertions).
13. [skip] <!-- fresh:skip: brainstorm-only, deliberately report-local --> e2e-level help smoke in pkg/provider (it builds the real binary
    already).
14. [skip] <!-- fresh:skip: brainstorm-only, deliberately report-local --> Interactive/alias surfaces (`docs-gates verbs`) — low value.
15. [skip] <!-- fresh:skip: brainstorm-only, deliberately report-local --> Fleet rollout acceleration (pbx-artmann migration) — untouched
    here, unaffected by the help surface (wrappers exec valid verbs only).

## g) Questions for the owner

1. **Version-vs-help precedence**: when both `--version` and `--help` ride one
   invocation, version wins today (pinned by test, reasoning: probe scripts).
   Keep that contract, or flip to the conventional help-first?
2. **Daemon-wrapped trains**: this session's work landed as three heuristic
   `chore: auto-commit` entries. Do you want a labeling strategy (annotated
   tag per session train, or sessions committing their own trains
   pre-daemon), or are heuristic entries acceptable history?
3. **Next Pareto slice**: is the `--json` machine face the right next feature
   (frees the provider from text parsing), or does a new gate family (counts
   gate / link checker) outrank it?

4. **The `(f) Next` list above is this report's §f section** (the `## f)`
   heading is what the freshness gate harvests): [todo] rows 1-8 are
   harvested into TODO_LIST §7, [roadmap] row 9-10 live in ROADMAP.md,
   [skip] rows 11-15 carry individual `<!-- fresh:skip -->` markers.
