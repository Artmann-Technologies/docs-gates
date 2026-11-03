# AGENTS.md — docs-gates

The docs-health battery gates as one standalone stdlib-only Go binary.
Extracted from the pbx-artmann deployment flake (`pbx-tools` docs gates)
2026-11-03; the extraction plan lives in that repo
(`docs/planning/2026-11-03_20-34_SUPERB-docs-gates-fleet-extraction.md`).
Public repo, MIT, module `github.com/Artmann-Technologies/docs-gates`.

## Critical constraints

- **stdlib-only is the module contract.** The ROOT go.mod carries zero
  requires — that is what keeps the nix build hermetic (`vendorHash =
  null`). The planned BuildFlow provider MUST be a separate go module
  (`pkg/provider` with its own go.mod requiring go-finding/toolsdk); it is
  consumed by BuildFlow's blank import, never by this binary.
- **JSON, not YAML, for config** (2026-11-03 decision): YAML would need
  go-faster/yaml and break the stdlib contract; JSON is jq-able, which the
  fleet tooling speaks. The tagliatelle lint rule pins snake_case json
  tags to match.
- **Exit-code contract: 0 clean / 1 findings / 2 usage-environment.**
  Everything funnels through internal/gatecli's constants; a verb that
  needs a new failure class maps it onto these three, never a new code.
- **Zero-value Settings never reach gate logic.** Every config knob rides
  a per-package Settings projection whose zero values fall back to the
  historical constants (docsfreshness.Settings, docsrouting's own,
  docscadence.SettingsFromConfig, config.Defaults). A repo WITHOUT a
  config file gets byte-for-byte the pre-extraction behavior — that is
  the parity guarantee. Tests rely on it; do not change a default without
  re-proving parity against the source fleet.
- **The five legacy names are API.** docs-freshness/docs-annotate/
  docs-routing/docs-cadence/docs-strike ride flake wrappers that exec
  `docs-gates <verb>`; each verb answers `--version` under its legacy
  name. Batteries and docs written against the old names keep working —
  never remove a wrapper without a fleet migration plan.
- **Parity discipline before any behavior change to a gate:** build this
  binary, run it over the source fleet's (pbx-artmann's) corpus, diff
  against that repo's in-tree gate binaries, expect ZERO delta. The
  planned golden fixture corpus (TODO parity-golden) makes this a check.
- **This repo has an auto-commit daemon** (commits ~every 3 min, pushes
  too): format and lint AT WRITE TIME, commit deliberate trains with the
  HEREDOC attribution form, and expect heuristic `chore: auto-commit`
  entries in the log. Never restore/checkout files you did not author.
- **Lint: the fleet golangci config runs at 0 findings.** Keep it there.
  Ported `//nolint:erraudit` directives are INERT until erraudit is
  adopted into the devShell (TODO erraudit-adopt) — adopt or strip, never
  add new ones.

## Non-obvious facts (paid for)

- **gatecli.CaptureOutput is a process-global stdout swap** — tests using
  it must never be t.Parallel (the paralleltest lint suppressions in
  battery_test.go are deliberate, opener-line form). Serial tests complete
  before parallel ones resume, so mixing serial capture tests with
  parallel plain tests is safe.
- **The battery's freshness + audit-all split is semantic, not
  redundant:** freshness checks ONLY the newest report (strictly, no
  ratchet); audit-all replays the whole corpus under the audit_max pin. A
  fixture wanting ratchet-only behavior needs a NEWER clean report, or
  freshness fails regardless of the pin (battery_test.go
  TestRunAuditMaxFromConfig paid for this).
- **docsstrike's target file comes from config** (`todo_file`) unless
  `--file` overrides; ParseArgs deliberately leaves Options.File empty.
- **Vars/params named `s` fail the fleet varnamelen config** when their
  scope grows — the Settings plumbing reads `settings`. Method receivers
  on Settings stay `s` (idiomatic, small scope, lint-clean).
- **`nix fmt` needs the tracked treefmt.toml** (fleet lesson: without it
  the formatter resolves a read-only store copy and formats nothing while
  exiting 0).
- **Darwin is deliberately absent** from the flake systems: current
  nixos-unstable no longer evaluates x86_64-darwin; no consumer rides
  macOS. Revisit only on a real consumer (ROADMAP).

## Commands

```console
nix build .#docs-gates -o /tmp/dg-result     # package; checkPhase runs go test ./...
nix develop -c go test ./...                 # the suite alone
nix develop -c golangci-lint run ./...       # the fleet config, 0 findings expected
nix develop -c golangci-lint fmt ./...       # golines/gofumpt/gci imports
nix fmt                                      # nixfmt-tree over *.nix (via treefmt.toml)
nix flake check --all-systems --no-build     # both-arch eval
go run ./cmd/docs-gates run <repo-root>      # the battery over any repo
```
