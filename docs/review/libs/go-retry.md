# go-retry utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v0.9.0 (latest local tag, fetched 2026-11-23) + HEAD a0f931dee9ddc3cad8b27cdd962292eb760f37b7 (`v0.9.0` is HEAD's ancestor tag; `git tag --list` confirms). Nested module: `tools` (untagged — dev-tool pins only, not a capability).
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 1 (resilience/error family) subagent of the multi-library utilization audit fleet.

## Verdict

NOT-APPLICABLE — go-retry is an exponential-backoff/jitter retry loop for transient failures, and docs-gates has no retryable boundary by design: every verb is a single-shot, deterministic, byte-pinned run whose failure policy is fail-fast-and-name-it, not retry.

## Score

- Inputs N = 11 (inventory rows below).
- A = 0 (no row addresses a problem docs-gates has; the nearest adjacency — bounded waiting — is a timeout problem the consumer already solves with one-shot watchdogs).
- M = 0.
- Score = **0** (0/0; by convention 0 when nothing is applicable) — band **low**. For a NOT-APPLICABLE library the score is noise; the verdict is the signal.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | `retry.Do` — bounded retry loop over an `AttemptFunc` | No — no operation is ever re-run; gates run once (`no first-failure abort` means all gates run once, not that any runs twice) | N/A | go-retry/retry.go:99; consumer internal/battery (single pass, FAIL rows recorded, never re-executed) |
| 2 | `DoWithValue[T]` — retry returning the success value | No | N/A | go-retry/retry.go:159 |
| 3 | `Config` + `DefaultConfig()` — MaxAttempts/InitialDelay/MaxDelay/Multiplier | No — there is no second attempt to space | N/A | go-retry/config.go:20,73 |
| 4 | Exponential backoff + jitter — `Backoff`, `ComputeDelay`, `JitterStrategy`, `WithRandomSource` | No — and actively hostile: jitter/sleep would inject nondeterminism into a battery whose output is byte-pinned by parity goldens | N/A | go-retry/retry.go:275-307, options.go:51-90; consumer internal/battery/parity_test.go |
| 5 | Pluggable `IsRetryable` predicate (defaults to `errorfamily.IsRetryable`) | No — nothing branches on retryability; see the go-error-family report for why classification is out of domain | N/A | go-retry/config.go (IsRetryable field); go-error-family/family.go:177 |
| 6 | `OnRetry` / `OnExhausted` hooks | No — the consumer's "exhausted" behavior is one FATAL line and an exit code, immediately | N/A | go-retry/options.go:38,44 |
| 7 | `DelayFunc` per-attempt override | No | N/A | go-retry/options.go:32 |
| 8 | Context-ending semantics — `ErrCanceled`, `ErrDeadlineExceeded`, `awaitBackoff` | Adjacent only — docs-gates has bounded waits, but as one-shot TIMEOUTS (watchdogs), not waits between retries | N/A — different problem, already solved | go-retry/retry.go:55-68,183; consumer internal/battery/battery_watchdog.go:33, internal/gatecli/gatecli_capture.go:141 |
| 9 | `ErrExhausted` typed terminal error (errorfamily Infrastructure) | No — terminal failures map onto the frozen 0/1/2 exit contract | N/A | go-retry/retry.go:45; consumer internal/gatecli/gatecli.go:29-32 |
| 10 | `FromPolicy(errorfamily.RetryPolicy)` — advisory-policy → loop-config bridge | No — no policy producer, no loop consumer | N/A | go-retry/config.go:86 |
| 11 | `tools` nested module (actionlint/govulncheck pins) | No — dev tooling of the library itself, not a capability | N/A | go-retry/tools/tools.go |

## What docs-gates does today

Zero usage, zero presence in either module graph. Verified:

```console
$ rg -n "go-retry|goretry|larsartmann/go-retry" /home/lars/projects/docs-gates -g '!docs/review/**' --type go
# rc=1, no matches

$ rg -n "go-retry" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1, no matches
```

(The only textual mention in the tree is `scripts/citation-sweep.py` — this audit fleet's scaffolding.)

What docs-gates does instead at every boundary where a lesser setup would reach for retry:

- **No sleep, no backoff, anywhere in non-test code** — verified: `rg -n "time\.Sleep" ... | rg -v _test` → rc=1. The only timer constructions are two one-shot watchdogs: `internal/battery/battery_watchdog.go:33` (`time.NewTimer(timeout / halfBudgetDivisor)` — informational half-budget warning on stderr) and `internal/gatecli/gatecli_capture.go:141` (`time.NewTimer(timeout)` — the per-gate cap that turns a wedged gate into a FAIL row `rc=timeout`). Both are timeouts, the opposite of retries: they bound a single attempt instead of funding more of them.
- **Git execs are single-shot with graceful degradation, not retries** — `internal/docsannotate/git.go:18-40` (`gitRunner.run`): one `exec.CommandContext`, an `*exec.ExitError` maps to the exit code, a missing git (nix sandbox) maps to `rcNoGit` so hash/tag rules warn-skip instead of crashing. Same single-shot shape at `internal/docsarchive/docsarchive.go:203` and `internal/docsfreshness/preflight_land.go:74`.
- **The provider execs the battery once per Detect** — `pkg/provider/provider.go:148-171` (`runBattery`): one `exec.CommandContext`, exit code captured, decode failure becomes its own finding. Re-running is BuildFlow's prerogative (Detect is read-only), not an in-process loop.
- **Wedge doctrine is loud, not patient** — AGENTS.md: "the capture-pipe deadlock (fixed by concurrent drainers; `gate_timeout_sec` bounds it; a new wedge = suspect a NEW park)". The resilience posture is deterministic fail-fast with named evidence — retrying would mask exactly the wedge class the battery exists to surface, and any retry-printed noise would threaten the byte-pinned goldens.

There is no `file:line` in docs-gates where a backoff loop, jitter, or attempt counter is hand-rolled — nothing to deduplicate against go-retry.

## Gap analysis — should docs-gates adopt it?

No. The fleet argument scores zero here on both prongs: docs-gates does not re-implement a retry loop (there is none), and the problems go-retry solves (transient network/db failures, at-least-once callers) are problems docs-gates deliberately does not have — "No net/http, no network calls, no retry/backoff, no time.Sleep in non-test code" is the stated architecture. For a CI gate, determinism IS the product: byte-pinned parity goldens, a frozen exit-code contract, point-in-time verdicts. A retry loop with jitter is not merely unused surface; importing its semantics (OnRetry prints, re-executed gates) would be a regression against the goldens and against the fail-fast doctrine.

Root adoption is barred by the zero-require contract (`scripts/stack-drift-check.sh`) and would transitively pull go-error-family. Provider adoption has no target: `runBattery`'s one failure mode worth tolerating (binary missing from PATH) is a permanent condition until the operator installs it — retrying a `LookPath` miss is noise, and `provider.go:291-297` already turns it into an actionable HealthCheck error.

RECONSIDER-WHEN: docs-gates gains a genuinely transient external boundary — the realistic first candidate is a network-facing integration (a git-hosting API for verdict-link resolution, a remote report corpus). That module should adopt go-retry from day one (with retry telemetry kept off the byte-pinned stdout paths), while the root stays stdlib-only. If the battery ever grows an optional `--retry-flaky-gates` developer convenience, go-retry belongs behind that flag in the provider module, never in golden-pinned paths.

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
```
