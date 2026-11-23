# Status Report — CI-red recovery, archive ritual, lockfile restore (docs-gates)

> **Date:** 2026-11-23 02:39 CET · **Session:** resume of the eight-issue
> fix wave (`2026-11-23_00-17_eight-issue-fix-wave-pr7-merge.md`) under
> an explicit "finish the whole board" directive. The three owner
> questions left open by the previous session were decided
> autonomously on evidence (decisions recorded in §3).

**Repos touched:** `docs-gates` only. Everything pushed; CI green on
the final pushes.

## 1) What this session found on resume

The world had moved under the previous session's summary:

- The daemon had **pushed the "unpushed" train** — origin/main sat at
  `22e99a0`, not `f9965c4`.
- **CI was RED on main twice** (`f9965c4`, `22e99a0`): 9 golangci
  findings in the `go` job + the `dogfood` job failing on the cadence
  ARCHIVE-OVERDUE (16 unarchived vs threshold 15).
- **Four local daemon commits** carried a coherent tenth gate
  (`internal/docsagents`, AGENTS.md mtime age bars) authored by another
  session mid-outage — carried forward and verified, never reverted.
- The GitHub API had recovered; `gh` worked again.
- Issues #2–#11 were **already closed by the owner himself** at
  23:05Z (before the outage) — the pending W10d item was moot.

## 2) Fully done

| Work | Evidence |
| --- | --- |
| Diagnosed both red CI runs to the exact findings | `gh run view` 45081343809, 45086415501 |
| Fixed all 17 local golangci findings (drill 17 → 6 → 0) | `buildflow -s golangci-lint --format finding` total: 0 |
| `docsvocabulary.Scan` gocognit 47 → decomposed (`scanFile`/`statusUsage`/`statusColumnIndex`, `fileScan` struct) | parity goldens byte-UNCHANGED (pure refactor, `internal/battery` tests green) |
| `config.validateContentGates` cyclop 13 → four per-gate validators | `internal/config/content_gates.go` |
| `docsstrike.PlanStrike` cyclop 15 → `uniqueRowMatching` + `applyStrike` | `internal/docsstrike/strike.go`, tests green |
| godox/goconst/mnd/prealloc/unparam/varnamelen/wsl fixes | classify.go comment reword; `severityWarn`/`severityFail`; `nearMissWidth`; `writeEnforcedConfig`; routing `settings` rename |
| flake.lock erraudit damage (3rd occurrence) — tripwire caught rev 8831ce2 (also broke devShell build); restored pin dccfd10 from `fb25c41` | commit `2cebb40`, `scripts/check-erraudit-rev.sh` PASS |
| Archive ritual: audit-all proof (31 reports, 0 unharvested), per-report dump-items counts, 8 × 11-05/11-06 reports bannered + `git mv` to archived/ | commit `b0987e7` (banners amended in after the pre-commit-stash trap) |
| Battery over this repo: **10/10 PASS** (freshness, audit-all, annotate, routing, cadence, stub, citations, stamps, vocabulary, agents) | `go run ./cmd/docs-gates run .` rc=0, twice |
| Full local verification: root + provider suites, `go vet`, erraudit rc=0, nolint-audit 17 needed/0 stale, `nix flake check --all-systems --no-build` | all green |
| Pushed everything; **CI SUCCESS** on `b0987e7` and `443a065` | runs 45093982368 + final watch |
| AGENTS.md memory: 11-23 lockfile recurrence + 2 new traps (buildflow format ≠ golangci fmt coverage; pre-commit stashes unstaged edits) | commit `443a065` |
| nextd queue sync | zero docs-gates items tracked; the 8 closures registered as drift-closed |

## 3) The three owner questions — decided on evidence

1. **Push or hold** → archive-first-then-push: the ritual landed before
   the push, so origin's dogfood job ran green (it had been the red
   leg).
