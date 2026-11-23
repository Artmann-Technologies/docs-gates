# Status Report — Fleet library audit run: 23 reports, the gates' blind spots, and this session's own mistakes (docs-gates)

> **Date:** 2026-11-24 00:12 CET · **Session:** the 2026-11-23 fleet
> library utilization audit (21 libraries, 23 subjects) and its close-out.
> The brutal self-review is folded into sections d) and e) on request, not
> written as a separate file.

## Scope

**Verified for this report** (every number below was re-derived between
2026-11-24 00:03 and 00:40 on tree `2258fe9` plus this session's
uncommitted edits, not remembered):

- `python3 scripts/replay-verify.py` — 196 of 196 verify commands in the 23
  reports replay exactly (strict, order-insensitive).
- `python3 scripts/citation-sweep.py` — 702 `path:line` citations resolve
  and are in range.
- `go run ./cmd/docs-gates run .` — all ten gates PASS.
- `go run ./cmd/docs-gates freshness --land-parity .` — the 26 TODO rows of
  §10 and §11 landed byte-exact in the formatter's padded bytes (SAFE batch
  2026-11-23T23:22:01Z); the 24 ROADMAP lines landed verbatim in the
  earlier batch (SAFE 23:16:51Z, 46 content lines).
- `freshness --dump-items` over each of the 12 unarchived reports, plus a
  bullet count of each `f)` section — the basis of d2.
- `go run ./cmd/docs-gates freshness .` on this report before any
  harvest — "0 unharvested open §f item(s), FRESH" with all 50 items
  extracted; `--explain-item` on five of them — the basis of d2b.
- `go run ./cmd/docs-gates rows TODO_LIST.md` — exit 1, two tables
  flagged; `go run ./cmd/docs-gates purge .` — dry-run only, drafts one row.
- `git log -S` on TODO_LIST.md for the struck-row shape — first appears
  2026-11-04; ten such rows already present at this session's baseline
  `adfa9c8`.
- `buildflow --build-mode dev` (2026-11-23 19:42) — rc 69, two error
  findings; `buildflow -s go-structure-linter --format finding` re-run at
  00:10 — "AGENTS.md has 229 lines (maximum: 220, excess: 9)".
- Headless render of `docs/review/libs/index.html` (1440x8997) read back
  by a vision subagent in five passes.
- `git rev-list --count origin/main..HEAD` — 34.

**NOT verified here:**

- **No Go test suite was run by me.** No Go code changed, so I relied on
  that; the 34 session commits are unpushed, so CI has not seen any of
  this work either.
- **The 23 reports' prose was sampled, not re-read end to end.** The two
  instruments check every citation and every verify command mechanically;
  the narrative claims were spot-checked (first report per group plus every
  claim quoted in the dashboard).
- **Library-side observations are the auditors', not mine** (cmdguard
  suggester, go-codec knobs, go-credentials-store drift). I did not
  re-verify them in the library repos.
- **The TODO_LIST.md corruption was diagnosed, not repaired.** Its
  mechanism was reproduced in a scratch file; which session wrote the bad
  strikes was not determined.
- **Whether the three alternate model providers still have no credits** —
  observed once, not re-checked.
- **Archived reports were not opened** (ritual rule).

## a) FULLY DONE

