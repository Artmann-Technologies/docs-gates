# AGENTS.md — docs-gates

The docs-health battery gates as one standalone stdlib-only Go binary,
extracted from the pbx-artmann deployment flake 2026-11-03. Public, MIT,
module `github.com/Artmann-Technologies/docs-gates`. **This repo
dogfoods itself:** its living docs (TODO_LIST, ROADMAP, CHANGELOG) are
the corpus its gates judge — run the battery (`go run ./cmd/docs-gates
run .`) after touching them and keep the rituals (strike landed rows,
archive old reports). Full incident narratives: `docs/agent-notes.md`.

## Layout and control flow

One binary, fourteen verbs; `cmd/docs-gates` dispatches (verb tables in `verbs.go`), everything else is
a library:

| Package                                                    | Role                                                                                                                                                                                                                                |
| ---------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cmd/docs-gates`                                           | verb dispatch (run, freshness, annotate, rows, routing, cadence, stub, citations, stamps, vocabulary, agents, strike, purge, archive), verb-to-legacy-name map, help surface (`--help`/`-h`/`help [verb]`, did-you-mean), `--version` (ldflags-stamped, `"dev"` bare)          |
| `internal/gatecli`                                         | exit-code constants, boolean/valued `--flag` parsing (`CutIntFlag`), the config prologue (`LoadConfig`), repo-root resolution, `Outf`/`Errf` + kin (the ONE home for unchecked `fmt.Fprint*`), `Truncate`, `CaptureOutput`, `Livef` |
| `internal/battery`                                         | the `run` verb: ordered gates from config, run INDEPENDENTLY with captured output, PASS/FAIL/SKIP matrix — no first-failure abort                                                                                                   |
| `internal/config`                                          | `.docs-gates.json` loader: decode over `Defaults()`, unknown keys WARN, malformed/invalid exit 2                                                                                                                                    |
| `internal/docs{freshness,annotate,routing,cadence,strike}` | the five historical gates + docs-purge: each a `Run([]string) int` whose argv matches the historical standalone binaries, each with a `Settings` projection (zero values → historical constants)                              |
| `internal/docs{stub,citations,stamps,vocabulary,agents}`    | the five content/rollout gates (2026-11-21 census + #11 + agents): glossary checksums, citation resolution, stamp age bars, FEATURES legend tokens, AGENTS.md mtime bars; warn-mode `*_enforce` knobs, same `Run`/`Settings` shape                                       |
| `internal/docsarchive`                                     | the archive verb (2026-11-23): the cadence ritual's banner + git mv for zero-open-§f reports, dry-run default, `--apply` writes, double-banner FATAL, open reports SKIP                                            |
| `pkg/provider`                                             | BuildFlow toolsdk provider — a SEPARATE go module that EXECS the binary from PATH, parses `run --json` into findings (strict decode), registers via `toolsdk.Register`                                                              |

Control flow: `main.go` → verb's `Run(rest)` → `gatecli.Parse` →
`Settings` → gate logic → verdict lines → exit code. `--version`/`--help`
are intercepted BEFORE any gate parser; help is pinned per gate by
usage_test.go.

## Critical constraints

- **stdlib-only is the root module contract** — zero requires
  (`scripts/stack-drift-check.sh` fails CI on drift); keeps the nix
  build hermetic. Third-party deps live in a separate go module with a
  STACK.md row. JSON, not YAML, for config: stdlib contract, jq-able.
- **Exit codes: 0 clean / 1 findings / 2 usage-environment** — everything
  funnels through gatecli constants; new failure classes map onto these.
- **Zero-value Settings never reach gate logic** — projections fall back to
  the historical constants (`cmp.Or`, `OrSlice`, `> 0`); config-less repos
  behave byte-for-byte like pre-extraction.
- **The five legacy names are API** (the newer gates carry no legacy
  wrapper): flake wrappers exec `docs-gates <verb>`; each answers
  `--version` under its legacy name. Never remove a wrapper without a
  migration plan.
- **The battery runs TEN gates** (2026-11-23): freshness, audit-all,
  annotate, routing, cadence, stub, citations, stamps, vocabulary,
  agents — `KnownGates` is the whole `skip_gates` vocabulary; the JSON
  face's gate rows are generic so the provider absorbs a new row.
- **Parity is enforced**: `TestParityGolden` pins byte output of
  fifteen invocations over `testdata/parity/corpus`. INTENTIONAL change
  re-pins via `PARITY_UPDATE=1 go test ./internal/battery/ -run
  TestParityGolden`, then hand-diff — a golden bump is a contract change.
- **Auto-commit daemon** (~2-3 min cadence, NEVER pushes — PUSH YOUR OWN
  TRAINS): format/lint at write time, heuristic `chore: auto-commit`
  entries. Never restore/checkout files you did not author; NEVER `git
  revert HEAD` blind — explicit SHA only; pin your baseline before
  before/after diffs — HEAD moves mid-session.
- **Lint: fleet golangci config at 0 findings.** `//nolint:erraudit`
  directives are LIVE, guarded by the `nolint-audit` + `erraudit-lint`
  flake checks; golangci's unknown-nolint warning is expected.
  `erraudit lint ./... --max` green since 2026-11-12: generic-error
  returns carry typed refusal carriers (unexported struct + `newXxxError`
  + `Unwrap`; the idiomatic `error` return KEPT). Never cite a directive
  COUNT from memory — run `erraudit nolint-audit .`

