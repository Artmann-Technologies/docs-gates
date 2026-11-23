# httputil utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v1.9.0 (latest local tag, fetched 2026-11-23) + HEAD 34dbf05f670d28830e61170a087f552070279b74 (`git describe` → `v1.9.0-24-g34dbf05`). Nested modules in the tree: `server_timing` v1.0.3 (own report), `endpoints` v0.1.0 (own report), plus the non-module `httpspec` and `apidiff` packages inside the root module.
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 2 (HTTP/serialization) subagent of the multi-library utilization audit fleet.
- Fleet stance addressed: "if all my projects re-implement everything, I will only fix bugs all day and none benefit from collective improvements" — does stdlib-only insulation here avoid real re-implementation debt, or leave docs-gates hand-rolling something the fleet solved?

## Verdict

NOT-APPLICABLE — httputil is an `net/http` middleware/server toolkit (CORS, compression, CSRF, rate limiting, security headers, server lifecycle), and docs-gates is an exit-code CLI with no HTTP surface at all: zero `net/http` imports anywhere in either of its modules, so there is no shared problem for the fleet argument to act on.

## Score

- Inputs N = 18 (inventory rows below).
- A = 1 (the panic-isolation concept row — "one unit's crash must not kill the whole run", the idea under `Recovery`; the battery is the only place the concept could land).
- M = 0.
- Score = round(100 × 0/1) = **0** — band **low**. Expected for NOT-APPLICABLE; the verdict carries the meaning, the number measures correctly-unrealized leverage.

