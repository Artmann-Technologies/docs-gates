# go-business-rules utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v2.2.1 (latest local tag, fetched 2026-11-23) + HEAD 75a1a8cfd2928d79aca7340f6cf018767ee35fef; root module path `github.com/LarsArtmann/go-business-rules/v2`. Submodules at HEAD: `adapters/cqrslite`, `listeners/otel`, `examples/sse` (untagged).
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 1 (resilience/error family) subagent of the multi-library utilization audit fleet.

## Verdict

NOT-ADOPTED — go-business-rules is a severity-aware rules engine over Go values, while docs-gates' "rules" are bespoke markdown-corpus scanners whose severity doctrine (WARN/FAIL/RATCHET/SKIP) is deliberately woven into byte-pinned output lines and per-repo JSON rollout knobs; the overlap is conceptual only, and encoding it into rule objects would move rollout policy out of the config file the fleet controls.

## Score

- Inputs N = 13 (inventory rows below).
- A = 3 (severity levels, result bucketing, config cross-field validation — problems docs-gates genuinely has, solved in-domain).
- M = 0.
- Score = round(100 × 0/3) = **0** — band **low**. Expected for NOT-ADOPTED; the verdict carries the meaning.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | Declarative `Rule`/`RuleImpl`/`ContextRule` model — named checks returning violations | No — gates are imperative corpus scans, not value predicates; a rule object adds nothing the `Run([]string) int` shape lacks | N/A | go-business-rules/rule.go:8,27,93 |
| 2 | ~25 builder validators over Go values (`NonNegative`, `InRange`, `NotEmpty`, `Matches`, `MinLength`…) | No — the only values docs-gates validates are a handful of config thresholds, checked with plain `if`s | N/A | go-business-rules/builders.go:82-321, builders_collection.go:30-99 |
| 3 | Severity levels — Info/Warning/Error/Critical (`Severity` = `finding.Severity`) | Yes in concept — docs-gates findings carry WARN/FAIL (and RATCHET/SKIP) severity | N/A (solved in-domain) — severity is spelled as line prefixes + exit-code selection, byte-pinned by goldens | go-business-rules/severity.go:6-22; consumer internal/docsagents/docsagents.go:36-40 (`severityWarn`/`severityFail`), internal/docsstub/docsstub.go:11 |
| 4 | `ValidatorBuilder` + `ValidationResultError` — `Errors()`/`Warnings()`/`Info()`/`Critical()`/`BySeverity()` bucket selection | Yes in concept — the battery bucket-selects PASS/FAIL/SKIP rows and aggregates to one exit code | N/A (solved in-domain) — a three-way switch over exit codes IS the whole model; the JSON face exposes it directly | go-business-rules/validator.go:15-22, validation_result.go:19-77; consumer internal/battery/battery.go:270-276 |
| 5 | Config cross-field validation (thresholds relate sensibly) | Yes — docs-gates validates warn-bar ≤ fail-bar per gate | N/A (solved in-domain) — five `invalidf` ifs, each producing a golden-pinned FATAL message | go-business-rules (rule composition generally); consumer internal/config/content_gates.go:59-88 |
| 6 | `ViolationError` typed violation values | No — gate findings are printed lines, not error values; gates return `int` | N/A | go-business-rules/errors.go:12-51 |
| 7 | Events + `Listener` (`RuleEvaluated`, `ValidationCompleted`) | No — no observability consumer; the printed matrix IS the report | N/A | go-business-rules/events.go:8-82 |
| 8 | Composites `All`/`Any`/`When`/`Not`, `OneOf` | No — no rule composition anywhere; gate order is a config list, evaluation independent by design | N/A | go-business-rules/builders_composite.go:64-116, :10 |
| 9 | Collection/generic validators (`NotEmptySlice`, `DivisibleBy`, generics) | No | N/A | go-business-rules/builders_collection.go:30,99 |
| 10 | Network-format builders (`IPAddress`, `CreditCard`, `PhoneNumber`, `PostalCode`) | No — nothing network-shaped is validated | N/A | go-business-rules/builders_network.go:23-144 |
| 11 | `adapters/cqrslite` — rule evaluation as CQRS command-guard | No — no command bus | N/A | go-business-rules/adapters/cqrslite/adapter.go:1 |
| 12 | `listeners/otel` — validation events → OTel spans/metrics | No — no OTel, no metrics | N/A | go-business-rules/listeners/otel/adapter.go:1 |
| 13 | Dependency convergence: requires `go-finding` v1.15.1 — the exact library+version `pkg/provider` already uses | Adjacent — if the provider ever needed severity-typed findings, `finding.Severity` is already in its graph WITHOUT go-business-rules | N/A | go-business-rules/go.mod:8; consumer pkg/provider/go.mod:5 |

## What docs-gates does today

Zero usage, zero presence in either module graph. Verified:

```console
$ rg -n "business-rules|businessrules" /home/lars/projects/docs-gates -g '!docs/review/**' --type go
# rc=1, no matches

$ rg -n "business-rules" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1, no matches
```

What docs-gates does instead — the rules-engine-shaped surface it hand-rolls (the duplication check the fleet argument demands):

