# go-daemon utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v0.4.0 (latest local tag, fetched 2026-11-23) + HEAD 6b9ee1d383c365d0e7579ed51042f61149755a1f (`git describe` shows `v0.4.0-39-g6b9ee1d`). Single module `github.com/LarsArtmann/go-daemon`; deps go-systemd/v22, go-codec, go-flightrecorder, go-sse/sseparse, prometheus/client_golang plus a heavy indirect tree (go-daemon/go.mod:5-25).
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f baseline — root module stdlib-only (scripts/stack-drift-check.sh:9) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 3 (I/O + runtime + output) subagent of the multi-library utilization audit fleet.

## Verdict

NOT-APPLICABLE — docs-gates is a run-to-completion CLI with no socket, no listener, no signal handling, and no supervision surface for go-daemon's lifecycle machinery to own, and the one daemon adjacent to the repo (the fleet auto-commit daemon) lives outside it entirely.

## Score

- Inputs N = 9 (inventory rows below).
- A = 0 (no row maps onto any behavior docs-gates has or wants; the nearest adjacency — the per-gate watchdog timeout — is a bounded-wait concern already solved with one-shot timers inside the run).
- M = 0.
- Score = **0** (0/0; by convention 0 when nothing is applicable) — band **low**. For a NOT-APPLICABLE library the score is noise; the verdict is the signal.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | `NewSocketServer`/`NewServer` — full socket lifecycle: stale-socket cleanup, 0600 perms, ordered `Lifecycle` hooks | No — the binary never listens on anything | N/A | go-daemon/socket.go:158-237 |
| 2 | Graceful shutdown: `sd_notify STOPPING=1`, drain within `WithShutdownGracePeriod` | No — there is nothing to drain; each verb writes its verdict lines and exits | N/A | go-daemon/socket.go:126-128, :207 |
| 3 | `NotifyReady`/`NotifyStopping`/`Watchdog` — sd_notify integration with observability counters | No — not a systemd unit; `WatchdogInterval` would report `false` everywhere by design | N/A | go-daemon/notify.go:59-102 |
| 4 | `NewHTTPClient`/`NewHTTPClientNoTimeout` — unix-socket HTTP transports | No — the consumer has zero `net/http` imports (verified below); its only IPC is exec-ing itself from pkg/provider | N/A | go-daemon/transport.go:25-41; pkg/provider/provider.go:149 |
| 5 | Content negotiation — `ResolveRequestCodec`/`ResolveResponseCodec` (JSON/CBOR, q-values, `q=0` refusals) | No — the battery speaks exactly two fixed faces: pinned human lines and one snake_case JSON document | N/A | go-daemon/negotiation.go:73-98; internal/battery/battery.go:264 |
| 6 | `WriteEncoded`/`WriteJSON`/`WriteErrorf` — negotiated, pooled-buffer response writes | No — no `http.ResponseWriter` exists in the tree | N/A | go-daemon/write.go:29-66 |
| 7 | `ParseSSEData` — WHATWG-conformant SSE framing parser (WPT-corpus-pinned) | No — no SSE anywhere (that is go-datastar/go-sse territory; see that report) | N/A | go-daemon/sse.go:42 |
| 8 | `FlightRecorder` — execution-trace recorder with slow-request triggers and Prometheus metrics | No — the battery's observability surface is its exit code plus byte-pinned stdout/stderr | N/A | go-daemon/flightrecorder.go:57-165 |
| 9 | `IsSocketAlive` liveness probe | No — no socket file to probe | N/A | go-daemon/socket.go:418 |

## What docs-gates does today

Zero usage, zero presence in either module graph. Verified:

```console
$ rg -n "go-daemon|godmn" /home/lars/projects/docs-gates -g '!docs/review/**' --type go
# rc=1, no matches
$ rg -n "go-daemon" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1, no matches
```

(The word "daemon" in the repo refers to the FLEET auto-commit daemon — external infrastructure, ~2-3 min cadence, never pushed — and to "foreign daemons" in the point-in-time verdict doctrine, AGENTS.md:55-59 and :188-189. No Go code in this repo implements, supervises, or talks to any daemon.)

