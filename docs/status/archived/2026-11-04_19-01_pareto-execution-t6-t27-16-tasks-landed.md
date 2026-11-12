# Pareto Execution — 16 In-Repo Tasks Landed (T6–T27), Owner-Gated Tier Open

> **Live remainder:** owner-gated tier (§c) lives in TODO_LIST §1–§5 +
> the 07:14 Pareto plan; §f rows carry their TODO §N homes; §g owner
> questions live with the owner. Archived 2026-11-12 (cadence ritual).

2026-11-04 19:01 — docs-gates, session report. Resumes the 16:45 report
(`2026-11-04_16-45_pareto-execution-t3-t4-t5-t7-t9-t10-plant-incident.md`);
that one waited on 3 owner answers (deploy auth, sibling-repo permission,
CHANGELOG granularity) — none answered, so this session continued the
in-repo tier exactly as its §f ordered.

## a) FULLY DONE

Head `acbc98a`, pushed, CI green both jobs (run 41878848461). Every task
landed red→green with discrimination proof, lint 0 both modules, parity
goldens byte-identical throughout.

- **Pre-work**: pending docs commit pushed; flake verification debt from
  the 16:45 report closed (fmt / both-arch eval / build, all green).
- **AGENTS.md lessons written** (3): blind-revert-under-daemon
  (`git log -1 --stat` first, revert by SHA, commit own work fast);
  erraudit system-installed on fleet hosts (CI-passing tests may be
  silent skips) + finding domain is legacyerrors only (sentinel
  errors.Is exempt) + flake-source probes must STAGE plants; buildflow's
  managed `reports/` gitignore swallows corpus fixtures (scoped
  negations, check-ignore ritual). Stale CI paragraph fixed on sight.
- **CHANGELOG/TODO backlog**: rows for T3/T5/T7/T10 + the session's
  corpora; TODO §2.1/§3.1 struck with evidence.
- **T6 erraudit-lint flake check**: findings axis hermetically gated
  (`erraudit lint ./... --type-aware`, nolint-audit sandbox plumbing);
  planted errors.As → check red naming the site; sentinel Is exempt
  (probed, not assumed).
- **T8 erraudit-rev tripwire** (scripts/check-erraudit-rev.sh as a flake
  check): pins the locked rev vs flake.lock (jq); wrong pin → red naming
  both revs; go-codec precedent honored.
- **T17 flake hygiene**: `erraudit.inputs.nixpkgs.follows` dedupes the
  second nixpkgs lock node (same rev both sides, zero behavior delta);
  `GOFLAGS=-mod=mod` proven unneeded and dropped from both erraudit
  checks; duplicated `errauditPkg` bindings collapsed into one top-level
  function (Nix list-application paren lesson along the way).