2. **Archive scope** → all eight 11-05/11-06 era reports (not just the
   minimum one): audit-all proved 0 unharvested corpus-wide; 8
   unarchived remain (comfortable headroom under 15).
3. **Closure voice** → moot: the owner had already closed all eight
   himself at 23:05Z with no comments; adding mine would be noise
   against his own closures.

## 4) Partially done / not started

- The previous report's optional follow-ups remain untouched:
  go-taskqueue `annotate_corpus` adoption, `vocabulary_enforce` fleet
  flip, go-verdict OUT_OF_SCOPE prose relocation, an archive-ritual
  helper verb, the release train. Owner-decision or separate work
  items, none blocking.
- The agents gate's CHANGELOG/FEATURES coverage was committed by the
  daemon in the other session's sweep (`c0f42e8` touched CHANGELOG) —
  not independently verified this session.

## 5) What I could have done better (honest ledger)

- **Banner staging trap**: committed the archive renames while the
  banners sat unstaged; pre-commit stashed them, so the first ritual
  commit was pure renames. Caught by reading the commit stat (0
  insertions), fixed with an amend before push. Should have staged
  everything first.
- **Wrong-run watch**: the first `gh run watch` grabbed the OLD failed
  run because I queried the run id before the new push registered.
  Should have matched headSha first.
- **grep-before-format**: counted "open items" with `grep -c '^OPEN'`
  against dump-items output that uses a different prefix — the zero
  counts were a format mismatch, not evidence. Caught by printing raw
  output; should have inspected the format before trusting a filter.
- **Pieced verification instead of one canonical loop**: build/vet/
  tests/lint-drill/erraudit/nolint/flake-check/battery were each run
  green, but a final single `buildflow --build-mode dev` pass after ALL
  changes was not — the legs covered it, the canonical loop would have
  proven it in one document.
- LSP ghost diagnostics (phantom `VocabularyFiles undefined` typecheck
  against a green build) persisted all session; correctly ignored per
  AGENTS.md, but the wrapper caches were never cleared — the ghosts
  will haunt the next session too.
- Two minor tool slips (`job_output` without `wait`; a redundant
  early `git status` read) — noise, no damage.

## 6) Nothing totally fucked up

No destructive event: no reverts, no resets, no lost work. The nearest
miss was the banner-staging gap above — caught pre-push by stat-reading
the commit. The lockfile damage was recovered forward per the
documented ritual, tripwire green before and after.

## 7) Next up (candidate board, owner-filtered)

- Verify the agents gate's CHANGELOG/FEATURES/README rows are complete
  and honest (daemon-committed, unreviewed).
- Release train: cut v0.3.0 (eight-issue wave + vocabulary + agents
  gates, parity re-pins).
- Decide the enforce flips for this repo: `vocabulary_enforce`,
  `agents_enforce`.
- go-taskqueue: adopt the `annotate_corpus` knob (draft exists).
- go-verdict: relocate the OUT_OF_SCOPE prose so the vocabulary gate
  goes green under enforcement.
- Archive-ritual helper verb candidate: banner + git mv as one gated
  command (dry-run default, like purge).
- BuildFlow upstream: the lockfile-rewrite repair damage is now a
  3-time recurrence (11-06, 11-14, 11-23) — file it upstream
  (verify-before-filing first: reproduce from the run logs).
- erraudit upstream: vendorHash-staleness incident row (05-43 report
  remainder, TODO_LIST §3).
- Pre-commit leg candidate: fire `check-erraudit-rev.sh` whenever
  flake.lock is staged (mechanical guard for the recurrence class).
- Clear the LSP wrapper caches so the typecheck ghosts stop seeding
  distrust in future sessions.
- TODO_LIST §9 (agents-gate harvest): strike the rows this session's
  verification completed.
- Cross-repo legs still open in TODO_LIST §3: SystemNix float, pbx
  pin bump, BuildFlow binary refresh.
