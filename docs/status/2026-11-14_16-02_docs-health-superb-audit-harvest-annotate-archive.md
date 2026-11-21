# docs-health superb audit — harvest, annotate, archive (session status)

**Scope line.** This report VERIFIES with fresh runs: the battery 5/5
PASS over this repo (before, mid, and after the sweep), the erraudit-rev
tripwire (red on arrival — flake.lock drifted to a broken master rev —
restored forward to the verified 5c841cf2 pins, own commit `2aa134b`),
erraudit nolint-audit 17/17 needed 0 stale, `nix flake check
--all-systems --no-build` all checks passed, the pkg/provider suite ok,
`--check-verdict-outside` CLEAN over 199 struck blocks, and
`--land-parity` OK over every pre-flighted batch (25 lines verbatim).
NOT verified: CI on GitHub hardware for this tree (unpushed), erraudit's
master state beyond the pinned rev, and the sibling-repo legs the board
tracks. Trigger: the owner's explicit docs-health demand ("view ALL
2026-* files, execute the skill superbly") — the standing 2026-11-04
"only at cadence pressure or on demand" ruling's on-demand leg.

## a) FULLY DONE

1. **Corpus read + classified.** All 20 unarchived 2026-* files (14
   status + 4 planning + 2 drafts) read via three analysis passes; every
   open item inventoried with evidence; archived/ deliberately unopened
   (closed history per the skill).
2. **ANNOTATE sweep — the TODO §1 row closed.** 200+ demonstrably-landed
   items struck inline across 14 reports with resolvable verdicts (hash
   claims all validated against git before use; annotate gate 0 findings
   at 199 struck blocks). The three June-era-style bulk-annotation
   hazards avoided: specs built from emit-keys/verified output, per-cell
   table strikes, open rows left bare.
3. **ARCHIVE batch.** The 11-04 trio (19-25, 22-52, 23-06) + planning v1
   (superseded by v2) moved with `git mv` behind `Live remainder` /
   `Superseded` banners naming the owning board rows; gates green
   pre-move. Cadence 14 → 12 unarchived (threshold 15).
4. **HARVEST batch 1** (the 11-12/11-14 ledgers, all pre-flighted SAFE,
   land-parity OK): TODO §4b (six owner questions incl. v0.3.1 ruling,
   adoption path, cadence policy, --max enforcement, release-claim
   protocol) + §5 (thirteen work rows incl. the pingen-spike adoption —
   whose explicit binding un-flipped a WOULD-DRIFT the first pre-flight
   caught).
5. **HARVEST batch 2** (the concurrent 15-52 recovery report's §f, which
   its own §c.1 deferred to this pass): TODO §6 (21 rows: erraudit +
   BuildFlow pushes, race-watch, tripwire generalization, io-surface
   +962% investigation, pre-commit arming...) + four more §4b owner
   rows (cross-repo pushes, session exclusivity, golangci posture,
   cmdguard v4→v5). Its §f rows that this pass resolves are struck
   (harvest, meta-docs review, coordination, the annotate sweep itself).
6. **Board repair.** The two malformed `| ~~ |` rows (erraudit re-pin,
   Cut v0.2.0) rebuilt as proper strikes; F94 (AGENTS distill
   `39bed4d`), dependabot PR #1 (`c8e8fbc`), both WONT-DO rulings, and
   the §1 sweep row struck — then the FIRST real `docs-gates purge
   --apply` run (7 rows → CHANGELOG [Unreleased] → ### Changed,
   pre-flighted by the verb itself; TODO now carries 0 struck rows).
7. **Living-doc drift fixed.** README (seven verbs + purge row +
   `--land-parity` + v0.3.0 + CONTRIBUTING/SECURITY/agent-notes links),
   FEATURES (fifteen parity invocations, seven verbs, purge verb,
   `--check-verdict-outside`, race CI lane, releases through v0.3.0),
   AGENTS ("six invocations" → fifteen), DOMAIN_LANGUAGE (verb = five
   gates + run + purge), ROADMAP (release-cadence bullet through v0.3.0,
   io-surface d1/b7 verdicts corrected to the clean-build re-run —
   resolving a TODO↔ROADMAP split brain — plus fail-fast, sandbox-safe
   battery, ledger-shape, pre-flight/margins convergence, JSON-face,
   version-verb fuel), CHANGELOG ([Unreleased] seeded: repo meta,
   land-parity case fix, check repairs, lock restores).
8. **Empty-section decay removed** from TODO_LIST (§2/2b/2d/2e/§7
   scaffolding); audit-all re-run green after the removals (the
   un-modeled-removal risk checked empirically, pre-commit).

## b) PARTIALLY DONE

1. **The 11-05/11-06 reports stay unarchived** — every remaining open
   row is owner-gated or sibling-scoped and live on the board; they are
   annotated (landed rows struck) but not banner-moved. The cadence gate
   is comfortable (12/15), so no pressure to force them.
2. **Drafts** (annotate-archived-scope knob = issue #4;
   freshness-tokenizer slash-join) remain open proposals — annotated as
   still-pending, deliberately untouched.
3. **The two 11-12 ledgers are now partially struck** (answered/routed
   rows) — their remaining brainstorm rows stay bare by design.

