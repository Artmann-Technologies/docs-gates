# AGENTS.md — docs-gates

The docs-health battery gates as one standalone stdlib-only Go binary.
Extracted from the pbx-artmann deployment flake (`pbx-tools` docs gates)
2026-11-03; the extraction plan lives in that repo
(`docs/planning/2026-11-03_20-34_SUPERB-docs-gates-fleet-extraction.md`).
Public repo, MIT, module `github.com/Artmann-Technologies/docs-gates`.

**This repo dogfoods itself:** its living docs (TODO_LIST, ROADMAP,
CHANGELOG) are exactly the corpus its gates judge. Run the battery over
this repo after touching the docs — `go run ./cmd/docs-gates run .` — and
keep the rituals (strike landed rows, archive old reports) or the tool
goes red over the repo that defines green.

## Layout and control flow

One binary, six verbs; `cmd/docs-gates` dispatches, everything else is a
library:

| Package                                                    | Role                                                                                                                                                                                                                                                              |
| ---------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cmd/docs-gates`                                           | verb dispatch (`run`, `freshness`, `annotate`, `routing`, `cadence`, `strike`), the verb-to-legacy-name map, `--version`. Version is ldflags-stamped by the flake build; `"dev"` bare-built.                                                                      |
| `internal/gatecli`                                         | shared CLI scaffolding: exit-code constants, boolean `--flag` parsing, repo-root resolution, `Outf`/`Outln`/`Errf`/`Errln` (the ONE home for unchecked `fmt.Fprint*`), UTF-8-safe `Truncate`, `CaptureOutput`                                                     |
| `internal/battery`                                         | the `run` verb: ordered gate list from config, gates run INDEPENDENTLY with captured output, PASS/FAIL/SKIP verdict matrix — no first-failure abort                                                                                                               |
| `internal/config`                                          | `.docs-gates.json` loader: decode over `Defaults()` (the historical fleet constants), unknown keys WARN, malformed/invalid exit 2                                                                                                                                 |
| `internal/docs{freshness,annotate,routing,cadence,strike}` | the five gates: each a `Run([]string) int` whose argv shape matches the historical standalone binaries, each with a per-package `Settings` projection (zero values → historical constants)                                                                        |
| `pkg/provider`                                             | BuildFlow toolsdk provider — a SEPARATE go module that EXECS the `docs-gates` binary from PATH (internal/ is unimportable across module boundaries), parses the verdict matrix into findings, registers via `toolsdk.Register` at init; consumers blank-import it |

Gate-verb control flow: `main.go` → verb's `Run(rest)` → `gatecli.Parse`
(flags + root) → `Settings` projection → gate logic → verdict lines via
gatecli → one of the three exit codes.

`--version` is intercepted in `cmd/docs-gates` BEFORE any gate parser
sees argv: the parsers reject unknown flags and `--version` would die as
one.

## Critical constraints

- **stdlib-only is the root module contract.** The ROOT go.mod carries
  zero requires — that is what keeps the nix build hermetic (`vendorHash
  = null`) and what `scripts/stack-drift-check.sh` enforces (CI fails the
  moment the root go.mod grows a require/replace/exclude). Third-party
  deps go in a SEPARATE go module under this repo, with a STACK.md row.
- **JSON, not YAML, for config** (2026-11-03 decision): YAML would need
  go-faster/yaml and break the stdlib contract; JSON is jq-able, which
  the fleet tooling speaks. The tagliatelle lint rule pins snake_case
  json tags to match.
- **Exit-code contract: 0 clean / 1 findings / 2 usage-environment.**
  Everything funnels through internal/gatecli's constants; a verb that
  needs a new failure class maps it onto these three, never a new code.
- **Zero-value Settings never reach gate logic.** Every config knob rides
  a per-package Settings projection whose zero values fall back to the
  historical constants (docsfreshness.Settings, docsrouting's own,
  docscadence.SettingsFromConfig, config.Defaults). A repo WITHOUT a
  config file gets byte-for-byte the pre-extraction behavior — that is
  the parity guarantee. Tests rely on it; do not change a default without
  re-proving parity.
- **The five legacy names are API.** docs-freshness/docs-annotate/
  docs-routing/docs-cadence/docs-strike ride flake wrappers that exec
  `docs-gates <verb>`; each verb answers `--version` under its legacy
  name. Batteries and docs written against the old names keep working —
  never remove a wrapper without a fleet migration plan.
- **Parity is enforced, not aspirational.** `TestParityGolden` pins the
  exact byte output of six invocations (five gates + battery) over the
  fixture corpus `testdata/parity/corpus`; goldens live in
  `testdata/parity/golden`. An INTENTIONAL output change re-pins via the
  env var (not a flag): `PARITY_UPDATE=1 go test ./internal/battery/
  -run TestParityGolden`, then diff by hand — a golden bump is a contract
  change.
- **This repo has an auto-commit daemon** (commits ~every 3 min, pushes
  too): format and lint AT WRITE TIME, commit deliberate trains with the
  HEREDOC attribution form, and expect heuristic `chore: auto-commit`
  entries in the log. Never restore/checkout files you did not author —
  and verify `git status` before claiming the daemon committed your work
  (it has a known blind spot).
- **Lint: the fleet golangci config runs at 0 findings.** Keep it there.
  Ported `//nolint:erraudit` directives are INERT until erraudit is
  adopted into the devShell (TODO erraudit-adopt) — adopt or strip, never
  add new ones.

