# ROADMAP

Long-term direction and raw ideas. Refined ideas graduate to TODO_LIST.

- **Fleet rollout** — pilots seeded (crm + go-appkit carry adoption
  ratchets, 2026-11-03); SystemNix carries the PATH integration. Next:
  the pbx-artmann migration (the extraction's source fleet), then widen
  to the remaining fleet repos. Config friction found along the way
  folds back into the schema.
- **Schema versioning** — if the config schema ever breaks shape, add a
  `schema_version` key BEFORE changing any key's meaning; the loader
  refuses newer majors loudly.
- **JSON-face contract growth** — before the `run --json` document ever
  changes shape: a `schema_version` field (the config doctrine applied
  to the machine face); then candidate additions, each a contract change
  re-pinning the golden + provider tests together: a `watchdog` field on
  the gate row (the half-budget warning is invisible to stdout-only
  consumers), per-gate captured findings in the provider's finding
  messages, the matrix's `rc=N` rendered into the row message, and a
  `skip_gates` parity arm for the battery-json golden; and `--json`
  faces for the purge/strike verbs (only `run` has one today).
- **Perf-net shape series** — the audit-all benchmark pins ONE corpus
  shape (CV-shaped 1,600×2,000); a size series (100/400/1600 reports)
  would make the cliff SHAPE visible instead of one point.
- **Gate family growth** — candidates worth a design pass: a generic
  counts gate (claims in FEATURES vs the tree), a link/reference checker
  for living docs, a report-freshness age gate distinct from cadence, and
  a false item-boundary detector (§f continuation lines that begin with
  `\d+[.)]` — the 2026-11-04 false-split class). The `--json` verdict
  output LANDED 2026-11-06 (`run --json` machine face + the provider
  migrated off matrix-text parsing). Nothing else here is committed;
  each wants a Pareto pass first.
- **io-surface pilot data point (2026-11-06, input to owner question
  #3 in TODO_LIST §4 — --json face vs a new gate family):** the pilot
  over this repo's 18 extraction sites was USEFUL — an accurate BOM,
  teeth proven both directions (a planted dead claim failed, a live
  claim passed), and the claims workflow works — but an UNGATED manifest
  rots (the InboxClean b4 lesson), so adoption is gated on io-surface
  persisting `--dirs` in the manifest; until then CI pinning
  `--dirs cmd,internal,pkg` in every invocation is the only safe shape.
  Pilot findings live in io-surface's tracking (TODO_LIST §8 row 1);
  the scratch verdicts: d1 NOT reproduced at HEAD (claims survived
  init --force), b7 stands corrected (the failure face at HEAD is human
  text, not JSON), and the extractor blind spot stands (os.Stdout/os.Stderr
  writes are unwatched seams). TODO_LIST 8 row 1 carries the full text.
- **Pre-flight model growth** — multi-line/range replacements (block
  swaps, not just single-line `--replace`) and pure removals
  (`--remove=OLD` — deleting a line is not a rewrite) if fleet rewrites
  need them; per-batch stamp recording instead of last-write-wins
  (split sessions overwrite each other's stamps), the stamp recording
  the living-docs git SHA for stronger post-hoc-edit detection, and
  land-parity as an opt-in battery gate row (`skip_gates` vocabulary).
- **Margins/routing convergence** — `--margins --json` for tooling
  consumption (jq-able, the config-format precedent), and docsrouting
  Gap rows carrying harvest suggestions (the margins vocabulary applied
  to unrouted rows), and `--pre-flight` composed with `--margins` so
  candidates get near-bar warnings in the same run (two diagnostics,
  one pass).
- **`docs-gates doctor` verb** — PATH, config, flake state in one
  command; the provider HealthCheck hint is the seed.
- **Board-mechanics fuel** — a split-check verb (verify one item's
  verdict without a full run), `CutStringFlag` (the string sibling of
  gatecli.CutIntFlag, re-needed by every new valued flag), a
  `docs-gates version` verb alongside `--version` (scriptability), and a
  findings-gate cache-state summary (say WHAT the result cache did to
  this run) — small mechanics that keep surfacing across sessions.
- **Status-report ledger shape** — the 11-12/11-14 ledgers use a `## f)`
  heading the §f extractor does not match, so brainstorms stay invisible
  to freshness (green over-claims). Pick ONE definition: normalize the
  report template to the extractable shape (and harvest immediately), or
  formally accept ledger-invisibility as the brainstorm convention.
- **Fleet-wide nolint staleness tripwires** — erraudit got its audit;
  errcheck/gosec directives have none (a BuildFlow-provider-shaped
  fleet idea).
- **Gate decoupling** — docsrouting imports docsfreshness for the shared
  token corpus (Tokens/Squash/GenericDF). Fine while the gates version
  together; if they ever version independently, extract
  `internal/tokencorpus` first.
- **Provider-contract versioning** — decide how BuildFlow consumers pin
  the provider↔binary handshake (toolsdk Spec version vs a `--version`
  handshake check in HealthCheck) and document it for adopters.
- **Release cadence** — v0.1.x/v0.2.0/v0.3.0 are tagged; the next cut
  (v0.3.1 vs riding to a minor) rides the owner ruling in TODO_LIST 4b;
  candidate trigger unchanged (a fleet migration proving the fleet
  contract, or the first schema addition).
- **Windows/macOS consumers?** — the flake targets Linux only today
  (nixos-unstable no longer evaluates x86_64-darwin). Revisit only when a
  real consumer appears; the Go code is portable already.

## Non-goals

- **erraudit as a golangci-lint module plugin** — standalone binary is
  the fleet posture; the repo documents the expected cross-tool
  `nolint_filter` warning instead of fighting it.
- **Public-repo erraudit tooling in README/CI** — the flake input is a
  private git+ssh dependency by design; CI never touches the flake and
  the public README stays silent about it.
