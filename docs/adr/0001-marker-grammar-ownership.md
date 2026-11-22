# ADR 0001 — Docs-health marker grammar ownership

**Status:** Accepted, 2026-11-15
**Deciders:** owner (Lars Artmann) + the docs-health tooling audit
**Context report:** `docs/status/2026-11-15_10-13_annotate-asset-vs-docs-gates-merge-analysis.md`
(in the crush-config repo)
**Plan:** `docs/planning/2026-11-15_10-45_SUPERB-docs-health-tooling-split-and-grammar-ownership.md`
(in the crush-config repo)

## Context

The docs-health ritual has two programs:

- **Writer:** the docs-health skill's single-file `assets/annotate.go`
  (`markerFor` renders verdict phrases; `rows`/`prose`/`repair` mutate files).
- **Judge:** this repository's `docs-gates` binary (`markerRe` recognises
  verdict phrases; the five gates judge a corpus).

They share a grammar that existed only as a coincidence: seven rendered
phrases happened to contain seven keywords the judge matched. Nothing
executed that agreement, and the writer shipped a 12-day-old prebuilt binary
that no longer matched its own source.

## Decision

1. **The writer owns rendering.** `markerFor` is the only place a verdict
   phrase is constructed. The judge never sees the kind letter.
2. **The judge owns recognition.** `internal/docsannotate/markerRe` defines
   which rendered phrases satisfy rule A (a struck item carries a verdict).
3. **The contract is executable.** `internal/docsannotate/markers_test.go`
   holds one fixture line per kind, a round-trip test that runs the gate over
   them, and a sabotage case that fails when a phrase is renamed. A marker
   change without the other side is a test failure, not a review catch.
4. **Routed verdicts are not strike verdicts.** The writer's `r` kind renders
   `**→ verdict**` and never strikes; rule A therefore cannot see it. The
   read-only row gate classifies the arrow form as ROUTED (resolved). This is
   pinned as intended behaviour, not treated as a hole.
5. **The split is by read/write, not by repository.** The read-only row gates
   moved here (`docs-gates rows`, `docs-gates rows --bare`). The destructive
   mutators stay in the skill until an input model, shim, and rollback path
   are proven and every caller is converted.

## Non-goals

- Merging the skill and this repository into one codebase.
- Porting `rows`, `prose`, `status-items`, `repair`, `unwrap`,
  `repair-mangle`, or `repair-partial`. A public binary may judge a corpus;
  it may not rewrite one.
- Replacing the writer's `self-test`; it pins mutation behaviour this
  repository deliberately does not own.

## Consequences

- A new marker kind requires three edits in one change: `markerFor`, the
  fixture, and (when the phrase needs a keyword) `markerRe`.
- `docs-gates rows` is the supported read-only surface; `annotate.go
  check-rows` / `bare-rows` remain aliases for callers pinned to the writer.
- The writer is invoked as `go run annotate.go`; a committed binary beside
  the source is refused by the skill's `check-skills.sh`.

## Revisit trigger

Revisit when either (a) a second consumer needs to render markers (promote
the grammar to a shared module), or (b) the mutator migration is designed and
every caller has moved (re-evaluate porting the write half).

## Verification

- `go test ./internal/docsannotate -run 'TestMarker|TestStruck|TestSabotage|TestRouted'`
- `go test ./internal/docsrows` (includes the writer-oracle parity test where
  the local skill fan-out exists)
- `bash scripts/check-skills.sh` in the SKILLS repo (subcommand map pin +
  self-test + no-binary guard)
