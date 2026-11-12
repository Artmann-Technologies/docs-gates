# Status Report — erraudit --max typed refusals + cadence archive ritual

**2026-11-12, 12:30 CET** · session: single train, "superb error handling!" — the user
pasted an `erraudit lint ./... --max` run (exit 2, three `generic_return` findings)
and demanded READ/UNDERSTAND/RESEARCH/REFLECT → execute step-by-step → verify →
repeat until everything works.

## Scope line

**VERIFIED this session (fresh commands, this session):**

- `erraudit lint ./... --max` over the ROOT module: **rc=0, zero findings** (was
  exit 2 with 3 findings at session start). Same invocation over
  `pkg/provider` (separate module): rc=0.
- `erraudit lint ./... --type-aware`: rc=0. `erraudit nolint-audit .`:
  "16 directives: 16 needed, 0 stale".
- `nix develop -c go vet ./...` clean; `go test ./...` 10/10 packages ok;
  `golangci-lint run ./...` 0 issues; `golangci-lint fmt` made zero changes.
- `nix flake check --all-systems --no-build`: "all checks passed!".
- Dogfood battery `timeout 300 go run ./cmd/docs-gates run .`:
  before the archive ritual — freshness/audit-all/annotate/routing PASS,
  cadence FAIL (16 unarchived > 15); after the ritual — **5/5 PASS, rc=0**.
- CI: push of `afe7eab` (typed refusals) ran the dogfood job RED on the same
  pre-existing cadence row (run 43208204565); push of `7db3152` (archive
  ritual) is **GREEN** (run 43208588818, success, both modules + dogfood).
- Predicate research read from erraudit source
  (`~/projects/erraudit/internal/ast/predicates.go`): `hasGenericErrorReturn`
  fires iff bare-`error` result AND body itself calls
  `errors.New`/`errors.Join`/`fmt.Errorf`/`.Wrap[f]`, minus the
  interface-method and `new*`/`wrap*`/`parse*` name exemptions.
- Pre-flight of the CHANGELOG candidate lines: first run REFUSED (exit 2,
  `### Changed` already on disk), rerun with `--duplicates`: **SAFE**, boundary
  tokens named (error→DF13, anchor→8, return→8, refusal→9), zero historical
  item flips to DRIFT.

**NOT verified / not attempted (and why):**

- The two daemon auto-commits `b0dbdee`/`f23b78b` that wrapped my mid-train
  edits were soft-reset away and re-landed as `afe7eab`; I did NOT audit the
  reflog for anything else the daemon may have staged mid-edit (tree was clean
  at each check, and `git log baseline..HEAD --name-only` showed ONLY my five
  files — read once, trusted).
- `pkg/provider`'s test suite was NOT re-run (module untouched; only its
  `--max` audit was re-run).
- NO fresh research into board rows outside this session's work (user
  instruction: report only what this session ran and noticed). Carried rows
  below cite the 07:33 report and TODO_LIST as of last session.
- The new carriers have NO dedicated unit tests yet (see b) — their message
  bytes are pinned only transitively, via the existing suites.

## a) FULLY DONE

1. **`config.Load` typed refusal carrier** (`internal/config/config.go:135`):
   new unexported `loadError` struct — `loadStep` enum (`loadRead`/`loadParse`/
   `loadValidate`, zero value deliberately unrepresentable and loud via the
   default `String()` branch) + path + cause, `Unwrap()` preserving the
   historical `%w` reach. The four inline `fmt.Errorf` sites are gone; FATAL
   text is byte-identical (`reading/parsing/invalid <path>: <cause>`).
2. **`docspurge.insertBullets` typed carrier**
   (`internal/docspurge/docspurge.go:289`): new `missingAnchorError` +
   `anchorScope` enum (`anchorUnreleasedHeading`/`anchorChangedSection`),
   whose `String()` carries the exact anchor phrase so both historical
   refusal messages stay byte-for-byte ("has no ## [Unreleased] heading …" /
   "has no ### Changed section under [Unreleased] …"). The `[Unreleased]`
   fragment is now DERIVED (`strings.TrimPrefix(unreleasedHeading, "## ")`) —
   one source of truth instead of a hardcoded literal.
