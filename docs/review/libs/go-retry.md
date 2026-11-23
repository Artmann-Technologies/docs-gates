# go-retry utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v0.9.0 (latest local tag, fetched 2026-11-23) + HEAD a0f931dee9ddc3cad8b27cdd962292eb760f37b7 (`git tag --list` confirms v0.9.0; HEAD is post-tag CI sync). Nested module: `tools` (untagged — dev-tool pins only, not a capability).
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 1 (resilience/error family) subagent of the multi-library utilization audit fleet.
- Fleet stance addressed per library: "if all my projects re-implement everything, I will only fix bugs all day and none benefit from collective improvements" — does stdlib-only insulation here avoid real re-implementation debt, or leave docs-gates hand-rolling something the fleet solved?

## Verdict

NOT-APPLICABLE — go-retry is an exponential-backoff/jitter retry loop for transient failures, and docs-gates has no retryable boundary by design: every verb is a single-shot, deterministic, byte-pinned run whose failure policy is fail-fast-and-name-it, not retry.

## Score

- Inputs N = 13 (inventory rows below).
- A = 0 (no row addresses a problem docs-gates has; the nearest adjacency — bounded waiting — is a timeout problem the consumer already solves with one-shot watchdogs).
- M = 0.
- Score = **0** (0/0; by convention 0 when nothing is applicable) — band **low**. For a NOT-APPLICABLE library the score is noise; the verdict is the signal.

- Applicable rows are exactly the inventory rows graded **Missed Opportunity** (the consumer has the underlying need; the capability is hand-rolled, solved in-domain, or declined by contract). Every **Not Applicable** row carries a one-line reason. `M` counts rows graded Fully Leveraged.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | `retry.Do` — bounded retry loop over an `AttemptFunc` | No — no operation is ever re-run; gates run once ("no first-failure abort" means all gates run once, not that any runs twice) | N/A | go-retry/retry.go:99; consumer internal/battery (single pass; FAIL rows recorded, never re-executed) |
| 2 | `DoWithValue[T]` — retry returning the successful attempt's value | No | N/A | go-retry/retry.go:159 |
| 3 | `Config` + `DefaultConfig()` — MaxAttempts (including first call), InitialDelay 100ms, MaxDelay 5s, Multiplier 2.0 | No — there is no second attempt to space | N/A | go-retry/config.go:20-73 |
| 4 | Exponential backoff math — `Backoff`, `ComputeDelay` | No | N/A | go-retry/retry.go:275-307 |
| 5 | Jitter — additive jitter with `JitterStrategy` (none/default/decorrelated), `WithRandomSource` for determinism in tests | No — and actively hostile: jitter/sleep injects nondeterminism into a battery whose output is byte-pinned by parity goldens | N/A | go-retry/options.go:51-90; consumer internal/battery/parity_test.go |
| 6 | Pluggable `IsRetryable` predicate (defaults to `errorfamily.IsRetryable`) | No — nothing branches on retryability; see the go-error-family report for why classification is out of domain | N/A | go-retry/config.go; go-error-family/family.go:177 |
| 7 | `OnRetry` hook (after each failed attempt, before sleeping) | No | N/A | go-retry/options.go:38 |
| 8 | `OnExhausted` hook (once, after all attempts fail, with the last error as returned) | No — the consumer's "exhausted" behavior is one FATAL line and an exit code, immediately | N/A | go-retry/options.go:44 |
| 9 | `DelayFunc` per-attempt delay override (>0 replaces backoff; 0 = default) | No | N/A | go-retry/options.go:32 |
| 10 | Context-ending semantics — `ErrCanceled`, `ErrDeadlineExceeded`, `awaitBackoff` returns on ctx end | No — adjacent only: docs-gates has bounded waits, but as one-shot TIMEOUTS (watchdogs), not waits between retries | N/A — different problem, already solved | go-retry/retry.go:55-68,183; consumer internal/battery/battery_watchdog.go:33, internal/gatecli/gatecli_capture.go:141 |
| 11 | `ErrExhausted` typed terminal error (errorfamily Infrastructure) | No — terminal failures map onto the frozen 0/1/2 exit contract | N/A | go-retry/retry.go:45; consumer internal/gatecli/gatecli.go:29-32 |
| 12 | `FromPolicy(errorfamily.RetryPolicy)` — advisory-policy → loop-config bridge | No — no policy producer, no loop consumer | N/A | go-retry/config.go:86 |
| 13 | `tools` nested module (actionlint/govulncheck pins) | No — dev tooling of the library itself, not a capability | N/A | go-retry/tools/tools.go |

(Runtime dependency on go-error-family v0.12.0 — see that report; root adoption would be transitively barred even if a retry boundary existed.)

## What docs-gates does today

Zero usage, zero presence in either module graph. Verified:

```console
$ rg -n "go-retry|goretry|larsartmann/go-retry" /home/lars/projects/docs-gates -g '!docs/review/**' --type go
# rc=1, no matches

$ rg -n "go-retry" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1, no matches
```

(The only textual mention in the tree is `scripts/citation-sweep.py` — this audit fleet's own
scaffolding.)

What docs-gates does instead at every boundary where a lesser setup would reach for retry:

### No sleep, no backoff, anywhere in non-test code

Verified: `rg -n "time\.Sleep" … | rg -v _test` → rc=1. The only timer constructions are two
one-shot watchdogs — `internal/battery/battery_watchdog.go:33`:

```go
timer := time.NewTimer(timeout / halfBudgetDivisor)   // informational half-budget warning (stderr)
```

and `internal/gatecli/gatecli_capture.go:141`:

```go
timer := time.NewTimer(timeout)   // the per-gate cap; expiry => FAIL row (rc=timeout)
```

Both are timeouts, the opposite of retries: they bound a single attempt instead of funding more
of them. The watchdog line is stderr-only and informational by design; the capture cap turns a
wedged gate into a named FAIL row so the battery moves on.

### Git execs are single-shot with graceful degradation, not retries

`internal/docsannotate/git.go:18-40` (`gitRunner.run`): one `exec.CommandContext`, an
`*exec.ExitError` maps to the exit code, and a missing git (nix sandbox) maps to `rcNoGit` so
hash/tag rules warn-skip instead of crashing. Same single-shot shape at
`internal/docsarchive/docsarchive.go:203` and `internal/docsfreshness/preflight_land.go:74`.

### The provider execs the battery once per Detect

`pkg/provider/provider.go:148-171` (`runBattery`): one `exec.CommandContext`, exit code
captured, decode failure becomes its own finding. Re-running is BuildFlow's prerogative (Detect
is read-only), not an in-process loop.

### Wedge doctrine is loud, not patient

AGENTS.md, paid-for section: "the capture-pipe deadlock (fixed by concurrent drainers;
`gate_timeout_sec` bounds it; a new wedge = suspect a NEW park)". The resilience posture is
deterministic fail-fast with named evidence — retrying would mask exactly the wedge class the
battery exists to surface, and any retry-printed noise would threaten the byte-pinned goldens.

There is no `file:line` in docs-gates where a backoff loop, jitter, or attempt counter is
hand-rolled — nothing to deduplicate against go-retry.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** It scores zero on both prongs: docs-gates does not
re-implement a retry loop (there is none), and the problems go-retry solves (transient
network/db failures, at-least-once callers) are problems docs-gates deliberately does not have —
"No net/http, no network calls, no retry/backoff, no time.Sleep in non-test code" is stated
architecture, not accident. For a CI gate, determinism IS the product: byte-pinned parity
goldens, a frozen exit-code contract, point-in-time verdicts ("timestamp the run"). A retry loop
with jitter is not merely unused surface; importing its semantics (OnRetry prints, re-executed
gates) would be a regression against the goldens and against the fail-fast doctrine. This is the
insulation half of the fleet argument working exactly as intended: the library solves a problem
docs-gates does not have.

**Root-contract bar.** Barred by the zero-require contract (`scripts/stack-drift-check.sh`
fails CI) and would transitively pull go-error-family.

**pkg/provider check.** No target: `runBattery`'s one failure mode worth tolerating (binary
missing from PATH) is a permanent condition until the operator installs it — retrying a
`LookPath` miss is noise, and `provider.go:291-297` already turns it into an actionable
HealthCheck error carrying the pinned install line.

RECONSIDER-WHEN: docs-gates gains a genuinely transient external boundary — the realistic first
candidate is a network-facing integration (a git-hosting API for verdict-link resolution, a
remote report corpus). That module should adopt go-retry from day one, with retry telemetry kept
off the byte-pinned stdout paths, while the root stays stdlib-only. If the battery ever grows an
optional `--retry-flaky-gates` developer convenience, go-retry belongs behind that flag in the
provider module, never in golden-pinned paths.

## Ranked opportunities

None.

## How to verify

```console
$ git -C /home/lars/projects/go-retry log -1 --format='%H'
a0f931dee9ddc3cad8b27cdd962292eb760f37b7
$ git -C /home/lars/projects/go-retry tag --list | tail -2
v0.8.0
v0.9.0

$ rg -n "go-retry|goretry|larsartmann/go-retry" /home/lars/projects/docs-gates -g '!docs/review/**' --type go
# rc=1 (zero Go-file hits)
$ rg -n "go-retry" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 (not in any module graph)

$ rg -n "time\.Sleep" /home/lars/projects/docs-gates --type go | rg -v _test
# rc=1 — no sleep in non-test code
$ rg -n "time\.NewTimer|time\.After|time\.Ticker" /home/lars/projects/docs-gates --type go | rg -v _test
internal/battery/battery_watchdog.go:33:  timer := time.NewTimer(timeout / halfBudgetDivisor)
internal/gatecli/gatecli_capture.go:141:  timer := time.NewTimer(timeout)
# — two one-shot watchdogs; no retry loops (the docsarchive.go:143 `for` is a glob-entry iteration)

$ sed -n 18,40p /home/lars/projects/docs-gates/internal/docsannotate/git.go
# — single-shot exec.CommandContext git runner; ExitError → exit code; no-git sandbox → rcNoGit warn-skip
$ rg -n "exec.CommandContext" /home/lars/projects/docs-gates --type go | rg -v _test
# → git.go:25, docsarchive.go:203, preflight_land.go:74, pkg/provider/provider.go:149 (all single-shot)
```
