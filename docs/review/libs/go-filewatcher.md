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

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** Both prongs score zero. docs-gates re-implements nothing go-filewatcher offers — there is no watch, no debounce, no filter chain, no event loop anywhere in the tree. And the problems go-filewatcher solves (long-lived reactive processes, editor-save coalescing, ENOSPC resilience, NFS fallback) are problems docs-gates deliberately refuses to have: it is a single-shot, deterministic, exit-code CLI whose outputs are byte-pinned by parity goldens (internal/battery/parity_test.go:188). A `docs-gates watch` mode would add a daemon with nothing to do — its whole value proposition is that a red gate today is the docs ritual due NOW, not an event to debounce.

**Root-contract bar.** Barred regardless: fsnotify + gogenfilter + x/time are hard requires, and the module even demands `GOEXPERIMENT=jsonv2` under its pinned toolchain. No pkg/provider target exists either — the provider is read-only by contract and execs the battery once per Detect.

RECONSIDER-WHEN: docs-gates ever grows a `--watch` developer mode (re-run the battery on living-doc save, locally only, human face only). That feature should live in a nested module or a wrapper binary that owns go-filewatcher and simply re-execs `docs-gates run` — never inside the root module, and never on the golden-pinned output paths.

## Ranked opportunities

None.

## How to verify

```console
$ git -C /home/lars/projects/go-filewatcher log -1 --format='%H'
3c074addbe083d8061d47476e02867d3686c8b5b
$ git -C /home/lars/projects/go-filewatcher describe --tags
v2.5.0-7-g3c074ad
$ git -C /home/lars/projects/go-filewatcher tag --list | tail -2
v2.4.3
v2.5.0

$ rg -n "go-filewatcher|filewatcher|fsnotify" /home/lars/projects/docs-gates -g '!docs/review/**' --type go
# rc=1 (zero Go-file hits)
$ rg -n "filewatcher|fsnotify" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 (not in any module graph)

$ rg -n "time\.Ticker|fsnotify|Watch\(" /home/lars/projects/docs-gates/internal /home/lars/projects/docs-gates/cmd /home/lars/projects/docs-gates/pkg --type go | rg -v _test
# rc=1 — no tickers, no watchers, no event loops in non-test code

$ sed -n '186,189p' /home/lars/projects/docs-gates/AGENTS.md
# — "Gate verdicts are point-in-time under foreign daemons — timestamp the
#    run; one run is not steady state."
```
