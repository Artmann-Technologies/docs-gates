# Residual Execution: the Five Sanctioned Rows, the Push-Lag Autopsy

2026-11-05 01:15 CET — the execution session that resumed after the 23-06
self-review's WAIT directive. Scope: the only sanctioned in-repo work (the
five harvested TODO rows), nothing from the owner-gated tier. Start state:
HEAD `6717741`, tree clean, battery 5/5, CI green. End state: HEAD
`b9b41a7` pushed to origin (10 commits), CI run `41931531064` success on
BOTH jobs (each now carrying its new golangci-lint leg), tree clean,
battery 5/5, suites green, golangci 0/0 both modules, erraudit rc=0,
nolint-audit 13/13 needed 0 stale, flake check all-systems pass, build
yields 6 binaries stamped `b9b41a7`.

## a) FULLY DONE

1. ~~Verify the handoff state~~ done at `b9b41a7` — HEAD `6717741` clean,
   remote synced, battery 5/5 confirmed before any edit.
2. ~~R1: pin the margins item-token exclusion in the byte-parity net~~
   done at `9fa4a3a`, ritual `4fec344`. The row as written was VACUOUS:
   corpus-replace rode `generic_df: 4`, and the near-bar window
   `[bar-2, bar)` over a one-line living doc can only ever hold DF-1
   tokens — so a bare margins arm would print `0 token(s)` with or
   without the exclusion. Fixed config-only: corpus-replace drops
   `generic_df` to 2, every living token enters the window, and the new
   `margins-replace` golden names only the unclaimed trio (cycle, notes,
   ships — the third was a discovery: the `# NOTES` heading token) while
   the four claimed tokens would headline the report the moment
   `claimedItemTokens` regresses. The T10 preflight-replace golden is
   byte-identical (traced every output line: no boundary crossings, the
   missing-tokens reason keeps all six tokens because found stays empty,
   quorum 2 is never met by a single surviving anchor). The T9
   gitignore-trap suite fired exactly as designed when the new golden
   existed untracked — the trap caught its first real intruder tonight.
3. ~~R2: CI golangci-lint for root AND provider~~ done at `e3226d7`,
   ritual `1fa9b19`. golangci-lint-action v9.3.0 pinned by SHA
   (`d583c34f…`), version `v2.15.0` matching the flake pin exactly; the
   provider job lints the nested module through the same walked-up
   `.golangci.yml` the repo's own command uses. Action ref and version
   syntax verified against the upstream repo via `gh api` before writing
   (no guessed URLs); YAML parse-checked. Proven green on live CI in the
   final push (both jobs).
4. ~~R3: script the strike-sweep safe mechanics~~ done at `4e87b4a`,
   ritual `6245181`. `scripts/safe-strike.sh`: exactly-one-line text
   anchors (full untruncated inventory printed on refusal), anchors must
   carry four or more letters (digits and regex scaffolding can select
   any numbering space), `[^|]*`-bounded cell rewrites, clean-commit-
   boundary check, dry-run default, `--gate` runs the annotate gate per
   batch. Proven against a nine-case refusal matrix on a scratch git
   board, never the live tree — then dogfooded: the script struck its
   own TODO row (task cell + status flip + evidence flip, gate per
   batch) and the annotate gate stayed CLEAN through both applies.
5. ~~R4: the nix provider check proves its e2e skip arm~~ done at
   `2508f53`, ritual `0150b26`. The check's phase now greps the verbose
   test log for the sentinel's sanctioned skip ("root module source not
   adjacent"), so a green check is evidence the end-to-end coverage is
   accounted for. Red-controlled INSIDE the real sandbox: a bogus
   sentinel string fails the build while the Go suite itself still
   prints PASS.
6. ~~R5 wrapper half: per-process analysis cache~~ done at SystemNix
   `59bb7cc35` (daemon-swept commit; all pre-commit gates — gitleaks,
   deadnix, statix, treefmt, flake eval — had already passed over the
   staged content in my first attempt). The wrapper now mints a fresh
   session dir per launch under the same cache root and sweeps session
   dirs older than 48h at launch; probe-verified (a 72h-old session dir
   swept, a 1h-old one kept, each launch gets a fresh dir).
7. ~~R5 diagnosis half: the 22-commit push lag~~ done at `b9b41a7`. The
   daemon NEVER pushed from this host: SystemNix sets `autoPush = false`
   in configuration.nix (since 2026-06-01; the upstream module default
   is false too). The afternoon session waited on a phantom push.
   Timeline reconstructed from the journal: 18:00–18:11 FastFlowLM
   timeouts, 18:30–19:04 a fleet-wide batch-queue-full storm, 19:05 HOST
   REBOOT (fresh boot id), 19:06:31 the restarted daemon's startup sweep
   reports `pushed=0`. AGENTS.md's "pushes too" claim is corrected to
   the truth: every push in this repo's history was a manual session
   push.
