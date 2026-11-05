# Fleet Dogfood — First External Verdicts (go-health, CV) and the audit-all Perf Cliff

2026-11-05 02:15 — docs-gates, session report. Scope: this session ran
the battery over two fleet repos on request and analyzed what came back.
No code in this repo changed; the deliverables are verdicts, a
performance diagnosis, and this report's harvest.

**Verified:** full battery `run` over go-health (complete matrix, exit 1);
freshness / annotate / routing --strict / cadence run individually over
CV with battery-identical argv (battery's own audit-all leg was killed
after ~15 min of no progress); ps forensics over every docs-gates
process on the box. **NOT verified:** CV audit-all (never completed —
see d); whether go-health/CV carry `.docs-gates.json` ratchet pins (no
config warnings printed, but presence was not checked directly; the
FAIL rows are definitive either way, since ratchets print RATCHET and
these printed findings); nothing in either target repo was modified.

## a) FULLY DONE

- **go-health battery, complete matrix (exit 1):** freshness PASS
  (newest report clean); audit-all FAIL — 32 unharvested open items
  across the 66-report corpus; annotate FAIL — 376 findings over 80
  files / 1,794 struck items; routing FAIL — 4 unrouted rows of 36
  unarchived scanned; cadence FAIL — 36 unarchived (threshold 15),
  oldest 27 days (threshold 21), plus 5 half-struck DONE rows.
- **CV gates, individually (battery argv mirrored):** freshness PASS on
  the 02:02 report (0 unharvested); routing FAIL — 1 unrouted row of
  370 scanned; cadence FAIL — 370 unarchived, oldest 34 days;
  annotate FAIL — 717 findings over 1,841 files / 6,015 struck items
  (F/spacing dominates, including `scripts/*.sh`).
- **Root-cause read:** both repos are red for the same driver — archive
  debt. Everything else (unharvested tails, half-struck rows) is
  downstream of reports never being closed out. go-health's tail even
  repeats itself: the "CV internal/agent sqlite restart-durability" and
  "re-verify architecture-gate.yml" obligations each appear in TWO
  reports — rediscovered, never harvested.
- **Perf-cliff diagnosis:** audit-all's cost is roughly
  reports × items × tokens × living-doc lines. CV is 1,643 reports
  (370 unarchived + 1,271 archived) against ~2,000 TODO/ROADMAP/
  CHANGELOG lines — ~25× go-health's report count and ~4× its living
  corpus, turning a 2-minute gate into a multi-hour one.
- **Forensics:** four foreign docs-gates processes found on the box —
  a CV run wedged since YESTERDAY 19:15 (7 h, 0% CPU), webphone runs
  at 2 h (0%) and active (2% CPU), and a BuildFlow run at 1.5 h (0%).
  Yesterday's session already hit this wall; mine is the second
  occurrence.

## b) PARTIALLY DONE

- **CV verdict matrix:** four of five gates have verdicts, but they
  came from individual verb invocations, not the battery path; the
  audit-all leg was killed and its verdict is unknown. The mirror used
  battery-identical argv, so coverage is equivalent minus that one
  gate.
- **Findings durability:** everything above lived only in chat until
  this report — the go-health/CV rows below are routed here, but their
  OWN repos' TODO_LISTs have not been touched (not asked, cross-repo).

## c) NOT STARTED

- Harvesting the go-health/CV reds into those repos' TODO_LISTs or
  running their archive rituals (their sessions' work; sibling
  permission is a standing owner gate).
- Any audit-all performance fix in this repo (diagnosed, not designed).
- Disposal of the wedged foreign processes (not mine to kill).
- CV's two HTML status reports outside the .md scan (gate WARN; their
  obligations unrouted).

## d) TOTALLY FUCKED UP (session mistakes, owned)

1. **Launched the CV battery without sizing the corpus.** go-health's
   ~2-minute run was the cost signal in hand; a ten-second
   `ls | wc -l` (372 + 1,271 reports) would have predicted hours. I
   started a multi-hour job silently in the background and set no
   duration expectation — the user had to ask "why is this taking
   forever" before I investigated.
