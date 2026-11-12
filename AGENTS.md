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

One binary, seven verbs; `cmd/docs-gates` dispatches, everything else is a
library:

| Package                                                    | Role                                                                                                                                                                                                                                                                                                                                                 |
| ---------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cmd/docs-gates`                                           | verb dispatch (`run`, `freshness`, `annotate`, `routing`, `cadence`, `strike`, `purge`), the verb-to-legacy-name map, the help surface (`--help`/`-h`/`help [verb]`, did-you-mean verb suggestions), `--version`. Version is ldflags-stamped by the flake build; `"dev"` bare-built.                                                                 |
| `internal/gatecli`                                         | shared CLI scaffolding: exit-code constants, boolean `--flag` parsing, valued `--flag=N` splitting (`CutIntFlag`), the config prologue (`LoadConfig`: load + `config:` warnings, one shape for every verb), repo-root resolution, `Outf`/`Outln`/`Errf`/`Errln` (the ONE home for unchecked `fmt.Fprint*`), UTF-8-safe `Truncate`, `CaptureOutput`   |
| `internal/battery`                                         | the `run` verb: ordered gate list from config, gates run INDEPENDENTLY with captured output, PASS/FAIL/SKIP verdict matrix — no first-failure abort                                                                                                                                                                                                  |
| `internal/config`                                          | `.docs-gates.json` loader: decode over `Defaults()` (the historical fleet constants), unknown keys WARN, malformed/invalid exit 2                                                                                                                                                                                                                    |
| `internal/docs{freshness,annotate,routing,cadence,strike}` | the five gates: each a `Run([]string) int` whose argv shape matches the historical standalone binaries, each with a per-package `Settings` projection (zero values → historical constants)                                                                                                                                                           |
| `pkg/provider`                                             | BuildFlow toolsdk provider — a SEPARATE go module that EXECS the `docs-gates` binary from PATH (internal/ is unimportable across module boundaries), parses the `run --json` verdict document into findings (strict decoder: unknown fields fail, shape drift becomes findings), registers via `toolsdk.Register` at init; consumers blank-import it |

Gate-verb control flow: `main.go` → verb's `Run(rest)` → `gatecli.Parse`
(flags + root) → `Settings` projection → gate logic → verdict lines via
gatecli → one of the three exit codes.

`--version` and `--help`/`-h` are intercepted in `cmd/docs-gates`
BEFORE any gate parser sees argv: the parsers reject unknown flags
and both would die as one. `help [verb]`/`<verb> --help` print the
gate packages' exported `Usage` texts (usage_test.go per gate pins
Usage against the parser's flag vocabulary, both directions — a help
that lies is worse than none). Usage-class FATALs ride
`gatecli.FatalUsage(verb, err)` (FATAL line + `usage: docs-gates <verb>
--help` pointer); config/environment failures keep the bare
`FATAL: %v` shape. The parity goldens pin none of the help/FATAL-refusal
paths — the re-pin ritual does not apply to them.

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
  docscadence.SettingsFromConfig, config.Defaults). String fallbacks spell
  `cmp.Or(override, config.Defaults().X)` and list fallbacks
  `config.OrSlice(...)` — the one-line convention (2026-11-04); int
  fallbacks with `> 0` semantics stay explicit if-forms (cmp.Or would let
  negatives through). A repo WITHOUT a config file gets byte-for-byte the
  pre-extraction behavior — that is the parity guarantee. Tests rely on
  it; do not change a default without re-proving parity.
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
- **This repo has an auto-commit daemon** (commits per the host config
  `/etc/projects-management-automation/service.yaml` — `min_commit_interval: 120s`
  - `debounce_duration: 60s` (read 2026-11-06), so ~2-3 min cadence, NEVER
    pushes — the host config sets `autoPush = false` (SystemNix
    configuration.nix, since 2026-06-01; the afternoon session's
    "22-commit push lag" root cause: the daemon was doing exactly what
    its config says while sessions believed it pushes and waited —
    every push in this repo's history was a manual session push, so
    PUSH YOUR OWN TRAINS; the phantom "pushes too" belief cost one
    session a 22-commit surprise): format and lint AT WRITE TIME, commit deliberate trains with the
    HEREDOC attribution form, and expect heuristic `chore: auto-commit`
    entries in the log. Never restore/checkout files you did not author —
    and verify `git status` before claiming the daemon committed your work
    (it has a known blind spot). NEVER run `git revert HEAD` blind here:
    between your push and the revert the daemon may have wrapped your own
    work in a `chore: auto-commit` commit, so `HEAD` is not necessarily the
    commit you mean (paid for 2026-11-04: a blind revert under the daemon
    removed a T5 test commit instead of the CI plant it targeted, leaving
    the break on origin ~9 h). Always `git log -1 --stat` first, revert by
    explicit SHA, and commit your own work immediately — an uncommitted
    train is what the daemon wraps into the wrong SHA. Pin your verification
    baseline (`git
  rev-parse HEAD`) BEFORE any before/after diff — HEAD moves under you
    mid-session, and an unpinned "pre-split" comparison can silently target
    a daemon commit (paid for 2026-11-04: one misdirected losslessness
    check).
- **Lint: the fleet golangci config runs at 0 findings.** Keep it there.
  erraudit rides the devShell since 2026-11-04 (flake input — see
  Non-obvious facts): the `//nolint:erraudit` directives are LIVE, and
  two flake checks keep them honest — `nolint-audit` fails the moment a
  directive goes stale, and `erraudit-lint` runs erraudit's own
  type-aware legacyerrors audit (`errors.As`/`Is` → `AsType`) over the
  root module, so a NEW migratable site fails `nix flake check` too
  (both discrimination-proven: planted stale directive / planted
  `errors.As` turn the check red, naming the site). Never
  add a directive without a code-adjacent reason; the check and erraudit's
  own audit keep them honest. golangci's runner still WARNS
  `[runner/nolint_filter] Found unknown linters in //nolint directives:
  erraudit` — expected cross-tool artifact (erraudit is a standalone
  binary, not a golangci linter; the fleet runs it standalone, no plugin),
  warning-only with 0 issues and rc=0. **The strict `erraudit lint
  ./... --max` audit (`generic_return` enforced) is green over both
  modules since 2026-11-12**: the three flagged generic-error returns
  (config.Load, docspurge.insertBullets, docsstrike.validateStrikeInputs)
  carry typed refusal carriers — unexported struct + `newXxxError`
  constructor + `Unwrap` where a cause exists (the config
  loadError/invalidError pattern), refusal messages byte-identical, and
  the idiomatic `error` return KEPT (a concrete `*T` return is the
  typed-nil landmine; constructor-shaped names like ParseArgs ride
  erraudit's name exemptions). Keep the pattern for any new refusal.
  Live directive count: 16
  non-test directives, 0 stale (nolint-audit counts NON-TEST files;
  2026-11-12, at erraudit rev 5c841cf2) — never cite a count from
  memory, run `erraudit nolint-audit .` when a doc or report needs the
  number.

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
- The `--json` machine face (`docs-gates run [root] --json`) puts EXACTLY
  one JSON document on stdout (snake_case tags, `duration_ms` integer
  columns, per-gate `stdout`/`stderr` embedded verbatim); every human
  line rides stderr; the exit contract is unchanged. The battery-json
  parity golden pins the document byte-for-byte, and pkg/provider parses
  THIS face (its decoder refuses unknown fields — row-shape drift
  surfaces as `battery-exit-without-report` findings, never silent zero
  values). Add a knob to the document = intentional contract change:
  re-pin the golden AND the provider tests together.
- A gate burning past HALF its `gate_timeout_sec` budget triggers the
  live watchdog line (`gate X over half its <cap> budget — abandoned at
  the cap`): human face prints it on the REAL stdout mid-capture via
  `gatecli.Livef` (the capture swap would otherwise park it in the
  gate's own buffer); the JSON face keeps stdout pure and warns on
  stderr. Fire-only, informational — the matrix arithmetic never reads
  it.

## Non-obvious facts (paid for)

- **Run `docs-freshness --pre-flight` BEFORE ANY living-docs edit
  batch** (TODO_LIST/ROADMAP/CHANGELOG — paid TWICE, 2026-11-05/06): the
  06-39 session genericized two TODO evidence anchors skipping the
  pre-flight, and the `4cf0e50` docs train went out ungated — only the
  afterwards battery run caught both. stdin the candidate lines, rewrite
  only on SAFE, and run the dogfood battery BEFORE the commit, never
  after. The pre-flight's universe is EXACTLY that trio (its
  loadLivingDocs corpus and its --replace targets both) — an
  AGENTS.md/FEATURES.md edit has no pre-flight surface at all (a
  --replace of an AGENTS line fails "--replace target(s) not found";
  paid 2026-11-12) and gates via the battery + dprint only; and the
  on-disk refusal is a lowercased SUBSTRING test over the joined
  corpus blob, so rewrapped reuse of removed lines still collides
  until the rewrite lands through --replace targets or genuinely new
  wording.
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
- **Report-vocabulary sharing:** docsfreshness exports the shared
  status-report definitions docsrouting consumes (`LivingDocsLines`,
  `SectionFRe`, `HeadingRe`, plus the older `Tokens`/`Squash`/`GenericDF`).
  Never re-copy these in a second gate — one definition is what keeps the
  freshness harvest and the routing triage from disagreeing about where
  §f begins or what the living-docs corpus is.
- **freshness ExtractItems treats any line-start `\d+[.)]` as an item
  boundary** (proseRe): a wrapped report line that begins `4211653)`
  splits one §f item into two false items (burned 2026-11-04 — an
  unharvested "tail" finding whose head carried the `fresh:skip` marker,
  which exempts only the item text containing it). Keep report
  continuation wraps away from a leading number+paren.
- **The five `//art-dupl:accept` directives** (docsannotate items.go ×2,
  spacing.go ×1, settings.go ×2) are the dedup baseline, not decoration:
  at `-t 1` the
  art-dupl report must show 0 groups (verified 2026-11-04, re-proven
  after the settings-projection accepts). A
  clone gets extracted or accepted-in-source with a rationale; multi-word
  text after the marker is a bare accept (a single hex token would be a
  hash pin). The 2026-11-04 pass extracted `gatecli.CutIntFlag`,
  `gatecli.LoadConfig`, `docsfreshness.LivingDocsLines`, exported
  `SectionFRe`/`HeadingRe`, and `cmp.Or`/`config.OrSlice` fallbacks; the
  settings-projection pair (loadSettings + cmp.Or accessors duplicated
  across gate packages) is ACCEPTED, not extracted — each gate's
  Settings is domain-specific and a generic plumbing helper would couple
  gates for six lines.
- **`gatecli.Truncate` stays a hand-rolled 10-liner** (deliberate
  non-fix, triaged 2026-11-04): the canonical candidate
  (go-output/nom.TruncateVisible over charmbracelet/x/ansi) would break
  the stdlib-only root contract, and its semantics differ anyway —
  visible-width budget with an ellipsis vs Truncate's byte-budget
  UTF-8-safe cut that the parity goldens pin byte-for-byte.
- **`nix fmt` needs the tracked treefmt.toml** (fleet lesson: without it
  the formatter resolves a read-only store copy and formats nothing while
  exiting 0).
- **Darwin is deliberately absent** from the flake systems: current
  nixos-unstable no longer evaluates x86_64-darwin; no consumer rides
  macOS. Revisit only on a real consumer (ROADMAP).
- **erraudit rides the flake as a PRIVATE git+ssh input** (2026-11-04):
  the repo is private (gh-verified), so the input forces SSH auth for
  `nix` commands from fresh clones; CI never touches the flake, so only
  local dev pays. The input tracks master with the rev pinned in
  flake.lock (BuildFlow's 2026-10-05 track-latest input policy) and
  consumes erraudit's OWN hermetic flake package — its private transitive
  deps never touch this repo's stdlib-only module. The `nolint-audit`
  check needs a real go toolchain in the sandbox (package loading), hence
  the pinned go + tmp GOCACHE/GOPATH + GOPROXY=off (the hermeticity
  proof). erraudit is ALSO system-installed on fleet hosts
  (`/run/current-system/sw/bin/erraudit`), so path-dependent erraudit
  probes and the nolint-audit discrimination test run bare there but skip
  on CI runners (no binary) — a test that "passes" on CI may be a silent
  skip; prove it locally before trusting it. erraudit's finding domain is
  the LEGACYERRORS migration (errors.As → AsType; sentinel errors.Is
  against known sentinels is deliberately exempt — an unchecked Close is
  errcheck's job, not erraudit's), plus the `generic_return` check under
  `--max` (typed refusal carriers per the Critical-constraints erraudit
  bullet); discrimination probes must plant
  `errors.As(err, &target)`. Flake-source probes must STAGE the plant
  (`git add -N`): flake checks read the git tree, untracked files are
  invisible to them.
- **Buildflow's managed `.gitignore` block contains a broad `reports/`
  rule that silently swallows ANY `reports/` directory** — including new
  corpus fixtures under `testdata/parity/*/reports/`. The files exist in
  the working tree (so tests pass locally) but never get staged by `git
  add <dir>`, and CI fails with golden drift only (paid for 2026-11-04).
  Scoped negations (`!testdata/parity/*/reports/` +
  `!testdata/parity/*/reports/**`) were appended OUTSIDE the managed
  block — buildflow owns the block itself. After adding any testdata or
  corpus files, run `git check-ignore -v` / `git status --ignored` and
  verify staged counts BEFORE trusting a green local suite.
- **Pre-commit nolint-audit leg (2026-11-04, bank-sync pattern):**
  git-hooks.nix drives the SAME `erraudit nolint-audit .` staleness check
  at commit time (hook config in flake.nix `preCommitCheckFor`, shared by
  the sandboxed `pre-commit-check` flake check and the devShell hook
  install). Red-controlled live (planted stale directive → hook FAIL
  naming the count). The SAME config carries the `docs-battery` leg
  (2026-11-06): a commit staging any living doc runs the WHOLE dogfood
  battery before the commit lands — the 4cf0e50 ungated-train lesson is
  now mechanical. One local blocker: `core.hooksPath` is explicitly
  set to its own default (`.git/hooks`), and pre-commit refuses to
  install over it — unset it locally to arm auto-install; until then run
  the hooks manually (command above). `.pre-commit-config.yaml` is a
  generated store symlink, gitignored by intent — never commit it.
- **go1.26+ compile-rejects mixing literals with a slice spread in one
  variadic call** (`f("a", "b", slice...)` is `too many arguments`, proven
  on both pinned toolchains 1.26.8 and 1.27.3 — it even hits
  `fmt.Println`): build the argv with `append` first. Bites exactly the
  kind of `exec.CommandContext(t.Context(), "git", args...)` probe a
  test naturally writes.
- **`git check-ignore` never reports TRACKED files as ignored** — an
  ignore-arm over committed fixtures needs `--no-index` (the gitignore
  arm's red control proved the plain form blind), and `git ls-files`
  prints paths relative to the INVOKING cwd, not the repo root — resolve
  each line with `filepath.Abs` or a repoRoot join silently leaves the
  tree.
- **LSP ghost diagnostics are a CACHE artifact, not a harness bug
  (root-caused 2026-11-04, do not file upstream):** the
  `golangci_lint_ls` pane served `undefined: <old-name>` for a symbol
  renamed in the SAME session — 0 occurrences repo-wide, real
  `golangci-lint run` + `go build` clean, and the error SURVIVED
  `lsp_restart`. Root cause: the host-managed wrapper
  (`~/.local/bin/golangci-lint-lsp-wrapper`) pins
  `GOLANGCI_LINT_ANALYSIS_CACHE` to a persistent dir whose entries do
  not invalidate on the edit; clearing the cache dirs + one restart
  cured it (diagnostics went clean). The templ PreToolUse caveat hook
  (crush-config) encodes the same family for templ. Protocol: suspect
  staleness → prove with the real gates → clear the wrapper caches +
  restart; an upstream issue was draft-ready and REFUTED by this
  experiment (verify-before-filing Gate 4).
- **Striking report-ledger rows by ROW NUMBER is a corruption machine
  (paid for 2026-11-04, reverted in ef14e11):** the §f "up to 50 things"
  ledgers share their `| N |` row shape with board-mirror tables in the
  SAME report, so a row-number sed hits the wrong table, and a greedy
  `(.*) (\|)$` capture collapses every middle cell into the struck text.
  If a strike sweep is ever redone: anchor each strike on a UNIQUE TASK
  TEXT PREFIX, bound the capture with `[^|]*` (cells never contain
  pipes), run the annotate gate after EVERY batch, and expect the daemon
  to commit a corrupt intermediate state within ~3 minutes (recovery is
  a forward `git restore --source=<good-sha>`, never a history rewrite).
  The ledgers hold 28-49 open rows each — a full sweep is a false-strike
  factory; only rows with demonstrably-landed work get struck. These
  mechanics are scripted: `scripts/safe-strike.sh` enforces them
  (one-line unique text anchors that must carry ≥4 letters, `[^|]*`-
  bounded cell rewrites, clean-commit-boundary check, dry-run default,
  `--gate` runs the annotate gate per batch) — use it instead of hand
  sed.
- **Known BuildFlow tool bug (2026-11-04, do not "fix" here):** the
  BuildFlow `nix-fmt`/`treefmt` roundtrip steps fail over THIS repo with
  `failed to create walker: path … not inside the tree root /nix/store`.
  In-repo `nix fmt` and `nix flake check` are green — the defect is the
  roundtrip copy's environment (its fresh `flake.lock` and tree-root
  resolution), not this flake. Diagnosis lives in the BuildFlow repo
  (`tools/providers/format_roundtrip.go` +
  `nixFmtRoundTripConfigFiles`); the failing steps do not gate anything
  this repo's own commands miss.
- **Known BuildFlow tool bug (2026-11-04, do not "fix" here):** the
  file-size-check result cache key is CONSTANT — the provider declares
  an empty FileContract, and the key's content hash covers only
  `tool.Files.Inputs`, i.e. no files. After splitting oversized files
  (this session), `buildflow -s file-size-check` replayed the pre-split
  findings for files that no longer existed. The honest verdict needs
  the bypass: `BUILDFLOW_NO_RESULT_CACHE=1 buildflow -s file-size-check
  --format finding`. Filed upstream as LarsArtmann/BuildFlow#87 (draft
  at BuildFlow
  `docs/drafts/file-size-check-empty-filecontract-constant-cache-key.md`);
  once fixed upstream, drop the bypass from any verification ritual.
- **Buildflow repair steps rewrite lockfiles mid-run; the daemon wraps
  the damage (paid for 2026-11-06):** a failing buildflow run's repair
  phase ran `nix-flake-update`, bumping the erraudit input to a master
  rev whose own go-modules vendorHash was stale (erraudit's auto-daemon
  had committed a go.mod/go.sum sweep without re-pinning its flake
  hash) — the devShell and every erraudit-dependent check went red, and
  the daemon wrapped the bumped lockfile into a `chore: auto-commit`
  within minutes. The `scripts/check-erraudit-rev.sh` tripwire caught
  the drift exactly as designed. Recovery is a forward
  `git restore --source=<good-sha> -- flake.lock` plus your OWN commit;
  and after ANY buildflow run that rewrote lockfiles, re-run the
  tripwire before trusting the tree. The upstream root cause gets fixed
  in erraudit itself (`buildflow -s nix-hash-fix --fix` there), never by
  hand-pasting a hash into this repo.
- **go-auto-upgrade warnings over this repo are deliberate non-fixes
  (2026-11-06, the full 26-finding set reduces to three classes):** the
  samber/lo suggestions stay withheld — lo is in no
  go.mod here, the root module is stdlib-only by contract, and
  pkg/provider is a thin toolsdk shim where one 8-line filter does not
  justify a new dependency (the gatecli.Truncate judgment). The
  jsonv1tov2 notes are the tool declining to migrate at the root's go
  1.26 floor, and pkg/provider stays on encoding/json v1 ON PURPOSE:
  `DisallowUnknownFields` strict decode is the row-shape contract, v1
  semantics are what the provider tests pin, and a v2 swap is exactly
  the runtime-invisible break the finding describes (the root's JSON
  face is additionally byte-pinned by the parity goldens). The six
  cross-file-wrapper warnings name the cmp.Or accessors
  (dir/archivedSub/reportGlob/reportDir) as "no call sites in this
  file" — that is the ACCEPTED settings-projection pair (2026-11-04:
  each gate's Settings is domain-specific; a generic plumbing helper
  would couple gates for six lines), not dead code.
