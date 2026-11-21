# AGENTS.md — docs-gates

The docs-health battery gates as one standalone stdlib-only Go binary,
extracted from the pbx-artmann deployment flake (`pbx-tools` docs gates)
2026-11-03 (plan in that repo's docs/planning/). Public, MIT, module
`github.com/Artmann-Technologies/docs-gates`.

**This repo dogfoods itself:** its living docs (TODO_LIST, ROADMAP,
CHANGELOG) are the corpus its gates judge — run the battery over this
repo after touching them (`go run ./cmd/docs-gates run .`) and keep the
rituals (strike landed rows, archive old reports). Full incident
narratives: `docs/agent-notes.md`.

## Layout and control flow

One binary, ten verbs; `cmd/docs-gates` dispatches (verb tables in `verbs.go`), everything else is
a library:

| Package                                                    | Role                                                                                                                                                                                                                                |
| ---------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cmd/docs-gates`                                           | verb dispatch (`run`, `freshness`, `annotate`, `routing`, `cadence`, `strike`, `purge`), verb-to-legacy-name map, help surface (`--help`/`-h`/`help [verb]`, did-you-mean), `--version` (ldflags-stamped, `"dev"` bare)             |
| `internal/gatecli`                                         | exit-code constants, boolean/valued `--flag` parsing (`CutIntFlag`), the config prologue (`LoadConfig`), repo-root resolution, `Outf`/`Errf` + kin (the ONE home for unchecked `fmt.Fprint*`), `Truncate`, `CaptureOutput`, `Livef` |
| `internal/battery`                                         | the `run` verb: ordered gates from config, run INDEPENDENTLY with captured output, PASS/FAIL/SKIP matrix — no first-failure abort                                                                                                   |
| `internal/config`                                          | `.docs-gates.json` loader: decode over `Defaults()`, unknown keys WARN, malformed/invalid exit 2                                                                                                                                    |
| `internal/docs{freshness,annotate,routing,cadence,strike}` | the five historical gates: each a `Run([]string) int` whose argv matches the historical standalone binaries, each with a `Settings` projection (zero values → historical constants) + docs-purge                                               |
| `internal/docs{stub,citations,stamps}` | the three content gates over the domain-language glossary (2026-11-21 census): bare-template checksums, code-citation resolution, verification-stamp age bars; warn-mode rollout knobs (`*_enforce`), same `Run`/`Settings` shape                                   |
| `pkg/provider`                                             | BuildFlow toolsdk provider — a SEPARATE go module that EXECS the binary from PATH, parses `run --json` into findings (strict decode), registers via `toolsdk.Register`                                                              |

Control flow: `main.go` → verb's `Run(rest)` → `gatecli.Parse` →
`Settings` → gate logic → verdict lines → exit code. `--version`/`--help`
are intercepted BEFORE any gate parser; help is pinned per gate by
usage_test.go. Usage FATALs ride `gatecli.FatalUsage`; config failures
keep bare `FATAL: %v`.

## Critical constraints

- **stdlib-only is the root module contract** — zero requires
  (`scripts/stack-drift-check.sh` fails CI on drift); keeps the nix
  build hermetic. Third-party deps live in a separate go module with a
  STACK.md row.
- **JSON, not YAML, for config** (2026-11-03): stdlib contract, jq-able;
  tagliatelle pins snake_case json tags.
- **Exit codes: 0 clean / 1 findings / 2 usage-environment** — everything
  funnels through gatecli constants; new failure classes map onto these.
- **Zero-value Settings never reach gate logic** — projections fall back to
  the historical constants (`cmp.Or`, `OrSlice`, `> 0`); config-less repos
  behave byte-for-byte like pre-extraction. A default change re-proves parity.
- **The five legacy names are API** (the three content gates carry no legacy wrapper): flake wrappers exec
  `docs-gates <verb>`; each answers `--version` under its legacy name.
  Never remove a wrapper without a migration plan.
- **Parity is enforced**: `TestParityGolden` pins byte output of
  fifteen invocations (the five gates + battery + battery-json + the
  margins/dump-items/preflight arms, incl. intended non-clean exits)
  over `testdata/parity/corpus`. INTENTIONAL change re-pins via
  `PARITY_UPDATE=1 go test ./internal/battery/ -run TestParityGolden`,
  then hand-diff — a golden bump is a contract change.
- **Auto-commit daemon** (~2-3 min cadence, NEVER pushes — PUSH YOUR OWN
  TRAINS): format/lint at write time, commit trains with the HEREDOC
  attribution form, expect heuristic `chore: auto-commit` entries. Never
  restore/checkout files you did not author; NEVER `git revert HEAD`
  blind — explicit SHA only (paid 2026-11-04); pin your baseline (`git
  rev-parse HEAD`) before before/after diffs — HEAD moves mid-session.
- **Lint: fleet golangci config at 0 findings.** `//nolint:erraudit`
  directives are LIVE, guarded by the `nolint-audit` (staleness) and
  `erraudit-lint` (new migratable site) flake checks; golangci's runner
  warning about the unknown nolint linter is expected. `erraudit lint
  ./... --max` green since 2026-11-12: generic-error returns carry typed
  refusal carriers (unexported struct + `newXxxError` + `Unwrap`; the
  idiomatic `error` return KEPT — concrete `*T` is the typed-nil
  landmine). Never cite a directive COUNT from memory — run
  `erraudit nolint-audit .`

## Verdict arithmetic (battery)

- Exits 1 iff ≥1 gate row FAILs; SKIP rows exit-clean; every gate always
  runs (one foreign red never masks later gates). A gate exiting 2
  INSIDE the battery is a FAIL row `(rc=2)`; only the battery's own
  failures exit 2; findings live in the gate's output.
- `--json`: one snake_case JSON document on stdout (`duration_ms` ints,
  gate streams verbatim), humans on stderr; byte-pinned by the
  battery-json golden, the provider's strict decoder turns row-shape
  drift into findings — a knob means re-pin golden AND provider.
- Half the `gate_timeout_sec` budget triggers the watchdog line via
  `gatecli.Livef` (JSON face: stderr only) — informational.

## Non-obvious facts (paid for)

- **Pre-flight before ANY living-docs edit batch** (`docs-freshness
  --pre-flight`, stdin candidates, SAFE only, battery BEFORE the commit —
  paid twice; universe is exactly the trio; the refusal is a lowercased
  substring test over the joined corpus, so rewrapped removed lines
  still collide until the rewrite lands).
- **Root `./...` never sees `pkg/provider`** — test/lint it from its own
  directory. Floors: root go 1.26, provider 1.27; the flake pins ONE
  toolchain (go_1_27 + GOTOOLCHAIN=local, one `goPkgAttr` binding —
  never fork; `go-floors` asserts).
- **gatecli.CaptureOutput is a process-global stdout swap** — its tests must never be t.Parallel (deliberate suppressions).
- **Freshness = newest report only (strict); audit-all = whole corpus
  under the audit_max pin.** Ratchets are pins, not floors (pass only
  when >= 0; tolerated findings print RATCHET).
- **docsstrike's target file comes from config** (`todo_file`); vars named `s` fail
  varnamelen when scope grows — plumbing reads `settings`, receivers stay `s`.
- **Report vocabulary is shared, never re-copied**: docsfreshness exports
  `LivingDocsLines`/`SectionFRe`/`HeadingRe`/`Tokens`/`Squash`/`GenericDF`
  — one definition keeps harvest and triage aligned on §f.
- **ExtractItems treats any line-start `\d+[.)]` as an item boundary** —
  keep report continuation wraps away from a leading number+paren.
- **Known split brains** (2026-11-15): the marker grammar (writer renders
  `markerFor`, judge matches `markerRe`, bound by `internal/docsannotate/
  markers_test.go`) and the docs-health binary (deleted; callers `go run`).
  See `docs/adr/0001-marker-grammar-ownership.md` and the split plan in
  `docs/planning/2026-11-15_marker-grammar-and-tooling-split.md`.
- **The five `//art-dupl:accept` directives are the dedup baseline**
  (`-t 1`: 0 groups). The settings-projection pair is ACCEPTED —
  domain-specific Settings; a generic helper would couple gates.
- **`gatecli.Truncate` stays hand-rolled** (non-fix): the canonical
  candidate breaks the stdlib-only contract; goldens pin its semantics.
- **`nix fmt` needs the tracked treefmt.toml.** **Darwin is deliberately absent** from flake systems (revisit on a real consumer).
- **erraudit rides the flake as a PRIVATE git+ssh input** tracking
  master (rev pinned in flake.lock; its OWN hermetic package). Probes
  may SILENTLY SKIP where the binary is absent (CI) — prove locally;
  flake-source probes must STAGE plants (`git add -N`).
- **The erraudit-rev tripwire** (`scripts/check-erraudit-rev.sh`): a
  per-input pin table (erraudit, go-filewatcher, go-nix-helpers,
  gogenfilter) — every lock rev must equal its expected rev; updates
  bump silently and rules ARE behavior — check on the new rev, THEN
  bump one input at a time.
- **Buildflow's repair phase rewrites lockfiles; the daemon wraps the
  damage** (paid 11-06 AND 11-14 — six inputs). Recovery: forward
  `git restore --source=<good-sha> -- flake.lock` + OWN commit, tripwire
  before trusting; hash fixes happen in erraudit via
  `buildflow -s nix-hash-fix --fix` (done 11-14), never here.
- **Stale buildflow binary is the #1 silent root cause** (11-14: masked
  upstream fixes for hours) — `buildflow doctor`, rebuild + reinstall.
