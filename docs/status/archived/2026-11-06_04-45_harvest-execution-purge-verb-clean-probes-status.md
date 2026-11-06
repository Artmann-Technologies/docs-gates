> **Superseded within the hour** by the full session close
> (2026-11-06_05-01_harvest-purge-verb-clean-probes-full-close.md), which
> carries this report's §h harvest ledger forward as its own §h and adds
> the final verification sweep. Archived by the cadence ritual (the
> corpus crossed max_unarchived=15); nothing in here is unrecoverable.

# Status Report — Whole-TODO-Board Harvest & Execution: Purge Verb, Gated Trains, Clean-Build io-surface Probes

**Report:** 2026-11-06 (session ~02:00–04:45 CET) · **Repo:** docs-gates @ `cb870f7` (main, clean tree) · Origin: pushed to `cb870f7` minus the closing trains (see scope line)

**Scope line — what this session VERIFIED:** every claim below rode a live
command: dogfood battery 5/5 PASS before EVERY docs commit (`c7c0a4d`,
`09be419`, `b222ef9`; the `cb870f7` purge commit ran the battery after —
its content was gated by the purge verb's OWN internal pre-flight SAFE +
cadence, noted in §d); root suite `go test -count=1 ./...` 9/9 packages;
provider suite green from its own directory; golangci 0 issues on the
touched packages (docspurge/cmd/docsfreshness/gatecli/battery) and 0 on
the full sweep at session end; the safe-strike harness 13/13 cases
behaved; the parity goldens re-pinned via `PARITY_UPDATE=1` and the delta
hand-diffed to exactly `est ~<10ms` → `est ~105µs` ×3; the io-surface
clean-clone probes ran against HEAD `f0e9ca5` built from /tmp
(d1/b7/extract verdicts in §a.8) with every extraction location checked
against source by script; the pre-flight SAFE gates ran BEFORE every
living-docs batch (harvest candidates, CHANGELOG bullets, row updates —
four SAFE gates, zero skips); `docs-gates purge . --apply` executed its
first real purge on this repo. **NOT verified here:** v0.2.0 NOT cut
(owner-gated, now 5th ask); dependabot PR #1 NOT merged (owner-gated
push); NOTHING in io-surface was written (its tree is dirty under an
active session — all filing blocked); the flake `pre-commit-check` runs
in the final sweep only; nixpkgs/SystemNix/crm untouched (owner-gated).

---

## a) FULLY DONE

