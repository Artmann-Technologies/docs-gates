# T1 fail-closed freshness — mid-train status (plan v3, first execution session)

2026-11-24 22:32 CET. Execution session against
docs/planning/2026-11-24_01-07_SUPERB-pareto-execution-plan-v3.md (21
tasks, 135 micro-tasks). This report covers ONLY this session's run.

## Scope

Verified by command, this session: root go test ./... green (three full
runs); probes A/B recorded before and after the judge change (output in
the session log, re-runnable as documented below); battery run over this
repo (audit-all red with 47 findings, then re-derived twice as verdicts
landed); pre-flight SAFE + land-parity OK (12/12 lines verbatim) for the
harvest batch. NOT verified: golangci-lint / erraudit / nolint-audit
over the changed files (not yet run — T1.8's lint leg is open); the
README/CHANGELOG rows for the behavior change (not yet written); the
remaining 34 fresh:skip markers (not yet written, battery audit-all
still red until they land); provider module untouched; no tags, no
proxy checks, nothing pushed. Repo HEAD at report time: 86727ae
(auto-commit daemon; the session's own detailed T1 commit is still
owed — see d2).

## a) Fully done

1. Orientation: TODO_LIST.md re-read line by line; the board matches the
   plan baseline exactly (161 lines, ten corrupted struck rows at the
   plan's exact lines L27/29/31/32/33/38/58/73/87/121; 228-line AGENTS.md).
2. T1.1 + T1.2 probes (the hole, documented): fixture repos under /tmp
   proved the pre-fix judge scores FRESH for (a) a report with no f)
   section, (b) a bullet-only f) section, and (c) an item whose two
   distinctive tokens match two DIFFERENT living lines. Post-fix, (a)
   and (b) WARN on gate stdout and (c) fails with the new reason
   "tokens never share one living line" plus per-token attributions.
3. T1.3 + T1.4 + T1.5 code, all tests green:
   - internal/docsfreshness/tokens.go: one shared line-level predicate
     (lineMatchesKey), Squash key hoisted out of the per-line loop.
   - internal/docsfreshness/corpus.go: memoized inverted index
     (lineHits) + Corpus.SameLineQuorum implemented as a flat
     (line, token) pair sort-and-walk — allocation-light on purpose.
   - internal/docsfreshness/check.go: itemVerdict requires the quorum
     on ONE living line (two matched tokens, one distinctive, same
     line); new drift reason names the coincidence class.
   - internal/docsfreshness/nearmiss.go: FindSectionFGaps +
     WarnSectionFGaps (missing-heading and bullet-item blind spots),
     WARN on gate stdout — the battery's PASS rows keep it visible.
   - Wired into freshness (checkNewest), audit-all, and routing; the
     routing gate's routed() now rides the same SameLineQuorum bar
     (T1.5 alignment decided: same-line, not corpus-wide).
4. Six new unit tests pin the bar (separate-lines-are-coincidence,
   same-line pass, quorum helper, single-token distinctive bar, gap
   matrix, freshness WARN-on-PASS). Three existing fixtures honestly
   redesigned so their baselines PASS the new bar (preFlightRoot pair
   anchor; docspurge zorphuser anchor; attribution DF 8 to 9).
5. TestReplaySpeedGuard caught my first SameLineQuorum implementation
   (per-item map) regressing the CV-shaped replay (2.4 s vs the 1 s
   budget); rewritten flat, guard green 3/3, benchmark allocations back
   near baseline. The guard did exactly its job.
6. Parity goldens UNTOUCHED — the fifteen pinned invocations are
   byte-identical under the new judge (the parity corpus was honest, so
   the planned T1.8 re-pin became a verified no-op; only better).
7. The 47 audit-all findings the fix exposed were individually
   classified with evidence checks (ci.yml concurrency absent, FEATURES
   rows present, one nixpkgs node in flake.lock, erraudit dist/ present
   with no ignore directive, safe-strike carries --cell/--set-cell).
8. Harvest batch landed through the full ritual: 12 candidate lines
   (one §3 row, a new §12 section with four rows, four ROADMAP bullets)
   pre-flighted SAFE (after two verifiably-done 04-00 items were
   skip-marked so their tripwire tokens could not be genericized by the
   batch — the pre-flight caught that interaction, exactly its purpose)
   and verified with --land-parity: 12/12 landed verbatim.

## b) Partially done

