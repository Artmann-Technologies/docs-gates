# go-idempotency utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v0.6.0 (latest local tag, fetched 2026-11-23) + HEAD 8fd593eef46b1f50f6ee7138f639c915a5392f08 (`git describe` from HEAD → `idemtest/v0.1.1-5-g8fd593e`; the nested module's tag is nearest — root tag `v0.6.0` exists and was verified). Nested module: `idemtest` v0.1.1.
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 1 (resilience/error family) subagent of the multi-library utilization audit fleet.
- Fleet stance addressed per library: "if all my projects re-implement everything, I will only fix bugs all day and none benefit from collective improvements" — does stdlib-only insulation here avoid real re-implementation debt, or leave docs-gates hand-rolling something the fleet solved?

## Verdict

NOT-APPLICABLE — go-idempotency deduplicates retried operations behind an atomic key store for at-least-once delivery systems (CQRS command handling, HTTP retries), a problem class docs-gates structurally does not have: no network, no retries, no concurrency across callers, no delivery layer — and its guarded mutators already achieve at-most-once semantics with dry-run defaults plus in-domain markdown predicates.

## Score

- Inputs N = 13 (inventory rows below).
- A = 1 (only the repeat-application-guard concept row is a problem docs-gates actually has; it is solved in-domain).
- M = 0.
- Score = round(100 × 0/1) = **0** — band **low**. The number measures library leverage, which is correctly zero; the verdict carries the meaning.

- Applicable rows are exactly the inventory rows graded **Missed Opportunity** (the consumer has the underlying need; the capability is hand-rolled, solved in-domain, or declined by contract). Every **Not Applicable** row carries a one-line reason. `M` counts rows graded Fully Leveraged.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | `Store.Seen(ctx, key)` — membership probe (racy alone; documented TOCTOU) | No — no concurrent callers, no key space, no pluggable backend | N/A — single-threaded CLI, one process, no delivery | go-idempotency/store.go:201 |
| 2 | `Store.Record(ctx, key, ttl)` — claim write | No | N/A | go-idempotency/store.go:249 |
| 3 | `Store.CheckAndRecord(ctx, key, ttl)` — the atomic primitive: check+record under one lock, concurrent same-key callers serialized, exactly one wins | No — the precondition (concurrent retries of one logical operation) cannot arise | N/A | go-idempotency/store.go:271, README "How it works" |
| 4 | Typed sentinels `ErrDuplicate`, `ErrInvalidTTL` (errorfamily `Conflict`/`Rejection`) | No — consumer refusals are exit-2 CLI text, not branchable sentinels | N/A | go-idempotency/store.go:21,41; consumer internal/gatecli/gatecli.go:29-32 |
| 5 | `MemoryStore` (deprecated, dev/test only) with background sweep goroutine | No — upstream itself deprecates it; nothing to dedup in-process | N/A | go-idempotency/store.go:157,294 |
| 6 | TTL/expiry semantics — `Forever` sentinel (math.MaxInt64), `isLive`, `expiryFor` | No — no claims expire; nothing is keyed | N/A | go-idempotency/store.go:63,227,233 |
| 7 | `WithClock` injectable clock for deterministic TTL tests | No — no clock-driven behavior to test | N/A | go-idempotency/clock.go:29 |
| 8 | `contract/` conformance suite for validating a third-party `Store` implementation (invariants, negative harness, timing, forever) | No — consumer implements no Store | N/A | go-idempotency/contract/contract.go, contract_invariants.go |
| 9 | `middleware/` net/http `Idempotency-Key` adapter | No — no net/http anywhere in the consumer | N/A | go-idempotency/middleware/http.go |
| 10 | `middleware` command-wrapper idempotency (CQRS command dedup) | No — no command bus; verbs are direct argv dispatch | N/A — cmd/docs-gates/verbs.go dispatches `Run([]string) int` | go-idempotency/middleware/middleware.go |
| 11 | `idemtest` v0.1.1 drills — `RequireAllowsAfterExpiry`, `SendConcurrently`, per-drill `Executions` counters | No — test kit for Store/pipeline implementers only | N/A | go-idempotency/idemtest/drill.go, executions.go, http.go |
| 12 | `idemtest` pinned-clock + sabotage harness — `PinnedClock`, `RequireAllowsAfterExpiryPinned`, negative proofs asserting failure messages name the violated contract | No | N/A | go-idempotency/idemtest/clock.go, negative_pinned_test.go |
| 13 | The concept: turn at-least-once application into at-most-once | Yes — strike/purge/archive are guarded mutators that must never apply twice | Missed Opportunity (solved in-domain) — the markdown itself is the claim record; guards are parse predicates + dry-run default, needing no store, key, TTL, or lock | Consumer sites in "What docs-gates does today" |