3. **`docsstrike.validateStrikeInputs` typed carrier**
   (`internal/docsstrike/parse.go:120`): new `strikeInputsError` +
   `strikeRefusal` enum (`refusalMissingRow`/`refusalMissingVerdict`/
   `refusalPrefixedVerdict`) + opener field. All three refusal texts
   byte-identical (the existing `strings.Contains(err.Error(), "verdict word")`
   test still passes untouched).
4. **Design honored the loaded go-error-modernization skill**: idiomatic
   `error` return KEPT (a concrete `*T` return is the typed-nil landmine),
   unexported struct + `newXxxError` constructor, `Unwrap` where a cause
   exists. This extends the file's own `invalidError`/`invalidf` pattern —
   which is exactly why `validate()` was never flagged while `Load` was.
5. **Five orphaned `//nolint:err113` directives removed** (2 in docspurge,
   3 in docsstrike) — nolint-audit stays green: 16 directives, 0 stale.
6. **`erraudit --max` clean over BOTH modules** — the user's exact failing
   invocation (exit 2) now exits 0.
7. **All local gates green**: gofmt-clean without changes, vet clean, root
   suite 10/10, golangci 0 issues (only the documented expected
   `[runner/nolint_filter] erraudit` warning), `--type-aware` rc=0, flake
   check --all-systems all passed.
8. **Docs landed with the full ritual**: CHANGELOG `[Unreleased]` →
   `### Changed` entry (pre-flighted SAFE first, including the
   `--duplicates` rerun for the deliberate second structural heading);
   AGENTS.md Critical-constraints erraudit bullet now records the `--max`
   posture (typed carriers, error-return-kept rationale, constructor-name
   exemptions) and the Non-obvious-facts erraudit-input bullet now names
   `generic_return` as part of erraudit's finding domain.
9. **One deliberate commit `afe7eab`** ("fix: typed refusal errors so the
   strict erraudit audit is clean") — the daemon had wrapped steps 1–2 and 3
   into two auto-commits (`b0dbdee`, `f23b78b`); `git reset --soft` to the
   pinned baseline `040873d` (allowed form) rewound and re-landed the whole
   train as one attributed commit. Pushed; `040873d..afe7eab` on origin.
10. **Cadence archive ritual executed exactly as the gate prescribes**:
    verification leg (freshness + audit-all + routing + annotate all green
    over the corpus — the same leg the existing archived banners document),
    `git mv` of the oldest report
    (`2026-11-04_05-38_art-dupl-dedup-pass-status.md`) into
    `docs/status/archived/` with the `> **Live remainder:**` banner
    (commit `7db3152`), battery re-run: **5/5 PASS** (15 unarchived =
    threshold 15, ARCHIVE-HEALTHY). Pushed; **CI green on main**
    (run 43208588818).
11. This status report, per the status-report skill (md default, scope line,
    daemon pickup — no manual commit per the Crush clause).

## b) PARTIALLY DONE

1. **Typed-refusal test coverage**: the existing suites pin the messages
   transitively (docsstrike's contains-assertion, gatecli's FATAL-line
   expectations, full green suites), but there are NO direct tests yet for
   the new carriers: no `errors.AsType` classification round-trip
   (step/kind/opener fields), no explicit table pinning every `Error()`
   byte, no out-of-range-kind test proving the loud default branches. The
   FATAL contract is enforced only indirectly.
2. **`--max` cleanliness is real but unenforced**: green today, wired
   nowhere — no flake check, no CI leg, no pre-commit leg carries
   `erraudit lint ./... --max`. It stays clean only as long as a human
   remembers to run it manually (this session started BECAUSE the user ran
   it by hand).
3. **The 07:33 report's three §g questions**: still unanswered (release
   claim protocol; v0.2.1 fast-follow vs next MINOR; SURFACED ruling
   priority). Owner-gated; carried forward unchanged.
4. **The 07:33 report's §f ledger harvest**: still owed — that ledger is
   machine-invisible (`## f)` header doesn't match the §f extraction shape,
   so freshness passes vacuously) and I did NOT harvest it this session
   (user instruction scoped this session to the erraudit train).
5. **The `main` branch spent ~3 minutes red on CI** (my first push carried
   the pre-existing cadence red until the follow-up archive commit). End
   state green; the window is owned in d).

## c) NOT STARTED (this session did not touch these)

