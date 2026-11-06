# Status Report — Harvest, Purge Verb, Gated Trains, Clean-Build Probes: Full Session Close

**Report:** 2026-11-06 05:01 CET · **Repo:** docs-gates @ main (deliberate
trains `c7c0a4d`…`8e55e12` + this report; the daemon interleaves
`chore: auto-commit` entries per its config) · **Session window:**
~02:00–05:01 CET · **Supersedes** the 04-45 intermediate report (archived
as superseded-within-the-hour; its §h harvest ledger is carried forward
HERE as §h).

**Scope line — what this session VERIFIED (fresh command behind every
claim):** dogfood battery 5/5 PASS BEFORE every docs commit (`c7c0a4d`,
`09be419`, `b222ef9`, `8e55e12`; the one exception `cb870f7` was gated by
the purge verb's OWN internal pre-flight SAFE + a green afterwards run —
§d.1); root suite `go test -count=1 ./...` → **10/10 packages ok**
(docspurge joins as the 10th); provider suite green from its own
directory; golangci **0 issues on BOTH modules**; `erraudit lint ./...
--type-aware` rc=0; `erraudit nolint-audit .` → **14 directives, 0
stale**; `scripts/check-erraudit-rev.sh` PASS (flake.lock matches
3d1f7ed); `nix fmt` zero changes; `nix flake check --all-systems
--no-build` → **all checks passed!** (incl. the new docs-battery hook
leg riding pre-commit-check); the skill fan-out guard → **OK: 51
symlinks, no store targets**; safe-strike harness **13/13 cases
behaved** (the two new --battery legs included); parity goldens
re-pinned via `PARITY_UPDATE=1` with the delta hand-diffed to exactly
`est ~<10ms` → `est ~105µs` ×3 (deterministic: 1×7×15µs); the
io-surface clean-clone probes built from /tmp at sibling HEAD `f0e9ca5`
with every extraction location script-checked against source (0 drift);
FOUR pre-flight SAFE gates ran before the four living-docs batches —
zero ungated edits this time. **NOT verified here:** v0.2.0 NOT cut
(owner-gated, 6th ask); dependabot PR #1 NOT merged (owner-gated push);
NOTHING written into io-surface (dirty tree, active session); the
closing trains are NOT pushed (see §g Q3); CI never ran locally (CI
judges on push); the docs-battery hook is NOT armed locally
(`core.hooksPath` blocker, pre-existing, documented).

---

## a) FULLY DONE

1. **Resume state pinned before touching anything**: HEAD == origin/main
   (the six trains from the previous session ARE pushed — §f.2 of the
   01-36 report is moot); the two daemon commits on top are a formatter
   pass + the 01-36 report landing (verified benign by diff).
2. **HARVEST of the 01-36 report §f — all 50 items dispositioned**
   (ledger in §h below; commit `c7c0a4d`): owner rows minted for the
   dependabot PR merge (§4) and the post-release pbx pin bump (§3); the
   §8 io-surface filing row refreshed; the stale ROADMAP probe paragraph
   corrected; the tail captured as ROADMAP fuel (new JSON-face contract
   growth bullet, perf-net shape series bullet). Freshness afterwards:
   **0 unharvested open §f items** on the 01-36 report; audit-all 0
   across the corpus.
3. **Third purge pass** (3 struck probe rows → CHANGELOG, `c7c0a4d`) —
   the last MANUAL one, because…
