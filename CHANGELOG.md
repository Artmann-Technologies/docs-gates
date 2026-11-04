# CHANGELOG

All notable changes. Dates are 2026 unless stated; no tagged releases
before v0.1.0.

## 2026-11-04

### Added

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
