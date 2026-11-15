# Mutator migration design (docs-health writer → docs-gates)

**Date:** 2026-11-15
**Status:** Design only — nothing in this document is implemented
**Context:** [ADR 0001](../adr/0001-marker-grammar-ownership.md) split the
tooling by read/write. This document is the gate that must be passed before
any destructive verb moves.

## Inventory — all twelve writer subcommands

| Subcommand                 | Class  | Input                                              | Write scope                              | May move?                             |
| -------------------------- | ------ | -------------------------------------------------- | ---------------------------------------- | ------------------------------------- |
| `rows`                     | mutate | file + section prefix + `<n>:<kind>:<value>` specs | table rows in one section                | later, with coverage                  |
| `prose`                    | mutate | file + section prefix + specs                      | prose items in one section               | later, with coverage                  |
| `status-items`             | mutate | file + spec file                                   | every item in one whole file             | later, with coverage                  |
| `repair`                   | mutate | TSV spec file (`path<TAB>lines<TAB>kind[<value>]`) | any file, any line                       | later, with coverage                  |
| `unwrap`                   | mutate | `<file>:<lines>...`                                | struck table rows a formatter re-wrapped | later, with coverage                  |
| `repair-mangle`            | mutate | files + `--dry-run`                                | single-strike/formatter damage           | later, with coverage                  |
| `repair-partial`           | mutate | files + `--dry-run`                                | bare cells of partially-struck rows      | later, with coverage                  |
| `check-rows --normalize`   | mutate | files                                              | drops beyond-count trailing empty cells  | later, with coverage                  |
| `emit-keys`                | read   | file + line numbers                                | none                                     | candidate, low value                  |
| `prose-items`              | read   | files                                              | none                                     | candidate, low value                  |
| `self-test`                | read   | none                                               | none                                     | stays (pins writer-only behaviour)    |
| `check-rows` / `bare-rows` | read   | files                                              | none                                     | **already moved** (`docs-gates rows`) |

## Decisions

### D1 — Explicit paths, not a repo root

Every mutator takes explicit file paths. The moved read-only verbs already do
this. A repo-root corpus model was rejected: the writer's section scoping
(`## f)` by default, `--section "### a)#2"`) is the caller's decision, and a
gate that guessed a corpus would silently judge files the caller did not
name.

### D2 — Destructive verbs default to dry-run

`rows`, `prose`, `status-items`, `repair`, `unwrap`, and the two repair verbs
already refuse to write without an explicit apply/dry-run choice in the
current tool; a port must keep that. `--apply` is the only write switch.

### D3 — No verb named `annotate`

`docs-gates annotate` is already the annotation-quality gate; putting mutators
under it would silently redefine an established contract. A future write
surface uses a distinct verb (candidate: `write`) with the writer's names as
subcommands, or stays a separate binary.

### D4 — Shim contract

While both surfaces exist, a shim is an argument-passthrough:

```
docs-gates write <subcommand> [args...]
  → exec the writer (source path resolved from DOCS_HEALTH_ANNOTATE_GO or
    the skill fan-out) with identical argv
  → identical stdout/stderr
  → exit code passthrough, except launch failures (2)
```

The shim must never re-implement a mutator. Its only job is to make the
migration invisible to callers and reversible in one env var.

### D5 — Rollback

One env var (`DOCS_HEALTH_ANNOTATE_GO=/path/to/annotate.go`) restores the
writer path. Every caller already accepts a source override (crm's
`ANNOTATE_GO`, CV's `DOCS_HEALTH_SKILL_DIR`, pbx-artmann's
`DOCS_HEALTH_ANNOTATE_GO`, webphone's `DOCS_HEALTH_ANNOTATE_GO`). No corpus
migration is involved, so rollback is a config change, not a data restore.

### D6 — Preconditions before the first mutator moves

1. `docs-gates rows` parity has been green for a full release cycle.
2. Every caller in the inventory resolves the new surface by default.
3. Round-trip tests exist for the specific verb (repair: mangle → clean,
   idempotent, content-loss exit 1 — already written for `repair-mangle`).
4. The writer's `self-test` assertions for that verb are ported to Go tests.
5. A documented rollback drill has been run once.

## Migration checklist (per caller)

| Step                                                     | Owner       | Verify                                             |
| -------------------------------------------------------- | ----------- | -------------------------------------------------- |
| Point the caller at the new surface behind an env switch | caller repo | dry-run output identical                           |
| Run the caller's own self-test                           | caller repo | exit 0                                             |
| Flip the default                                         | caller repo | one real pass, no diffs outside the intended files |
| Remove the old path                                      | caller repo | `git grep` finds one resolution                    |
| Record the landing in the caller's changelog/AGENTS      | caller repo | link to this document                              |

## Gap check

All twelve subcommands are accounted for above. Nothing in the writer's
dispatch is unclassified; the two that already moved are named explicitly and
their writer aliases remain.

## Open questions

- Does `prose-items` belong in the read-only port at all, or is it a
  measurement pass that only the corpus sweep needs? (Current answer: leave
  it, no consumer asked for it.)
- Should `check-rows --normalize` move with the read-only verb, or stay with
  the writer because it writes? (Current answer: stays; `docs-gates rows`
  deliberately has no write flag.)
