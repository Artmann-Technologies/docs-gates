# go-error-family utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v0.12.0 (latest local tag, fetched 2026-11-23) + HEAD 01d155efa67f35ea75ced0942ab2c4b1b406bde3 (`git describe` → `v0.12.0-13-g01d155e`). Submodules at HEAD: agent v0.2.5, bridge v0.3.4, diagnose v0.2.6, diagnose/git v0.5.4, diagnose/postgres v0.5.4, examples v0.3.2.
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f — root module stdlib-only (zero non-stdlib requires, `scripts/stack-drift-check.sh` in CI) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 1 (resilience/error family) subagent of the multi-library utilization audit fleet.

## Verdict

NOT-ADOPTED — zero API use anywhere in docs-gates (root or pkg/provider); it rides in `pkg/provider/go.sum` only as a transitive indirect of go-finding, and the root module's hand-rolled typed refusal carriers are a deliberate, standing-decision alternative that keeps the stdlib-only contract.

## Score

- Inputs N = 16 (inventory rows below).
- A = 6 (rows whose problem docs-gates actually has: typed error carriers, error interfaces, exit-code mapping, CLI boundary rendering, wrap hygiene, transitive presence).
- M = 0 (no row fully leveraged — by construction, the carrier slice is solved differently and the rest is out of domain).
- Score = round(100 × 0/6) = **0** — band **low**. Expected for NOT-ADOPTED; the verdict carries the meaning, not the number.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | `Family` behavioral classification — `Classify(err)` into six families (Rejection/Conflict/Transient/Corruption/Infrastructure/Orchestration), multi-error worst-of | Yes in principle (docs-gates returns errors) — but every refusal is a terminal CLI exit, never a retry/tone/audience decision | N/A — no caller ever branches on error class; failures map to a fixed 3-code contract | go-error-family/family.go:27, go-error-family/classify.go; consumer exits via `internal/gatecli/gatecli.go:29-32` (`ExitClean/ExitFindings/ExitUsage`) |
| 2 | Six small interfaces (`Coded`, `Classified`, `Contextual`, `Retryable`, `ExitCoder`, `HTTPStatuser`) | Partially — consumer carriers implement only `Error()` (+`Unwrap` where a cause exists) | N/A — carriers are unexported and exist to satisfy erraudit's generic-error lint, not a cross-package protocol | go-error-family/interfaces.go:16-62; consumer `internal/config/errors.go:13-23`, `internal/gatecli/gatecli.go:69-76` |
| 3 | `Error` reference implementation + constructors (`NewRejection` …), `WithContext`/`WithContextAny` | No — each consumer carrier is 3-8 lines, unexported, message spelled for a golden-pinned FATAL line | N/A — standing decision (AGENTS.md "Deliberate non-fixes" / erraudit carriers) | go-error-family/error.go, go-error-family/constructors.go; consumer carriers listed in "What docs-gates does today" |
| 4 | `ExitCode(err)` — BSD sysexits.h mapping (1/64/66/71/75…) | No — actively conflicts: docs-gates pins 0/1/2 forever | N/A — adopting would break the exit-code contract and every wrapper test | go-error-family/family.go:203; consumer `internal/gatecli/gatecli.go:24-32` |
| 5 | `HandleError(err)` CLI boundary — What/Why/Fix/WayOut structured rendering, message templates | Yes in principle — docs-gates renders FATAL lines at a CLI boundary | N/A — output is byte-pinned by parity goldens; human text is a contract, not a library default | go-error-family/handle.go:121; consumer FATAL lines pinned by `internal/battery/parity_test.go` |
| 6 | `Registry` + `Clone()` + `RegisterClassifier` predicate classification | No — no dynamic third-party error taxonomy anywhere | N/A — zero foreign error types to classify | go-error-family/registry.go:31 |
| 7 | `RegisterStdlibDefaults` (context/sql/os sentinel families) | No — the only stdlib errors handled (`*exec.ExitError`, `*json.SyntaxError`) are handled structurally, by type assertion | N/A — go-error-family/stdlib.go:30; consumer `pkg/provider/provider.go:165` (ExitError), `internal/config/errors.go:29` (SyntaxError) |
| 8 | `WrapOnce` — idempotent wrap against double-wrapping at API boundaries | Marginal — docs-gates has no layered API boundaries; errors are created once at the refusal site | N/A — the failure mode does not exist in a single-binary CLI | go-error-family/wraponce_test.go (API surface); no consumer double-wrap sites |
| 9 | `LogError` / `LogErrorContext` (slog structured fields) | No — root module has no slog and no logging surface; diagnostics ARE the output | N/A — would add a require for zero behavioral gain | go-error-family/log.go:22 |
| 10 | `HTTPStatus(err)` / `HTTPHandler` net/http middleware | No — no net/http anywhere in the consumer (root or provider) | N/A | go-error-family/http.go:17,54 |
| 11 | `Family.RetryPolicy()` advisory retry defaults (attempts + backoff, loop not run) | No — no retry, no backoff, no time.Sleep in non-test code; timeouts are one-shot watchdogs | N/A — see go-retry report for the retry-side analysis | go-error-family/retry.go:18-26; consumer `internal/battery/battery_watchdog.go:33` |
| 12 | `errorfamilytest` assertion helpers (`AssertFamily`, `AssertCode`, …) | No — root tests are stdlib-`testing` only, exercised through real `Run` functions | N/A — test-only dep would still break the zero-require root go.mod | go-error-family/errorfamilytest/errorfamilytest.go |
| 13 | `agent` submodule — root-cause analysis agent over diagnostic context (experimental v0.x) | No — no diagnosis loop; a failed git exec is a terminal refusal | N/A | go-error-family/agent/agent.go:1 |
| 14 | `bridge` submodule — samber/oops ↔ error-family adapters | No — samber/lo already withheld from root for stdlib-only reasons (standing decision, AGENTS.md) | N/A — same class of decline | go-error-family/bridge/bridge.go:1 |
| 15 | `diagnose` (+`diagnose/git`, `diagnose/postgres`) — deterministic live-system rules, structured `Fix{Summary, Command}` | No — docs-gates never investigates a live system; the git exec sites are verdict resolution, not diagnosis | N/A — go-error-family/diagnose/git checks repo/remotes state; consumer `internal/docsannotate/git.go:25` only resolves hashes | go-error-family/diagnose/diagnose.go:1, go-error-family/diagnose/git/git.go |
| 16 | Transitive presence in `pkg/provider` (indirect via go-finding) | Yes — it is in the provider module graph today | N/A (correctly inert) — go-finding's `FindingError` implements `errorfamily.Coded`/`Classified`, but provider.go never touches `FindingError` or any errorfamily symbol | `pkg/provider/go.mod:10` (`// indirect`), `pkg/provider/go.sum:11-12`; go-finding/errors.go:85-103 |