- **23 per-library reports** at `docs/review/libs/*.md` (21 libraries;
  httputil's two submodules carry their own reports). Uniform seven-section
  skeleton, every one recording the tag and commit reviewed. Evidence: the
  recount script — 274 capability rows, 38 applicable, 236 not applicable,
  zero invariant violations, all 23 reports.
- **The dashboard** `docs/review/libs/index.html` plus its render trace
  `index-screenshot.webp`. Evidence: self-consistency gate (rows numbered
  1–23, sorted by verdict then name, sums 274 and 38, trigger table in the
  same order, 53 links resolving); visual verdict PASS.
- **Two mechanical instruments**: `scripts/citation-sweep.py` and
  `scripts/replay-verify.py` (stdlib Python, read-only guard). Both green
  at HEAD.
- **Findings routed, not entombed**: four hardening rows in TODO_LIST §10
  (pre-flight SAFE, land-parity OK); the fifth landed in AGENTS.md.
- **AGENTS.md carries the durable outcome**: the audit verdict, the
  re-audit triggers and the torn-write decision (AGENTS.md:123-134).
- **Report defects the instruments caught and I fixed**: 58 non-literal
  transcripts made literal, 30 more in the strict pass, 33 tip queries
  pinned to the reviewed commit; wrong-content citations corrected
  (`flags_suggest.go:13-42` to `13-84`, `pkg/provider/go.mod:5` to `:6`,
  `docsstub.go:19` to `:20`, go-daemon's tag `v0.3.1` to `v0.3.0`, a
  claimed `text/template` hit that does not exist, two regexes that did
  not parse); status vocabulary reconciled across 18 reports; three extra
  top-level sections demoted; two rule-F spacing faults repaired.
- **AGENTS.md**: one exact-duplicate BuildFlow#87 bullet merged, the
  "no result-cache bypass" nuance carried into the surviving bullet.
- **Two of my own incidents reverted before they could ship** — see d6.

## b) PARTIALLY DONE

- **Quality gate.** Works: the battery is 10/10 and the two instruments
  are green. Open: `buildflow` still exits 69 on two error findings
  (AGENTS.md size, the documented md5 delta). Blocker: the distill is the
  owner's in-flight work. Effort: S for the lines, M if the 220 rule is
  to be revisited.
- **The four hardening items.** Works: found, evidenced, routed. Open: no
  code written (TODO §10 rows 1–4). Blocker: none, awaiting a go-ahead
  (question 3). Effort: S each.
- **The audit instruments.** Works: strict replay, citation sweep. Open:
  hardcoded `/home/lars/projects` paths, no tests, nothing invokes them,
  the interpreter leaks in from the host profile rather than the devShell,
  and three helper scripts I ran from `/tmp` were never shipped (fleet
  facts, transcript literalizer, dashboard builder). Effort: S–M each.
- **Dashboard versus reports.** Works: consistent today. Open: the
  verdict matrix and trigger table are hand-copied from the reports with
  no generator or check between them. Effort: M.
- **Discoverability.** Works: AGENTS.md names the dashboard. Open: neither
  README nor `docs/agent-notes.md` points at the corpus, and compressing
  AGENTS.md dropped the scripts' names. Effort: S.
- **CHANGELOG and FEATURES.** Undecided whether the scripts and the
  corpus count as shipped tooling; nothing recorded. Effort: S once
  decided.
- **This report's own harvest.** Works: 22 rows in TODO_LIST §11 and 24
  lines in a new ROADMAP block, ledger appended below. Open: nothing.
- **Push.** 34 commits sit ahead of origin/main; nothing was pushed (the
  daemon never pushes and I was not asked).

## c) NOT STARTED

- The TODO_LIST.md struck-row repair (ten rows, corrupted since
  2026-11-04). Deprioritized: foreign data, and a repair from git history
  deserves a reviewed plan.
- A fail-closed freshness judge and a standing check for shifted struck
  rows — both are product changes with golden impact (see d2, d3).
- The purge collector's stray-pipe bullet (d3).
- Wiring the instruments into CI, a pre-commit leg or a flake check.
- Porting the instruments to Go verbs — parked on question 1.
- Formatting the four pre-existing unformatted files and marking the
  draft's Go block illustrative — foreign in-flight files.
- Skill feedback (library-deep-dive runbook, SKILLS repo) and the
  cross-project lessons for crush-config.
- Everything library-side: the cmdguard, go-codec and
  go-credentials-store follow-ups.
- Publishing the dashboard on the project website — never asked.

## d) TOTALLY FUCKED UP

- **d1. TODO_LIST.md's struck rows are structurally corrupted, and every
  standing gate is blind to it.** Ten rows in two shapes: nine read
  `| ~~ | <task> | <status> | <priority> | <effort> |` — the strike marker
  alone in the Task cell, every column shifted one left, the Evidence cell
  gone — and one row is struck _before_ its leading pipe (`TODO_LIST.md:87`),
  so it is no longer a table row at all. Severity: the living ledger has
  lost evidence cells; recoverable from history. Present since
  2026-11-04, ten rows already at this session's baseline — not caused
  here. The cadence gate reports "0 struck DONE rows" and the battery is
  green; only the on-demand `rows` verb sees it (exit 1, two tables). Root
  cause, reproduced in a scratch file: a strike written _before_ the row's
  leading pipe gives the row a surplus cell, and the table formatter then
  keeps that marker as the Task cell, shifts every column left and drops
  the Evidence cell. docs-gates' own `strike` verb is not the source — it
  writes inside the Task cell and refuses double strikes
  (`internal/docsstrike/strike.go:170`); the shape is the "strike by row
  number" anti-pattern AGENTS.md already warns about. Which session wrote
  it was not determined. Mitigation: none.
- **d2. The freshness gate is satisfiable by omission.** Of 12 unarchived
  reports, four (all of 2026-11-23) have no `f)` section and two have
  bullet-only `f)` sections from which zero items are extracted. Their
  next steps are forced by nothing: the 05-33 report's six follow-ups (the
  md5 exception, BuildFlow#115, the AGENTS distill) appear in neither
  TODO_LIST nor ROADMAP today. Severity: the gate that exists to force
  harvest passes vacuously. Root cause: the `f)` grammar plus numbered-only
  extraction, meeting free-form reports. Mitigation: manual harvest.
- **d2b. And when a report does carry numbered items, they pass on
  incidental vocabulary.** This very report was judged FRESH with "0
  unharvested open §f item(s)" before a single item was harvested.
  `--explain-item` shows why: the markdown-lint item passed on the words
  "on-demand" and "quality" while `markdown-lint` itself is absent from the
  living docs; the go-credentials-store item passed on "commits", "ahead"
  and "checkout" while the library's name is absent. Two matched tokens
  anywhere in TODO_LIST, ROADMAP or CHANGELOG, one of them merely below the
  generic bar, satisfy any item. Root cause: the quorum is counted
  corpus-wide instead of on one living-doc line, and my Impact, Effort and
  Category columns added shared vocabulary. Severity: the matcher cannot
  tell a harvest from a coincidence. The fifty items in section f) were
  therefore routed by hand, not by the gate.
