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
  Pilot findings live in io-surface's tracking (TODO_LIST §8 row 1); the
  json failure path (text on stderr, empty stdout, rc=1) and the
  init --force defect still need their scratch probes (§8 rows 2-3).
- **Pre-flight model growth** — multi-line/range replacements (block
  swaps, not just single-line `--replace`) and pure removals
  (`--remove=OLD` — deleting a line is not a rewrite) if fleet rewrites
  need them.
- **Margins/routing convergence** — `--margins --json` for tooling
  consumption (jq-able, the config-format precedent), and docsrouting
  Gap rows carrying harvest suggestions (the margins vocabulary applied
  to unrouted rows).
- **`docs-gates doctor` verb** — PATH, config, flake state in one
  command; the provider HealthCheck hint is the seed.
- **Board-mechanics fuel** — a split-check verb (verify one item's
  verdict without a full run), `CutStringFlag` (the string sibling of
  gatecli.CutIntFlag, re-needed by every new valued flag), and a
  findings-gate cache-state summary (say WHAT the result cache did to
  this run) — small mechanics that keep surfacing across sessions.
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
- **Release cadence** — v0.1.0/v0.1.1 are tagged; decide when v0.2.0
  cuts (candidate trigger: the pbx-artmann migration proving the fleet
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
