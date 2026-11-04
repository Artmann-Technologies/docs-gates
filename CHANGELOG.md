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
- docs/DOMAIN_LANGUAGE.md: the gate vocabulary (verdict matrix, ratchet,
  struck row, harvest, distinctive token, living/historical docs) in one
  greppable glossary.

### Changed

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

### Fixed

- `skip_gates` validation: a typo'd gate name used to be a silent no-op
  (the battery matches via `slices.Contains` and never notices); it now
  exits 2 naming the unknown entry and the five valid gates. Empty
  entries refuse the same way. The battery's matrix names ride the new
  `config.Gate*` constants, so the vocabulary cannot drift. Also fixed
  `annotate_max` missing from the known-keys list, which made every repo
  pinning it get a spurious "unknown config key" warning.

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