8. ~~Final gates + push + CI~~ done at `b9b41a7` — both module suites,
   golangci 0/0 with the expected erraudit warning, erraudit lint rc=0,
   nolint-audit 13 needed 0 stale, flake check all-systems, package
   build with stamped `--version`, battery 5/5, manual push (the daemon
   never will), CI success on go AND provider.
9. ~~Live-PATH probe~~ done (report-time observation): `docs-gates` and
   the five legacy wrappers ARE on the live system PATH
   (`/run/current-system/sw/bin/`) riding version `6ed909f` — the
   pre-hardening rev. T3's headline goal (binaries on live PATH) is
   already materially true; the version is stale pending T2.

## b) PARTIALLY DONE

1. The SystemNix wrapper fix is committed and gate-verified but
   INERT: master is 4 ahead / 0 behind, unpushed (pushing would also
   publish another session's three in-flight daemon commits — not my
   call), and unactivated (the fix deploys at the next switch). Until a
   fresh langserver run is observed minting a `session-*` dir, the fix
   is evidence-backed code, not live behavior.
2. The ghost-diagnostics cure is an ISOLATION bet: the 2026-11-04
   experiment cleared BOTH cache dirs; only the analysis cache was
   made per-process. If ghosts survive the switch, the general
   `GOLANGCI_LINT_CACHE` is the next suspect and needs its own
   experiment.
3. The CI lint legs are green but carry a DRIFT VECTOR: when the flake
   bumps golangci (v2.15.0 → next), the workflow's `version:` input must
   move in the same train or CI and the fleet lint silently diverge.
   Nothing enforces that today.
4. safe-strike.sh's refusal matrix lives only in this session's
   transcript — no in-repo test harness pins it; a future edit to the
   script has no regression net.
5. The live `docs-gates` on PATH rides `6ed909f`, pre-dating the
   hardening and today's five gates' worth of fixes — live consumers
   run stale gates until the T2 float + deploy.
6. Workflow YAML validation was ad hoc (python yaml in the devshell);
   no actionlint anywhere in the toolchain.

## c) NOT STARTED

1. The entire owner-gated tier, by standing instruction: T1 v0.2.0
   release cut, T2 SystemNix float to the hardened flake, T3 deploy
   smoke, T7 fleet ratchet re-runs, T8 pbx-artmann migration, T11 pilot
   ratchet drains, T13 rpi3-dns eval, T14 BuildFlow roundtrip upstream
   issue, T15 BuildFlow#87, T21 skill-text check, T22 nixpkgs
   maintainers PR, F94 AGENTS simplification. Asked five times in the
   previous session, never answered; untouched tonight by design.
2. The docs-health skill cross-check against the new flags (19:01
   report §f.45, lives in crush-config): adjacent to tonight's
   crush-config leg and tempting while I was in sibling territory, but
   it was NOT among the five sanctioned rows — deliberately skipped.
3. Upstream BuildFlow work (roundtrip walker bug, the file-size cache
   key): documented non-fixes here; upstream actions remain open.
4. ROADMAP fuel (margins threshold knob, doctor verb, split-check,
   continuation-digit gate, CutStringFlag, result-cache identity,
   findings-gate cache-state summary): ideas, not commitments; the
   margins knob gained a concrete argument tonight (the bar-2 window
   insight) but nothing was started.

## d) TOTALLY FUCKED UP

1. I wrote five distinct bugs into safe-strike.sh before it worked:
   a bash parse break (the apostrophe in "row's" closed the
   single-quoted perl program — found only by running it), a missing
   `~~…~~` wrap in whole-line mode (the strike silently became a bare
   verdict append), an exclusive-branch combo that DROPPED the cell
   strike when `--set-cell` was present (exactly the TODO-ritual
   combination), an ordinal rail that only caught bare digits
   (`^\| 3 \|` sailed through into a real strike of line 5 during
   testing), and a perl `die` exiting 255 against the script's own
   exit-code contract. The nine-case matrix caught every one BEFORE
   landing and before the script touched the live board — the control
   system worked — but five bugs in one small script is a shameful
   write quality, and two of them (the missing wrap, the dropped cell)
   would have been silent data corruptions had I applied-first.
