# go-atomic-write utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v0.7.0 (latest local tag, fetched 2026-11-23) + HEAD 38dbb9debf0b38c39ea2d7e4b82eed2feca1d66a (`git describe` shows `v0.7.0-2-g38dbb9d` — HEAD is two commits past the tag). Single module `github.com/larsartmann/go-atomic-write`; runtime deps cespare/xxhash/v2 v2.3.0 + gofrs/flock v0.13.1 (go-atomic-write/go.mod:5-8).
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f baseline — root module stdlib-only (zero non-stdlib requires, enforced by scripts/stack-drift-check.sh:9) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 3 (I/O + runtime + output) subagent of the multi-library utilization audit fleet.
- Fleet stance addressed: "if all my projects re-implement everything, I will only fix bugs all day and none benefit from collective improvements" — does stdlib-only insulation here avoid real re-implementation debt, or leave docs-gates hand-rolling something the fleet solved?

## Verdict

NOT-ADOPTED — with a documented residual risk: the four guarded mutator writes are plain `os.WriteFile` with no temp+rename/fsync, and the library's core guarantees are exactly what those sites lack, but the stdlib-only root contract, the dry-run-default design, and git-as-recovery-layer make the residual risk small and already adjudicated in-repo (an io-surface judgment records `no-atomic-write` as a standing exception).

One nuance keeps this from being a clean NOT-APPLICABLE: unlike the fleet's retry/daemon/watcher libraries, go-atomic-write addresses a failure class (crash-torn writes and lost updates) that docs-gates' write sites genuinely ARE exposed to. The verdict below is therefore an evidence-backed acceptance of residual risk, not a "wrong domain" dismissal.

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

### The guarantees, precisely (and how they were earned)

The library's offering is not a naive temp+rename — its own history documents why each leg exists, which matters when judging whether a stdlib re-creation would be "equivalent":

- **Single-rename atomicity**: v0.2.0 REMOVED a `.bak` two-rename pattern whose intermediate window could show ENOENT or lose the target on a crash between renames — "the library named 'atomic-write' was not atomic" (go-atomic-write/CHANGELOG.md:248-252). Today the commit is one `os.Rename` (go-atomic-write/rename_unix.go:12).
- **Content durability**: the staged file is fsync'd before the rename (go-atomic-write/stage.go:88), and the target DIRECTORY is fsync'd after it, POSIX-only (go-atomic-write/rename_unix.go:17-33) — the pair that makes the post-crash file content, not just the name, durable.
- **Concurrent-writer isolation**: each writer stages through a `crypto/rand`-suffixed unique temp (go-atomic-write/stage.go:44-49), because a shared `path + ".tmp"` once let two writers interleave bytes in the staging file itself (same v0.2.0 section).
- **TOCTOU verification**: `WriteVerified` re-reads the target under an exclusive flock and compares the xxhash64 fingerprint taken at read time; mismatch returns the `ErrConcurrentModification` sentinel (go-atomic-write/commit.go:13, :36-60). A zero fingerprint means "must not exist yet" — concurrent creation is a conflict, not a silent skip (go-atomic-write/commit.go:21-34).
- **Failure hygiene**: a failed rename removes the staged temp so no `.tmp` residue sits next to the target (go-atomic-write/commit.go:67-76; the leak was fixed in v0.6.1, go-atomic-write/CHANGELOG.md:78).
- **Mode preservation**: the staged file is created with the existing file's own permission bits, 0644 for fresh files (go-atomic-write/stage.go:16-31).
- **Known residual (theirs)**: a hard crash still leaves an orphaned unique `.tmp` (unique names prevent interference but do not self-clean) — v0.2.0 Caveats, go-atomic-write/CHANGELOG.md:248.

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

Before the sites: one repo-hygiene nuance. go-atomic-write DOES appear in docs-gates' flake.lock — but only as a transitive input of the PRIVATE erraudit master-tracking flake (its own inputs list go-atomic-write pinned at v0.6.0, rev 2234e1d4), never as a direct input of docs-gates' flake.nix and never in either Go module graph. The erraudit-rev tripwire (scripts/check-erraudit-rev.sh) is what pins that rev; nothing in docs-gates' own code can see the library. So even the fleet's nix-level proximity does not make this a dependency of the audited consumer.