1. **Standing `--max` gate**: a flake check `erraudit-max` mirroring
   `erraudit-lint`, and/or a devShell command entry in AGENTS' Commands
   block. Not started (needs an owner decision, see g-2).
2. **HARVEST of either report's ledger into TODO_LIST**
   (docs-health HARVEST mode). Not started.
3. **SURFACED rulings** (dependabot PR #1 / D1 / D2 / io-surface
   adoption): owner-gated, untouched.
4. **TODO_LIST row "push erraudit 5c841cf2, then bump input + re-pin
   tripwire"** (TODO_LIST line 54): untouched, blocked on erraudit-repo
   work.
5. **Pre-commit hook arming** (`git config --unset core.hooksPath` so the
   nolint-audit + docs-battery legs auto-install): still undone, still the
   documented one-liner.
6. **v0.2.1 fast-follow** (the v0.2.0 tag's tree carries the stale README
   release line; the fix only lives on main since `b2f1f7f`): decision
   still open.
7. **Classification/byte tests for the new carriers** (b-1): not started.
8. **A second pre-archive** to absorb the cadence red THIS report will
   re-cause (see e-4 / g-1): not started — deliberate, awaiting your
   policy call.

## d) TOTALLY FUCKED UP (all self-inflicted, all caught, none landed)

1. **I pushed a knowingly-red main.** My local battery run BEFORE the first
   push showed cadence FAIL (16 > 15) and I pushed `afe7eab` anyway,
   treating the red as pre-existing — which it was — and fixed it with a
   second commit (`7db3152`) three minutes later. Better: the gate had
   already printed the full archive ritual in my own terminal; I should
   have executed it BEFORE the push and landed one green train. Cost: ~3
   minutes of red CI on main and one wasted CI run. End state identical;
   the process was sloppy and I own it.
2. **The daemon wrapped my deliberate train mid-flight** (`b0dbdee` +
   `f23b78b`) because I batched three file-edits before my first commit.
   The recovery (`git reset --soft` to the pinned baseline + one attributed
   commit) worked cleanly — nothing was lost, nothing foreign was caught
   in the wrapper commits (verified by `--name-only`) — but the situation
   was avoidable: commit each verified step immediately, per the AGENTS
   rule the repo already paid for on 2026-11-04.
3. **Pre-flight first-run refusal** (exit 2, "candidates ALREADY on disk")
   on the `### Changed` heading: friction, not damage — the refusal's own
   message prescribed `--duplicates` and the rerun was SAFE. Minor: I
   could have anticipated that a structural heading collides with the
   lowercased-substring on-disk test and run `--duplicates` from the start.
4. **Nothing else.** No lost work, no corrupted state, no false-strike, no
   reverted foreign edits, no force-push, no history rewrite (soft-reset
   only ever touched unpushed daemon commits, per the allowed form).

## e) WHAT WE SHOULD IMPROVE

1. **Gate before push, not after**: run the dogfood battery and RESOLVE
   owed rituals before every push. "Pre-existing red" is still red main.
2. **Commit per verified step** — never let the auto-commit daemon wrap a
   deliberate train (the 2026-11-04 lesson, re-learned at small scale).
3. **Wire `--max` into a mechanical gate** — a class of finding that only
   exists when a human remembers to run a command WILL regress silently.
4. **Cadence churn is structural**: EVERY status report pushes unarchived
   count past `max_unarchived: 15` again — this very report re-crosses it
   the moment the daemon commits it (15 + 1 = 16 > 15). Either raise the
   threshold, adopt archive-on-report cadence, or accept red-until-ritual
   as deliberate friction. Needs a policy, not another one-off archive.
5. **Pin the new carriers directly**: classification round-trips
   (`errors.AsType`), byte-exact `Error()` tables, out-of-range enum
   behavior. Today's safety is transitive.
6. **§f ledger shape**: the previous report's 50-item ledger is invisible
   to the freshness harvest (`## f)` header). Either normalize report
   templates to the extractable §f shape (and then HARVEST immediately,
   or freshness goes red on 40+ open items) or formally accept that
   ledgers are invisible — but stop letting green freshness over-claim.
