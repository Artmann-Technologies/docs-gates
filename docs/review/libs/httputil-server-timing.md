# httputil/server_timing utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: server_timing/v1.0.3 (latest local tag, fetched 2026-11-23); the submodule sits in the httputil checkout at HEAD 34dbf05f670d28830e61170a087f552070279b74, module `github.com/larsartmann/httputil/server_timing`, package `servertiming`.
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 2 (HTTP/serialization) subagent of the multi-library utilization audit fleet.
- Fleet stance addressed: does stdlib-only insulation here avoid re-implementation debt, or leave docs-gates hand-rolling something the fleet solved?

## Verdict

NOT-APPLICABLE — server_timing serializes per-request sub-metrics into the W3C `Server-Timing` HTTP response header, and docs-gates has no HTTP responses; its timing story (per-gate wall clock, `duration_ms` rows, a half-budget watchdog, a goroutine-abandonment timeout) is in-process stdlib `time` arithmetic over CLI gate functions — adjacent in concept, unrelated in mechanism and destination.

## Score

- Inputs N = 8 (inventory rows below).
- A = 1 (the per-unit wall-clock duration-capture concept row — the battery genuinely measures per-gate durations; satisfied by two stdlib calls).
- M = 0.
- Score = round(100 × 0/1) = **0** — band **low**. Expected for NOT-APPLICABLE; the verdict carries the meaning.

- Applicable rows are exactly the inventory rows graded **Missed Opportunity** (the consumer has the underlying need; the capability is hand-rolled, solved in-domain, or declined by contract). Every **Not Applicable** row carries a one-line reason. `M` counts rows graded Fully Leveraged.

## What the library offers

| # | Capability                                                                                                                     | Applies to docs-gates?                                                                                           | Status                                                                                                                             | Evidence                                                                                                  |
| - | ------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| 1 | `ServerTimingMiddleware` — installs a concurrency-safe collector per request, stamps the `Server-Timing` header at first write | No — no `http.Handler`, no `ResponseWriter`, no response headers                                                 | N/A                                                                                                                                | httputil/server_timing/server_timing.go:400; httputil/server_timing/README.md:53-55                       |
| 2 | `MeasureServerTiming(ctx, name)` timed-region helper returning a stop func                                                     | No — the concept row 8 below is the non-HTTP analog                                                              | N/A                                                                                                                                | httputil/server_timing/server_timing.go:258                                                               |
| 3 | `RecordServerTiming(ctx, name, desc, dur)` for already-measured durations                                                      | No                                                                                                               | N/A                                                                                                                                | httputil/server_timing/server_timing.go:237                                                               |
| 4 | `total` auto-metric (time to first byte)                                                                                       | No — no first byte; gates finish before any output is released                                                   | N/A                                                                                                                                | httputil/server_timing/README.md:43-46                                                                    |
| 5 | Conditional instrumentation (`ServerTimingMiddlewareWhen(pred)`) — gate the header to admin/debug requests                     | No — nothing to gate; no requests                                                                                | N/A                                                                                                                                | httputil/server_timing/server_timing.go:416                                                               |
| 6 | Injection-safe header serialization (RFC 7230 token sanitization, quoted-string escaping, fuzz-tested)                         | No — no header values are ever constructed                                                                       | N/A                                                                                                                                | httputil/server_timing/README.md:64                                                                       |
| 7 | Capability-preserving writer wrapper (Flush/Hijack/Push/Unwrap delegation, `http.ResponseController` transparency)             | No — the consumer's writer swapping is `os.Stdout`/`os.Stderr` pipe swapping, not `http.ResponseWriter` wrapping | N/A                                                                                                                                | httputil/server_timing/server_timing.go:270, :309                                                         |
| 8 | Per-unit wall-clock duration capture (the concept under the header)                                                            | Yes — the battery times every gate                                                                               | Missed Opportunity (solved with stdlib) — `time.Now`/`time.Since` per gate, printed as heartbeat lines and `duration_ms` JSON rows | consumer internal/battery/battery.go:197-203, :255; httputil/server_timing exports no non-HTTP timing API |

Module shape worth noting: this submodule is itself stdlib-only with zero
dependencies (httputil/server_timing/go.mod:1-3, httputil/server_timing/README.md:9, :61),
so unlike the httputil root it would not drag a dependency tree — the bar to
adoption is purely "there is no HTTP response to stamp", not hermeticity.

## What docs-gates does today

Zero usage, zero presence. Verified:

```console
$ rg -n "server_timing|servertiming|Server-Timing" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**' --type go
# rc=1, no matches
$ rg -n "github.com/larsartmann/httputil/server_timing" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 — not in any module graph
```

The consumer's timing machinery, examined honestly for the adjacency the brief
asks about (`gate_timeout_sec` watchdog + `CaptureOutputWithTimeout`):

- **Per-gate duration**: `runGates` wraps each gate in `time.Now()`/`time.Since`
  (internal/battery/battery.go:197-203) and carries the result on the row
  (`duration`), surfaced as the human heartbeat "done in %s" and the machine
  `duration_ms` field (internal/battery/battery.go:222, :255, :287).
