# ADR 0002 — Destructive authority: `strike` vs `status-items`

**Status:** Accepted, 2026-11-15
**Deciders:** owner + the docs-health tooling audit
**Related:** [ADR 0001](0001-marker-grammar-ownership.md)

## Context

Two commands can put a verdict on a document:

- `docs-gates strike` (this repo) — strikes ONE row in the living
  `TODO_LIST.md`, addressed by a unique text anchor. Dry-run by default;
  refuses ambiguous anchors, double strikes, and verdict-opener targets.
  `docs-gates purge` is its cleanup counterpart.
- `status-items` (docs-health writer) — strikes every item in a WHOLE
  historical report from a generated spec file. Refuses duplicate keys and
  open-family verdicts; atomic per file.

They overlap in output shape (`~~original~~ verdict`) and nowhere else.

## Decision

Keep two authorities. The boundary is the **target class**:

| Authority                     | Target                        | Address model                      | Failure model                                   |
| ----------------------------- | ----------------------------- | ---------------------------------- | ----------------------------------------------- |
| `docs-gates strike` / `purge` | living board (`TODO_LIST.md`) | unique text anchor                 | refuse on ambiguity; reversible only by editing |
| writer `status-items`         | historical report             | key per item (`N` / `N@substring`) | refuse duplicate keys; atomic per file          |

Neither may grow into the other's target. A living-doc row is struck through
the anchored command so the board's order and counts stay stable; a
historical report is annotated through the writer because its item numbering
and section scoping are the writer's contract.

The `purge` asymmetry is intentional: `purge` collects struck TODO rows into
a CHANGELOG entry, while no counterpart exists for historical reports (their
archive manifest is the record). Do not add a "purge" for reports.

## Consequences

- Callers pick by file class, not by convenience.
- `docs-gates strike` never touches `docs/status/`; `status-items` never
  touches `TODO_LIST.md`.
- A future write surface (see the mutator design) must preserve this split.

## Verification

- `internal/docsstrike` tests pin the refusal matrix (ambiguous, double,
  verdict-opener).
- `docs-health/assets/annotate-status-items_test.sh` pins the writer's 46
  assertions (duplicate keys, open verdicts, dry-run, atomicity).
