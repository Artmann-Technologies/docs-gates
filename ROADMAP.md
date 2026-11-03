# ROADMAP

Long-term direction and raw ideas. Refined ideas graduate to TODO_LIST.

- **BuildFlow provider** — self-register the battery as a BuildFlow tool
  via go-finding/toolsdk: `pkg/provider` as its own go module (the root
  module stays stdlib-only), BuildFlow blank-imports it, Detect on
  `.docs-gates.json`/`docs/status`, Run = the battery. BuildFlow's own
  AGENTS.md is authoritative for its side (one blank import + require +
  regression test).
- **Golden parity corpus** — testdata/parity: synthetic docs/status +
  living docs with pinned per-gate outputs, run in checkPhase. The
  contract that lets pbx-artmann (and any fleet consumer) relock this
  repo without re-proving the gates by hand.
- **Fleet rollout** — seed `.docs-gates.json` + first battery runs in the
  pilot repos (crm, go-appkit for the extraction train), then widen.
  Config friction found in pilots folds back into the schema.
- **Schema versioning** — if the config schema ever breaks shape, add a
  `schema_version` key BEFORE changing any key's meaning; the loader
  refuses newer majors loudly.
- **Gate family growth** — candidates worth a design pass: a generic
  counts gate (claims in FEATURES vs the tree), a link/reference checker
  for living docs, a report-freshness age gate distinct from cadence.
  Nothing here is committed; each wants a Pareto pass first.
- **Windows/macOS consumers?** — the flake targets Linux only today
  (nixos-unstable no longer evaluates x86_64-darwin). Revisit only when a
  real consumer appears; the Go code is portable already.
