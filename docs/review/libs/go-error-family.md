# go-error-family utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v0.12.0 (latest local tag, fetched 2026-11-23) + HEAD 01d155efa67f35ea75ced0942ab2c4b1b406bde3 (`git describe` → `v0.12.0-13-g01d155e`). Submodules at HEAD: agent v0.2.5, bridge v0.3.4, diagnose v0.2.6, diagnose/git v0.5.4, diagnose/postgres v0.5.4, examples v0.3.2.
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f — root module stdlib-only (zero non-stdlib requires, `scripts/stack-drift-check.sh` in CI) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 1 (resilience/error family) subagent of the multi-library utilization audit fleet.
- Fleet stance addressed per library: "if all my projects re-implement everything, I will only fix bugs all day and none benefit from collective improvements" — does stdlib-only insulation here avoid real re-implementation debt, or leave docs-gates hand-rolling something the fleet solved?

## Verdict

NOT-ADOPTED — zero API use anywhere in docs-gates (root or pkg/provider); go-error-family rides in `pkg/provider/go.sum` only as a transitive indirect of go-finding, and the root module's hand-rolled typed refusal carriers are a deliberate, standing-decision alternative that keeps the stdlib-only contract.

Supporting facts, each verified this session:

- No consumer Go file names any errorfamily symbol (`rg` over `--type go`, rc=1 — see How to verify).
- The library's behavioral surface (retry decisions, HTTP statuses, BSD sysexits, What/Why/Fix rendering, live-system diagnosis) targets service applications; docs-gates is a single-shot exit-code CLI with byte-pinned output.
- The consumer's carrier pattern is enforced by erraudit in CI — and erraudit is itself built on go-error-family, so the fleet's error discipline already reaches this repo at the tooling layer.

## Score

- Inputs N = 16 (inventory rows below).
- A = 4 (rows whose underlying problem docs-gates actually has: typed error carriers, error interfaces, exit-code mapping, CLI boundary rendering — all solved differently by standing decision or contract).
- M = 0 (no row fully leveraged — the carrier slice is solved differently by standing decision, the rest is out of domain).
- Score = round(100 × 0/4) = **0** — band **low**. Expected for NOT-ADOPTED; the verdict carries the meaning, not the number.

