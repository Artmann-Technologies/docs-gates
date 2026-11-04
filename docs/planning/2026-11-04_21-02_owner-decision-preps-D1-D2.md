# Owner decision preps — D1 (battery adoption scope) + D2 (go floor)

**2026-11-04 21:02 CET** · baseline `4777163` · input: TODO_LIST §4 rows +
plan v2 T23/T24. Decision-ready: each section ends in ONE question with a
recommendation, so the answer can be a one-liner.

---

## D1 — battery adoption on hosts: timer/PMA lane vs PATH-only

**The question:** should fleet hosts run the docs-gates battery
PERIODICALLY over their repos (a systemd timer or a pre-commit-adjacent
mechanism riding PMA), or keep the PATH-only scope (binaries available;
only repos that opt in — via BuildFlow detect, CI, or a manual run —
execute it)?

| Axis            | PATH-only (status quo)                                                                                  | Timer/PMA lane                                                                                                |
| --------------- | ------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| What runs today | Binary on PATH (post-T3 deploy); BuildFlow consumers auto-run it per build; evo-x2's own CI-style gates | Nothing extra — a timer would run `docs-gates run <repo>` on a cadence over selected repos                    |
| Failure surface | A repo nobody builds drifts silently                                                                    | Drift is caught within one timer period, fleet-wide                                                           |
| Noise           | Zero (nothing runs uninvited)                                                                           | Every red battery = a notification somebody must triage; repos WITHOUT living docs are permanently red        |
| Cost to operate | None                                                                                                    | Timer units + state + triage rota; reds in non-doctrine repos are UNACTIONABLE (no TODO_LIST to harvest into) |
| Preconditions   | Deploy (T3) only                                                                                        | Deploy + a per-repo allowlist + a doctrine-freshness bar (the battery only makes sense where the rituals run) |
| Rollback        | Trivial (nothing to disable)                                                                            | Mask/disable timer; state cleanup                                                                             |
| Reversibility   | High                                                                                                    | Medium (notification channels and triage expectations accrete)                                                |

**Interaction with the pilot ratchets (T11):** crm + go-appkit run the
battery under adoption ratchets seeded at measured debt. A timer lane
running the battery fleet-wide BEFORE the pilots demonstrate the drain
ritual would judge repos the doctrine has not onboarded — the 1%-tier
ships the binaries, the pilots produce the data, and the lane decision
deserves that data.

**Recommendation:** PATH-only for one quarter (2026-Q1 2027 review).
Concretely: deploy the binaries (T3), keep battery execution inside
BuildFlow consumers + opted-in repos, let the pilot ratchet drains (T11)
accumulate red/green history, then re-decide with the pilots' numbers on
the table. If the pilots drain their ratchets to zero and HOLD for a
month, the timer lane's noise objection weakens and the fleet-wide
default becomes defensible.

**One-liner needed:** "PATH-only for Q1, re-decide with pilot data" — or
override with a target date for the timer lane.

---

## D2 — root go.mod floor: hold 1.26 or raise to 1.27

**The question:** the root module pins `go 1.26` while the flake builds
AND the provider module pin `go 1.27`. Raise the root floor to 1.27 so
one number rules all three surfaces (and AGENTS.md's toolchain paragraph
collapses to one sentence), or hold 1.26?

**Facts on the ground (verified 2026-11-04):**

- Root `go.mod`: `go 1.26`; stdlib-only, zero requires.
- `pkg/provider/go.mod`: `go 1.27` (go-finding/toolsdk needs it).
- flake: one `goPkgAttr = "go_1_27"` binding rules build + devShell +
  provider check; `GOTOOLCHAIN=local` everywhere.
- CI: plain `go` on ubuntu-latest (currently ≥ 1.27 anyway) — the floor
  is the constraint only for EXTERNAL consumers.
- Both 1.26.8 and 1.27.3 toolchains are exercised in this session; the
  suite is green on both (the 1.26 devShell path is what the parity
  guarantee originally rode).

| Axis                  | Hold 1.26 (status quo)                                                                                                       | Raise root floor to 1.27                                                                                                                                    |
| --------------------- | ---------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `go install …@latest` | Works for go ≥ 1.26 users                                                                                                    | Requires go ≥ 1.27 (one minor older toolchain excluded; Go users auto-upgrade fast, but the install line in the README becomes a requirement, not a nicety) |
| AGENTS.md story       | Three-sentence toolchain paragraph explaining the 1.26/1.27 split + why local                                                | One sentence: "the floor is 1.27 everywhere" — the split explanation (the longest AGENTS paragraph) disappears                                              |
| Mismatch surface      | The root floor and every real build disagree TODAY — a go 1.26 build of the root works but nothing in the fleet exercises it | Floor = pin; no build anywhere disagrees with the declared floor                                                                                            |
| Risk                  | A 1.26-only consumer hits a stdlib API newer than their floor (silent drift class)                                           | A 1.26 consumer loses `go install` until they upgrade (loud, one-time)                                                                                      |
| Precedent             | Provider already forced 1.27 as the EFFECTIVE floor for anything importing it                                                | —                                                                                                                                                           |

**Recommendation:** raise the root floor to 1.27 at the next release
(v0.2.0, T1). The provider made 1.27 the effective floor on day one; the
declared floor should not claim less than every real consumer needs. The
CHANGELOG carries a "go: root floor 1.27" line under that release. After
the tag, AGENTS.md's toolchain paragraph collapses to one sentence (F94
— owner-gated, lands with the decision).

**One-liner needed:** "floor to 1.27 with v0.2.0" — or "hold 1.26
because <reason>".

---

_Point-in-time snapshot; annotate, never rewrite. Both recommendations
ride owner questions that have been asked three times today (07:11 §g,
19:01 §g, 19:25 §g); nothing here executes without the answer._