## c) NOT STARTED (deliberately)

1. v0.3.1 cut (owner ruling in §4b rides it).
2. Any sibling-repo leg (SystemNix window, pbx bump, erraudit/BuildFlow
   pushes) — all board rows, all owner-gated.

## d) TOTALLY FUCKED UP

1. **The `| tail` exit-code mask** bit me once mid-sweep (an annotate
   §c call failed invisibly behind a pipe; the output's own "exit
   status 1" tail line was the tell). Unmasked re-run caught it.
2. **Two pre-flight refusals were self-inflicted printf mangling** (the
   candidate never reached the gate; "Present: --" was the mangled line
   itself). `%s` form fixed it — candidates go through files, never
   printf formats.
3. **A concurrent session landed mid-sweep** (the 15-52 six-layer
   recovery; the battery went red on its unharvested §f.9). Handled
   forward: full read, harvest batch 2, re-green — but the sweep's
   mid-flight battery snapshot was briefly red through no fault of the
   corpus state.

## e) WHAT WE SHOULD IMPROVE

1. The blob-substring on-disk refusal flags any candidate containing
   `--` (flag names) — the TODO §5 row to make structural headings
   implicit duplicates is the same family; a word-boundary matcher would
   kill the class.
2. The fail-fast flip class (23-06 §f.48 genericizing on every batch
   that says "battery"/"local") cost three pre-flight rounds — fragile
   anchors get re-found by every pass until ROADMAP carries the bullet
   (it does now).

## f) Next

| # | Item                                                                                      | Routes to                                         |
| - | ----------------------------------------------------------------------------------------- | ------------------------------------------------- |
| 1 | Everything this pass opened or carried                                                    | TODO_LIST §1-§8 (the board IS the routed backlog) |
| 2 | The owner questions (v0.3.1, adoption path, cadence policy, pushes, posture, exclusivity) | TODO_LIST 4b                                      |
| 3 | Unshipped fuel (sandbox battery, ledger shape, fail-fast, version verb)                   | ROADMAP                                           |

## g) Archive manifest (the 2026-11-14 sweep)

| File                                                           | Classification     | Deciding reason                                                                                                 |
| -------------------------------------------------------------- | ------------------ | --------------------------------------------------------------------------------------------------------------- |
| docs/status/2026-11-04_19-25_evening-docs-health-close.md      | ARCHIVED           | 22 landed items struck; remainder (float/deploy/switch, D1/D2) routed to TODO 3/4 pre-move; banner names owners |
| docs/status/2026-11-04_22-52_plan-v2-execution-session.md      | ARCHIVED           | landed §f/§g struck; owner tail routed to the 04-50 §g.5 window                                                 |
| docs/status/2026-11-04_23-06_session-self-review-and-status.md | ARCHIVED           | 28 landed items struck; sibling/owner tail routed TODO 3/4; ROADMAP fuel stays                                  |
| docs/planning/2026-11-04_07-14_SUPERB-pareto-execution-plan.md | ARCHIVED           | superseded by v2 (its header records execution); in-repo tier done; banner cites TODO 3/4                       |
| docs/status/2026-11-04_21-02_owner-decision-preps-D1-D2.md     | LEAVE ALONE (open) | D1/D2 rulings still pending; the doc's whole value is the pending decision                                      |
| docs/planning/2026-11-04_19-28_...v2.md, 2026-11-05_02-50 plan | ANNOTATED, kept    | executed tiers struck at parent level; owner/sibling rows still open                                            |
| docs/drafts/*                                                  | LEAVE ALONE (open) | live proposals (issue #4; tokenizer) — not decidable here                                                       |

## h) Harvest ledger (batches 1+2, dispositions)

| Source                                        | Disposition     | Destination                                                              |
| --------------------------------------------- | --------------- | ------------------------------------------------------------------------ |
| 11-14 §f.2/11/21/23/26/27/30/32 + §g Q1-Q3    | new rows        | TODO §4b/§5                                                              |
| 11-14 §f.12/22/24/25/45/50                    | declined/routed | ROADMAP fuel or foreign tracking (BuildFlow convention row)              |
| 12-30 §f.3/4/5/17/20-22/27/33                 | new rows        | TODO §4b/§5 (--max ruling; carrier tests; FATAL UX; duplicates-implicit) |
| 12-30 §f.24                                   | declined        | overtaken by the 39bed4d AGENTS distill                                  |
| 07-33 §f.23/24/25/34/35/38/39/47              | new rows        | TODO §5                                                                  |
| 07-33 §f.9/10 (release authority)             | done            | struck in-report (releases cut)                                          |
| 15-52 §f.1/5-11/18/20/22-25/27-29/42/43/45/46 | new rows        | TODO §6 + §4b                                                            |
| 15-52 §f.8/35/44/50                           | existing rows   | TODO §3 / ROADMAP (dedupe)                                               |
| 15-52 §f.38-40/47                             | declined        | foreign-repo fuel — belongs in BuildFlow's tracking                      |
| 15-52 §f.14/32                                | declined        | standing rules already in AGENTS                                         |
| 15-52 §f.15                                   | done in-audit   | SECURITY/CONTRIBUTING/.editorconfig/.gitattributes reviewed accurate     |