- Applicable rows are exactly the inventory rows graded **Missed Opportunity** (the consumer has the underlying need; the capability is hand-rolled, solved in-domain, or declined by contract). Every **Not Applicable** row carries a one-line reason. `M` counts rows graded Fully Leveraged.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | `Family` behavioral classification — `Classify(err)` into six families (Rejection/Conflict/Transient/Corruption/Infrastructure/Orchestration); classification order multi-error → interface → sentinels → classifiers → default | No — every refusal is a terminal CLI exit, never a retry/tone/audience decision | N/A — no caller ever branches on error class; failures map to a fixed 3-code contract | go-error-family/family.go:27, go-error-family/classify.go; consumer exits via internal/gatecli/gatecli.go:29-32 |
| 2 | Multi-error support — `errors.Join` + Classify picks the worst Family deterministically | No — no joined errors in the consumer (each refusal is single-cause) | N/A | go-error-family/README.md "What It Gives You" |
| 3 | Six small interfaces (`Coded`, `Classified`, `Contextual`, `Retryable`, `ExitCoder`, `HTTPStatuser`) | Partially — consumer carriers implement only `Error()` (+ `Unwrap` where a cause exists) | Missed Opportunity (declined) — carriers are unexported and exist to satisfy erraudit's generic-error lint, not a cross-package protocol | go-error-family/interfaces.go:16-62; consumer internal/config/errors.go:13-23 |
| 4 | `Error` reference implementation + constructors (`NewRejection`, `NewTransient`, …), `WithContext`/`WithContextAny` | Yes — docs-gates has 11 typed refusal carriers; each is 3-8 lines, unexported, message spelled for a golden-pinned FATAL line | Missed Opportunity (declined: standing decision) — AGENTS.md erraudit-carrier doctrine | go-error-family/error.go, go-error-family/constructors.go |
| 5 | `Error.JSON()` — canonical JSON view for API boundaries | No — the battery's JSON face carries verdict rows, not error objects | N/A | go-error-family/README.md |
| 6 | `ExitCode(err)` — BSD sysexits.h mapping (1/64/66/71/75…) | Yes in principle — docs-gates maps every refusal to an exit code — but it actively conflicts: docs-gates pins 0/1/2 forever | Missed Opportunity (declined: conflicts) — adopting would break the exit-code contract and every wrapper test | go-error-family/family.go:203; consumer internal/gatecli/gatecli.go:24-32 |
| 7 | `HandleError(err)` CLI boundary — What/Why/Fix/WayOut structured rendering, message templates | Yes in principle — docs-gates renders FATAL lines at a CLI boundary | Missed Opportunity (declined) — output is byte-pinned by parity goldens; human text is a contract, not a library default | go-error-family/handle.go:121; consumer FATAL lines pinned by internal/battery/parity_test.go |
| 8 | `MessageTemplate` registry + `TemplateForCode` for HTTP/gRPC boundaries | No — no such boundary | N/A | go-error-family/handle.go:94,418 |
| 9 | `Registry` + `Clone()` + `RegisterClassifier` predicate classification | No — no dynamic third-party error taxonomy anywhere | N/A | go-error-family/registry.go:31-260 |
| 10 | `RegisterStdlibDefaults` (context/sql/os sentinel families) | No — the only stdlib errors handled (`*exec.ExitError`, `*json.SyntaxError`) are handled structurally, by type assertion | N/A | go-error-family/stdlib.go:30; consumer pkg/provider/provider.go:165, internal/config/errors.go:29 |
| 11 | `WrapOnce` — idempotent wrap against double-wrapping at API boundaries | No — no layered API boundaries; errors are created once at the refusal site | N/A — the failure mode does not exist in a single-binary CLI | go-error-family/wraponce_test.go |
| 12 | `LogError` / `LogErrorContext` (slog structured fields) | No — root module has no slog and no logging surface; diagnostics ARE the output | N/A | go-error-family/log.go:22-28 |
| 13 | `HTTPStatus(err)` / `HTTPHandler` net/http middleware | No — no net/http anywhere in the consumer | N/A | go-error-family/http.go:17,54 |
| 14 | `Family.RetryPolicy()` advisory retry defaults (loop not run) | No — no retry, no backoff, no time.Sleep in non-test code; timeouts are one-shot watchdogs | N/A — see the go-retry report | go-error-family/retry.go:18-26; consumer internal/battery/battery_watchdog.go:33 |
| 15 | `errorfamilytest` assertion helpers (`AssertFamily`, `AssertCode`, `AssertRetryable`, …) | No — root tests are stdlib-`testing` only, exercised through real `Run` functions | N/A — a test-only dep would still break the zero-require root go.mod | go-error-family/errorfamilytest/errorfamilytest.go |
| 16 | Transitive presence in `pkg/provider` (indirect via go-finding); `agent`/`bridge`/`diagnose`(+git,+postgres)/`examples` submodules — experimental analysis/bridge/diagnosis surfaces | No — nothing to leverage: the presence is inert (an indirect of go-finding), and the submodules answer needs docs-gates does not have (no diagnosis loop, no oops stack, no agent runtime) | N/A (correctly inert) — go-finding's `FindingError` implements `errorfamily.Coded`/`Classified`, but provider.go never touches `FindingError` or any errorfamily symbol | pkg/provider/go.mod:10, pkg/provider/go.sum:11-12; go-finding/errors.go:85-103 |

## What docs-gates does today

Zero direct usage — verified across the entire consumer including pkg/provider:

```console
$ rg -n "errorfamily|go-error-family" /home/lars/projects/docs-gates --type go -g '!docs/review/**'
# (no matches, rc=1)

$ rg -n "errorfamily|go-error-family" /home/lars/projects/docs-gates/pkg/provider --type go
# (no matches, rc=1)
```

The only occurrences in the whole consumer tree are dependency metadata, not code:

- `pkg/provider/go.mod:10` — `require github.com/larsartmann/go-error-family v0.12.0 // indirect`
- `pkg/provider/go.sum:11-12` — the two v0.12.0 hashes
- `flake.lock:184-198` — a `flake:false` source input pinned at ref v0.11.0 (rev 2d0679e0), feeding the erraudit tool build; erraudit's own go.mod requires go-error-family v0.12.0 (`/home/lars/projects/erraudit/go.mod:13`)

