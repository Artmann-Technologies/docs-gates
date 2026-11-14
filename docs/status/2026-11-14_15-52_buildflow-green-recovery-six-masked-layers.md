# 2026-11-14 15:52 — BuildFlow Green Recovery: the Six-Layer Masked-Latent Dig

Scope line: this report VERIFIES the final state with these artifacts: the
final `buildflow --build-mode full` run (91 success / 0 failed, exit 0,
"passed with warnings", `/tmp/final2.log`), `nix flake check --all-systems`
("all checks passed!"), `go test -race ./...` green across the root module
and `pkg/provider`, `bash scripts/check-erraudit-rev.sh` PASS, golangci at
0 issues, erraudit `lint`/`nolint-audit` green (17 directives, 0 stale),
dogfood battery PASS, parity goldens byte-identical, `git status -sb` clean
and pushed (`origin/main == HEAD` at 9005c5d). NOT verified: CI on GitHub
hardware (the new `test (race)` step has never run there), the erraudit and
BuildFlow upstream pushes (both still local), the concurrent session's own
work (89d7dc6 cmdguard-adoption report — out of scope per instruction), and
who/what pushed mid-session (I only observed the ref move).

Session shape: the user's `buildflow --fix --build-mode=full` run failed
(exit 69; nix-fmt, treefmt, nix-build-verify red; failure loops at 95%+).
Fixing the first layer kept unmasking the next: six layers total — lockfile
damage, BuildFlow tool bugs, docs-gates' own dead checks, fleet-policy
findings, real data races, oversized files. Each layer had been invisible
because the layer above it aborted the run first.

## a) FULLY DONE

1. **flake.lock recovery** — the repair phase's `nix-flake-update` had
   bumped six inputs (erraudit → `75cfc156` with a stale upstream
   vendorHash, cmdguard v4.1.0→v5.0.0 major, nixpkgs, go-filewatcher,
   go-nix-helpers, gogenfilter), daemon-wrapped in `2adbc73`. Forward
   restore from `0ef4b12` + own commit `e1b8ca3`; tripwire PASS. Same
   runbook as the 2026-11-06 incident.
2. **BuildFlow flake-pin-consistency false-FAIL fixed** — the checker's
   declaration regexes required line-start `<name>.url` / `<name> = {`,
   missing the equivalent flattened `inputs.<name>.url` form our flake
   uses. Fixed with an optional `inputs.` prefix + 4 regression tests.
