# Session status — plan v2 execution: 11 tasks landed, 1 corruption caught and reverted, CI green

**2026-11-04 23:06 CET** · baseline `d6d1702` (pushed, CI run 41920240341
success) · scope: this session only — the plan-v2 zero-permission
execution + this report.

**Verified here:** battery 5/5 PASS rc=0 (`go run ./cmd/docs-gates run .`,
re-run after every doc edit); root suite 8/8 ok (`go test ./... -count=1`);
provider suite ok; golangci 0 issues on BOTH modules; `nix flake check
--all-systems --no-build` all-pass; the four erraudit-era checks built
green individually (`nolint-audit`, `erraudit-lint`, `erraudit-rev`,
`pre-commit-check`); `nix build .#default` yields all six binaries with
stamped `--version` under both names; CI success read back via `gh run
list` for `d6d1702`; red controls run LIVE for every new gate (gitignore
both legs, e2e sentinel decision table, planted-stale pre-commit hook,
rewrite-mode SAFE/UNSAFE discrimination); corruption revert `ef14e11`
verified by annotate CLEAN (13 struck, 0 findings).

**NOT verified here:** anything in sibling repos (SystemNix lock state,
BuildFlow board, crm/go-appkit corpora) — out of scope, owner-gated; the
daemon's push-lag CAUSE (22 commits accumulated before my manual push —
observed, not diagnosed); the nix-sandbox "skipped" arm of the e2e
sentinel (the nix check passes but its non-verbose log does not prove
WHICH arm fired — the decision table proves the mapping instead); CI log
contents beyond the success badge; ROADMAP items (untouched by design).

## a) FULLY DONE

