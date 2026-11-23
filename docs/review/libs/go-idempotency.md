# go-idempotency utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v0.6.0 (latest local tag, fetched 2026-11-23) + HEAD 8fd593eef46b1f50f6ee7138f639c915a5392f08 (`git describe` from HEAD → `idemtest/v0.1.1-5-g8fd593e`; the submodule tag is nearest). Nested module: `idemtest` v0.1.1.
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 1 (resilience/error family) subagent of the multi-library utilization audit fleet.

## Verdict

NOT-APPLICABLE — go-idempotency deduplicates retried operations behind an atomic key store for at-least-once delivery systems, a problem class docs-gates structurally does not have (no network, no retries, no concurrency, no delivery layer), and its guarded mutators already achieve at-most-once semantics with dry-run defaults plus in-domain markdown predicates.

## Score

- Inputs N = 11 (inventory rows below).
- A = 1 (only the repeat-application-guard concept row is a problem docs-gates actually has; solved in-domain).
- M = 0.
- Score = round(100 × 0/1) = **0** — band **low**. The number measures library leverage, which is correctly zero; the verdict carries the meaning.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | `Store` interface — `Seen`/`Record`/`CheckAndRecord` with the atomicity contract (check+record under one lock, concurrent same-key callers serialized) | No — no concurrent callers, no key space, no pluggable backend | N/A — single-threaded CLI, one process, no delivery | go-idempotency/store.go:77 |
| 2 | Typed sentinels `ErrDuplicate`, `ErrInvalidTTL` (errorfamily `Conflict`/`Rejection`) | No — consumer refusals are exit-2 CLI text, not branchable sentinels | N/A — go-idempotency/store.go:21,41; consumer exit contract internal/gatecli/gatecli.go:29-32 | |
| 3 | `MemoryStore` (deprecated, dev/test only) | No | N/A — upstream itself deprecates it; nothing to dedup in-process | go-idempotency/store.go:157 |
| 4 | TTL/expiry semantics — `Forever` sentinel, `isLive`, sweep goroutine | No — no claims expire; nothing is keyed | N/A | go-idempotency/store.go:63,227,294 |
| 5 | `WithClock` injectable clock for deterministic TTL tests | No — no clock-driven behavior to test | N/A | go-idempotency/clock.go:29 |
| 6 | `contract/` conformance suite for validating a `Store` implementation | No — consumer implements no Store | N/A | go-idempotency/contract/contract.go |
| 7 | `middleware/` net/http `Idempotency-Key` adapter | No — no net/http anywhere in the consumer | N/A | go-idempotency/middleware/http.go |
| 8 | `middleware` command-wrapper idempotency (CQRS command dedup) | No — no command bus; verbs are direct argv dispatch | N/A — cmd/docs-gates/verbs.go dispatches `Run([]string) int` | go-idempotency/middleware/middleware.go |
| 9 | `idemtest` v0.1.1 — drill kit (`RequireAllowsAfterExpiry`, `SendConcurrently`, sabotage harness, `PinnedClock`) | No — test kit for Store/pipeline implementers only | N/A | go-idempotency/idemtest/drill.go |
| 10 | Runtime dependency on go-error-family v0.12.0 | No — see the go-error-family report; root adoption would transitively break stdlib-only twice over | N/A | go-idempotency/go.mod:8 |
| 11 | The concept: turn at-least-once application into at-most-once | Yes — strike/purge/archive are guarded mutators that must never apply twice | N/A (solved in-domain) — the markdown itself is the claim record; guards are parse predicates + dry-run default, needing no store, key, TTL, or lock | Consumer sites in "What docs-gates does today" |

## What docs-gates does today

Zero usage — not in code, not in either module's dependency graph. Verified:

```console
$ rg -n "idempotency|go-idempotency|idemtest" /home/lars/projects/docs-gates --type go -g '!docs/review/**'
# rc=1, no matches

$ rg -n "go-idempotency" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1, no matches
```

The one textual hit in the tree is this audit fleet's own scaffolding — `scripts/citation-sweep.py:31` lists `go-idempotency` among the reviewed library names — not product code.

What docs-gates does instead for at-most-once semantics (the honest duplication check the fleet argument demands):

