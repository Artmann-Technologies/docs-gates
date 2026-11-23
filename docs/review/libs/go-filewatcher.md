# go-filewatcher utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v2.5.0 (latest local tag, fetched 2026-11-23) + HEAD 3c074addbe083d8061d47476e02867d3686c8b5b (`git describe` shows `v2.5.0-7-g3c074ad`). Single module `github.com/larsartmann/go-filewatcher/v2`; deps fsnotify v1.10.1, gogenfilter/v3, go-gitignore, x/text, x/time (go-filewatcher/go.mod:5-11); requires Go 1.26.7 + `GOEXPERIMENT=jsonv2` (README install note).
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f baseline — root module stdlib-only (scripts/stack-drift-check.sh:9) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 3 (I/O + runtime + output) subagent of the multi-library utilization audit fleet.

## Verdict

NOT-APPLICABLE — docs-gates is a point-in-time batch judge by doctrine (every verb runs, renders a pinned verdict, exits 0/1/2, and dies); it contains no file watching, no hot reload, and no event loop, and its own paid-for note says a watcher's premise — steady-state reactivity — is the opposite of its guarantee: "Gate verdicts are point-in-time under foreign daemons — timestamp the run; one run is not steady state" (AGENTS.md:188-189).

## Score

- Inputs N = 9 (inventory rows below).
- A = 0 (no row addresses a problem docs-gates has; the nearest adjacency — noticing when docs change — is a scheduling concern the fleet solves outside this binary: CI on push, BuildFlow detect, the auto-commit daemon).
- M = 0.
- Score = **0** (0/0; by convention 0 when nothing is applicable) — band **low**. For a NOT-APPLICABLE library the score is noise; the verdict is the signal.
- Roadmap check, for honesty: TODO_LIST/ROADMAP carry no watch/hot-reload/live-reload row at any priority (the word "watch" appears only for CI-run watching and the cadence gate's metaphorical "watches this file") — nothing planned is being blocked by this non-adoption.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | `New(paths, opts...)` — recursive fsnotify watch setup with path validation and walk | No — no path is ever watched; every verb takes a repo-root argument and scans once | N/A | go-filewatcher/watcher.go:237 |
| 2 | `Watch(ctx)` event loop returning a read-only event channel | No — the binary has no long-lived goroutine feeding a consumer loop; the only background goroutines are per-gate watchdogs that end with the gate | N/A | go-filewatcher/watcher.go:369; internal/gatecli/gatecli_capture.go:141 |
| 3 | 20 built-in filters + AND/OR/NOT composition (`WithExtensions`, `WithIgnoreDirs`, …) | No — corpus selection is one glob over `report_dir` plus config lists (`living_docs`), evaluated once per run | N/A | go-filewatcher/options.go:24-418; internal/docsfreshness/reports.go |
| 4 | Debouncing (global + per-path) | No — no repeated events exist to coalesce | N/A | go-filewatcher/options.go:24-46 |
| 5 | 19 middleware (logging, recovery, rate limiting, metrics) | No — gates fail fast with named evidence; recovery/rate-limit semantics would mask exactly the wedge class the battery exists to surface | N/A | go-filewatcher/middleware.go:39-123 |
| 6 | Resilience: ENOSPC watch-limit handling, circuit breaker + `Healthy()` | No — no inotify watches to exhaust | N/A | go-filewatcher/health.go:32; go-filewatcher/options.go:104 |
| 7 | Self-healing re-registration loop (`WithSelfHeal`) | No — nothing to re-register | N/A | go-filewatcher/watcher_selfheal.go:11 |
| 8 | NFS/FUSE polling fallback (`WithPolling`, `WithPollInterval`) | No — one-shot directory reads don't need change detection | N/A | go-filewatcher/options.go:209-232 |
| 9 | .gitignore-aware event skipping | No — ironically adjacent: the battery's one gitignore interaction is `git check-ignore` for the pre-flight stamp note (internal/docsfreshness/preflight_land.go:74), a one-shot exec, not a watcher rule | N/A | go-filewatcher/watcher_gitignore.go; internal/docsfreshness/preflight_land.go:74 |

### What the library is, in one paragraph

A production wrapper around fsnotify: `New` walks paths and registers watches (go-filewatcher/watcher.go:237), `Watch(ctx)` runs the event loop into a buffered channel (go-filewatcher/watcher.go:369), and every event passes a composable filter chain plus up to 19 middleware (logging, panic recovery, rate limiting) before delivery. Its differentiators are resilience legs the raw library lacks: ENOSPC watch-limit handling with a circuit breaker and a `Healthy()` probe (added v2.5.0, go-filewatcher/CHANGELOG.md:5), periodic re-registration of failed paths (go-filewatcher/watcher_selfheal.go:11-20), and an NFS/FUSE polling fallback (go-filewatcher/options.go:209-232). It is built for processes whose MAIN LOOP is event consumption — build tools, config reloaders, log shippers.

### The option surface, grouped (to show the full breadth being declined)

- **Delivery shaping**: `WithDebounce`/`WithPerPathDebounce` (coalesce editor-save bursts), `WithBuffer`/`WithErrorBufferSize` (channel sizing), `WithEventChannelMode` (go-filewatcher/options.go:24-36, :137-150, :311).
- **Event selection**: `WithExtensions`, `WithIgnoreDirs`, `WithIgnoreHidden`, `WithIgnorePatterns`, `WithFilter` composition, `WithGitignore` (go-filewatcher/options.go:46-72, :195, :349).
- **Resilience**: `WithCircuitBreaker`, `WithMaxWatches` (+ safety fraction), `WithSelfHeal`-driven retry loop, `WithPolling`/`WithPollInterval`, `WithCaseSensitivity` (go-filewatcher/options.go:104, :380-398, :209-232, :370).
- **Observability/error paths**: `WithMiddleware` (19 built-ins; recovery, logging, rate-limit at go-filewatcher/middleware.go:39-123), `WithOnError`/`WithErrorHandler`/`WithOnAdd`, `WithDebug` slog, `WithContentHashing` for change detection beyond mtime (go-filewatcher/options.go:118-173, :245, :418).

Every group presumes a live watcher instance held across time; every group is inert in a binary whose every execution path ends in `os.Exit`-equivalent within seconds.

## What docs-gates does today

Zero usage, zero presence in either module graph. Verified:

```console
$ rg -n "go-filewatcher|filewatcher|fsnotify" /home/lars/projects/docs-gates -g '!docs/review/**' --type go
# rc=1, no matches
$ rg -n "filewatcher|fsnotify" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1, no matches
```

(The only repo-tree mentions are flake.lock — a transitive pin under the private erraudit input's own dependency set — and this fleet's scripts/citation-sweep.py.)

There is also no hand-rolled watcher to deduplicate — the design has no equivalent to find:

- **No polling**: `rg -n "time\.Ticker" internal/ cmd/ pkg/ --type go` over non-test code returns nothing (the watchdogs use one-shot `time.NewTimer`; see the go-retry group-1 report for that distinction).
- **No event channels**: the only channel plumbing in the battery is `gatecli.CaptureOutput`'s concurrent drainers (the capture-pipe deadlock fix), which exist to END a gate's output stream, not to react to file events.
- **Re-runs are the caller's job**: CI runs the battery on push (AGENTS.md:218-223), BuildFlow's detect step execs it via pkg/provider/provider.go:149, and the dogfood ritual re-runs it after living-doc edits. "Timestamp the run; one run is not steady state" (AGENTS.md:188-189) is the explicit doctrine that a watcher would contradict.

### Debounce and middleware semantics, spelled out

The two legs a reactive consumer would actually use, so the N/A is grounded rather than assumed:

- **Debounce**: `WithDebounce` coalesces bursts globally; `WithPerPathDebounce` per path (go-filewatcher/options.go:24-36) — five editor events on one save become one delivery. docs-gates never receives an event, so there is nothing to coalesce; the closest analogue (a human re-running the battery twice) is handled by the battery being idempotent and cheap (~1.5 ms judge, README.md:70-72).
- **Middleware**: `Middleware func(Handler) Handler`, applied in REVERSE order — last added runs first (go-filewatcher/middleware.go:39-41) — with built-ins for logging (outermost), panic recovery (innermost), rate limiting, and a deduplicator tuned by event-count/cap heuristics (go-filewatcher/middleware.go:20-38). This is a serving-pipeline abstraction; the battery's cross-cutting concerns are exit-code mapping and stream policy, both settled in gatecli (internal/gatecli/gatecli_output.go:32-40).

### Why the doctrine excludes watching (not just the code)

The exclusion is architectural, not incidental. The binary's contract is "reads and judges a corpus; it never rewrites one" (README.md:65-66) — a judge whose verdicts are meaningful BECAUSE they are point-in-time snapshots of a corpus that foreign processes (editors, the auto-commit daemon, BuildFlow) mutate concurrently. A watcher-based mode would blur exactly that semantics: a verdict would no longer name a moment but a window, and the timestamp discipline (AGENTS.md:188-189) would lose its meaning. The battery even bounds each gate's wall clock (`gate_timeout_sec`) so that NOTHING inside a run waits on the world — the opposite posture of an event loop whose job is to wait on the world.

### The re-run triggers that exist today (the watcher's job, already assigned)

| Trigger | Who owns it | How docs-gates is invoked |
|---|---|---|
| Push to GitHub | CI dogfood job (AGENTS.md:218-223) | `go run ./cmd/docs-gates run .` |
| Local doc edit | the human's ritual (AGENTS.md:6-9) | `go run ./cmd/docs-gates run .` after touching living docs |
| BuildFlow quality loop | BuildFlow detect step | exec via pkg/provider/provider.go:149 |
| Commit time | pre-commit docs-battery leg (AGENTS.md:160-165) | pre-commit hook over all files |

Four independent re-run triggers, none owned by this repo's code — the scheduling layer is already built, fleet-side. Adding an in-process watcher would duplicate the fleet's scheduling, the exact anti-pattern the fleet argument warns about, in the one place it would be wrong.

### Toolchain and contract mismatch, for the record

Even setting the zero-requires contract aside, the shapes grind:

- **Toolchain**: go-filewatcher pins `go 1.26.7` and imports `encoding/json/v2` behind `GOEXPERIMENT=jsonv2` (its README install note; default under 1.27 where docs-gates lives).
- **Weight**: five direct requires (fsnotify, gogenfilter/v3, go-gitignore, x/text, x/time) for a consumer whose heaviest current import is `regexp`.
- **Maintenance class**: its own v2.4.x changelog is a history of case-folding/NFC path-key fixes — the surface you earn by holding filesystem state over time. docs-gates' reading of the filesystem is one glob and a handful of `os.ReadFile`/`os.Stat` calls per run, torn down at exit; it has no path keys to normalize because it holds no paths across time.

The dependency would import a stateful problem class into a stateless design.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** Both prongs score zero. docs-gates re-implements nothing go-filewatcher offers — there is no watch, no debounce, no filter chain, no event loop anywhere in the tree. And the problems go-filewatcher solves (long-lived reactive processes, editor-save coalescing, ENOSPC resilience, NFS fallback) are problems docs-gates deliberately refuses to have: it is a single-shot, deterministic, exit-code CLI whose outputs are byte-pinned by parity goldens (internal/battery/parity_test.go:188). A `docs-gates watch` mode would add a daemon with nothing to do — its whole value proposition is that a red gate today is the docs ritual due NOW, not an event to debounce.

**Root-contract bar.** Barred regardless: fsnotify + gogenfilter + x/time are hard requires, and the module even demands `GOEXPERIMENT=jsonv2` under its pinned toolchain. No pkg/provider target exists either — the provider is read-only by contract and execs the battery once per Detect.

RECONSIDER-WHEN: docs-gates ever grows a `--watch` developer mode (re-run the battery on living-doc save, locally only, human face only). That feature should live in a nested module or a wrapper binary that owns go-filewatcher and simply re-execs `docs-gates run` — never inside the root module, and never on the golden-pinned output paths.

**Fleet-argument scorecard, per prong.** (a) "Re-implementation debt": zero — there is no watch/debounce/filter/poll code in the tree to deduplicate (the sweeps above are the proof, not an assertion). (b) "Collective improvement": nothing to inherit — the library's improvements (self-heal, circuit breakers, case-aware path keys per its v2.4.0 changelog) benefit processes that watch; docs-gates would only start benefiting the day the flip condition lands, and the natural adopter then is the WRAPPER, not this binary. This is the insulation half of the fleet argument working as designed: the consumer's shape (batch, deterministic, byte-pinned) does not overlap the library's shape (reactive, long-lived, event-driven), so not adopting costs nothing.

## Ranked opportunities

None.

### Cross-references within this audit fleet

- go-atomic-write's `WriteIfChanged` advertises "no mtime bump, no file-watcher trigger" as a feature — the two libraries are fleet-complements for reactive tooling; docs-gates needs neither half.
- go-datastar (this group) is the other reactive-surface library; both share the same flip condition (a live dashboard/watch mode) and the same answer (adopt in the wrapper, not here).
- The go-retry report (group 1) covers the one-shot-timer vs loop distinction for the watchdogs mentioned above.

## How to verify

```console
$ git -C /home/lars/projects/go-filewatcher log -1 --format='%H'
3c074addbe083d8061d47476e02867d3686c8b5b
$ git -C /home/lars/projects/go-filewatcher describe --tags
v2.5.0-7-g3c074ad
$ git -C /home/lars/projects/go-filewatcher tag --list | tail -2
v2.4.3
v2.5.0
$ head -12 /home/lars/projects/go-filewatcher/go.mod
# module github.com/larsartmann/go-filewatcher/v2 — fsnotify v1.10.1, gogenfilter/v3,
# go-gitignore, x/text, x/time; go 1.26.7 + GOEXPERIMENT=jsonv2 per README

$ rg -n "go-filewatcher|filewatcher|fsnotify" /home/lars/projects/docs-gates -g '!docs/review/**' --type go
# rc=1 (zero Go-file hits)
$ rg -n "filewatcher|fsnotify" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 (not in any module graph)

$ rg -n "time\.Ticker|fsnotify|Watch\(" /home/lars/projects/docs-gates/internal /home/lars/projects/docs-gates/cmd /home/lars/projects/docs-gates/pkg --type go | rg -v _test
# rc=1 — no tickers, no watchers, no event loops in non-test code

$ sed -n '186,189p' /home/lars/projects/docs-gates/AGENTS.md
# — "Gate verdicts are point-in-time under foreign daemons — timestamp the
#    run; one run is not steady state."
$ sed -n '65,66p' /home/lars/projects/docs-gates/README.md
# — "docs-gates reads and judges a corpus; it never rewrites one."
$ rg -in "watch|hot.reload|live.reload" /home/lars/projects/docs-gates/ROADMAP.md /home/lars/projects/docs-gates/TODO_LIST.md
# only unrelated hits (CI-run watching; the cadence gate metaphorically "watches this
# file") — no reactive feature planned at any priority
$ git -C /home/lars/projects/go-filewatcher log -1 --format='%ci %h'
# fetched 2026-11-23; 0 behind upstream per the fleet manifest
```
