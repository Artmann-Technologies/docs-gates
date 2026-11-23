# go-cqrs-lite utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: HEAD 35856a709c8de00eb2ad447aa93c4fd073ea114e (no root version tags; 105 nested `go.mod` files — the surface is ~100 independently-versioned v4 submodules plus `core/v5`; latest meaningful submodule tags per the fleet manifest: system/v4 v4.19.0, metaengine/v4 v4.26.0, catalog/v4 v4.15.0, cmd/cqrs-lint/v4 v4.22.2, metaengine/sqliteengine/v4 v4.9.1; fetched 2026-11-23). Note: the manifest calls the root go.mod "literally module name `library`" — the checkout's root go.mod is in fact `module github.com/larsartmann/go-cqrs-lite` (go-cqrs-lite/go.mod:2); audited as found.
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f baseline — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 5 (UI + CQRS/application frameworks) subagent of the multi-library utilization audit fleet.

## Verdict

NOT-APPLICABLE — go-cqrs-lite is an event-sourcing/CQRS engine family (event journal, deciders, projection host with replay/checkpoints/dead-letter, cost-based query planner), and docs-gates is a stateless exit-code battery: the orchestrator's own bar ("an exit-code markdown battery gains nothing from event sourcing") holds exactly, and the honest duplication hunt below finds conceptual rhyme in the consumer's replays — but no shared mechanism.

## Score