1. **The 01-36 report §f harvested** (the session's headline debt):
   items 3–12 routed — owner rows minted for dependabot PR #1 (§4) and
   the post-release pbx pin bump (§3); the §8 filing row refreshed with
   probe results; the stale ROADMAP io-surface paragraph corrected; the
   unharvested tail captured as ROADMAP fuel (JSON-face contract growth
   bullet, perf-net shape series). Freshness confirms **0 unharvested
   open §f items** on the 01-36 report; the full harvest ledger is §h
   below. Commit `c7c0a4d`.
2. **Third purge pass** (3 struck probe rows → CHANGELOG, `c7c0a4d`) —
   and the LAST manual one: …
3. **`docs-gates purge` verb shipped** (`32e5b04`, the board's §f.9):
   strike-collect → dated CHANGELOG bullet per row (verdict tail cut at
   the table pipe) → docs-freshness pre-flight gate → write only on
   SAFE; dry-run default; exit 1 on UNSAFE, exit 2 on the on-disk/
   anchor refusals. Tests pin BOTH refusal legs red-controlled (the
   exit-1 fixture is a DF-flip construction proven against the real
   binary before encoding). **Live-proven**: `purge . --apply` executed
   its first real purge on this very board (`cb870f7`) — the tool
   purged the repo that defines it. Seven verbs now.
4. **Pre-commit `docs-battery` leg** (`09be419`, §f.8): a commit staging
   any living doc runs the whole battery BEFORE it lands — the 4cf0e50
   ungated-train lesson is now mechanical, riding the same
   git-hooks.nix config as nolint-audit (local arming still blocked by
   `core.hooksPath`, documented).
5. **§f.7 paid twice → fact widened**: the AGENTS pre-flight fact now
   scopes to ALL living-docs edits and names the battery-before-commit
   ordering (`c7c0a4d`). Every subsequent batch this session obeyed it
   (four pre-flight SAFE gates).
6. **Honest µs durations** (§f.27): `<1µs` / exact-µs / 100µs-steps
   tiers below 10ms; goldens re-pinned to the deterministic
   `est ~105µs`; matrix/heartbeat lines unchanged (value-normalized).
   Visible live: the matrix now prints `in 79µs`.
7. **§f.24 + §f.25**: the safe-strike harness carries the --battery legs
   (fake docs-gates on PATH proves invocation; without it the fallback
   fails LOUDLY — never a silent skip); the audit-all benchmark family
   gained the go-appkit pilot shape (12×150, ~1ms/op) beside the CV
   cliff.
8. **Clean-build io-surface probes (§f.6) — and they REVISE the pilot**:
   clone at HEAD `f0e9ca5`, built from /tmp:
   - **d1 NOT reproduced**: hand-written claims SURVIVE `init --force`
     — the dirty pilot build's destruction was evidently repaired by
     the sibling's committed scaffold work.
   - **b7 revised AGAIN**: `check --json` on a red tree at HEAD puts
     human error text on STDOUT and NO JSON (pilot's dirty build had
     valid JSON there) — the machine-face-on-failure contract does NOT
     hold on the committed tree.
   - **extractor**: 26 sites / 34 locations over docs-gates, all
     resolved-tier, **all verified at their exact source lines, 0
     drift** — closing §8.3 as verified (no context-trust needed).
   The §8 filing row carries all of this; filing itself stays blocked
   (active sibling session).
9. **§f.23 fleet skill text**: docs-health SKILL.md battery section now
   names the --json face (stdout-only consumers parse THAT, not the
   human matrix), the half-budget warning, the µs render, and the purge
   verb — committed in the SKILLS repo (`647e897`), fan-out verified by
   readlink through the delivering layer.

## b) PARTIALLY DONE

1. **io-surface filing** (§8.1 row): evidence is now clean-build-based,
   but the write stays blocked on the sibling's dirty tree + active
   session. Effort S once quiet.
2. **CHANGELOG purge-entry collapse** (§f.26): still rides the v0.2.0
   section cut (owner-gated); the §4 row names it.
3. **Provider handshake** (§f.19): ROADMAP bullet unchanged; the Spec's
   prose "v0.2.0+" still names a tag that does not exist.

## c) NOT STARTED (owner-gated / external — unchanged)

1. **v0.2.0** (§4, 5th ask) — the JSON contract the provider names is
   only in [Unreleased]; announcement + pbx/SystemNix floats queue
   behind it.
2. **Dependabot PR #1** (§4, new row): drift-verified safe; merge is a
   push.
3. SystemNix owner window (float + evo-x2 + wrapper switch), nixpkgs
   maintainers PR, crm harvest authority, D1/D2 rulings: §3/§4 rows
   unchanged.

## d) HONEST DEFECTS OF THIS SESSION

1. **One battery-after-commit** (`cb870f7`): the compound command
   committed before the sweep ran. Mitigated: the content was gated by
   the purge verb's own pre-flight SAFE + cadence, and the afterwards
   battery was green — but the ordering discipline I wrote into AGENTS
   hours earlier deserves better. Next session: `git commit` is the
   LAST command of a train, always.
2. **Probe setup burned two rounds**: the first d1 re-run appended to a
   guessed claims path (false DESTROYED verdict — the claim was never
   written) and the first b7 lacked a manifest (rc=2 usage, not the
   failure path). Caught by refusing to trust the first reading and
   inspecting what init actually creates; the final verdicts are from
   the corrected setups.
3. **The LSP ghost diagnostics recurred** (the documented
   `golangci_lint_ls` cache class): stale parse.go EOF + phantom lint
   warnings while the real gates were green. Protocol followed (prove
   with real gates; no upstream filing — cache-root cause known).

## e) WHAT WE SHOULD IMPROVE NEXT

1. **The purge verb's bullet shape is minimal** — one bullet per row,
   verdict verbatim. The manual purges composed GROUPED narratives
   (quality-machinery batch, battery batch…). Consider a `--group` mode
   or leaving grouping to a manual pass over the verb's draft.
2. **The battery hook in the flake check sandbox** costs a cold `go
   run` compile; if `pre-commit-check` slows noticeably, prebuild the
   binary into the hook env (docsGatesFor) instead of `go run`.
3. **§e.5/6/7 of the 01-36 report stand** (watchdog JSON field, Spec
   version in prose, estimate µs — the last one is DONE now).

## f) Up to 20 things to get done next

1. Cut v0.2.0 (owner, 5th ask; collapses the purge entries at the cut).
2. Merge dependabot PR #1 (owner; drift-verified).
3. File the io-surface findings when its tree quiets down (row §8.1,
   clean-build evidence attached).
4. Post-release: pbx pin bump (§3 new row) + SystemNix float window.
5. The ROADMAP tail is current (JSON-face contract growth, perf-net
   shape series); no orphaned §f items — freshness 0 unharvested.

## g) The three questions (unchanged, carried forward)

1. **Push authority** — MOOT: all trains through `cb870f7` are on
   origin (the owner/session pushed; origin/main == local HEAD at
   session start). The closing trains await the same treatment.
2. **v0.2.0 release authority** — 5th ask.
3. **Dependabot PR #1 merge** — owner call.

## h) HARVEST ledger — 01-36 report §f dispositions

| §f item                       | Disposition     | Destination / reason                                                                 |
| ----------------------------- | --------------- | ------------------------------------------------------------------------------------ |
| 1 cut v0.2.0                  | existing row    | §4 v0.2.0 (evidence extended: 5th ask, §f.26 collapse note)                           |
| 2 push trains                 | done in code    | origin/main carries them (verified at session start) — moot                           |
| 3 archive next report         | executed here   | this report triggered it; 01-36 archived with the live-remainder banner               |
| 4 merge PR #1                 | new row         | §4 (owner-gated push)                                                                 |
| 5 file io findings            | existing row    | §8.1 (evidence: b6 cleared → dirty-tree blocker → clean-build revisions)              |
| 6 re-run probes clean         | done in code    | §a.8 of this report (clean clone f0e9ca5)                                             |
| 7 widen pre-flight fact       | done in code    | AGENTS first fact widened (`c7c0a4d`)                                                 |
| 8 pre-commit battery leg      | done in code    | flake.nix docs-battery hook (`09be419`)                                               |
| 9 purge verb                  | done in code    | `32e5b04` + live dogfood `cb870f7`                                                    |
| 10 io stdout/stderr note      | merged          | into §8.1 (5-finding list)                                                            |
| 11 io d1 fix upstream         | merged          | into §8.1 — now MOOT-ish: d1 not reproduced at HEAD                                   |
| 12 pbx pin bump               | new row         | §3 (blocked on the release)                                                           |
| 13 SystemNix owner window     | existing row    | §3 consolidation rows                                                                 |
| 14 dependabot-auto-configure  | existing (ROADMAP-adjacent) | §a.15 of 01-36 verified fix upstream unreleased; fleet-side wait          |
| 15 nixpkgs maintainers PR     | existing row    | §3 E22 (githubId resolved)                                                            |
| 16 D1 ruling                  | existing row    | §4 (matrix linked)                                                                    |
| 17 D2 ruling                  | existing row    | §4 (go-floors enforces mechanically)                                                  |
| 18 io-surface adoption ruling | existing row    | §4                                                                                    |
| 19 provider handshake         | existing        | ROADMAP provider-contract bullet                                                      |
| 20 JSON schema_version        | new (ROADMAP)   | JSON-face contract growth bullet                                                      |
| 21 watchdog field             | new (ROADMAP)   | same bullet                                                                           |
| 22 size-series bench          | new (ROADMAP)   | perf-net shape series bullet                                                          |
| 23 fleet skill text pass      | done in code    | SKILLS repo `647e897`                                                                 |
| 24 safe-strike --battery test | done in code    | harness cases 12–13                                                                   |
| 25 go-appkit bench case       | done in code    | BenchmarkCheckReportReplayGoAppkitShape                                               |
| 26 collapse purge entries     | existing row    | §4 v0.2.0 evidence (at the cut)                                                       |
| 27 estimate µs precision      | done in code    | FormatDuration tiers + goldens re-pinned                                              |
| 28 AGENTS compression (F94)   | existing row    | §2c owner-gated tier                                                                  |
| 29/30 post-switch LSP rows    | existing rows   | §2c (blocked on the SystemNix switch)                                                 |
| 31/32 crm/go-appkit ratchets  | existing rows   | §3                                                                                    |
| 33–37 doctor/margins-json/CutStringFlag/split-check/cache-state | existing | ROADMAP bullets (verified current)          |
| 38–43 gate-family candidates  | existing        | ROADMAP gate-family bullet                                                            |
| 44/45 BuildFlow upstream      | existing rows   | §3 (concurrency constraint noted)                                                     |
| 46 golangci merge-train       | declined        | the golangci-drift flake check owns the pin mechanically; nothing to do               |
| 47 wrap-or-strike ritual      | executed here   | the board carries zero bare-DONE rows; purge verb covers the cycle                    |
| 48 skip_gates parity arm      | new (ROADMAP)   | JSON-face contract growth bullet                                                      |
| 49 provider findings message  | new (ROADMAP)   | same bullet                                                                           |
| 50 rc=N in row message        | new (ROADMAP)   | same bullet                                                                           |

---

_Point-in-time snapshot; the daemon moves HEAD continuously. The §f list
above is the next session's HARVEST input (deliberately short — the
board and ROADMAP carry the real queue)._
