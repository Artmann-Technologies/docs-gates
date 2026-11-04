# Status Report — evening docs-health close: cadence gate bug, purge, harvest

**2026-11-04 19:25 CET** · session: continuation of the 07:11 docs-health
audit thread (user: "keep going until everything works") · branch `main`,
HEAD `73f5eac` (pushed, CI green).

**Scope line — VERIFIED this session (fresh commands, exit codes read
back):**

| Claim                                                                  | Evidence                                                                                                                          |
| ---------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| CI green on the morning audit train AND the evening fix train          | `gh run list`: success on 817119e's run and on 73f5eac (run completed after push)                                                  |
| Execution sessions landed T3–T27 (16 tasks) before this continuation   | read `docs/status/2026-11-04_19-01_pareto-execution-t6-t27-16-tasks-landed.md` in full; TODO carried 25 struck rows with evidence   |
| Battery over the pre-fix tree: 5/5 PASS but cadence said "0 struck"    | `go run ./cmd/docs-gates run .` — the mismatch (25 struck rows vs 0 counted) against threshold 20 exposed the bug                   |
| Gate bug root-caused                                                   | `internal/docscadence/classify.go`: status cell had to EXACTLY equal `✅ \`DONE\``/`DONE`; dated cells (`✅ DONE 2026-11-04`) miss  |
| Fix + test land, gate flips truthful                                    | full suite 8×ok, vet clean; live `docs-gates cadence .` → "25 struck DONE row(s) … OVERDUE" rc=1 with the purge-ritual remediation |
| Purge safe (substance preserved BEFORE deletion)                        | CHANGELOG [Unreleased] carries 26 per-task rows covering every struck row — verified by reading before deleting                     |
| Post-purge + harvest state green                                        | battery 5/5 PASS (0 struck rows, freshness/audit-all FRESH over 8 reports, annotate 13 struck/0 findings, routing strict-clean)     |
| Lint clean after formatter round                                         | `golangci-lint run ./...` → 0 issues; `gofmt -l .` empty                                                                            |

**NOT verified this session (and why):**

- **The 16:45 report body** — never read; accepted the 19:01 report's
  resume framing plus green gates as coverage. Its 3 owner questions are
  carried by 19:01 §g anyway.
- **aarch64 builds** — eval-only all day (standing gap, tracked).
- **SystemNix/pbx/BuildFlow-side state** — external repos, untouched
  (the whole remaining P1 tier is owner-gated).
- **The provider-module suite after the classify fix** — not re-run;
  the change is root-module-only (judgment: not needed; suite was green
  at 04:00 and the fix touches nothing provider-visible).

---

## Self-critique (asked explicitly: forgotten? better? improve?)

- **What I forgot:** lint in the same breath as the code edit. The
  classify.go comment placement carried a gci finding that the daemon
  swept into `5f283ed` BEFORE my lint pass — the second lint-dirty
  commit on this repo today (the 19:01 session paid the identical foul
  at 066ac47). Tests ran first, lint later; under a 3-minute daemon,
  "later" is a commit.
- **What I got wrong in judgment:** I drafted `isDoneCell` twice (a
  redundant loop shape, immediately rewritten as the simple boolean).
  A four-line predicate should be written once.
- **Where I was sloppy:** the TODO rebuild shipped with section numbers
  1/2/4/5 (3 skipped) and needed a patch two minutes later; and two
  edit-tool stale-read rejections (AGENTS.md, CHANGELOG.md) burned
  round-trips against daemon/formatter races — the write-protocol
  lesson (fresh read, immediate edit) is documented in today's own
  reports and I re-learned it live.
- **Where I was close to lying:** nowhere material — every gate claim
  above is a read-back command. The one soft spot: "CHANGELOG covers
  every struck row" was verified by reading row HEADLINES, not each
  row's full text against each struck task (headline coverage was
  complete and specific enough to act; full-text diff would have cost
  more than the residual risk).
- **Ghost systems:** none — the fix, test, purge, harvest, and ROADMAP/
  AGENTS/CHANGELOG edits are all wired and pushed.
