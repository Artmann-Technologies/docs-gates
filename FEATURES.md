# FEATURES

Honest status inventory. DONE = proven by a test or a live run in this
repo; PARTIALLY DONE = works but incomplete; PLANNED = designed, not
built.

## DONE

- **docs-freshness gate** — newest-report open-§f harvest check,
  `--audit-all` whole-corpus replay, the `--audit-max=N` ratchet,
  `--explain-item` / `--explain-token` diagnostics, `--pre-flight`
  candidate replay (additions + rewrite-mode `--replace`, the
  `--duplicates` modeling leg, split-heading warning), `--margins`
  token-budget report, `--dump-items` extraction dump. Ported from the
  source fleet with identical output semantics; config-plumbed (report
  dir, living docs, distinctive-token bar).
- **audit-all cost model** — the prepared Corpus memoizes token presence
  and document frequency per corpus (CV-shaped replay: 2,753ms → 47ms,
  59×; memo≡naive equivalence test guards every verdict class), the
  replay prints its shape + worst-case estimate BEFORE running
  (reports × living lines × the measured per-report-line constant), and
  TestReplaySpeedGuard pins the memoization contract (>1s = regression).
- **docs-annotate gate** — every struck item carries a resolvable verdict
  (hash/tag claims via git, repo-path claims), rule F inline-code spacing
  scan over configurable extensions, config-driven report dir + extra
  globs + rule-F toggle, and the `annotate_max` adoption ratchet (born
  from the 2026-11-03 pilots: repos with pre-doctrine strike corpora seed
  the measured count and drain down; strict -1 default preserves the
  pre-extraction behavior byte-for-byte).
- **docs-routing gate** — non-§f routing-gap lint (`--strict`), pointer
  convention, generic-token bar shared with freshness.
- **docs-cadence gate** — struck-row purge threshold + archive cadence
  (unarchived count + oldest age), config-driven thresholds, TODO file,
  report dir.
- **docs-strike tool** — TODO_LIST row-strike helper (dry-run default,
  ambiguity refusals, double-strike refusals, verdict-opener refusals),
  target file from config unless `--file` overrides.
- **Battery runner** (`docs-gates run`) — five gates as one independent
  matrix (PASS/FAIL/SKIP per gate, no first-failure abort), exit 0/1,
  `skip_gates` + `audit_max` from repo config. Every gate runs under
  `gate_timeout_sec` (default 300s, 0 = off; an expired wedge is a FAIL
  row `rc=timeout`, never a hang), a heartbeat names each gate's elapsed
  time (`done in <dur>` + the matrix `(in <dur>)` column) and a
  watchdog names the cap live once a gate burns past half its budget.
  The `--json` machine face puts exactly one verdict document on stdout
  (snake_case: per-gate verdict, exit_code, duration_ms, timed_out,
  embedded captured streams; humans ride stderr; exit contract
  unchanged) — pinned byte-for-byte by the battery-json parity golden.
  Live-proven over the source fleet's corpus: 5/5 PASS.
- **BuildFlow provider** — pkg/provider as its own go module
  (go-finding/toolsdk): blank-import self-registration, Detect exec's
  the binary from PATH with `run --json` and decodes the verdict
  document into one finding per FAIL row (the decoder refuses unknown
  fields, so a row-shape change surfaces as a shape-break finding, never
  a silent zero value; rows win over the pass summary); read-only by
  contract (Repair nil). Released as `pkg/provider/v0.1.1`; HealthCheck
  covered by its own test; the module's suite rides a flake `provider`
  check and its e2e leg runs the REAL binary.
- **Parity goldens** — TestParityGolden pins fourteen invocations (five
  gates + battery + battery-json + the margins/dump-items/preflight
  arms, including their intended non-clean exits) byte-for-byte over
  the fixture corpora testdata/parity/* (custom config, `{{DATE}}`
  date-token substitution so ages stay stable); goldens under
  testdata/parity/golden; re-pin via `PARITY_UPDATE=1`; a gitignore-trap
  test fails the suite the moment a corpus fixture is untracked.
- **Config schema** — `.docs-gates.json` loader: defaults for everything,
  unknown-key warnings, malformed/invalid = exit 2.
- **Multi-verb binary** — six verbs, per-verb `--version` under the
  legacy standalone names; `docs-gates help [verb]`, `--help`/`-h`
  (root and per-verb) print the gate packages' Usage texts with
  did-you-mean suggestions for unknown verbs (prefix + edit distance),
  and `help <verb> extra` names the ignored arguments.
- **CI workflow** — build / vet / stack-drift / test plus the audit-all
  cliff benchmark (CV-shaped bench at 100 iterations under an OS
  `timeout 45s` — the ~10x-regression kill switch) and the fleet
  golangci-lint (v2.15.0, matching the flake pin) on push + PR; a
  `provider` job runs the nested module (build / vet / lint / test); a
  `dogfood` job runs the battery over THIS repo every push (full
  history checkout so annotate's hash claims resolve). No standalone
  gofmt step — golangci gates format on both jobs via gofumpt.
- **stack-drift check** — `scripts/stack-drift-check.sh` fails CI the
  moment the root go.mod grows a require/replace/exclude (the
  stdlib-only contract, mechanically enforced).
- **safe-strike rails** — `scripts/safe-strike.sh` (unique-text anchors
  only, `[^|]*`-bounded cell rewrites, clean-commit-boundary check,
  dry-run default, `--gate` runs annotate per batch, `--battery` runs
  the whole five-gate battery per batch) plus
  `scripts/safe-strike-test.sh` pinning the refusal matrix on a scratch
  git repo.
- **flake.nix** — buildGoModule (vendorHash-free, stdlib-only), five
  legacy-name wrappers, both-Linux systems, devShell (go_1_27 pinned +
  GOTOOLCHAIN=local + golangci-lint + erraudit + nixfmt-tree +
  actionlint + shellcheck), checks: wrapper surface + legacy-wrappers
  smoke + nolint-audit staleness + erraudit-lint (type-aware
  legacyerrors audit) + erraudit-rev (pinned input-rev tripwire) + the
  provider module suite + golangci-drift (ci.yml action version ==
  locked nixpkgs golangci) + go-floors (root+provider floors within the
  pinned toolchain) + shell-lint (shellcheck + actionlint) + the
  skills guards.
- **Dependabot** — weekly grouped updates for the workflow action SHAs
  and both go modules (`.github/dependabot.yml`), so the checkout /
  setup-go / golangci pins never age without a renewal PR.
- **Releases** — v0.1.0 (the extraction train) and v0.1.1 (annotate_max
  ratchet + provider module) tagged; `go install …@latest` resolves.
- **Pilot rollout evidence** — crm + go-appkit seeded (2026-11-03) and
  green under adoption ratchets (audit_max, annotate_max, cadence bars at
  the measured debt); the ratchets drain as the rituals run.

## WORTH CONSIDERING

- A `numbers`-style counts gate (test/check counts vs docs claims) — the
  source fleet's docs-numbers is repo-specific; a generic shape needs a
  design pass first.
