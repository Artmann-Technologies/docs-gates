# go-atomic-write utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v0.7.0 (latest local tag, fetched 2026-11-23) + HEAD 38dbb9debf0b38c39ea2d7e4b82eed2feca1d66a (`git describe` shows `v0.7.0-2-g38dbb9d` — HEAD is two commits past the tag). Single module `github.com/larsartmann/go-atomic-write`; runtime deps cespare/xxhash/v2 v2.3.0 + gofrs/flock v0.13.1 (go-atomic-write/go.mod:5-8).
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f baseline — root module stdlib-only (zero non-stdlib requires, enforced by scripts/stack-drift-check.sh:9) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 3 (I/O + runtime + output) subagent of the multi-library utilization audit fleet.
- Fleet stance addressed: "if all my projects re-implement everything, I will only fix bugs all day and none benefit from collective improvements" — does stdlib-only insulation here avoid real re-implementation debt, or leave docs-gates hand-rolling something the fleet solved?

## Verdict

NOT-ADOPTED — with a documented residual risk: the four guarded mutator writes are plain `os.WriteFile` with no temp+rename/fsync, and the library's core guarantees are exactly what those sites lack, but the stdlib-only root contract, the dry-run-default design, and git-as-recovery-layer make the residual risk small and already adjudicated in-repo (an io-surface judgment records `no-atomic-write` as a standing exception).

## Score

- Inputs N = 10 (inventory rows below).
- A = 5 (crash-durable write, fsync-before-rename, TOCTOU/fingerprint verification, mode preservation, concurrent-writer lock — all address real exposure at the four write sites).
- M = 0.
- Score = **0** (round(100·0/5)) — band **low**. Expected for NOT-ADOPTED; the verdict and the residual-risk note are the signal, not the score.

## What the library offers

| # | Capability | Applies? | Status | Evidence |
|---|---|---|---|---|
| 1 | `Write(path, data)` — stage to unique temp file, fsync, single atomic `rename(2)` | Yes — the crash-torn-write class exists at all four consumer write sites | Declined (contract) | go-atomic-write/atomicwrite.go:17; go-atomic-write/stage.go:73 |
| 2 | fsync of the staged file before rename (crash durability of content) | Yes — `os.WriteFile` buffers only | Declined (contract) | go-atomic-write/stage.go:88 |
| 3 | Directory fsync after rename (durable rename, POSIX) | Yes in principle; the targets are git-tracked files whose real durability layer is git | Declined (contract) | go-atomic-write/rename_unix.go:17-33 |
| 4 | Mode preservation from the existing file (default 0644) | Yes — and the consumer DOES hand-roll this slice, three times | Hand-rolled locally | go-atomic-write/stage.go:16-31; internal/docspurge/docspurge_collect.go:185-196; internal/docsarchive/docsarchive.go:222-231; internal/docsstrike/docsstrike.go:146-153 |
| 5 | `WriteVerified` — xxhash64 fingerprint re-check under flock; `ErrConcurrentModification` sentinel | Yes — the auto-commit daemon and a human editor can race the read-modify-write (AGENTS.md: "The daemon RACES staging") | Declined (discipline instead) | go-atomic-write/atomicwrite.go:65; go-atomic-write/commit.go:20-60; internal/docspurge/docspurge.go:108-205 |
| 6 | `WriteIfChanged` — skip identical rewrites (no mtime bump) | Marginally — the pre-flight stamp rewrite (internal/docsfreshness/preflight_land.go:58) can rewrite identical bytes | N/A — one gitignored stamp, no observable benefit | go-atomic-write/atomicwrite.go:91-116 |
| 7 | Cross-process `flock` serialization of writers | Partially — see row 5; the guard is the one-writer ritual, not a lock | Declined (discipline instead) | go-atomic-write/commit.go:36-44 |
| 8 | `MkdirAll` of missing parent dirs (0755) on every write | No — consumers assume existing trees; archive creates its dir explicitly | N/A | go-atomic-write/stage.go:39-42; internal/docsarchive/docsarchive.go:195 |
| 9 | Streaming `WriteFunc`/`WriteFuncVerified` (64KB buffered callback) | No — payloads are single living-doc files held in memory already | N/A | go-atomic-write/atomicwrite.go:125-137 |
| 10 | Windows `MoveFileEx` retry on sharing violations | No — the fleet is Linux (Darwin deliberately absent from flake systems) | N/A | go-atomic-write/rename_windows.go (README platform table) |

