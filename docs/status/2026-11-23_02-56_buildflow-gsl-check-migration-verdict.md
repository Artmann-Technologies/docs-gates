# Status Report — BuildFlow/gsl check-migration verdict & session status (docs-gates)

- Written: 2026-11-23 02:56 CET
- Session window: ~22:30 (11-22) → 02:56 (11-23), interrupted by a session
  exit (~4 h wall-clock gap) mid-subagent
- Scope: ONE question — "which checks that ~/projects/BuildFlow and/or
  ~/projects/go-structure-linter perform should be in docs-gates instead?"
- Code changes this session: **zero** (pure cross-repo analysis; no gate
  was touched, no golden re-pinned, nothing implemented)

## 1) The question and the method

Two read-only subagents inventoried both repos; every load-bearing claim
in the final verdict was then verified firsthand in this repo and in
BuildFlow's `internal/cli/docs_check*.go`, `internal/config/`, and
gsl's rule registry/README. Delivered in two parts (BuildFlow verdict
first, gsl correction after its report landed — the gsl subagent hit a
provider rate limit once and was resumed).

## 2) The verdict (carried here so it outlives the chat)

**Boundary drawn:** BuildFlow = code/build/env health + MECHANICAL doc
hygiene (format, style lint, network link check, spelling, presence +
repair). docs-gates = SEMANTIC doc health (freshness, drift, verdicts,
cadence, glossary truth). gsl = Go repo-structure linting.

### Move INTO docs-gates (priority order)

1. **AGENTS.md size knob** (smallest, kills a live split brain): gsl
   `agent-config` (exist CRITICAL + mtime ≤14d HIGH + ≤220 lines HIGH,
   `internal/rules/agent_config_rule.go`) and BuildFlow doctor
   `agents-md-size` (≤220) both duplicate this repo's `agents` gate,
   which already owns existence (missing = fail severity) and mtime
   bars `agents_warn_days=7` / `agents_max_days=14` — the 14-day
   constant now exists in three independent copies. Add
   `agents_max_lines` here; gsl + BuildFlow delegate.
2. **BuildFlow stamps → the `stamps` gate**: BuildFlow `freshness`
   (30-day `verified ... on` WARN over FEATURES/TODO_LIST/
   DOMAIN_LANGUAGE, `docs_check.go:135`) and `benchmarks-freshness`
   (90-day hard-FAIL over docs/BENCHMARKS.md, `docs_check.go:146`)
   are expressible today via `stamp_files` + `stamp_warn_days` /
   `stamp_fail_days` (`internal/config/config_defaults.go:29-39`);
   this repo's `stampRe` (`internal/docsstamps/docsstamps.go:173`) is
   a superset of BuildFlow's `verifiedOnPattern`
   (`docs_check.go:127`). Two gaps: (a) ONE global bar pair — no
   per-file bars (30-warn vs 90-fail); (b) a corpus file with no
   stamps is a SKIP here (`docsstamps.go:214`) but a missing
   BENCHMARKS stamp FAILs in BuildFlow → needs a require-stamp knob.
3. **BuildFlow `gotcha-anchor` → the `citations` gate** (same class:
   doc pointer at code must resolve). Anchor pattern
   `docs_check.go:131`, body `docs_check_3.go:14-110`; this gate
   already resolves `path:line` and `path:lo-hi` over configurable
   `citation_files`. To absorb: line ≤ EOF bounds check (unverified
   here), skip rules (bare filenames, `..`-elided upstream paths),
   corpora AGENTS.md + docs/GOTCHAS.md.

### New-gate candidates

- **`integrity`** from BuildFlow doctor `docs-integrity`: tracked
  `.md` husks (empty on disk, non-empty blob in git history — the
  daemon data-loss signature), WARN with byte-exact recovery command,
  SKIP outside a git repo. Hermetic, docs-scoped, fits.
- **`links`** from BuildFlow `cross-ref` (file-existence markdown link
  resolution + allowlist): BORDERLINE — hermetic and stdlib-able, but
  drags an allowlist policy while its natural companions (lychee)
  are network-bound and must stay in BuildFlow.

### Stays in BuildFlow

