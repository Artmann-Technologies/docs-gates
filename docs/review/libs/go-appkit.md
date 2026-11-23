# go-appkit utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: root v0.13.0 (latest local tag; `git describe` from the manifest HEAD → `cqrs/v0.10.0-7-g74e4d80`, i.e. satellites tagged independently) + HEAD 74e4d8011c6c275bb9e3115384ab951ed9d66d0a per the fleet manifest (the checkout tip carries later auto-commits past the manifest HEAD — a moving target under the fleet's auto-commit daemon — none touching module code I cite). 15 modules: root core, config v0.1.1, cqrs v0.10.0, docs v0.3.4, errorpages v0.1.3, flightrecorder v0.2.0, flightrecorderhealth v0.1.7, health v0.3.2, httpx v0.2.0, otel v0.4.0, preset v0.1.1, realtime v0.2.0, security v0.3.0, systemd v0.2.0, integration.
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f baseline — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 5 (UI + CQRS/application frameworks) subagent of the multi-library utilization audit fleet.

## Verdict

NOT-ADOPTED — go-appkit is the fleet's HTTP service framework ("A NEW fleet service starts on go-appkit — owner ruling", go-appkit/README.md:7), and docs-gates is not a service: the one capability family that genuinely maps onto a docs-gates problem (struct-tag config loading, the `config` module) is deliberately re-solved in ~300 stdlib lines because the root contract forbids its dependency chain — justified insulation, not neglect.

## Score

- Inputs N = 12 (inventory rows below).
- A = 1 (the config-loading row — the only problem docs-gates actually has).
- M = 0.
- Score = round(100 × 0/1) = **0** — band **low**. The verdict and gap analysis carry the substance; note the relationship between these two repos is already strong in the *inverse* direction (below).

- Applicable rows are exactly the inventory rows graded **Missed Opportunity** (the consumer has the underlying need; the capability is hand-rolled, solved in-domain, or declined by contract). Every **Not Applicable** row carries a one-line reason. `M` counts rows graded Fully Leveraged.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | Core: `appkit.Service` — HTTP server lifecycle, middleware stack, graceful drain/shutdown, health endpoints | No — no server, no listener, no long-running process; the binary runs gates and exits | N/A | go-appkit/README.md:36-56 |
| 2 | `preset` — one-call composition (core + systemd + health dashboard + errorpages, env-switched otel/security/realtime legs) | No — the battery composes gates from config, not services; see gap analysis | N/A | go-appkit/preset/README.md:3-5 |
| 3 | `cqrs` — event service over go-cqrs-lite `system` (store, projections, DLQ, typed C/Q facade) | No — no events, no store; judged fully in the go-cqrs-lite report | N/A | go-appkit/cqrs/README.md:1-5 |
| 4 | `config` — env-first struct-tag loading, precedence flag > env > file > default, strict YAML/JSON layer | **Yes in problem shape** — docs-gates loads `.docs-gates.json` onto typed settings; deliberately re-solved stdlib-only | Missed Opportunity (declined by contract, see gap analysis) | go-appkit/README.md:98; internal/config/config.go:3-14 |
| 5 | `docs` — auto-generated OpenAPI/AsyncAPI/D2 docs from Go types | No — docs-gates has no API surface to document | N/A | go-appkit/docs/README.md:3 |
| 6 | `errorpages` — pretty classified error pages (HTML) + JSON contracts | No — errors are exit codes 0/1/2 + stderr prose | N/A | go-appkit/README.md:89 |
| 7 | `realtime` — SSE hub (broadcast, replay, heartbeat, Last-Event-ID resume) | No — no connections, nothing streamed | N/A | go-appkit/README.md:90 |
| 8 | `flightrecorder` (+ health bridge) — on-demand runtime/trace capture on failure | No — wedged gates are abandoned at the wall-clock cap, not diagnosed in-process; capture is pipe-draining (internal/gatecli/gatecli_capture.go:121) | N/A | go-appkit/README.md:92-93 |
| 9 | `health` — probes, startup latch, real-time dashboard | No — health IS the exit code; the matrix is the dashboard | N/A | go-appkit/README.md:94 |
| 10 | `security` — CSRF, rate limiting, body limits, CSP nonce | No — no requests, no origins | N/A | go-appkit/README.md:95 |
| 11 | `systemd` — sd_notify READY/STOPPING/WatchdogSec keepalive | No — not a daemon; CI owns its lifecycle | N/A | go-appkit/README.md:96 |
| 12 | `httpx` — handler DX (ResultHandler, bind+validation, content negotiation) + `integration`/testkit | No — no `http.Handler` exists in either consumer module | N/A | go-appkit/README.md:97 |

