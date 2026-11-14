# AGENTS.md — docs-gates

The docs-health battery gates as one standalone stdlib-only Go binary.
Extracted from the pbx-artmann deployment flake (`pbx-tools` docs gates)
2026-11-03; plan lives in that repo
(`docs/planning/2026-11-03_20-34_SUPERB-docs-gates-fleet-extraction.md`).
Public repo, MIT, module `github.com/Artmann-Technologies/docs-gates`.

**This repo dogfoods itself:** its living docs (TODO_LIST, ROADMAP,
CHANGELOG) are exactly the corpus its gates judge. Run the battery over
this repo after touching the docs — `go run ./cmd/docs-gates run .` — and
keep the rituals (strike landed rows, archive old reports) or the tool
goes red over the repo that defines green.

## Layout and control flow

One binary, seven verbs; `cmd/docs-gates` dispatches, everything else is a
library:

| Package                                                    | Role                                                                                                                                                                                                          |
| ---------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cmd/docs-gates`                                           | verb dispatch (`run`, `freshness`, `annotate`, `routing`, `cadence`, `strike`, `purge`), verb-to-legacy-name map, help surface (`--help`/`-h`/`help [verb]`, did-you-mean), `--version` (ldflags-stamped, `"dev"` bare) |
| `internal/gatecli`                                         | shared CLI scaffolding: exit-code constants, boolean/valued `--flag` parsing (`CutIntFlag`), the config prologue (`LoadConfig`), repo-root resolution, `Outf`/`Outln`/`Errf`/`Errln` (the ONE home for unchecked `fmt.Fprint*`), UTF-8-safe `Truncate`, `CaptureOutput`, `Livef` |
| `internal/battery`                                         | the `run` verb: ordered gates from config, run INDEPENDENTLY with captured output, PASS/FAIL/SKIP matrix — no first-failure abort                                                                                      |
| `internal/config`                                          | `.docs-gates.json` loader: decode over `Defaults()`, unknown keys WARN, malformed/invalid exit 2                                                                                                              |
| `internal/docs{freshness,annotate,routing,cadence,strike}` | the five gates: each a `Run([]string) int` whose argv matches the historical standalone binaries, each with a `Settings` projection (zero values → historical constants) + docs-purge                                       |
| `pkg/provider`                                             | BuildFlow toolsdk provider — a SEPARATE go module that EXECS the `docs-gates` binary from PATH, parses `run --json` into findings (strict decode: unknown fields fail, drift becomes findings), registers via `toolsdk.Register` |

Gate-verb control flow: `main.go` → verb's `Run(rest)` → `gatecli.Parse`
→ `Settings` projection → gate logic → verdict lines → exit code.
`--version`/`--help` are intercepted in `cmd/docs-gates` BEFORE any gate
parser (parsers reject unknown flags; both would die as one); help is
pinned per gate by usage_test.go against the parser's flag vocabulary,
both directions — a help that lies is worse than none. Usage-class
FATALs ride `gatecli.FatalUsage` (FATAL line + `--help` pointer);
config/environment failures keep the bare `FATAL: %v` shape. The parity
goldens pin none of the help/FATAL-refusal paths.

## Critical constraints

- **stdlib-only is the root module contract** — root go.mod carries ZERO
  requires (`scripts/stack-drift-check.sh` fails CI on require/replace/
  exclude); that is what keeps the nix build hermetic (`vendorHash =
  null`). Third-party deps go in a separate go module with a STACK.md row.
- **JSON, not YAML, for config** (2026-11-03): YAML would break the
  stdlib contract; JSON is jq-able for fleet tooling. tagliatelle pins
  snake_case json tags.
- **Exit-code contract: 0 clean / 1 findings / 2 usage-environment.**
  Everything funnels through internal/gatecli constants; new failure
  classes map onto these three, never a new code.
- **Zero-value Settings never reach gate logic** — every knob rides a
  per-package Settings projection whose zero values fall back to the
  historical constants. String fallbacks spell
  `cmp.Or(override, config.Defaults().X)`, list fallbacks
  `config.OrSlice(...)` (the one-line convention, 2026-11-04); int
  fallbacks with `> 0` semantics stay explicit if-forms (cmp.Or would
  admit negatives). A repo WITHOUT a config file gets byte-for-byte the
  pre-extraction behavior — tests rely on it; never change a default
  without re-proving parity.
- **The five legacy names are API** (docs-freshness/docs-annotate/
  docs-routing/docs-cadence/docs-strike): flake wrappers exec
  `docs-gates <verb>`; each verb answers `--version` under its legacy
  name. Never remove a wrapper without a fleet migration plan.
- **Parity is enforced, not aspirational**: `TestParityGolden` pins exact
  byte output of six invocations (five gates + battery) over
  `testdata/parity/corpus`, goldens in `testdata/parity/golden`.
  INTENTIONAL output change re-pins via env (not a flag):
  `PARITY_UPDATE=1 go test ./internal/battery/ -run TestParityGolden`,
  then diff by hand — a golden bump is a contract change.
- **Auto-commit daemon** (~2-3 min cadence per host config, NEVER pushes
  — every push in this repo's history was a manual session push, so PUSH
  YOUR OWN TRAINS; the "pushes too" belief once cost a 22-commit surprise):
  format/lint AT WRITE TIME, commit deliberate trains with the HEREDOC
  attribution form, expect heuristic `chore: auto-commit` entries. Never
  restore/checkout files you did not author; verify `git status` before
  claiming the daemon committed (known blind spot). NEVER `git revert
  HEAD` blind: between your push and the revert the daemon may have
  wrapped your work — `git log -1 --stat` first, revert by explicit SHA,
  commit your own work immediately (paid 2026-11-04: a blind revert
  removed a test commit instead of its CI target). Pin your verification
  baseline (`git rev-parse HEAD`) BEFORE any before/after diff — HEAD
  moves under you mid-session.
- **Lint: the fleet golangci config runs at 0 findings.** The
  `//nolint:erraudit` directives are LIVE, guarded by two flake checks:
  `nolint-audit` (fails on stale directives) and `erraudit-lint`
  (erraudit's type-aware legacyerrors audit over the root module — a NEW
  migratable site fails `nix flake check`). golangci's runner WARNING
  about `erraudit` being an unknown nolint linter is expected (standalone
  binary, not a plugin). The strict `erraudit lint ./... --max` audit
  (`generic_return` enforced) is green over both modules since
  2026-11-12: flagged generic-error returns carry typed refusal carriers
  (unexported struct + `newXxxError` constructor + `Unwrap`; idiomatic
  `error` return KEPT — a concrete `*T` return is the typed-nil
  landmine); keep the pattern for new refusals. Never add a directive
  without a code-adjacent reason, and never cite a directive COUNT from
  memory — run `erraudit nolint-audit .`.

