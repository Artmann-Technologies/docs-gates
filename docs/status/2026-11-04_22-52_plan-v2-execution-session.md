# Plan v2 execution — 14 tasks landed, one gate-surface train, one corruption caught

**2026-11-04 22:52 CET** · baseline `4cc99af` (sweep-restored tree; battery
5/5 PASS, annotate CLEAN) · input: `docs/planning/2026-11-04_19-28_…-v2.md`
(24 tasks / 94 micro-tasks ≈ 22.75 h). This session took the
ZERO-PERMISSION critical path (the owner gates — release authority,
sibling permission, deploy window — remain unanswered, asked 4× today).

## a) 5-number summary

- **Verdicts:** battery 5/5 PASS; root suite 8/8 ok; provider suite ok;
  golangci 0 issues both modules; `nix flake check --all-systems` all-pass.
- **Tasks:** 14 of plan v2's 24 touched — T4 T5 T6 T9 T10 T12 T16 T18 T19
  landed; T20 root-caused (NOT filed); T23+T24-prep decision doc written;
  T17 attempted, corrupted, reverted, scoped down.
- **New fixtures/net:** parity goldens grew from 9 to 12 pinned
  invocations (`margins`, `dump-items`, `dump-items-refusal`), the
  corpus grew by `corpus-replace` + a dated-DONE struck row + a struck
  report item; the parity harness gained per-gate expected exit codes and
  corpus-root normalization.
- **New gates:** a battery regression arm for the `.gitignore`
  `reports/` trap (tracked + `--no-index` legs, red-controlled live), an
  e2e-skip sentinel + decision table in the provider module, a commit-time
  `nolint-audit` pre-commit leg (git-hooks.nix, sandboxed flake check +
  devShell install), and `--dump-items` now explains struck rows.
- **Incidents:** one self-inflicted corruption (strike sweep by row
  number wrecked three report ledgers; annotate hit 31 findings; reverted
  byte-identical in `ef14e11`, battery green) — full lesson in AGENTS.md.

## b) Board summary

- Landed rows struck in TODO_LIST tonight: §1 ×9, §2 ×3, §3 ×1 (LSP
  ghost row closed as root-caused-not-filed). Open rows: 32 → 19.
- Three NEW TODO rows' evidence updated in place (ANNOTATE sweep scoped
  down; the two §4 owner decisions now point at the decision doc).
- Owner questions unchanged: release authority (T1), sibling permission
  (T7/T8/T11/T14/T15/T21), deploy window (T3). The decision preps for
  battery adoption + go floor are written and decision-ready:
  `docs/planning/2026-11-04_21-02_owner-decision-preps-D1-D2.md`.

## c) Notable finds

- **go1.26+ compile-rejects mixing literals with a slice spread in one
  variadic call** (`f("a", "b", slice...)` = `too many arguments`;
  proven on both pinned toolchains, hits even `fmt.Println`). Build argv
  with append first. AGENTS.md non-obvious fact.
- **`git check-ignore` never reports TRACKED files as ignored** — the
  T9 red control proved the plain form blind on committed fixtures;
  `--no-index` is mandatory there. `git ls-files` prints paths relative
  to the INVOKING cwd, not the repo root.
- **The LSP ghost is a cache artifact, not a harness bug:** the
  `golangci_lint_ls` wrapper's persistent `GOLANGCI_LINT_ANALYSIS_CACHE`
  served `undefined: <renamed symbol>` across restarts; clearing the
  cache dirs cured it. An upstream issue was draft-ready and REFUTED by
  the cold-cache experiment (verify-before-filing Gate 4 did its job).
- **The provider e2e leg's skip arms are now unforgeable:** ledger +
  sentinel + decision table — only the two sanctioned sandbox reasons can
  skip; anything else fails the suite.

## d) Deviations

- T17 (ANNOTATE sweep) is scoped DOWN, not done: the §f ledgers hold
  28-49 open rows each (they are harvest fuel, correctly routed), no
  report is fully-resolvable, and the first sweep attempt proved the
  mechanics corrupt under the daemon's ~3-min commit cadence. Safe
  mechanics documented (unique task-text anchors, `[^|]*` cell bounds,
  annotate-gate after every batch). Redo at cadence pressure.
- T20 filed nothing, deliberately: the verified root cause is local
  (host wrapper cache), so an upstream issue would have been a premise
  error. The row is closed on that evidence.
- The daemon carried parts of every train (expected); deliberate commits
  where the narrative matters: `8a6435f` (T4 ritual), `d92f7f8` (T5),
  `ff7cfcd` (T6), `94350f5` (T9), `510ff58` (T10), `87d34fe` (T12
  ritual), `4777163` (T19), `ef14e11` (corruption revert).

## e) Verification

- Red controls, live: gitignore arm (swallowed fixture → FAIL naming
  itself; untracked fixture → tracked-leg FAIL), e2e sentinel
  (decision-table UNACCOUNTED rows), pre-commit hook (planted stale
  directive → FAIL naming "1 stale"), rewrite-mode golden (same
  candidate SAFE without `--replace`, UNSAFE with).
- Parity: goldens re-pinned once per intentional change and hand-diffed
  line by line (margins/dump/cadence/annotate/battery diffs are exactly
  the contract changes described above; freshness/audit-all/routing
  untouched).
- Dogfood: battery re-run after every doc edit (kept green throughout,
  including over the corruption — which is exactly why the annotate
  gate's 31 findings were the catching mechanism, not the battery).

## f) Next tasks (ranked; HARVEST fuel)

1. Cut v0.2.0 (T1) — 27+ Unreleased rows now; the release gate is still
   the owner's one-liner. Everything in the 1% tier rides this.
2. Owner answers (4th ask): release authority, sibling permission,
   deploy window — T2/T3/T7/T8/T11/T13/T14/T15/T21 all wait on them.
3. Fleet ratchet corpora re-run (T7) — the Corpus semantics converged
   again today (dated-DONE classify + routing glob prep); crm/go-appkit
   ratchets never re-run against it.
4. Sweep-safe ANNOTATE pass per report on demand (T17 residue): only
   with the AGENTS.md safe mechanics; only rows with demonstrable
   landing evidence.
5. BuildFlow triage batch (T14) — 9-tools history, preflight
   attribution, binary rebuild, `buildflow format` once.
6. docs-surface follow-up: README example for `--dump-items` refusal
   shape is covered; next doc debt is the fleet skill cross-check (T21,
   sibling-gated).
7. Harvest this report's §f into TODO_LIST before session end (done —
   see the TODO_LIST rows this session added/updated).

## g) Questions for the owner

1. **Release authority for v0.2.0** (4th ask): may this repo cut the
   annotated tag? The Unreleased section is release-ready; the decision
   preps doc also needs your D1/D2 one-liners.
2. **Sibling permission** for the fleet legs (SystemNix float/deploy,
   pbx migration, BuildFlow repo triage, pilot drains, crush-config
   skill check).
3. **D1/D2 one-liners** — the matrix is in the decision preps doc; a
   reply of "PATH-only Q1; floor 1.27 with v0.2.0" unblocks everything.

_Point-in-time snapshot; annotate, never rewrite. Verschlimmbesserung
guard: T17's scoping and T20's non-filing are explicit judgment calls
with evidence, not silently dropped work._
