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

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | Patch model — every DOM patch a first-class value (`Patch` → `sse.Event`), buildable without a live connection | No — nothing in the consumer renders DOM or speaks the DataStar wire protocol | N/A | go-datastar/patch.go:25 |
| 2 | `Response` — element/signal/script patch writers over an `sse.Stream` (`PatchElements`, `PatchSignals`, `ExecuteScript`, `Redirect`, …) | No — requires an HTTP response stream; the consumer has zero `net/http` imports | N/A | go-datastar/response.go:40-110 |
| 3 | `ReadSignals`/`LastEventID` — inbound client-state parsing (query/body signals, reconnect IDs) | No — the CLI's inbound surface is argv, a JSON config file, and one stdin pre-flight mode | N/A | go-datastar/inbound.go:26, :107 |
| 4 | `broadcast` submodule — fan-out to N connections, `WithReplayCapacity` reconnection replay, `Hub` sharing, heartbeats | No — the battery has exactly one consumer per run (the terminal or the provider reading stdout); there is no fan-out topology | N/A | go-datastar/broadcast/broadcaster.go:57; go-datastar/broadcast/options.go:63 |
| 5 | `ScriptHandler`/`ScriptTag` — serve the embedded DataStar JS client with ETag/Cache-Control | No — the binary serves nothing | N/A | go-datastar/script_handler.go:37, :88 |
| 6 | `static` module — the versioned, embedded JS client bundle as its own asset module | No — no web assets exist in the consumer | N/A | go-datastar/static/static.go |
| 7 | `datastartest` module — SSE parsing, typed decoding, and assertion helpers for E2E-testing DataStar handlers | No — nothing to test end-to-end over SSE; gates are tested through their real `Run` functions with captured stdout | N/A | go-datastar/datastartest/assert.go |
| 8 | Templ/GoStar rendering adapters for patch content | No — the only "renderer" the docs domain knows is the markdown viewer a human points at docs/status/*.md, produced by the status-report skill, not this binary | N/A | go-datastar/adapters.go |

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

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** Both prongs score zero. docs-gates re-implements nothing go-datastar offers — there is no HTML fragment, no SSE framing, no signals store, no reconnection semantics anywhere in the tree (that absence is what the group's zero-hit sweeps above prove). And the problems go-datastar solves (live DOM patching, broadcast fan-out, replay on reconnect) are problems only a web-served product has; docs-gates' consumers are a terminal, a CI runner, and one BuildFlow plugin. The fleet argument's insulation half works exactly as intended here: the library solves a problem this consumer deliberately does not have — "no net/http, no network" is stated architecture, and it keeps the nix build hermetic and the goldens deterministic.

**Root-contract bar.** Barred on arrival: the root module tolerates zero requires, and go-datastar transitively rides go-sse + go-error-family. The pkg/provider escape hatch is equally wrong a home: it is a read-only BuildFlow tool registration, not a web service.

**The honest adjacent thought, declined.** A "docs-health dashboard" that live-updates as reports land would be a legitimate go-datastar product — but it belongs to a separate fleet service that EXECS `docs-gates run --json` (the exact pattern pkg/provider already demonstrates), never to the gate binary itself. Nothing in the consumer's ROADMAP or TODO_LIST points at a dashboard.

RECONSIDER-WHEN: the fleet ever builds a live docs-health dashboard (web page showing gate verdicts/report freshness as files change). That project should adopt go-datastar (plus go-filewatcher for the change source) from day one and shell out to this binary for verdicts — the consumer itself stays terminal-only.

## Ranked opportunities

None.

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
```