## Verdict arithmetic (battery)

- `docs-gates run` exits 1 iff at least one gate row is FAIL. SKIP rows
  (`skip_gates`) are exit-clean. Every gate always runs — the matrix
  exists so one foreign red never masks later gates. A gate exiting 2
  INSIDE the battery becomes a FAIL row carrying `(rc=2)`; only the
  battery's OWN failures (bad root, bad config) exit 2. Findings live in
  the gate's own output (stdout on green, stderr on red).
- The `--json` machine face puts EXACTLY one JSON document on stdout
  (snake_case tags, `duration_ms` integers, per-gate stdout/stderr
  verbatim); humans on stderr; exit contract unchanged. The battery-json
  parity golden pins it byte-for-byte, and pkg/provider's strict decoder
  turns row-shape drift into findings, never silent zero values. Add a
  knob = intentional contract change: re-pin golden AND provider tests.
- A gate burning past HALF its `gate_timeout_sec` budget triggers the
  live watchdog line (`gate X over half its <cap> budget — abandoned at
  the cap`): human face prints on the REAL stdout mid-capture via
  `gatecli.Livef`; JSON face keeps stdout pure, warns on stderr.
  Fire-only — the matrix never reads it.

## Non-obvious facts (paid for)

- **Run `docs-freshness --pre-flight` BEFORE ANY living-docs edit batch**
  (TODO_LIST/ROADMAP/CHANGELOG — paid TWICE, 2026-11-05/06): stdin the
  candidate lines, rewrite only on SAFE, run the dogfood battery BEFORE
  the commit. Its universe is exactly that trio (an AGENTS.md/FEATURES.md
  edit has no pre-flight surface; gates via battery + dprint only). The
  on-disk refusal is a lowercased SUBSTRING test over the joined corpus
  blob — rewrapped reuse of removed lines still collides until the
  rewrite lands through --replace targets or genuinely new wording.