Why it is indirect: go-finding's root module requires go-error-family v0.12.0 directly
(`/home/lars/projects/go-finding/go.mod`), and its `FindingError` carries `ErrorCode() string`
(implements `errorfamily.Coded`) and `ErrorFamily() errorfamily.Family` (implements
`errorfamily.Classified`) — go-finding/errors.go:85-103. The errorfamily **type** `Family` is
therefore technically reachable through go-finding's exported API into pkg/provider's import
closure — but provider.go builds plain `finding.NewFinding` values (provider.go:241-289) and
returns plain `fmt.Errorf` errors (provider.go:230, 293); no errorfamily symbol is named.

What docs-gates does INSTEAD (the hand-rolled equivalent the fleet argument must beat) — typed
refusal carriers: unexported struct + constructor + optional `Unwrap`, one per refusal class,
message frozen for the FATAL line. Representative shape, `internal/config/errors.go:98-113`:

```go
type loadError struct {
    step  loadStep
    path  string
    cause error
}

func (e *loadError) Error() string { return fmt.Sprintf("%s %s: %v", e.step, e.path, e.cause) }
func (e *loadError) Unwrap() error { return e.cause }

func newLoadError(step loadStep, path string, cause error) *loadError {
    return &loadError{step: step, path: path, cause: cause}
}
```

The full carrier census (verified via `rg "type \w+Error struct|func new\w+Error|func invalidf" internal/`):

- `internal/config/errors.go:13-23` `syntaxError` (JSON syntax decorated with line/column, `Unwrap` keeps `*json.SyntaxError` reachable); `:98-113` `loadError`/`newLoadError` (+ `Unwrap`); `:118-127` `invalidError`/`invalidf` (no Unwrap by design — user-facing CLI text, no errors.Is consumers)
- `internal/gatecli/gatecli.go:69-76` `notDirectoryError`, `unknownFlagsError`; `:179-184` `valuedFlagError`
- `internal/docsrows/rows.go:87-89` `usageError`
- `internal/docspurge/parse.go:15-19` `purgeError`; `internal/docspurge/docspurge_collect.go:161-179` `missingAnchorError`/`newMissingAnchorError`
- `internal/docsstrike/strike.go:15-19` `strikeError`; `internal/docsstrike/parse.go:150-178` `strikeInputsError`/`newStrikeInputsError`
- `internal/docsfreshness/corpus.go:14-21` `candidatesError` (+ `Unwrap`)
- `internal/docsagents/docsagents.go:177-181` `directoryAgentsError`

Total: ~11 types across 7 files, each single-purpose. This is the pattern AGENTS.md names
("typed refusal carriers … satisfying erraudit while keeping idiomatic error returns") and it is
lint-enforced: `erraudit lint ./... --max` green since 2026-11-12, directives guarded by the
nolint-audit + erraudit-lint flake checks.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** The argument is strong when a project hand-rolls a *shared,
evolving* implementation. Test it against what docs-gates hand-rolls here:

- The carriers are not a re-implementation of go-error-family. They implement no families, no
  registry, no exit-code derivation, no retry decision — they are message structs. What they
  duplicate from go-error-family is only the `Error() string` + optional `Unwrap()` shape, which
  is stdlib `errors` itself, not library surface. There is no `file:line` where go-error-family's
  actual capabilities (Classify, ExitCode, HandleError, Registry, diagnose rules) are being
  re-implemented — those problems (behavioral branching on error class, sysexits exits,
  What/Why/Fix rendering, live-system diagnosis) are problems docs-gates does not have.
- A single-shot CLI with a frozen 0/1/2 exit contract and byte-pinned output has no consumer for
  a classification protocol: nothing downstream of a docs-gates error branches on anything but
  the exit code.

**Root-contract bar.** Adopting into the ROOT would break the zero-require contract (CI
stack-drift fails), force re-pinning every FATAL-line golden whose wording moved through
`HandleError` templates, and import a six-family taxonomy whose only observable effect would be
`ExitCode(err)` values that then have to be overridden back to 1/2 everywhere. Negative leverage.

**pkg/provider check.** The one legitimate third-party home — and even there the fit is absent:
provider failures become `finding.Finding` values with severity (provider.go:241-289), not process
exits or HTTP responses, the two things errorfamily's boundary helpers produce. The one marginal
idea (classifying "binary not on PATH" as Rejection vs "exec failed" as Transient for
BuildFlow-side retry) adds a require for information no BuildFlow consumer currently acts on.