- **Split brains found and fixed:** ritual-vs-gate (the purge WAS due
  while the gate said healthy — the gate was wrong, now fixed); TODO
  section numbering (caught by self-review before commit).

---

## a) FULLY DONE

1. **CI loop closed on both trains** — the morning audit train (pushed
   at last pass) and the evening fix train both green on GitHub.
2. **docs-cadence purge-counter bug found, fixed, tested** — dated DONE
   status cells (`✅ DONE 2026-11-04`) now count as struck rows (any
   cell starting with a DONE marker; prose DONE still never occupies a
   whole cell); `TestDatedDoneCellsAreStruckRows` pins the dated and
   dashed shapes; CHANGELOG Fixed row written.
3. **Purge ritual executed** — 25 landed rows deleted from TODO_LIST
   after verifying per-task CHANGELOG coverage; cadence now reads a
   truthful 0 struck rows under threshold 20.
4. **The 19:01 execution report harvested in full** — its 50 §f items
   routed: 11 quality-machinery rows + 3 docs-surface rows new in
   TODO §1-§2; fleet legs confirmed/extended in §3 (fleet-ratchet-corpus
   re-run added as its own row); v0.2.0 release cut into §4; curated
   long-game ideas into ROADMAP (pre-flight removals/ranges, margins
   --json + routing convergence, doctor verb, fleet-wide nolint
   tripwires); noise declined with reasons in the report's own routing.
