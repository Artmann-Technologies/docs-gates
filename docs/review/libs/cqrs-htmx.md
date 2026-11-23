# cqrs-htmx utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: root v4.19.0 (latest local tag; tags v4.18.0, v4.18.1, v4.19.0 present; fetched 2026-11-23) + HEAD 4b9803326337945cce87f68a040bfa8bae9d7474 (verified as checkout HEAD). Main modules: usermgmt v4.21.0 (+oauth2 v4.13.1, totp v4.12.1, webauthn v4.13.0), adminui v4.18.0, auditlog v4.13.0, dashboardui v4.18.0, datastar v4.13.0, health v4.15.0, huma v4.1.0, identity-model v4.18.0, loginpage v4.19.0, setup v4.18.0, systemadapter v4.14.0. Root go.mod's module line is `github.com/larsartmann/cqrs-htmx/v4` (cqrs-htmx/go.mod:2) — the manifest's "module line is identity-model" refers to the lint-ignore comment above it (cqrs-htmx/go.mod:1). SSE-only since ADR 0046 dropped WebSocket.
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f baseline — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 5 (UI + CQRS/application frameworks) subagent of the multi-library utilization audit fleet.

## Verdict

NOT-APPLICABLE — cqrs-htmx wires CQRS commands/queries to `net/http` + HTMX (decoding, Casbin authorization, validation, partial rendering, SSE fan-out, plus event-sourced identity and ready-made UIs), and docs-gates has no HTTP surface, no users, no sessions, and no HTML: zero overlap for the fleet argument to act on.

## Score

- Inputs N = 11 (inventory rows below).
- A = 1 (the readiness/point-in-time-verdict concept row — row 11).
- M = 0.
- Score = round(100 × 0/1) = **0** — band **low**. Expected for NOT-APPLICABLE; like the group-2 httputil report, the verdict carries the meaning and the number measures correctly-unrealized leverage.

- Applicable rows are exactly the inventory rows graded **Missed Opportunity** (the consumer has the underlying need; the capability is hand-rolled, solved in-domain, or declined by contract). Every **Not Applicable** row carries a one-line reason. `M` counts rows graded Fully Leveraged.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | Core `app.Command`/`app.Query` HandlerOption wiring: decode → authorize → validate → respond, one declarative line per endpoint | No — no endpoints; verbs dispatch from a plain map (cmd/docs-gates/verbs.go:77) | N/A | cqrs-htmx/README.md:38 |
| 2 | HTMX response building (HX-Trigger/Push-URL, partial-vs-full rendering), CSRF, method override | No — no browser, no responses to enhance | N/A | cqrs-htmx/README.md:72-73 |
| 3 | SSE-only realtime: broadcaster, reconnect, offline sync protocol (ADR 0046 dropped WebSocket) | No — nothing streams; output is one JSON document + stderr prose | N/A | cqrs-htmx/docs/adr/0046-drop-websocket-sse-only.md:1, :16 |
| 4 | `usermgmt` — event-sourced identity (register/login/sessions/roles/tenants/bots/lockout/audit log) + webauthn/totp/oauth2 satellites | No — no users; the battery's actor is CI | N/A | cqrs-htmx/README.md:95-96, cqrs-htmx/usermgmt/README.md:3-4 |
| 5 | `adminui` — one-call admin dashboard (users, tenants, members, audit log; embedded Tailwind build) | No — the verdict matrix IS the admin surface, 40 lines of text | N/A | cqrs-htmx/adminui/README.md:3, :8 |
| 6 | `dashboardui` — event-store observability (journal, projections, dead letters, time-travel) | No — no event store to observe | N/A | cqrs-htmx/README.md:17-19 |
| 7 | `setup` — one-call full stack (root app + usermgmt + adminui + dashboardui + loginpage) | No — the consumer's "one-call" is `docs-gates run .` | N/A | cqrs-htmx/README.md:139, :312-314 |
| 8 | `systemadapter` — identity domain bridged onto go-cqrs-lite `system.New` + metaengine presets | No — no system to adapt to | N/A | cqrs-htmx/systemadapter/README.md:3-8 |
| 9 | `huma`, `health`, `auditlog`, `datastar`, `loginpage`, `identity-model`, `eventcatalog`/OpenAPI collectors | No — no API spec, no probes, no login page, no identity model | N/A | cqrs-htmx/README.md:55 (module map) |
| 10 | Error-to-status mapping with HTMX-aware auth redirects | No — the 0/1/2 exit-code contract is the whole error surface (internal/gatecli constants, AGENTS.md:38-39) | N/A | cqrs-htmx/README.md:71 |
| 11 | Concept: readiness GATING on projection catch-up (block traffic until read models are current) | **Concept applies** — docs-gates' verdicts are point-in-time snapshots under foreign daemons (timestamp the run; one run is not steady state); judged below as rhyme, not duplication | Missed Opportunity (declined by doctrine: rhyme, not duplication — see gap analysis) | cqrs-htmx/projection_readiness.go:11-16; AGENTS.md:188-189 |

Dependency note: the root module's require graph (go-cqrs-lite v4 family,
Casbin, templ, go-appkit satellites — cqrs-htmx/go.mod) is an entire web stack;
any import into docs-gates' root breaks the zero-require contract outright.

## What docs-gates does today

Zero usage. Verified:

```console
$ rg -n "larsartmann/cqrs-htmx" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**'
# rc=1, no matches — not in go.mod (root has no requires), not in
# pkg/provider/go.mod:5-10 (go-finding, toolsdk, go-error-family only)
$ rg -n "html/template|text/template" /home/lars/projects/docs-gates --type go
# rc=1 — the consumer renders no HTML of any kind
```