- **The --margins near-bar arithmetic (documented 2026-11-05, §2c
  row):** the window is `[bar-2, bar)` — a token at DF bar-2 is
  "two more lines genericize it", bar-1 "one more" — and the DF counts
  LIVING-DOC LINES ONLY (TODO_LIST/ROADMAP/CHANGELOG), never report
  lines, so report edits never push budget. The corpus-replace parity
  fixture's `generic_df: 2` is load-bearing TWICE: it powers the
  rewrite-mode delta arm AND puts the one-line living doc's whole
  vocabulary inside the near-bar window for the margins-replace golden
  (the claimed-token exclusion pin). Heading tokens are deliberately
  NOT excluded: a heading word that reaches the bar DOES genericize
  items, and the report is informational — hiding a real path would be
  worse than listing it.
- **The wedge class was the capture-pipe deadlock (root-caused
  2026-11-05 by live reproduction, E6+E12):** gate output past the
  ~64KB pipe buffer parked its write on the full capture pipe while
  io.ReadAll waited for the gate to return first — netpoll-parked at
  0% CPU forever (the autopsied fleet wedge: 13/14 threads futex-idle,
  one netpoll thread in epoll, ~2 s CPU, buildflow parent). Fixed by
  concurrent drainers in gatecli (both capture functions); the per-gate
  timeout (gate_timeout_sec, default 300) stays as the bound for any
  future class. If a battery run wedges again, suspect a NEW park
  first, not this one.
