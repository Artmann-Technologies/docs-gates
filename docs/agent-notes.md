# Agent notes — the paid-for incident narratives

Full detail behind every rule in AGENTS.md "Non-obvious facts": what
happened, when, and the exact mechanics. Extracted 2026-11-14 when
AGENTS.md was distilled to its 220-line budget (the buildflow
agents-md-size ratchet); nothing was deleted, only relocated.

---

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

