# go-health-dashboard utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v0.13.3 (latest local tag, fetched 2026-11-23) + HEAD 19af9cf4993011afb71451ca0927e70ca0acba67. Single module `github.com/larsartmann/go-health-dashboard`.
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 4 (ops/health/auth/CLI-framework) subagent of the multi-library utilization audit fleet.
- Fleet stance addressed: "if all my projects re-implement everything, I will only fix bugs all day and none benefit from collective improvements" — does stdlib-only insulation here avoid real re-implementation debt, or leave docs-gates hand-rolling something the fleet solved?

## Verdict

NOT-APPLICABLE — go-health-dashboard is a real-time browser status page (templ + Datastar SSE + OIDC + Prometheus) that renders a live go-health probe, and docs-gates owns no HTTP or HTML render surface at all: its entire human surface is byte-pinned plain-text matrix output and its machine surface is one JSON document on stdout, so there is nothing for the dashboard to render and no re-implementation to retire.

## Score

- Inputs N = 12 (inventory rows below).
- A = 1 (the "verdict rendering" concept row — a human-triage view over check results; the battery matrix is the only place the concept lives, and it is deliberately terminal text).
- M = 0.
- Score = round(100 × 0/1) = **0** — band **low**. Expected for NOT-APPLICABLE.

- Applicable rows are exactly the inventory rows graded **Missed Opportunity** (the consumer has the underlying need; the capability is hand-rolled, solved in-domain, or declined by contract). Every **Not Applicable** row carries a one-line reason. `M` counts rows graded Fully Leveraged.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | Live HTML dashboard over a go-health `Prober` (`/health`, badges, severity tables) | No — requires a go-health probe; docs-gates has none (see go-health report) | N/A | go-health-dashboard/README.md:8-11 |
| 2 | Datastar SSE real-time updates (`/health/sse`, push-on-change, heartbeats) | No — no server, no browser clients, no long-lived process | N/A | go-health-dashboard/README.md:444-455 |
| 3 | Content negotiation on `/health` (HTML vs JSON) | No — the battery's `--json` flag already splits machine/human faces by stream, byte-pinned | N/A | go-health-dashboard/README.md:20; internal/battery/battery.go:97-99 |
| 4 | Kubelet probe passthrough (`/healthz`, `/readyz`, `/startupz`) | No — no kubelet; exit codes are the contract | N/A | go-health-dashboard/README.md:112-114 |
| 5 | Health trend sparkline + history export (`WithTrend`, `/health/trend`, CSV/JSON) | No — docs-gates keeps history in timestamped report files by doctrine; a sparkline has no host | N/A | go-health-dashboard/README.md:340-347 |
| 6 | Prometheus exposition (`/health/metrics`, latency histograms) | No — nothing scrapes docs-gates; findings are CI-time, not scrape-time | N/A | go-health-dashboard/README.md:302-338 |
| 7 | Webhook push on transitions (`WithWebhook`, HMAC signing) | No — no transitions; every run is independent and point-in-time | N/A | go-health-dashboard/README.md:358-375 |
| 8 | Auth: middleware hook + native OIDC login (PKCE, JWKS, HMAC sessions) | No — nothing served to authenticate | N/A | go-health-dashboard/README.md:247-300 |
| 9 | Rate limiting, request logging, CSP, public-mode anonymization | No — no requests | N/A | go-health-dashboard/README.md:168-190 |
| 10 | Aggregate/federation rendering (one card per source, worst-of badges) | No — no probes to merge (and the fleet already has go-health's `federation` for the service side) | N/A | go-health-dashboard/README.md:377-391 |
| 11 | `health-hub` federation hub binary + NixOS module (env-configured, no code) | No — docs-gates publishes no health document to federate | N/A | go-health-dashboard/README.md:393-442 |
| 12 | Human-triage VIEW over check results (failures-first grouping, evidence strip) | Concept only — the battery matrix is the terminal-text equivalent, deliberately | Missed Opportunity (solved in-domain: the battery matrix is the terminal-text view, deliberately) | go-health-dashboard/README.md:205-245; internal/battery/battery.go:317-340 |

Dependency note: the dashboard exists precisely because its dependency
weight would pollute go-health — 16 direct requires including templ,
chromedp, go-datastar, go-sse, templ-components, and prometheus libs
(go-health-dashboard/go.mod:5-20; the "Why a Separate Repo" rationale at
go-health-dashboard/README.md:22-27). Its `[Unreleased]` section is active
(native OIDC login, SSE lifecycle metrics, per-IP rate limits —
go-health-dashboard/CHANGELOG.md:16 onwards). Nothing here can land in a
stdlib-only module.

## What docs-gates does today

Zero usage, zero presence. Verified:

```console
$ rg -n "go-health-dashboard" -g '*.go' -g 'go.mod' -g 'go.sum' /home/lars/projects/docs-gates
# rc=1, no matches in either module
```

The only "go-health-dashboard" string in the repo is a planning-doc
mention of it as an UNKNOWN runner class
(docs/planning/2026-11-15_marker-grammar-and-tooling-split.md:79) and the
audit scaffolding list in scripts/citation-sweep.py:44 — prose, not code.

The hand-rolled-equivalent hunt (any HTML/render surface the repo owns)
found none to name, verified from both directions:

- No template engine anywhere: `rg -l "text/template|html/template" -g '*.go'`
  over the repo matches only internal/docsstub/docsstub.go — whose
  "template" is a checksum of unfilled doc scaffolds (crypto/md5 identity
  matching, internal/docsstub/docsstub.go:17-19), not a render library.
  The "html" hits in internal/docsfreshness/explain.go:39-49 are the
  HtmlBlindSpot detector (flagging stray `2*.html` report snapshots the
  `.md` glob misses) — reading HTML filenames, never writing HTML.
- The render surfaces docs-gates owns are all plain text, by contract:
  the root usage built with a `strings.Builder`
  (cmd/docs-gates/main.go:165-182), the per-gate `Usage` constants pinned
  by internal/docsfreshness/usage_test.go:15 and
  internal/battery/usage_test.go:15, the verdict matrix
  (internal/battery/battery.go:317-340), and the one-shot JSON document
  (internal/battery/battery.go:264-313) whose snake_case shape the provider
  strict-decodes (pkg/provider/provider.go:92-102). All byte-pinned by
  parity goldens — an HTML render layer would be a new contract, not a
  replacement for an old one.
- The JSON face is integration-ready by design: its gate rows are generic
  so a new battery gate lands without a provider change (AGENTS.md
  battery row) — any future dashboard consumer would read this document,
  not scrape the human matrix. That is the whole integration surface a
  status-board deployment would need, and it already exists.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** The dashboard's job is to stop every
fleet *service* from hand-rolling a status page — a genuinely shared
problem for the owner's long-running HTTP projects, and the right answer
there is exactly this library. docs-gates has no page to serve: it is a
judge whose output contract is byte-stable terminal text plus one JSON
document, consumed by CI jobs, the BuildFlow provider, and humans reading
a log. Where a service would render a dashboard, docs-gates prints a
matrix — and that matrix is pinned by goldens whose entire value is that
they do not change. There is no duplication: nothing in docs-gates renders,
serves, streams, or authenticates anything. The insulation is not avoiding
a solved problem; there is no problem.

**Root-contract bar.** Absolute: 16 direct requires
(go-health-dashboard/go.mod:5-20) against a zero-require module
(scripts/stack-drift-check.sh:8-9 fails CI on any `require` line), plus a
build-time `GOEXPERIMENT=jsonv2` requirement the root module deliberately
does not carry. pkg/provider is equally wrong a home — its budget is
go-finding/toolsdk only, and its output is findings, not HTML.

**The honest adjacent thought (declined, with reason):** "a fleet dashboard
over every repo's battery verdicts" is a conceivable future — but it would
be a NEW deployment (the `health-hub` shape: one hub fetching per-repo
verdicts), built on go-health + go-health-dashboard in their own module,
with docs-gates only needing to keep its `run --json` face stable. That is
a RECONSIDER-WHEN, not a gap: nothing in the current consumer would change
to enable it.

RECONSIDER-WHEN: the owner wants a standing visual board of docs-health
across repos (fleet battery verdicts, trend over time). Then build a
separate hub module that adopts go-health-dashboard (or its `health-hub`
binary) and consumes stored `run --json` documents; docs-gates itself
should stay render-free — the JSON face is the integration contract, and
it already exists.

## Ranked opportunities

None. (A fleet verdict dashboard is a new deployment, not a change to this
consumer; recorded under RECONSIDER-WHEN.)

## How to verify

```console
$ git -C /home/lars/projects/go-health-dashboard log -1 --format='%H'
19af9cf4993011afb71451ca0927e70ca0acba67
$ git -C /home/lars/projects/go-health-dashboard tag --sort=-v:refname | head -1
v0.13.3
$ sed -n '5,20p' /home/lars/projects/go-health-dashboard/go.mod
# 16 direct requires: templ, chromedp, go-humanize, go-datastar (+static),
# go-health v0.10.0, go-sse (+ssetest), templ-components x4,
# prometheus client_model + common, samber/do/v2

$ rg -n "go-health-dashboard" -g '*.go' -g 'go.mod' -g 'go.sum' /home/lars/projects/docs-gates
# rc=1 — zero code/module-graph references
$ rg -l "text/template|html/template" -g '*.go' /home/lars/projects/docs-gates
internal/docsstub/docsstub_test.go
internal/docsstub/docsstub.go
# both are scaffold-checksum matching, not rendering:
$ rg -n "crypto/md5" /home/lars/projects/docs-gates/internal/docsstub/docsstub.go
19:	"crypto/md5" //nolint:gosec // checksum identity, not security

# the consumer's actual render surfaces (plain text, golden-pinned):
$ rg -n "func rootUsage\(" /home/lars/projects/docs-gates/cmd/docs-gates/main.go
165:func rootUsage() string {
$ rg -n "func renderMatrix\(|func renderJSON\(" /home/lars/projects/docs-gates/internal/battery/battery.go
264:func renderJSON(results []result, root string) int {
317:func renderMatrix(results []result) bool {
```