- **The root `./...` never sees `pkg/provider`** (nested module): test
  and lint it from its own directory. Its go.mod pins go 1.27 (root pins
  1.26); the flake pins ONE toolchain (go_1_27) over both surfaces.
- **The flake pins go_1_27 + GOTOOLCHAIN=local everywhere** — go_1_26
  would reject the 1.27 floor under local, and without local, `auto`
  silently downloads toolchains at dev time. The toolchain rides one
  `goPkgAttr` binding shared by buildGoModule.override and the devShell —
  do not fork it; `go-floors` check asserts both floors stay within pin.
- **gatecli.CaptureOutput is a process-global stdout swap** — tests using
  it must never be t.Parallel (the paralleltest suppressions are
  deliberate, opener-line form). Serial tests complete before parallel
  ones resume, so mixing is safe.
- **The battery's freshness + audit-all split is semantic**: freshness
  checks ONLY the newest report (strict); audit-all replays the whole
  corpus under the audit_max pin. A fixture wanting ratchet-only behavior
  needs a NEWER clean report, or freshness fails regardless.
- **Ratchets are pins, not floors**: `audit_max`/`annotate_max` pass to
  the gate only when >= 0 (strict -1 omits the flag). Tolerated findings
  print as RATCHET, not PASS.
- **docsstrike's target file comes from config** (`todo_file`) unless
  `--file` overrides; ParseArgs deliberately leaves Options.File empty.
- **Vars named `s` fail the fleet varnamelen config when scope grows** —
  the Settings plumbing reads `settings`; Settings method receivers stay
  `s` (idiomatic, small, lint-clean).
- **Report-vocabulary sharing**: docsfreshness exports the shared
  status-report definitions docsrouting consumes (`LivingDocsLines`,
  `SectionFRe`, `HeadingRe`, plus `Tokens`/`Squash`/`GenericDF`). Never
  re-copy them in a second gate — one definition keeps freshness harvest
  and routing triage from disagreeing about §f boundaries or the corpus.
- **freshness ExtractItems treats any line-start `\d+[.)]` as an item
  boundary** (proseRe): a wrapped report line beginning `4211653)` splits
  one §f item in two (burned 2026-11-04). Keep continuation wraps away
  from a leading number+paren.
- **The five `//art-dupl:accept` directives** (docsannotate items.go ×2,
  spacing.go ×1, settings.go ×2) are the dedup baseline: at `-t 1` the
  art-dupl report must show 0 groups. A clone gets extracted or
  accepted-in-source with rationale; multi-word text after the marker is
  a bare accept. Extracted 2026-11-04: `gatecli.CutIntFlag`,
  `gatecli.LoadConfig`, `docsfreshness.LivingDocsLines`,
  `SectionFRe`/`HeadingRe`, `cmp.Or`/`config.OrSlice`. The
  settings-projection pair (loadSettings + cmp.Or accessors duplicated
  across gate packages) is ACCEPTED — each gate's Settings is
  domain-specific; a generic plumbing helper would couple gates for six
  lines.