3. **The nix-fmt/treefmt roundtrip wedge fixed upstream** (red since
   2026-11-04, "known bug" status retired). Three stacked causes, each
   live-probed in /tmp sandboxes before fixing: (1) no git repo in the
   sandbox → treefmt tree-root discovery refused explicit paths ("path .
   not inside the tree root /nix/store") or walked the ENTIRE filesystem
   (3.5M files reproduced); (2) with a git repo, nix's clean-source filter
   refuses untracked inputs; (3) missing `flake.lock` coupled formatting
   verification to CURRENT upstream state. Fix: sandbox is git-init'd,
   every copied file staged, flake.lock rides along on every lane. Steps
   went 18ms/8ms green immediately and stayed green (23ms/3.6s) in the
   final run.
4. **BuildFlow pre-commit audit compliance** — my first commit there was
   REJECTED by its own audits (raw `exec.Cmd.CombinedOutput`, 2 em dashes
   in new code, two files over the 350-line ceiling). Refactored to
   `bfrunner.CombinedOutputCmd`, split `format_roundtrip_git.go` +
   `format_roundtrip_sandbox_test.go`, ASCII punctuation. Committed
   `1d474a028`.
5. **golangci-drift check repaired** — unpassable since 2026-11-04: it
   counted pins AFTER `sort -u` (two identical job pins counted 1 ≠ 2
   demanded) and two distinct pins could never equal the single flake
   version. Now counts raw pins, then compares the collapsed unique value
   (`4572726`).
6. **shell-lint repaired** — bare `actionlint` requires a git project
   above the workflows dir; the store source has none. Explicit workflow
   paths skip discovery.
7. **docs-battery pre-commit leg scoped to real commits** — git-hooks'
   sandbox git-inits a synthetic ONE-COMMIT repo, so annotate's
   verdict-hash evidence ("done at <sha>") was red by construction there.
   `preCommitCheckFor` gained a `sandboxed` flag; the devShell keeps the
   leg (`c1b364d`).
8. **Full nix flake check green**: `nix flake check --all-systems` →
   "all checks passed!" on both architectures (nolint-audit, erraudit-lint,
   provider-suite, golangci-drift, go-floors, shell-lint,
   pre-commit-check, legacy-wrappers smoke, skills guards).
9. **erraudit upstream vendorHash fixed** at the root: `buildflow -s
   nix-hash-fix --fix` in the erraudit repo re-pinned its stale
   go-modules hash (its daemon had swept go.mod/go.sum without re-pinning).
   Result commit `f5a40b3a` builds green.
10. **buildflow binary rebuilt and installed** — stale binary (built
    02:56, HEAD at 09:00) had masked upstream fixes for hours; final
    install is the nix-stamped `1d474a0` via atomic rename (plain `cp`
    hit "Text file busy").
11. **cue-vet error cleared** — fleet cue policy requires `permissions:`
    in CI workflows; added `permissions: contents: read` to ci.yml
    (daemon-committed `987aea7`).
12. **flake-meta-checker error cleared** — provider-suite check meta
    gained homepage/license/maintainers.
13. **go-structure-linter error cleared** — CI gained a `test (race)`
    step (`go test -race ./...`) with the incident as rationale.
14. **Three REAL data races fixed** (found by test-race running for the
    first time in months once the nix steps stopped blocking it):
    a. `TestLandParityFailsOnRetypedLanding` combined `t.Parallel()` with
       `gatecli.CaptureOutput` — the documented "capture tests stay
       serial" invariant had a holdout racing every parallel printer.
    b. The half-budget watchdog goroutine fired mid-capture and read the
       swapped `os.Stderr` via `Errf`. gatecli gained `liveErr` (stderr
       twin of `liveOut`) + `LiveErrf`; the JSON face hands the watchdog
       the snapshot writer.
    c. `CaptureOutputWithTimeout`'s timeout path and the abandoned gate
       goroutine BOTH ran `pipes.restore()` — a straggler's late global
       reassignment raced the next capture. Restore is single-shot
       (`sync.Once`).
    `go test -race ./...` green across root + provider; parity goldens
    byte-identical.
15. **TestReplaySpeedGuard made race-aware** — the absolute 1s budget
    ignored that race instrumentation alone slows the loop ~18× (50ms →
    0.92s) and fired under any machine load (3.3s during my parallel nix
    builds). Budget scales to 6s under the race detector via build-tag
    consts (`race_on_test.go`/`race_off_test.go`).
16. **Six oversized files split** under the 350-line ceiling (file-size
    findings gate): gatecli.go → `gatecli_output.go` + `gatecli_capture.go`
    (505→255+105+156); gatecli_test.go → capture tests out (405→339+73);
    battery.go → `battery_watchdog.go` (361→320+48); battery_test.go →
    `battery_json_test.go` (483→323+170); config.go →
    `config_defaults.go` (363→324+42); docspurge.go →
    `docspurge_collect.go` (395→267+134). No behavior change; suites,
    lint, erraudit, battery, goldens all green after.
17. **AGENTS.md distilled to the 220-line ratchet** — 486 → 219 lines;
    every distinct concept preserved, full incident narratives moved
    verbatim to `docs/agent-notes.md`; session facts folded in (repaired
    checks, roundtrip fix, docs-battery scope, stale-binary lesson,
    erraudit upstream fix). Dogfood battery PASS at 219.
18. **`buildflow format` run** — cleared the formatter-fingerprint drift
    (cached verdicts re-verified under new toolchain identity); dprint
    repaired 2 files (reflow + one budget overage, reclaimed).
19. **Stray lint repairs landed deliberately** — `preflight_land.go`
    gosec/paralleltest fixes + `.golangci.yml` `go: 1.27` from the
    earlier buildflow run's repair phase were sitting uncommitted in the
    worktree; verified and committed as an own train (`e5eaa07`) instead
    of letting the daemon fossilize them.
20. **All docs-gates trains pushed** — `origin/main == HEAD` at `9005c5d`;
    final run verdict "BuildFlow passed with warnings", 91/0 failed,
    doctor 38 ok / 3 warn / 0 fail (the 3 warnings are the documented
    accepted states: Truncate non-fix, fingerprint drift, dprint remote
    plugins).

## b) PARTIALLY DONE

1. **erraudit push** — `f5a40b3a` (the vendorHash fix) is committed
   LOCALLY in the erraudit repo, unpushed. Until it lands on GitHub, any
   consumer's `nix flake update erraudit` still resolves the broken
   master. Deliberately held: cross-repo pushes need your say-so.
2. **BuildFlow push** — the checker + roundtrip fixes are committed
   locally (my `1d474a028` carried only the last 2 files; the auto-daemon
   there swept the bulk into heuristic commits `deef5fef3`/`34a5bedd1`),
   unpushed. Local installs work; the fleet does not benefit until pushed.
3. **Attribution hygiene in BuildFlow history** — my fix commit's message
   describes the whole change but carries only its tail; the bulk sits in
   heuristic daemon commits. Content is all at HEAD; the history shape is
   ugly.
4. **io-history-ratchet warning** — fired during the broken runs (14 files
   hashed vs median 4), vanished in the final run. Treated as transient
   repair-run noise; never root-caused. It self-resolved, but I cannot
   distinguish "transient" from "will return" without the --profile-io
   dig I skipped.
5. **dprint-remote-plugins** — accepted as a warning (fetch is one-time,
   cached). The documented remedy (local plugin bundle via
   `BUILDFLOW_DPRINT_PLUGINS`) is NOT set up.
6. **go-auto-upgrade warnings** — the accepted non-fixes (samber/lo
   withheld, jsonv1tov2 declined, settings-projection pair) still fire
   every run. Documented, not silenced.
7. **canonical-helper Truncate warning** — documented non-fix (stdlib-only
   contract), still firing.
8. **treefmt sandbox noise** — with the git-init'd sandbox, treefmt now
   prints "Cowardly refusing to install hooks with `core.hooksPath` set"
   during roundtrip runs. Cosmetic (step passes), noted, not fixed or
   upstreamed.
9. **Binary freshness horizon** — the installed binary is stamped at
   `1d474a0`, but BuildFlow HEAD has moved since (daemon commits).
   Next runs may show a small staleness warning again until a rebuild.
10. **Timing-regression verdicts** — go-mod-tidy +800%, golangci-lint
    +97%, go-auto-upgrade +133% etc. were judged machine-load noise
    (they fired during the parallel nix build storm) and never
    re-baselined on a quiet machine.

## c) NOT STARTED

1. docs-health HARVEST of this report's §f list into TODO_LIST/ROADMAP.
2. BuildFlow#87 (file-size-check constant cache key) — upstream status
   check; I used the documented `BUILDFLOW_NO_RESULT_CACHE=1` bypass only.
3. erraudit repo's own AGENTS.md note about its vendorHash incident and
   the nix-hash-fix resolution (the lesson currently lives only in
   docs-gates' memory and erraudit's git history).
4. Investigating the io-surface regression (+962%) visible in BuildFlow's
   own pre-commit run output.
5. Arming pre-commit hooks locally (`git config --unset core.hooksPath`)
   and exercising the docs-battery devShell leg manually at least once.
6. Verifying the concurrent session's cmdguard-adoption status report
   (`89d7dc6`) claims against code.
7. Re-baselining `buildflow timings` on a quiet machine.
8. Darwin support revisit (ROADMAP, unchanged by this session).

## d) TOTALLY FUCKED UP

1. **I broke HEAD and the daemon committed it.** The battery.go watchdog
   extraction: I cut by line numbers, but my own LiveErrf edit earlier in
   the same file had shifted everything — the off-by-N deletion removed
   the `startHalfBudgetWatchdog` DEFINITION while `runGates` kept calling
   it, and the auto-commit daemon swept the broken worktree into commits
   `8f6e954`/`1c6c360` before I noticed. Local HEAD was briefly
   broken-compiling; the daemon's "fossilize a botched edit" hazard — the
   exact one AGENTS.md warns about — caught ME. Forward-recovered from
   `b755d4c` (luckily already pushed). Cost: ~4 extra cycles and a
   temporarily broken local history.
2. **I misdiagnosed the test-race failures for two cycles.** First pass:
   reproduced only the four named tests, saw no DATA RACE, built a
   "race-overhead timing" theory, load-simulated twice, and fixed the
   timing guard — all before running the FULL package under `-race`,
   which immediately surfaced the three REAL data races (and showed the
   "timing" failures were just those races poisoning parallel tests).
   Whole-package evidence first; per-test reproduction second.
3. **The store-walking repro.** My first `nix fmt` sandbox repro ran
   bare (whole-tree mode) instead of buildflow's explicit-file-args
   shape — twice — walking 3.5M/6M files and ~2min each of store-wide
   formatting attempts. Informative, but sloppy execution of my own
   hypothesis test.
4. **AGENTS.md line-budget whack-a-mole: seven iterations** (486→360→
   275→235→229→222→221→219). I never checked the checker's exact
   counting formula (`newlines + 1`, so `wc -l` must be ≤219) until late,
   and dprint's markdown pass re-aligned the table + rewrapped lines,
   silently pushing me back over budget TWICE after I "finished".
5. **Exit-code shells that lied.** `cmd | tail; echo RC=$?` captured
   tail's exit code twice (the 02F flake check printed "CHECK-RC=0" while
   the check had FAILED). Only reading the actual output saved me.