- **d3. The purge collector would write a malformed bullet.** Its dry run
  drafts `- TODO_LIST purge (2026-11-24): | Write erraudit's own …` — a
  stray leading pipe from the row struck before its pipe — and targets a
  row whose status cell still reads as open.
- **d4. The BuildFlow findings gate has no signal.** It is permanently red
  on the documented md5 delta; the AGENTS.md size error stacks on top. An
  always-red gate cannot announce a new error.
- **d5. AGENTS.md grew from 223 to 228 lines** (linter: 229) against a 220
  ceiling. The memory rule of the audit skill requires the verdict there;
  I compressed my addition to seven lines and removed one duplicate, and
  it is still worse than I found it.
- **d6. My own mistakes this session.** Each was caught and reverted, each
  cost time: (i) a transcript fixer dropped the prompt prefix and
  corrupted 139 verify entries in the working tree — restored from commit
  `612e0e7`; (ii) I ran the formatter outside the devShell, so it fetched
  remote dprint plugins and reformatted four foreign files including a
  write-once status report — restored from explicit SHAs; (iii) I
  "corrected" two AGENTS.md citations from a mis-read, out-of-order `sed`
  batch and swapped them — the auditors had been right; (iv) my brief to
  the auditors carried a wrong line number and no fixed status vocabulary,
  which is why 18 reports needed reconciliation; (v) dashboard rows
  mis-sorted and a render trace generated from an outdated render — both
  caught by my own gates; (vi) two vision passes misread small text until
  I cropped at 2x; (vii) I pre-flighted and landed the TODO §10 and §11
  rows in compact bytes instead of the formatter's padded bytes — the
  exact lesson AGENTS.md records — and had to re-judge them in rewrite mode.
