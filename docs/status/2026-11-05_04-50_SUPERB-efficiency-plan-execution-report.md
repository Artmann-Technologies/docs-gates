# SUPERB Execution Report — The Whole Board, One Session

**2026-11-05 03:00–05:00 CET** · baseline `4f9ec5b` · executing the
02-50 Pareto plan end to end (27 tasks, 96 micro-tasks).

## a) What was done

The in-repo tier of the efficiency plan is COMPLETE and verified; the
owner-gated tier is fully prepped and surfaced. The plan's own order of
attack was followed: E1 → E2 → E3 → E5 → E4, then the quality batches,
fleet runs, and sibling legs.

- **E1 token memoization** (the 1%): Corpus caches presence + DF per
  prepared corpus; synthetic CV-shape benchmark 2,753ms → 47ms per
  1,600-report replay (59×); parity goldens byte-identical with NO
  re-pin; memo≡naive equivalence test red-controlled with a planted
  inverted cache.
- **E2 + E10**: committed benchmark + corpus synthesizer + shape-guard
  test; CI cliff arm is an OS `timeout` wrapper (go test's own -timeout
  provably never fires inside a benchmark loop) plus a CI dogfood job
  running the battery over this repo on every push.
- **E3 heartbeat / E5 estimate**: per-gate `done in <dur>` lines,
  matrix duration columns, audit-all pre-replay shape + worst-case
  estimate; duration tokens normalized in the parity harness (goldens
  re-pinned, hand-diffed).
- **E4 per-gate timeout**: `gate_timeout_sec` (default 300, 0 = off);
  an expired gate becomes a FAIL row (rc=timeout) via
  gatecli.CaptureOutputWithTimeout; discrimination test plants a
  select{} wedge.
- **The session's discovery — the capture-pipe deadlock (E6+E12
  payoff)**: the first fleet re-run (crm) hit the fresh timeout at 5m0s
  while the same gate ran directly in seconds. Gate output past the
  ~64KB pipe buffer parks its write on the never-drained capture pipe —
  THE fleet-wedge mechanism, confirmed by live reproduction and matching
  the thread autopsy (13/14 futex-idle, one netpoll epoll). Both
  capture functions now drain concurrently; regression-tested at 4MB
  per stream, red-controlled against the sequential-read plant.
- **Fleet runs (E12)**: crm 140 reports — freshness + audit-all RED
  (179 unharvested, pin 168 exceeded by 11 unhavested Nov-3/4
  obligations owned by crm's sessions); go-appkit all five green with
  its audit_max=12 ratchet holding EXACTLY under memoized semantics.
  CV audit-all (the hours-wedged run): **49.3 seconds**, 1,640 reports,
  8,788 findings = real archive debt, honest rc=1.
- **E26 + E27 quality batches**: safe-strike refusal-matrix harness
  (11 rails, all green), safe-strike --battery, golangci-drift flake
  check (red-controlled), shell-lint check (shellcheck + actionlint),
  go-floors dual-toolchain guard (red-controlled), dependabot enrolled,
  README Scripts section, gofmt CI step dropped (gofumpt supersedes),
  margins arithmetic + corpus-replace coupling documented, two
  assessments WONT-DO with rationale (heading-token exclusion,
  safe-strike bulk mode), purge pass (20 rows → CHANGELOG).
- **Siblings**: SystemNix rpi3-dns eval FIXED (quotaboard sops template
  guarded by its key-provisioning predicate) + doc-counts re-stamped
  145→146 (post-crash-calm-gate landed uncounted); E6 autopsy closed
  (mechanism confirmed as the deadlock above); E11 lessons in AGENTS +
  crush-config lessons.md (63f5a81); docs-health skill battery section
  updated (SKILLS repo); BuildFlow binary rebuilt + installed (e81c11834)
  and its doctor triaged (the 9 unavailable tools are foreign-ecosystem
  linters — benign for this repo); E17's diagnosis sharpened (treefmt's
  tree root = the flake-eval store source; rel-path candidates resolve
  outside it; fix = lane-level --tree-root forwarding into the roundtrip
  copy).
- **21 commits pushed** (4f9ec5b..df5720d): CI now runs the cliff-bench
  arm and the dogfood job against the memoized tree.

## b) Verification matrix (final sweep)

