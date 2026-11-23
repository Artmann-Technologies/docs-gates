# go-health utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v0.10.0 (latest local tag, fetched 2026-11-23) + HEAD 3ca66446e1f8164f28e1ab788ad03727f6c99307. Modules: root `github.com/larsartmann/go-health`; nested `tools/doanalyzerv2` (untagged, private-analyzer runner).
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 4 (ops/health/auth/CLI-framework) subagent of the multi-library utilization audit fleet.
- Fleet stance addressed: "if all my projects re-implement everything, I will only fix bugs all day and none benefit from collective improvements" — does stdlib-only insulation here avoid real re-implementation debt, or leave docs-gates hand-rolling something the fleet solved?

## Verdict

NOT-APPLICABLE — go-health is a Kubernetes health-probe SDK (HTTP `/healthz`/`/readyz`/`/startupz` handlers over a samber/do container for long-running services), and docs-gates is a point-in-time exit-code CLI with no server, no container, and no process that outlives a run; the only conceptual neighbor (many independent checks → one verdict) is already owned by the battery under contracts go-health cannot serve.

## Score

- Inputs N = 13 (inventory rows below).
- A = 2 (the check-execution/aggregation concept row and the panic-recovery concept row — the two ideas that could in principle land in the battery).
- M = 0.
- Score = round(100 × 0/2) = **0** — band **low**. Expected for NOT-APPLICABLE; the number measures correctly-unrealized leverage, the verdict carries the meaning.

- Applicable rows are exactly the inventory rows graded **Missed Opportunity** (the consumer has the underlying need; the capability is hand-rolled, solved in-domain, or declined by contract). Every **Not Applicable** row carries a one-line reason. `M` counts rows graded Fully Leveraged.

## What the library offers

| #  | Capability                                                                                    | Applies to docs-gates?                                                                                                                          | Status                                                                                                     | Evidence                                                                                                                    |
| -- | --------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| 1  | Three-probe HTTP SDK: liveness/readiness/startup handlers with 200/503 semantics              | No — no HTTP server, no kubelet; health IS the 0/1/2 exit code                                                                                  | N/A                                                                                                        | go-health/README.md:197-217; go-health/handlers.go                                                                          |
| 2  | samber/do v2 injector integration (`health.New(injector, ...)`, eager critical invocation)    | No — no DI container anywhere in either module                                                                                                  | N/A                                                                                                        | go-health/probe.go:161; go-health/README.md:93-128                                                                          |
| 3  | Critical vs non-critical classification (fail→503 vs warn→200)                                | Concept only — the battery's FAIL/SKIP rows and the gates' WARN-before-enforce knobs solve the "degraded but green" question in exit-code space | Missed Opportunity (solved in-domain — exit-code space already carries the degraded-but-green distinction) | go-health/README.md:247-249; internal/battery/battery.go:272-279                                                            |
| 4  | Named concurrent checks with per-check duration, panic recovery, batch deadline (`NewChecks`) | Concept yes — the battery's closest analog; mechanism unusable (see gap analysis)                                                               | Missed Opportunity (declined: mechanism unusable — see gap analysis)                                       | go-health/README.md:444-451; go-health/checks.go:15                                                                         |
| 5  | Background result caching + refresh loop (1s default)                                         | No — gates are point-in-time by doctrine; a cached verdict would lie                                                                            | N/A                                                                                                        | go-health/README.md:307-315                                                                                                 |
| 6  | Shutdown awareness / two-phase drain (`MarkShuttingDown`, grace period)                       | No — the binary runs and exits; nothing to drain                                                                                                | N/A                                                                                                        | go-health/README.md:399-413                                                                                                 |
| 7  | Programmatic health API (`Status()`, `Ready()`, `AwaitReady`, `Healthz()`)                    | No — `Run([]string) int` return values are the whole status surface                                                                             | N/A                                                                                                        | go-health/README.md:415-434                                                                                                 |
| 8  | `checks` batteries: Disk, Memory, HTTP, Database probe functions                              | No — no disks, heaps, URLs, or DB handles to probe                                                                                              | N/A                                                                                                        | go-health/checks/checks.go:27; go-health/checks/checks.go:58; go-health/checks/checks.go:93; go-health/checks/checks.go:127 |
| 9  | `aggregate` sub-package: merge N in-process probes, worst-of status, namespaced keys          | No — one process, one battery; the matrix is not a merge of probes                                                                              | N/A                                                                                                        | go-health/aggregate/aggregate.go:74; go-health/README.md:482-503                                                            |
| 10 | `federation` sub-package: fetch remote health documents over HTTP, merge untrusted input      | No — the root module has no network by contract                                                                                                 | N/A                                                                                                        | go-health/federation/federation.go:147; go-health/README.md:505-551                                                         |
| 11 | Observability seam (`WithEvaluationHook`, Prometheus recipe)                                  | No — no polling consumers; findings ride the gate's own stdout                                                                                  | N/A                                                                                                        | go-health/README.md:453-468                                                                                                 |
| 12 | Build-version endpoint (`VersionHandler`, `WithVersion`)                                      | No — `--version` interception already answers this under legacy names                                                                           | N/A                                                                                                        | go-health/README.md:254; internal/gatecli/gatecli.go:109-117                                                                |
| 13 | `tools/doanalyzerv2`: private AST analyzer for samber/do anti-patterns                        | No — no samber/do usage to analyze                                                                                                              | N/A                                                                                                        | go-health/tools/doanalyzerv2/main.go:1-16                                                                                   |