- **T11 provider flake check**: nested module's suite gates under
  `nix flake check` (own buildGoModule, vendorHash pinned over public
  deps, GOTOOLCHAIN=local). Surfaced + fixed the fake-binary shebang
  (`/usr/bin/env` doesn't exist in nix sandboxes → `#!/bin/sh`).
- **T12 provider e2e**: TestDetectEndToEndOverRealBinary builds the real
  cmd/docs-gates, Detect over a harvest-red fixture → exactly 1 finding,
  rule freshness. Caught en route: zero reports is a vacuous PASS (the
  first fixture proved nothing) and the test silently SKIPPED on a wrong
  relative path until run verbose.
- **T19 erraudit posture sweep**: fix dry-run zero fixes both modules;
  --max audit root 34 / provider 2 violations, recorded as posture
  (never a gate); the 1 CRITICAL (context_loss, docscadence bare return)
  judged false positive in-source — the sole caller frames the FATAL
  with the report dir; the "enforces samber/oops" ERROR class is
  deliberate (stdlib-only forbids the dep).
- **T20 code health**: docsannotate settings.go projection (Settings +
  loadSettings, Run off raw config.Config); docscadence pre-split into
  classify.go + archive.go (326→192 lines, goldens byte-identical);
  cmp.Or string-fallback convention applied (ints stay explicit).
  Parity trap documented: annotate_extra_globs needs NO OrSlice fallback
  (explicit [] means none; the corpus pins it).
- **T21 shared-helper round 2**: docsfreshness.Corpus is the ONE
  prepared corpus (raw+squashed, Contains/Count) — docsrouting's three
  hand-rolled predicates deleted; single unarchivedReports glob site via
  Settings.reportGlobPath. CutStringFlag survey: deliberate non-fix
  (docsstrike's next-arg style is fleet API; only CutIntFlag had two
  consumers).
- **T22 triage**: art-dupl --show-suppressed (6 groups, all test-file
  similarity) + --structural spot-check (no hidden actionable clones);
  the -t 1 self-scan caught the session's OWN new projection pair →
  accepted-in-source ×2 (accepts now 5, 0-shown baseline re-proven).
  gatecli.Truncate triage: deliberate non-fix (go-output/nom helper
  breaks stdlib-only + width-budget/ellipsis semantics vs the byte-cut
  the goldens pin).
- **T24 legacy-wrappers v2**: check asserts the stamped version STRING
  per name (red+green proven by planting a wrong stamp). In-repo half
  done; upstream half open (see b/c).
- **T25 pre-flight rewrite-mode + heading-awareness**: `--replace=OLD`
  (repeatable, = form) removes the old line before the replay so
  rewrites are judged by new words alone; delta baseline stays the
  PRE-edit corpus (an item anchored only by the replaced line flips —
  discrimination-proven both ways); heading candidates carry the
  split-landing advisory; replaced-line candidates exempt from the
  on-disk refusal.
- **T26 --margins**: proactive near-bar token report (DF within 2 of the
  bar, worst first, remaining budget) — dogfood run names 56 tokens.
- **T27 --dump-items + harvest attributions**: §f extraction inspectable
  without test scaffolding (items + tokens; unreadable/no-§f are usage
  refusals); UNHARVESTED findings carry matched-token attributions (DF,
  generic tag, example living lines). All informational — goldens
  untouched.

## b) PARTIALLY DONE

- **T24 upstream half**: nixpkgs `lib.maintainers.LarsArtmann` still
  absent (re-verified by eval 2026-11-04); flake meta.maintainers
  intentionally not touched until the upstream PR lands. EXTERNAL,
  owner-gated.
- **T27 attribution leg covers checkNewest only**: the auditAll replay
  renderer (audit.go:25) still prints UNHARVESTED without
  harvestAttributions — the battery's audit-all leg lacks the lines the
  single-report leg gained. The battery stayed green only because the
  repo's own corpus findings never take that path. One-line follow-up.
- **CHANGELOG granularity (owner Q3)**: defaulted to per-task rows this
  session; the Unreleased section is now large — a release cut (section
  - annotated tag, go-release ritual) would settle the question de facto.
- **Routing-triage semantics convergence**: docsrouting now rides the
  freshness squash-arm condition (key-length, not raw-length). For
  hyphenated tokens that is a real (intended) behavior convergence;
  tokenized golden corpus is unaffected — proven — but the fleet-wide
  ratchet corpora (crm, go-appkit) have not been re-run against the
  converged semantics.

## c) NOT STARTED (all owner-gated or board rows untouched this session)

- **T1** SystemNix float (`nix flake update docs-gates` + expected-locks
  - evo-x2 toplevel proof) — P1, needs sibling-repo permission.
- **T2** evo-x2 deploy (`nix run .#deploy`, live-PATH smoke) — P1,
  owner-gated quiescent window.
- **T13** pbx-artmann migration onto this binary (the fleet migration
  plan the wrapper-contract names) — needs sibling permission.
- **T14** crm / **T15** go-appkit pilot ratchet drains — sibling repos.
- **T16** rpi3-dns toplevel eval + **T18** SystemNix post-integration
  flake check + SystemNix TODO row — sibling repos.
- **T23** BuildFlow upstream items (roundtrip walker bug, BuildFlow#87
  cache key, binary rebuild/format, health-check triage) — sibling repo,
  includes filing upstream.
- **D1/D2 owner decisions**: battery adoption scope (timer/PMA vs
  PATH-only); root go.mod floor 1.26 vs 1.27. Prepared as
  recommendations, never agent-decided.
- **Pre-commit nolint-audit leg** (bank-sync pattern) — board row,
  untouched; the flake check now exists, the commit-time leg doesn't.
- **Harness LSP ghost-diagnostics issue** (gopls phantom errors for
  moved files) — the session hit it ~15 times (zz_plant ghosts all day);
  reproduce + file, board row open.

## d) TOTALLY FUCKED UP (owned, each with the fix that landed)

1. **Rewrite-mode shipped-inverted delta baseline — almost.** The first
   implementation computed baseFailing on the POST-removal corpus, which
   makes an item killed by the removal count as "already failing" → the
   exact class the feature targets reports SAFE. Caught by MY OWN
   discrimination test, but the test was written AFTER the feature
   (inverted order vs the plan's red→green discipline) — the probe
   binary run is what exposed it. Fix: baseline moved to the pre-edit
   corpus; discrimination test is permanent (preflight_rewrite_test.go).
2. **lint=1 committed (066ac47).** The daemon urgency lesson ("commit
   own work immediately") overrode the lint-at-write-time lesson: the
   accept-marker gci finding rode into a commit; formatter-normalized
   `//art-dupl:accept` → `// art-dupl:accept` was unverified against the
   detector. Fix landed within one commit (800d4cd) and art-dupl still
   recognizes the spaced marker (re-proven). Lesson: verify AFTER the
   formatter touches suppression-like markers, BEFORE committing.
3. **Three write/edit races lost to the daemon/formatter** (e2e_test.go
   rewrite rejected twice — a stale multi-line exec.Command version
   persisted on disk for one whole verify round; rewriteRoot fixture
   edit rejected once), plus **three daemon commit-splits** (7f2cfe4 took
   flake.nix mid-T11; 76df0db took the T21 code; 64e53d2/8add341 took
   stragglers) — work sits in `chore: auto-commit` commits instead of
   deliberate trains. Content all landed and verified; history attribution
   is the damage. Mitigation for next session: read-back after EVERY
   write near a formatter/daemon tick; stage+commit at task granularity,
   not file-granularity.
4. **The T12 e2e test was silently worthless twice**: wrong relative
   path (`../` vs `../..`) made it SKIP while the suite "passed", and
   the first fixture (no reports) was a vacuous PASS. Only running it
   verbose — because the ritual says prove the discrimination leg ran —
   caught both. Fix landed (path + qxvorlag harvest-red fixture); the
   general rule now in AGENTS: a test that "passes" on CI may be a
   silent skip; prove it ran.
5. **The 03-34 report's carried §f row went red at session END** — the
   nolint-audit-cadence row should have been struck when T7 landed
   (same-session §f hygiene), not discovered by the final battery.
   Fix: struck inline (acbc98a) with the landing evidence. Cost: one
   ritual round-trip.
6. **flake.nix sed produced invalid Nix** — `errauditPkg system` inside
   a list is two items without parens; nix fmt silently reflowed it and
   only eval caught it. Fix: `(errauditPkg system)` ×3. Cheap, but a
   hand-edit of a formatter-owned file is exactly the class the
   treefmt.toml lesson warns about.
7. **AGENTS.md multiedit ambiguity burned an edit** (one-line anchor
   matched two "Known BuildFlow tool bug" bullets) — caught by the
   follow-up grep, fixed with wider context. The exact-match discipline
   (3–5 context lines) was skipped once; it failed exactly as documented.

## e) WHAT WE SHOULD IMPROVE

- **Discrimination test FIRST, always.** T25's inverted order cost a
  probe cycle; T12's fixture taught it too. The plan's red→green exists
  because the feature author cannot see their own baseline bug.
- **Formatter/daemon write-protocol**: after any write, re-read the
  region before the next edit; commit at task boundaries with a final
  lint BEFORE the commit, not after.
- **Silent-skip blindness**: every host-dependent test (erraudit on
  PATH, adjacent root source) should print its skip loudly or be paired
  with a CI-visible assertion that the skip HAPPENED — the provider
  check currently loses the e2e leg silently in the sandbox.
- **One renderer, one feature**: attribution was wired into one of two
  render sites; a sweep for sibling render sites belongs in the task's
  definition-of-done (grep the verb's output literals).
- **Same-session §f hygiene**: landing a TODO row named in an old
  report's §f should strike that report row in the SAME commit train.
- **README/docs lag the surface**: three new user-facing flags
  (--margins/--dump-items/--replace) shipped without README or
  DOMAIN_LANGUAGE updates — the docs lagged by design this session
  (CHANGELOG-first) and that debt is now real.

## f) UP TO 50 NEXT THINGS

1. Wire harvestAttributions into the auditAll replay renderer (audit.go
   UNHARVESTED leg) — parity-safe, one-line + test.
2. Cut the release: section the Unreleased CHANGELOG rows, annotated tag
   (v0.2.0 shape), go-release ritual pre-verification.
3. README: document --margins, --dump-items, --replace (rewrite-mode)
   and the exit-code contract per verb.
4. DOMAIN_LANGUAGE.md: add the new vocabulary (rewrite-mode, margins,
   dump-items, harvest attribution, Corpus, settings-projection).
5. FEATURES.md rows for the three new freshness flags + the three flake
   checks (erraudit-lint, erraudit-rev, provider).
6. Pre-commit nolint-audit leg (bank-sync pattern) catching directive rot
   at commit time, complementing the flake check.
7. Parity corpus arm for rewrite-mode: golden `--replace` through the
   BINARY (the unit tests cover it; the byte-parity net does not).
8. Golden the --margins output shape (near-bar corpus fixture).
9. Golden --dump-items (report fixture incl. the no-§f refusal bytes).
10. CI-visible assertion for silent skips: the provider check's e2e leg
    vanishes in the nix sandbox — make the skip print a count or assert
    skips happened, so coverage loss is visible.
11. Re-run the fleet ratchet corpora (crm, go-appkit) against the
    converged Corpus semantics (routing squash-arm change).
12. T1 (board): SystemNix float — nix flake update docs-gates +
    expected-locks + evo-x2 toplevel proof.
13. T2 (board): evo-x2 deploy to live PATH + six-binary --version smoke.
14. T13 (board): pbx-artmann migration onto docs-gates (the fleet
    migration plan), then fleet announcement.
15. T14/T15 (board): crm + go-appkit pilot ratchet drains (audit_max,
    annotate_max, cadence bars tighten as debt drains).
16. T16 (board): rpi3-dns toplevel eval after the lars-packages change.
17. T18 (board): SystemNix post-integration standalone flake check.
18. SystemNix TODO_LIST row for the deploy-pending integration.
19. T23 (board): file the BuildFlow roundtrip `/nix/store` walker issue
    upstream (diagnosis already in this repo's AGENTS.md).
20. T23 (board): fix BuildFlow#87 upstream (constant result-cache key
    over the empty FileContract), then drop the
    BUILDFLOW_NO_RESULT_CACHE bypass from AGENTS.md.
21. T23 (board): rebuild the BuildFlow binary (binary-freshness advisory
    74a9212 vs 4211653) + one `buildflow format` pass.
22. T23 (board): triage buildflow's "9 tools unavailable" health-check
    failure via buildflow history --last-error.
23. Owner D1: battery adoption scope on hosts (timer/PMA lane vs
    PATH-only) — prepare the recommendation matrix.
24. Owner D2: root go.mod floor unification (1.26 vs 1.27) — prepare the
    recommendation.
25. nixpkgs: submit lib.maintainers.LarsArtmann upstream PR, then
    reference it in flake meta.maintainers.
26. Harness: reproduce + file the LSP ghost-diagnostics issue (gopls
    phantom errors for moved/deleted files; zz_plant.go ghosts ran all
    session).
27. ANNOTATE sweep over the older unarchived reports (docs/status holds
    7 unarchived vs cadence threshold 15) — resolve carried rows inline,
    archive the resolved ones.
28. Extend the rewrite-mode model to multi-line replacements (a range or
    block swap) if fleet rewrites need it — ROADMAP-grade idea.
29. Consider `--margins --json` for tooling consumption (jq-able, the
    config-format precedent).
30. Consider docsrouting Gap rows carrying harvest suggestions (the
    --margins vocabulary applied to unrouted rows).
31. docsannotate: apply the settings-projection accept pattern to
    docsrouting's private settingsFromConfig for consistency (or export
    it like the other gates).
32. Flake check naming: consider one meta-check aggregating the three
    erraudit checks for single-invocation fleets.
33. Measure whether the vendorHash FOD slows `nix flake check` enough
    to warrant a check toggle (provider suite off in quick loops).
34. Survey remaining `//nolint` directives fleet-wide for the same
    staleness treatment erraudit got (errcheck/gosec directives have no
    tripwire).
35. Add the erraudit-rev bump ritual line to AGENTS Commands (the script
    header documents it; the command list doesn't).
36. battery_test.go: a corpus arm asserting the audit-all attribution
    lines once item 1 lands.
37. Gatecli: survey whether Errf/Errln write failures should ever count
    (currently policy: never) — document the stderr-side rationale next
    to Outf's.
38. Pre-flight: model REMOVALS without replacement (`--remove=OLD`) —
    deleting a line is not a rewrite; today only --replace exists.
39. TokenMargins: exclude tokens already present in the NEWEST report's
    §f items (those are already harvested — the budget warning is only
    interesting for unclaimed vocabulary).
40. DumpItems: add the struck-row arm (why a `~~row~~` is exempt) so the
    extraction decision is fully explainable.
41. The 16:45 report's §g Q3 (CHANGELOG granularity): settle by doing
    item 2 (release cut) rather than by prose decision.
42. AGENTS.md: the Critical-constraints list is growing long — consider
    moving paid-for war stories to a docs/lessons file with one-line
    pointers.
43. Flake: erraudit input still tracks master by policy — revisit
    pinning a tag once erraudit cuts releases.
44. pkg/provider: tag the nested module (pkg/provider/v0.1.x) if the
    provider surface changed for consumers (it didn't this session —
    verify before tagging).
45. Cross-check the fleet docs-health skill text against the new flags
    (the skill's battery section predates --replace/--margins).
46. go.mod floor: when D2 lands, simplify the AGENTS toolchain paragraph
    (one floor sentence instead of the exception story).
47. Consider a `docs-gates doctor` verb (PATH, config, flake state) —
    the provider HealthCheck hint is the seed.
48. Sweep the temp-probe debris convention: /tmp/dg-* files accumulate
    across sessions; a cleanup line in the ritual (or mktemp -d uniform)
    would keep probes self-describing.
49. battery_test: add a regression arm for the `reports/` gitignore trap
    (assert corpus fixtures stay tracked — `git check-ignore` in CI).
50. Announce the session's gate features fleet-wide AFTER the release
    cut (item 2), not before — feature announcements ride versions.

## g) THREE QUESTIONS I CANNOT ANSWER MYSELF

1. **Release authority**: should I cut v0.2.0 now (CHANGELOG section +
   annotated tag + push, item 2) to settle the granularity question and
   version the new flags — or do you want a quiet accumulation period
   first?
2. **Sibling-repo permission**: the whole P1 tier (T1/T2 SystemNix float
   - deploy, T13 pbx-artmann migration, T14/T15 pilot drains, T18,
     T23 BuildFlow upstream) is blocked on your go-ahead to work directly
     in SystemNix / pbx-artmann / crm / go-appkit / BuildFlow. Grant
     blanket permission for the tier, or gate repo-by-repo?
3. **Battery adoption scope (D1)**: you own the decision recorded in
   TODO §5 — do you want the gates running as a timer/PMA lane over
   fleet repos (autonomous red-on-drift), or PATH-only (humans invoke)?
   My recommendation is PATH-only for a quarter (let the pilots' ratchet
   drains produce data first), but the call changes T2's smoke shape.