## What docs-gates does today

Zero direct usage. Verified with (run from the consumer root):

```console
$ rg -n "errorfamily|go-error-family" /home/lars/projects/docs-gates --type go -g '!docs/review/**'
# (no matches, rc=1)

$ rg -n "errorfamily|go-error-family" /home/lars/projects/docs-gates/pkg/provider --type go
# (no matches, rc=1)
```

The only occurrences in the whole consumer tree are dependency metadata, not code:

- `pkg/provider/go.mod:10` — `require github.com/larsartmann/go-error-family v0.12.0 // indirect`
- `pkg/provider/go.sum:11-12` — the two v0.12.0 hashes
- `flake.lock:184-198` — a `flake:false` source input pinned at ref v0.11.0 (rev 2d0679e0), feeding the **erraudit** tool build; erraudit's own go.mod requires go-error-family v0.12.0 (`/home/lars/projects/erraudit/go.mod:13`)

Why indirect: go-finding's root module requires go-error-family v0.12.0 directly (`/home/lars/projects/go-finding/go.mod:6`), and its `FindingError` carries `ErrorCode() string` (implements `errorfamily.Coded`) and `ErrorFamily() errorfamily.Family` (implements `errorfamily.Classified`) — go-finding/errors.go:85-103. So the errorfamily **type** `Family` is technically reachable through go-finding's exported API into `pkg/provider/provider.go`'s import closure — but provider.go constructs plain `finding.NewFinding` values and returns plain `fmt.Errorf` errors; no errorfamily symbol is named (verified above).

What docs-gates does INSTEAD (the hand-rolled equivalent the fleet argument must beat): typed refusal carriers — unexported struct + constructor + optional `Unwrap`, one per refusal class, message frozen for the FATAL line:

- `internal/config/errors.go:13-23` `syntaxError` (+ line/column decoration, `Unwrap`); `:98-113` `loadError`/`newLoadError` (+ `Unwrap`); `:118-127` `invalidError`/`invalidf`
- `internal/gatecli/gatecli.go:69-76` `notDirectoryError`, `unknownFlagsError`; `:179-184` `valuedFlagError`
- `internal/docsrows/rows.go:87-89` `usageError`
- `internal/docspurge/parse.go:15-19` `purgeError`; `internal/docspurge/docspurge_collect.go:161-179` `missingAnchorError`/`newMissingAnchorError`
- `internal/docsstrike/strike.go:15-19` `strikeError`; `internal/docsstrike/parse.go:150-178` `strikeInputsError`/`newStrikeInputsError`
- `internal/docsfreshness/corpus.go:14-21` `candidatesError` (+ `Unwrap`)
- `internal/docsagents/docsagents.go:177-181` `directoryAgentsError`