- **The nix-fmt/treefmt roundtrip wedge is FIXED upstream** (2026-11-14;
  sandbox git-inits, staged files, lock rides along).
- **docs-gates flake checks repaired 2026-11-14** (unpassable since
  11-04, vendorHash-masked): golangci-drift counted pins after sort -u;
  shell-lint ran bare actionlint; BuildFlow's `inputs.<name>.url`
  false-FAIL fixed upstream — do not reformat flake.nix to dodge it.
- **The docs-battery pre-commit leg rides REAL COMMITS ONLY** (devShell;
  c1b364d): git-hooks' sandbox git-inits a synthetic one-commit repo, so
  verdict-hash evidence is red by construction there; CI's dogfood job
  is the push-time leg.
- **Buildflow's managed `.gitignore` block swallows ANY `reports/` dir**
  (fixtures unstaged → CI golden-drift-only failures). After adding
  testdata: `git check-ignore -v` (`--no-index` for TRACKED files);
  `git ls-files` is cwd-relative.
- **Pre-commit legs** (`preCommitCheckFor`): nolint-audit everywhere,
  docs-battery devShell-only. Auto-install needs the (dead-global)
  `core.hooksPath` unset; `.pre-commit-config.yaml` is a generated
  store symlink — never commit.
- **go1.26+ rejects mixing literals with a slice spread in one variadic
  call** (`f("a", "b", slice...)`): build the argv with `append`.