- **Dry-run default, `--apply` opt-in** — the primary device; nothing is written until the operator says so: `internal/docsstrike/docsstrike.go:153` (plain `os.WriteFile`), `internal/docsarchive/docsarchive.go:229`, `internal/docspurge/docspurge_collect.go:192`, `internal/docsfreshness/preflight_land.go:58` (the pre-flight stamp).
- **Refuse-the-repeat predicates over the corpus state** — the markdown is the claim record:
  - never double-strike: `internal/docsstrike/strike.go:166` ("row %d already carries a strike marker — never double-strike", exit 2); the guard contract is stated at `internal/docsstrike/docsstrike.go:25`.
  - never double-archive: `internal/docsarchive/docsarchive.go:183` ("already carries the archive banner — refusing to double-archive"; the corruption narrative at `:14`).
  - purge refuses struck rows whose status cell reads OPEN: shared predicates `internal/docscadence/classify.go:52` (`IsDoneCell`) and `:66` (`IsOpenCell`).
- **Line-level collision detection in pre-flight** — `internal/docsfreshness/preflight_refusals.go:59` (`handleCollisions`) refuses a landing batch whose content lines already exist in the living docs (with the `--duplicates` escape hatch) — text dedup on a candidate list, not operation dedup.
- **Read-only battery by contract** — `pkg/provider/provider.go:224` (`Outputs: nil`, "read-only: the battery never writes the tree it judges"): BuildFlow may re-run Detect at will because it is purely idempotent in the functional sense — no claims needed.

None of these duplicate a store, key, TTL, or atomic check-and-record — the four things go-idempotency actually implements. There is no `file:line` in docs-gates where a go-idempotency mechanism is being hand-rolled.

## Gap analysis — should docs-gates adopt it?

No. The fleet argument ("stop re-implementing; let every project benefit from collective fixes") bites only where a project maintains a private copy of a shared, evolving mechanism. Here there is no copy: docs-gates' repeat-guards are domain predicates (is this row already struck? does this report carry the banner?) over data that is itself the persistent state — exactly the case where a generic key-value claim store adds a dependency, a clock, a sweep goroutine, and a second source of truth without removing one line of logic. The library's own framing confirms the mismatch: it exists for at-least-once delivery with concurrent retries (CQRS command handling, HTTP clients resubmitting) — docs-gates has no delivery layer, no retries (verified: zero `time.Sleep`/backoff loops in non-test code; see the go-retry report), and no concurrency beyond one watchdog goroutine (`internal/battery/battery_watchdog.go`).

Adoption into the ROOT is doubly barred: it breaks the zero-require contract (`scripts/stack-drift-check.sh` fails CI) and drags go-error-family in transitively. Adoption into `pkg/provider` has no target: the provider is read-only (`provider.go:224`, Repair nil) — the classic reason to need idempotency keys (a retried mutation) cannot arise.

RECONSIDER-WHEN: docs-gates (or its provider) gains a mutating, retried, or asynchronous surface — e.g., a Repair phase fed by BuildFlow retries, a daemon applying strikes from a queue, or any endpoint where the same logical change can be delivered twice. That new module should adopt go-idempotency with a Store implemented over the repo state, keeping the root stdlib-only.

## Ranked opportunities

None.

## How to verify

```console
$ git -C /home/lars/projects/go-idempotency log -1 --format='%H'
8fd593eef46b1f50f6ee7138f639c915a5392f08
$ git -C /home/lars/projects/go-idempotency tag --list | tail -3
v0.5.0
v0.6.0
$ git -C /home/lars/projects/go-idempotency describe --tags
idemtest/v0.1.1-5-g8fd593e   # nearest tag is the nested idemtest/v0.1.1; root tag v0.6.0 exists

$ rg -n "idempotency|go-idempotency|idemtest" /home/lars/projects/docs-gates --type go -g '!docs/review/**'
# rc=1 (zero Go-file hits)
$ rg -n "go-idempotency" /home/lars/projects/docs-gates -g '!docs/review/**'
scripts/citation-sweep.py:31:    "go-idempotency",        # audit-fleet scaffolding only
$ rg -n "go-idempotency" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 (not in any module graph)

# in-domain repeat guards (spot-check):
$ rg -n "already carries|refusing to double" /home/lars/projects/docs-gates/internal --type go | rg -v _test
internal/docsarchive/docsarchive.go:183: ... refusing to double-archive ...
internal/docsstrike/strike.go:166: ... never double-strike ...
```