7. **FATAL UX opportunities noticed while typing the carriers**: the
   config parse refusal could surface `json.SyntaxError` offset as a
   line:column; the missing-anchor refusal could point at the nearest
   heading line; the strike refusal could print a correct example
   invocation. All cheap, all user-facing polish.
8. **Two status reports in one day inflates the archive corpus** — batch
   one report per train going forward where feasible.
9. **AGENTS.md erraudit bullet is now the longest bullet in Critical
   constraints** — approaching the point where a dedicated subsection
   scans better.
10. **erraudit UX nit (own repo)**: `lint <dir>` prints the
    "promoting directory … to package pattern" notice on every strict
    audit run; quieter default would tidy scripted output.

## f) Up to 50 things we should get done next (BRAINSTORM — HARVEST fuel, not commitments)

Sorted roughly by impact; owner-gated and blocked items marked. Most of
these are ROADMAP fuel, not promises.

1. Execute the cadence ritual for THIS report (it re-crosses 16 > 15 on
   daemon commit): pre-archive one more report or get the policy answer
   (g-1) first.
2. Decide the `max_unarchived` policy: raise threshold vs
   archive-on-report cadence vs keep ritual churn (g-1).
3. Wire `erraudit lint ./... --max` as a standing gate — flake check
   `erraudit-max` (red-control: plant a generic-return, prove red) +
   devShell Commands entry (g-2).
4. Add direct carrier tests: `errors.AsType` classification round-trips
   for `loadError` (step+path+cause), `missingAnchorError` (scope),
   `strikeInputsError` (kind+opener).
5. Add byte-exact `Error()` table tests for all three carriers (pin the
   FATAL contract explicitly, not transitively).
6. HARVEST the 07:33 report's 50-item §f ledger into TODO_LIST (manual —
   header invisible to the extractor).
7. HARVEST THIS report's §f into TODO_LIST (docs-health HARVEST).
8. Normalize the status-report §f header shape so ledgers become
   freshness-visible (extractor vs template — pick ONE definition).
9. Answer §g-1 of 07:33: release-ownership/claim protocol after the
   two-session v0.2.0 tag collision (OWNER).
10. Answer §g-2 of 07:33: v0.2.1 fast-follow for the tag's stale README
    line vs ride to next MINOR (OWNER).
11. Answer §g-3 of 07:33 + g-3 here: which SURFACED ruling first
    (OWNER).
12. Ruling: dependabot PR #1 (owner-gated).
13. Ruling: D1 (owner-gated).
14. Ruling: D2 (owner-gated).
15. Ruling: io-surface adoption (owner-gated).
16. Push erraudit `5c841cf2` upstream, then bump this repo's flake input +
    re-pin `scripts/check-erraudit-rev.sh` expected_rev (run the erraudit
    checks on the new rev first) — TODO_LIST line 54 (blocked on
    erraudit repo).
17. Arm pre-commit hooks locally: `git config --unset core.hooksPath`,
    verify the nolint-audit + docs-battery legs install and fire.
18. Evaluate an erraudit CI leg (needs the binary on runners; private
    input today) — decide public-launcher vs local-only posture, record
    it.
19. Decide whether pkg/provider gets ANY erraudit gate (currently
    nothing pins its clean `--max`).
20. Config FATAL UX: surface `json.SyntaxError` offset as line:column in
    the parse refusal.
21. Purge FATAL UX: missing-anchor refusal names the nearest heading line
    number.
22. Strike FATAL UX: missing `--row`/`--verdict` refusal prints a correct
    example invocation.
23. Add an out-of-range enum test (construct an invalid `loadStep`/
    `anchorScope`/`strikeRefusal` via unsafe cast or internal constructor)
    pinning the loud default branch (tiny, optional).
24. Split the AGENTS.md erraudit bullet into a dedicated subsection
    (scanability).
25. erraudit repo (own): quiet the "promoting directory to package
    pattern" notice under scripted use.
26. erraudit repo (own): consider emitting the live directive count in a
    machine-readable form so the AGENTS count line can be checked, not
    retyped.
27. docs-freshness (own): pre-flight special-case structural repeat
    headings (`### Changed` second copy) so `--duplicates` is implicit
    for them.
28. Keep the release ritual: next tag must cut the now-non-empty
    `[Unreleased]`; nested provider tag rides the same commit.
