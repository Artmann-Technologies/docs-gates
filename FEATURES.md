# FEATURES

Honest status inventory. DONE = proven by a test or a live run in this
repo; PARTIALLY DONE = works but incomplete; PLANNED = designed, not
built.

## DONE

- **docs-freshness gate** — newest-report open-§f harvest check,
  `--audit-all` whole-corpus replay, the `--audit-max=N` ratchet,
  `--explain-item` / `--explain-token` diagnostics, `--pre-flight`
  candidate replay with bind + duplicate legs. Ported from the source
  fleet with identical output semantics; config-plumbed (report dir,
  living docs, distinctive-token bar).
- **docs-annotate gate** — every struck item carries a resolvable verdict
  (hash/tag claims via git, repo-path claims), rule F inline-code spacing
  scan over configurable extensions, config-driven report dir + extra
  globs + rule-F toggle, and the `annotate_max` adoption ratchet (born
  from the 2026-11-03 pilots: repos with pre-doctrine strike corpora seed
  the measured count and drain down; strict -1 default preserves the
  pre-extraction behavior byte-for-byte).
- **docs-routing gate** — non-§f routing-gap lint (`--strict`), pointer
  convention, generic-token bar shared with freshness.
- **docs-cadence gate** — struck-row purge threshold + archive cadence
  (unarchived count + oldest age), config-driven thresholds, TODO file,
  report dir.
- **docs-strike tool** — TODO_LIST row-strike helper (dry-run default,
  ambiguity refusals, double-strike refusals, verdict-opener refusals),
  target file from config unless `--file` overrides.
- **Battery runner** (`docs-gates run`) — five gates as one independent
  matrix (PASS/FAIL/SKIP per gate, no first-failure abort), exit 0/1,
  `skip_gates` + `audit_max` from repo config. Live-proven over the
  source fleet's corpus: 5/5 PASS.
- **Config schema** — `.docs-gates.json` loader: defaults for everything,
  unknown-key warnings, malformed/invalid = exit 2.
- **Multi-verb binary** — six verbs, per-verb `--version` under the
  legacy standalone names.
- **flake.nix** — buildGoModule (vendorHash-free, stdlib-only), five
  legacy-name wrappers, both-Linux systems, devShell (go 1.26,
  golangci-lint, nixfmt-tree), nix fmt via tracked treefmt.toml.

## PARTIALLY DONE

- **Parity discipline** — the extracted battery matched the source
  battery's Go-gate verdicts live (5/5 PASS over the real corpus);
  the byte-identical golden-fixture corpus (testdata/parity) is still
  owed. (TODO_LIST)

## PLANNED

- **BuildFlow provider** — toolsdk self-registration (`pkg/provider`
  submodule) so the battery rides BuildFlow runs as a provider step.
- **Pilot rollout evidence** — crm + go-appkit seeded (2026-11-03) and
  green under adoption ratchets (audit_max, annotate_max, cadence bars at
  the measured debt); the ratchets drain as the rituals run.
- **Golden parity fixtures** — a synthetic corpus with pinned verdicts
  guarding the gate semantics against drift.

## WORTH CONSIDERING

- A `numbers`-style counts gate (test/check counts vs docs claims) — the
  source fleet's docs-numbers is repo-specific; a generic shape needs a
  design pass first.
- A `--json` verdict output for CI annotations.