**Ecosystem note.** The enforcement tool behind the consumer's carrier discipline — erraudit —
is itself built on go-error-family, and this repo's flake carries it as a pinned source input.
The fleet's error protocol already governs docs-gates' code shape without docs-gates importing
the protocol library. That is the "share the protocol, not the implementation" stance the
library's own README advocates, realized one level up.

Conclusion: stdlib-only insulation is CORRECT here — the library solves server/service
error-protocol problems docs-gates does not have, and the one shared problem (typed, golden-
pinned refusals under erraudit) is already solved with less machinery.

RECONSIDER-WHEN: erraudit's rules (or a fleet convention) start requiring errorfamily interface
implementations (`Coded`/`Classified`) rather than mere typed carriers — then adopt first in
`pkg/provider` (which already carries the indirect) and only consider root if the stack-drift
contract is renegotiated fleet-wide; or when docs-gates grows a long-running/networked component
that actually branches on error class.

## Ranked opportunities

None. (The adjacent idea — errorfamily classification inside pkg/provider's Detect/HealthCheck
failure paths — was considered and rejected above: findings already carry severity, and no
BuildFlow consumer branches on family.)

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
85:// ErrorCode implements errorfamily.Coded.
90:// ErrorFamily implements errorfamily.Classified, mapping ErrorCategory to
92:func (e *FindingError) ErrorFamily() errorfamily.Family {
95:		return errorfamily.Rejection
97:		return errorfamily.Conflict
99:		return errorfamily.Transient
101:		return errorfamily.Infrastructure
103:		return errorfamily.Transient
$ rg -n "go-error-family" /home/lars/projects/go-finding/go.mod
github.com/larsartmann/go-error-family v0.12.0
$ rg -n "go-error-family" /home/lars/projects/erraudit/go.mod      # flake input consumer
13:	github.com/larsartmann/go-error-family v0.12.0

$ rg -n "type \w+Error struct|func new\w+Error|func invalidf" /home/lars/projects/docs-gates/internal --type go | rg -v _test
/home/lars/projects/docs-gates/internal/docsagents/docsagents.go:177:type directoryAgentsError struct {
/home/lars/projects/docs-gates/internal/docsrows/rows.go:87:type usageError struct{ msg string }
/home/lars/projects/docs-gates/internal/docspurge/docspurge_collect.go:161:type missingAnchorError struct {
/home/lars/projects/docs-gates/internal/docspurge/docspurge_collect.go:179:func newMissingAnchorError(scope anchorScope, changelog string, nearestLine int) *missingAnchorError {
/home/lars/projects/docs-gates/internal/docspurge/parse.go:15:type purgeError struct {
/home/lars/projects/docs-gates/internal/docsstrike/strike.go:15:type strikeError struct {
/home/lars/projects/docs-gates/internal/docsstrike/parse.go:150:type strikeInputsError struct {
/home/lars/projects/docs-gates/internal/docsstrike/parse.go:178:func newStrikeInputsError(kind strikeRefusal, opener string) *strikeInputsError {
/home/lars/projects/docs-gates/internal/config/errors.go:13:type syntaxError struct {
/home/lars/projects/docs-gates/internal/config/errors.go:98:type loadError struct {
/home/lars/projects/docs-gates/internal/config/errors.go:111:func newLoadError(step loadStep, path string, cause error) *loadError {
/home/lars/projects/docs-gates/internal/config/errors.go:118:type invalidError struct {
/home/lars/projects/docs-gates/internal/config/errors.go:125:func invalidf(format string, args ...any) *invalidError {
/home/lars/projects/docs-gates/internal/docsfreshness/corpus.go:14:type candidatesError struct {
/home/lars/projects/docs-gates/internal/gatecli/gatecli.go:69:type notDirectoryError struct{ path string }
/home/lars/projects/docs-gates/internal/gatecli/gatecli.go:74:type unknownFlagsError struct{ flags []string }
/home/lars/projects/docs-gates/internal/gatecli/gatecli.go:179:type valuedFlagError struct {

$ sed -n 98,113p /home/lars/projects/docs-gates/internal/config/errors.go   # carrier shape excerpt
type loadError struct {
	step  loadStep
	path  string
	cause error
}
func (e *loadError) Error() string {
	return fmt.Sprintf("%s %s: %v", e.step, e.path, e.cause)
}
func (e *loadError) Unwrap() error { return e.cause }
// newLoadError builds one loadError with the FATAL-line message.
func newLoadError(step loadStep, path string, cause error) *loadError {
	return &loadError{step: step, path: path, cause: cause}
}
```