6. **Import-fix whack-a-mole on every split.** Each file split needed
   2-3 import cleanup rounds because I ran `golangci fmt` only at the
   end instead of right after each extraction.

## e) WHAT WE SHOULD IMPROVE

1. **Extract by content anchors, never line numbers** — `lsp_replace_symbol`
   or anchored regex; line numbers rot the moment the file has been
   touched in the same session (paid today, twice).
2. **Commit verified work immediately.** The daemon commits every ~2-3
   min and wraps intermediate states without judgment — it fossilized a
   lockfile bump AND a broken battery.go today. The window between
   "verified" and "committed" is where damage happens.
3. **Whole-suite evidence before per-item theories.** The full-package
   `-race` run found in one command what two per-test cycles missed.
4. **Timing guards race-aware from birth.** Absolute wall-clock budgets
   are wrong under `-race` (~18× here) and under load; the build-tag
   const pattern (`raceDetector`) should be the fleet snippet.
5. **Know the checker's arithmetic before writing to a budget** —
   `strings.Count(data,"\n")+1` vs `wc -l` is a one-line detail that cost
   six iterations.
6. **Formatter immediately after generated/split code**, not at the end
   of a batch — dprint/gofumpt would have made each split one-shot.
7. **Masked-latent layers are the failure mode of a quality pipeline.**
   Every red layer hid the next one (nix build cascade → dead checks →
   policy findings → races → file sizes). A periodic "full gate with all
   layers forced to run" (e.g. `--fail-on none` inventory run) would
   surface the whole queue at once instead of one per incident.