- **Per-finding severity as printed verdict words**: `WARN %s:%d: … (vocabulary_enforce=false, warn mode)` — `internal/docsvocabulary/docsvocabulary.go:275`; `WARN %s cites %s — target does not exist …` — `internal/docscitations/docscitations.go:105`; severity consts `severityWarn`/`severityFail` — `internal/docsagents/docsagents.go:36-40`; stamps two-bar model (warn bar informational, fail bar exits 1) — `internal/docsstamps/docsstamps.go:8-15`.
- **The rollout doctrine** (the "docsstub doctrine"): five content gates share "findings WARN and exit 0 until `*_enforce` flips" — `internal/docsstub/docsstub.go:11,39,100,121`; knobs `stub_enforce`/`citation_enforce`/`stamp_enforce`/`vocabulary_enforce`/`agents_enforce` declared at `internal/config/known_keys.go:34-47` and `internal/config/config.go:112-117`. This is docs-gates' severity policy engine: repo-controlled JSON booleans, not code.
- **RATCHET as a tolerated-findings class** — green-but-named: `internal/docsfreshness/audit.go:73` ("RATCHET — %d unharvested tolerated (pinned --audit-max=%d)").
- **Verdict bucketing + aggregation** — `internal/battery/battery.go:37-39` (`verdictPass/Fail/Skip`), `:270-276` (bucket by exit code/skip list), exit arithmetic "0 all green, 1 any gate FAIL" (`:105`).
- **Config cross-field validation** — `internal/config/content_gates.go:59-88`: warn/fail bar ordering rules per gate, refusals via `invalidf` (typed carrier, `internal/config/errors.go:118-127`).

None of this evaluates `Rule` objects over Go values, composes rules, or emits events — go-business-rules' actual machinery. What overlaps is the *concept* of severity, and there the consumer's implementation is not a re-implementation of the library but a projection onto a different contract surface (printed lines, exit codes, JSON keys). AGENTS.md additionally records the family-wide disposition of genericizing per-gate shapes: "The settings-projection pair is ACCEPTED — domain-specific Settings; a generic helper would couple gates" (the five `//art-dupl:accept` baseline).

## Gap analysis — should docs-gates adopt it?

No. Weighing both sides:

- **Fleet argument**: the one thing docs-gates hand-rolls that go-business-rules names is severity-aware findings. But the library's severity lives in `ValidationResultError` buckets over rule objects; docs-gates' severity lives in byte-pinned output prefixes (`WARN …`), a frozen 0/1/2 exit contract, and JSON rollout knobs that fleet repos flip per-repo without touching the binary. Inserting an engine between config and output could only be judged correct if every golden stays byte-identical — the engine's entire observable surface would be "none", which is the definition of dead weight. The library solves *value validation with degrees of failure*; docs-gates' validation-shaped code is five config ifs, below any reasonable adoption threshold.
- **Root contract**: adoption breaks zero-requires (CI stack-drift), and drags go-finding + go-error-family transitively into the root.
- **pkg/provider**: the legitimate third-party home already shares go-finding v1.15.1 with this library, but runs no rule evaluation — it classifies pre-computed battery rows with a three-way switch (`provider.go:233-289`). Adding go-business-rules there would classify… one verdict string. (Note for any future attempt: the library README still says `GOEXPERIMENT=jsonv2` is required; the provider already builds `encoding/json/v2` on go 1.27 with no experiment set, so the constraint looks historical — re-verify at adopt time.)

The stdlib-only insulation here rightly avoids re-implementation debt: the library's problem (typed-value rule evaluation for application input) is not docs-gates' problem (corpus scanning with a printable contract), and the doctrine docs-gates does share with it is deliberately encoded where the fleet needs it — `.docs-gates.json`.

RECONSIDER-WHEN: a gate's logic becomes genuinely rule-shaped and user-authored — e.g., repos defining their own declarative rule tables (per-path regex rules, per-file threshold matrices) that docs-gates must evaluate generically. Then the evaluation engine belongs behind the config loader; if it needs go-business-rules, it goes in a separate module (provider-side or a new one), never the stdlib-only root — or the fleet renegotiates the root contract wholesale.

## Ranked opportunities

None. (The severity-word unification across gates — WARN/RATCHET prefixes — is an internal consistency topic for docs-gates itself, not an adoption of this library.)

## How to verify

```console
$ git -C /home/lars/projects/go-business-rules log -1 --format='%H'
75a1a8cfd2928d79aca7340f6cf018767ee35fef
$ git -C /home/lars/projects/go-business-rules tag --list | tail -2
v2.2.0
v2.2.1

$ rg -n "business-rules|businessrules" /home/lars/projects/docs-gates -g '!docs/review/**' --type go
# rc=1 (zero Go-file hits)
$ rg -n "business-rules" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 (not in any module graph)

# consumer's severity/rollout surface (spot-check):
$ rg -n "_enforce" /home/lars/projects/docs-gates/internal/config/known_keys.go
34: "stub_enforce": true, … 37: "citation_enforce": true, … 41: "stamp_enforce": true,
43: "vocabulary_enforce": true, … 47: "agents_enforce": true,
$ rg -n "severityWarn|verdictPass|RATCHET" /home/lars/projects/docs-gates/internal --type go | rg -v _test
internal/docsagents/docsagents.go:39: severityWarn = "warn"
internal/battery/battery.go:37-39: verdictPass/verdictFail/verdictSkip
internal/docsfreshness/audit.go:73: "…RATCHET — %d unharvested tolerated…"

# shared dependency (convergence note, row 13):
$ rg -n "go-finding" /home/lars/projects/go-business-rules/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod
go-business-rules/go.mod:8: github.com/larsartmann/go-finding v1.15.1
pkg/provider/go.mod:5: github.com/larsartmann/go-finding v1.15.1
```