The bare name appears in exactly two non-review places: the audit scaffolding
list (scripts/citation-sweep.py:50) and the docs-gates adoption survey, which
counts cqrs-htmx among the nine repos where the docs-health tooling rides
PATH/wrappers (docs/planning/2026-11-15_marker-grammar-and-tooling-split.md:77).
That is the INVERSE relationship again: cqrs-htmx is (per the survey) a
docs-gates consumer, not a dependency — the fleet argument's collective
improvements flow from this battery into cqrs-htmx's repo, not the other way.

Duplication hunt (fleet argument), full sweep:

- No HTTP handler/middleware/response code of any shape exists — the nearest
  thing to "response building" is the battery's verdict-matrix renderer and
  the snake_case JSON face (internal/battery/battery.go:128-130), a data
  contract, not a view.
- No auth/session/identity concepts — gates are unauthenticated by design and
  run in CI.
- No SSE/streaming — the one long-lived write is the watchdog's live warning
  line (internal/battery/battery_watchdog.go:25-48), a stderr write, not a
  stream protocol.
- No UI templating — see the templ-components report (same evidence, zero
  hits).

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** cqrs-htmx deletes the endpoint glue every
GOTH-stack web app rewrites (decode, authorize, validate, HTMX headers,
error-to-status — cqrs-htmx/README.md:66-75). docs-gates is not that app, and
its architecture is the negative image of it: no transport, no identity, no
rendering, no long-lived connections — a judge that answers in ~1.5 ms and
exits (README.md:69-72). There is no re-implementation to retire and no
file:line to name: no duplication exists because no overlap exists.

**The one honest rhyme, ruled.** cqrs-htmx's projection readiness gates
request traffic on read-model catch-up, because a projection still replaying
would serve stale answers as if current (cqrs-htmx/projection_readiness.go:11-16).
docs-gates has the same epistemic discipline expressed as doctrine instead of
machinery: gate verdicts are point-in-time snapshots under foreign auto-commit
daemons, so runs are timestamped and "one run is not steady state"
(AGENTS.md:188-189); freshness checks only the newest report, audit-all replays
the whole corpus under the ratchet (docs/DOMAIN_LANGUAGE.md:28-29). Same
insight (state moves under the reader), different layer: cqrs-htmx solves it
with runtime gating in a server; docs-gates solves it with corpus semantics
and ritual. Conceptual rhyme, not duplication — nothing to adopt, nothing to
retire.

**Root-contract bar.** Importing any cqrs-htmx module breaks
scripts/stack-drift-check.sh:8-13 on the first `go get` and drags an entire
web stack into a binary that serves nothing. pkg/provider is equally wrong a
home: its contract is one local exec plus a strict JSON decode
(pkg/provider/provider.go:148-171) — adding HTTP/UI wiring there would only
endanger the row-shape contract (pkg/provider/provider.go:92-102).

**What docs-gates would have to become.** A web application — e.g. a fleet
docs-health console (login, per-repo battery history, verdict timelines,
audit trail). If that ever exists, it is a textbook cqrs-htmx + templ-components
product (identity via usermgmt, rendering via templ-components, engine via
systemadapter onto go-cqrs-lite), built in its own repo that EXECS or
shell-outs to the stdlib battery — exactly the pkg/provider pattern, one layer
up. The battery stays a battery.

RECONSIDER-WHEN: a fleet-facing docs-gates console (auth + dashboards +
history over accumulated runs) is greenlit. Until then the insulation is
correct, and the existing relationship (cqrs-htmx consuming the battery per
the PATH/wrapper survey) already moves collective improvements in the
direction the owner wants.

## Ranked opportunities

None. (The point-in-time doctrine note above is already written into
AGENTS.md:188-189; there is no code opportunity.)

## How to verify

```console
$ git -C /home/lars/projects/cqrs-htmx log -1 --format='%H'
4b9803326337945cce87f68a040bfa8bae9d7474
$ git -C /home/lars/projects/cqrs-htmx tag --list 'v4.*' | sort -V | tail -3
v4.18.0 v4.18.1 v4.19.0
$ head -2 /home/lars/projects/cqrs-htmx/go.mod
//cqrs-lint:ignore(E003) identity-model is intentionally a cohesive pure-domain package (IDs, events, commands, fold); splitting would break usermgmt's alias/re-export architecture
module github.com/larsartmann/cqrs-htmx/v4

$ rg -n "larsartmann/cqrs-htmx" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**'
# rc=1 — zero module-path hits
$ rg -n "cqrs-htmx" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**'
docs/planning/2026-11-15_marker-grammar-and-tooling-split.md:77:| PATH / wrapper                   | 9           | cqrs-htmx, go-taskqueue, GoReleaser-Wizard, monitor365-ui-sdk, nixos-modules, ratsgedaechtnis, storbi, timesheets, touchbase                               |
scripts/citation-sweep.py:50:    "cqrs-htmx",
# the adoption survey + the audit scaffolding list; zero imports

# the rhyme, library side (readiness gating on projection catch-up):
$ sed -n '11,16p' /home/lars/projects/cqrs-htmx/projection_readiness.go
// Projection worker states as reported by [ProjectionStatusEntry.Status].
# ... "the string form is the published wire vocabulary (ProjectionStatusHandler
#     JSON, health bridges, readiness gates)"
# the rhyme, consumer side (point-in-time doctrine, no machinery):
$ sed -n '188,189p' /home/lars/projects/docs-gates/AGENTS.md
# "Gate verdicts are point-in-time under foreign daemons — timestamp the
#  run; one run is not steady state."
$ cat /home/lars/projects/docs-gates/pkg/provider/go.mod | sed -n '5,10p'
# require ( go-finding v1.15.1 + toolsdk v1.16.0 ) + go-error-family v0.12.0
# indirect — no cqrs-htmx anywhere in the provider's graph
```