Dependency note for any future adoption: go-health is not stdlib-only —
go-health/go.mod:5-9 requires `samber/do/v2`, `go-branded-id`, and
`samber/go-type-to-string`. Root-module adoption would fail
scripts/stack-drift-check.sh on the first `go get` and would pull the
`net/http` machinery the hermetic nix build deliberately excludes.

## What docs-gates does today

Zero usage, zero module-graph presence. Verified:

```console
$ rg -n "github.com/larsartmann/go-health" /home/lars/projects/docs-gates/go.mod \
    /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=2 for the root files (no go.sum exists at root — zero requires), no matches
$ rg -l "larsartmann/go-health" -g '*.go' /home/lars/projects/docs-gates
# rc=1, no matches
```

The only "go-health" strings in the repo are prose: planning docs and
archived status reports recording the 2026-11-05 dogfood run of the
docs-gates battery OVER the go-health repo (e.g.
docs/status/archived/2026-11-05_02-15_fleet-dogfood-go-health-cv-audit-all-perf-cliff.md:15)
— go-health is a _subject_ of the battery, not a dependency. No samber
imports anywhere: `rg -n "samber" go.mod pkg/provider/go.mod` → rc=1.

The hand-rolled-equivalent hunt (battery aggregation + provider health
hooks), with the exact code that owns each concern:

- **Battery aggregation** — internal/battery/battery.go:58-81 builds the
  ordered ten-gate list; internal/battery/battery.go:174-234 runs each gate
  INDEPENDENTLY under a per-gate wall-clock cap with captured stdio;
  internal/battery/battery.go:317-340 renders the PASS/FAIL/SKIP matrix;
  internal/battery/battery.go:264-313 renders the `--json` machine face. This
  is the functional neighbor of go-health's `NewChecks` + classification, but
  it is exit-code-and-stdio shaped, byte-pinned by
  internal/battery/parity_test.go:84's invocation table.
- **Provider health hook** — pkg/provider/provider.go:291-297 implements the
  toolsdk `Spec.HealthCheck` as `exec.LookPath("docs-gates")` plus an install
  hint. Judged honestly against go-health's model: this is an
  _installation_ check ("is the tool present"), not a _dependency health_
  check ("is a runtime service reachable"). go-health's
  check/aggregate/classify machinery answers a question the provider does
  not have — the provider's only dependency is a local binary, and
  `LookPath` is the complete correct answer. No overlap worth adopting.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** go-health exists to stop every fleet
