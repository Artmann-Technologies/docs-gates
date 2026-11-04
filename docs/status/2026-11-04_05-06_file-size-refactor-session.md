# Status Report — file-size refactor session (2026-11-04 05:06)

**Scope line.** VERIFIED this session, with commands: the five oversized-file
findings eliminated (`BUILDFLOW_NO_RESULT_CACHE=1 buildflow -s
file-size-check --format finding` → 0 findings); declaration-losslessness of
the two big splits (grep-diff of top-level declaration names against the true
pre-split commits); root module build/vet/test (`go build ./... && go vet
./... && go test ./... -count=1` → 8/8 ok); provider module tests (`go test
./...` from `pkg/provider`); lint (`nix develop -c golangci-lint run ./...` →
0 issues); gofmt (`gofmt -l .` → clean); erraudit (`erraudit lint ./...
--type-aware` rc=0; `nolint-audit .` → 15 directives, 15 needed, 0 stale);
dogfood battery (`go run ./cmd/docs-gates run .` → all gates PASS, run twice,
incl. after the AGENTS.md edit). NOT verified: `nix build .#docs-gates`
(checkPhase in the nix sandbox), `nix flake check --all-systems --no-build`,
`nix develop -c golangci-lint fmt ./...` (golines pass — gofmt+lint clean
makes a drift unlikely but it was not run), the CI run after the daemon's
push, and the byte-exactness of moved code beyond the declaration-set and
compiler/test/parity evidence. Those gaps are listed in (f).

## a) FULLY DONE