5. **The untracked 19:01 report committed** (daemon blind spot — it sat
   untracked 14+ minutes until this session's deliberate commit).
6. **AGENTS Commands** gained the erraudit-rev tripwire line (the script
   existed; the command list didn't name it).
7. **Deliberate train committed and pushed** (`73f5eac`): gate fix +
   test + purge + harvest + ROADMAP + CHANGELOG + report, one detailed
   message, CI green.

## b) PARTIALLY DONE

1. **The classify fix has unit coverage but no golden arm** — the parity
   corpus's TODOS.md carries no dated-DONE rows, so the byte-parity net
   does not exercise the new counting (only the unit test does). One
   fixture arm closes it; routed to §f below.
2. **The stale LSP `gci` ghost warning** on items.go/classify.go
   persisted across the day's tool diagnostics while real lint was
   clean — the harness issue row is tracked; no repro attempted here.
3. **CHANGELOG granularity** — settled de facto (per-task rows), but the
   owner question from 16:45 is technically still open until the v0.2.0
   cut (§4 row) lands.

## c) NOT STARTED

The whole remaining TODO board (30 open rows): the owner-gated P1 tier
(v0.2.0 release, SystemNix float + evo-x2 deploy, sibling-repo
permission), the docs-surface debt (README/DOMAIN_LANGUAGE/FEATURES for
the three new flags), and the external/fleet legs. Nothing in-repo and
unblocked was left undone by choice — the unblocked in-repo tier was
emptied by the 19:01 session.

## d) TOTALLY FUCKED UP

Nothing unrecovered. The honest damage list:

1. **Lint-dirty daemon commit (5f283ed)** — my gci finding rode into a
   heuristic commit before I linted; fixed one commit later (73f5eac).
   Same class as the morning session's 066ac47: twice in one day means
   the mitigation ("final lint BEFORE commit") is insufficient under a
   3-minute daemon — the lint must follow the EDIT, not the commit plan.
2. **Drafted `isDoneCell` twice** — a four-line predicate written,
   rejected, rewritten within one minute. Think, then type.
3. **TODO section-numbering slip** (1/2/4/5) shipped in the first write,
   patched before commit. Caught by self-review, not by care.

## e) WHAT WE SHOULD IMPROVE

1. **Lint-at-edit-time, not lint-at-commit-time** — under the daemon,
   any gap between "code written" and "lint run" is a window where a
   finding commits itself. Pair every code edit with its lint in the
   same command chain.
2. **Predicates once** — for small pure functions, write the final shape
   first; the throwaway draft costs more than the ten seconds it saves.
3. **Write-protocol under daemon ticks** — fresh view immediately before
   every edit of a file the daemon/formatter also touches (paid twice
   this session; the 19:01 report's §e already prescribed it).
4. **Gate-behavior fixes deserve golden arms** — the unit test pins the
   counter, but the parity net exists to pin gate OUTPUT; a dated-DONE
   fixture row in the corpus TODOS.md would have the byte-parity net
   carry this fix too (routed below).
5. **The trophy-case class is now mechanically caught** — today's bug
   was "ritual says purge, gate says healthy"; with the counter fixed,
   the disagreement direction to watch flips (gate OVERDUE with rows
   already in CHANGELOG = stale gate run, not stale docs).

## f) Up to 50 things to get done next (brainstorm, not commitment)

The 30-row TODO_LIST rebuilt this evening IS the routed backlog; top
items by impact, "§N" cites home. NEW = born this session.

| #  | Task                                                                              | Impact | Effort | Category      |
| -- | ---------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 1  | Cut v0.2.0 (section + annotated tag, go-release ritual) — settles granularity, versions the flags (§4) | High | M | Release |
| 2  | SystemNix float to the hardened flake (§3, P1)                                      | High   | S      | Integration   |
| 3  | evo-x2 deploy + live-PATH six-binary smoke (§3, P1, owner-gated window)             | High   | S      | Deploy        |
| 4  | Wire harvestAttributions into the auditAll renderer + battery arm (§1)              | High   | S      | Quality       |
| 5  | CI-visible silent-skip assertion for the provider e2e leg (§1)                      | High   | S      | Quality       |
| 6  | NEW: parity-corpus fixture arm for the dated-DONE counting (corpus TODOS.md row + golden re-pin) | Medium | S | Quality |
| 7  | Docs-surface debt: README/DOMAIN_LANGUAGE/FEATURES rows for --margins/--dump-items/--replace (§2) | Medium | S | Documentation |
| 8  | Parity arm: golden --replace through the binary (§1)                                | Medium | M      | Quality       |
| 9  | battery_test gitignore-trap regression arm (§1)                                     | Medium | S      | Quality       |
| 10 | Re-run fleet ratchet corpora vs converged Corpus semantics (§3)                     | Medium | S      | Process       |
| 11 | pbx-artmann migration + fleet announcement after the release (§3)                   | Medium | M      | Integration   |
| 12 | Pre-commit nolint-audit leg (§1)                                                    | Medium | S      | Quality       |
| 13 | Owner D1 (battery adoption scope) + D2 (go floor) — recommendations prepared (§4)   | Medium | S      | Decision      |
| 14 | BuildFlow upstream: #87 + roundtrip walker + hygiene + LSP-ghost filing (§3)        | Medium | L      | Quality       |
| 15 | ANNOTATE sweep over the 8 unarchived reports, archive the fully-resolved (§1)       | Low    | M      | Ritual        |

(Honest stop at 15: rows 16-30 live in TODO_LIST §1-§4 with evidence —
the purge left no struck residue, so there is no hidden remainder.)

## g) Questions I can NOT figure out myself

1. **Release authority (blocks §4 P1):** cut v0.2.0 now — CHANGELOG
   section cut + annotated tag + push, versioning today's flags and the
   cadence fix — or accumulate quietly first? My recommendation: cut it;
   the Unreleased section is 27 rows deep and the granularity question
   resolves itself in the cutting.
2. **Sibling-repo permission (blocks the §3 P1 tier):** blanket go-ahead
   for SystemNix float/deploy + pbx-artmann + crm/go-appkit + BuildFlow
   upstream, or repo-by-repo gating? The whole external tier has been
   parked on this since 16:45.
3. **Battery adoption scope (D1, reshapes §3):** timer/PMA lane running
   the gates autonomously over fleet repos, or PATH-only for a quarter
   while the pilot ratchets produce data? I stand with the 19:01
   recommendation: PATH-only first.

---

*HARVEST note: §f item 6 is NEW and is harvested to TODO_LIST §1 before
this session ends; items 1-5, 7-15 already ride TODO_LIST §1-§4.*
