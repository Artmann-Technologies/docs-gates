# Eight-issue fix wave + PR#7 merge — session status

Session start 2026-11-22 ~22:40 CET, report cut 2026-11-23 00:17 CET.
Scope: the whole open board — PR #7 (dependabot) + issues #2 #3 #4 #5
#6 #9 #10 #11 — planned as 26 tracked work items, executed in
importance order (data-loss class first, new gate last).

## What landed (verified)

| Item                     | Verdict  | Evidence                                                                                                                                                                                                                                                                                           |
| ------------------------ | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| PR #7 go-finding bump    | MERGED   | stale-base diagnosed (dogfood red was pre-fix corpus); `update-branch` → 3/3 green → merge c4cc05f; provider suite green locally                                                                                                                                                                   |
| #3 routed verdict marker | DONE     | 08f13b7; markers_test pins both contracts; annotate suite green                                                                                                                                                                                                                                    |
| #9 near-miss §f WARN     | DONE     | 9c20b3d + wiring eb8a455; freshness+routing+audit-all, stdout-on-purpose; unit tests both gates                                                                                                                                                                                                    |
| #10 purge open-row guard | DONE     | f7a2480; refuses OPEN-status strikes (exit 2, nothing written), warns unrecognized; IsOpenCell/IsDoneCell shared from docscadence                                                                                                                                                                  |
| #6 header-aware strike   | DONE     | bcc9985; Task/Status/Evidence by header NAME; dry-run proven on InboxClean 4-col + own 5-col; no-Status tables strike task-only                                                                                                                                                                    |
| #4 annotate_corpus knob  | DONE     | 6754250; mirrors audit_corpus; validation + gate-level scope tests + README                                                                                                                                                                                                                        |
| #5 tokenizer slash split | DONE     | 5507ca1; split spelling BINDS (2026-10-18 contract deliberately flipped, test renamed + comment records the flip); parity goldens verified UNCHANGED (corpus has no slash tokens)                                                                                                                  |
| #11 vocabulary gate      | DONE     | dea0759; battery's 9th row via exported docsrows.Tables/SplitCells; legend = table + prose; cell-aware (JSON-in-Notes never flags); unused = INFO; warn-mode rollout; battery parity re-pinned, hand-diff = new PASS stanza + matrix row only; dogfood 9/9 PASS; go-verdict corpus validates clean |
| #2 verdict-rules docs    | DONE     | 8ec0947; README A–G + P table, marker vocabulary incl. routed                                                                                                                                                                                                                                      |
| CHANGELOG [Unreleased]   | DONE     | pre-flighted SAFE, battery-verified, landed in 8ec0947                                                                                                                                                                                                                                             |
| Quality train            | DONE     | buildflow dev: 91 success / 0 failed; erraudit rc=0; nolint-audit 17 needed / 0 stale; vendorHash repaired (PR#7 rode without it — FOD broke, fixed in f9965c4); two file-size cap breaches split properly (purge guard tests → own file; config load-error types → errors.go)                     |
| Trains pushed            | PARTIAL  | c4cc05f..f9965c4 pushed 23:5x; AGENTS.md memory (11b8bc9) + tokenRe wart fix (46a1b59) still local                                                                                                                                                                                                 |
| Issue closure comments   | NOT DONE | blocked: GitHub API outage (gh fails "error connecting" since ~00:05; SSH push still works) + CI-on-push unverified for the same reason                                                                                                                                                            |

## What I got wrong along the way (caught in-session)

1. Typo `n\t\t\tgatecli.Errf` introduced in the purge guard edit —
   caught by build before anything ran.
2. My docsrows export edit used `isTableLine` as anchor and CLOBBERED
   it — caught by grep of references, restored immediately.
3. tokenRe shipped one commit carrying a dead `ERROR|` regex
   alternative (drafting wart) — removed in 46a1b59 (unpushed).
4. File-size caps: docspurge_test.go 395 and config.go 355 lines —
   my additions; fixed by splitting, not suppression.
5. Merged PR #7 without noticing the dependabot branch does not carry
   flake.nix's vendorHash — the provider-suite FOD broke one buildflow
   run; repaired and recorded as an AGENTS.md gotcha (future merges of
   provider dep bumps must bump the pin in the same train).
6. IsOpenCell first draft had a TrimSpace/Trim order bug (leading space
   blocked the backtick strip) — caught by a probe test before commit.
7. Renamed/rewrote TestPreFlightBindsNothingWhenSlashTokenIsSplit —
   DELIBERATE contract flip (#5), documented in the test comment; the
   joined-spelling-required era is over by design, not by drift.

## Deliberate non-fixes / open judgment calls

- The 2026-10-18 test comment's T11 pin survives: the unstruck writer
  routed phrase still never matches markerRe (no "routed" word).
- Vocabulary gate: warn-mode by default (docsstub doctrine) — the fleet
  flip rides a later decision, not this session.
- This report carries NO §f section on purpose: the next-work items
  below are recommendations pending owner decisions (the three
  questions); routing them into TODO_LIST without those answers would
  guess. §f-less reports are clean by construction.

## Battery state at cut time

9/9 PASS immediately before this file landed. Landing it makes 16
unarchived reports against the 15 threshold — cadence ARCHIVE-OVERDUE
is EXPECTED on the next run; the archive ritual is owed (next-work
item #1). Not a regression: the session produced the 16th report.

## Next up (plain list, deliberately unharvested — see questions)

1. Archive ritual: oldest reports (11-05 era) → archived/ after
   check-rows verifies every item struck/routed; battery re-green.
2. Close issues #2–#11 with terse evidence comments (commits named in
   the table above) once gh API returns; verify CI on main (go /
   provider / dogfood) for the pushed train.
3. Push 11b8bc9 + 46a1b59 (with the report).
4. nextd: mark the docs-gates items done IF the queue tracks this org
   (Artmann-Technologies is not a LarsArtmann/* repo — verify first).
5. Fleet rollout decisions: vocabulary_enforce flip timing;
   annotate_corpus adoption for go-taskqueue (drop its skip_gates).
6. go-verdict: move the OUT_OF_SCOPE prose definition under its
   "## Status legend" heading so the vocabulary gate judges it as
   legend, not usage (their file, their call — the INFO lines already
   name the asymmetry).
7. io-surface / timesheets / storbi: re-run their batteries to feel
   #9/#5/#3 land on the corpora that paid for the fixes.
8. Consider an ARCHIVE ritual helper verb (the manual git-mv + banner
   dance is the last un-tooled ritual) — candidate for TODO_LIST.
9. FEATURES.md of this repo: prose-only sections today; if it ever
   grows status tables, give it a "## Status legend" or the vocabulary
   gate stays SKIP (honest, but worth a decision).
10. CHANGELOG cut + version bump when the owner wants a release train
    (nine-gate battery is the headline).

## Session metrics

17 commits on main this session (8 authored trains + 9 daemon-wrapped
heuristic commits). Zero force-anything. One merge commit (PR #7).
Parity goldens re-pinned exactly once (battery + battery-json), hand-
diffed. Lint/format/test/race/nix-build all green at final buildflow
run; erraudit rc=0; 17 nolint directives all needed.