1. **T4 docs-surface train** — README freshness-diagnostics section
   (`--margins`/`--dump-items`/`--pre-flight`+`--replace`) + per-verb
   exit-code table (battery's gate-rc2-to-FAIL nuance included);
   DOMAIN_LANGUAGE +6 terms (rewrite-mode, margins, dump-items, harvest
   attribution, corpus, settings-projection) and the pre-flight row
   corrected to "report corpus"; FEATURES freshness/flake/provider rows.
   Evidence: daemon `88d5a9f` + `8a6435f`.
2. **T5 audit-all attribution parity** — the replay renderer now prints
   the same matched-token/example-line attributions as the newest-report
   leg; battery corpus arm `TestRunAuditAllAttributions` asserts the
   lines (red run output rides stderr). Evidence: commit `d92f7f8`.
3. **T6 e2e silent-skip visibility** — ledger + sentinel
   (`TestE2ELegNeverVanishesSilently`) + 7-case decision table
   (`TestE2ELegVerdictTable`); only the two sanctioned sandbox reasons
   can skip; live proof "E2E LEG: ran" locally, nix provider check green
   with the sanctioned skip. Evidence: commit `ff7cfcd`.
4. **T9 gitignore-trap regression arm** — `TestCorpusFixturesStayTracked`
   with two legs: tracked (`git ls-files`) + unignored (`check-ignore
   --no-index`); both legs red-controlled live. It earned its keep
   immediately by catching T10's unstaged fixtures. Evidence: commit
   `94350f5`.
5. **T10 rewrite-mode parity golden** — `corpus-replace` fixture + pinned
   UNSAFE bytes (`preflight-replace.txt`: WOULD-DRIFT naming the vanished
   anchors) + discrimination arm proving the same candidate is SAFE
   without `--replace`; harness gained per-gate `wantCode`. Evidence:
   commit `510ff58`.
6. **T12 pre-commit nolint-audit leg** — git-hooks.nix input +
   `preCommitCheckFor` shared by the sandboxed `pre-commit-check` flake
   check and the devShell hook install; red-controlled live (planted
   stale directive → "13 needed, 1 stale" FAIL). Evidence: daemon
   `ff462df` + `87d34fe`.
7. **T16 goldens batch** — goldens 9→12 (`margins`,
   `dump-items`, `dump-items-refusal`); `--dump-items` gained the
   struck-row arm (WHY a ~~row~~ is exempt); corpus carries a dated-DONE
   struck row + a struck report item (cadence/annotate/battery goldens
   re-pinned by hand-diffed, exactly-described diffs); harness gained
   `argsAfterRoot` + corpus-root normalization. Evidence: daemon
   `7a16061`.
8. **T18 margins unclaimed-vocabulary scoping** — open-item tokens of
   the newest report are excluded from the budget report; struck and
   fresh:skip items stay unclaimed; two tests cover both arms; dogfood
   intersection check empty (no claimed token leaks). Evidence: daemon
   `3e23d60`.
9. **T19 routing projection export** — `SettingsFromConfig` exported
   like annotate/freshness/cadence; same zero-values-fallback contract.
   Evidence: commit `4777163`.
10. **T20 LSP ghost root-caused, NOT filed** — the `golangci_lint_ls`
    stale `undefined: <renamed symbol>` survives `lsp_restart`; clearing
    the host wrapper's persistent `GOLANGCI_LINT_ANALYSIS_CACHE` cured
    it → root cause is LOCAL (wrapper cache), the upstream issue was
    draft-ready and REFUTED (verify-before-filing Gate 4). AGENTS.md
    carries the protocol. Evidence: the cold-cache experiment in-session.
11. **T23 + F93 decision preps** — `docs/planning/2026-11-04_21-02_owner-decision-preps-D1-D2.md`:
    D1 PATH-only-vs-timer matrix (7 axes) with the PATH-only-for-a-quarter
    recommendation; D2 floor-1.27 recommendation with tradeoffs. Both §4
    TODO rows now point at it. Evidence: commit `d6d1702`.
12. **Session status report** — `docs/status/2026-11-04_22-52_plan-v2-execution-session.md`,
    committed + pushed, CI green; its §f routed to existing rows (audit-all
    green is the mechanical proof).
13. **Four new AGENTS non-obvious facts** — variadic-spread compile rule
    (proven on 1.26.8 AND 1.27.3), check-ignore tracked-blindness +
    ls-files cwd-relative output, LSP-ghost cache root cause, strike-sweep
    corruption machine (with the safe mechanics).
14. **Ritual discipline held** — CHANGELOG row per landed change, TODO
    strikes with verdict+evidence per landed row, battery green after
    every doc edit, lint at edit time.

## b) PARTIALLY DONE

1. **T17 ANNOTATE sweep** — attempted, corrupted three report ledgers,
   reverted byte-identical (`ef14e11`), scoped down. What works: the
   revert (annotate CLEAN), the honest re-scope on the row (ledgers hold
   28-49 open rows each; no report fully-resolvable), the safe mechanics
   documented in AGENTS.md. What remains: the actual per-row verdict
   strikes on ~70 demonstrably-landed rows across 02-24..05-38, and the
   archiving clause (nothing qualifies yet). Blocker: none but scope —
   the false-strike risk under the daemon cadence is the reason.
   Effort to finish: M per report, on demand.
2. **T24 D2 floor decision** — F93 (recommendation note) done; F94 (the
   AGENTS one-floor-sentence simplification) is owner-gated on the
   decision itself. Effort: S once decided.
3. **TODO_LIST harvest of THIS report's §f** — the 22:52 report's items
   routed to existing rows; the NEW candidates surfaced by tonight's
   self-review (see §f items marked [NEW]) are being harvested now, at
   the end of this report. Effort: S.
4. **Daemon push health** — 22 commits accumulated unpushed; I pushed
   manually once. The push itself works; WHY the daemon lagged is
   undiagnosed (not attempted — out of session scope). Effort: S to
   diagnose, in crush-config territory.

## c) NOT STARTED (all owner-gated or sibling-scoped; priority per plan v2)

1. **T1 v0.2.0 release cut** — waits on release authority (asked 4×).
   Still wanted: yes, it versions 27+ Unreleased rows and the three new
   flags.
2. **T2 SystemNix float** — sibling permission; mechanical once granted.
3. **T3 evo-x2 deploy + live smoke** — owner quiescent window.
4. **T7 fleet ratchet corpora re-run** — sibling; MORE urgent now (the
   Corpus semantics converged again today: dated-DONE classify, margins
   scoping, routing glob prep).
5. **T8 pbx-artmann migration** — sibling; rides T1's tag.
6. **T11 pilot ratchet drain** — sibling.
7. **T13 SystemNix eval legs** — sibling; rides T2.
8. **T14 BuildFlow triage batch** (9-tools history, preflight
   attribution, binary rebuild, `buildflow format` once) — sibling.
9. **T15 BuildFlow upstream** (#87 + roundtrip walker, then drop the
   AGENTS bypasses) — sibling.
10. **T21 skill cross-check in crush-config** — sibling; now ALSO the
    place where the LSP-wrapper cache fix would land (see §f).
11. **T22 nixpkgs maintainers PR** — owner-gated external PR.
12. **F94 AGENTS toolchain paragraph simplification** — rides the D2
    decision.
13. **ROADMAP fuel items** (doctor verb, margins threshold knob,
    split-check verb, continuation-digit gate, CutStringFlag,
    result-cache identity, findings-gate cache-state summary) — not
    started by intent; they are ideas, not commitments.

## d) TOTALLY FUCKED UP

1. **The strike sweep corrupted three status reports.** Anatomy, in
   order of blame:
   - **The inventory was truncated.** My first dump piped `head -12` and
     I believed each §f ledger held ~12 open items. Reality: 28-49 each.
     The whole sweep was sized on wrong data — the plan was doomed before
     the first sed ran.
   - **I mapped dump-item ordinals to table row numbers 1:1.** They are
     different numbering spaces; for 03-34/04-00 my first batch struck
     the WRONG rows (open work struck as "done", landed rows missed).
   - **The sed anchor matched the wrong table.** The §f ledgers share
     their `| N |` row shape with board-mirror tables in the same file;
     row-number patterns hit both.
   - **Greedy capture collapsed cells.** `(.*) (\|)$` swallowed every
     middle cell into the struck text — table structure destroyed, not
     just strikes added.
   - **I applied first and gated after.** My own T9/T6 protocol is
     red-control BEFORE trusting — inverted here. The annotate gate's 31
     findings was the only thing that caught it.
   - **The daemon committed the corrupt intermediates** (`b285980`,
     `b5ddb5b`, `61be3f3`) before I noticed — corrupt states entered
     local history. Recovery was a forward `git restore --source=` (the
     sanctioned shape; never a history rewrite) and a push carrying the
     fix (`ef14e11`).
   - Damage: contained (annotate CLEAN, battery green, origin healed),
     but 3 bad commits exist in local history and one manual push
     carried both the corruption and its fix.
2. **Smaller burns, same root pattern (predict-then-discover):**
   - First `--replace` fixture built on a WRONG quorum model (I predicted
     a flip that cannot happen — `calibration` substring-matches inside
     `recalibration`); burned a golden pin cycle before building
     `corpus-replace`.
   - Margins exclusion first claimed STRUCK items' tokens (over-exclusion)
     — caught by parity drift, corrected to open-items-only.
   - Mixed literal+spread variadic calls failed to compile twice before I
     probed the rule; and I wrote "go1.26+" in a comment from two data
     points before verifying it hit `fmt.Println` too.
   - `\r\n`-ended old_strings failed two multiedits before I read the
     file properly with `view`.
   - RC-capture pipelines misread once (`grep` exits 0 on match — my
     "want nonzero" check printed a meaningless RC).

## e) WHAT WE SHOULD IMPROVE

1. **Untruncated inventory before any mass edit.** The single change
   that prevents the entire §d incident.
2. **One-row dry-run + gate, THEN batch.** Apply the transformation to
   one row, run the gate, inspect the diff, and only then scale.
3. **Never leave unverified multi-file state in the tree** — the daemon
   commits within ~3 minutes; corrupt intermediates become history.
   Verify + deliberate-commit beats daemon-swept hope.
4. **Read the gate code, not the docstring, before predicting fixture
   semantics** — check.go's substring presence is what decided the
   fixture, and I predicted from prose.
5. **Probe toolchain rules empirically FIRST** (a 30-second /tmp repro)
   before writing version claims into comments or fighting compile
   errors three times.
6. **Verify your own verification pipelines** — RC captures through grep
   pipes lie; capture the test's exit explicitly.
7. **Time-box judgment sweeps at the inventory step** — when the real
   size is 6× the estimate, scope down before writing any transformation.
8. **Encode the safe strike mechanics in a script** (`scripts/`), so the
   next sweep cannot repeat the sed mistakes even by accident (candidate
   row, §f).
9. **The e2e sentinel's sandbox arm deserves a log-visible proof** — the
   nix check passing does not show WHICH arm fired; assert the skip line
   in the check's log output.

## f) 50 things to get done next

Brainstorm, not commitment (HARVEST must apply routing rigor; [T] =
already tracked in TODO_LIST, [NEW] = surfaced tonight, [SIBLING] /
[OWNER] = gates, [ROADMAP] = fuel, not commitments).

1. [T][OWNER] Cut v0.2.0: fold Unreleased (27+ rows), annotated tag,
   go-release pre-verify, push, proxy/pkg.go.dev confirmation.
2. [T][OWNER] Answer D1: battery adoption PATH-only-for-a-quarter vs
   timer lane (matrix ready in the decision doc).
3. [T][OWNER] Answer D2: root go floor 1.27 with v0.2.0 (tradeoffs in
   the decision doc).
4. [T][SIBLING] SystemNix float: `nix flake update docs-gates` +
   expected-locks + toplevel proof.
5. [T][SIBLING][OWNER] evo-x2 deploy + six-binary live smoke + SystemNix
   TODO row for the deployed switch state.
6. [T][SIBLING] Fleet ratchet corpora re-run (crm, go-appkit) — Corpus
   semantics converged AGAIN tonight (dated-DONE, margins scoping).
7. [T][SIBLING] pbx-artmann migration onto this binary (rides T1's tag)
   - fleet announcement.
8. [T][SIBLING] BuildFlow triage batch: 9-tools `history --last-error`,
   preflight FAIL attribution, binary rebuild + reinstall,
   `buildflow format` once.
9. [T][SIBLING] BuildFlow upstream: #87 fix + roundtrip walker fix, then
   drop both AGENTS bypasses.
10. [T][SIBLING] Pilot ratchet drain in crm + go-appkit (bars tighten as
    debt drains).
11. [T][SIBLING] SystemNix eval legs: rpi3-dns toplevel + standalone
    all-systems check.
12. [T][SIBLING] Cross-check the fleet docs-health skill text vs the new
    flags (crush-config repo).
13. [T][OWNER] nixpkgs maintainers PR + flake meta.maintainers reference.
14. [T][OWNER] After D2: collapse the AGENTS toolchain paragraph to one
    floor sentence (F94).
15. [T] T17 residue: per-report verdict strikes with the AGENTS safe
    mechanics (unique-text anchors, cell-bounded, gate per batch) — only
    rows with demonstrable landing evidence; archive when a report
    resolves fully.
16. [NEW] Pin the margins exclusion in the parity net: a `margins` gate
    over `corpus-replace` (all its near-bar vocabulary is item-claimed →
    pins the exclusion byte-for-byte).
17. [NEW] CI lint jobs: golangci-lint (root + provider) currently run
    locally only; CI runs vet/gofmt — a lint job would close the gap the
    provider job left.
18. [NEW] Encode the safe strike mechanics as `scripts/safe-strike.sh`
    (anchor-on-unique-text, cell-bounded, gate-per-batch) so the next
    sweep cannot repeat §d by accident.
19. [NEW] e2e sentinel sandbox-arm proof: assert the "E2E LEG: skipped"
    line inside the nix provider check's log (the check passes today
    without proving which arm fired).
20. [SIBLING] Fix the `golangci-lint-lsp-wrapper` analysis cache in
    crush-config (bust or version-key `GOLANGCI_LINT_ANALYSIS_CACHE`) —
    the verified root cause of the ghost-diagnostics class.
21. [OWNER/local] Unset `core.hooksPath` locally to arm pre-commit
    auto-install (currently explicitly set to its own default; pre-commit
    refuses to install over it).
22. [T] Diagnose the daemon push lag (22 commits accumulated) — push
    works, the daemon's push cadence did not.
23. [ROADMAP] Margins threshold knob (`marginNearBar` as a config key) —
    this repo's own report lists 57 near-bar tokens; a knob would let
    repos tune signal vs noise.
24. [ROADMAP] Doctor verb: `docs-gates doctor <root>` — config
    validation, corpus health, gitignore-trap probe in one.
25. [ROADMAP] `split-check` verb: dry-run "would splitting this file
    change any golden byte?" via the parity harness.
26. [ROADMAP] File-size gate: `fixStrategy: split` hints carrying the
    LSP-symbol seams.
27. [ROADMAP] File-size gate: higher threshold for `*_test.go`.
28. [ROADMAP] `gatecli.CutStringFlag` sibling for docsstrike's valued
    string flags (only if a second consumer appears).
29. [ROADMAP] New gate class: §f continuation lines beginning
    `\d+[.)]` (the false item-boundary class).
30. [ROADMAP] Result-cache key should include repo identity (BuildFlow
    issue #67 class).
31. [ROADMAP] Findings-gate summary naming the result-cache state
    (hit/miss) per tool.
32. [SIBLING] Verify the 05-06 §f item-9 fresh:skip claim against the
    BuildFlow board's actual state.
33. [T] Fleet-wide search for >350-line Go files proactively (crm,
    go-appkit, bank-sync) instead of waiting for their next run.
34. [ROADMAP] `--json` verdict output for CI annotations (FEATURES
    WORTH CONSIDERING).
35. [ROADMAP] Generic counts gate (numbers-style) — needs a design pass.
36. [T] After T1: bump README "Releases" line + `go install @v0.2.0`
    mention (F5).
37. [T] After T1: CHANGELOG `[v0.2.0]` section date + release link (F2).
38. [T] Provider module: re-run its suite + lint after the next
    toolsdk bump (only checked locally tonight).
39. [ROADMAP] Config: consider warning when `living_docs` names a file
    that does not exist (currently silently empty corpus?).
40. [ROADMAP] Strike verb: `--batch` mode reading a verdict table from
    stdin (would make safe sweeps scriptable).
41. [ROADMAP] annotate: rule for struck rows whose verdict cites a
    DELETED path (currently path-claims must exist; deleted-since is
    legitimate but unverifiable).
42. [ROADMAP] freshness: `--dump-items` shorthand accepting repo root
    (dump the newest report implicitly).
43. [ROADMAP] Battery: print per-gate durations (the matrix names
    verdicts, not cost).
44. [ROADMAP] Corpus: a corpus-corrupt fixture (malformed config) as a
    golden to pin the exit-2 path end-to-end.
45. [T] Re-run `buildflow --build-mode dev` once after the daemon-heavy
    session (fingerprint drift accumulated?).
46. [T] Verify `git check-ignore -v` output shape used by the T9 arm
    stays stable across git majors (the arm depends on exit codes, which
    are contractual — low risk, noted).
47. [ROADMAP] Pre-flight: `--stdin-file <path>` so rewrites can be
    judged from a saved candidate file (scripting ergonomics).
48. [ROADMAP] Battery: `--fail-fast` escape hatch for local loops
    (default stays no-abort).
49. [T] Annotate the 22-52 report's §b/§d claims when T14/T15 land (the
    binary-rebuild + format items resolve observations made tonight).
50. [T] Keep the owner-question cadence: 5th ask for release authority +
    sibling permission rides the next natural contact point (T1 blocks
    the 1% tier).

## g) Questions for the owner

1. **Release authority (5th ask):** may I cut v0.2.0 now — fold
   Unreleased, annotated tag, push, fleet announcement? Everything in
   the 1% tier (and the D2 floor recommendation) rides this one yes.
2. **Sibling permission, blanket or per-repo:** SystemNix (float +
   deploy), BuildFlow (triage + upstream #87/walker), crm/go-appkit
   (ratchet re-runs + drains), crush-config (skill cross-check + the
   wrapper-cache fix)? A single blanket "yes, this week" unblocks nine
   tasks; tell me if any repo is off-limits.
3. **The golangci LSP wrapper cache** (root-caused tonight to
   `GOLANGCI_LINT_ANALYSIS_CACHE` in your host-managed wrapper): should
   I fix it in crush-config (bust/version-key the cache so renames
   invalidate), or keep the cache for speed and live with a documented
   clear-and-restart protocol when diagnostics look stale?

_Point-in-time snapshot; annotate, never rewrite. This report's §f is
HARVEST fuel — the [NEW] in-repo items are being routed into TODO_LIST
now; the ROADMAP-tagged items stay here deliberately._
