# CHANGELOG

All notable changes. Dates are 2026 unless stated; no tagged releases
before v0.1.0.

## 2026-11-04

### Added

- erraudit adoption (the TODO erraudit-adopt coin): the `//nolint:erraudit`
  directives are live via a private git+ssh flake input + devShell package
  (erraudit's own hermetic flake build, rev pinned in flake.lock), and a
  `nolint-audit` flake check staleness-checks every directive on each
  `nix flake check` (root module; sandboxed with the pinned go toolchain,
  GOPROXY=off). All 15 directives verified needed, 0 stale; the check is
  discrimination-proven against a planted stale directive.
- BuildFlow provider: `pkg/provider` as its own go module (go-finding
  v1.15.0 + toolsdk v1.16.0 — the root module stays stdlib-only).
  Blank-import self-registration via toolsdk.Register; Detect exec's the
  battery from PATH and parses the verdict matrix into one finding per
  red gate; read-only by contract (Repair nil).
- Golden parity corpus: TestParityGolden pins six invocations (five gates
  + battery) byte-for-byte over testdata/parity/corpus (custom config,
  `{{DATE}}` token substitution keeps report ages stable); goldens under
  testdata/parity/golden, re-pin via `PARITY_UPDATE=1`.
- CI workflow (build / vet / gofmt / stack-drift / test on push + PR) and
  `scripts/stack-drift-check.sh`, which fails CI the moment the root
  go.mod grows a require/replace/exclude (the stdlib-only contract).

### Changed

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

## 2026-11-03 (late)

### Added

- `annotate_max` adoption ratchet (docs-annotate gate + `--annotate-max=N`
  CLI + battery pass-through): tolerates up to N findings green-but-named,
  REGRESSION past the pin. Born from the pilot rollouts (crm: 134
  pre-doctrine strike findings; go-appkit: 255); the strict default
  preserves the pre-extraction behavior byte-for-byte (parity re-verified
  over the source corpus after the change).
- Pilot rollouts seeded: crm + go-appkit carry `.docs-gates.json` with
  audit_max / annotate_max / cadence bars at their measured debt; both run
  5/5 PASS.
## 2026-11-03

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