- **Deadline enforcement**: `gatecli.CaptureOutputWithTimeout`
  (internal/gatecli/gatecli_capture.go:121) runs the gate in a goroutine and
  abandons it via `select` on a `time.Timer` — an expired gate becomes a FAIL
  row and the battery moves on (internal/battery/battery.go:209-216). The
  gate functions deliberately take no `context`
  (internal/gatecli/gatecli_capture.go:116-119), so no cancellation propagates.
- **Half-budget watchdog**: `startHalfBudgetWatchdog`
  (internal/battery/battery_watchdog.go:25) fires a stderr-only informational
  line when a gate passes half its wall-clock budget.

**Adjacency judgment:** concept-adjacent, mechanism-unrelated. Both libraries
answer "how long did this unit take, and what do we do when it exceeds a
budget" — but server_timing's answer is _serialize sub-metrics into an HTTP
response header for browser DevTools/RUM to read_ (httputil/server_timing/README.md:7),
while docs-gates' answer is _print a duration in a CLI matrix / JSON row and
abandon the straggler goroutine_. There is no `http.ResponseWriter` in
docs-gates to wrap, no header to stamp, and no consumer of a Server-Timing
header (the only machine consumer of battery output is pkg/provider, which
parses JSON — pkg/provider/provider.go:92-102). The battery's three timing
pieces are also each ~10-30 lines of direct stdlib `time` usage with
byte-pinned output; nothing in them re-implements header serialization, token
sanitization, or writer delegation — the parts server_timing actually
packages.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** server_timing solves a shared problem for
the fleet's _HTTP services_ — injectable request-phase timing visible in
DevTools without vendor agents. docs-gates is not such a service, and its
timing needs are already met by three small, golden-pinned stdlib constructs.
The honest answer to "is docs-gates hand-rolling something the fleet solved?"
is no: the overlap is the trivial half (a `time.Duration` measurement — two
stdlib calls), while everything the library packages (the header format,
escaping, writer delegation, TFB finalization) has no docs-gates counterpart
to replace. Naming a duplicated file:line here would be manufacturing overlap
that does not exist.

**Root-contract bar.** Even though this submodule is itself dependency-free,
importing it into the root module would still add a `require` line and break
`scripts/stack-drift-check.sh` CI (the contract fails on any require,
stack-drift-check.sh:8-9) — for a package whose entire exported surface
type-anchors on `net/http`. There is also nothing to time through its API:
`MeasureServerTiming` takes a `context.Context` that must carry the
middleware's collector (httputil/server_timing/server_timing.go:222-235), and
no docs-gates gate runs under such a context.

**pkg/provider check.** No target: Detect runs the battery binary, parses
`run --json` (already carrying `duration_ms`), and reports findings. Adding
Server-Timing would mean adding an HTTP server first — a different project.

RECONSIDER-WHEN: docs-gates (or a sibling module here) exposes the battery
over HTTP — a fleet status endpoint or a dashboard backend. Server-Timing on
that endpoint's per-gate phases is then an exact fit, and because the
submodule is zero-dependency it can even live in a stdlib-adjacent module
without hermeticity pain. Until an HTTP surface exists, the insulation is
correct.

## Ranked opportunities

None.

## Release shape (context for any future adoption)

`server_timing` carries a day-one-v1 posture (v1.0.3 is the fourth v1.x tag;
the httputil owner ruled it the stable tier, versus `endpoints`' pre-1.0
iteration tier — httputil/CHANGELOG.md:61), and its README is explicit that
the module exists precisely so services can add Server-Timing _without_ the
root httputil dependency set (httputil/server_timing/README.md:9, :61). The
consumption pattern for the fleet's services is settled: middleware in the
chain, `MeasureServerTiming` from handlers, `ServerTimingMiddlewareWhen` to
gate it in production (httputil/server_timing/README.md:62). None of that
surface has a CLI-battery analog, which is the whole verdict in one line.

## How to verify

```console
$ git -C /home/lars/projects/httputil tag --list 'server_timing/*' | sort | tail -1
server_timing/v1.0.3
$ cat /home/lars/projects/httputil/server_timing/go.mod
module github.com/larsartmann/httputil/server_timing
go 1.27
# (no requires — the submodule is stdlib-only)

$ rg -n "server_timing|servertiming|Server-Timing" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**' --type go
# rc=1 — zero hits
$ rg -n "github.com/larsartmann/httputil" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 — not in any module graph

# the consumer's own timing machinery (all stdlib, no header serialization):
$ rg -n "start := time.Now|elapsed := time.Since|DurationMS" /home/lars/projects/docs-gates/internal/battery/battery.go
197:		start := time.Now()
203:		elapsed := time.Since(start)
255:	DurationMS int64  `json:"duration_ms"`
287:			DurationMS: row.duration.Milliseconds(),
$ rg -n "func CaptureOutputWithTimeout|func startHalfBudgetWatchdog" /home/lars/projects/docs-gates/internal/gatecli/gatecli_capture.go /home/lars/projects/docs-gates/internal/battery/battery_watchdog.go
/home/lars/projects/docs-gates/internal/battery/battery_watchdog.go:25:func startHalfBudgetWatchdog(gateName string, timeout time.Duration, warnf func(string, ...any)) func() {
/home/lars/projects/docs-gates/internal/gatecli/gatecli_capture.go:121:func CaptureOutputWithTimeout(fn func() int, timeout time.Duration) (CapturedOutput, bool) {
```
