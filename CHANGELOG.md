# CHANGELOG

All notable changes. Dates are 2026 unless stated. Version sections
match the annotated git tags; anything after the newest tag sits under
[Unreleased].

## [Unreleased]

### Added

- erraudit adoption (the TODO erraudit-adopt coin): the `//nolint:erraudit`
  directives are live via a private git+ssh flake input + devShell package
  (erraudit's own hermetic flake build, rev pinned in flake.lock), and a
  `nolint-audit` flake check staleness-checks every directive on each
  `nix flake check` (root module; sandboxed with the pinned go toolchain,
  GOPROXY=off). All 15 directives verified needed, 0 stale; the check is
  discrimination-proven against a planted stale directive.
- Golden parity corpus: TestParityGolden pins the five gates + battery
  byte-for-byte over testdata/parity/corpus (custom config, `{{DATE}}`
  token substitution keeps report ages stable); goldens under
  testdata/parity/golden, re-pin via `PARITY_UPDATE=1`. A second corpus
  (corpus-ratchet: annotate_max pinned at 2 over two unannotated strikes)
  goldens the annotate RATCHET arm byte-for-byte, with a discrimination
  assertion that a stricter pin fails REGRESSION. A third corpus
  (corpus-warn) goldens the `config: unknown key` warning through the
  cadence verb, and `preflight-safe` goldens the strike pre-flight
  stdin candidate leg (a swapped os.Stdin fixture) plus its
  already-on-disk refusal arm (exit 2).
- CI workflow (build / vet / gofmt / stack-drift / test on push + PR) and
  `scripts/stack-drift-check.sh`, which fails CI the moment the root
  go.mod grows a require/replace/exclude (the stdlib-only contract); a
  second `provider` job runs build/vet/test from pkg/provider so the
  nested module is gated too (go-finding verified public, no CI creds).
- gatecli unit coverage pins the shared CLI scaffolding's contract:
  CutIntFlag last-wins + exact FATAL bytes + untouched-rest sentinels, and
  LoadConfig warnings-before-error ordering with clean-config silence
  (captured-stderr assertions).
- nolint-audit discrimination test (internal/battery): plants a
  finding-free `//nolint:erraudit` directive in a temp module and asserts
  erraudit reads it STALE with a failing rc — proof the flake's
  nolint-audit check cannot rot into a no-op. Wired as a leg of that
  check; skips on hosts without erraudit on PATH (CI runners).
- Freshness debuggability train: `--margins` prints the proactive
  token-margin report (every living-docs token within 2 DF of the
  generic bar, worst first, with its remaining line budget — so drafts
  avoid boundary words first-try instead of through a failed gate);
  `--dump-items <report>` makes the §f extraction inspectable (section
  found, items extracted, per-item tokens) without test scaffolding; and
  every UNHARVESTED finding now carries harvest attributions — matched
  tokens with their DF (generic-tagged at the bar) plus example living
  lines, the preflight culprit idea extended to the harvest findings
  themselves. All informational: goldens untouched.
- Pre-flight rewrite-mode + heading-awareness (the preflight-bind-check
  remaining legs): `--replace=OLD` removes the named living-doc line
  before the replay, so a REWRITE is judged by its new words alone — an
  item anchored only by the replaced line flips to WOULD-DRIFT, the
  class the additions-only model could not see (discrimination-proven:
  the same candidate is SAFE without --replace, UNSAFE with it). The
  delta baseline stays the pre-edit corpus, so damage the rewrite did
  not cause stays unblamed and a rewrite that repairs drift binds
  visibly. Heading candidates carry the split-landing advisory (land
  heading + batch through ONE run); a candidate equal to its own replace
  target is exempt from the on-disk refusal.
- legacy-wrappers smoke v2: the flake check now asserts each of the six
  names stamps the ACTUAL version string (the ldflags `-X main.version`
  value from the same `version` binding the package builds with), not
  just exit 0 — proven both directions (a wrong expected stamp fails the
  check naming both values). The nixpkgs maintainer-list still lacks a
  LarsArtmann entry (re-verified by eval), so the meta.maintainers half
  of that row stays open pending the owner-gated upstream PR.
- art-dupl calibration audit: `--show-suppressed` shows six groups — all
  test-file similarity, the intentional class — and a `--structural`
  spot-check found no hidden actionable clones (the one non-test group is
  const-declaration shape). The audit's `-t 1` self-scan also caught the
  fresh docsannotate settings projection mirroring docsfreshness's: both
  now carry in-source accepts with rationale (baseline re-proven at 0
  shown groups). The `gatecli.Truncate` canonical-helper triage landed as
  a deliberate non-fix — the width-budgeted go-output helper would break
  the stdlib-only contract and the byte-precise golden shapes.
- Shared-helper round 2 (zero behavior change, goldens byte-identical):
  docsfreshness.Corpus is now the ONE prepared living-docs corpus (raw +
  squashed lines, Contains/Count) — docsrouting's hand-rolled
  tokenIn/squashContains/livingLineCount predicates are deleted and the
  routing triage rides the same squash-arm semantics as the harvest
  (for tokenized input the two conditions were equivalent, and now they
  cannot drift); the double report-dir glob collapsed into one
  unarchivedReports site via Settings.reportGlobPath. The CutStringFlag
  survey ruled docsstrike's next-arg value style a deliberate non-fix:
  its `--file X` shape is fleet API and only CutIntFlag ever had two
  consumers.
- Code-health train (zero behavior change, goldens byte-identical):
  docsannotate gained the settings.go projection (Settings +
  DefaultSettings/SettingsFromConfig/loadSettings, Run no longer touches
  config.Config directly); docscadence.go pre-split along its two seams
  into classify.go (TODO purge counting) and archive.go (report
  inventory), the host file down to 192 lines ahead of the 350 gate; and
  docscadence's TodoFile/ReportDir fallbacks now spell the
  `cmp.Or(override, Defaults)` convention (int cadence bars stay explicit
  if-forms). One parity trap documented in-source: annotate_extra_globs
  needs NO OrSlice fallback — an explicit `[]` means "no extra globs",
  and OrSlice's len>0 check would resurrect the planning-glob defaults.
- erraudit posture sweep (audit-only, never a gate): `erraudit fix`
  dry-run finds zero auto-fixable sites in either module; the `--max`
  audit (suppressions bypassed) reports root 34 / provider 2 violations —
  the dominant "project enforces samber/oops" class is deliberate (the
  stdlib-only contract forbids the dependency), and the one CRITICAL
  (context_loss at docscadence's bare err return) is judged a false
  positive in-source: the sole caller frames the FATAL with the report
  dir, so a wrap would duplicate the context.
- `erraudit-lint` flake check: the findings axis joins the staleness
  axis — erraudit's type-aware legacyerrors audit (errors.As/Is → AsType)
  runs over the root module inside the same hermetic sandbox, so a new
  unmigrated site fails `nix flake check`. Discrimination-proven (planted
  errors.As turns the check red, naming the site; sentinel errors.Is is
  deliberately exempt).
- Provider e2e battery run: TestDetectEndToEndOverRealBinary builds the
  real cmd/docs-gates from the root module and runs provider Detect over
  a fixture that is red in exactly one gate (an unharvestable §f item —
  zero reports is a vacuous PASS, which the probe caught), asserting one
  finding per red gate against the live binary instead of canned matrix
  text.
- `provider` flake check: the nested module's suite now gates under
  `nix flake check` — its own buildGoModule (go 1.27 override, pinned
  vendorHash over the public go-finding deps, GOTOOLCHAIN=local), so
  provider regressions fail the nix gate too, not just CI/local. The
  suite's fake-binary helper moved from `#!/usr/bin/env bash` to
  `#!/bin/sh`: nix sandboxes have no `/usr/bin`, so every Detect test
  failed with `fork/exec … no such file or directory` until the shebang
  went portable POSIX.
- `erraudit-rev` tripwire (scripts/check-erraudit-rev.sh as a flake
  check): the erraudit input tracks master, so a `nix flake update` can
  silently swap the rule set — the check pins the locked rev and fails
  until the pin is bumped deliberately (go-codec's
  check-erraudit-version.sh precedent). Proven both directions at the
  flake level (wrong pin → red naming both revs; restored → green).
- docs/DOMAIN_LANGUAGE.md: the gate vocabulary (verdict matrix, ratchet,
  struck row, harvest, distinctive token, living/historical docs) in one
  greppable glossary.

### Changed

- Per-gate wall-clock timeout in the battery (2026-11-05): every gate
  runs under a cap (config `gate_timeout_sec`, default 300, 0 =
  disabled) via gatecli.CaptureOutputWithTimeout — an expired gate
  becomes a FAIL row `(rc=timeout)` naming the gate and its cap while
  the battery moves on, so a wedged gate can never hang the whole run
  (the CV wedge class, now bounded). Gates take no context, so an
  expired gate is ABANDONED, not killed: it keeps running as a
  goroutine whose late writes fall through to the restored streams
  (cosmetic at worst; the observed class blocks at 0% CPU and never
  writes again) and os.Exit discards it. Discrimination-proven with a
  planted select{} wedge: runGates returns past the cap, the row fails
  rc=timeout, the gate behind the wedge still runs, the battery exits
  1. Run split into runGates/renderMatrix for exactly that injection.
- audit-all performance cliff fixed (2026-11-05 fleet dogfood): the
  prepared Corpus now memoizes token presence and raw-line document
  frequency per corpus, so a replay pays one full-corpus scan per
  DISTINCT token instead of per occurrence per report — the same tokens
  recur across every report, and CV-shaped corpora made every call a
  full rescan. Measured on the committed CV-shaped benchmark
  (1,600 reports × 2,000 living lines): 2,753ms → 47ms per replay
  (59×); a memo≡naive equivalence test guards every verdict class
  (red-controlled with a planted inverted cache), and the parity goldens
  stayed byte-identical with NO re-pin — a pure cache of a pure
  function. Corpus now flows as one value through the freshness gate
  (CheckReport/itemVerdict/harvestAttributions/TokenTable/preflight
  ride it; the lines+squashed+DF-closure triple is gone), so
  docsrouting shares the same warmed caches.
- TODO_LIST purge (20 struck rows → this entry): the 2026-11-04
  freshness-diagnostics, parity-net, CI-lint, safe-strike and
  golangci-lint-lsp-wrapper batches left the board — harvest
  attributions, --margins golden, --dump-items and the dated-DONE
  struck-row parity arm, the gitignore-trap regression net, the
  pre-commit nolint-audit leg, the settings-projection export, the CI
  lint jobs, the e2e sentinel proof: substance lives in the commits
  named in their evidence cells (d92f7f8, ff7cfcd, 510ff58, 94350f5,
  7a16061, 4777163, 9fa4a3a, e3226d7, 4e87b4a, 2508f53) and the
  CHANGELOG entries above; the board reopens at two freshly struck rows
  (the token-memoization win and its subsumed audit-all short-circuit
  sibling, both closed 2026-11-05).
- Three diagnostic surfaces joined the byte-parity net: `--margins` pins
  its near-bar token report, `--dump-items` pins the extraction dump
  (now including a struck-row arm — WHY a `~~row~~` is exempt — plus the
  no-§f refusal bytes), and the corpus carries a dated-DONE struck row
  plus a struck report item so the cadence dated-cell classify fix and
  the strike exemption ride the golden contract. The parity harness
  gained per-gate expected exit codes (the rewrite/refusal arms are
  SUPPOSED to be non-clean) and corpus-root normalization for
  path-carrying outputs.
- Directive honesty now has a commit-time leg (the bank-sync pre-commit
  pattern): git-hooks.nix drives the same `erraudit nolint-audit .`
  staleness check on every Go-touching commit — one hook config shared
  by the sandboxed `pre-commit-check` flake check and the devShell hook
  install, red-controlled live (a planted stale directive fails the
  hook, naming the count). Auto-install needs `core.hooksPath` unset
  locally; the manual command is in AGENTS.md.
- The pre-flight `--replace` rewrite-mode contract is pinned in the
  byte-parity net: a dedicated corpus-replace fixture drives the binary
  to the UNSAFE bytes (the flipped item naming its vanished anchors),
  and a discrimination arm proves the same candidate is SAFE without the
  flag — the flip is provably the rewrite's delta, not the candidate's
  words. The parity harness now supports per-gate expected exit codes
  (the rewrite arm is SUPPOSED to be a findings exit).
- The `--margins` item-claimed exclusion is byte-pinned through the
  binary: a margins-replace parity arm rides corpus-replace, whose
  generic_df drops to 2 so the one-line living doc's whole vocabulary
  sits inside the near-bar window — the golden names only the unclaimed
  tokens (cycle, notes, ships) while the four item-claimed ones would
  headline the report the moment the exclusion regresses. The
  rewrite-mode UNSAFE golden is byte-untouched by the bar change.
- The `.gitignore` `reports/` trap is now a failing suite instead of a
  surprise: a battery test asserts every testdata file is tracked
  (`git ls-files`) and none matches an ignore rule (`git check-ignore
  --no-index` — the plain form is blind to tracked files), both legs
  red-controlled live (a swallowed fixture names itself with the exact
  2026-11-04 failure shape). Rides the existing `go test` CI job.
- The fleet golangci-lint now gates CI (it ran on developer machines
  only, so a lint regression could land behind a green CI): v2.15.0 —
  the flake's pin — over the root module and over the nested provider
  module through the same walked-up config the repo's lint command
  uses. The runner's unknown-linter warning about the
  `//nolint:erraudit` directives is expected and warning-only.
- The strike-sweep safety rails are now a script, not tribal memory:
  `scripts/safe-strike.sh` refuses anchors that match zero or several
  lines (printing the full inventory), anchors that carry no task text
  (the bare-row-number corruption machine), cell rewrites that could
  swallow a neighboring cell, and writes to a file that is not at a
  clean commit boundary; dry-run is the default and `--gate` runs the
  annotate gate after every applied batch. Landed by striking its own
  TODO row through the script.
- The nix provider check now proves its end-to-end skip arm: the
  check's phase greps the test log for the sentinel's sanctioned skip
  ("root module source not adjacent" — the sandbox vendors only
  pkg/provider), so a green check is evidence the e2e coverage is
  accounted for rather than a leg that silently vanished or skipped
  for the wrong reason. Red-controlled in the sandbox: a bogus
  sentinel string fails the build while the Go suite itself passes.
- The provider's end-to-end Detect leg can no longer vanish silently: a
  skip inside the nix provider sandbox (root source not adjacent, no go
  on PATH) records a ledger note that a sentinel cross-checks against an
  independent environment probe — any other skip reason, a leg that ran
  where it could not, or a deleted leg now fails the suite instead of
  hiding in a green `ok` line; a decision-table test pins every
  combination (the CI-visible form of the 19:01 §e.3 silent-skip
  observation).
- Audit-all findings now carry the same harvest attributions as the
  newest-report leg (matched tokens with DF + example living-docs lines
  under each UNHARVESTED row); a battery corpus arm asserts the lines so
  the two render sites cannot drift apart again. Parity goldens
  untouched — the fixture corpus replays FRESH, the change only colors
  red runs.
- File-size refactor (26 files, zero behavior change): the two oversized
  gate files and their tests split along package seams — docsfreshness.go
  (1203 lines → 8 files), docsannotate.go (853 → 4), their test files
  mirroring the split, docsrouting gaining settings.go, docsstrike
  gaining parse.go + strike.go. Declaration-losslessness proven against
  the pre-split commits; parity goldens untouched and green. Verified
  through the `BUILDFLOW_NO_RESULT_CACHE=1` bypass (upstream
  LarsArtmann/BuildFlow#87, constant result-cache key — see AGENTS.md).
- Deduplication pass (art-dupl at threshold 1 → zero actionable groups):
  five shared helpers extracted — `gatecli.CutIntFlag` (one valued
  `--flag=N` parser for `--audit-max`/`--annotate-max`, identical FATAL
  wording), `gatecli.LoadConfig` (one config-load + `config:` warnings
  shape for all six verbs; warnings now also print on the malformed-config
  FATAL path instead of being dropped, and cadence/strike/battery switch
  from `WARN:` to the shared `config:` wording), `docsfreshness.LivingDocsLines`
  (the lowercased raw-line living-docs corpus has exactly one reader for
  freshness + routing), exported `SectionFRe`/`HeadingRe` (one §f/heading
  definition, so the two report-scanning gates cannot disagree), and
  stdlib `cmp.Or` + `config.OrSlice` spelling the zero-value-Settings
  fallbacks as one-liners. Three idiom clusters (flush/continue boundaries,
  covered-line bookkeeping, per-file skip arms) accepted in place via
  `//art-dupl:accept` rationales — acceptance is marked in source, not prose.
- flake.nix hardened: one pinned toolchain (`go_1_27` via
  `buildGoModule.override` + devShell) with `GOTOOLCHAIN=local` and
  `CGO_ENABLED=0` in the devShell — kills go auto-mode's silent go1.27
  network download into the module cache (pkg/provider pins go 1.27,
  which go_1_26 + local rejects). devShell switched to `mkShellNoCC`.
- flake hygiene pass: `erraudit.inputs.nixpkgs.follows = "nixpkgs"`
  evaluates erraudit against this flake's nixpkgs (both sides sat on the
  same rev — zero behavior delta), removing the nixpkgs_2 lock node; the
  `GOFLAGS=-mod=mod` env flag dropped from both erraudit checks (proven
  unneeded: the stdlib-only root module never needs write mode); the two
  duplicated `errauditPkg` bindings collapsed into one top-level
  `system:` function.
- `meta.platforms = lib.platforms.linux`; lib bound once (no more
  `pkgs.lib.`); checks extended with `default` (wrapper surface) and
  `legacy-wrappers` (smoke: each of the five legacy names +
  `docs-gates` answers `--version` rc=0 from the installed PATH).
- pkg/provider lint hygiene: the `//nolint:gosec,noctx` directive moved
  onto the `exec.CommandContext` call-start line (golangci reports G204
  there — the closing-paren placement never matched, so the gosec finding
  it meant to suppress was firing), the genuinely unused `noctx` token
  dropped (CommandContext cannot trigger noctx), and the matrix parser's
  magic `2` became the `minMatrixFields` constant.
- docsannotate's `headingRe` now carries a comment distinguishing it from
  `docsfreshness.HeadingRe` (`^#{1,6}` whole-report scan vs `^#{1,2}`
  §-structure scan) so a future dedup pass cannot merge them.
- Docs-health audit pass (2026-11-04): TODO_LIST rebuilt by domain with
  every §f item from the five status reports routed (blanket ROUTE rows
  dissolved), ROADMAP landed rows removed, AGENTS.md gained the
  baseline-pinning daemon lesson (pin `git rev-parse HEAD` before
  diffing — the daemon moves HEAD mid-session).

### Changed

- Docs surface caught up with the flags (they shipped CHANGELOG-first):
  the README gains a freshness-diagnostics section (`--margins`,
  `--dump-items`, `--pre-flight` with `--replace` rewrite-mode) and a
  per-verb exit-code table (battery's gate-rc2-to-FAIL nuance included);
  DOMAIN_LANGUAGE defines the six new terms (rewrite-mode, margins,
  dump-items, harvest attribution, corpus, settings-projection) and
  corrects pre-flight's replay target to the report corpus; FEATURES
  rows cover the three freshness flags, the full flake check list
  (erraudit-lint, erraudit-rev, provider), and the provider suite's
  flake-check wiring.

### Fixed

- `skip_gates` validation: a typo'd gate name used to be a silent no-op
  (the battery matches via `slices.Contains` and never notices); it now
  exits 2 naming the unknown entry and the five valid gates. Empty
  entries refuse the same way. The battery's matrix names ride the new
  `config.Gate*` constants, so the vocabulary cannot drift. Also fixed
  `annotate_max` missing from the known-keys list, which made every repo
  pinning it get a spurious "unknown config key" warning.
- docs-cadence purge counter blind to dated DONE cells: a status cell
  had to exactly equal `✅ \`DONE\``/`DONE` to count, so the dated strike
  shape (`✅ DONE 2026-11-04`) landed by the execution sessions counted
  as 0 — 25 struck rows sat under a 20 threshold the gate could never
  trip. The counter now accepts any cell starting with a DONE marker
  (prose DONE still never occupies a whole cell); found by the evening
  docs-health pass when the ritual said purge and the gate said healthy.

## [v0.1.1] - 2026-11-04

### Added

- `annotate_max` adoption ratchet (docs-annotate gate + `--annotate-max=N`
  CLI + battery pass-through): tolerates up to N findings green-but-named,
  REGRESSION past the pin. Born from the pilot rollouts (crm: 134
  pre-doctrine strike findings; go-appkit: 255); the strict default
  preserves the pre-extraction behavior byte-for-byte (parity re-verified
  over the source corpus after the change).
- BuildFlow provider: `pkg/provider` as its own go module (go-finding
  v1.15.0 + toolsdk v1.16.0 — the root module stays stdlib-only),
  blank-import self-registration via `toolsdk.Register`; Detect exec's the
  battery from PATH and parses the verdict matrix into one finding per
  red gate; read-only by contract (Repair nil). Released as the nested
  module tag `pkg/provider/v0.1.1`.
- Pilot rollouts seeded: crm + go-appkit carry `.docs-gates.json` with
  audit_max / annotate_max / cadence bars at their measured debt; both run
  5/5 PASS.

## [v0.1.0] - 2026-11-03

### Added

- Extraction train: the five docs gates (freshness, annotate, routing,
  cadence, strike) ported from the source fleet's `pbx-tools` package,
  import paths rewritten, per-package config plumbing (`Settings`
  projections with zero-value fallbacks to the historical constants).
- `.docs-gates.json` config schema (JSON, stdlib-only): report dir/glob,
  living docs, annotate extras, rule-F toggle + extensions, the
  distinctive-token bar, the audit ratchet, cadence thresholds, TODO file,
  per-gate skips. Unknown keys warn; malformed/invalid values exit 2.
- Battery runner `docs-gates run`: five gates as one independent verdict
  matrix, `skip_gates` + `audit_max` from repo config. Live-proven 5/5
  PASS over the source fleet's real corpus.
- Multi-verb binary `cmd/docs-gates` with per-verb `--version` under the
  legacy standalone names.
- flake.nix: buildGoModule package (vendorHash-free), five legacy-name
  wrappers, Linux-only systems, devShell (go 1.26 + golangci-lint +
  nixfmt-tree), tracked treefmt.toml.
- Fleet golangci config, brought to 0 findings (static config errors,
  named default constants, parallel tests, tagliatelle snake_case rule
  for the jq-able config schema).
- GitHub repo (public, MIT), meta registration.