8. **Cross-repo fix latency.** erraudit's stale vendorHash was known
   2026-11-06 and re-struck today because the upstream fix flow is
   manual; pushes of upstream fixes should be part of "fixed", not a
   flagged afterthought.
9. **Sandbox coverage honesty.** The sandboxed pre-commit-check now
   covers only nolint-audit; the full battery rides real commits and CI.
   Honest, but thinner — a sandbox-safe battery variant (env-gated
   hash-rule skip) would restore coverage.
10. **Do not trust pipeline exit codes relayed through pipes** — read the
    output, or use `${PIPESTATUS[@]}` / write logs to files.

## f) TOP 50 THINGS TO GET DONE NEXT

P0 — unblocks the fleet / immediate:

| #  | Task                                                                                                                                                                                                                       |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | Push erraudit `f5a40b3a` (vendorHash fix) — consumers' `nix flake update erraudit` still resolves broken master until then.                                                                                                  |
| 2  | Push the BuildFlow fix train (flake-pin-consistency + roundtrip sandbox) — the fleet's nix-fmt/treefmt stays red everywhere else until then.                                                                                |
| ~~3~~  | ~~docs-health HARVEST: route this report's §f items into TODO_LIST (actionable) and ROADMAP (fuel).~~ done — in the 2026-11-14 docs-health pass (TODO_LIST 5/6 + 4b rows) |
| 4  | Strike the TODO_LIST/report rows this session landed (safe-strike.sh) — the flake-check repairs, race fixes, and splits deserve landed verdicts with hashes.                                                                |
| 5  | Watch the first GitHub CI run on `9005c5d`: the `test (race)` step runs on CI hardware for the first time (6s speed-guard budget under race + CI slowness is the risk point). **→ routed TODO_LIST 6 (race-watch row)** |