1. T1.7 (battery green here): 2 of ~36 fresh:skip markers written; 34
   remain (verdicts already composed with per-item evidence). Until
   they land, the audit-all gate is red with the named backlog — this
   is the honest state the fix was supposed to produce.
2. T1.7 docs: README note + CHANGELOG rows for the same-line bar and
   the new WARNs are not yet written ("provider rows and the README
   note" leg).
3. T1.8 lint leg: golangci-lint fmt + run + erraudit + nolint-audit
   over the changed files not yet executed this session.
4. T1.9 commit: not landed under my own message (see d2).

## c) Not started

T2 repair ten corrupted struck rows; T3 standing struck-row check;
T4 purge collector fixes; T5 battery panic guard; T6 release cut;
T7/T12 fleet roll-out and sweep; T8 owner decision sheet + SystemNix
window; T9 refusal UX pair; T10 strike the 20 done rows + purge;
T11 AGENTS.md distill + doc batch; T13 hygiene batch; T14 BuildFlow#115;
T15 BuildFlow local tooling; T16 md5 scoped exception; T17 flake
tripwire extension; T18 cross-project notes; T19 formatter config +
four files; T20 library-side batch; T21 io-surface batch. All 21
task entries, 135 micro-tasks: the session spent its budget inside T1.

## d) TOTALLY FUCKED UP (caught, fixed, recorded)

1. My skip-marker script located the 04-00 report's item "| 3 |" by
   NUMBER and marked an innocent FINDINGS-table row (a different table
   in the same report) as done. Caught by post-edit verification
   (grep showed the marker beside the wrong text; the §f item kept
   failing). Fixed: marker removed from the wrong row, re-applied by
   content match on the real §f row. Lesson (standing now): report
   edits match on item TEXT, never on position or bare number — the
   same class as the fleet lesson "never read batched sed ranges as
   argument-ordered".
2. The auto-commit daemon raced the session four times: the whole T1
   code change + harvest batch landed as five bare chore: auto-commit
   entries (86727ae and earlier) before my planned detailed T1 commit
   (T1.9) existed. Nothing is lost and the tree is clean, but the
   detailed message (probes, diff, fixture rationale) now has no
   commit to ride. Improvement: commit the moment a task goes green —
   do not batch commits across a long task.

## e) WHAT WE SHOULD IMPROVE

1. Commit discipline under the daemon: stage + commit immediately after
   each green battery, not at task end.
2. Marker/annotation edits: content-match always (see d1).
3. Prefer verified no-op parity over planned re-pins: when a change can
   be proven byte-stable against the goldens, keep the corpus honest
   and skip the re-pin entirely.
4. Fixture design for same-line tests: every pre-flight/purge fixture
   now needs a pair-anchor line so its baseline passes the bar — encode
   that pattern in the helper comment (done for preFlightRoot and the
   docspurge fixture) so the next fixture author does not rediscover it.
5. The audit-all backlog (47 items) proves archived reports were never
   truly verified pre-fix; the per-item verdict ritual just performed
   is the template for any future replay-size change (same doctrine as
   the crush-config audit_max drain: never raise a pin, verdict every
   item).

## f) Next (up to 50, plan order)

1. Land the 34 remaining fresh:skip verdict markers (composed, evidence
   per item) — then battery audit-all green.
2. README note + CHANGELOG rows for the same-line bar + WARNs.
3. golangci-lint fmt + run + erraudit + nolint-audit over changed files.
4. T1.9-style detailed commit for the T1 train (or a follow-up
   documenting the daemon split).
5. T3.1 planted fixture: shifted struck row + struck-before-pipe, prove
   cadence is blind today.
6. T3.2-T3.5 docscadence detection of mis-strikes, FAIL row both faces,
   golden re-pin hand-diffed, CHANGELOG, commit.
7. T4.1 red test: struck-before-pipe row drafts a stray-pipe bullet.
8. T4.2 collector parses cells; strip the leading pipe.
9. T4.3 refuse open-status rows via IsOpenCell (exit 2, nothing written).
10. T4.4 applyPurge writes CHANGELOG first, TODO second; order asserted.
11. T4.5 parity dry-run unchanged; T4.6 CHANGELOG + commit.
12. T2.1 inventory script over the ten corrupted rows.
13. T2.2 git log -S recovery of each row's original cells.
14. T2.3 verdict per row from the striking session's report.
15. T2.4 repair script, dry-run default.
16. T2.5-T2.6 apply one commit per row, battery green each.
17. T2.7 rows verb exit 0; cadence counts the struck rows again.
18. T2.8 agent-notes line + final commit.
19. T5 panic guard: red subprocess test, deferred recover, FAIL row,
    JSON face, pin, CHANGELOG.
20. T9.1 lift editDistance into one shared helper.
21. T9.2-T9.3 unknown-key WARN suggests nearest key, tie rule.
22. T9.4 strike refusal example invocation (pin + hand-diff).
23. T9.5 README examples + CHANGELOG + commit.
24. T13.5 out-of-corpus candidate refusal check (unit test).
25. T13.1 writePreservingMode dedupe into gatecli (three call sites).
26. T13.2 rows verb over docs/review/libs + this report.
27. T13.3 fix or accept what rows flags.
28. T13.4 markdown-lint over the audit corpus.
29. T10.1 re-check the 20 done-rows' evidence at execution time.
30. T10.2-T10.4 strike batches via safe-strike --batch.
31. T10.5 purge dry-run then apply; verify CHANGELOG entry.
32. T6.1 cut CHANGELOG version section.
33. T6.2 provider install hint re-pin + tests.
34. T6.3 annotated tags, push, fresh-cache proxy check.
35. T6.4 cmdguard CI pin bump.
36. T6.5 strike fix rows with the tag; agent-notes line.
37. T7.1 release-doctrine date confirmation; T7.2 agents_enforce flip
    with planted-stale AGENTS.md test.
38. T7.3 re-measure blast radius with the fixed binary.
39. T7.4-T7.6 sweep repos 1-45 (battery per repo, verdict logged).
40. T7.7 stale go-structure-linter suppressions check.
41. T8.1 owner decision sheet (11 rulings, recommendation + cost of
    deferring each) — the session cannot answer these itself.
42. T15.1 rebuild buildflow binary; T15.2 full dev run, read warnings.
43. T15.3 severity-gap issue with run 45112672861 evidence.
44. T11 AGENTS.md distill to <=220 lines + doc batch.
45. T12 sweep 46-67 + per-adoption seeding + ratchet pilots.
46. T17 tripwire extension to every master-tracking input.
47. T18 cross-project lessons (crush-config, SKILLS runbook).
48. T14 BuildFlow#115 guard + PR; T16 md5 scoped exception ask.
49. T19 dprint config regeneration (owner-gated for the write-once
    pair); T20 library-side issues; T21 io-surface after ruling.
50. Battery + tripwire + push at every train boundary (dogfood CI run
    should stay green throughout).

## g) Questions I cannot answer myself

1. The daemon pre-empted T1's commit (five chore entries). Want a
   follow-up empty-change commit carrying the detailed T1 message for
   history's sake, or is the chore trail acceptable as-is?
2. The 34 pending fresh:skip markers will leave ~15 write-once reports
   dprint-dirty (same accepted class as the four known files). Format
   them inside T19 anyway, or leave write-once files dirty forever?
3. T8's eleven rulings (battery adoption scope, io-surface, adoption
   path, cadence policy, erraudit --max, release protocol, cmdguard
   report placement, session exclusivity, cmdguard v5, nixpkgs entry,
   SystemNix push): answer them in one sitting now, or should the
   session continue the T2-T5 train first and batch the sheet later?

## f) §f next tasks (the harvest input)

1. Land the 34 remaining fresh:skip verdict markers for the fail-closed
   replay backlog so audit-all returns green (TODO §12 carries the
   landed rows; the verdicts are composed with evidence).
2. Write the README note and CHANGELOG rows for the same-line freshness
   bar and the new harvest-blind WARNs (plan T1 docs leg).
3. Run golangci-lint fmt + run + erraudit + nolint-audit over the T1
   changed files and land the owed detailed commit for the train.
4. Execute T3 (standing struck-row check) and T4 (purge collector
   fixes) before T2 repairs the ten corrupted rows (plan order
   T3, T4, T2).
5. Execute T5 (battery panic guard) to close the first six-hour train.
6. Execute T10 (strike the 20 done rows, then purge) and T6 (release
   cut) to carry the judge-integrity fixes out to the fleet.
7. Prepare the T8 owner decision sheet (eleven rulings) — it gates T11,
   T21 and the flip date.
8. Sweep the fleet per T7/T12 on the released binary and record the
   adopt / not-owed verdict per repo.