29. After any HARVEST lands: re-run the dogfood battery BEFORE the commit
    (the docs-battery pre-commit leg would do it if hooks were armed —
    see 17).
30. Record the typed-refusal pattern as the convention for ALL future
    refusals (AGENTS bullet exists; verify the next new refusal follows
    it).
31. Consider `--max --disable generic_return` posture docs: if a genuine
    FP ever appears (identity-callback class), the disable is per-module
    and must be recorded, not silent (skill canon).
32. CI visibility: make a red dogfood row louder (job summary step or
    annotation) so "docs ritual due" is unmistakable in the PR/branch
    view.
33. Process: one status report per train (two same-day reports inflated
    the archive corpus today).
34. Archived-banner pointers: make "owning sections live in TODO_LIST §N
    and ROADMAP.md" specific (§ numbers) at the next archive.
35. TODO_LIST bookkeeping: completed ad-hoc work (this train) had no
    board row — decide whether ad-hoc trains get struck-row records or
    CHANGELOG-only records is the accepted bookkeeping.
36. ROADMAP fuel: the typed-refusal seam could eventually back an
    `explain`/`why` mode that prints structured refusal context (fields,
    not just text).
37. ROADMAP fuel: one typed FATAL seam across gatecli (`FatalUsage` +
    config FATALs + carrier Error()) — large, only if refusal taxonomy
    keeps growing.
38. ROADMAP fuel: config parse refusals could carry the offending JSON
    key (decoded from `json.UnmarshalTypeError.Field`) — structured
    carrier already has the slot shape for it.
39. Verify AGENTS claims after the next erraudit input bump: directive
    count line (16 today) + rev — the docs-health VERIFY pass.
40. Dependabot posture: grouped weekly PRs will keep arriving; the
    ruling in 12 sets the triage template.
41. Keep `golangci-drift` + `go-floors` checks in mind for the next
    golangci action bump (existing checks; no action until then).
42. When CHANGELOG's `[Unreleased]` grows further, re-run pre-flight per
    batch (standing rule; noted so it isn't forgotten at cut time).
43. Provider note for the next release train: provider strict-decoder
    contract — any JSON-face knob means re-pin golden AND provider tests
    together (standing rule, unchanged).
44. Battery-watchdog idea: a synthetic slow-gate test proving the
    half-budget warning + cap-abandon path still fire (the drainer fix's
    guard rails; untested since 2026-11-05).
45. `gatecli.LoadConfig` stays a pure forwarder (exempt); if callers ever
    need config-error classification, the typed seam is `loadError` —
    document that path when it first matters.
46. Consider whether `Load`'s `loadStep` should ALSO distinguish the two
    parse passes (raw-map vs typed decode) — same message today; a
    distinct step would sharpen classification if a consumer ever cares
    (tiny, opinion).
47. Session hygiene: pin the baseline SHA at session start (done today
    after one aborted probe — the aborted `nix eval` probe wasted a
    background shell; avoid speculative nix evals in compound probes).
48. README/FEATURES: no updates were needed this session (internal
    change) — VERIFY on the next docs-health pass that FEATURES still
    matches reality.
49. Next battery run over a foreign repo: keep the size-first discipline
    (`ls docs/status | wc -l` before launching) — standing rule.
50. When the SURFACED rulings (12–15) are decided, strike their TODO_LIST
    rows via `scripts/safe-strike.sh --battery`, one batch each.

## g) Questions I can NOT figure out myself

1. **Cadence policy**: every status report re-crosses `max_unarchived: 15`
   (this report does it again on daemon commit). Do you want the threshold
   raised, an archive-on-report cadence, or is the ritual churn
   deliberate friction worth keeping? I can implement any of the three;
   the intended friction level is a preference I cannot derive.
2. **`--max` enforcement posture**: now that `erraudit lint ./... --max`
   is clean over both modules, should it become a standing flake check
   (mechanical enforcement — but note `--max` bypasses EVERY
   `//nolint:erraudit`, so it also forbids future suppressions of this
   class), or stay a manual strict audit? Enforcement constrains future
   code; that's a policy call, not a technical one.
3. **Carried from 07:33 §g-3**: which SURFACED ruling do you want first —
   dependabot PR #1, D1, D2, or io-surface adoption? All four are
   owner-gated; I cannot rank your priorities for you.