2. **Broken exit-code capture in all four CV gate runs:** the shape
   `cmd 2>&1 | tail -N; echo RC=$?` prints TAIL's exit code. All four
   rows showed RC=0 while three gates FAILED. The verdicts I reported
   were read from output text (correct), but the capture was broken and
   unflagged.
3. **Burned the user's time polling instead of diagnosing:** two
   blocking job_output waits before running the 5-second ps forensics
   that actually answered the question.
4. **Misleading table cell:** my verdict table printed "CV freshness
   PASS*" — at battery time it was FAIL (7 unharvested); it flipped
   only because CV's daemon wrote a newer (clean) report at 02:02,
   mid-session. The asterisk said so, but the cell led with PASS.
5. **Config presence in the targets assumed, not checked:** no
   `.docs-gates.json` lookup was made in either repo before claiming
   "defaults ran". The FAIL rows are definitive regardless, but the
   claim rested on an unchecked assumption.

## e) WHAT WE SHOULD IMPROVE

- **Size before launch:** corpus counts are one `ls` away; any gate
  run over an unfamiliar repo gets counted first, and the launch
  message carries a duration estimate.
- **Shell discipline: never pipe-then-`$?`** — capture PIPESTATUS or
  drop the pipe. The repo's own scripts know this; my ad-hoc calls
  didn't.
- **Battery UX gap (product):** a gate that runs for hours prints
  nothing after its start line — wedges are invisible without ps
  forensics. A heartbeat/elapsed line and a per-gate timeout are cheap
  and would have saved this whole session's confusion.
- **Point-in-time gates under daemons:** freshness flipped FAIL→PASS
  between two runs minutes apart because the target repo's corpus
  changed underneath. When gating daemon-active repos, record the
  newest-report identity alongside the verdict.

## f) Next things (this repo's rows harvested to TODO_LIST §2d; sibling rows stay here, scoped to their own repos)