`markdownlint` (style, opt-in via config presence), `lychee` (network),
`codespell` (dictionary), `securitymd`/`licenseforge` (presence +
repair — this battery is judge-only), the two lychee FP-risk doctor
checks (they guard BuildFlow's own lychee step), and the claim-vs-code
trio (`provider-count`, `module-count`, `module-list`,
`command-coverage`): their ground truths are BuildFlow-internal (DI
`ProviderList`, go.work, cobra tree). A generic `claims` gate needs a
config-named truth command — biggest design lift, HOLD until a second
consumer exists.

### gsl: nothing else moves

65 rules (7 groups, `internal/rules/registry.go`): Go layout, go.mod,
gitignore/attributes, CI, test structure, security — code-scoped. Its
docs-scoped surface (`readme`, `readme-install-section`, `changelog`,
`contributing-file`, `code-of-conduct-present`, `domain-language-md`)
is presence/scaffolding + one-regex shape floors, all paired with
`--fix` template generation — repair semantics do not fit a judge-only
battery, and the glossary already gets far deeper judgment here (stub
checksums, citations, stamps, vocabulary) than gsl's 10–500-line /
≥3-terms floor. Confirmed absent from gsl: TODO_LIST/ROADMAP
inspection, report freshness, docs-vs-code drift.

### Incidental find

BuildFlow's provider registration comment still says "five-gate
battery" (`tools/providers/sdk_imports.go`) — stale (ten gates since
agents joined) but harmless: the JSON face's gate rows are generic, so
the provider absorbs rows without drift (vocabulary and agents proved
it).

## 3) a) Fully done

- Both inventories delivered: BuildFlow (4 layers: ~110 pipeline
  providers, in-process SDK modules, 46 doctor checks with a
  15-member gate profile, external tool-sdk providers; 42-row check
  table) and gsl (65 rules, 7 groups, per-rule implementation paths).
- Firsthand verification of every load-bearing overlap: this repo's
  ten-gate table (`internal/battery/battery.go:70-80`), stamps/citations
  config knobs (`known_keys.go:31-41`, `config_defaults.go:29-39`),
  stamps skip semantics (`docsstamps.go:204,214`), citations shapes
  (`docscitations.go` + tests), both stamp regexes read and compared,
  BuildFlow freshness / benchmarks-freshness / gotcha-anchor bodies
  read, gsl README rule list + registry-backed inventory.
- Consolidated verdict delivered (3 moves, 2 new-gate candidates of
  which 1 borderline, explicit stay lists, priority order).
- Subagent crash recovery: both agents were interrupted by the session
  exit; both resumed via `agent_send` with full reports delivered.

## 4) b) Partially done

- `docscitations` line-bounds semantics: NOT yet verified whether it
  bounds-checks line numbers against EOF (gotcha-anchor does); named
  as gap #5 on the board, not resolved.
- Stamp-regex compatibility: superset claim is from reading both
  regexes, not from a test corpus proving byte behavior on BuildFlow's
  phrasings (`verified via ... on`, `Last verified on`) — and this
  repo's `stampRe` excludes `|` (table-cell bounded) while
  BuildFlow's is line-wide `[^|\n]`-free; edge cases unproven.
- The claim-vs-code trio and cross-ref bodies: taken from the
  BuildFlow subagent report, not personally read (`docs_check_2.go`,
  `docs_check_3.go:111+` unread by me) — flagged as subagent-sourced
  in the verdict, not yet verified.