## What docs-gates does today

Zero usage in either module graph. Verified:

```console
$ rg -n "go-atomic-write|atomicwrite|atomic-write" /home/lars/projects/docs-gates -g '!docs/review/**'
# hits only in: flake.lock (a TRANSITIVE pin under the erraudit flake input's own inputs,
# rev 2234e1d4, ref v0.6.0 — never consumed by docs-gates' own Go/flake code),
# docs/planning/2026-11-15_marker-grammar-and-tooling-split.md:76 (a pinned-repo survey row),
# docs/status/archived/2026-11-06_00-33_io-surface-pilot-evaluation-session.md:66-67,
# and scripts/citation-sweep.py (this fleet's scaffolding)
$ rg -n "atomic" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 — not in any module graph
```

The four non-test write sites, each verified open:

1. **internal/docsstrike/docsstrike.go:153** — `docs-gates strike --apply` rewrites the whole TODO_LIST after in-place row surgery. Mode hand-preserved at internal/docsstrike/docsstrike.go:146-150. Guarded: dry-run default (internal/docsstrike/docsstrike.go:137-138), exactly-one-row selector, double-strike refusal, and `scripts/safe-strike.sh` outside the binary.
2. **internal/docsarchive/docsarchive.go:229** — `writePreservingMode` (internal/docsarchive/docsarchive.go:222) prepends the archive banner to each report before `git mv` (internal/docsarchive/docsarchive.go:203). Guarded: `--apply` opt-in (internal/docsarchive/docsarchive.go:106-112); the file move itself rides git.
3. **internal/docspurge/docspurge_collect.go:192** — `writePreservingMode` (internal/docspurge/docspurge_collect.go:185) rewrites TODO_LIST then CHANGELOG from internal/docspurge/docspurge.go:199 and internal/docspurge/docspurge.go:205. Guarded: dry-run default; writes only on a SAFE draft; anchor refusal exits 2 with nothing written.
4. **internal/docsfreshness/preflight_land.go:58** — the land-parity stamp (gitignored local state, 0600); the write error is deliberately swallowed ("the stamp is a guard, never the gate").

There is no temp-file/rename/fsync logic anywhere in the consumer's non-test code (`rg -n "os.Rename|\.tmp|Sync\(\)"` over internal/ + cmd/ returns nothing non-test).

## Gap analysis — should docs-gates adopt it?

**The prior adjudication.** This exact question was already judged in-repo: the 2026-11-06 io-surface pilot recorded `docsstrike.Run: exception "no-atomic-write", rationale = stdlib-only contract forbids go-atomic-write, blast radius one ledger row, guarded by dry-run default + scripts/safe-strike.sh` (docs/status/archived/2026-11-06_00-33_io-surface-pilot-evaluation-session.md:66-67), and the open question it left — "decision or deferral?" — is answered by architecture: the stdlib-only root contract is deliberate (AGENTS.md:34-37, scripts/stack-drift-check.sh:9), and go-atomic-write is a two-dependency library (xxhash + flock; gofrs/flock itself drags golang.org/x/sys), so root adoption would break zero-requires for a benefit the write pattern already fences three other ways.

**The honest residual risk.** Machine death mid `purge --apply` yields two distinct windows: (a) a torn file — `os.WriteFile` truncates then writes, so power loss mid-write can destroy both TODO_LIST and CHANGELOG content; (b) the two-write gap — TODO_LIST rewritten at internal/docspurge/docspurge.go:199 but CHANGELOG not yet at internal/docspurge/docspurge.go:205 means struck rows deleted without their CHANGELOG bullets. go-atomic-write closes (a) fully (rows 1-3) and shrinks (b) to a one-rename window. The consumer's counterweights: every target is git-tracked and the last commit is the recovery point (`git restore --source=<good-sha>`); the writes are `--apply` opt-in behind dry-run defaults; the rituals run one mutator per commit boundary (safe-strike doctrine); and the auto-commit daemon snapshots the tree every ~2-3 minutes — though note the daemon cuts both ways: it can commit the half-applied two-write purge state (window b), making "restore from HEAD" recover a torn ritual. The blast radius is a docs ledger, never code or config, and the owner has already priced exactly this trade (the io-surface exception). This audit concurs with NOT-ADOPTED and asks only that the residual be stated where the exception lives.