1. **All 5 file-size findings resolved by pure file-splitting** — zero
   behavior change; gate now reports 0 findings (verified with the cache
   bypass; the plain re-run serves a stale cached verdict, see (e)/#87).
   - `internal/docsfreshness/docsfreshness.go` (1203 → 9 files: CLI entry,
     `reports.go`, `tokens.go`, `check.go`, `corpus.go`, `audit.go`,
     `explain.go`, `preflight.go`, `preflight_report.go`; 97–173 lines).
   - `internal/docsannotate/docsannotate.go` (853 → CLI entry + `items.go`,
     `claims.go`, `git.go`, `spacing.go`; 51–283 lines).
   - `internal/docsfreshness/docsfreshness_test.go` (754 → 6 files mirroring
     the production split: `tokens_test.go`, `check_test.go`,
     `explain_test.go`, `preflight_test.go`, `preflight_bind_test.go`,
     `run_test.go`; 72–212 lines).
   - `internal/docsrouting/docsrouting.go` (371 → 320 + `settings.go` 57,
     mirroring docsfreshness's settings layout).
   - `internal/docsstrike/docsstrike.go` (355 → CLI entry 134 + `parse.go`
     153 + `strike.go` 81).
2. **Losslessness proven for the split**: declaration-name set identical to
   the pre-split commits (`docsannotate` checked against `c8afcb7`, the true
   pre-split commit, after the daemon moved HEAD mid-check); compiler
   resolves every symbol; test inventory 26 = 23 split-file tests + 3 in
   `settings_test.go`, all pass; parity goldens untouched and green.
3. **Quality loop clean**: golangci-lint 0 issues, gofmt clean, erraudit
   green with all 15 `//nolint:erraudit` directives live and 0 stale, dogfood
   battery PASS twice, provider module green.
4. **BuildFlow result-cache defect: root-caused at source level, filed,
   documented.** `file-size-check` declares an empty `FileContract`, and the
   result-cache key hashes only `tool.Files.Inputs`
   (`execution/pipeline_builder.go:213` → `execution/result_cache_1.go:59`),
   so its cache key is constant — the gate replayed pre-split findings for
   files that no longer existed. Filed upstream as
   [LarsArtmann/BuildFlow#87](https://github.com/LarsArtmann/BuildFlow/issues/87)
   (verified landed), voice-checked draft at BuildFlow
   `docs/drafts/file-size-check-empty-filecontract-constant-cache-key.md`
   (checker: 0 FAIL, 0 WARN), known-artifact entry with the
   `BUILDFLOW_NO_RESULT_CACHE=1` workaround recorded in this repo's
   AGENTS.md ("Non-obvious facts").
5. **AGENTS.md updated** with the known tool bug + workaround (committed by
   the daemon, 955efb9/fa458f8).
6. **Largest .go file repo-wide is now 326 lines** (docscadence,
   pre-existing, under the 350 gate).

## b) PARTIALLY DONE

1. **Loop-closure on the cache defect** — diagnosis + issue + consumer-repo
   documentation are done; the actual FIX upstream is not (BuildFlow had a
   concurrent session active, HEAD moved 74a9212 → 4211653 mid-session, so I
   deliberately did not patch it there). Workaround documented instead.
2. **`buildflow --build-mode dev`** — every step green except the two
   documented known-bug steps (`nix-fmt`/`treefmt` roundtrip walker error,
   pre-existing, AGENTS.md says do not fix here). The loop as a whole
   therefore does not exit green.
3. **Session learnings in memory** — AGENTS.md carries the cache-defect
   entry, but the baseline-pinning lesson (pin the pre-edit commit BEFORE
   splitting, because the auto-commit daemon moves HEAD under you) is only
   in this report, not yet distilled into AGENTS.md or a lessons file.

## c) NOT STARTED

1. **HARVEST** — section (f) below is not yet distilled into
   `TODO_LIST.md`/`ROADMAP.md` (this report is its input).
2. **CHANGELOG row** for the file-size refactor (the repo dogfoods its own
   gates; a structural change of this size deserves a CHANGELOG line).
3. **nix-side verification**: `nix build .#docs-gates`, `nix flake check
   --all-systems --no-build`, and the `golangci-lint fmt` (golines) pass.
4. **CI confirmation** after the daemon's push (root-module CI: build/vet/
   gofmt/stack-drift/test).
5. **Triage of the four preflight WARN/FAIL signals** buildflow surfaced
   every run this session (see (f) items 6–9) — noticed, not addressed.
6. **Upstream fix for BuildFlow#87** and, once fixed, dropping the bypass
   from the verification ritual.

## d) TOTALLY FUCKED UP

Nothing. No data loss, no broken gates, no left-behind breakage; the two
near-misses were self-caught within seconds: (1) the first docsfreshness
write batch left `runExplainItem` unplaced (gopls flagged it before the
build), (2) the first losslessness check for docsannotate compared against a
HEAD the daemon had already moved — recovered by diffing against the true
pre-split commit `c8afcb7`. Both are lessons, not damage: see (e).

## e) WHAT WE SHOULD IMPROVE

1. **Pin the verification baseline before editing** in repos with an
   auto-commit daemon: record `git rev-parse HEAD` first, then diff against
   that pin, never against a moving HEAD. (Paid for this session: one
   misdirected check.)
2. **Split-then-verify ordering**: writing replacement files in batches
   guarantees transient "undefined symbol" noise; writing the leaf files
   first and the referencing file last (or one `go build` between batches)
   keeps every intermediate state compileable.
3. **Stale tool caches in verification rituals**: any buildflow single-step
   verification over a mutated tree should default to
   `BUILDFLOW_NO_RESULT_CACHE=1` until #87 is fixed — document that next to
   the command, not just in AGENTS.md prose.
4. **Preflight signals should not scroll past**: the same four WARN/FAIL
   lines appeared on every run this session (flake-pin disagreement,
   canonical-helper reimplementation, formatter fingerprint drift, binary
   freshness). They were all out of scope for this task, but "noticed every
   run, triaged never" is how known artifacts rot — they are now itemized
   in (f) with owners.
5. **`canonical-helper-reimplementation` on `gatecli.Truncate`** — the
   preflight says gatecli re-implements a canonical helper from `constants`;
   even if it is a false positive, it deserves a one-line AGENTS.md verdict
   (deliberate non-fix with rationale) so future sessions stop re-reading it.

## f) Up to 50 things to get done next (brainstorm, not commitment)

Immediate, this session's loose ends:

1. Fix BuildFlow#87 upstream (`NewFileSizeCheckProvider` input contract, or
   hash the language glob when `Trigger.Files` is nil) — then re-verify this
   repo's gate WITHOUT the env bypass and drop the workaround from AGENTS.md.
2. HARVEST this report: land items worth tracking into `TODO_LIST.md`, the
   rest into `ROADMAP.md` (docs-health HARVEST mode).
3. Add a CHANGELOG row for the file-size refactor (26-file reorganization,
   no behavior change).
4. Run `nix build .#docs-gates -o /tmp/dg-result` to prove the nix sandbox
   checkPhase still passes over the split packages.
5. Run `nix develop -c golangci-lint fmt ./...` (golines/gofumpt/gci) and
   confirm a zero-diff `git status`.
6. Reconcile the `nixpkgs_2` flake.nix/flake.lock disagreement the preflight
   FAIL names (input locked at ref nixos-unstable, no matching `url` line in
   flake.nix) — or record it as deliberate.
7. Triage `quality/canonical-helper-reimplementation` on `gatecli.Truncate`:
   fix by importing the canonical helper, or document as a deliberate
   non-fix with rationale.
8. Clear `quality/formatter-fingerprint-drift` (templ-fmt identity drift) by
   running `buildflow format` once so cached verdicts re-verify.
9. <!-- fresh:skip: sibling-repo work (BuildFlow binary rebuild lives in that repo, not this board) --> Rebuild/reinstall the BuildFlow binary (`nix build . &&
   .#reinstall`) to clear the binary-freshness advisory (74a9212 vs 4211653) —
   coordinate with the concurrent BuildFlow session first.
10. Distill the baseline-pinning lesson (e.1) into AGENTS.md "Non-obvious
    facts" or the global lessons file.
11. Confirm CI green on GitHub after the daemon's push (root-module
    workflow), since provider regressions only surface locally.
12. `nix flake check --all-systems --no-build` once, to re-prove both-arch
    eval after the new files.