P1 — close this session's open threads:

| #  | Task                                                                                                                                                                                                                       |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 6  | Write the erraudit-repo AGENTS.md note: its daemon's go.mod sweep broke the flake hash; nix-hash-fix is the fix; its own "vendor hash drift gate" check should have caught it — why didn't it?                                |
| 7  | File/track the treefmt "Cowardly refusing to install hooks with core.hooksPath set" noise inside BuildFlow roundtrip sandboxes (side effect of the git-init fix; cosmetic but noisy in every nix run).                       |
| 8  | Check BuildFlow#87 status (file-size-check constant cache key): still open? Re-verify the bypass is still needed. **→ duplicate — TODO_LIST 3 carries the BuildFlow#87 row** |
| 9  | Investigate the io-surface regression (+962%) from BuildFlow's own pre-commit run output.                                                                                                                                   |
| 10 | `git config --unset core.hooksPath` locally + run `pre-commit run docs-battery --all-files` once — the devShell leg has never been exercised end to end.                                                                     |
| 11 | Re-baseline `buildflow timings --regressions` on a quiet machine; confirm the +800%/+133%/+97% regressions were load noise, not real.                                                                                       |
| 12 | Adjudicate the daemon-carried `.golangci.yml` `run.go: 1.27` change (root floor stays 1.26): intended fleet policy or repair side effect? Document or revert. **→ routed TODO_LIST 4b (golangci posture ruling)** |
| 13 | Adjudicate go-structure-linter's `go-version` warning (go.mod 1.26 vs available 1.27): document the floor as deliberate, or bump root floor in ONE train with the flake pin (the check's own remedy text). **→ routed TODO_LIST 4b (golangci posture ruling, folds D2)** |
| 14 | Rebuild/reinstall buildflow if binary-freshness warns again (HEAD moved past `1d474a0` via daemon commits). **→ standing rule — AGENTS stale-binary lesson carries it** |
| ~~15~~ | ~~Review the BuildFlow-generated SECURITY.md (v0.3.0 currency claim), CONTRIBUTING.md, .editorconfig, .gitattributes landed by daemon commits today.~~ done — reviewed 2026-11-14 — SECURITY.md/CONTRIBUTING.md/.editorconfig/.gitattributes accurate (v0.3.0 is current) |