**The one genuine re-implementation slice.** Mode preservation is the library's offering the consumer hand-rolled — three copies of a `Stat`-then-`Perm`-fallback helper (inventory row 4). That is 3×7 duplicated lines inside one repo, not fleet-wide duplication of the library; the dedup fix is an internal `gatecli` helper, not a dependency. It does not move the adoption needle.

**pkg/provider check.** The provider writes nothing (read-only by contract, README.md:139); there is no write path there to harden.

RECONSIDER-WHEN: (1) any mutator grows an unguarded or multi-file write sequence that git cannot cheaply unwind — e.g. purge batching across repos, or a write to a NON-tracked file (the stamp today, anything user-configurable tomorrow); (2) a fleet decision ever relaxes stdlib-only for a vetted allowlist — go-atomic-write is then the first candidate for the four sites, ideally vendored-by-stdlib-reimplementation (temp+rename+fsync is ~40 lines) rather than the dependency itself; (3) a second fleet consumer adopts the purge ritual on NFS/FUSE where torn writes bite harder.

## Ranked opportunities

1. **Deduplicate the three `writePreservingMode` copies into one `gatecli` helper** (internal/docspurge/docspurge_collect.go:185, internal/docsarchive/docsarchive.go:222, plus the inline variant at internal/docsstrike/docsstrike.go:146-153) — pure intra-repo cleanup, no dependency, kills the only real duplication found. LOW effort, no golden impact (behavior identical).
2. **Document the torn-write residual next to the io-surface exception** (one line in AGENTS.md's deliberate-non-fixes list, alongside the Truncate entry): "plain os.WriteFile under --apply; git is the atomicity layer for tracked files; the two-write purge window is accepted." Makes the standing decision legible to the next auditor without re-deriving it.
3. **Order-swap in purge apply** (CHANGELOG first, TODO_LIST second at internal/docspurge/docspurge.go:199-205): a crash then leaves a duplicate CHANGELOG bullet — detectable, re-purgeable — instead of silently deleted ledger rows. Zero dependencies, shrinks window (b) to the harmless side. Optional; goldens do not pin the write order.

## How to verify

```console
$ git -C /home/lars/projects/go-atomic-write log -1 --format='%H'
38dbb9debf0b38c39ea2d7e4b82eed2feca1d66a
$ git -C /home/lars/projects/go-atomic-write describe --tags
v0.7.0-2-g38dbb9d
$ git -C /home/lars/projects/go-atomic-write tag --list | tail -2
v0.6.2
v0.7.0

$ rg -n "go-atomic-write|atomicwrite|atomic-write" /home/lars/projects/docs-gates -g '!docs/review/**'
# flake.lock:24,134,140,147 (transitive under the erraudit input, ref v0.6.0);
# docs/planning/2026-11-15_marker-grammar-and-tooling-split.md:76;
# docs/status/archived/2026-11-06_00-33_io-surface-pilot-evaluation-session.md:66,67,185;
# scripts/citation-sweep.py:39 — zero Go-code or module-graph usage
$ rg -n "atomic" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1

$ rg -n "os\.WriteFile" /home/lars/projects/docs-gates --type go | rg -v _test
# internal/docsarchive/docsarchive.go:229, internal/docsstrike/docsstrike.go:153,
# internal/docspurge/docspurge_collect.go:192, internal/docsfreshness/preflight_land.go:58
# (all other hits are _test.go fixtures)
$ rg -n "os\.Rename|\.tmp|\.Sync\(\)" /home/lars/projects/docs-gates/internal /home/lars/projects/docs-gates/cmd --type go | rg -v _test
# rc=1 — no temp+rename/fsync logic anywhere non-test

$ sed -n '195,210p' /home/lars/projects/docs-gates/internal/docspurge/docspurge.go
# — TODO_LIST written at :199, CHANGELOG at :205: the two-write purge window
$ bash /home/lars/projects/docs-gates/scripts/stack-drift-check.sh
# stack-drift: root go.mod is dependency-free — OK
```