13. Verify `git status` is clean (daemon picked up everything; it has a
    known blind spot).

Structural follow-ons in docs-gates (same refactor family):

14. `internal/docscadence/docscadence.go` is 326 lines — the next file that
    will trip the 350 gate; decide now whether to pre-split it along the
    same seams or let the gate force it later.
15. Give `docsannotate` a settings.go-style projection (it consumes
    `config.Config` directly in `Run`; docsfreshness and docsrouting both
    have the projection) — consistency + smaller Run.
16. `internal/battery/battery.go` (150) and `gatecli.go` (217) are fine, but
    `docsannotate_test.go` (146) + `config_test.go` (61) could mirror the
    production split if the fleet convention is file-per-subject.
17. Consider a `//nolint` inventory audit: erraudit's 15 directives are
    honest, but nothing documents WHERE each lives after the split — a
    one-table AGENTS.md appendix would make the next split cheaper.
18. `internal/docsrouting` still imports `docsfreshness` for Tokens/Squash/
    MinTokenLen/GenericDF — fine today, but if the fleet wants gates
    independently versionable, a shared `internal/tokencorpus` package would
    decouple them.
19. The parity fixture corpus (`testdata/parity/corpus`) exercises the
    gates through `Run`; after the split, consider adding one fixture that
    exercises a `--pre-flight` stdin path for byte-level parity of that mode
    too (it is currently covered by unit tests, not goldens).

Docs/ritual hygiene:

20. Strike landed rows in older status reports per the ritual (the battery
    PASSed, but the annotate gate only checks the newest report strictly).
21. Add the `BUILDFLOW_NO_RESULT_CACHE=1` caveat inline next to the
    `buildflow -s file-size-check` line wherever the fleet runbooks mention
    single-step verification (the fleet SKILL.md files, not just this
    AGENTS.md).
22. Update `docs-gates` README if it names any of the five pre-split
    filenames (check: README.md, FEATURES.md — not verified this session).
23. File the `nix-fmt`/`treefmt` roundtrip walker bug as a BuildFlow issue
    if it is not already filed (AGENTS.md documents the diagnosis but I did
    not check for an existing issue number).

BuildFlow upstream (separate repo, coordinate with the concurrent session):

24. Consider a generic guard: warn (or fail doctor) when a detector declares
    `Trigger.Files: nil` AND an empty `FileContract` — the class condition
    that produced #87.
25. Extend #87's fix to audit for OTHER providers with empty input
    contracts walking the tree themselves (`file_counts`,
    `filesystem_check`, `lockfile_check` are candidates worth a grep).
26. Result-cache key could include the repo identity — same class as open
    issue #67 (cross-repo leakage).
27. Make the findings-gate summary name the result-cache state (hit/miss)
    per tool so stale-verdict symptoms are visible without `-v`.

Same-class hygiene noticed but unaddressed (from run outputs):

28. The preflight `env/goexperiment-jsonv2` WARN: GOEXPERIMENT=jsonv2 is set
    in the shell but this repo does not import encoding/json/v2 — harmless,
    but a `.buildflow.yml` `env:` note or doctor whitelist would silence it.
29. `docs/status/` currently holds 3 unarchived reports (cadence says
    PURGE-HEALTHY at threshold 15) — no action needed; listed so the next
    ritual does not re-derive it.
30. Check whether `pkg/provider`'s verdict-matrix parser needs awareness of
    new gate package layout — verified green this session, but its
    FINDING-parsing tests exec the binary; one end-to-end `docs-gates run`
    through the provider path (not just unit tests) would close the loop.

Longer-shot ideas (ROADMAP fuel, do not schedule):

31. File-size gate could ignore `*_test.go` at a higher threshold (test
    files legitimately run longer) instead of one 350 bar for all.
32. File-size gate findings could carry `fixStrategy: split` with a hint of
    the LSP-symbol seams (the split points this session chose manually).
33. A `docs-gates split-check` verb that dry-runs "would splitting this file
    change any golden byte?" using the parity harness.
34. gopls/LSP diagnostics in the harness reported ghost errors for the
    deleted test file until restart — worth an issue against the harness
    LSP integration if it reproduces.
35. Fleet-wide: search other repos for >350-line Go files proactively
    instead of waiting for their next buildflow run.

(Deliberately stopped at 35 — items beyond that would be padding; the
first 13 are the real queue.)

## g) Questions I cannot figure out myself

1. **BuildFlow#87**: the BuildFlow repo has a concurrent session (HEAD moved
   74a9212 → 4211653 during this task). Should I implement the fix there
   now, or leave the whole file to the other session to avoid a collision?
2. **`nixpkgs_2` flake-pin disagreement**: is the flake.nix/flake.lock
   mismatch (preflight FAIL on every run) a deliberate mid-edit state from
   another session, or do you want me to reconcile it here?
3. **Scope of the harvest**: should items 6–9 + 14–19 land in
   `TODO_LIST.md` now (they are triaged but unowned), or do you want to
   pick winners first and keep TODO_LIST to what you actually intend to do?