## Verdict arithmetic (battery)

- Exits 1 iff ≥1 gate row FAILs; SKIP rows exit-clean; every gate always
  runs (one foreign red never masks later gates). A gate exiting 2
  INSIDE the battery is a FAIL row `(rc=2)`; findings live in the gate's
  own output. `--json`: one snake_case document on stdout, humans on
  stderr, byte-pinned by the battery-json golden — a knob means re-pin
  golden AND provider. Half the `gate_timeout_sec` budget triggers the
  watchdog line (stderr-only, informational).

## Non-obvious facts (paid for)

- **`routed` counts as a rule-A marker** (#3): docsrouting treats a
  struck routed row as closed; markers_test pins both contracts.
- **Near-miss §f headings WARN on gate STDOUT** (#9): the canon-rejected
  `## §f` shape fires `WarnSectionFNearMiss` in freshness/audit-all/routing
  — stdout, because the battery swallows a PASS row's stderr.
- **The purge collector refuses struck rows whose status cell reads
  OPEN** (#10): `IsOpenCell`/`IsDoneCell` are the shared predicates;
  refusal = exit 2, nothing written.
- **docs-strike locates Task/Status/Evidence BY HEADER NAME** (#6):
  fleet TODO_LISTs are 4-6 column with Status first, last or absent;
  cell count must equal the table's header count.
- **The tokenizer splits tokens on `/`** (#5): a split spelling of a
  slash-joined token BINDS; path-like tokens match as segments (no
  golden moved — the parity corpus carries none).
- **`annotate_corpus`** mirrors `audit_corpus` ("all" default /
  "unarchived" drops the archived subdir) — validateCorpus guards both.
- **A dependabot bump to pkg/provider does NOT carry the flake's
  vendorHash** (PR #7 lesson): the merge train must also bump
  flake.nix's `docs-gates-provider-suite` vendorHash.
- **Pre-flight before ANY living-docs edit batch** (`docs-freshness
  --pre-flight`, stdin candidates, SAFE only, battery BEFORE the commit
  — paid twice; universe is exactly the trio; the refusal is a lowercased
  substring test, so rewrapped removed lines still collide).
- **Root `./...` never sees `pkg/provider`** — test/lint it from its own
  directory. Floors: both go 1.27; the flake pins ONE toolchain (go_1_27
  + GOTOOLCHAIN=local, one `goPkgAttr` binding — never fork).
- **gatecli.CaptureOutput is a process-global stdout swap** — its tests
  must never be t.Parallel. **Freshness = newest report only (strict);
  audit-all = whole corpus under the audit_max pin.** Ratchets are pins,
  not floors (tolerated findings print RATCHET).
- **docsstrike's target file comes from config** (`todo_file`); vars
  named `s` fail varnamelen when scope grows. **Report vocabulary is
  shared, never re-copied**: docsfreshness exports
  `LivingDocsLines`/`SectionFRe`/`HeadingRe`/`Tokens`/`Squash`/`GenericDF`.
  **ExtractItems treats any line-start `\d+[.)]` as an item boundary** —
  keep report continuation wraps away from a leading number+paren.
- **Known split brains** (2026-11-15): the marker grammar (writer renders
  `markerFor`, judge matches `markerRe`, bound by markers_test.go) and
  the docs-health binary (deleted; callers `go run`). See ADR 0001 + the
  split plan under docs/planning/.
- **The five `//art-dupl:accept` directives are the dedup baseline**
  (`-t 1`: 0 groups). The settings-projection pair is ACCEPTED —
  domain-specific Settings; a generic helper would couple gates.
- **Deliberate non-fixes**: `gatecli.Truncate` stays hand-rolled (stdlib-only;
  goldens pin semantics); samber/lo withheld (stdlib-only root); jsonv1tov2
  declined (`DisallowUnknownFields` is the row-shape contract, byte-pinned);
  crypto/md5 in docsstub is census-fingerprint matching, not security (the
  library-policy ERROR is a standing known-delta, documented here); the
  mutators use plain `os.WriteFile` under `--apply` (git is the atomicity
  layer; the purge two-write window is accepted — TODO_LIST §10).
- **Fleet library audit** (2026-11-23, `docs/review/libs/index.html`): NO
  LarsArtmann library is adopted, correctly — 16 NOT-APPLICABLE, 7
  NOT-ADOPTED (real overlap, barred by the root contract). New serving, watch
  or daemon faces are NEW sibling modules. **Re-audit when** cmdguard ships a
  zero-dependency core, stdlib-only is relaxed, or such a module exists.
- **`nix fmt` needs the tracked treefmt.toml.** **Darwin is deliberately absent** from flake systems (revisit on a real consumer).
- **erraudit rides the flake as a PRIVATE git+ssh master-tracking input**
  (its OWN hermetic package; rev pinned in flake.lock). Probes may
  SILENTLY SKIP where the binary is absent (CI) — prove locally;
  flake-source probes must STAGE plants (`git add -N`). The
  erraudit-rev tripwire pins every master-tracking input's rev — rules
  ARE behavior, so check on the new rev, THEN bump one input at a time.
- **Buildflow's repair phase rewrites lockfiles; the daemon wraps the
  damage** (six events 11-06..11-23). ROOT CAUSE PROVEN 11-23: the
  `nix-flake-update` step is a blindRepairer over BARE `nix flake
  update` — `buildflow --fix` fires it; plain runs and bare
  `nix flake check` are safe. BuildFlow#115 filed. Recovery: forward
  `git restore --source=<good-sha> -- flake.lock` + OWN commit, tripwire
  before trusting; hash fixes happen in erraudit via
  `buildflow -s nix-hash-fix --fix`, never here. A STALE buildflow
  binary is the #1 silent root cause besides — doctor, rebuild.
- **`buildflow format` does NOT cover golangci fmt** (golines/gci) —
  finish with `nix develop -c golangci-lint fmt ./...` and re-run the
  drill. **Pre-commit stashes unstaged edits** — stage every ritual edit
  before committing. **The daemon RACES staging** (paid 11-23 ×5): it
  commits your just-written files mid-edit, so "nothing to commit"
  usually means the daemon already landed it — check `git log`, then push.
- **Annotated tags return the TAG OBJECT sha** from `git rev-parse
  <tag>`/`ls-remote` — compare commits via `<tag>^{commit}` before
  declaring a moved tag (the 11-23 false alarm).
- **`go get` on a nested module caches negative resolution** — a first
  attempt before proxy indexing poisons retries; retest with a fresh
  GOMODCACHE (paid proving pkg/provider@v0.5.0).
- **Fixed upstream, do not re-litigate**: the nix-fmt/treefmt roundtrip
  wedge (2026-11-14); BuildFlow#87 (result cache, no bypass; dprint bundle
  rides `nix build .#dprint-plugins` + BUILDFLOW_DPRINT_PLUGINS); the
  `inputs.<name>.url` false-FAIL (do not reformat flake.nix to dodge it).
- **Pre-commit legs** (`preCommitCheckFor`): nolint-audit everywhere,
  docs-battery + erraudit-rev-tripwire on flake.lock staging; the
  docs-battery leg is devShell-ONLY (the sandbox's synthetic repo makes
  verdict-hash evidence red by construction). Auto-install needs
  `core.hooksPath` unset; `.pre-commit-config.yaml` is a generated
  store symlink — never commit.
- **go1.26+ rejects mixing literals with a slice spread in one variadic
  call**: build the argv with `append`. **LSP ghost diagnostics are a
  CACHE artifact**: prove with real gates, then clear
  `~/.cache/golangci-lint-analysis` and restart.
- **Striking ledger rows by ROW NUMBER is a corruption machine**: use
  `scripts/safe-strike.sh` (unique ≥4-letter anchors, dry-run default,
  one strike per commit boundary, `--gate`/`--battery` per batch).
- **dprint-managed tables × pre-flight: judge the FORMATTED bytes against
  the PRE-edit tree** — the formatter re-aligns cells AFTER the verdict
  (paid 2026-11-14).
- **The --margins near-bar window is `[bar-2, bar)`**, counts LIVING-DOC
  LINES ONLY; the corpus-replace fixture's `generic_df: 2` is
  load-bearing twice; headings are NOT excluded. The wedge class was
  the capture-pipe deadlock (fixed by concurrent drainers;
  `gate_timeout_sec` bounds it; a new wedge = suspect a NEW park).
  Buildflow's managed `.gitignore` block swallows ANY `reports/` dir —
  after adding testdata: `git check-ignore -v` (`--no-index` for
  TRACKED files).
- **Size a foreign corpus BEFORE launching the battery** (`ls docs/status
  | wc -l`; memoized replays: one scan per DISTINCT token, hours → 49s).
  Gate verdicts are point-in-time under foreign daemons — timestamp the
  run; one run is not steady state.

- **CHANGELOG prose is a living-doc edit — run the pre-flight ritual on
  it too** (paid 2026-11-24): the match predicate is raw-substring, so
  "drafting" feeds token `draft` and "routing" feeds `route`; two DF-7
  tokens genericized and flipped two PASSING items to UNHARVESTED
  mid-leg. Margins-first (`--margins` names the DF-6/7 tokens), then
  either dodge the substring or mark the verifiably-done items — the
  tripwire case's precedent, now thrice.

- **Testing**: stdlib `testing` only; gates are exercised through their
  real `Run` functions with `gatecli.CaptureOutput`, never mocked
  output. Two parity layers: byte goldens + the defaults-are-historical
  invariant. Corpus fixtures use a `{{DATE}}` token so report ages stay
  stable.

## Commands

```console
buildflow --build-mode dev                        # the quality loop (BuildFlow owns it)
nix build .#docs-gates -o /tmp/dg-result          # package; checkPhase runs go test ./...
nix develop -c go test ./...                      # root suite (provider excluded — see gotcha)
(cd pkg/provider && nix develop -c go test ./... && nix develop -c golangci-lint run ./...)  # provider's own suite + lint
nix develop -c golangci-lint run ./...            # fleet config, 0 findings expected
nix develop -c golangci-lint fmt ./...            # golines/gofumpt/gci (buildflow format does NOT cover this)
nix develop -c erraudit lint ./... --type-aware   # rc=0 expected (nolint directives live)
nix develop -c erraudit nolint-audit .            # directive staleness (also a flake check)
nix develop -c pre-commit run docs-battery --all-files  # commit-time dogfood leg (needs core.hooksPath unset)
go run ./cmd/docs-gates purge <repo-root>          # board purge (dry-run default; --apply writes)
bash scripts/check-erraudit-rev.sh                # rev tripwire (also a flake check + pre-commit leg)
nix fmt                                           # nixfmt-tree over *.nix (via treefmt.toml)
nix flake check --all-systems --no-build          # both-arch eval (all checks + the skills guards)
timeout 300 go run ./cmd/docs-gates run <repo-root>  # battery over any repo (run over THIS repo after doc edits)
go run ./cmd/docs-gates run <repo-root> --json    # the machine face: one JSON document on stdout, humans on stderr
PARITY_UPDATE=1 go test ./internal/battery/ -run TestParityGolden   # re-pin goldens after an INTENTIONAL output change
```

CI (`.github/workflows/ci.yml`, push/PR): root `go` job (build/vet/
stack-drift/tests + the audit-all cliff benchmark arm), a `provider`
job, and the `dogfood` job (battery over THIS repo every push). The
flake never runs on CI. `golangci-drift` keeps ci.yml's action version
and the nixpkgs golangci pin in one train; dependabot (grouped, weekly)
keeps action SHA pins fresh.