2. Two of my red controls were INVALID on first run: the ordinal-rail
   test used an anchor shape the rail was never meant to catch (so the
   "pass" proved nothing), and the bounds test hit the dirty-tree rail
   instead of the rail under test. A red control that does not fire for
   the RIGHT reason is theater; both had to be re-run on a clean file
   with a real ledger-row anchor before the rails were honestly proven.
3. I nearly corrupted the corpus-replace fixture by design: my first
   plan added lines to NOTES.md carrying claimed tokens — which the
   quorum arithmetic (matchQuorum 2) would have kept the rewrite-arm
   item anchored, flipping the T10 pin from UNSAFE to SAFE and
   silently destroying the rewrite discrimination golden. Caught by
   reading itemVerdict before editing; the config-only bar change was
   the save. One more step down that path and T10 would have needed a
   re-pin with a weaker contract.
4. The daemon beat me twice tonight and I let it: SystemNix swept my
   wrapper fix into a 20-file chore commit (my attribution lost, though
   the gates had passed on the content), and docs-gates swept the
   AGENTS/TODO docs train into `9155251` (recovered by amending to
   `b9b41a7` — local-only, but the amend of a daemon commit was a
   judgment call, not a sanctioned move). The "commit immediately"
   protocol exists precisely for this; I ran battery-and-format between
   edit and commit and paid both times.
5. Memory maintenance miss: the margins DF arithmetic (window
   `[bar-2, bar)`, DF counts living-doc LINES only, near-bar means
   DF-1 in a one-line corpus) is exactly the non-obvious fact AGENTS.md
   exists for, and I did not write it there at discovery time — it
   survived only in the parity test comment and now in this report.
   The update-at-the-moment rule, violated.
6. Small burns, same root pattern: `head` on the pma binary dumped ELF
   garbage into the transcript (use `file` first); I grep-verified the
   provider check by grepping its RESULT touch-dir — a meaningless
   check that returned 0 and told me nothing (the real proof was the
   build rc plus the in-phase assertions); the R1 fixture work went
   eight tool-calls deep into NOTES.md surgery before I computed the
   near-bar window arithmetic that made the whole question trivial;
   the SystemNix commit raced HEAD twice and burned two full pre-commit
   flake evals (~3 min each) on failed lock attempts; the first T9/T4
   harness invocation ordered flags after positionals and died on
   usage — my own script's convention, which I then violated in my
   own test command.