## Verdict arithmetic (battery)

- `docs-gates run` exits 1 iff at least one gate row is FAIL. SKIP rows
  (`skip_gates`) are exit-clean and never count. Every gate always runs —
  the matrix exists so one foreign red never masks the later gates (the
  2026-10-30 failure class the shell battery was rebuilt for).
- A gate exiting 2 INSIDE the battery becomes a FAIL row carrying
  `(rc=2)`: the battery as a whole still exits 1. Only the battery's OWN
  failures (bad root, bad config) exit 2 from `docs-gates run`.
- Findings live in the gate's own output (stdout on green, forwarded to
  stderr on red); the matrix names gates, not findings.

## Non-obvious facts (paid for)

- **The root `./...` never sees `pkg/provider`** — it is a nested go
  module. `go test ./...`, `go vet ./...`, the nix checkPhase, and CI
  (which runs all of those from the root) all exclude it. Test the
  provider from its own directory: `(cd pkg/provider && go test ./...)`.
  Its go.mod pins go 1.27 (the root pins 1.26); the flake pins one
  toolchain (go_1_27) over BOTH surfaces so the devShell runs its suite
  hermetically.
- **The flake pins go_1_27 + GOTOOLCHAIN=local everywhere (build,
  devShell)** — go_1_26 (the root floor) would reject pkg/provider's
  `go 1.27` directive under local, and WITHOUT local go's default `auto`
  mode silently downloads go1.27 into the module cache at dev time
  (network fetch inside `nix develop`, verified 2026-11-04: local +
  go1.26.10 fails `go.mod requires go >= 1.27`). A future module floor
  past the pinned line now fails LOUDLY instead of fetching. The toolchain
  rides one `goPkgAttr` binding shared by buildGoModule.override and the
  devShell — do not fork it.
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
- **Ratchets are pins, not floors:** `audit_max` and `annotate_max` are
  passed to the gate only when >= 0 (the strict -1 default omits the
  flag entirely, matching the gate CLIs' own semantics). Tolerated
  findings print as RATCHET, not PASS.
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
- **Known BuildFlow tool bug (2026-11-04, do not "fix" here):** the
  BuildFlow `nix-fmt`/`treefmt` roundtrip steps fail over THIS repo with
  `failed to create walker: path … not inside the tree root /nix/store`.
  In-repo `nix fmt` and `nix flake check` are green — the defect is the
  roundtrip copy's environment (its fresh `flake.lock` and tree-root
  resolution), not this flake. Diagnosis lives in the BuildFlow repo
  (`tools/providers/format_roundtrip.go` +
  `nixFmtRoundTripConfigFiles`); the failing steps do not gate anything
  this repo's own commands miss.

## Testing

- stdlib `testing` only (no testify — the stdlib-only contract),
  table-driven where it fits; gates are exercised through their real
  `Run` functions with `gatecli.CaptureOutput`, never via mocked output.
- Two parity layers: the golden fixtures (exact byte output, see Critical
  constraints) and the defaults-are-historical invariant (config-less
  repos reproduce pre-extraction behavior).
- The corpus fixtures use a `{{DATE}}` token in filenames and contents,
  substituted with the run date, so report ages stay stable across time —
  both in the corpus and in the goldens.

## Commands

```console
buildflow --build-mode dev                        # the quality loop: lint+format+test (BuildFlow owns it; .buildflow.yml skips the JS/TS + docstring steps by intent)
nix build .#docs-gates -o /tmp/dg-result          # package; checkPhase runs go test ./...
nix develop -c go test ./...                      # root-module suite (provider excluded — see gotcha)
(cd pkg/provider && nix develop -c go test ./...) # provider module's own suite
nix develop -c golangci-lint run ./...            # the fleet config, 0 findings expected
nix develop -c golangci-lint fmt ./...            # golines/gofumpt/gci imports
nix fmt                                           # nixfmt-tree over *.nix (via treefmt.toml)
nix flake check --all-systems --no-build          # both-arch eval
go run ./cmd/docs-gates run <repo-root>           # the battery over any repo (run it over THIS repo after doc edits)
PARITY_UPDATE=1 go test ./internal/battery/ -run TestParityGolden   # re-pin goldens after an INTENTIONAL output change
```

CI (`.github/workflows/ci.yml`, push/PR): `go build` / `go vet` /
`gofmt -l` / `scripts/stack-drift-check.sh` / `go test ./...` — root
module only, so provider regressions surface only in a local run from
`pkg/provider/`.