- Inputs N = 10 (inventory rows below; the ~100 submodules grouped by concern).
- A = 1 (the replay/projection concept row — row 10; the only place the fleet argument could bite).
- M = 0.
- Score = round(100 × 0/1) = **0** — band **low**. Expected for NOT-APPLICABLE; the gap analysis carries the real content.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | `system` composition root: `system.New(ctx, DomainConfig, DeploymentConfig)` auto-wires store, bus, projections, queries, snapshots, timers, health, lifecycle | No — the consumer's only composition is a 24-line static gate table built from JSON config (internal/battery/battery.go:58-81) | N/A | go-cqrs-lite/README.md:54, go-cqrs-lite/README.md:309 |
| 2 | Event journal: immutable, versioned events; optimistic concurrency; time-travel queries | No — reports are immutable by ritual (annotate, never rewrite), but they are read as files, never journaled, versioned, or folded | N/A | go-cqrs-lite/README.md:34-41 |
| 3 | Decider pattern (load · fold · decide), typed command/event handler signatures | No — no commands; gates are argv → verdict batch functions | N/A | go-cqrs-lite/README.md:34-41 |
| 4 | Projection host: replay, crash-restart, checkpoints, dead-letter | Rhyme only — see row 10 and gap analysis; the consumer's "replay" is memoized batch scanning with no persistence, no catch-up, no subscriptions | N/A | go-cqrs-lite/README.md:39-40 |
| 5 | `metaengine` cost-based planner: routes queries across 11 interchangeable engines by measured cost | No — one engine (substring scan over a []string corpus); no engine choice exists | N/A | go-cqrs-lite/metaengine/v4/README.md:1-6 |
| 6 | Production primitives: event signing (HMAC/Ed25519), payload encryption, OTel tracing/metrics | No — no events to sign, no telemetry surface (exit code is the signal) | N/A | go-cqrs-lite/README.md:65-67 |
| 7 | Honest 6-family error taxonomy with sentinels and `%w` | No — the consumer's contract is exactly three exit codes plus typed refusal carriers it already has (AGENTS.md:38-39, :63-66) | N/A | go-cqrs-lite/README.md:68-69 |
| 8 | Long tail (grouped): ~90 domain/plumbing submodules — event, command, query, decider, dispatcher, dedup, idempotency, kv, graph, scheduling, engines (sqlite/pebble/bbolt/postgres/mysql/duckdb/...), cmd/cqrs-lint + generators, catalog, benchkit | No — every one presumes a running event-sourced application | N/A | go-cqrs-lite/CHANGELOG.md:28 (Unreleased sweep across the family) |
| 9 | `cmd/cqrs-lint`, module-graph discipline, multi-module isolation itself | No — the consumer enforces its own (stricter) zero-dep discipline via scripts/stack-drift-check.sh:8-13 | N/A | go-cqrs-lite/go.mod:2 |
| 10 | Concept: deterministic REPLAY over an immutable corpus to derive current state (projections/read models), so reads never mutate the log | **Concept applies** — this is the shape of audit-all/pre-flight; judged below as rhyme, not duplication | N/A (concept satisfied in-library by the consumer's own memoized scan) | go-cqrs-lite/README.md:34-41; internal/docsfreshness/corpus.go:81-96 |

## What docs-gates does today

Zero usage. Verified:

```console
$ rg -n "larsartmann/go-cqrs-lite" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**'
# rc=1, no matches — not in the root go.mod (no requires exist), not in
# pkg/provider/go.mod:5-10 (go-finding, toolsdk, go-error-family only)
```

Bare-name hits outside docs/review are exactly two: the audit scaffolding list
(scripts/citation-sweep.py:49) and a status-report note that the race
build-tag idiom (`RaceEnabled`) was contributed FROM go-cqrs-lite INTO
go-finding/toolsdk — the dep pkg/provider already uses
(docs/status/2026-11-14_18-00_whole-board-second-batch-ux-external-release.md:23).
The consumer itself uses that same idiom in its freshness tests
(internal/docsfreshness/race_off_test.go:1-8) — a fleet-shared testing
convention flowing through docs-gates' existing dep graph, not a go-cqrs-lite
adoption.

## Gap analysis — should docs-gates adopt it?

The orchestrator asked for an honest ruling on four specific rhymes. Rulings,
with the exact consumer code:

**1. Memoized replays (one scan per DISTINCT token) vs projections/read
models — RHYME, NOT DUPLICATION.** The consumer's `Corpus` memoizes two pure
functions per prepared corpus: `Contains` (token → present) and `Count`
(token → raw-line document frequency) over the living-docs lines
(internal/docsfreshness/corpus.go:81-96, :103-140). This is what took the CV
audit-all cliff from hours to 49s — one scan per DISTINCT token instead of one
per (token × report) (AGENTS.md:186-187). A go-cqrs-lite projection is a
different machine: a persisted, checkpointed, incrementally-caught-up read
model over an append-only event journal, restarted on crash, with dead-letter
quarantine (go-cqrs-lite/README.md:39-40). The consumer's caches are
per-process, rebuilt from scratch every run, keyed by token, and deliberately
unsynchronized because the gate path is single-goroutine by contract
(internal/docsfreshness/corpus.go:77-80). Cross-run persistence would buy
nothing: the living docs change between runs, so any checkpoint invalidates
wholesale; the corpus is tiny (hundreds of lines); and correctness is pinned
byte-for-byte by parity goldens, which a checkpoint-resume path could only
endanger. No file:line of the consumer re-implements event sourcing — there is
no journal, no fold, no subscription, no rebuild trigger.

**2. Ratchet arithmetic — NO OVERLAP AT ALL.** The adoption ratchet
(`audit_max`/`annotate_max`: tolerated findings print RATCHET and stay green;
REGRESSION past the pin fails — internal/docsannotate/docsannotate.go:229-239)
is a debt-drain migration pattern. go-cqrs-lite has no analogue; the fleet's
ratchet doctrine originated on the docs-gates side and flows outward (the crm
and go-appkit pilots pin their counts against it, TODO_LIST.md:30).

**3. Watchdog/timeout capture vs projection-host supervision — RHYME ONLY.**
The battery bounds each gate with a wall-clock cap: an expired gate becomes a
FAIL row and the battery moves on, with a half-budget early-warning goroutine
(internal/gatecli/gatecli_capture.go:121,
internal/battery/battery_watchdog.go:25-48). The projection host
crash-restarts workers and dead-letters poison events. Shared idea (a wedged
unit must not hang or mask the rest — AGENTS.md:70-72); utterly different
mechanisms (pipe-drain abandonment vs supervised event consumption). Nothing
to retire.

**4. "Projection-like report scans" — TWO SENSES, NEITHER CQRS.** (a) The
DOMAIN_LANGUAGE "settings-projection" (docs/DOMAIN_LANGUAGE.md:43) names
config→struct field mapping with zero-value fallbacks — a projection in the
linear-algebra sense, not a read model. (b) The freshness harvest is genuinely
projection-shaped in concept: it derives "which §f items do the living docs
reflect" by replaying the report corpus against current living-doc state,
point-in-time, never mutating reports (docs/DOMAIN_LANGUAGE.md:28-29,
internal/docsfreshness/check.go:38-57). But it is a batch scan with substring
token matching (internal/docsfreshness/check.go:61-94), run on demand, state
discarded at exit. Conceptual rhyme — deterministically re-deriving a view
from an immutable log — honestly acknowledged, and honestly not duplication:
making it an actual projection would ADD a store, checkpoints, and a daemon to
a battery whose whole value is being none of those things.

**Root-contract bar.** Any adoption (even `event` alone, the lightest module)
breaks zero-requires on the first `go get` — the `event` module's only
third-party dep is CBOR, but "only" is already "one too many" for
scripts/stack-drift-check.sh. And pkg/provider has no use for it: its job is
one exec plus a strict decode (pkg/provider/provider.go:148-171).

**What docs-gates would have to become.** A stateful, long-running application
whose findings derive from an append-only event log consumed by multiple read
models — i.e., a docs-event pipeline with subscriptions, catch-up UIs, and
time-travel over verdict history. That is a real product idea (a fleet verdict
timeline), and if it is ever built, the fleet doctrine says build it on
`system.New` + metaengine rather than hand-rolling stores and projections —
in a NEW repo/module, exec'ing this binary the way pkg/provider does.

RECONSIDER-WHEN: a verdict-history/timeline product grows out of docs-gates —
multiple consumers needing derived views over accumulated battery runs
(audit-trend dashboards, delta queries between runs). Until then the
insulation is correct: the battery's replay semantics are 60 lines of memoized
pure functions, pinned by goldens, with no infrastructure to share.

## Ranked opportunities

None. (The RaceEnabled build-tag idiom already flows into the consumer's dep
graph via go-finding/toolsdk — that fleet benefit is banked without any
import.)

## How to verify

```console
$ git -C /home/lars/projects/go-cqrs-lite log -1 --format='%H'
35856a709c8de00eb2ad447aa93c4fd073ea114e
$ find /home/lars/projects/go-cqrs-lite -name go.mod -not -path '*/.git/*' | wc -l
105
$ head -2 /home/lars/projects/go-cqrs-lite/go.mod
//cqrs-lint:ignore(E003) library code or intentional pattern
module github.com/larsartmann/go-cqrs-lite
$ git -C /home/lars/projects/go-cqrs-lite tag --list 'system/v4/v4*' | sort -V | tail -1
system/v4/v4.19.0

$ rg -n "larsartmann/go-cqrs-lite" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**'
# rc=1 — zero module-path hits
$ rg -n "go-cqrs-lite" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**' -g '!docs/planning/**'
scripts/citation-sweep.py:49:    "go-cqrs-lite",
docs/status/2026-11-14_18-00_whole-board-second-batch-ux-external-release.md:23:- **go-finding/toolsdk**: the race build-tag idiom (go-cqrs-lite's canonical `RaceEnabled`) contributed as `race_on.go`/`race_off.go` — fleet providers stop copying it locally; green under both `-race` and without. Pushed 8562d6b.
# scaffolding list + one fleet-flow status note; zero imports

# the rhyme, consumer side (memoized per-distinct-token replay — NOT a projection):
$ sed -n '81,86p' /home/lars/projects/docs-gates/internal/docsfreshness/corpus.go
type Corpus struct {
	lines    []string
	squashed []string
	present  map[string]bool // token -> Contains verdict (memoized)
	df       map[string]int  // token -> raw-line document frequency (memoized)
$ rg -n "memoized replays" /home/lars/projects/docs-gates/AGENTS.md
187:  | wc -l`; memoized replays: one scan per DISTINCT token, hours → 49s).
$ rg -n "RATCHET" /home/lars/projects/docs-gates/internal/docsannotate/docsannotate.go | head -2
229:// adoption ratchet: tolerated pins print RATCHET and stay green, counts
239:			"docs-annotate: RATCHET — %d finding(s) tolerated (pinned --annotate-max=%d); red only past %d\n",
```
