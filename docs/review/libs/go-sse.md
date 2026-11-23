# go-sse utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v0.7.0 (latest local tag, fetched 2026-11-23) + HEAD d010c2222cae90e01e3f8859992611204e0fecd4 (`git describe` → `v0.7.0-28-gd010c22`). Modules: root `github.com/larsartmann/go-sse` v0.7.0; `sseparse` v0.2.1; `ssetest` v0.5.0.
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 2 (HTTP/serialization) subagent of the multi-library utilization audit fleet.
- Fleet stance addressed: does stdlib-only insulation here avoid re-implementation debt, or leave docs-gates hand-rolling something the fleet solved?

## Verdict

NOT-APPLICABLE — go-sse is a Server-Sent Events transport (wire serialization, connection lifecycle, fan-out broadcasting, reconnection replay) and docs-gates is a batch CLI that holds no connections, streams nothing, and parses no event protocol; the verified nearest neighbor is a whole-file Markdown table scanner with a different grammar, which is not SSE re-implementation in any sense.

## Score

- Inputs N = 10 (inventory rows below; submodules covered).
- A = 1 (the line-oriented stateful text-scanning concept row — docs-gates does run line-driven state machines over text corpora; satisfied by `strings.Split` + small scanner structs, no streaming).
- M = 0.
- Score = round(100 × 0/1) = **0** — band **low**. Expected for NOT-APPLICABLE; the verdict carries the meaning.

- Applicable rows are exactly the inventory rows graded **Missed Opportunity** (the consumer has the underlying need; the capability is hand-rolled, solved in-domain, or declined by contract). Every **Not Applicable** row carries a one-line reason. `M` counts rows graded Fully Leveraged.

## What the library offers

| #  | Capability                                                                                                         | Applies to docs-gates?                                                            | Status                                                                                                                                              | Evidence                                                                  |
| -- | ------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| 1  | `Event`/`EventID` wire types + `WriteEvent` allocation-minimized serializer (multi-line data, CRLF)                | No — no SSE wire output; stdout is matrix lines and one JSON doc                  | N/A                                                                                                                                                 | go-sse/event.go:29, :91, :167                                             |
| 2  | `Stream` — single connection lifecycle: headers, send, heartbeat, context, `Last-Event-ID`                         | No — no connections, no long-running process                                      | N/A                                                                                                                                                 | go-sse/stream.go:38, :66, :83                                             |
| 3  | `Broadcaster[T]` fan-out with subscribe/broadcast/close, drop hooks, bounded buffers                               | No — no subscribers; each run is single-consumer                                  | N/A                                                                                                                                                 | go-sse/fanout.go:24-74                                                    |
| 4  | `EventStore` + `Replay` reconnection replay (missed events on reconnect)                                           | No — no reconnecting clients; the provider execs fresh runs                       | N/A                                                                                                                                                 | go-sse/replay.go; go-sse/README.md:22                                     |
| 5  | `SubscribeFilter` / `FilteredEventStore` predicate push-down                                                       | No                                                                                | N/A                                                                                                                                                 | go-sse/filter_test.go surface; go-sse/README.md:21, :23                   |
| 6  | Heartbeat/comment writer to defeat idle-proxy timeouts                                                             | No — no proxies; the battery's own anti-wedge is the timeout watchdog, in-process | N/A                                                                                                                                                 | go-sse/event.go:239; consumer internal/battery/battery_watchdog.go:25     |
| 7  | `sseparse` (submodule v0.2.1): zero-dependency WHATWG-conformant wire-format parser (`ReadEvents`, `StreamReader`) | No — nothing parses `event:`/`data:`/`id:` lines; see row 10                      | N/A                                                                                                                                                 | go-sse/sseparse/README.md:3-4; go-sse/sseparse/reader.go:90, :197         |
| 8  | `sseparse` WPT conformance corpus (embedded, regenerable vectors)                                                  | No — no protocol conformance duty                                                 | N/A                                                                                                                                                 | go-sse/sseparse/corpus.go:10-18                                           |
| 9  | `ssetest` (submodule v0.5.0): httptest-based collectors + assertions for SSE handlers                              | No — no handlers to exercise; gates test through real `Run` + captured stdio      | N/A                                                                                                                                                 | go-sse/ssetest/collect.go:25, :129; go-sse/ssetest/compat.go:145          |
| 10 | Line-oriented stateful scanning of a text corpus (the parsing concept underneath)                                  | Yes in concept — the gates scan Markdown line by line                             | Missed Opportunity (solved with stdlib) — whole-file reads + `strings.Split` + small scanner structs; no incremental `io.Reader` streaming anywhere | consumer internal/docsrows/rows.go:211; internal/docsrows/scan.go:150-166 |

Dependency note: the root module requires `templ`, `go-branded-id`, and
`go-error-family` (go-sse/go.mod:5-9) — root-module adoption in docs-gates
fails scripts/stack-drift-check.sh:8-9 on sight. `sseparse` alone is
zero-dependency (go-sse/sseparse/go.mod), but it parses a protocol
docs-gates never speaks.

## What docs-gates does today

Zero usage, zero presence. Verified:

```console
$ rg -n "github.com/larsartmann/go-sse" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**'
# rc=1 — zero module-path hits (the repo's prose mentions of "go-sse" are the
# 2026-11-21 audit-all drift incident narrative in docs/drafts/ and CHANGELOG,
# where go-sse was the *measured corpus*, not an import)
$ rg -n "go-sse|sseparse|ssetest" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 — not in any module graph
```

The streaming/event-parse scan the brief asks for, run and cleared:

```console
$ rg -n 'text/event-stream|event-stream|"sse"' /home/lars/projects/docs-gates --type go
# rc=1 — no SSE wire vocabulary anywhere
$ rg -n 'Scanner' /home/lars/projects/docs-gates --type go
# only internal/docsrows/scan.go — tableScanner, the Markdown table state machine
```

What that `tableScanner` actually is (the honest adjacency judgment): a
line-driven state machine over a whole file — fence tracking, pipe-row
accumulation, block flush at boundaries (internal/docsrows/scan.go:150-166) —
fed by `strings.Split(string(data), "\n")` from a full `os.ReadFile`
(internal/docsrows/rows.go:211). Mechanically it resembles a line parser the
way any FSM does; substantively it shares nothing with SSE: the grammar is
Markdown pipes/fences rather than `field: value` event fields, the input is
bounded local files rather than an unbounded network stream, and there is no
incremental `io.Reader` consumption anywhere in the consumer (even stdin
pre-flight candidates are read whole via `io.ReadAll`,
internal/docsfreshness/corpus.go:26). No connection object, no flusher, no
heartbeat, no replay buffer, no event serializer exists to cite as
duplicated — the duplication answer is none.

One reverse-direction fact worth recording: go-sse is a _docs-gates
consumer_ — its repo runs the battery over its own living docs, and the
2026-11-22 corpus-scope fix in docs-gates (CHANGELOG: the 122 → 123 → 120
prose-drift class go-sse measured) came from exactly that dogfooding. The
relationship is fleet-tooling, not library dependency.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** go-sse exists because "every Go project
that serves SSE reinvents the same four pieces" (go-sse/README.md:11) —
event serialization, connection lifecycle, fan-out, replay. Those four pieces
are real shared debt for the fleet's _live_ services (go-datastar, dashboards,
activity feeds). docs-gates serves nothing: it is a judge-and-exit battery
whose inter-process contract is one JSON document consumed by a local exec.
There is no fifth piece for it to re-implement; insulating here avoids
carrying a transport the binary can never open. The honest duplication
question ("name the file:line of hand-rolled streaming") has an empty answer,
verified above.

**Root-contract bar.** Root adoption is doubly barred (require block,
go-sse/go.mod:5-9; `net/http`-anchored Stream/Broadcaster types in a
network-free CLI). Even `sseparse` — zero-dependency — would be a require
line against a protocol never spoken.

**pkg/provider check.** No target: Detect's lifetime is one exec + one
strict decode (pkg/provider/provider.go:92-102, :148-171). If BuildFlow ever
wanted _live gate telemetry_ (a gate starting/finishing as it happens), the
shape would be a new long-running surface in a new module — and then go-sse
would be the right transport there, not here.

RECONSIDER-WHEN: docs-gates grows a watching/streaming mode — e.g. a
`docs-gates watch` daemon streaming gate verdicts to dashboards, or a fleet
activity feed of battery runs. That surface belongs in a separate module
(beside pkg/provider) and should build on go-sse + sseparse from day one
rather than hand-rolling an event wire. Until a live consumer exists, the
insulation is correct, and the current relationship (go-sse dogfoods the
battery) is the direction that actually pays.

## Ranked opportunities

None.

## Release shape (context for any future adoption)

Root v0.7.0 (2026-11-04) with `sseparse` v0.2.1 and `ssetest` v0.5.0 as
nested-module tags; `[Unreleased]` at HEAD moves the DataStar activity-feed
example to the companion go-datastar repo with the stated scope rationale
that go-sse "stays a protocol-free transport" (go-sse/CHANGELOG.md:26-37) —
the same boundary discipline that makes it the right _transport-only_ pick
for a future live surface. The root module's two small dependencies
(`go-branded-id` for typed event IDs, `go-error-family`) are the only weight
an adopter carries (go-sse/README.md:7); `sseparse` alone is zero-dependency
if only wire parsing is ever needed.

## How to verify

```console
$ git -C /home/lars/projects/go-sse log -1 --format='%H'
d010c2222cae90e01e3f8859992611204e0fecd4
$ git -C /home/lars/projects/go-sse describe --tags
v0.7.0-28-gd010c22
$ git -C /home/lars/projects/go-sse tag --list 'sseparse/*' | sort -V | tail -1; git -C /home/lars/projects/go-sse tag --list 'ssetest/*' | sort -V | tail -1
sseparse/v0.2.1
ssetest/v0.5.0

$ rg -n "github.com/larsartmann/go-sse" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**'
# rc=1 — zero module-path hits
$ rg -n "go-sse|sseparse|ssetest" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 — not in any module graph
$ rg -n 'text/event-stream|event-stream|"sse"' /home/lars/projects/docs-gates --type go
# rc=1 — no SSE vocabulary

# the nearest neighbor, confirmed non-SSE (whole-file Markdown FSM):
$ rg -n "type tableScanner|strings.Split" /home/lars/projects/docs-gates/internal/docsrows/scan.go /home/lars/projects/docs-gates/internal/docsrows/rows.go
/home/lars/projects/docs-gates/internal/docsrows/rows.go:211:	lines := strings.Split(string(data), "\n")
/home/lars/projects/docs-gates/internal/docsrows/scan.go:73:	parts := strings.Split(strings.Trim(masked, "|"), "|")
/home/lars/projects/docs-gates/internal/docsrows/scan.go:155:type tableScanner struct {
$ rg -n "io.ReadAll\(os.Stdin\)" /home/lars/projects/docs-gates/internal/docsfreshness/corpus.go
26:	data, err := io.ReadAll(os.Stdin)
```