- **Size a foreign corpus BEFORE launching the battery** (paid for
  2026-11-05, the CV wedge): `ls docs/status | wc -l` plus a living-doc
  line count costs one second and would have predicted the hours-long
  audit-all. The 2026-11-05 corpus memoization made replay cost one
  full-corpus scan per DISTINCT token (CV 1,643 reports × ~2,000 lines:
  hours → 49s; synthetic CV-shape benchmark 2,753ms → 47ms per
  1,600-report replay, 59×), audit-all now prints its replay shape +
  worst-case estimate BEFORE replaying, and the battery caps every gate
  at `gate_timeout_sec` (default 300s, FAIL row rc=timeout) — estimate
  and cap are circuit breakers, not permission to launch blind.
- **Gate verdicts are point-in-time under foreign daemons** (paid for
  2026-11-05): a battery run over a repo whose auto-commit/report
  daemon is active can flip between two invocations minutes apart — CV's
  freshness went FAIL→PASS mid-session when its daemon wrote a newer
  report. When reporting verdicts about a repo you do not own, record
  the run's timestamp alongside the matrix, and never treat one run as
  that repo's steady state.

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
(cd pkg/provider && nix develop -c golangci-lint run ./...) # provider lint — root ./... never sees the nested module
nix develop -c golangci-lint run ./...            # the fleet config, 0 findings expected
nix develop -c golangci-lint fmt ./...            # golines/gofumpt/gci imports
nix develop -c erraudit lint ./... --type-aware   # error-handling audit, rc=0 expected (//nolint:erraudit directives live)
nix develop -c erraudit nolint-audit .            # directive staleness over the root module (also the nolint-audit flake check)
nix develop -c pre-commit run nolint-audit --all-files   # the commit-time leg of the same gate (bank-sync pattern); auto-install needs `git config --unset core.hooksPath` (set to the default .git/hooks here, which pre-commit refuses to override)
nix develop -c pre-commit run docs-battery --all-files  # the commit-time dogfood leg: the whole battery over the repo when living docs are staged
go run ./cmd/docs-gates purge <repo-root>          # the board purge as one gated command: strike-collect → draft → pre-flight → apply (dry-run default; --apply writes)
bash scripts/check-erraudit-rev.sh                # erraudit-rev tripwire (also a flake check): run after bumping the erraudit input
nix fmt                                           # nixfmt-tree over *.nix (via treefmt.toml)
nix flake check --all-systems --no-build          # both-arch eval (checks: nolint-audit, erraudit-lint, provider-suite, golangci-drift, go-floors, shell-lint + the skills guards)
timeout 300 go run ./cmd/docs-gates run <repo-root>  # the battery over any repo (run it over THIS repo after doc edits; timeout bounds a standalone run the way gate_timeout_sec bounds the battery's gates)
go run ./cmd/docs-gates run <repo-root> --json    # the machine face: one JSON verdict document on stdout, humans on stderr
timeout 300 go run ./cmd/docs-gates <verb> <repo-root>  # a standalone gate (same timeout discipline — standalone runs have NO cap of their own)
PARITY_UPDATE=1 go test ./internal/battery/ -run TestParityGolden   # re-pin goldens after an INTENTIONAL output change
```

CI (`.github/workflows/ci.yml`, push/PR): `go build` / `go vet` /
`scripts/stack-drift-check.sh` / `go test ./...` plus the audit-all
cliff benchmark arm (`go test -bench BenchmarkCheckReportReplay
-benchtime=100x` under an OS `timeout 45s`; the ~10x-regression kill
switch for the corpus memoization; no standalone gofmt step — golangci
gates format on both jobs via gofumpt) from the root (`go` job), a
`provider` job running build/vet/test from `pkg/provider`, and the
`dogfood` job running the battery over THIS repo every push — both
modules gated, provider regressions fail CI, and a red dogfood row is
the docs ritual coming due, never a pin. The flake (and its erraudit
checks) never runs on CI: local-only quality loop. The ci.yml golangci
action `version:` and the flake's nixpkgs golangci pin are kept in one
train by the `golangci-drift` flake check; `shell-lint` shellchecks
every script + actionlints the workflow YAML; `go-floors` asserts both
go.mod floors stay within the pinned toolchain; dependabot (actions +
both go modules, grouped, weekly) keeps the action SHA pins fresh.