P2 — hygiene / robustness:

| #  | Task                                                                                                                                                                                                                       |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 16 | Sandbox-safe battery variant: env-gated annotate hash-rule skip so the sandboxed pre-commit-check regains full-battery coverage. **→ routed ROADMAP (sandbox-safe battery bullet)** |
| 17 | Scope BuildFlow's roundtrip `git add -A .` to the lane's copied file set (perf on large staged-only lanes).                                                                                                                  |
| 18 | Promote the `raceDetector` build-tag const pattern to a fleet snippet (go-finding/toolsdk candidate) — every repo with timing guards needs it.                                                                              |
| 19 | Watch CI duration after the race step; if the go job doubles, move race to its own job or per-package scope. **→ routed TODO_LIST 6 (race-watch row)** |
| 20 | Add `-race` coverage to the provider CI job (currently plain `go test`).                                                                                                                                                    |
| 21 | Link `docs/agent-notes.md` from README/docs index so HARVEST/VERIFY know it exists.                                                                                                                                          |
| 22 | Reset/prune buildflow's nix-fmt (31/32) and treefmt (36/38) failure-history counters — the 95%-failure alerts are now stale noise post-fix.                                                                                 |
| 23 | After BuildFlow pushes: `nix flake update` nothing, but re-run a full buildflow on a second machine to confirm the roundtrip fix travels.                                                                                    |
| 24 | AGENTS.md regrowth trigger: alert when >200 lines (the 220 ratchet holds today at 219).                                                                                                                                      |
| 25 | Extend the erraudit-rev tripwire pattern to the other master-tracking lock inputs (go-filewatcher, gogenfilter, go-nix-helpers) — same silent-bump class.                                                                    |
| 26 | Adjudicate the reverted cmdguard v4→v5 major bump: is the fleet still meant to stay on v4.x here, or was v5 the intended next step that needs a deliberate update + validation?                                               |
| 27 | Deliberately update the master-pinned inputs one at a time (gogenfilter, go-nix-helpers, go-filewatcher) with checks per step, instead of waiting for the repair phase to bulk-bump them again.                              |
| 28 | Set up the local dprint plugin bundle (`nix build .#dprint-plugins` in BuildFlow + `BUILDFLOW_DPRINT_PLUGINS`) to kill the cold-fetch warning.                                                                              |
| 29 | One deliberate `--json` timeout run to eyeball that the watchdog's stderr line routes to real stderr (post-LiveErrf) and the JSON stdout stays pure end to end.                                                              |
| 30 | Consider a visible note when annotate skips hash verification (synthetic git repos look real; currently silent skip is only for NO git).                                                                                     |
| 31 | Dogfood: record the "flake checks were red for 10 days, masked by a cascade" incident as a full narrative in docs/agent-notes.md (cautionary tale for check-layer audits).                                                   |
| 32 | `buildflow doctor` re-run once BuildFlow HEAD settles — confirm binary-freshness + fingerprint verdicts settle too. **→ standing rule — AGENTS stale-binary lesson carries it** |
| ~~33~~ | ~~Coordinate with the concurrent session (cmdguard adoption + land-parity report `89d7dc6`) before harvesting TODO_LIST, to avoid clobbering its in-flight rows.~~ done — resolved — the docs-health pass coordinated on fresh HEAD; no rows clobbered |
| 34 | Verify the io-history-ratchet warning stays gone on the next two runs; if it returns, do the `--profile-io` dig that was skipped.                                                                                            |

P3 — ROADMAP fuel:

| #  | Task                                                                                                                                                                                                                       |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 35 | ROADMAP: revisit Darwin absence only on a real consumer (unchanged). **→ duplicate — ROADMAP Darwin bullet carries it** |
| 36 | ROADMAP: provider jsonv1tov2 reconsideration trigger — when the provider floor moves past the v1 pin, re-run the migration probe. **→ routed ROADMAP (jsonv1tov2 trigger)** |
| 37 | ROADMAP: the sandbox-safe full battery (env-gated) as a feature, not a policy split. **→ routed ROADMAP (sandbox-safe battery bullet)** |
| 38 | BuildFlow: record roundtrip sandbox state (git-init'd, files staged, lock present) in the formatter RunRecord for debuggability. **→ foreign-repo fuel — belongs in BuildFlow tracking** |
| 39 | BuildFlow: failure-loop detector could auto-quarantine a step after N consecutive identical failures instead of warning forever. **→ foreign-repo fuel — belongs in BuildFlow tracking** |
| 40 | BuildFlow: auto-rebaseline the formatter fingerprint after a successful `buildflow format` (the drift warning re-fired even after format until the next doctor). **→ foreign-repo fuel — belongs in BuildFlow tracking** |
| ~~41~~ | ~~docs-health ANNOTATE pass over the 13 unarchived status reports (oldest 10d, threshold 21d — comfortable, but they will hit it).~~ done — in the 2026-11-14 docs-health pass — 200+ items struck corpus-wide, oldest batch archived |
| 42 | gatecli_test.go at 339 lines — one test away from the ceiling; keep an eye or pre-split.                                                                                                                                     |
| 43 | Make scripts/safe-strike.sh discoverable from the docs-battery failure output (it is the tool the ritual names).                                                                                                             |
| 44 | erraudit: re-run the strict `--max` audit after the next deliberate input bump (today's refactors added no new generic-return sites — verified). **→ standing rule — AGENTS erraudit bullet (never cite counts from memory)** |
| 45 | Fleet audit: which other covered repos use the flattened `inputs.<name>.url` flake form (they were all false-FAILing pin-consistency until the upstream fix lands for them).                                                 |
| 46 | Share the "repair-phase lock bump + daemon wrap" recovery runbook to the fleet lessons file (crush-config `references/lessons.md`, by commit).                                                                               |
| 47 | BuildFlow: expose failure-history retention/reset so step alert fatigue can be managed per repo. **→ foreign-repo fuel — belongs in BuildFlow tracking** |
| 48 | docs-gates: benchmark the CI race step's wall time on GitHub hardware; if it dominates the go job, split it out. **→ routed TODO_LIST 6 (race-watch row)** |
| ~~49~~ | ~~Consider annotating (not rewriting) the 2026-11-04/06 entries that claimed flake-check green — they were true for the checks that existed then, but the suite has grown since; a verify stamp would prevent inherited-claim drift.~~ done — the 2026-11-14 pass struck the landed rows in the 11-04/05/06 reports with evidence |
| 50 | ROADMAP: explore making the battery's watchdog line structured (JSON event) so machine consumers can surface it without scraping stderr. **→ duplicate — ROADMAP watchdog-field bullet carries it** |

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Cross-repo pushes:** Should I push erraudit (`f5a40b3a`) and the
   BuildFlow fix train now? docs-gates' "PUSH YOUR OWN TRAINS" policy is
   explicit for THIS repo, and the global rule says never push unasked —
   for sibling repos I held off. The fixes only help the fleet once
   pushed, so this is a policy call, not a technical one.
2. **The concurrent session:** another session is actively working in
   docs-gates (cmdguard-adoption status report `89d7dc6`, its own
   erraudit lock restore `2aa134b`, and something pushed mid-session).
   Should I treat parallel sessions as normal and just rebase my
   doc-harvest work on fresh pulls, or do you want session exclusivity
   while a big recovery like this one is in flight?
3. **The golangci `run.go: 1.27` + root floor question:** the repair
   phase bumped `.golangci.yml` to analyze at Go 1.27 while the root
   go.mod floor stays 1.26 (deliberate contract). Is analyzing at 1.27
   the intended fleet posture (lint at newest, floor pinned lower), or
   should the floor move to 1.27 in one deliberate train (which would
   also retire the go-version warning and change the jsonv1tov2
   calculus)?