Total hand-rolled carrier surface: ~11 types across 7 files, each single-purpose. This is the pattern AGENTS.md names ("typed refusal carriers … satisfying erraudit while keeping idiomatic error returns") and it is lint-enforced (`erraudit lint ./... --max` green since 2026-11-12). Notably, the enforcement tool itself (erraudit) is built on go-error-family — the fleet's error discipline already flows through the library at the tooling layer, while the consumer's code layer stays carrier-local.

## Gap analysis — should docs-gates adopt it?

The fleet argument ("if all my projects re-implement everything, only bugs get fixed all day") is strong when a project hand-rolls a **shared, evolving** implementation. Test it against what docs-gates hand-rolls here:

- The carriers are not a re-implementation of go-error-family. They implement no families, no registry, no exit-code derivation, no retry decision — they are message structs. What they duplicate from go-error-family is only the `Error() string` + optional `Unwrap()` shape, which is stdlib `errors` itself, not library surface. There is no `file:line` where go-error-family's actual capabilities (Classify, ExitCode, HandleError, Registry, diagnose rules) are being re-implemented — those problems (behavioral branching on error class, sysexits exits, What/Why/Fix rendering, live-system diagnosis) are problems docs-gates does not have. A single-shot CLI with a frozen 0/1/2 exit contract and byte-pinned output has no consumer for a classification protocol.
- Adopting into the ROOT would break the zero-require contract (CI stack-drift fails), force re-pinning every FATAL-line golden whose wording moved through `HandleError` templates, and import a 6-family taxonomy whose only observable effect would be `ExitCode(err)` mapping that then has to be overridden back to 1/2 everywhere. That is negative leverage.
- pkg/provider is the legitimate third-party home, and even there the fit is absent: provider failures become `finding.Finding`s with severity (provider.go:241-289), not process exits or HTTP responses — the two things errorfamily's boundary helpers produce. The one marginal idea (classifying "binary not on PATH" as Rejection vs "exec failed" as Transient for BuildFlow-side retry) adds a require for information BuildFlow does not currently act on.

Verdict: stdlib-only insulation is CORRECT here. The library solves server/service error-protocol problems; docs-gates' error problem (typed, golden-pinned refusals under erraudit) is already solved with less machinery, and the lint that enforces the discipline is itself where the library already sits in this repo's toolchain.

RECONSIDER-WHEN: erraudit's rules (or a fleet convention) start requiring errorfamily interface implementations (`Coded`/`Classified`) rather than mere typed carriers — then adopt first in `pkg/provider` (already carrying the indirect) and only consider root if the stack-drift contract is renegotiated fleet-wide; or when docs-gates grows a long-running/networked component that actually branches on error class.

## Ranked opportunities

None. (The only adjacent idea — errorfamily classification inside `pkg/provider`'s Detect/HealthCheck failure paths — was considered and rejected above: findings already carry severity, and no BuildFlow consumer branches on family.)

## How to verify

```console
$ git -C /home/lars/projects/go-error-family log -1 --format='%H'
01d155efa67f35ea75ced0942ab2c4b1b406bde3
$ git -C /home/lars/projects/go-error-family describe --tags
v0.12.0-13-g01d155e

$ rg -n "errorfamily|go-error-family" /home/lars/projects/docs-gates --type go -g '!docs/review/**'
# rc=1, no matches (zero API use, root + provider)

$ rg -n "go-error-family" /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
pkg/provider/go.mod:10:require github.com/larsartmann/go-error-family v0.12.0 // indirect
pkg/provider/go.sum:11:github.com/larsartmann/go-error-family v0.12.0 h1:HEtocvwaPe9udDusqDqV0pxN9Wx98KHQBK0MEHZupzg=
pkg/provider/go.sum:12:github.com/larsartmann/go-error-family v0.12.0/go.mod h1:gB43rKC69SoPilihuze43OFqKYMGQxDNUi/25PF9MgE=

$ rg -n "errorfamily\." /home/lars/projects/go-finding/errors.go   # type surfacing through go-finding
go-finding/errors.go:85:// ErrorCode implements errorfamily.Coded.
go-finding/errors.go:92:func (e *FindingError) ErrorFamily() errorfamily.Family {
$ cat /home/lars/projects/go-finding/go.mod | rg error-family       # why it is indirect
github.com/larsartmann/go-error-family v0.12.0
$ rg -n "go-error-family" /home/lars/projects/erraudit/go.mod       # flake input consumer
13:	github.com/larsartmann/go-error-family v0.12.0

$ rg -n "type \w+Error struct|func new\w+Error|func invalidf" /home/lars/projects/docs-gates/internal --type go | rg -v _test
# → the 11 carrier sites listed in "What docs-gates does today"
```