4. **`docs-gates purge` verb shipped and LIVE-proven** (`32e5b04`, the
   board's §f.9): collects struck TODO rows, drafts one dated CHANGELOG
   bullet per row (verdict tail cut at the table pipe — a real bug the
   first draft had, caught by the UNSAFE-fixture work), gates the draft
   through `docsfreshness.PreFlight`, writes only on SAFE; dry-run
   default; exit 1 UNSAFE / exit 2 on-disk-or-anchor refusals; wired
   into all five dispatch/help tables (the cmd suite's consistency pin
   covers it automatically). **Nine tests**, including the two
   red-controlled refusal legs: the exit-1 fixture is a DF-flip
   construction (zorphuser at DF 7 + candidate bullet → 8 = the
   GenericDF bar) proven against the REAL binary before being encoded,
   and the exit-2 leg pins the on-disk refusal surfacing with nothing
   written. **First live purge executed on this repo** (`cb870f7`):
   the verb collected the struck clean-build row, drafted, pre-flighted
   SAFE, purged; cadence confirmed 0 struck rows. Seven verbs now.
5. **§f.7 — the pre-flight fact widened after its second payment**:
   AGENTS' first non-obvious fact now scopes to ALL living-docs edits
   and names battery-BEFORE-commit (`c7c0a4d`). The rest of the session
   obeyed it: four pre-flight SAFE gates before four living-docs
   batches.
6. **§f.27 — honest µs durations**: FormatDuration renders `<1µs` /
   exact µs / 100µs-steps below 10ms (the old `<10ms` bucket was honest
   but useless on the estimate line); goldens re-pinned to the
   deterministic `est ~105µs`; matrix/heartbeat lines unchanged
   (value-normalized by the parity regex, whose now-dead `<10ms`
   alternation was removed with its comment). Visible live: the matrix
   prints `in 79µs` / `in 6.6ms`.
7. **§f.8 — pre-commit `docs-battery` leg** (`09be419`): git-hooks.nix
   `preCommitCheckFor` now carries a second hook; a commit staging
   TODO_LIST/CHANGELOG/ROADMAP/AGENTS/FEATURES runs the whole battery
   before the commit lands. Rides the same hermetic env plumbing as
   nolint-audit (pinned go, GOTOOLCHAIN=local, GOPROXY=off, /tmp
   caches); proven green inside the flake-check sandbox; local arming
   still blocked by `core.hooksPath` (pre-existing, documented).
8. **§f.23 — fleet skill text pass**: docs-health SKILL.md's battery
   section now names the `--json` machine face (consumers parse THAT,
   never the human matrix), the half-budget warning, the µs render, and
   the purge verb. Committed in the SKILLS repo (`647e897`); the
   fan-out delivering layer verified by readlink + content-through-
   fan-out + the crush-config guard (51 clean symlinks).
9. **§f.24 — safe-strike --battery legs**: harness cases 12 (fake
   docs-gates on PATH → stub invoked, marker line asserted) and 13
   (without the binary the go-run fallback fails LOUDLY in a non-module
   scratch repo — never a silent skip). 13/13 cases behaved.
10. **§f.25 — go-appkit corpus shape in the benchmark family**:
    `BenchmarkCheckReportReplayGoAppkitShape` (12×150, ~1.0ms/op) pins
    the pilot-board cost point beside the CV cliff, so a memoization
    regression surfaces on the small shape too.
11. **§f.6 + §8.3 — clean-build io-surface probes, and they REVISE the
    pilot verdicts**: clean clone at sibling HEAD `f0e9ca5`, binary
    built from /tmp:
    - **d1 NOT reproduced**: a hand-written claims.cue append SURVIVES
      `init --force` at HEAD — the dirty pilot build's destruction was
      evidently repaired by the sibling's since-committed scaffold
      work. The upstream-filing urgency drops sharply.
    - **b7 revised AGAIN, in the bad direction**: `check --json` on a
      red tree at HEAD puts human error text on STDOUT and NO JSON at
      all (the pilot's dirty build had valid JSON on stdout there) —
      the machine-face-on-failure contract does NOT hold on the
      committed tree. That is the strongest filing candidate.
    - **extractor**: 26 sites / 34 locations over docs-gates (the
      corpus grew from the pilot's 18 — docspurge added 4), all
      resolved-tier, **every location verified at its exact source
      line, 0 drift** — §8.3 closes as VERIFIED, no context-trust
      needed.
12. **Closing docs**: the 04-45 intermediate report + the 01-36 archive
    (live-remainder banner, `git mv`, corpus 15→15 with the new newest
    report); CHANGELOG caught up (purge verb, hook leg, µs durations);
    AGENTS counts seven verbs, documents the hook leg + the purge
    command; FEATURES/ROADMAP current.
13. **Full end-of-session verification sweep — every leg green**: root
    suite 10/10, provider suite ok, golangci 0+0, erraudit rc=0,
    nolint-audit 14/0 stale, rev tripwire PASS, `nix fmt` clean, flake
    check all-passed, fan-out guard OK, harness 13/13, dogfood battery
    5/5.

## b) PARTIALLY DONE

1. **io-surface filing** (§8.1 row): the evidence is now clean-build-
   based and STRONGER (b7's broken JSON failure face is the headline
   finding; d1 is near-moot), but the write stays blocked on the
   sibling's dirty tree + active session. Effort S once quiet.
2. **CHANGELOG purge-entry collapse** (§f.26): three same-day purge
   entries now sit in [Unreleased]; collapsing them rides the v0.2.0
   section cut (owner-gated). The §4 row names it.
3. **Provider handshake** (§f.19): the Spec's prose "v0.2.0+" still
   names a tag that does not exist; the mechanical-handshake decision
   lives in ROADMAP untouched.
4. **Local hook arming**: the docs-battery leg runs in the flake-check
   sandbox but NOT on local commits (`core.hooksPath` set to its own
   default; pre-commit refuses to install over it). The blocker and the
   one-line fix are documented in AGENTS; unsetting it is a local host
   decision.
5. **The µs estimate still prints for tiny corpora** — legible now
   (`est ~105µs`), but the 01-36 §e.7 alternative (suppress under a
   threshold) remains open polish; not obviously wanted anymore.

## c) NOT STARTED (owner-gated / external — the standing queue)

1. **v0.2.0 release** (§4, 6th ask): versions the --json contract the
   provider names; collapses the purge entries at the cut; the fleet
   announcement + pbx/SystemNix floats queue behind it.
2. **Dependabot PR #1** (§4, new row): golangci-lint-action SHA bump
   only, drift-verified safe, CI green on the PR; merging is a push.
3. **SystemNix owner window** (§3): float to the hardened flake (needs
   the v0.2.0 tag) + evo-x2 deploy (live-PATH staleness is also the
   un-fixed-wedge binary) + the wrapper switch — one window covers all.
4. **nixpkgs `lib.maintainers.LarsArtmann` PR** (§3): githubId
   resolved; drafting is owner-gated.
5. **crm harvest + ratchet drains** (§3): blocked on the standing
   cross-repo authority ruling; go-appkit is genuine-open (12/12).
6. **D1 (host battery adoption) / D2 (go.mod floor)** rulings (§4):
   matrix + recommendation written in docs/planning.
7. **io-surface adoption ruling** (§4): pilot + clean-probe data all on
   the table.
8. **F94 AGENTS compression pass** (§2c, owner tier): the
   non-obvious-facts section grew again (widened fact, hook leg, seven
   verbs).
9. **Post-switch LSP rows** (§2c ×2): blocked on the SystemNix wrapper
   switch.
10. **BuildFlow upstream legs** (§3): #87 cache-key fix + the
    nix-fmt/treefmt roundtrip --tree-root forwarding — their repo
    carries active sessions; coordinate before any PR.

## d) TOTALLY FUCKED UP (all recovered; listed so they are not repeated)

1. **One battery-after-commit** (`cb870f7`): a compound
   `git commit … && battery` inverted the ordering I had codified hours
   earlier. Mitigated in fact (the commit's content was gated by the
   purge verb's internal pre-flight SAFE + cadence, and the afterwards
   battery was green) but not in discipline. The fix is mechanical, not
   willpower — see §e.3.
2. **Probe setups that produced a FALSE verdict**: the first d1 re-run
   appended to a guessed claims path (`manifest/claims.cue` — wrong;
   init actually creates `docs/architecture/io-surface/claims.cue`),
   failed the append silently, and then reported "claim DESTROYED" —
   which was true only because the claim was never written. The first
   b7 ran without a manifest and reported rc=2 (usage), which I nearly
   misread as the failure path. Both were caught by refusing to trust
   the first reading and inspecting what the tool actually creates;
   every verdict in §a.11 comes from the corrected setups. The lesson
   is probe-shaped, not tool-shaped (§e.2).
3. **The LSP ghost-diagnostics class recurred** (documented cache
   artifact): stale `parse.go EOF` typecheck errors and phantom lint
   warnings persisted across file rewrites while `go build`, the real
   `golangci-lint run`, and the suites were green. Protocol followed
   (prove with real gates; never trust the pane); the wrapper-cache
   clearing from the 2026-11-04 protocol would have helped — I worked
   around instead of curing.
4. **Edit-tool exact-match churn under a live formatter**: the daemon
   re-aligns markdown tables continuously, so three exact-match edits
   failed (one for a genuine typo — `354` vs `355` — and two for
   formatter reflow), and two multiedit batches left partial-apply
   states: a duplicated function header in parse.go (caught by
   `go build` immediately) and a mangled candidates file (caught by
   re-reading before the pre-flight). Line-index Python edits are the
   robust shape for formatter-owned files.
5. **Daemon message-heuristic wrapping**: the bulk of the µs and purge
   trains landed inside `chore: auto-commit` entries (the daemon
   commits every ~2–3 min and I was mid-file). All work IS in history
   in order — but three feature-sized changes carry heuristic messages.
   Accepted (never rewrite under the daemon); the lesson is commit
   fast, in small deliberate units, the moment a unit is green.

## e) WHAT WE SHOULD IMPROVE

1. **`docs-gates purge --group`**: the verb writes one bullet per row
   (faithful but flat); the manual purges composed grouped narratives.
   Either add a grouping mode or document one-bullet-per-row as THE
   shape and stop composing narratives by hand.
2. **A probe protocol lesson**: "inspect the artifact the tool created
   BEFORE asserting a verdict" — the d1 false-DESTROYED class deserves
   a row in crush-config's references/lessons.md (cross-project: any
   probe against a tool's output layout). Needs a commit there, not an
   in-session write.
3. **Make battery-before-commit mechanical**: a `scripts/dogfood-commit.sh`
   (battery → commit, failing closed) or a git-hooks entry — the
   §d.1 inversion should be impossible, not discouraged. The docs-
   battery hook already exists; arming it locally is the cheaper half.
4. **Report §f discipline**: the 01-36 report's 50-item brainstorm cost
   half a session to route. This report deliberately carries a SHORT
   §f — the board + ROADMAP are the queue; reports should route
   as-they-land, not accumulate brainstorms.
5. **Hook-env prebuild**: if `pre-commit-check` slows measurably from
   the battery leg's cold `go run`, prebuild the binary into the hook
   env (docsGatesFor) instead of compiling per run. Not measured yet —
   the sweep passed comfortably.
6. **LSP cache protocol**: follow the 2026-11-04 cure (clear the
   wrapper caches + restart) at the FIRST ghost instead of working
   around; the pane cost real rounds this session.

## f) Up to 50 things we should get done next

(Routing: 1–12 are the actionable frontier; 13–50 are the standing
board/ROADMAP inventory this session verified current — each already
lives in TODO_LIST/ROADMAP with evidence; listed here so the next
session's HARVEST has the full picture without re-reading the board.)

| #  | Task                                                                                     | Impact   | Effort | Category      |
| -- | ---------------------------------------------------------------------------------------- | -------- | ------ | ------------- |
| 1  | Cut v0.2.0 (section cut collapses the purge entries; go-release ritual; fleet announce)  | Critical | M      | Release       |
| 2  | Merge dependabot PR #1 (drift-verified; the merge is a push)                             | High     | S      | Release       |
| 3  | Push the session's deliberate trains (owner ruling — see §g Q3)                          | High     | S      | Release       |
| 4  | File the io-surface findings when its tree quiets (clean-build data: b7 headline, d1 moot) | High    | S      | External      |
| 5  | Post-release: bump pbx-artmann's pin (§3 row)                                            | Medium   | S      | External      |
| 6  | SystemNix owner window: float + evo-x2 deploy + wrapper switch (one window)              | High     | M      | External      |
| 7  | nixpkgs lib.maintainers.LarsArtmann PR (owner-gated)                                     | Low      | S      | External      |
| 8  | crm harvest authority ruling; then the ratchet drain                                     | Medium   | M      | External      |
| 9  | D1 ruling: host battery adoption vs PATH-only (matrix exists)                            | Medium   | S      | Decision      |
| 10 | D2 ruling: root go.mod floor unification (recommendation written)                        | Low      | S      | Decision      |
| 11 | io-surface adoption ruling (§4; ROADMAP data point current)                              | Medium   | S      | Decision      |
| 12 | Arm the local pre-commit hooks (unset core.hooksPath) so docs-battery guards local work  | Medium   | S      | Quality       |
| 13 | `scripts/dogfood-commit.sh`: battery-then-commit, failing closed (§e.3)                  | Medium   | S      | Quality       |
| 14 | `docs-gates purge --group` mode or canonicalize one-bullet-per-row (§e.1)                | Low      | S      | Feature       |
| 15 | Probe-protocol lesson → crush-config references/lessons.md (§e.2, needs a commit there)  | Medium   | S      | Documentation |
| 16 | F94: AGENTS non-obvious-facts compression pass (owner tier)                              | Low      | M      | Documentation |
| 17 | JSON face: `schema_version` field BEFORE any shape change (ROADMAP doctrine)             | Medium   | S      | Feature       |
| 18 | JSON face: `watchdog` gate-row field (contract change → re-pin golden + provider)        | Low      | M      | Feature       |
| 19 | Provider: per-gate captured findings into finding.Message                                | Low      | M      | Feature       |
| 20 | JSON row message: render `rc=N` (human-face parity)                                      | Low      | S      | Polish        |
| 21 | battery-json golden: `skip_gates` parity arm                                             | Low      | S      | Quality       |
| 22 | Audit-all size-series benchmark (100/400/1600) — cliff SHAPE                             | Medium   | M      | Quality       |
| 23 | Provider handshake: mechanical version check (Spec version vs `--version` probe)         | Medium   | M      | Feature       |
| 24 | go-appkit/crm: tighten ratchets as owning sessions land items (standing doctrine)        | Medium   | M      | Ritual        |
| 25 | Post-switch: LSP-ghost protocol becomes "restart suffices" (§2c.1)                       | Low      | S      | Documentation |
| 26 | Post-switch: cold-cache experiment IF ghosts survive (§2c.2)                             | Low      | M      | Quality       |
| 27 | dependabot-auto-configure release carrying HasUnmodeledGroups; then bump BuildFlow's dep | Low      | S      | External      |
| 28 | BuildFlow#87 fix upstream; drop the BUILDFLOW_NO_RESULT_CACHE bypass                     | Low      | L      | External      |
| 29 | BuildFlow roundtrip --tree-root forwarding (coordinate with their sessions)              | Low      | M      | External      |
| 30 | `docs-gates doctor` verb (PATH, config, flake state in one command)                      | Low      | M      | Feature       |
| 31 | `--margins --json` for tooling consumption                                               | Low      | M      | Feature       |
| 32 | `CutStringFlag` (string sibling; re-needed by every new valued flag)                     | Low      | S      | Feature       |
| 33 | Split-check verb (verify one item's verdict without a full run)                          | Low      | M      | Feature       |
| 34 | Findings-gate cache-state summary                                                        | Low      | S      | Feature       |
| 35 | Pre-flight model growth: block replacements + `--remove` (ROADMAP)                       | Low      | M      | Feature       |
| 36 | Margins/routing convergence: Gap rows carrying harvest suggestions                       | Low      | M      | Feature       |
| 37 | False item-boundary detector (§f continuation `\d+[.)]` class)                           | Low      | M      | Feature       |
| 38 | Generic counts gate design pass (FEATURES claims vs the tree)                            | Low      | L      | Feature       |
| 39 | Link/reference checker for living docs (design pass)                                     | Low      | L      | Feature       |
| 40 | Report-freshness age gate distinct from cadence (design pass)                            | Low      | M      | Feature       |
| 41 | errcheck/gosec fleet-wide nolint staleness tripwires                                     | Low      | M      | Quality       |
| 42 | tokencorpus extraction IF gates ever version independently                               | Low      | M      | Cleanup       |
| 43 | §1 board row: annotate sweep over older unarchived reports — only at cadence pressure    | Low      | M      | Ritual        |
| 44 | Estimate-line suppression under a threshold (µs render may have settled it — decide)     | Low      | S      | Polish        |
| 45 | Verify the docs-battery hook cost inside pre-commit-check; prebuild if slow              | Low      | S      | Quality       |
| 46 | Archive the next report when it lands (standing ritual; corpus at 15/15 again)           | Medium   | S      | Ritual        |
| 47 | Windows/macOS consumers revisit only on a real consumer (ROADMAP)                        | Low      | -      | Non-goal      |
| 48 | Sub-10ms estimate honest-but-useless follow-up: drop if µs render satisfies              | Low      | S      | Polish        |
| 49 | Keep the §8 filing row's clean-build evidence fresh if the sibling lands more work       | Low      | S      | External      |
| 50 | Retire the ROADMAP provider-contract bullet the day the handshake decision lands         | Low      | S      | Documentation |

## g) Three questions I can NOT figure out myself

1. **Release authority, 6th ask: may I cut v0.2.0?** The provider Spec
   prose names a tag that does not exist; the §f.26 collapse decision
   (three same-day purge entries → one) settles at the cut; the pbx
   bump + SystemNix float queue behind it. Everything on the board says
   GO — I need your word (and then I run the go-release ritual).
2. **May I merge dependabot PR #1?** Verified this session's baseline:
   only the golangci-lint-action SHA moves, `version: v2.15.0` and the
   flake pin are untouched, the golangci-drift check owns the contract
   mechanically, CI is green on the PR. Merging is a push — your call.
3. **Push authority for THIS session's trains** (`c7c0a4d`…`8e55e12` +
   this report): the previous session's trains reached origin (you or a
   session pushed them), so the pattern exists — but my harness still
   wants an explicit ask per train. Say the word and everything local
   goes up in one push.

---

_Point-in-time snapshot; the daemon moves HEAD continuously. §h is the
canonical harvest ledger for the 01-36 report (the 04-45 intermediate
report was archived as superseded by THIS report)._

## h) HARVEST ledger — 01-36 report §f dispositions

| §f item                                        | Disposition   | Destination / reason                                                        |
| ---------------------------------------------- | ------------- | --------------------------------------------------------------------------- |
| 1 cut v0.2.0                                   | existing row  | §4 (evidence extended: 6th ask, §f.26 collapse note)                        |
| 2 push trains                                  | done          | origin carried them at session start — moot                                 |
| 3 archive next report                          | executed      | 01-36 archived this session; corpus 15/15 again                             |
| 4 merge PR #1                                  | new row       | §4 (owner-gated push)                                                       |
| 5 file io findings                             | existing row  | §8.1 (three evidence revisions this session)                                |
| 6 re-run probes clean                          | done          | §a.11 (clean clone f0e9ca5)                                                 |
| 7 widen pre-flight fact                        | done          | AGENTS first fact (`c7c0a4d`)                                               |
| 8 pre-commit battery leg                       | done          | flake.nix docs-battery hook (`09be419`)                                     |
| 9 purge verb                                   | done          | `32e5b04` + live dogfood `cb870f7`                                          |
| 10 io stdout/stderr note                       | merged        | into §8.1's finding list                                                    |
| 11 io d1 fix upstream                          | merged/moot   | d1 not reproduced at HEAD (§a.11)                                           |
| 12 pbx pin bump                                | new row       | §3 (blocked on the release)                                                 |
| 13 SystemNix owner window                      | existing row  | §3 consolidation                                                            |
| 14 dependabot-auto-configure release           | existing      | fix verified upstream-unreleased (01-36 §a.15); fleet-side wait             |
| 15 nixpkgs maintainers PR                      | existing row  | §3 E22                                                                      |
| 16 D1 ruling                                   | existing row  | §4 + planning matrix                                                        |
| 17 D2 ruling                                   | existing row  | §4 + go-floors mechanical enforcement                                       |
| 18 io-surface adoption ruling                  | existing row  | §4                                                                          |
| 19 provider handshake                          | existing      | ROADMAP bullet                                                              |
| 20 JSON schema_version                         | new (ROADMAP) | JSON-face contract growth bullet                                            |
| 21 watchdog field                              | new (ROADMAP) | same bullet                                                                 |
| 22 size-series bench                           | new (ROADMAP) | perf-net shape series bullet                                                |
| 23 fleet skill text pass                       | done          | SKILLS `647e897`, fan-out verified                                          |
| 24 safe-strike --battery test                  | done          | harness cases 12–13                                                         |
| 25 go-appkit bench case                        | done          | BenchmarkCheckReportReplayGoAppkitShape                                     |
| 26 collapse purge entries                      | existing row  | §4 v0.2.0 evidence (at the cut)                                             |
| 27 estimate µs precision                       | done          | FormatDuration tiers + goldens re-pinned                                    |
| 28 AGENTS compression (F94)                    | existing row  | §2c owner tier                                                              |
| 29 post-switch LSP "restart suffices"          | existing row  | §2c.1                                                                       |
| 30 post-switch cold-cache experiment           | existing row  | §2c.2                                                                       |
| 31 crm harvest                                 | existing row  | §3 (authority ruling)                                                       |
| 32 go-appkit/crm ratchet tightening            | existing row  | §3 doctrine                                                                 |
| 33 doctor verb                                 | existing      | ROADMAP                                                                     |
| 34 --margins --json                            | existing      | ROADMAP                                                                     |
| 35 CutStringFlag                               | existing      | ROADMAP                                                                     |
| 36 split-check verb                            | existing      | ROADMAP                                                                     |
| 37 cache-state summary                         | existing      | ROADMAP                                                                     |
| 38 false item-boundary detector                | existing      | ROADMAP gate-family bullet                                                  |
| 39 generic counts gate                         | existing      | ROADMAP gate-family bullet                                                  |
| 40 link/reference checker                      | existing      | ROADMAP gate-family bullet                                                  |
| 41 report-freshness age gate                   | existing      | ROADMAP gate-family bullet                                                  |
| 42 errcheck/gosec tripwires                    | existing      | ROADMAP fleet bullet                                                        |
| 43 tokencorpus extraction                      | existing      | ROADMAP decoupling bullet                                                   |
| 44 BuildFlow#87                                | existing row  | §3                                                                          |
| 45 BuildFlow roundtrip --tree-root             | existing row  | §3                                                                          |
| 46 golangci merge-train                        | declined      | the golangci-drift flake check owns the pin mechanically — nothing to do    |
| 47 wrap-or-strike ritual                       | executed      | zero bare-DONE rows; the purge verb now covers the DONE→strike→purge cycle  |
| 48 skip_gates parity arm                       | new (ROADMAP) | JSON-face contract growth bullet                                            |
| 49 provider findings message                   | new (ROADMAP) | same bullet                                                                 |
| 50 rc=N in row message                         | new (ROADMAP) | same bullet                                                                 |