- **LSP ghost diagnostics are a CACHE artifact** (host wrapper's pinned
  analysis cache; survives lsp_restart): prove with real gates, then
  clear the wrapper caches and restart; do not file upstream.
- **Striking ledger rows by ROW NUMBER is a corruption machine**: use
  `scripts/safe-strike.sh` (unique ≥4-letter anchors, dry-run default,
  `--gate`/`--battery` per batch, `--batch` for multi-anchor sweeps).
- **BuildFlow#87 is FIXED upstream** (live-proven 2026-11-14) — no
  result-cache bypass; dprint bundle rides `nix build .#dprint-plugins` +
  BUILDFLOW_DPRINT_PLUGINS (SystemNix home-base.nix).
- **dprint-managed tables × pre-flight: judge the FORMATTED bytes against
  the PRE-edit tree** — the formatter re-aligns cells AFTER the verdict,
  so the landed bytes never match a typed draft's verdict (paid
  2026-11-14); post-landing re-pre-flight hits the on-disk refusal.
- **go-auto-upgrade warnings are deliberate non-fixes**: samber/lo
  withheld (stdlib-only root; thin provider shim); jsonv1tov2 declined
  (encoding/json v1 ON PURPOSE — `DisallowUnknownFields` is the row-shape
  contract, byte-pinned); the six cross-file-wrapper warnings ARE the
  accepted settings-projection pair.