_service_ from hand-rolling restart-cascade-prone health endpoints — a real
shared problem across the owner's HTTP projects (15 direct consumers,
go-health/README.md:677-684). docs-gates is not in that class: it is a
judge that runs, decides, and exits. Its "aggregation" (the battery) is not
a re-implementation of go-health's probe loop — the two share only the
abstract shape "N checks → one verdict". The battery's concrete
obligations are alien to go-health: gates communicate through byte-pinned
stdio captured by a **process-global stdout swap**
(internal/gatecli/gatecli_capture.go — AGENTS.md: "gatecli.CaptureOutput is
a process-global stdout swap"), which is why gates run sequentially; go-health's
`NewChecks` wins by running checks concurrently, and that win is
unimportable here without rewriting every gate's output path. The verdict
arithmetic (exit 1 iff ≥1 FAIL, SKIP never counts, rc=2 inside the battery
becomes a FAIL row) is pinned by parity goldens and the provider's strict
decoder. There is no file:line of duplicated probe logic to name — the
honest answer is that the overlap is conceptual, not implementational.

**Root-contract bar.** Outright: go-health's three requires
(go-health/go.mod:5-9) break the zero-require contract
(scripts/stack-drift-check.sh:8-9 fails CI on any `require` line) and the
vendorHash-free hermetic nix build. Root adoption would also be dead code:
no listener, no kubelet, no long-running process (AGENTS.md: "No net/http,
no network").

**pkg/provider check.** Also no: the provider module's dependency budget
is deliberately tiny (go-finding + toolsdk only), and its HealthCheck
surface is one `LookPath` — importing a probe SDK to wrap a binary-presence
check would be dependency theater.

**The one honest adjacent observation (named, not adoption-shaped):** like
the httputil report (group 2) found, the battery has no panic guard —
`rg -n "recover\(\)" -g '*.go'` over the whole consumer returns rc=1, so a
panicking gate would crash the run and mask later gates, contradicting the
"one foreign red never masks later gates" doctrine. go-health's `NewChecks`
proves the fleet solved panic-isolation for check batches
(go-health/README.md:444: "recovers a panicking check into that check's
error"), but the docs-gates-shaped fix is a three-line stdlib
`defer recover()` in the gate-capture wrapper, not an import. No gate panic
has been paid for to date.

RECONSIDER-WHEN: docs-gates (or a sibling module in this repo) grows a
long-running surface — a battery daemon, a fleet status endpoint, or a
watch-mode that re-runs gates on corpus changes. That module should adopt
go-health from day one in a separate go module (the pkg/provider pattern),
with the gates exposed as check functions; until then the insulation is
correct, not neglect.

## Ranked opportunities

None. (The panic-guard note is a stdlib hardening item inside gatecli,
already recorded in the httputil report; it is not a go-health adoption.)

## How to verify

```console
$ git -C /home/lars/projects/go-health rev-parse 3ca66446e1f8164f28e1ab788ad03727f6c99307^{commit}
3ca66446e1f8164f28e1ab788ad03727f6c99307
# — the reviewed commit, pinned: the repository tip drifts under the fleet's auto-commit daemon
$ git -C /home/lars/projects/go-health tag --sort=-v:refname | head -1
v0.10.0
$ sed -n '5,9p' /home/lars/projects/go-health/go.mod
require (
	github.com/larsartmann/go-branded-id v0.7.1
	github.com/samber/do/v2 v2.1.0
	github.com/samber/go-type-to-string v1.8.0
)

$ rg -l "larsartmann/go-health" -g '*.go' /home/lars/projects/docs-gates
# rc=1 — zero Go-file references in either module
$ rg -n "samber" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod
# rc=1 — no samber anywhere in either module graph
$ rg -n "recover\(\)" -g '*.go' /home/lars/projects/docs-gates
# rc=1 — no panic guard (the gap-analysis observation)

# the consumer's aggregation (sequential, stdio-captured, exit-code shaped):
$ rg -n "func gates\(|func runGates\(|func renderMatrix\(|func renderJSON\(" \
    /home/lars/projects/docs-gates/internal/battery/battery.go
58:func gates(root string, cfg config.Config) []gate {
174:func runGates(list []gate, skipGates []string, timeout time.Duration, jsonMode bool) []result {
264:func renderJSON(results []result, root string) int {
317:func renderMatrix(results []result) bool {

# the provider's health hook (installation check, not dependency health):
$ rg -n "HealthCheck: func" /home/lars/projects/docs-gates/pkg/provider/provider.go
291:	HealthCheck: func(_ context.Context) error {
```
