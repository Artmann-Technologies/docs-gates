# go-datastar utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v0.7.0 (latest local tag, fetched 2026-11-23) + HEAD 4fca56a9fbf1a0cfdfc5b7a988b3d00099ffe543 (`git describe` shows `static/v0.8.0-10-g4fca56a`; the static submodule's v0.8.0 tag is the newest of the family, root/broadcast/datastartest are v0.7.0). Modules: root `github.com/larsartmann/go-datastar` (requires go-sse v0.7.0 + go-error-family v0.12.0 + static v0.8.0), `broadcast` v0.7.0, `datastartest` v0.7.0, `static` v0.8.0.
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f baseline — root module stdlib-only (scripts/stack-drift-check.sh:9) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 3 (I/O + runtime + output) subagent of the multi-library utilization audit fleet.

## Verdict

NOT-APPLICABLE — go-datastar is the server side of a hypermedia protocol (SSE-pushed DOM patches, signals, and scripts to a browser client), and docs-gates has no browser, no HTTP surface, and no interactive state: it is a terminal-and-exit-code CLI whose two output faces are pinned text and one JSON document.

## Score

- Inputs N = 8 (inventory rows below).
- A = 0 (no row addresses a problem docs-gates has; the nearest adjacency — streaming progress to a live terminal — is solved in-repo with `gatecli.Livef` writing to the real stdout during a capture, no protocol involved).
- M = 0.
- Score = **0** (0/0; by convention 0 when nothing is applicable) — band **low**. For a NOT-APPLICABLE library the score is noise; the verdict is the signal.
- Roadmap check, for honesty: TODO_LIST/ROADMAP carry no dashboard/web/live-view row — the docs domain's "UI" is the markdown files themselves, judged by this battery, viewed in any editor.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | Patch model — every DOM patch a first-class value (`Patch` → `sse.Event`), buildable without a live connection | No — nothing in the consumer renders DOM or speaks the DataStar wire protocol | N/A | go-datastar/patch.go:25 |
| 2 | `Response` — element/signal/script patch writers over an `sse.Stream` (`PatchElements`, `PatchSignals`, `ExecuteScript`, `Redirect`, …) | No — requires an HTTP response stream; the consumer has zero `net/http` imports | N/A | go-datastar/response.go:40-110 |

  (Row 2's full method set — `PatchElementsTempl`, `RemoveElement(ByID)`, `ConsoleLog/Error`, `DispatchCustomEvent`, `ReplaceURL(Querystring)`, `Prefetch` — extends through go-datastar/response.go:152; every one writes to a browser connection.)
| 3 | `ReadSignals`/`LastEventID` — inbound client-state parsing (query/body signals, reconnect IDs) | No — the CLI's inbound surface is argv, a JSON config file, and one stdin pre-flight mode | N/A | go-datastar/inbound.go:26, :107 |
| 4 | `broadcast` submodule — fan-out to N connections, `WithReplayCapacity` reconnection replay, `Hub` sharing, heartbeats | No — the battery has exactly one consumer per run (the terminal or the provider reading stdout); there is no fan-out topology | N/A | go-datastar/broadcast/broadcaster.go:57; go-datastar/broadcast/options.go:63 |
| 5 | `ScriptHandler`/`ScriptTag` — serve the embedded DataStar JS client with ETag/Cache-Control | No — the binary serves nothing | N/A | go-datastar/script_handler.go:37, :88 |
| 6 | `static` module — the versioned, embedded JS client bundle as its own asset module | No — no web assets exist in the consumer | N/A | go-datastar/static/static.go |
| 7 | `datastartest` module — SSE parsing, typed decoding, and assertion helpers for E2E-testing DataStar handlers | No — nothing to test end-to-end over SSE; gates are tested through their real `Run` functions with captured stdout | N/A | go-datastar/datastartest/assert.go |
| 8 | Templ/GoStar rendering adapters for patch content | No — the only "renderer" the docs domain knows is the markdown viewer a human points at docs/status/*.md, produced by the status-report skill, not this binary | N/A | go-datastar/adapters.go |

### What the library is, in one paragraph

The Go server-side vocabulary of the DataStar hypermedia protocol: DOM patches, signal updates, and script executions pushed to the browser as Server-Sent Events over go-sse. Its differentiating bet is that "patches are values, not method calls" — every patch implements `Patch` and can be built offline from any connection (go-datastar/patch.go:25), then routed through a broadcaster with reconnection replay (go-datastar/broadcast/broadcaster.go:57, ring-buffer capacity at go-datastar/broadcast/options.go:63), persisted, or filtered per subscriber. The [Unreleased] section is an activity-feed showcase exercising exactly that pipeline (go-datastar/CHANGELOG.md:8-19). Everything it offers presumes an HTTP response stream and a browser on the other end.

### The error taxonomy, examined for fleet value

go-datastar classifies every error three ways — stable code (`datastar.<operation>_<failure>`), errorfamily family (Rejection/Transient/Orchestration, chosen by whose fault and whether retrying helps), and sentinel (`errors.Is`) — documented at go-datastar/errors.go:10-33. This is genuinely good fleet engineering, and the go-error-family group's report covers where the consumer stands on that axis: docs-gates maps every failure onto three exit codes (internal/gatecli/gatecli.go:27-29) with typed refusal carriers where a caller must distinguish. For a wire protocol with remote clients, codes/families pay for themselves; for a CLI whose callers branch on one integer, they would be dead weight. Nothing here changes the verdict; it is noted so the fleet knows the library's best ideas were looked at, not skipped.

### datastartest, the E2E half

`Collect`-family helpers drive a real `http.Handler` and gather SSE events for assertion (go-datastar/datastartest/collect.go:26-121). The consumer's closest analogue is `gatecli.CaptureOutput` — swapping a process's stdout/stderr to exercise a gate's REAL `Run` end to end (AGENTS.md:191-195). Same philosophy (test the real surface, never a mock), different transport (function-in-process vs HTTP handler). No shared debt.

### The inbound half, also absent

Row 3 covers `ReadSignals` parsing client state from query or body (go-datastar/inbound.go:26-107) and `LastEventID` reconnect headers (go-datastar/inbound.go:107). docs-gates' entire inbound surface: argv (parsed by gatecli.Parse), one JSON config file (`internal/config`, decode-over-defaults with unknown-key WARN), and stdin read exactly once for `--pre-flight` candidate lines. There is no request, no session, no reconnect state — the concepts do not transpose even partially.

### Contrast table: what a DataStar consumer is vs what docs-gates is

| Axis | DataStar consumer | docs-gates |
|---|---|---|
| Process shape | long-lived HTTP server | run-to-completion CLI |
| Delivery | SSE stream, push | stdout/stderr bytes, then exit code |
| State | signals, live DOM, subscribers | none between runs (one gitignored stamp) |
| Client | browser JS runtime | terminal human / provider parser |
| Change propagation | patches on event | re-exec on trigger (CI, BuildFlow, human) |

### Release shape and the official-SDK comparison (context, not a case for adoption)

The library cuts its four modules in lockstep — v0.7.0 was a deliberate "full four-module lockstep cut" repairing a root-only v0.6.3 release defect (go-datastar/CHANGELOG.md:27-31). Against the official `starfederation/datastar-go` SDK it differentiates on broadcast/replay/filtering plus error-family classification, and concedes built-in compression and first-party cadence (its own README comparison). None of those axes exist for a consumer with no HTTP handler: the comparison is included only to show the library's entire design center — connection lifecycles, subscriber fan-out, reconnect state — is orthogonal to a terminal verb dispatcher.

## What docs-gates does today

Zero usage, zero presence in either module graph. Verified:

```console
$ rg -n "datastar" /home/lars/projects/docs-gates -g '!docs/review/**' --type go
# rc=1, no matches (case-insensitive too)
$ rg -in "datastar" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1, no matches
```

No hand-rolled hypermedia-equivalent exists either — the absence is structural:

- **No HTTP anywhere**: `rg -n "net/http|text/html|html/template|EventSource" internal/ cmd/ pkg/ --type go` → rc=1 for non-test code. The provider's IPC is `exec.CommandContext` of the docs-gates binary (pkg/provider/provider.go:149); everything else is function calls within the process.
- **No interactive state to sync**: "signals" in this repo's vocabulary would collide with the docs domain's `report_dir` tokens; the closest thing to live state is the half-budget watchdog line, and its transport is stderr via `gatecli.LiveErrf` — one fire-and-forget write, no session.
- **Output faces are exactly two, both pinned**: the human verdict matrix (internal/battery/battery.go:317) and the `--json` machine document (internal/battery/battery.go:264-295), byte-pinned by parity goldens (internal/battery/parity_test.go:188). Neither has a browser consumer; the provider is the sole machine consumer and it wants one JSON document on stdout, the opposite of an event stream.

### The consumer's entire user-facing surface

Worth enumerating to prove no interactive channel hides anywhere: fourteen verbs dispatched from a fixed table (cmd/docs-gates/verbs.go:28-44), each rendering verdict/finding lines to stdout/stderr through the `gatecli` write family and exiting 0/1/2 (internal/gatecli/gatecli.go:27-29); one JSON face; stdin consumed exactly once (`--pre-flight` candidate lines); four guarded file writes (the go-atomic-write report's four sites). No port, no socket, no long-lived connection, no push — the complete inventory of "delivery" this binary knows is: process exit and file bytes. A hypermedia protocol has nothing to attach to.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** Both prongs score zero. docs-gates re-implements nothing go-datastar offers — there is no HTML fragment, no SSE framing, no signals store, no reconnection semantics anywhere in the tree (that absence is what the group's zero-hit sweeps above prove). And the problems go-datastar solves (live DOM patching, broadcast fan-out, replay on reconnect) are problems only a web-served product has; docs-gates' consumers are a terminal, a CI runner, and one BuildFlow plugin. The fleet argument's insulation half works exactly as intended here: the library solves a problem this consumer deliberately does not have — "no net/http, no network" is stated architecture, and it keeps the nix build hermetic and the goldens deterministic.

**Root-contract bar.** Barred on arrival: the root module tolerates zero requires, and go-datastar transitively rides go-sse + go-error-family. The pkg/provider escape hatch is equally wrong a home: it is a read-only BuildFlow tool registration, not a web service.

**The honest adjacent thought, declined.** A "docs-health dashboard" that live-updates as reports land would be a legitimate go-datastar product — but it belongs to a separate fleet service that EXECS `docs-gates run --json` (the exact pattern pkg/provider already demonstrates), never to the gate binary itself. Nothing in the consumer's ROADMAP or TODO_LIST points at a dashboard.

The four modules reviewed, for completeness:

- Root `github.com/larsartmann/go-datastar` v0.7.0 — the protocol vocabulary (patches, Response, inbound, errors, adapters).
- `broadcast` v0.7.0 — fan-out, replay, hubs (go-datastar/broadcast/broadcaster.go:57).
- `datastartest` v0.7.0 — handler-driving E2E collectors (go-datastar/datastartest/collect.go:26).
- `static` v0.8.0 — the embedded JS client bundle as its own asset module (go-datastar/static/static.go).

Why NOT-APPLICABLE and not NOT-ADOPTED, in three lines:

1. No capability of the library maps to any surface of the consumer — the sweeps prove absence, not decline.
2. The consumer's delivery model (bytes then exit) cannot host a connection-oriented protocol without becoming a different program.
3. The fleet-level integration point already exists and is library-free: exec `run --json` from whatever web project wants the data.

RECONSIDER-WHEN: the fleet ever builds a live docs-health dashboard (web page showing gate verdicts/report freshness as files change). That project should adopt go-datastar (plus go-filewatcher for the change source) from day one and shell out to this binary for verdicts — the consumer itself stays terminal-only.

**Fleet-argument scorecard, per prong.** (a) "Re-implementation debt": zero — no HTML, no SSE framing, no signals store, no reconnect state anywhere in the tree (the group's zero-hit sweeps are the proof). (b) "Collective improvement": nothing to inherit under NOT-APPLICABLE — protocol/vocabulary gains in go-datastar (new patch kinds, replay semantics) benefit web surfaces this consumer will never have while it remains a CLI. The correct fleet-level move if a dashboard is ever wanted is adopting go-datastar in the dashboard repo and exec-ing `docs-gates run --json` there — the provider module already demonstrates that this binary integrates as a SUBPROCESS, which is the whole architectural answer.

## Ranked opportunities

None.

### Cross-references within this audit fleet

- go-sse (group 1) underlies this library's transport; docs-gates appears in that report's zero-hit list too.
- go-filewatcher (this group) is the natural change-source half of any future dashboard; both reports converge on the same flip condition.

## How to verify

```console
$ git -C /home/lars/projects/go-datastar log -1 --format='%H'
4fca56a9fbf1a0cfdfc5b7a988b3d00099ffe543
$ git -C /home/lars/projects/go-datastar describe --tags
static/v0.8.0-10-g4fca56a
$ git -C /home/lars/projects/go-datastar tag --list | tail -3
v0.6.2
v0.6.3
v0.7.0
$ head -10 /home/lars/projects/go-datastar/go.mod
# module github.com/larsartmann/go-datastar — requires static v0.8.0,
# go-error-family v0.12.0, go-sse v0.7.0

$ rg -n "datastar" /home/lars/projects/docs-gates -g '!docs/review/**' --type go
# rc=1 (zero Go-file hits)
$ rg -in "datastar" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 (not in any module graph)

$ rg -n "net/http|text/html|html/template|EventSource" /home/lars/projects/docs-gates/internal /home/lars/projects/docs-gates/cmd /home/lars/projects/docs-gates/pkg --type go | rg -v _test
# rc=1 — no HTTP, no HTML, no SSE in non-test code
$ rg -n "exec.CommandContext" /home/lars/projects/docs-gates --type go | rg -v _test
# internal/docsannotate/git.go:25, internal/docsarchive/docsarchive.go:203,
# internal/docsfreshness/preflight_land.go:74, pkg/provider/provider.go:149 —
# the process graph is four one-shot exec sites, all git or self
$ rg -in "dashboard|web view|browser" /home/lars/projects/docs-gates/ROADMAP.md /home/lars/projects/docs-gates/TODO_LIST.md
# rc=1 — no web-surface feature planned; the docs ARE the UI
$ git -C /home/lars/projects/go-datastar log -1 --format='%ci %h'
# fetched 2026-11-23; 0 behind upstream per the fleet manifest
```