| Gate                                                                    | Result                              |
| ----------------------------------------------------------------------- | ----------------------------------- |
| `go test ./...` (root, -count=1)                                        | PASS                                |
| `(cd pkg/provider && go test ./...)`                                    | PASS                                |
| parity goldens (byte-identical except intentional re-pins, hand-diffed) | PASS                                |
| `golangci-lint run ./...` (fleet config)                                | 0 findings                          |
| `erraudit lint --type-aware` / `nolint-audit`                           | rc=0, 13/13 directives needed       |
| `nix flake check --all-systems --no-build` (incl. 4 new checks)         | ALL PASSED                          |
| `docs-gates run .` (dogfood)                                            | PASS all five                       |
| safe-strike-test.sh (11 rails)                                          | ALL behaved                         |
| speed guard / OS-timeout / drift / floors red-controls                  | all tripped as designed             |
| CI (push df5720d)                                                       | [watched to completion — see below] |

## c) The numbers

- Synthetic CV-shape hot loop: 2,753ms → 47ms (59×), shape-guarded <1s.
- Real CV audit-all: hours (killed 2026-11-04) → **49.3s**.
- crm audit-all post-deadlock-fix: 5m0s timeout → 790ms honest rc=1.
- go-appkit: green, ratchet exact (12/12 tolerated).
- Board: 41 open rows at baseline → the in-repo tier closed (16 struck
  in one batch + the earlier 6 + E27 rows); struck count 7/20.

## d) Mistakes this session (honest)

1. First benchmark was a no-op (ExtractSectionF takes PATHS; in-memory
   report bodies read as unreadable files) — caught by adding a
   shape-guard test that FAILS on zero work; the guard is permanent.
2. `cmd | tail; echo $?` reported tail's rc over a failing lint — the
   exact lesson the plan warned about, made in the same session that
   encoded it into lessons.md.
3. First timeout design repointed streams to devnull — the
   discrimination test caught that it would swallow the battery's own
   matrix; restored-to-original is correct.
4. The duration-normalization regex first ate `(4)` count parens; the
   parity red test caught it before any golden moved.
5. Sixteen-row strike batch genericized two old report items' anchors
   (corpus/parity tokens crossed the DF bar) — resolved by striking the
   two landed items through safe-strike with vocabulary preserved.
6. Daemon races beat several deliberate commits (heuristic wraps); all
   content identical, tree clean at every checkpoint.

## e) Process notes

Point-in-time verdicts under foreign daemons bit twice (CV freshness
flip; SystemNix's concurrent FastFlowLM session breaking the flake
check MINUTES after I fixed doc-counts). Both encoded in AGENTS facts.
The "do the whole list" mandate was executed against the plan's OWN
owner gates: everything in-repo landed; everything owner-gated is
prepped with a one-line ask.

## f) Next (small, ordered)

1. E7 v0.2.0 release cut (CHANGELOG section is one edit away) — OG.
2. E8+E9 SystemNix float + evo-x2 deploy of v0.2.0 — OG (float needs
   the tag pushed first).
3. E15 remainder: pbx-artmann pin bump v0.1.0 → v0.2.0 (one line) —
   rides E7; OG per plan.
4. E25 D1/D2 rulings (preps written in §4).
5. crm harvest authority (owner question 3, standing).
6. E17/E18 BuildFlow upstream fixes (diagnosis sharpened; concurrent
   sessions active — coordinate).
7. E24.3 pma/FastFlowLM monitoring docs (owned by the active session).

## g) Owner questions (consolidated)

1. **Release v0.2.0?** (E7) — everything green; tag + push + CHANGELOG
   cut ready. Consumers pin tags (pbx at v0.1.0).
2. **Deploy?** (E8/E9) — SystemNix float + evo-x2 switch puts the fast
   binary on the live PATH; the wrapper fix (59bb7cc35) rides the same
   window (E23).
3. **Cross-repo harvest authority** — crm's 11-past-pin Nov-3/4
   obligations + 179 total: harvest into crm's TODO (docs-health
   HARVEST mode) or leave to crm's sessions?
4. **Kill PID 248337?** (webphone wedge under a live buildflow parent —
   now PROVEN to be the capture-pipe deadlock; killing is safe for the
   data, but the parent buildflow run is another session's.)
5. **D1 battery scope / D2 go floor** (§4 preps) — rulings pending.