1. audit-all perf: hoist the living-docs DF/line-match index to once-per-run — it is recomputed per report today — and benchmark over the go-health corpus before/after.
2. audit-all short-circuit: a report with zero open items skips token matching entirely.
3. Battery heartbeat: per-gate elapsed-time progress line so a wedged or crawling gate is visible in-band.
4. Battery per-gate wall-clock timeout mapping onto the FAIL row (rc=timeout) instead of hanging the whole battery forever.
5. audit-all preflight cost warning when reports × living-doc lines crosses a threshold (print an estimate before the replay starts).
6. Perf regression net: a benchmark pinning audit-all runtime vs corpus size, plus a CI arm over a synthesized ~1.6k-report corpus so the cliff stays visible.
7. Wedged-process diagnosis: strace one 0%-CPU docs-gates straggler before asserting the pipe-block class (Go's stdout SIGPIPE behavior may not be the mechanism).
8. AGENTS.md non-obvious fact: the audit-all cost model with the CV numbers (1,643 reports × ~2,000 living lines ≈ hours; go-health's 66 × ~500 ≈ 2 min).
9. go-health archive ritual: 36 unarchived reports vs threshold 15, oldest 27 days (sibling repo) <!-- fresh:skip -->
10. go-health harvest: the 32 unharvested audit-all items into its TODO_LIST; the sqlite restart-durability and architecture-gate.yml obligations each appear in TWO reports (sibling repo) <!-- fresh:skip -->
11. go-health routing: 4 unrouted rows — check-art-dupl.sh exec bit, T09 version-skew script, T22 shutdown decision record, Issue #6 annotation (sibling repo) <!-- fresh:skip -->
12. go-health annotate: 376 findings — 7 F/spacing in living docs (deprecation-policy, etag-rejection-design, two reports), the rest legacy struck-without-verdict in archived reports (sibling repo) <!-- fresh:skip -->
13. go-health cadence detail: 5 half-struck DONE rows need the verdict wrap docs-strike applies (sibling repo) <!-- fresh:skip -->
14. CV archive ritual: 370 unarchived reports vs threshold 15, oldest 34 days — the fleet's biggest debt pile (sibling repo) <!-- fresh:skip -->
15. CV routing: 1 unrouted row in the chromedp-takeover fleet-audit report (sibling repo) <!-- fresh:skip -->
16. CV annotate: 717 findings, F/spacing dominated including scripts shell files — one formatter pass likely clears most (sibling repo) <!-- fresh:skip -->
17. CV: route the obligations carried by the 2 HTML status reports outside the .md scan (sibling repo) <!-- fresh:skip -->
18. CV: one complete audit-all replay (post perf-fix or overnight) for the full-corpus verdict this session never got (sibling repo) <!-- fresh:skip -->
19. go-health and CV: consider `.docs-gates.json` ratchets (audit_max, annotate_max) so their reds show tolerated progress instead of monolithic FAIL (sibling repos) <!-- fresh:skip -->
20. Ops: dispose of the wedged docs-gates processes — owner call, see §g Q1 (owner action) <!-- fresh:skip -->

## g) Questions I cannot answer myself

1. **Kill the wedged processes?** Three 0%-CPU docs-gates stragglers
   (CV 7 h, webphone 2 h, BuildFlow 1.5 h) predate this session and are
   not mine; the fourth (webphone, 2% CPU) looks alive. Kill the three,
   or leave for their owning sessions?
2. **audit-all semantics:** is replaying the FULL corpus (CV: 1,271
   archived included) still the intended contract for repos this size,
   or is a scoped mode (unarchived + last N archived) an acceptable
   config knob? The freshness/audit-all split is documented as
   semantic, so I will not decide gate semantics unilaterally.
3. **Cross-repo authority:** should I harvest the go-health/CV reds
   into THOSE repos' TODO_LISTs and run their archive rituals, or does
   that stay with their own sessions? (Sibling-repo permission is a
   standing owner gate from earlier reports.)

## Appendix (2026-11-05 ~03:45): the wedged-process autopsy, observably closed

The 03:40 efficiency session re-inspected the foreign docs-gates
processes (read-only; the plan's E6):

- The 7 h CV straggler, the second webphone run, and the BuildFlow
  straggler are GONE (their sessions or the host ended them). One
  process remains: PID 248337, `docs-gates run
  /home/lars/projects/webphone`, parent **buildflow**, exe the
  fleet-installed `docs-gates-6ed909f` build, 28 min old, 2 s CPU
  total, 0.1% CPU.
- Per-thread wchan: 13 of 14 threads in `futex_do_wait` (idle Go
  runtime), exactly one in `do_epoll_wait` (the netpoll thread). NO
  thread is in `pipe_write`/`pipe_read`, no spinning, ~99 voluntary
  context switches total. `strace -p` is blocked by yama
  ptrace_scope=1 (buildflow owns it), so the exact syscall stays
  unproven — but the observable shape is decisive: the main goroutine
  is parked in netpoll-backed I/O that never arrives (the
  never-closing-pipe class), at 0% CPU, forever.
- Consequence for the battery: this wedge class never self-resolves,
  which is exactly what the per-gate timeout (E4, gate_timeout_sec
  default 300, FAIL row rc=timeout) now bounds — a wedged gate fails
  its row instead of hanging the run. The remaining PID is left alive
  for its owning buildflow run (kill still owner question 1).

### Appendix addendum (~04:05): mechanism CONFIRMED by live reproduction

The first post-fix fleet battery run (crm) reproduced the wedge under
the new code and closed the question: crm's audit-all (179 findings ×
attribution lines ≈ >64KB captured output) hit the per-gate timeout at
5m0s — and the direct, un-captured run of the same gate completed in
seconds. That is the capture-pipe deadlock: a gate writing past the
pipe buffer parks on the full pipe while io.ReadAll waits for the gate
to return first (the epoll-thread picture in the autopsy above is the
parked write's netpoll wait). Fixed the same hour: both capture
functions drain concurrently (readers only; EOF still gated on the gate
returning). crm audit-all now: 790ms, honest rc=1. Fleet re-runs with
the fixed binary: crm FAIL (freshness + audit-all 179 — real debt,
unpinned), go-appkit PASS with its audit_max=12 ratchet holding exactly
(memoized semantics byte-identical to the pin's calibration — the
parity guarantee holding across the fleet).