- **`gatecli.Truncate` stays a hand-rolled 10-liner** (deliberate
  non-fix): the canonical candidate (go-output/nom.TruncateVisible)
  breaks the stdlib-only contract and its semantics differ anyway — and
  the parity goldens pin Truncate's byte-budget cut byte-for-byte.
- **`nix fmt` needs the tracked treefmt.toml** — without it the formatter
  resolves a read-only store copy and formats nothing while exiting 0.
- **Darwin is deliberately absent** from the flake systems (nixpkgs no
  longer evaluates x86_64-darwin; no consumer rides macOS). Revisit on a
  real consumer (ROADMAP).
- **erraudit rides the flake as a PRIVATE git+ssh input** (forces SSH
  auth from fresh clones; CI never touches the flake). It tracks master
  with the rev pinned in flake.lock and consumes erraudit's OWN hermetic
  flake package. The `nolint-audit` flake check needs a real go toolchain
  in the sandbox (tmp GOCACHE/GOPATH + GOPROXY=off = the hermeticity
  proof). erraudit is ALSO system-installed on fleet hosts — path-
  dependent probes run bare there but may SILENTLY SKIP on CI; prove
  locally before trusting. erraudit's domain: the LEGACYERRORS migration
  (sentinel `errors.Is` against known sentinels deliberately exempt) plus
  `generic_return` under `--max`; discrimination probes must plant
  `errors.As`. Flake-source probes must STAGE the plant (`git add -N`) —
  flake checks read the git tree.
- **The buildflow repair phase rewrites lockfiles; the daemon wraps the
  damage** (paid 2026-11-06 AND 2026-11-14, when it bumped SIX inputs —
  erraudit to a rev with a stale upstream vendorHash, plus a cmdguard
  v4→v5 major). Recovery: forward
  `git restore --source=<good-sha> -- flake.lock` + your OWN commit, then
  re-run `bash scripts/check-erraudit-rev.sh` before trusting the tree.
  The upstream vendorHash root cause gets fixed in erraudit itself
  (`buildflow -s nix-hash-fix --fix` there — done 2026-11-14 at its
  f5a40b3a), never by hand-pasting a hash here.
- **The erraudit-rev tripwire** (`scripts/check-erraudit-rev.sh`, also a
  flake check + pre-commit leg): flake.lock's erraudit rev must equal
  `expected_rev` in the script. `nix flake update erraudit` bumps it
  silently and erraudit's rules ARE behavior — run the erraudit checks
  on the new rev first, THEN bump expected_rev.
- **Stale buildflow binary is the #1 silent root cause**: the running
  binary predating BuildFlow HEAD masked real upstream fixes for hours
  (2026-11-14). When behavior contradicts current code, `buildflow
  doctor` (binary-freshness), then rebuild + reinstall from the BuildFlow
  repo before debugging further.
- **The buildflow nix-fmt/treefmt roundtrip wedge is FIXED upstream**
  (2026-11-14, BuildFlow repo — was "known bug, do not fix here" since
  2026-11-04). Root causes, each live-probed: (1) the sandbox was not a
  git repo, so treefmt's tree-root discovery refused explicit paths
  ("path . not inside the tree root /nix/store") or walked the ENTIRE
  filesystem (3.5M files); (2) with a git repo, nix's clean-source filter
  refuses UNTRACKED files; (3) without flake.lock in the sandbox, a
  fresh-lock eval coupled formatting detection to CURRENT upstream state.
  Fix: the sandbox is git-init'd with every copied file staged, and
  flake.lock rides along on every lane.