(Runtime dependency on go-error-family v0.12.0 — see the go-error-family report; it would bar root adoption transitively even if a use existed.)

## What docs-gates does today

Zero usage — not in code, not in either module's dependency graph. Verified:

```console
$ rg -n "idempotency|go-idempotency|idemtest" /home/lars/projects/docs-gates --type go -g '!docs/review/**'
# rc=1, no matches

$ rg -n "go-idempotency" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1, no matches
```

The one textual hit in the tree is this audit fleet's own scaffolding — `scripts/citation-sweep.py:31`
lists `go-idempotency` among the reviewed library names — not product code.

What docs-gates does instead for at-most-once semantics (the honest duplication check the fleet
argument demands):

### Dry-run default, `--apply` opt-in — the primary device

Nothing is written until the operator says so; the write itself is a plain `os.WriteFile` at:

- `internal/docsstrike/docsstrike.go:153` (strike lands the edited TODO table)
- `internal/docsarchive/docsarchive.go:229` (archive's `git mv`/write path)
- `internal/docspurge/docspurge_collect.go:192` (purge's rewritten TODO_LIST)
- `internal/docsfreshness/preflight_land.go:58` (the pre-flight land-parity stamp)

### Refuse-the-repeat predicates — the markdown is the claim record

- never double-strike: `internal/docsstrike/strike.go:166` — `"row %d already carries a strike
  marker — never double-strike"` (exit 2); the guard contract is stated at
  `internal/docsstrike/docsstrike.go:25`.
- never double-archive: `internal/docsarchive/docsarchive.go:183` — `"FATAL: %s already carries
  the archive banner — refusing to double-archive"` (corruption narrative at `:14`).
- purge refuses struck rows whose status cell reads OPEN: shared predicates
  `internal/docscadence/classify.go:52` (`IsDoneCell`) and `:66` (`IsOpenCell`) — refusal = exit
  2, nothing written.

### Line-level collision detection in pre-flight

`internal/docsfreshness/preflight_refusals.go:59` (`handleCollisions`) refuses a landing batch
whose content lines already exist in the living docs (with the explicit `--duplicates` escape
hatch modeling a legitimate +1 DF push) — text dedup on a candidate list, not operation dedup.

### Read-only battery by contract

`pkg/provider/provider.go:224` — `Outputs: nil`, "read-only: the battery never writes the tree it
judges". BuildFlow may re-run Detect at will because it is purely idempotent in the functional
sense — no claims needed. The per-token memoization caches in
`internal/docsfreshness/corpus.go:84-85` are a performance cache inside one process's one run
(the 2026-11-05 audit-all cliff fix), not cross-run dedup state.

None of these duplicate a store, key, TTL, or atomic check-and-record — the four things
go-idempotency actually implements. There is no `file:line` in docs-gates where a
go-idempotency mechanism is being hand-rolled.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** The argument bites only where a project maintains a private
copy of a shared, evolving mechanism. Here there is no copy: docs-gates' repeat-guards are domain
predicates (is this row already struck? does this report carry the banner?) over data that is
itself the persistent state — exactly the case where a generic key-value claim store adds a
dependency, a clock, a sweep goroutine, and a second source of truth without removing one line
of logic. The library's own framing confirms the mismatch: it exists for at-least-once delivery
with concurrent retries; docs-gates has no delivery layer, no retries (verified: zero
`time.Sleep`/backoff loops in non-test code — see the go-retry report), and no concurrency
beyond one informational watchdog goroutine (`internal/battery/battery_watchdog.go`).

**Root-contract bar.** Adoption into the ROOT is doubly barred: it breaks the zero-require
contract (`scripts/stack-drift-check.sh` fails CI) and drags go-error-family v0.12.0 in
transitively.

**pkg/provider check.** No target: the provider is read-only (`provider.go:224`, Repair nil) —
the classic reason to need idempotency keys (a retried mutation) cannot arise. A BuildFlow-side
re-run of Detect re-executes the battery, which judges; judging twice is free of side effects by
contract.

Conclusion: the stdlib-only insulation rightly avoids re-implementation debt — the library
solves a problem docs-gates does not have, and the adjacent problem docs-gates does have
(at-most-once mutation) is solved with devices (dry-run default, state predicates) that are
simpler than the library's mechanism and pinned by the repo's own gates.

RECONSIDER-WHEN: docs-gates (or its provider) gains a mutating, retried, or asynchronous
surface — e.g., a Repair phase fed by BuildFlow retries, a daemon applying strikes from a queue,
or any endpoint where the same logical change can be delivered twice. That new module should
adopt go-idempotency from day one, with a `Store` implemented over the repo's own state; the
root stays stdlib-only.

## Ranked opportunities

None.

## How to verify

```console
$ git -C /home/lars/projects/go-idempotency log -1 --format='%H'
8fd593eef46b1f50f6ee7138f639c915a5392f08
$ git -C /home/lars/projects/go-idempotency tag --list | tail -3
v0.4.0
v0.5.0
v0.6.0
$ git -C /home/lars/projects/go-idempotency describe --tags
idemtest/v0.1.1-5-g8fd593e

$ rg -n "idempotency|go-idempotency|idemtest" /home/lars/projects/docs-gates --type go -g '!docs/review/**'
# rc=1 (zero Go-file hits)
$ rg -n "go-idempotency" /home/lars/projects/docs-gates -g '!docs/review/**'
/home/lars/projects/docs-gates/scripts/citation-sweep.py:31:    "go-idempotency",
$ rg -n "go-idempotency" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 (not in any module graph)

# in-domain repeat guards (spot-check):
$ rg -n "already carries|refusing to double" /home/lars/projects/docs-gates/internal --type go | rg -v _test
/home/lars/projects/docs-gates/internal/docsarchive/docsarchive.go:63:print as skips and never move; a report that already carries a banner
/home/lars/projects/docs-gates/internal/docsarchive/docsarchive.go:183:			gatecli.Errf("FATAL: %s already carries the archive banner — refusing to double-archive\n", report.rel)
/home/lars/projects/docs-gates/internal/docsstrike/strike.go:166:			msg: fmt.Sprintf("row %d already carries a strike marker — never double-strike", idx+1),
/home/lars/projects/docs-gates/internal/docsstrike/docsstrike.go:25://	  - A row that already carries ~~ is refused (never double-strike).
/home/lars/projects/docs-gates/internal/docsannotate/settings.go:74:	// Defaults() anyway, so s already carries the resolved list.
$ rg -n "func IsOpenCell|func IsDoneCell" /home/lars/projects/docs-gates/internal --type go
/home/lars/projects/docs-gates/internal/docscadence/classify.go:52:func IsDoneCell(cell string) bool {
/home/lars/projects/docs-gates/internal/docscadence/classify.go:66:func IsOpenCell(cell string) bool {

$ rg -n "os.WriteFile" /home/lars/projects/docs-gates/internal --type go | rg -v _test
/home/lars/projects/docs-gates/internal/docsarchive/docsarchive.go:229:	return os.WriteFile( //nolint:gosec // G306,G703: the mode IS the source file's; the path argument IS the tool's input contract
/home/lars/projects/docs-gates/internal/docspurge/docspurge_collect.go:192:	return os.WriteFile( //nolint:gosec // G306,G703: the mode IS the source file's; the path argument IS the tool's input contract
/home/lars/projects/docs-gates/internal/docsstrike/docsstrike.go:153:	if err := os.WriteFile( //nolint:gosec // G703: the file argument IS the tool's input contract
/home/lars/projects/docs-gates/internal/docsfreshness/preflight_land.go:58:	_ = os.WriteFile(stampPath(root), blob, stampFileMode)
```