7. The vendor/flake-pin-consistency preflight FAIL resurfaced in the
   23:25 buildflow run ("flake.nix and flake.lock disagree on 1
   input(s)") and I observed-and-moved-on again — it is tracked in the
   backlog as never-diagnosed, and tonight added one more sighting
   without one more bit of diagnosis.

## e) WHAT WE SHOULD IMPROVE

1. Compute the threshold arithmetic BEFORE designing fixtures. The
   near-bar window is two lines of algebra; the R1 investigation was
   an hour of fixture surgery that the algebra would have deleted.
   Read the gate code, then the numbers, then touch the fixture.
2. Red controls need a validity check themselves: name WHICH rail a
   control exercises and prove the rail fires for that reason, not for
   whichever rail happens to trip first (the dirty-tree rail masked
   two of mine). A control that passes for the wrong reason is worse
   than no control.
3. Write scripts with the refusal matrix IN HAND, not after: five bugs
   in 220 lines came from writing the perl branch logic top-down. The
   nine cases existed before the code (they came from §d of the 23-06
   report) — enumerating them as executable cases FIRST would have
   generated the branch structure instead of auditing it after.
4. Escape-audit quoted programs before running: the single-quote-in-
   perl failure class is now twice-burned (heredocs, embedded
   programs). `bash -n` immediately after writing, before any
   execution, would have saved the detour.
5. Commit trains within one tool call of finishing them: both daemon
   sweeps landed in the battery-format-then-commit window. The format
   pass can ride AFTER the commit as a follow-up train; attribution
   lost to a sweep is not recoverable without history surgery.
6. Verify-what-a-path-IS before probing it: `file` before `head`, and
   a grep against a touch-file result dir is not a verification. Two
   meaningless artifacts tonight came from probing shapes blindly.
7. Record non-obvious facts at discovery, not at report time — the
   AGENTS update rule exists because report-time records depend on the
   report being written; tonight's DF-arithmetic fact nearly died in
   working memory.
8. Belief-audit inherited AGENTS claims when they drive behavior: the
   "pushes too" belief cost a session a 22-commit surprise and an
   autopsy. One `git log origin..HEAD` check would have refuted it
   months ago. Claims about automation behavior deserve the same
   verify-before-trusting discipline as claims about APIs.
9. Prefer config-only fixture changes over payload surgery: the bar-2
   solution changed one JSON integer and preserved three goldens by
   construction. When a pin needs to discriminate, ask what KNOB puts
   it in the discriminating regime before reshaping the corpus.

## f) Up to 50 things we should get done next

1. SystemNix: activate the switch so the per-process analysis cache
   goes live, then probe a fresh langserver run for its `session-*`
   dir (the fix is inert until then).
2. Update AGENTS.md's LSP-ghost protocol after activation: "clear the
   cache dirs + restart" becomes "restart suffices" once every
   langserver starts cold.
3. If ghost diagnostics survive the new wrapper: isolate the general
   `GOLANGCI_LINT_CACHE` with the same cold-cache experiment class
   (the 23-06 cure cleared both dirs; only one was fixed).
4. SystemNix: publish master (4 ahead, 0 behind — carries the wrapper
   fix plus another session's three daemon commits; needs an owner
   ruling, see §g).
5. Attribute the 19:05 host reboot: confirm from deploy logs whether
   it was the evo-x2 deploy window (which would make the stale
   `6ed909f` on PATH the deploy's content, not an accident).
6. T2: float SystemNix to the hardened docs-gates flake — the live
   PATH binaries still ride pre-hardening `6ed909f` (live-verified
   tonight: `docs-gates --version` on `/run/current-system/sw/bin`).
7. T3: post-float smoke — the six binaries on live PATH answering
   `--version` with the hardened rev (headline goal already materially
   true at a stale rev).
8. Encode safe-strike's nine-case refusal matrix as an in-repo bash
   harness (`scripts/safe-strike-test.sh`) so the script's next edit
   has a regression net.
9. Add a `--battery` flag to safe-strike.sh: gate per batch should be
   able to run the whole battery, not only the annotate verb.
10. CI dogfood job: run the battery over THIS repo in CI
    (`go run ./cmd/docs-gates run .`) — the gates that define green
    currently judge themselves only on dev machines.
11. Golangci version-drift guard: assert the flake's golangci pin and
    ci.yml's action `version:` input move together (a tiny flake check
    or CI script).
12. actionlint + shellcheck into the devshell (workflow YAML and bash
    scripts have no standing validator; tonight's YAML check was a
    one-off python parse).
13. Record the margins DF arithmetic in AGENTS.md as a non-obvious
    fact (window `[bar-2, bar)`; DF counts living-doc lines; the
    corpus-replace bar-2 rationale) — the memory-maintenance miss of
    §d.6.
14. Fleet AGENTS audit for "pushes too"-class daemon beliefs: docs-gates
    carried one for five months; siblings that copy its AGENTS
    phrasing likely carry the same falsehood.
15. Hand the 18:00–19:04 queue-full storm + FastFlowLM timeout window
    to SystemNix monitoring docs (host-level; observed in the pma
    journal, out of this repo's scope).
16. T1: v0.2.0 release cut (owner-gated; the Unreleased section keeps
    growing — five more rows tonight).
17. T7: fleet ratchet corpora re-runs (crm, go-appkit).
18. T8: pbx-artmann migration onto this repo, then fleet announce.
19. T11: pilot ratchet drains in crm + go-appkit.
20. T13: rpi3-dns toplevel eval after the lars-packages change.
21. T14: file/fix the BuildFlow nix-fmt/treefmt roundtrip walker bug
    upstream (it failed again in tonight's 23:25 buildflow run).
22. T15: BuildFlow#87 upstream fix, then drop the
    BUILDFLOW_NO_RESULT_CACHE bypass from the verification ritual.
23. T21: cross-check the fleet docs-health skill text against the new
    flags (crush-config sibling; still pending, deliberately untouched
    tonight).
24. T22: submit `lib.maintainers.LarsArtmann` upstream to nixpkgs.
25. F94: AGENTS.md simplification pass (this file keeps accreting
    non-obvious facts; a periodic compression is owed).
26. Triage buildflow's "9 tools unavailable" via
    `buildflow history --step ... --last-error` (existing backlog row,
    never diagnosed).
27. Rebuild/reinstall the BuildFlow binary (freshness advisory 74a9212
    vs 4211653) and run `buildflow format` once for the fingerprint
    drift.
28. Diagnose the vendor/flake-pin-consistency preflight FAIL
    attribution (resurfaced tonight at 23:25; still observed-never-
    diagnosed).
29. Docs-health skill cross-check against --replace/--margins
    (duplicate pointer of T21 — kept because the battery's routing
    treats it as a distinct routed row).
30. Margins threshold knob: expose `marginNearBar` as config (the
    bar-2 window insight tonight is the concrete argument).
31. `docs-gates doctor` verb (PATH, config, flake state in one probe —
    tonight's live-PATH probe was by hand what a doctor verb would
    automate).
32. Continuation-digit gate: the §f wrapped-line trap (proseRe) that
    burned the 2026-11-04 session remains open on ROADMAP.
33. `--margins --json` for tooling (ROADMAP convergence row).
34. Split-check verb (ROADMAP fuel).
35. CutStringFlag extraction (ROADMAP fuel; CutIntFlag exists, its
    string sibling keeps being re-needed).
36. BuildFlow result-cache identity fix (ROADMAP fuel — the empty
    FileContract cache key class).
37. Findings-gate cache-state summary (ROADMAP fuel).
38. Dependabot enrollment for the action SHA pins (BuildFlow
    `-s dependabot-auto-configure`) — the two new action pins age
    without a renewal mechanism.
39. README "Scripts" section: safe-strike.sh, stack-drift-check.sh and
    check-erraudit-rev.sh are discoverable only through AGENTS.md.
40. Purge ritual: struck-row count reached 17 tonight against the
    20-row cadence threshold — schedule the docs-health purge pass
    before the gate turns it from warning to finding.
41. Verify whether the morning session's corrupt daemon commits
    (`b285980`, `b5ddb5b`, `61be3f3`) are unreachable from every ref
    and can be allowed to age out of local history (housekeeping, no
    urgency).
42. Re-verify the AGENTS "commits ~every 3 min" claim against the
    effective config (min_commit_interval 120s + debounce 60s — the
    claim is roughly right but uncited; cite the config or restate).
43. Consider a `skip_gates` entry or quieter output for the
    erraudit-unknown-linter warning noise in CI logs (cosmetic; the
    warning is expected and documented).
44. The R1 golden names `notes` (the `# NOTES` heading token) as
    near-bar vocabulary — a heading-token exclusion in margins (like
    the one dump-items/item parsing applies) is a candidate behavior
    refinement; assess whether headings SHOULD claim budget.
45. corpus-replace now serves two masters (rewrite-mode and margins);
    document the coupling in the fixture (a fixture README or the
    parity comment) so the next editor knows bar-2 is load-bearing for
    BOTH arms.
46. Keep report-ledger continuation lines away from leading digits in
    THIS report too (the proseRe trap is alive; this §f wraps dozens of
    lines — the battery just validated them, keep it that way on any
    future edit).
47. Consider promoting the annotate `--gate` invocation inside
    safe-strike to also strike-batch mode (`status-items`) so bulk
    sweeps get the same rails (today only single-row strikes are
    scripted).
48. The provider module's `go 1.27` floor vs the root's 1.26: the
    dual-floor story is documented, but a check that FAILS eval when
    the floors drift apart past the pinned toolchain would make the
    invariant mechanical (flake already pins go_1_27 over both; assert
    the relationship).
49. CI: the go job's gofmt step and the new golangci step overlap in
    formatting concerns (gofumpt inside golangci) — assess dropping
    the standalone gofmt step for latency now that lint gates format.
50. This report's §f rows must stay harvested after future rituals —
    when rows land, strike them through safe-strike with the battery
    gate, keeping the newest-report unharvested count at zero.

## g) Questions for the owner (cannot self-answer)

1. Release authority: is the T1 v0.2.0 cut authorized NOW (fold the
   Unreleased section — 30+ rows across five gates of user-visible
   work — into `[v0.2.0]`, annotated tag, push)? Asked five times
   yesterday without an answer; the section has only grown since.
2. Is `autoPush = false` DELIBERATE host policy (privacy, atomicity,
   review-before-push), or an accident of the 2026-06-01 module
   default that five months of phantom-push beliefs grew around? If
   deliberate, tonight's AGENTS correction stands and nothing changes;
   if accidental, the flip is one line in SystemNix
   configuration.nix — but it changes push behavior for the whole
   fleet, so I will not touch it without the ruling.
3. May I push SystemNix master (4 ahead: my wrapper fix riding inside
   a daemon commit, plus another live session's three daemon commits)
   and run the switch to activate the wrapper fix — or does the
   sibling stay on its own session's cadence, in which case the
   wrapper fix remains inert until whoever owns that train pushes and
   switches? (Related: if the 19:05 reboot WAS the deploy window, tell
   me whether the stale `6ed909f` on PATH is expected, so the T2/T3
   rows can be re-verified against reality.)
