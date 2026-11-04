# STACK.md — docs-gates decision record

Selected 2026-11-03 at repo inception (extraction train). Re-align only
via this file; a dependency that appears without a row here is drift
(`scripts/stack-drift-check.sh` fails the build on it).

## Foundation

| Concern   | Choice                                                                           | Owner            |
| --------- | -------------------------------------------------------------------------------- | ---------------- |
| Language  | Go 1.26 root floor (go.mod `go 1.26`); flake pins go_1_27 + GOTOOLCHAIN=local over BOTH modules (pkg/provider pins go 1.27) | how-to-golang    |
| Modules   | ZERO requires in the root go.mod (stdlib-only)    | this repo, ADR   |
| CLI shape | internal/gatecli (Parse/HandleVersion/Outf)       | this repo        |
| Errors    | one static type per package concern (+%w wraps)   | fleet go policy  |
| Config    | stdlib encoding/json, snake_case keys             | this repo, ADR   |
| Build     | flake.nix buildGoModule, vendorHash=null          | fleet nix policy |
| Lint      | fleet .golangci.yml, 0-findings bar               | BuildFlow        |
| Format    | gofmt + golangci-lint fmt (golines 120)           | BuildFlow        |
| Tests     | stdlib testing, table-driven, capture via gatecli | fleet go policy  |

## ADR: stdlib-only root module

JSON over YAML (no go-faster/yaml); no testify (stdlib testing); no
samber/lo (the stdlib2lo suggester findings are advisory noise); slices/
maps stdlib helpers over external libraries. Consequence: `vendorHash =
null` and a fully hermetic nix build. Escape hatch: a SEPARATE go module
under this repo (the planned `pkg/provider`) may require third-party
packages — the root stays clean (realized as `pkg/provider`, its own
go.mod pinning go 1.27).

## First-party SDK inventory

go-finding v1.15.0 + go-finding/toolsdk v1.16.0, consumed ONLY by the
pkg/provider submodule. A dependency that appears anywhere else is drift:
the root go.mod must stay require-free (`scripts/stack-drift-check.sh`
fails CI on it). Source-fleet provenance: the five gate packages are
ported from pbx-artmann's internal/{gatecli,docs*} — treat pbx-artmann's
docs-health conventions as the spec, this repo as the implementation.