- **d7. Infrastructure.** Rate limits killed two of the first two
  auditor groups and one group twice more; three alternate providers had
  no credits (gemini, hyper, huggingface); glm-4.6v rejected the vision
  request above a 32768-token limit. The run went fully sequential.
- **d8. Pre-existing, noticed, untouched:** dprint-format warnings in four
  files (AGENTS.md, TODO_LIST section 6, the 05-40 status report, the
  BuildFlow#115 draft); an `md-go-validator` warning on that draft's Go
  block; the go-credentials-store checkout three commits ahead of upstream.

## e) WHAT WE SHOULD IMPROVE

Process and design, with the self-review answered inline.

- **Fix the brief before dispatch.** Pin the status vocabulary and require
  literal, pinned-commit transcripts up front. Impact: three repair passes
  (58 + 30 + 33 entries) would not have existed.
- **"In range" is not "right content".** Five citations pointed at valid
  lines of the wrong bullet. Impact: a reader trusts a range check that
  cannot see meaning. Fix: a content mode in the sweep.
- **Subset-tolerant matchers hide rot.** The first replay harness passed
  abbreviated transcripts. Fix: strict by default, as shipped.
- **Formatters only inside `nix develop`.** Fix: one line in
  `docs/agent-notes.md`.
- **dprint pads every table to its widest cell.** Impact: a pre-flight run
  on hand-written rows judges bytes the formatter then changes. Fix: format
  first, then pre-flight those bytes (rewrite mode when the rows exist).
- **Never read batched `sed` ranges as if ordered by argument.** Fix:
  `awk` with printed line numbers, or re-derive after every edit.
- **Run the product's own on-demand verbs on the living docs** (`rows` on
  TODO_LIST.md would have exposed d1 any day since 11-04).
- **What did I forget?** The three things in b): the scripts' names when
  compressing AGENTS.md, the `/tmp` helpers, and that nothing invokes the
  instruments.
- **What is stupid that we do anyway?** Writing reports free-form against
  a judge that only reads numbered items (d2).
- **Did I lie?** No deliberate lie. I made false statements and corrected
  them in session (d6). One suspicion — that a dry-run purge rewrites the
  land-parity stamp — I checked before reporting and it was false.
- **Did we remove something useful?** Auditors' commentary lines were
  dropped when transcripts were literalized (minor framing loss); the
  duplicate AGENTS.md bullet's nuance was carried over.
- **Scope creep?** Yes. The instruments and the routing came from the
  audit skill's done bars; literalizing every transcript went beyond the
  ask. It raised trust in the reports and roughly tripled the session.
- **Ghost systems.** (1) `docs-gates rows` detects d1 but nothing runs it
  on TODO_LIST.md — integrate it. (2) The two instruments exist but no
  gate calls them. (3) `--margins` and `--explain-item` are used; not
  ghosts.
- **Split brains created.** Dashboard matrix versus report headers (hand
  copies); the audited-subjects list in three places (the sweep's
  `LIB_NAMES`, the dashboard, the report files); the status-rubric
  sentence in 25 files; two independent console-block parsers. Fix: one
  manifest and a generated or checked matrix.
- **Tests.** No Go code changed, so the suites were not re-run and CI has
  not seen this work. The two scripts have no tests. Fix: stdlib
  `unittest`, or Go tests if they become verbs.
- **Type model.** A machine-readable sidecar per report (verdict, counts,
  version, triggers) would replace prose parsing with a typed source of
  truth.
- **Inherited claims.** I re-verified against sources rather than
  re-reading documents: 196 commands, 702 citations, every AGENTS.md
  citation re-derived after each edit — which is how the wrong-content
  class surfaced.

## f) Up to 50 things to get done next

Ranked by impact. A brainstorm beyond the first ten, not a commitment
list: routing follows the harvest ledger below.

