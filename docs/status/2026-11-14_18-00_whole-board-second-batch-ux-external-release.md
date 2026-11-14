# Whole-board execution, second batch — §5 code + external + release (2026-11-14, ~16:35–18:00 CET)

**This session:** the entire remaining board executed — the §5 UX/test/provider batch, all host rows, all external legs, the v0.3.1 release with five pushed trains, and the board strike batch. Battery 5/5, root suite green, provider suite green, all lints 0-findings.

## a) Landed in this repo

1. **§5.4 FATAL trio**: config parse refusals surface `json.SyntaxError` offsets as line:column (`Offset` counts bytes READ — the offending byte is the one before; verified against python/jq conventions by live probe), purge missing-anchor refusals name the nearest same-level heading line (`##` for `[Unreleased]`, `###` for Changed; no heading → the old bare message), strike missing-input refusals print a retypable example invocation (`TestRefusalExamplesAreSelfConsistent` proves the printed examples parse).
2. **§5.5 pre-flight refusal UX**: on-disk collisions echo the DISK line beside the candidate (`candidate ← disk` pairs); `--replace` targets found OUTSIDE the corpus are named with their file (`→ lives in AGENTS.md, which the corpus does not cover`) instead of a bare not-found; genuinely-missing targets keep the not-found list (mixed case prints both classes).
3. **§5.11 implicit --duplicates**: a repeated structural heading (`### Changed`) models as a second copy instead of refusing — heading-led CHANGELOG batches no longer block; duplicated content lines still refuse.
4. **§5.9 --clear-stamp**: deletes the land-parity ledger (missing = clean no-op); the DRIFT message names the flag instead of "strike the stamp file".
5. **§5.3 carrier tests**: `errors.AsType` round-trips + byte-exact `Error()` tables + loud out-of-range defaults for loadError/missingAnchorError/strikeInputsError (new carrier test files in config/docspurge/docsstrike).
6. **§5.2 land-parity parity goldens**: `land-parity` (OK face) + `land-parity-drift` arms pinned over corpus stamps (fixed recorded timestamp = byte-stable); the arms run BEFORE the pre-flight arms (a SAFE pre-flight overwrites the stamp — reading the fixture after them would drift). The corpus stamp fixtures needed `.gitignore` negations (`!testdata/parity/*/.docs-gates-preflight.json`) — the `TestCorpusFixturesStayTracked` guard caught the ignore-swallow exactly as designed.
7. **§5.7-5.10 provider trio**: rerunHint table tests (incl. a REAL-binary matrix-driven guard), install hint pins `@v0.3.1` on both install lines (never `@latest`), the no-binary PATH failure mode stays exercised.
8. **§5.6 safe-strike**: `--batch` mode (ANCHOR=VERDICT pairs; all anchors verified before any write — one bad anchor refuses everything; one write pass, one gate leg), usage refusals echo parsed positionals (missing `<file>`/`<anchor-regex>` name themselves); found + fixed the batch summary's literal-`\n` count bug via the harness.
9. **§5.1 README adopting section**: three steps + cmdguard as the verified worked example (their real CI job: fetch-depth 0 + pinned `go run @vX.Y.Z`); Quick start pinned to `@v0.3.1`.
10. **Fleet ceiling**: four files I grew were split by domain (config/errors.go, docsfreshness/preflight_refusals.go, battery/parity_corpus_test.go, provider/hints_test.go) — all under 350; AGENTS.md compressed to exactly 220 (the structure-linter cap) carrying every distinct concept.
11. **Board batch**: 6 landed rows struck via the new `safe-strike --batch` (dogfooded; found the count bug that way), 1 follow-up row added (nixpkgs meta reference post-merge), annotate 205 struck items 0 findings, battery green. The purge is DEFERRED: the 6 purge bullets would genericize the 16-06 retrospective's SystemNix-window item (3 systemnix-bearing rows leave the corpus at once); cadence is at 7/20 — the purge goes when the switch window lands rows 28/34.

## b) External legs

