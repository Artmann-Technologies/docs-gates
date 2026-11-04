# Domain Language

Ubiquitous language for the docs-gates battery. Terms are used
consistently across code, config, output, and docs; a term here is the
contract — code is the source of truth for behavior, this file for
vocabulary.

## Glossary

| Term                | Definition                                                                                                                              | Where used                                              |
| ------------------- | --------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------- |
| gate                | One docs-health check exposed as a verb (freshness, annotate, routing, cadence, strike). Each rides `Run([]string) int`.                 | `internal/docs*` packages                               |
| verb                | A `docs-gates` CLI command: the five gate verbs plus `run`. Unknown flags are usage errors (exit 2).                                      | `cmd/docs-gates`                                        |
| battery             | `docs-gates run`: every gate executed independently, verdicts rendered as one matrix; no first-failure abort.                             | `internal/battery`                                      |
| verdict matrix      | The PASS/FAIL/SKIP table the battery prints — one row per gate, always all rows. Failed gates drive exit 1; SKIP rows never count.        | `internal/battery/battery.go`                           |
| exit-code contract  | 0 clean / 1 findings / 2 usage-environment. Every failure class maps onto these three; never a new code.                                  | `internal/gatecli` constants                            |
| report              | A timestamped status report file (`2*.md`) under the report dir. Point-in-time snapshot: annotate, never rewrite.                          | `report_dir` config                                     |
| §f section          | The "next tasks" section of a status report (heading matches `SectionFRe`). The harvest input.                                            | `internal/docsfreshness/reports.go`                     |
| living docs         | The files a repo maintains in place: TODO_LIST, ROADMAP, CHANGELOG (config: `living_docs`). Lowercased raw-line corpus for token checks.  | `internal/docsfreshness.LivingDocsLines`                |
| harvest             | Pulling a report's open §f items forward into the living docs. Freshness fails an item the living docs do not reflect.                     | `internal/docsfreshness/check.go`                       |
| distinctive token   | An item token whose document frequency across living-docs lines is below `generic_df` — matched on purpose, not by accident.               | `GenericDF`, `generic_df` config                        |
| generic token       | A token on >= `generic_df` separate living-doc lines; matching it alone is accidental overlap, not a harvest.                              | `internal/docsfreshness/check.go`                       |
| fresh:skip marker   | `<!-- fresh:skip: reason -->` exempts one §f item from the harvest check (prefix match; the reason is required hygiene).                    | `SkipMarker`                                            |
| routing gap         | An open row in a report's non-§f sections that no living doc represents. The routing gate flags them (`--strict`).                         | `internal/docsrouting`                                  |
| struck row          | A resolved TODO/report row: `~~item~~ verdict`. Struck rows purge into CHANGELOG at the cadence threshold.                                 | `internal/docsstrike`, `internal/docscadence`           |
| verdict             | The resolvable evidence a struck item must carry: a git hash/tag claim (checked via git) or a repo-path claim. Annotate fails bare strikes. | `internal/docsannotate`                                 |
| rule F              | The inline-code spacing scan (backtick spacing) over configurable extensions. Toggle + extensions from config.                             | `internal/docsannotate/spacing.go`                      |
| ratchet             | A pin, not a floor: `audit_max` / `annotate_max` tolerate up to N findings green-but-named (RATCHET verdict); a REGRESSION past the pin fails. Strict default `-1` omits the flag. | `--audit-max`, `--annotate-max`, config keys |
| audit-all           | Whole-corpus replay of the harvest check under the audit ratchet — versus freshness, which checks only the newest report, strictly.        | `--audit-all`                                           |
| cadence             | The archive/purge thresholds: `struck_rows` (struck TODO rows before purge is due), `max_unarchived`, `oldest_days`.                       | `internal/docscadence`                                  |
| purge               | Deleting struck DONE rows from TODO_LIST once they live in CHANGELOG.                                                                     | TODO_LIST ritual; cadence gate watches                  |
| archive             | Moving fully-resolved reports into `<report_dir>/archived/`. Annotate does not scan the archive by default.                               | `archived_subdir` config                                |
| skip_gates          | Config list of battery gate names to SKIP (exit-clean, never counted).                                                                    | `internal/battery`, config                              |
| legacy wrapper name | The historical standalone binary names (`docs-freshness`, …) kept as flake wrappers execing `docs-gates <verb>`; API — never remove one without a fleet migration plan. | `flake.nix`, `cmd/docs-gates`                |
| parity golden       | A pinned byte-exact gate output over `testdata/parity/corpus`; re-pin intentionally via `PARITY_UPDATE=1`.                                | `internal/battery/parity_test.go`                       |
| date token          | `{{DATE}}` in fixture filenames/contents, substituted with the run date so report ages stay stable across time.                            | `testdata/parity`                                       |
| pre-flight          | Freshness mode replaying PROPOSED living-doc lines against the newest report (bind + duplicate legs) before a doc edit lands.             | `--pre-flight`                                          |

## Bounded contexts

- **Gate output vs battery output** — a single gate names findings;
  the battery's matrix names gates, not findings (findings live in the
  gate's own output; forwarded to stderr when red).
- **Freshness vs audit-all** — freshness: newest report only, strictly,
  no ratchet. Audit-all: whole corpus under the `audit_max` pin.
- `headingRe` (docsannotate, `^#{1,6}`) vs `HeadingRe` (docsfreshness,
  `^#{1,2}`) — deliberately different regexes sharing a near-identical
  name: annotate scans whole reports for strikes; freshness finds §-top
  structure only. Do not merge.

## Deprecated terms

- "pbx-tools docs gates" — the pre-extraction home; this repo is the
  implementation now. Kept only in provenance notes (STACK.md, AGENTS.md).
