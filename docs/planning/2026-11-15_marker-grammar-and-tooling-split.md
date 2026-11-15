# Marker grammar and docs-health tooling split

**Date:** 2026-11-15
**Status:** Accepted and executing (read-only half landed)
**ADR:** [0001-marker-grammar-ownership.md](../adr/0001-marker-grammar-ownership.md)

## Problem

Two repositories owned one contract without an executable link:

- the docs-health skill's `assets/annotate.go` (writer: renders verdict
  markers, mutates reports), and
- this binary (judge: recognises verdict markers, judges a corpus).

The agreement held only because today's seven rendered phrases happened to
contain keywords `markerRe` matched. Separately, the skill shipped a prebuilt
`assets/annotate` binary that had drifted 12 days behind its source and was
missing four subcommands.

## Decision

Split by **read/write**, not by repository:

1. **Read-only, gate-shaped verbs move here.** `docs-gates rows` (the writer's
   `check-rows`) and `docs-gates rows --bare` (the writer's `bare-rows`) judge
   explicit file paths and own their exit contract (0 clean, 1 findings,
   2 usage/environment).
2. **Mutators stay in the skill.** `rows`, `prose`, `status-items`, `repair`,
   `unwrap`, `repair-mangle`, `repair-partial` (plus the read-only
   `emit-keys` and `prose-items`) are not ported until an input model, shim,
   and rollback path are designed. A public binary must not absorb unproven
   file-rewriting code.
3. **The grammar is executable.** `internal/docsannotate/markers_test.go`
   pins one rendered phrase per kind, runs the gate over all six strike
   verdicts, and fails when a marker is renamed (sabotage case). The routed
   kind `r` is pinned as intentionally unstruck: `rows` counts the arrow form
   as ROUTED instead.
4. **One writer artifact.** The prebuilt `assets/annotate` binary is deleted;
   every caller runs `go run annotate.go`. `check-skills.sh` refuses a rebuilt
   binary beside the source.

## What moved, what stayed

| Surface                                               | Owner now              | Evidence                                            |
| ----------------------------------------------------- | ---------------------- | --------------------------------------------------- |
| `rows` (completeness), `rows --bare`                  | this repo              | `internal/docsrows`, parity test against the writer |
| `check-rows`, `bare-rows` aliases                     | writer (kept)          | callers pinned to `annotate.go`                     |
| `self-test`, all mutators, `emit-keys`, `prose-items` | writer                 | `docs-health/SKILL.md` subcommand map               |
| marker rendering                                      | writer (`markerFor`)   | pinned by `markers_test.go`                         |
| marker recognition                                    | this repo (`markerRe`) | pinned by `markers_test.go`                         |

## Caller inventory (executable references)

| Repo        | File                            | Uses                                        | Failure posture                                   |
| ----------- | ------------------------------- | ------------------------------------------- | ------------------------------------------------- |
| crm         | `scripts/docs-health-gate.sh`   | `go run annotate.go check-rows`             | fails loudly when helper/go missing (2026-11-15)  |
| CV          | `scripts/docs-archive-sweep.sh` | `go run annotate.go`                        | `die` on missing source or go (2026-11-15)        |
| go-appkit   | `scripts/docs-annotate-pass.sh` | `go run annotate.go check-rows`             | fails loudly; `ANNOTATE_BIN` retired (2026-11-15) |
| pbx-artmann | `scripts/docs-gates.sh`         | `go run annotate.go repair-mangle`          | `fatal` when helper/go missing                    |
| pbx-artmann | `scripts/docs-archive.sh`       | `go run annotate.go status-items`           | notice + degraded phases; battery gate still runs |
| webphone    | `scripts/docs-gates.sh`         | `go run annotate.go repair-mangle`          | local: WARN + explicit SKIPPED row; CI: fatal     |
| webphone    | `scripts/cadence-candidates.sh` | builds `/tmp/annotool/annotate` from source | `exit 1` on build failure                         |

No executable caller used the deleted prebuilt binary after the go-appkit
conversion; prose references in status reports are historical.

## Install coverage of `docs-gates`

<!-- coverage-survey:start -->

Surveyed 2026-11-15 across every checkout in `/home/lars/projects` that
shows execution evidence (script, CI job, flake check, or `.docs-gates.json`).

| Provided by                      | Count       | Representative repos                                                                                                                                       |
| -------------------------------- | ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| PINNED (tag/rev/input)           | 13          | cmdguard, crm, go-aichat, go-atomic-write, graph-features-language, io-surface, mr-sync, pbx-artmann, pim-sync, typespec-go, webphone                      |
| PATH / wrapper                   | 9           | cqrs-htmx, go-taskqueue, GoReleaser-Wizard, monitor365-ui-sdk, nixos-modules, ratsgedaechtnis, storbi, timesheets, touchbase                               |
| ABSENT (annotate.go only)        | 2 checkouts | cv/CV, go-appkit                                                                                                                                           |
| UNKNOWN (config only, no runner) | 11          | go-cqrs-lite, go-filewatcher, go-finding, go-health-dashboard, InboxClean, journal, media-atlas, pbx-lens, SKILLS, nix-international-telephony, quotaboard |

Notes:

- `quotaboard/scripts` hardcodes `/home/lars/.local/bin/docs-gates`, which does
  not exist on this host (the fleet binary is
  `/run/current-system/sw/bin/docs-gates`); its workflow would fail before
  reaching any gate.
- No repo ships a prebuilt `annotate`; every consumer runs the source.
  `webphone/scripts/cadence-candidates.sh` builds the source into `/tmp`.
- The two ABSENT repos are the archive/annotation tools half, not full-battery
  consumers; they cannot adopt `docs-gates rows` until they call the binary.

Conclusion: the read-only half is safe to move — every full-battery consumer
has the binary pinned or on PATH. The write half still needs the ABSENT rows
converted plus the mutator design.

<!-- coverage-survey:end -->

## Sequencing

1. Land the read-only port + parity tests (`internal/docsrows`). **Done.**
2. Delete the stale binary + guard (`check-skills.sh`). **Done.**
3. Convert/harden every executable caller. **Done.**
4. Pin the grammar (`markers_test.go`) and document the two halves
   (`docs-health/SKILL.md`). **Done.**
5. Design the mutator migration (input model, shim, rollback). **Open** — the
   mutators stay where they are until that design is accepted.

## Rollback

- The read-only port is additive. `annotate.go check-rows` / `bare-rows`
  still work, so a caller can switch back with no data migration.
- Reverting the deletion is a `git revert` of the skill commit; the guard then
  fails by design and must be reverted with it.
- No corpus was rewritten by this change; the only file repaired was a
  formatter-damaged archived report (restored from history + per-cell
  strikes).

## Revisit triggers

- A second consumer needs to render markers → promote the grammar to a shared
  module.
- The mutator migration design is accepted and all callers converted →
  re-evaluate porting the write half.
- An external repo adopts the ritual without the docs-health skill → decide
  whether the writer becomes installable rather than fan-out-only.