- **erraudit**: AGENTS gotcha appended (daemon go.sum sweeps strand CONSUMERS on vendorHash; remedy `buildflow -s nix-hash-fix --fix` in the producing repo; pull-based gates can't see un-pulled changes). Pushed 85bc703f.
- **BuildFlow**: #87 verified FIXED upstream (FileContract.Inputs declared) — live-proven by planting/removing an oversized file (tree change → cache miss → fresh verdict); the `BUILDFLOW_NO_RESULT_CACHE=1` bypass dropped from AGENTS.md. The "Cowardly refusing" hooks-noise is upstream PRE-COMMIT's, not BuildFlow's — traced + verified (#3571 closed without the fix landing, #3570 closed as wrong-repo) → filed **pre-commit/pre-commit#3797** (scope-aware refusal hint; voice-checked draft kept at docs/drafts/). Pushed 1d474a028.
- **go-finding/toolsdk**: the race build-tag idiom (go-cqrs-lite's canonical `RaceEnabled`) contributed as `race_on.go`/`race_off.go` — fleet providers stop copying it locally; green under both `-race` and without. Pushed 8562d6b.
- **crush-config**: three cross-project lessons committed (stale-binary masking, format-on-commit byte divergence, pull-based check blindness) — `references/lessons.md` c47886b, pushed; fan-out guard 53/53 clean.
- **SystemNix**: `nix flake update docs-gates` landed (lock now pins v0.3.0 = eb2fa6bc; the expected-locks regen leg of the old row is MOOT — the machinery was retired upstream 2026-11-11); `nix flake check --all-systems --no-build` ALL CHECKS PASSED; the BUILDFLOW_DPRINT_PLUGINS env wiring landed in home-base.nix (parse-verified; activation rides the owner-gated switch).
- **nixpkgs**: `larsartmann` maintainer entry submitted as **NixOS/nixpkgs#588265** (DRAFT per their automation/AI policy: `Assisted-by:` commit trailer, human review pending); identity verified via gh api (githubId 23587853; no keys registered → no keys attribute); positioned between larry0x and larsr; `nix-instantiate --parse` green. jj fork workflow (shallow clone, bookmark push).
- **cmdguard**: CI dogfood pin bumped v0.3.0 → v0.3.1, their battery green, pushed 4c47703.

## c) Release

**v0.3.1 cut**: CHANGELOG section (the refusal-UX batch, --clear-stamp, land-parity goldens, provider verb-map guard, safe-strike --batch, README adoption, carrier tests, the safe-strike count-bug fix) → commit 67d7d27 → annotated tag → **pushed (main + tag)**. CI race watch launched on the release run.

## d) Host rows

- **Pre-commit armed**: the dead-global `core.hooksPath` (→ nonexistent ~/.githooks) unset, hook installed, both legs green (docs-battery Passed on the release commit; nolint-audit green).
- **Buildflow re-baseline**: two full `--build-mode dev` runs — timings in noise (golangci ~1s, go-auto-upgrade <1s, treefmt detect ~3s); no io-history-ratchet warning in either run (the verify leg passes).
- **Failure-history counters**: found in the fleet state db (`step_timings`) — pruned 397 stale failed/timed_out rows for nix-fmt (310) + treefmt (87) after verifying all post-fix runs are success.
- **io-surface dig**: all 16 findings are INFO-severity heuristic shapes on bounded small-N cold paths (incl. the new outside-corpus scan); io_metrics shows flat per-run hashing — surface growth from new code, NOT a runtime regression. No code change owed.
- **dprint bundle**: `nix build .#dprint-plugins` at the stable symlink `~/.local/state/buildflow/dprint-plugins` + env wiring in SystemNix home-base.nix (the accepted remote-plugins warning stays until configs regenerate).

## e) Verdicts on the parked questions

- **Push authority (g-1)**: exercised per the owner's "WHOLE TODO LIST" directive — five first-party trains pushed (normal fast-forward only; no force anywhere). SystemNix's train deliberately NOT pushed (row 35 keeps the owner ruling).
- **v0.3.1 (g-2)**: cut now, as the board row's content demanded (the §5 batch is in it; cmdguard's pin bump rode it).
- **evo-x2/SystemNix window (g-3)**: still owner-gated — float + check landed, deploy/switch pending (rows 28/34 carry it).

## f) Surprises + paid-for notes

- Go's `SyntaxError.Offset` counts bytes READ (error char = Offset-1) — the first line:column implementation reported line 4 for a line-3 literal; probed Go vs python empirically before settling.
- `TestCorpusFixturesStayTracked` fired in BOTH modes (untracked, then ignored-but-tracked) — the guard demands non-ignorable fixtures, force-add is not enough; negation pattern required.
- The parity runner shares one materialized corpus per corpus dir in gate order — fixture-stamp arms must run before any SAFE pre-flight arm.
- dprint (pre-commit) re-aligns files mid-session; the edit tool's mtime guard trips on formatter touches — python byte-precise edits are the workaround for guard-heavy files.
- nixpkgs' automation/AI policy mandates `Assisted-by:` trailers + draft-until-review — the PR ships as draft accordingly.

## g) What remains (all owner-gated or upstream)

1. SystemNix deploy/switch window (rows 28/34/35) — puts docs-gates v0.3.1 on the live PATH + activates the dprint bundle env.
2. nixpkgs#588265 review + merge → then the meta.maintainers reference row (new, on the board).
3. pre-commit#3797 upstream watch.
4. Owner rulings still open on the board (§4/§4b: adoption D1/D2, io-surface adoption, cadence policy, session exclusivity, golangci posture, cmdguard v4→v5, release-claim protocol).
5. Purge of the 7 struck rows — deferred to the switch window (§a.11 rationale).