- Applicable rows are exactly the inventory rows graded **Missed Opportunity** (the consumer has the underlying need; the capability is hand-rolled, solved in-domain, or declined by contract). Every **Not Applicable** row carries a one-line reason. `M` counts rows graded Fully Leveraged.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | CORS middleware (preflights, wildcard origins, LNA preflights) | No — no HTTP server, no browser clients | N/A | httputil/cors.go |
| 2 | Client-IP extraction behind proxies (`X-Forwarded-For` precedence) | No — no requests, no remote peers | N/A | httputil/clientip.go |
| 3 | `ResponseRecorder` writer wrapper (status capture, Flusher/Hijacker/Pusher delegation) | No — no `http.ResponseWriter` anywhere | N/A | httputil/recorder.go |
| 4 | `Chain` / `Compose` / `MiddlewareStack` middleware composition | No — no middleware chain; gates are `Run([]string) int` packages | N/A | httputil/compose.go, httputil/stack.go |
| 5 | Security headers, CSP nonce | No — no response headers | N/A | httputil/security.go, httputil/nonce.go |
| 6 | Request-ID propagation | No — no requests; runs are identified by timestamp line, not IDs | N/A | httputil/requestid.go |
| 7 | Panic recovery middleware (one handler's panic → 500, server lives) | No — the HTTP-handler version of row 18; no handler exists in the consumer | N/A | httputil/recovery.go |
| 8 | Request timeout middleware (context deadline per request) | No — gates take no context by design; the battery's timeout is goroutine abandonment, see gap analysis | N/A | httputil/timeout.go:11-25 |
| 9 | Structured per-request logging | No — `gatecli.Outf`/`Errf`/`Livef` are the whole output story, humans on stderr | N/A | httputil/logging.go; internal/gatecli/gatecli_output.go |
| 10 | Response compression + request decompression with bomb protection | No — no bodies, no encodings | N/A | httputil/compression.go, httputil/decompression.go |
| 11 | ETag conditional requests | No — addressed in the go-etag report (group 1) | N/A | httputil/README.md:310-324 |
| 12 | `Server` wrapper: configurable `http.Server`, lifecycle, graceful shutdown | No — the binary runs and exits; nothing long-lived to serve or drain | N/A | httputil/server.go |
| 13 | Kubernetes health handlers (`/health`, `/live`, `/ready`) | No — health is the exit code itself (0/1/2) | N/A | httputil/health.go |
| 14 | CSRF protection (nosurf, HTMX helpers) | No — no cookies, no forms, no origins | N/A | httputil/csrf.go |
| 15 | Keyed token-bucket rate limiting + trusted-proxy key extraction | No — nothing is called repeatedly by clients; the battery runs on demand | N/A | httputil/ratelimit_keyed.go |
| 16 | Content-language negotiation | No — output is English prose + JSON, no negotiation | N/A | httputil/language.go |
| 17 | `httpspec` behavioral spec suite (19 specs over any `http.Handler`) | No — nothing to wrap in an `http.Handler`; gates are tested through their real `Run` functions with captured stdio | N/A | httputil/README.md:138-162; AGENTS.md testing note |
| 18 | Panic isolation of independently-failing units (the concept under `Recovery`) | Yes in concept — battery gates are the units; currently NO `recover()` exists anywhere in docs-gates | Missed Opportunity (concept satisfied by nobody: a gate panic crashes the whole battery run) | consumer-wide `rg 'recover\(\)' --type go` → rc=1; httputil/recovery.go solves the HTTP-shaped version |

Dependency note for any future adoption: the root module is not stdlib-only —
httputil/go.mod:5-13 requires `justinas/nosurf`, `go-error-family`,
`go-etag/server`, `httputil/server_timing`, and `golang.org/x/time` (the README
calls this a minimal footprint, httputil/README.md:11). Adoption into the
docs-gates root module would break the zero-require contract outright.

## What docs-gates does today

Zero usage, zero presence. Verified:

```console
$ rg -n "github.com/larsartmann/httputil" /home/lars/projects/docs-gates --hidden -g '!.git'
# rc=1, no matches (the only "httputil" string in the repo is the audit-fleet
# scaffolding list in scripts/citation-sweep.py:35)

$ rg -c "net/http" /home/lars/projects/docs-gates --type go
# rc=1, zero files — no HTTP client, server, or handler code exists at all
```

`go.mod` (root) carries no requires (go.mod:1-3, enforced by
scripts/stack-drift-check.sh:8-9); pkg/provider/go.mod:5-10 requires only
`go-finding`, `go-finding/toolsdk`, and `go-error-family` — no httputil in
either module graph, no indirect traces in pkg/provider/go.sum.

Checked for hand-rolled HTTP utilities (the fleet-argument duplication
question) and found none to name:

- No HTTP client wrapper exists — the provider's only external interaction is
  a local process exec (`pkg/provider/provider.go:148-171`), not a transport.
- No server scaffolding exists — `cmd/docs-gates/verbs.go:77` dispatches
  argv verbs to `Run([]string) int` functions; there is no listener, mux, or
  handler anywhere.
- The nearest timing/deadline code — `gatecli.CaptureOutputWithTimeout`
  (internal/gatecli/gatecli_capture.go:114-156) and the half-budget watchdog
  (internal/battery/battery_watchdog.go:25) — is in-process goroutine
  abandonment over swapped `os.Stdout`/`os.Stderr` pipes, not request
  deadlines; judged in depth in the server_timing report. httputil's
  `Timeout` (httputil/timeout.go:11-25) is `context.WithTimeout` on a request
  context — a mechanism docs-gates gates deliberately cannot use ("The gate
  functions take no context", internal/gatecli/gatecli_capture.go:116-117).

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** The argument is at its weakest against this
pairing. httputil exists to stop every HTTP service in the fleet from
hand-rolling CORS/compression/CSRF/server-lifecycle — a real, shared, subtle
problem class across the owner's service projects. docs-gates has no HTTP
surface by deliberate architecture: it is an exit-code CLI whose two contracts
(byte-pinned stdio output and the 0/1/2 exit arithmetic) are defined by the
*absence* of a serving layer. There is no re-implementation to retire and no
file:line of duplicated middleware to name — the honest answer to the
orchestrator's question is that no duplication exists, because no overlap
exists.

**Root-contract bar.** Adoption into the root module would break
`scripts/stack-drift-check.sh` CI on the first `go get` (httputil's five
requires, httputil/go.mod:5-13) — and would import `net/http` machinery the
binary never executes. Dead weight on every `nix build`.

**pkg/provider check.** Also no target: the provider's job is one local exec
plus a strict JSON decode (pkg/provider/provider.go:92-102, :148-171). It
makes no HTTP requests and serves nothing, so no httputil capability has a
call site.

**The one honest adjacent gap (named, not adoption-shaped):** the battery has
no panic guard — `rg 'recover\(\)' --type go` over the whole consumer returns
zero files, so a panicking gate would crash the entire battery run and mask
the remaining gates, contradicting the "one foreign red never masks later
gates" doctrine (AGENTS.md). httputil's `Recovery` middleware proves the fleet
solved the *concept* for HTTP handlers, but the docs-gates-shaped fix is a
three-line stdlib `defer recover()` inside the gate-capture wrapper — not an
httputil import. No gate panic has been paid for to date (the documented wedge
classes were pipe deadlocks and races, both fixed), so even that stdlib change
is a hardening option, not owed debt.

RECONSIDER-WHEN: docs-gates (or a sibling module in this repo) grows an HTTP
surface — a fleet status endpoint serving battery verdicts, a dashboard
backend, or a remote report-corpus fetcher. That module should adopt httputil
from day one (it is precisely the fleet-shared problem class the owner names),
while the root CLI stays network-free; until then the insulation is correct,
not neglect.

## Ranked opportunities

None. (The panic-guard observation above is a stdlib hardening note inside
gatecli, not an httputil adoption.)

## How to verify

```console
$ git -C /home/lars/projects/httputil log -1 --format='%H'
34dbf05f670d28830e61170a087f552070279b74
$ git -C /home/lars/projects/httputil describe --tags
v1.9.0-24-g34dbf05
$ git -C /home/lars/projects/httputil tag --list | tail -1
v1.9.0

$ rg -n "github.com/larsartmann/httputil" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**'
scripts/citation-sweep.py:35:    "httputil",
# rc=0 only for the audit scaffolding list; zero code/module-graph hits
$ rg -c "net/http" /home/lars/projects/docs-gates --type go
# rc=1 — zero files import net/http (root or pkg/provider)
$ rg -n "recover\(\)" /home/lars/projects/docs-gates --type go
# rc=1 — no panic guard anywhere (the gap-analysis observation)
$ cat /home/lars/projects/docs-gates/pkg/provider/go.mod
# module .../pkg/provider; requires: go-finding v1.15.1, toolsdk v1.16.0,
# go-error-family v0.12.0 indirect — no httputil

# the consumer's deadline machinery (in-process, not request-shaped):
$ rg -n "func CaptureOutputWithTimeout|The gate functions take no context" \
    /home/lars/projects/docs-gates/internal/gatecli/gatecli_capture.go
116:// context, so an expired gate CANNOT be killed — it keeps running as an
121:func CaptureOutputWithTimeout(fn func() int, timeout time.Duration) (CapturedOutput, bool) {
$ rg -n "func startHalfBudgetWatchdog" /home/lars/projects/docs-gates/internal/battery/battery_watchdog.go
25:func startHalfBudgetWatchdog(gateName string, timeout time.Duration, warnf func(string, ...any)) func() {
```