1. **internal/docsstrike/docsstrike.go:153** — `docs-gates strike --apply` rewrites the whole TODO_LIST after in-place row surgery. Mode hand-preserved at internal/docsstrike/docsstrike.go:146-150. Guarded: dry-run default (internal/docsstrike/docsstrike.go:137-138), exactly-one-row selector, double-strike refusal, and `scripts/safe-strike.sh` outside the binary.
2. **internal/docsarchive/docsarchive.go:229** — `writePreservingMode` (internal/docsarchive/docsarchive.go:222) prepends the archive banner to each report before `git mv` (internal/docsarchive/docsarchive.go:203). Guarded: `--apply` opt-in (internal/docsarchive/docsarchive.go:106-112); the file move itself rides git.
3. **internal/docspurge/docspurge_collect.go:192** — `writePreservingMode` (internal/docspurge/docspurge_collect.go:185) rewrites TODO_LIST then CHANGELOG from internal/docspurge/docspurge.go:199 and internal/docspurge/docspurge.go:205. Guarded: dry-run default; writes only on a SAFE draft; anchor refusal exits 2 with nothing written.
4. **internal/docsfreshness/preflight_land.go:58** — the land-parity stamp (gitignored local state, 0600); the write error is deliberately swallowed ("the stamp is a guard, never the gate").

There is no temp-file/rename/fsync logic anywhere in the consumer's non-test code (`rg -n "os.Rename|\.tmp|Sync\(\)"` over internal/ + cmd/ returns nothing non-test).

### The guard stack actually wrapped around each write

Every site stacks the same four guards in place of write-level atomicity:

1. **Dry-run default** — strike prints the old→new row pair and exits without touching disk until `--apply` (internal/docsstrike/docsstrike.go:137-138); purge prints the draft plan; archive prints its candidate list (internal/docsarchive/docsarchive.go:106-107). The dangerous flag is opt-in, never defaulted.
2. **Refusal-first semantics** — ambiguous selectors, double strikes, missing verdicts, missing `[Unreleased]`/`Changed` anchors all exit 2 with NOTHING written (the purge collector's `missingAnchorError` refuses to guess an insertion point, internal/docspurge/docspurge_collect.go:160-180).
3. **git as the atomicity layer** — every target file (TODO_LIST, CHANGELOG, reports) is git-tracked; the archive verb even moves files via `git mv` so history keeps the rename (internal/docsarchive/docsarchive.go:16-17, :203). The last commit is always a full recovery point.
4. **Ritual rails outside the binary** — `scripts/safe-strike.sh` (unique ≥4-letter anchors, dry-run default, one strike per commit boundary) plus the auto-commit daemon snapshotting the tree every ~2-3 minutes (AGENTS.md:55-59).

### The crash scenarios, honestly priced

- **Torn single write** (machine dies mid-`os.WriteFile`): TODO_LIST or CHANGELOG can lose both old and new content. Recovery: `git restore` from the last commit — unless the daemon already committed a prior edit of the same file, in which case that daemon commit is the recovery point. Cost: minutes of ritual redo, zero silent corruption (the next battery run surfaces any damage).
- **The two-write gap** (crash between internal/docspurge/docspurge.go:199 and :205): struck rows deleted from TODO_LIST, their CHANGELOG bullets never landed — silent ledger loss until a human notices. Same git recovery, but nothing flags it automatically. This is the one window this audit would name as the genuine residual.
- **Concurrent write** (the daemon's format/lint-at-write-time racing a `--apply`): the read-modify-write at internal/docspurge/docspurge.go:108-205 reads the file, computes the new bytes, and writes the whole file back; a daemon edit landing in between is silently clobbered. In practice the daemon formats rather than edits prose and AGENTS.md's staging-race note prices the inverse race, but the exposure is real and is exactly `WriteVerified`'s domain (go-atomic-write/commit.go:36-60).

## Gap analysis — should docs-gates adopt it?

**The prior adjudication.** This exact question was already judged in-repo: the 2026-11-06 io-surface pilot recorded `docsstrike.Run: exception "no-atomic-write", rationale = stdlib-only contract forbids go-atomic-write, blast radius one ledger row, guarded by dry-run default + scripts/safe-strike.sh` (docs/status/archived/2026-11-06_00-33_io-surface-pilot-evaluation-session.md:66-67), and the open question it left — "decision or deferral?" — is answered by architecture: the stdlib-only root contract is deliberate (AGENTS.md:34-37, scripts/stack-drift-check.sh:9), and go-atomic-write is a two-dependency library (xxhash + flock; gofrs/flock itself drags golang.org/x/sys), so root adoption would break zero-requires for a benefit the write pattern already fences three other ways.

**The honest residual risk.** Machine death mid `purge --apply` yields two distinct windows: (a) a torn file — `os.WriteFile` truncates then writes, so power loss mid-write can destroy both TODO_LIST and CHANGELOG content; (b) the two-write gap — TODO_LIST rewritten at internal/docspurge/docspurge.go:199 but CHANGELOG not yet at internal/docspurge/docspurge.go:205 means struck rows deleted without their CHANGELOG bullets. go-atomic-write closes (a) fully (rows 1-3) and shrinks (b) to a one-rename window. Weighed against the consumer's counterweights:

| Exposure | go-atomic-write would | docs-gates' existing answer | Net |
|---|---|---|---|
| Torn single write (crash mid-`WriteFile`) | close entirely (temp+fsync+rename) | `git restore` from last commit; battery re-run surfaces damage | recovery vs prevention — accepted |
| Two-write purge gap | shrink to one rename window | dry-run preview + SAFE-only writes + safe-strike's one-mutator-per-commit-boundary | residual accepted; ordering fix available (opportunity 3) |
| Concurrent writer (daemon vs `--apply`) | detect via fingerprint + flock | single-writer ritual; daemon formats rather than edits prose | accepted; the cheapest real hardening |
| Stamp file (untracked) | nothing to recover from git | gitignored by design; a torn stamp is a named usage error with the recovery path printed | non-issue by construction |

**The one genuine re-implementation slice.** Mode preservation is the library's offering the consumer hand-rolled — three copies of a `Stat`-then-`Perm`-fallback helper (inventory row 4). That is 3×7 duplicated lines inside one repo, not fleet-wide duplication of the library; the dedup fix is an internal `gatecli` helper, not a dependency. It does not move the adoption needle.

**pkg/provider check.** The provider writes nothing (read-only by contract, README.md:139); there is no write path there to harden.

RECONSIDER-WHEN: (1) any mutator grows an unguarded or multi-file write sequence that git cannot cheaply unwind — e.g. purge batching across repos, or a write to a NON-tracked file (the stamp today, anything user-configurable tomorrow); (2) a fleet decision ever relaxes stdlib-only for a vetted allowlist — go-atomic-write is then the first candidate for the four sites, ideally vendored-by-stdlib-reimplementation (temp+rename+fsync is ~40 lines) rather than the dependency itself; (3) a second fleet consumer adopts the purge ritual on NFS/FUSE where torn writes bite harder.

**Fleet-argument scorecard, per prong.** (a) "Re-implementation debt": one real slice found — mode preservation hand-rolled three times (row 4); it is 3×7 lines inside ONE repo, so the dedup is an internal `gatecli` helper, not the library. Everything else the library offers is NOT re-implemented — declined, not duplicated. (b) "Collective improvement": if go-atomic-write ever hardens (e.g. self-cleaning stale temps, its documented residual), docs-gates would not inherit it under NOT-ADOPTED; the price is bounded because the protected write surface is four sites in guarded rituals, revisited cheaply by re-running this audit.

## Ranked opportunities

1. **Deduplicate the three `writePreservingMode` copies into one `gatecli` helper** (internal/docspurge/docspurge_collect.go:185, internal/docsarchive/docsarchive.go:222, plus the inline variant at internal/docsstrike/docsstrike.go:146-153) — pure intra-repo cleanup, no dependency, kills the only real duplication found. LOW effort, no golden impact (behavior identical). This is also the honest answer to the fleet argument for this library: the shareable slice is 7 lines, not the mechanism.
2. **Document the torn-write residual next to the io-surface exception** (one line in AGENTS.md's deliberate-non-fixes list, alongside the Truncate entry at AGENTS.md:123): "plain os.WriteFile under --apply; git is the atomicity layer for tracked files; the two-write purge window is accepted." Makes the standing decision legible to the next auditor without re-deriving it.
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