- **docs-gates flake checks fixed 2026-11-14 (both were unpassable since
  11-04, masked by the vendorHash cascade):** `golangci-drift` counted
  pins AFTER sort -u (two identical job pins counted 1 ≠ 2 demanded;
  distinct pins never equaled the single flake version) — now counts raw
  pins then compares the collapsed unique value; `shell-lint` ran bare
  actionlint, which requires a git project above the workflows dir (the
  store source has none) — now explicit paths. The BuildFlow
  flake-pin-consistency check ALSO had a false FAIL over the flattened
  `inputs.<name>.url` declaration form (its regexes demanded line-start
  `<name>.url` / `<name> = {`) — fixed upstream (optional `inputs.`
  prefix), do not reformat flake.nix to dodge it.
- **Buildflow's managed `.gitignore` block contains a broad `reports/`
  rule that silently swallows ANY `reports/` directory** — including new
  fixtures under `testdata/parity/*/reports/` (files exist locally, never
  staged, CI fails with golden drift only; paid 2026-11-04). Scoped
  negations were appended OUTSIDE the managed block. After adding
  testdata/corpus files: `git check-ignore -v` (with `--no-index` for
  TRACKED files — the plain form never reports them), `git status
  --ignored`, verify staged counts. `git ls-files` prints paths relative
  to the INVOKING cwd — resolve with filepath.Abs, not a repo-root join.
- **Pre-commit legs (bank-sync pattern, config in flake.nix
  `preCommitCheckFor`)**: `nolint-audit` (same staleness check at commit
  time) and `docs-battery` (the WHOLE dogfood battery when living docs
  are staged — the ungated-train lesson made mechanical). Local blocker:
  `core.hooksPath` is set to `.git/hooks`, which pre-commit refuses to
  override — unset it locally to arm auto-install; until then run hooks
  manually. `.pre-commit-config.yaml` is a generated store symlink,
  gitignored by intent — never commit it.
- **go1.26+ compile-rejects mixing literals with a slice spread in one
  variadic call** (`f("a", "b", slice...)` — even `fmt.Println`): build
  the argv with `append` first. Bites exactly the
  `exec.CommandContext(t.Context(), "git", args...)` probes tests write.
- **LSP ghost diagnostics are a CACHE artifact, not a harness bug**
  (root-caused 2026-11-04, do not file upstream): the host-managed
  golangci-lint-lsp wrapper pins a persistent analysis cache that does
  not invalidate on edits; it survives `lsp_restart`. Protocol: prove
  with the real gates → clear the wrapper caches + restart.
- **Striking report-ledger rows by ROW NUMBER is a corruption machine**
  (paid 2026-11-04, reverted): the §f ledgers share their `| N |` row
  shape with board-mirror tables, and a greedy capture collapses cells.
  The mechanics are scripted: `scripts/safe-strike.sh` (unique ≥4-letter
  text anchors, `[^|]*`-bounded rewrites, clean-commit-boundary check,
  dry-run default, `--gate` per batch) — use it instead of hand sed; only
  rows with demonstrably-landed work get struck.