Root module composition note: adoption of ANY row into the docs-gates root
module breaks the zero-require contract on the first import — even the lightest
candidate (`config`) requires `go-faster/yaml` + `go-error-family` plus four
indirects (go-appkit/config/go.mod:5-13).

## What docs-gates does today

Zero code usage; an unusually rich *inverse* relationship. Verified:

```console
$ rg -n "larsartmann/go-appkit" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**'
# rc=1, no matches — not in go.mod (root carries no requires), not in
# pkg/provider/go.mod:5-10, not in any source file
```

The bare name appears only as adoption/ritual prose, because **go-appkit is a
docs-gates consumer, one of the pilot fleet rollouts**:

- Known-consumers list: crm, CV, go-appkit, webphone, pbx-artmann
  (README.md:120-126).
- Pilot ratchet row: go-appkit runs the battery with `audit_max` pinned 12/12
  exact — genuine open obligations, nothing to drain (TODO_LIST.md:30).
- The repo's own benchmark family carries a dedicated go-appkit corpus shape —
  `BenchmarkCheckReportReplayGoAppkitShape`, 12 reports × ~150 living lines
  (internal/docsfreshness/bench_test.go:145-152) — so a memoization regression
  would show up on the pilot's shape, not only the CV cliff.
- Pilots seeded with `.docs-gates.json` at measured debt, 5/5 PASS
  (CHANGELOG.md:750-752).

Hand-rolled overlap hunt (the fleet-argument question), in full:

- **Config**: the one honest candidate. `internal/config` decodes one optional
  JSON file over `Defaults()` — unknown keys WARN, malformed JSON exits 2
  (internal/config/config.go:3-14). appkit/config exists precisely to kill a
  fleet-wide hand-rolled pattern (22 `envOr` packages,
  go-appkit/config/README.md:8-9) — but that pattern (env-first, multi-layer
  precedence, YAML trees) is not docs-gates' problem: the JSON-not-YAML
  decision is recorded as the stdlib contract itself (jq-able, hermetic), and
  the whole loader incl. defaults, key registry, and typed errors is ~1100
  lines of golden-pinned stdlib (internal/config/config.go,
  config_defaults.go, known_keys.go, errors.go). Rhyme, not duplication.
- **DI/composition**: no container exists. The battery's gate table
  (internal/battery/battery.go:58-81) is a static list built from config —
  the closest thing to a composition root, and it is 24 lines of slice
  literal, deliberately argv-shaped for parity.
- **Service lifecycle/preset**: nothing long-lived to compose, drain, or shut
  down; `runGates` is a for-loop with a per-gate wall-clock cap
  (internal/battery/battery.go:174-200).

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** The owner's argument is strongest where
many projects share one problem class — and here it already works, just in the
other direction: collective docs-gates improvements (memoized replays,
ratchet semantics, the JSON face) flow INTO go-appkit on every pilot battery
run, and go-appkit's corpus shape is pinned inside docs-gates' own benchmark
suite (internal/docsfreshness/bench_test.go:145-152). The fleet gets shared
upgrades without docs-gates importing anything. For the remaining eleven
capability rows there is no shared problem: docs-gates has no HTTP surface,
no processes to keep alive, nothing to observe at runtime — the same verdict
as httputil (group 2), one framework layer up.