| #  | Task                                                                                                                                                                                                                            | Impact   | Effort | Category      |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------ | ------------- |
| 1  | Battery panic guard: add a deferred `recover()` in the gate-capture wrapper so a panicking gate becomes a named FAIL row instead of crashing the whole run                                                                      | Critical | S      | Bug           |
| 2  | Config did-you-mean: lift the Levenshtein suggester into one shared helper and propose the right key on an unknown `.docs-gates.json` key                                                                                       | High     | S      | Feature       |
| 3  | Dedupe `writePreservingMode`: three hand-rolled copies (docspurge, docsarchive, docsstrike) collapse into one `gatecli` helper                                                                                                  | Medium   | S      | Cleanup       |
| 4  | Purge `--apply` write order: write the CHANGELOG before the TODO_LIST so a mid-purge crash cannot drop ledger rows                                                                                                              | Medium   | S      | Bug           |
| 5  | Finish the AGENTS.md distill: 229 lines against the 220 ceiling, 223 before this session                                                                                                                                        | High     | S      | Documentation |
| 6  | Repair the ten corrupted struck rows in TODO_LIST.md: nine whose strike marker sits alone in the Task cell with the Evidence cell dropped, one struck before its leading pipe; restore each row's cells from git history        | High     | M      | Bug           |
| 7  | Make the freshness judge fail closed: WARN on a report with no `f)` section or bullet-only next steps, and require a harvest's matched tokens to share one living-doc line — this report scored FRESH with zero items harvested | Critical | M      | Bug           |
| 8  | Fix the purge collector's drafted bullet: a row struck before its leading pipe yields a CHANGELOG line starting with a stray pipe, and the row's status still reads as open                                                     | Medium   | S      | Bug           |
| 9  | Add a standing-battery check for shifted struck rows: `docs-gates rows TODO_LIST.md` exits 1 today while the battery is green                                                                                                   | High     | S      | Feature       |
| 10 | Decide the instruments' home: keep `scripts/citation-sweep.py` and `scripts/replay-verify.py` as Python, port them to Go judge verbs, or move them to a sibling module (question 1)                                             | High     | M      | Architecture  |
| 11 | Wire the audit instruments into a quality gate: a devShell-only pre-commit leg or flake check that skips when the library checkouts are absent                                                                                  | High     | M      | Quality       |
| 12 | Port the instruments to Go as judge verbs with tests and goldens, once the home is decided                                                                                                                                      | Medium   | L      | Feature       |
| 13 | Make the scripts portable: replace the hardcoded `/home/lars/projects` paths with a `--projects-dir` flag or an environment variable                                                                                            | Medium   | S      | Quality       |
| 14 | Add stdlib `unittest` coverage for both Python scripts if they stay Python                                                                                                                                                      | Medium   | S      | Quality       |
| 15 | Provide `python3` inside the nix devShell: today the interpreter leaks in from the host profile                                                                                                                                 | Medium   | S      | Quality       |
| 16 | Teach `citation-sweep.py` to resolve bare follow-on cites such as `:188-189` after a filename — one such cite escaped every tool                                                                                                | Medium   | S      | Quality       |
| 17 | Add a content mode to the citation sweep that prints each cited line, so wrong-content cites can be reviewed                                                                                                                    | Medium   | M      | Quality       |
| 18 | Emit anchor-only output for living docs in verify transcripts (`rg -N`) so AGENTS.md and TODO_LIST edits stop drifting them                                                                                                     | Medium   | S      | Quality       |
| 19 | Ship the fleet-facts derivation as a script: semver-sorted tags per submodule, HEADs, upstream drift, module names read without the `head -1` comment bug                                                                       | Medium   | S      | Tooling       |
| 20 | Ship the transcript literalizer (`replay-verify --fix`, read-only guarded) that currently lives only in `/tmp`                                                                                                                  | Medium   | S      | Tooling       |
| 21 | One audited-subjects manifest consumed by the sweep's library list, the dashboard builder and the facts script — three lists today                                                                                              | Medium   | M      | Architecture  |
| 22 | Generate the dashboard's verdict matrix and trigger table from the report headers, or add a consistency check, to remove the hand-copied split brain                                                                            | Medium   | M      | Quality       |
| 23 | Point README.md and `docs/agent-notes.md` at `docs/review/libs/` and the two scripts                                                                                                                                            | Medium   | S      | Documentation |
| 24 | Decide and record whether the audit scripts and corpus belong in CHANGELOG `[Unreleased]` and FEATURES.md                                                                                                                       | Low      | S      | Documentation |
| 25 | Run dprint inside the devShell on the four pre-existing unformatted files (AGENTS.md, TODO_LIST section 6, the 05-40 status report, the BuildFlow#115 draft)                                                                    | Low      | S      | Cleanup       |
| 26 | Mark the Go block in `docs/drafts/2026-11-23_buildflow-nix-flake-update-blind-repair.md` illustrative so `md-go-validator` stops warning                                                                                        | Low      | S      | Documentation |
| 27 | Give the findings gate a signal again: a library-policy scoped exception for the deliberate md5 fingerprint so `buildflow` stops exiting 69 on a documented delta                                                               | High     | M      | Quality       |
| 28 | Document in `docs/agent-notes.md` that formatters must run inside `nix develop`: outside it dprint fetches remote plugins and formats a different tree                                                                          | High     | S      | Documentation |
| 29 | Document the pre-flight `--replace` recipe — judge formatted rows by removing them from the replay — in `docs/agent-notes.md`                                                                                                   | Medium   | S      | Documentation |
| 30 | Run the product's on-demand `rows` verb over `docs/review/libs/` and this report; it was never run on the audit corpus                                                                                                          | Medium   | S      | Quality       |
| 31 | Run the on-demand `markdown-lint` step over the audit corpus                                                                                                                                                                    | Low      | S      | Quality       |
| 32 | Record this session's cross-project lessons in crush-config `references/lessons.md`: out-of-order `sed` reads, subset-tolerant matchers, prompt-prefix parsing (repo: crush-config)                                             | Medium   | S      | Documentation |
| 33 | Feed the audit runbook: fix the status vocabulary, require literal transcripts and pinned-commit commands, add citation content verification (repo: SKILLS)                                                                     | High     | S      | Documentation |
| 34 | Push or deliberately park the three local commits ahead of upstream in the go-credentials-store checkout (repo: go-credentials-store)                                                                                           | Medium   | S      | Fleet         |
| 35 | Export the Levenshtein suggester or ship a zero-dependency core so the fleet shares one implementation (repo: cmdguard)                                                                                                         | Medium   | M      | Fleet         |
| 36 | Decide unknown-field rejection and case-sensitivity knobs for the JSON codec; the audit found adoption would remove capability (repo: go-codec)                                                                                 | Medium   | M      | Fleet         |
| 37 | Make the dprint plugin source explicit: refuse or warn loudly when `BUILDFLOW_DPRINT_PLUGINS` is unset (repo: BuildFlow)                                                                                                        | Medium   | S      | Fleet         |
| 38 | Land the BuildFlow#115 fix for the `nix-flake-update` blind repairer (repo: BuildFlow)                                                                                                                                          | High     | M      | Fleet         |
| 39 | Give AGENTS.md size one owner: conclude the go-structure-linter migration so the 220 rule stops duplicating this repo's agents gate (repo: go-structure-linter)                                                                 | Medium   | M      | Fleet         |
| 40 | Clamp `max_tokens` per model so vision dispatch on glm-4.6v stops failing above 32768 (repo: crush-config)                                                                                                                      | Low      | S      | Tooling       |
| 41 | Top up or accept single-provider risk: three alternate model providers had no credits during this run                                                                                                                           | Medium   | S      | Ops           |
| 42 | Re-audit the 23 subjects when a flip trigger fires: a cmdguard zero-dependency core, relaxed stdlib-only, or the first sibling serving module                                                                                   | Medium   | L      | Fleet         |
| 43 | A `docs-gates lib-audit` verb that derives the fleet manifest and runs the battery across a fleet of repos                                                                                                                      | Medium   | L      | Feature       |
| 44 | A gate that replays the verify blocks of every doc, not only audits, so stale commands surface                                                                                                                                  | High     | L      | Feature       |
| 45 | A content-anchored citation convention (`path:line "quoted phrase"`) so sweeps verify meaning, not just ranges                                                                                                                  | Medium   | M      | Quality       |
| 46 | A machine-readable sidecar per report (verdict, counts, version, triggers) so dashboards stop parsing prose                                                                                                                     | Medium   | M      | Architecture  |
| 47 | Placeholders plus a recorded environment for verify transcripts so another machine can replay them                                                                                                                              | Low      | S      | Quality       |
| 48 | A `--diff` mode on `replay-verify.py` that prints the line-level difference on mismatch                                                                                                                                         | Low      | S      | Quality       |
| 49 | Publish `docs/review/libs/index.html` on the project website                                                                                                                                                                    | Low      | M      | Documentation |
| 50 | A scheduled job that re-runs the sweep and replay across audit corpora and opens a TODO row on drift                                                                                                                            | Medium   | M      | Feature       |

## g) Questions I cannot answer myself

1. **Where should the two audit instruments live** — Python scripts as
   today, Go judge verbs inside docs-gates, or a sibling module? They
   judge and never write, which fits this binary's charter, but new verbs
   mean help pins, goldens and provider rows, and that is a product
   boundary only you can draw. I tried to settle it by reasoning about the
   stdlib-only contract and could not.
2. **May I finish the AGENTS.md distill myself?** Reaching 220 means
   condensing about nine lines of notes I did not author. I removed one
   exact duplicate and compressed my own addition to the minimum; the
   rest is your text, or the 220 ceiling is yours to retire in the
   go-structure-linter migration decision.
3. **Which of these may I execute without further check-ins** — the four
   TODO §10 hardening fixes (gate code, possible golden re-pins), the
   TODO_LIST.md struck-row repair from git history, or the fail-closed
   freshness WARN? Default if unanswered: none; they stay routed.

## Harvest ledger (2026-11-24)

Routed by hand: the freshness judge called this report FRESH before
anything was harvested (d2b). Pre-flight over the 46 new content lines:
SAFE, 36 reports, nothing flips to DRIFT; all 46 landed verbatim. The 26
TODO rows were then re-judged in rewrite mode after the formatter padded
them (SAFE again, land-parity OK). Every
premise was re-checked against TODO_LIST.md and ROADMAP.md on 2026-11-24
before rowing (topic greps plus the gate and instrument outputs above).

| Source     | Disposition  | Destination / reason                                                                                                |
| ---------- | ------------ | ------------------------------------------------------------------------------------------------------------------- |
| 11-24 §f1  | existing row | TODO_LIST §10 row 1, routed 2026-11-23 from the audit reports — premise re-checked 2026-11-24                       |
| 11-24 §f2  | existing row | TODO_LIST §10 row 2, routed 2026-11-23 from the audit reports — premise re-checked 2026-11-24                       |
| 11-24 §f3  | existing row | TODO_LIST §10 row 3, routed 2026-11-23 from the audit reports — premise re-checked 2026-11-24                       |
| 11-24 §f4  | existing row | TODO_LIST §10 row 4, routed 2026-11-23 from the audit reports — premise re-checked 2026-11-24                       |
| 11-24 §f5  | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f6  | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f7  | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f8  | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f9  | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f23 | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f25 | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f26 | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f27 | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f28 | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f29 | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f30 | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f31 | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f32 | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f33 | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f34 | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f35 | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f36 | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f37 | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f38 | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f39 | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f40 | new row      | TODO_LIST §11 (bounded, this repo or repo-tagged) — premise re-checked 2026-11-24                                   |
| 11-24 §f10 | new row      | ROADMAP, Fleet library audit follow-ups (decision or design fuel, not a commitment) — premise re-checked 2026-11-24 |
| 11-24 §f11 | new row      | ROADMAP, Fleet library audit follow-ups (decision or design fuel, not a commitment) — premise re-checked 2026-11-24 |
| 11-24 §f12 | new row      | ROADMAP, Fleet library audit follow-ups (decision or design fuel, not a commitment) — premise re-checked 2026-11-24 |
| 11-24 §f13 | new row      | ROADMAP, Fleet library audit follow-ups (decision or design fuel, not a commitment) — premise re-checked 2026-11-24 |
| 11-24 §f14 | new row      | ROADMAP, Fleet library audit follow-ups (decision or design fuel, not a commitment) — premise re-checked 2026-11-24 |
| 11-24 §f15 | new row      | ROADMAP, Fleet library audit follow-ups (decision or design fuel, not a commitment) — premise re-checked 2026-11-24 |
| 11-24 §f16 | new row      | ROADMAP, Fleet library audit follow-ups (decision or design fuel, not a commitment) — premise re-checked 2026-11-24 |
| 11-24 §f17 | new row      | ROADMAP, Fleet library audit follow-ups (decision or design fuel, not a commitment) — premise re-checked 2026-11-24 |
| 11-24 §f18 | new row      | ROADMAP, Fleet library audit follow-ups (decision or design fuel, not a commitment) — premise re-checked 2026-11-24 |
| 11-24 §f19 | new row      | ROADMAP, Fleet library audit follow-ups (decision or design fuel, not a commitment) — premise re-checked 2026-11-24 |
| 11-24 §f20 | new row      | ROADMAP, Fleet library audit follow-ups (decision or design fuel, not a commitment) — premise re-checked 2026-11-24 |
| 11-24 §f21 | new row      | ROADMAP, Fleet library audit follow-ups (decision or design fuel, not a commitment) — premise re-checked 2026-11-24 |
| 11-24 §f22 | new row      | ROADMAP, Fleet library audit follow-ups (decision or design fuel, not a commitment) — premise re-checked 2026-11-24 |
| 11-24 §f24 | new row      | ROADMAP, Fleet library audit follow-ups (decision or design fuel, not a commitment) — premise re-checked 2026-11-24 |
| 11-24 §f41 | new row      | ROADMAP, Fleet library audit follow-ups, Owner action — premise re-checked 2026-11-24                               |
| 11-24 §f42 | new row      | ROADMAP, Fleet library audit follow-ups, Longer term — premise re-checked 2026-11-24                                |
| 11-24 §f43 | new row      | ROADMAP, Fleet library audit follow-ups, Longer term — premise re-checked 2026-11-24                                |
| 11-24 §f44 | new row      | ROADMAP, Fleet library audit follow-ups, Longer term — premise re-checked 2026-11-24                                |
| 11-24 §f45 | new row      | ROADMAP, Fleet library audit follow-ups, Longer term — premise re-checked 2026-11-24                                |
| 11-24 §f46 | new row      | ROADMAP, Fleet library audit follow-ups, Longer term — premise re-checked 2026-11-24                                |
| 11-24 §f47 | new row      | ROADMAP, Fleet library audit follow-ups, Longer term — premise re-checked 2026-11-24                                |
| 11-24 §f48 | new row      | ROADMAP, Fleet library audit follow-ups, Longer term — premise re-checked 2026-11-24                                |
| 11-24 §f49 | new row      | ROADMAP, Fleet library audit follow-ups, Longer term — premise re-checked 2026-11-24                                |
| 11-24 §f50 | new row      | ROADMAP, Fleet library audit follow-ups, Longer term — premise re-checked 2026-11-24                                |

Nothing was declined. Items 5, 27 and 38 also give living-doc rows to
three follow-ups of the 2026-11-23 05-33 report that its bullet-style §f
never got harvested (the AGENTS distill, the md5 exception, BuildFlow#115).
The three questions in section g) are questions, not tasks, and were not
rowed.