- **Known BuildFlow file-size-check cache bug** (2026-11-04, filed
  upstream as LarsArtmann/BuildFlow#87): its result-cache key is CONSTANT
  (empty FileContract), so post-split findings replay for deleted files.
  Honest verdict needs `BUILDFLOW_NO_RESULT_CACHE=1 buildflow -s
  file-size-check --format finding` until fixed upstream.
- **go-auto-upgrade warnings over this repo are deliberate non-fixes**
  (2026-11-06, the full set reduces to three classes): samber/lo
  suggestions withheld (lo in no go.mod; stdlib-only root; provider is a
  thin shim — one 8-line filter doesn't justify a dep); jsonv1tov2
  declined at the go 1.26 floor (provider stays on encoding/json v1 ON
  PURPOSE: `DisallowUnknownFields` is the row-shape contract, v1
  semantics are what the provider tests pin, and the root JSON face is
  byte-pinned by parity goldens); the six cross-file-wrapper warnings
  name the cmp.Or accessors as "no call sites in this file" — that IS
  the accepted settings-projection pair, not dead code.
- **The --margins near-bar arithmetic**: the window is `[bar-2, bar)` and
  the DF counts LIVING-DOC LINES ONLY, never report lines. The
  corpus-replace parity fixture's `generic_df: 2` is load-bearing TWICE:
  it powers the rewrite-mode delta arm AND puts the one-line living doc's
  whole vocabulary inside the near-bar window. Heading tokens are
  deliberately NOT excluded: a heading word that reaches the bar DOES
  genericize items; hiding a real path would be worse than listing it.
- **The wedge class was the capture-pipe deadlock** (root-caused
  2026-11-05 by live reproduction): gate output past the ~64KB pipe
  buffer parked its write on the full capture pipe while io.ReadAll
  waited — netpoll-parked at 0% CPU forever. Fixed by concurrent
  drainers in gatecli; `gate_timeout_sec` stays the bound for any future
  class. If a battery run wedges again, suspect a NEW park first.
- **Size a foreign corpus BEFORE launching the battery** (the CV wedge,
  2026-11-05): `ls docs/status | wc -l` plus a living-doc line count
  costs one second and predicts hours-long audit-alls. Corpus
  memoization replays cost one full-corpus scan per DISTINCT token
  (hours → 49s on 1,643 reports); audit-all prints its replay shape +
  worst-case estimate BEFORE replaying, and `gate_timeout_sec` caps every
  gate. Estimate and cap are circuit breakers, not permission to launch
  blind.
- **Gate verdicts are point-in-time under foreign daemons** (paid
  2026-11-05): a battery over a repo with an active auto-commit/report
  daemon can flip between runs minutes apart. When reporting verdicts
  about a repo you do not own, record the run's timestamp; never treat
  one run as steady state.
- **dprint-remote-plugins is an accepted warning**: dprint.json fetches 4
  wasm plugins once, cached in ~/.cache/dprint. The remedy if it ever
  bites: `nix build .#dprint-plugins` in the BuildFlow repo +
  `BUILDFLOW_DPRINT_PLUGINS=<store>/share/dprint-plugins`.

## Testing

- stdlib `testing` only (the stdlib-only contract), table-driven where it
  fits; gates are exercised through their real `Run` functions with
  `gatecli.CaptureOutput`, never mocked output.
- Two parity layers: the golden fixtures (exact bytes, see constraints)
  and the defaults-are-historical invariant (config-less repos reproduce
  pre-extraction behavior).
- Corpus fixtures use a `{{DATE}}` token in filenames and contents,
  substituted with the run date, so report ages stay stable across time.

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
timeout 300 go run ./cmd/docs-gates run <repo-root>  # the battery over any repo (run it over THIS repo after doc edits)
go run ./cmd/docs-gates run <repo-root> --json    # the machine face: one JSON verdict document on stdout, humans on stderr
timeout 300 go run ./cmd/docs-gates <verb> <repo-root>  # a standalone gate (standalone runs have NO cap of their own)
PARITY_UPDATE=1 go test ./internal/battery/ -run TestParityGolden   # re-pin goldens after an INTENTIONAL output change
```

CI (`.github/workflows/ci.yml`, push/PR): `go build` / `go vet` /
`scripts/stack-drift-check.sh` / `go test ./...` plus the audit-all cliff
benchmark arm (`go test -bench BenchmarkCheckReportReplay
-benchtime=100x` under `timeout 45s`; the memoization kill switch) from
the root (`go` job), a `provider` job from `pkg/provider`, and the
`dogfood` job running the battery over THIS repo every push. No
standalone gofmt step — golangci gates format on both jobs. The flake
(and its erraudit checks) never runs on CI: local-only quality loop. The
ci.yml golangci action `version:` and the flake's nixpkgs golangci pin
are kept in one train by the `golangci-drift` flake check; `shell-lint`
shellchecks every script + actionlints the workflow (explicit paths —
the store source has no .git for actionlint's discovery); `go-floors`
asserts both go.mod floors stay within the pinned toolchain; dependabot
(actions + both go modules, grouped, weekly) keeps action SHA pins fresh.