**The config question, judged honestly.** If the root contract did not exist,
appkit/config would still be a poor fit: its strict-file-key semantics differ
from docs-gates' WARN-on-unknown-key sprawl guard, its env/flag layers have no
callers here, and its YAML dependency is exactly what the recorded format
decision excludes (internal/config/config.go:3-14). The consumer's loader is
small, stable, byte-pinned by the defaults-are-historical invariant
(AGENTS.md:40-42), and already shared fleet-wide as the thing being consumed
— re-homing it on appkit/config would move code to a dep to solve a problem
docs-gates does not have.

**What docs-gates would have to become.** A long-running service — e.g. a
scheduled battery runner daemon, or a fleet aggregator serving verdict
history. That service should start from `preset` per the owner ruling
(go-appkit/README.md:7), live in its own module (the pkg/provider precedent),
and exec the stdlib binary exactly the way pkg/provider does
(pkg/provider/provider.go:148-171). Adopting go-appkit into the root module
would also be circular: the gate taking a dependency on a repo it gates.

RECONSIDER-WHEN: docs-gates grows a daemon or serving face (battery
scheduler, fleet verdict aggregator, webhook receiver). Then: build it as a
separate module on appkit `preset` from day one; until then, insulation is
correct — and the existing inverse integration already delivers the fleet
argument's benefit.

## Ranked opportunities

None. (The go-appkit corpus-shape benchmark and pilot ratchet already exist
inside docs-gates; keeping them green IS the fleet-leverage play here.)

## How to verify

```console
$ git -C /home/lars/projects/go-appkit rev-parse 74e4d8011c6c275bb9e3115384ab951ed9d66d0a
74e4d8011c6c275bb9e3115384ab951ed9d66d0a   # fleet-manifest HEAD exists on the branch (the tip is deliberately not pinned: it drifts under the fleet's auto-commit daemon)
$ git -C /home/lars/projects/go-appkit describe --tags 74e4d8011c6c275bb9e3115384ab951ed9d66d0a
cqrs/v0.10.0-7-g74e4d80

$ rg -n "larsartmann/go-appkit" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**'
# rc=1 — zero module-path hits in code or module graphs
$ rg -n "go-appkit" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**' -g '!docs/planning/**' -g '!docs/status/**' | cut -d: -f1 | sort | uniq -c
      4 /home/lars/projects/docs-gates/CHANGELOG.md
      1 /home/lars/projects/docs-gates/FEATURES.md
      1 /home/lars/projects/docs-gates/internal/docsfreshness/bench_test.go
      1 /home/lars/projects/docs-gates/README.md
      1 /home/lars/projects/docs-gates/ROADMAP.md
      1 /home/lars/projects/docs-gates/scripts/citation-sweep.py
      1 /home/lars/projects/docs-gates/TODO_LIST.md

$ sed -n '3,8p' /home/lars/projects/docs-gates/internal/config/config.go
//
// Format decision (2026-11-03): JSON, not YAML — the module is stdlib-only
// by contract (hermetic nix builds, vendorHash-free), and JSON is jq-able,
// which matters in a fleet whose tooling speaks jq. YAML support is a
// deliberate exclusion; revisit only if a consumer cannot seed JSON.
//
$ sed -n '5,9p' /home/lars/projects/go-appkit/config/README.md
plain struct — with classified [go-error-family] failures that name the exact
field path.
It exists because the LarsArtmann fleet kept hand-rolling the same
`envOr`/`envBool`/`envDurationOr` pattern (22 config packages and 25 ad-hoc
$ cat /home/lars/projects/go-appkit/config/go.mod
module github.com/larsartmann/go-appkit/config
go 1.27
require (
	github.com/go-faster/yaml v0.4.6
	github.com/larsartmann/go-error-family v0.12.0
)
require (
	github.com/go-faster/errors v0.6.1 // indirect
	github.com/go-faster/jx v1.0.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
)
```