- Doctor checks (`docs-integrity`, `agents-md-size`, lychee-*, and
  BuildFlow's full 46-check doctor surface): subagent-sourced only.
- The verdict's durability: it lived only in chat; this report is the
  first artifact. A `docs/planning/` note is still unlanded (board
  item 35).

## 5) c) Not started

- All three migrations (agents_max_lines knob; stamps per-file bars +
  require-stamp; gotcha-anchor absorption). Nothing implemented.
- The new `integrity` gate; the `links` decision spike; the `claims`
  hold decision record.
- BuildFlow delegation PRs (stamps, agents-md-size retirement,
  gotcha-anchor retirement, stale comment fix) and the gsl agent-config
  shrink PR.
- TODO_LIST/ROADMAP harvest of the candidate board below (deliberately
  pre-triage: it sits outside §f per the 02:39 report's convention, so
  the harvest contract does not fire on untriaged candidates).
- Any AGENTS.md memory updates for this repo (candidate additions are
  in section 7 of this report, landed only on instruction).

## 6) d) Totally fucked up

Nothing. Near-miss ledger, honestly:

- The session exited mid-subagent (~4 h wall-clock lost); both
  subagents were interrupted with reports undelivered. Recovered: their
  sessions were kept and both completed after `agent_send` resume.
- First two resume attempts addressed the subagents by LABEL
  (`buildflow-checks`) instead of the returned NAME
  (`mindy-buildflow-checks` / `boris-gsl-checks`) — two failed calls
  before the correct names.
- The gsl subagent failed once with a provider rate limit; hedged by
  reading gsl's README rule list directly in parallel, so the final
  answer never depended on the retry.
- The FIRST `ls` of `internal/` this session returned a stale listing
  missing the five content-gate packages (docsvocabulary, docsagents,
  docsstub, docscitations, docsstamps) — contradicting AGENTS.md; the
  re-read listed all 15 packages. Unexplained tool glitch; caught only
  because AGENTS.md disagreed. No conclusion was built on the stale
  listing (the interim answer cited the ten-gate surface from AGENTS.md
  and the re-read).
- The interim BuildFlow-only verdict was delivered before the gsl
  report arrived — intentional pipelining, but a reader acting on it
  alone would have missed the gsl agent-config duplication (the
  strongest single finding).

## 7) e) What we should improve

- **What I forgot / under-verified:** line-bounds semantics of
  citations (gap #5); a stamp-phrasing test corpus (gap #14); personal
  reads of the claim-trio and doctor bodies. None blocked the verdict,
  but each is a landed-cost risk for the implementation phase.
- **Subagent naming:** record that resume must use the RETURNED name,
  not the label — cost one wasted round trip under time pressure.
- **Redundancy policy:** when a subagent fails once (rate limit),
  immediately do the cheap direct read in parallel instead of serial
  retries — the gsl README peek is the pattern to repeat.
- **Provenance discipline:** subagent-only facts were labeled as such
  in the verdict ("taken from the subagent report, not personally
  read") — keep that; it converts "forgot to verify" into an explicit
  board item instead of silent trust.
- **Cross-repo analyses should land a docs/planning note in the same
  session** — the analysis survived only because the status report was
  requested; a planning note is the durable home.
- **Stale tool output:** re-verify surprising reads against a second
  source before building on them (AGENTS.md disagreed with the stale
  ls — that disagreement is what made the glitch visible).

## 8) Next up (candidate board — pre-triage, deliberately outside §f)

Pre-triage candidates, NOT yet harvestable into TODO_LIST/ROADMAP;
triage lands rows on instruction. (Convention: the 02:39 report keeps
its candidate board outside §f too — untriaged items must not enter
the harvest contract, which requires a verdict or a landed row per
item.)

Migration implementation (docs-gates first):

1. Add `agents_max_lines` to the agents gate: known key, default
   decision (fleet 220 vs off), Settings projection, tests, usage text,
   parity re-pin. Kills the three-copy 220-line split brain.
2. Decide `agents_max_lines` default semantics (0/-1 = off vs always
   on) BEFORE writing the knob — data-model first.
3. stamps gate: design per-file age bars — owner decision between
   per-file override syntax inside `stamp_files` vs a named
   stamp-class table (files → class, class → warn/fail pair).
4. stamps gate: require-stamp knob (missing stamp FAILs when enforced;
   today it is a SKIP, `docsstamps.go:214`).
5. Verify `docscitations` line-bounds behavior against BuildFlow
   gotcha-anchor's line ≤ EOF check; add the bounds check if absent.
6. citations gate: absorb gotcha-anchor skip rules (bare filenames,
   `..`-elided upstream paths are unverifiable → skip, not fail).
7. citations gate: verify line-range bounds (lo ≤ hi ≤ EOF) exist; add
   if absent.
8. Stamp-phrasing test corpus: prove `stampRe` byte-behavior on
   BuildFlow's `verified via ... on` and `Last verified on` phrasings
   (including a table-row line, since this repo's regex excludes `|`).
9. Parity re-pin after each gate change
   (`PARITY_UPDATE=1 go test ./internal/battery/ -run TestParityGolden`,
   then hand-diff — a golden bump is a contract change).
10. Provider strict-decode proof after stamps/citations output-shape
    changes (new finding shapes must decode; the vocabulary/agents
    joins are the precedent).
11. New `integrity` gate: tracked `.md` husks (empty on disk, non-empty
    blob in git history), WARN naming the husk with the byte-exact
    recovery command (`git show <rev>:<path> > <path>`), SKIP outside
    a git repo, `integrity_enforce` warn-mode knob, legacy verb name
    decision (docs-integrity?) + parity goldens.
12. `links` gate decision spike: hermetic file-existence markdown link
    resolution + allowlist config; decide whether it earns a gate or
    stays BuildFlow-side with lychee.
13. Record the `claims`-gate HOLD decision (config-named ground-truth
    command; revisit on a second consumer) in ROADMAP so it stops
    being re-derived.
14. Dogfood the citations absorption HERE first: this repo's AGENTS.md
    carries many `file:line` references — trial `citation_files`
    including AGENTS.md before proposing the BuildFlow delegation.
15. Fleet .docs-gates.json guidance: stamp_files for
    FEATURES/TODO_LIST/docs/DOMAIN_LANGUAGE/docs/BENCHMARKS with the
    right bars (30-warn living / 90-fail benchmarks) once per-file
    bars land.

BuildFlow side (each needs its own PR; cross-repo, on instruction):

16. BuildFlow PR: delegate freshness + benchmarks-freshness to the
    docs-gates provider; keep `buildflow docs --check` as a thin
    wrapper so docs-gate.yml CI semantics stay intact.
17. BuildFlow PR: retire doctor `agents-md-size` (delegate to the
    agents gate + `agents_max_lines`).
18. BuildFlow PR: retire `gotcha-anchor` (delegate to citations via
    `citation_files`).
19. BuildFlow: fix the stale "five-gate battery" comment in
    `tools/providers/sdk_imports.go` (ten gates now) — trivial.
20. BuildFlow: map FAIL classes after delegation (docs --check exits 1
    on benchmarks expiry today; stamps needs `stamp_enforce` flipped
    there to keep the hard-FAIL semantics — warn-mode rollout note).
21. BuildFlow: verify-before-filing the lockfile-rewrite recurrence
    upstream (3rd occurrence 11-23; carried from the 02:39 report,
    still unfiled).

gsl side (cross-repo, on instruction):

22. gsl PR: shrink `agent-config` to existence (CRITICAL) + fixer;
    drop the mtime-14d and 220-line halves with a pointer to
    docs-gates.
23. gsl: decide the fate of the now-unused `FileMustBeFresh` helper
    (delete or keep as test DSL) after the shrink.
24. gsl: update the README rules table and per-rule docs pages after
    the shrink (docs/rules/<name>.md are generated — regenerate).

Battery/dogfood ritual for THIS session's artifact:

25. Run the battery over this repo after this report lands
    (`go run ./cmd/docs-gates run .`) — expect a red point-in-time
    until the board is triaged; timestamp the run.
26. Triage the board above into TODO_LIST (short/mid-term rows) and
    ROADMAP (long-term), then strike the landed rows here.
27. Carry forward the 02:39 report's still-open rows (release train
    v0.3.0, `vocabulary_enforce`/`agents_enforce` flip decisions,
    TODO_LIST §9 strike) — do not re-derive, strike on completion.
28. After any gate change: golangci 0-findings, `erraudit lint ./...
    --type-aware`, `erraudit nolint-audit .`, `nix flake check
    --all-systems --no-build`.
29. Decide whether AGENTS.md belongs in `citation_files` permanently
    (depends on items 5-7 and 14).
30. docs/status corpus size check before the next audit-all
    (`ls docs/status | wc -l`), archive-ritual cadence per
    `max_unarchived`.

Housekeeping carried from earlier sessions (unchanged by this one):

31. pre-commit leg candidate: fire `check-erraudit-rev.sh` whenever
    flake.lock is staged (mechanical guard for the 3-time recurrence).
32. Clear the LSP wrapper caches so typecheck ghosts stop seeding
    distrust in future sessions.
33. erraudit upstream: vendorHash-staleness incident row (from the
    05-43 report remainder, TODO_LIST §3).
34. go-taskqueue: adopt the `annotate_corpus` knob (draft exists).
35. Land a `docs/planning/2026-11-23_check-migration-verdict.md` note
    carrying section 2 of this report (the durable home for the
    analysis).

## 9) g) Owner questions (cannot be figured out from any repo)

1. **Sequencing and authority:** land all three migrations
   docs-gates-first, then delegate in BuildFlow/gsl as separate PRs
   (my recommendation — keeps this repo's contract changes reviewable
   and lets the provider absorb rows as proven), or one coordinated
   cross-repo train with the consumer deletions in the same window?
   This touches two other repos you own and decides whether
   `buildflow docs --check` must stay byte-compatible for CI mid-train.
2. **stamps data model:** for per-file age bars, per-file override
   syntax inside `stamp_files` (e.g. `"docs/BENCHMARKS.md:90"`) or a
   named stamp-class table (classes carry warn/fail pairs, files join
   a class)? Expensive to migrate once fleet configs exist — your
   call on the shape.
3. **claims gate:** build the generic doc-claims-vs-ground-truth gate
   here at all, or keep it BuildFlow-forever? It hinges on fleet
   knowledge I don't have: is there a second repo whose living docs
   carry numeric ground-truth claims (count-style) worth judging
   generically, or is BuildFlow's DI/go.work/cobra trio the only
   consumer?
