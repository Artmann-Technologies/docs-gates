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
  for living docs, a report-freshness age gate distinct from cadence, a
  false item-boundary detector (§f continuation lines that begin with
  `\d+[.)]` — the 2026-11-04 false-split class), and a `--json` verdict
  output for CI annotations (which would also free pkg/provider from
  parsing the human matrix text — the findings-as-data direction the
  fleet's tooling takes, cf. the dnsblockd io-surface brainstorm).
  Nothing here is committed; each wants a Pareto pass first.
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