There is likewise no hand-rolled daemon-equivalent to deduplicate:

- **No signal handling**: `rg -n "signal\.|os\.Signal|SIGTERM|SIGHUP|syscall\.SIG" internal/ cmd/ pkg/ --type go` → rc=1. A gate killed by SIGKILL is bounded by the caller (`timeout 300 go run …`, AGENTS.md:213); a wedged gate inside the battery is bounded by `gate_timeout_sec`, which converts it into a named FAIL row (`rc=timeout`) and moves on — the anti-daemon answer to a hang.
- **No listener or serve loop**: `rg -n "ListenAndServe|http\.Server|net\.Listen" internal/ cmd/ --type go` → rc=1. The process-graph edges are: `main.go` → verb `Run` → gates; `pkg/provider` execs the binary from PATH once per BuildFlow Detect (pkg/provider/provider.go:149); flake wrappers exec `docs-gates <verb>` under legacy names. Every edge is one-shot exec with an exit code.
- **No persistence between runs**: the only cross-run state is the gitignored land-parity stamp (internal/docsfreshness/preflight_land.go:28), explicitly "local editor state, never shared" — the opposite of a daemon's coordination state.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** Both prongs score zero. docs-gates re-implements nothing go-daemon offers — there is no socket, no graceful shutdown, no readiness protocol, no negotiation, no supervision anywhere in the tree. And the problems go-daemon solves (long-lived socket servers, systemd integration, content negotiation across clients, trace capture on slow requests) belong to a serving process; docs-gates' product is a deterministic batch verdict whose every output byte is pinned by parity goldens (internal/battery/parity_test.go:188) and whose contract is three exit codes (internal/gatecli/gatecli.go:27-29). Turning it into a served daemon would invert the architecture the extraction was FOR: fourteen verbs, each a library `Run([]string) int`, dispatched once.

**Root-contract bar.** Barred outright and doubly: five direct requires including prometheus/client_golang, and a socket server would reintroduce the long-running-process class the consumer's "no long-running process" boundary excludes. pkg/provider is equally wrong a home — it is a BuildFlow plugin that registers tools and execs a binary; it does not serve.

RECONSIDER-WHEN: docs-gates ever grows a resident service — the only realistic shape is a `docs-gates serve` mode exposing the battery over a unix socket for editors/CI to poll verdicts live. If that day comes, go-daemon (in a nested module) is the right chassis for it, and the parity goldens would need a deliberate carve-out for the served faces. Nothing in the current corpus or ROADMAP points there.

## Ranked opportunities

None.

## How to verify

```console
$ git -C /home/lars/projects/go-daemon log -1 --format='%H'
6b9ee1d383c365d0e7579ed51042f61149755a1f
$ git -C /home/lars/projects/go-daemon describe --tags
v0.4.0-39-g6b9ee1d
$ git -C /home/lars/projects/go-daemon tag --list | tail -2
v0.3.1
v0.4.0

$ rg -n "go-daemon|godmn" /home/lars/projects/docs-gates -g '!docs/review/**' --type go
# rc=1 (zero Go-file hits)
$ rg -n "go-daemon" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 (not in any module graph)

$ rg -n "signal\.|os\.Signal|SIGTERM|SIGHUP|syscall\.SIG" /home/lars/projects/docs-gates/internal /home/lars/projects/docs-gates/cmd /home/lars/projects/docs-gates/pkg --type go
# rc=1 — no signal handling anywhere
$ rg -n "ListenAndServe|http\.Server|net\.Listen" /home/lars/projects/docs-gates/internal /home/lars/projects/docs-gates/cmd --type go
# rc=1 — no listener, no serve loop

$ rg -n "ExitClean|ExitFindings|ExitUsage" /home/lars/projects/docs-gates/internal/gatecli/gatecli.go | head -3
# 27/28/29 — the whole failure vocabulary is three exit codes
```