- **The --margins near-bar window is `[bar-2, bar)`**, counts LIVING-DOC LINES ONLY; the
  corpus-replace fixture's `generic_df: 2` is load-bearing twice; headings are NOT excluded.
- **The wedge class was the capture-pipe deadlock** (fixed by concurrent
  drainers; `gate_timeout_sec` bounds it; a new wedge = suspect a NEW park).
- **Size a foreign corpus BEFORE launching the battery** (`ls docs/status
  | wc -l`). Memoized replays cost one scan per DISTINCT token (hours →
  49s); audit-all prints its estimate; the cap is a circuit breaker.
- **Gate verdicts are point-in-time under foreign daemons** — timestamp the run; one run is not steady state.

- **Testing**: stdlib `testing` only; gates are exercised through their
  real `Run` functions with `gatecli.CaptureOutput`, never mocked
  output. Two parity layers: byte goldens + the defaults-are-historical
  invariant. Corpus fixtures use a `{{DATE}}` token substituted with the
  run date, so report ages stay stable.

## Commands

```console
buildflow --build-mode dev                        # the quality loop: lint+format+test (BuildFlow owns it; .buildflow.yml skips the JS/TS + docstring steps by intent)
nix build .#docs-gates -o /tmp/dg-result          # package; checkPhase runs go test ./...
nix develop -c go test ./...                      # root-module suite (provider excluded — see gotcha)
(cd pkg/provider && nix develop -c go test ./...) # provider module's own suite
(cd pkg/provider && nix develop -c golangci-lint run ./...) # provider lint — root ./... never sees the nested module
nix develop -c golangci-lint run ./...            # the fleet config, 0 findings expected
nix develop -c golangci-lint fmt ./...            # golines/gofumpt/gci imports
nix develop -c erraudit lint ./... --type-aware   # error-handling audit, rc=0 expected (//nolint:erraudit directives live)
nix develop -c erraudit nolint-audit .            # directive staleness over the root module (also the nolint-audit flake check)
nix develop -c pre-commit run nolint-audit --all-files   # the commit-time leg (auto-install needs `git config --unset core.hooksPath`)
nix develop -c pre-commit run docs-battery --all-files  # the commit-time dogfood leg: the whole battery when living docs are staged
go run ./cmd/docs-gates purge <repo-root>          # the board purge as one gated command (dry-run default; --apply writes)
bash scripts/check-erraudit-rev.sh                # erraudit-rev tripwire (also a flake check): run after bumping the erraudit input
nix fmt                                           # nixfmt-tree over *.nix (via treefmt.toml)
nix flake check --all-systems --no-build          # both-arch eval (checks: nolint-audit, erraudit-lint, provider-suite, golangci-drift, go-floors, shell-lint + the skills guards)
timeout 300 go run ./cmd/docs-gates run <repo-root>  # battery over any repo (run over THIS repo after doc edits)
go run ./cmd/docs-gates run <repo-root> --json    # the machine face: one JSON verdict document on stdout, humans on stderr
timeout 300 go run ./cmd/docs-gates <verb> <repo-root>  # a standalone gate (standalone runs have NO cap of their own)
PARITY_UPDATE=1 go test ./internal/battery/ -run TestParityGolden   # re-pin goldens after an INTENTIONAL output change
```

CI (`.github/workflows/ci.yml`, push/PR): root `go` job (build/vet/
stack-drift/`go test ./...` + the audit-all cliff benchmark arm), a
`provider` job, and the `dogfood` job (the battery over THIS repo every
push; golangci gates format on both jobs). The flake never runs on CI
(local-only). `golangci-drift` keeps the ci.yml action version and the
flake's nixpkgs golangci pin in one train; `shell-lint` actionlints with
explicit paths; `go-floors` guards the go.mod floors; dependabot
(grouped, weekly) keeps action SHA pins fresh.
