# Status Report — Buildflow Lockfile-Drift Recovery: erraudit vendorHash Fix, 26-Finding Triage, Gated Docs Trains

**Report:** 2026-11-06 (session ~05:00–05:45 CET) · **Repo:** docs-gates @ `4c73c78` (main) + erraudit @ `5c841cf2` (local, unpushed) · Origin: docs-gates pushed to `ea7f380` minus the docs train; erraudit pushed to `c1902fea` (its fix is local-only)

**Scope line — what this session VERIFIED:** every load-bearing claim below
rode a live command run inside this session. The nix breakage was
reproduced at the offending rev (`nix build` of erraudit `c1902fe` → the
exact fixed-output hash mismatch, `skC3…` vs `xoyn…`), the lockfile
drift was read from the daemon commit's own diff (`git show 46f9b96 --
flake.lock`), and every green claim at session close is a fresh rc=0:
`bash scripts/check-erraudit-rev.sh` (PASS), `nix flake check
--all-systems --no-build` (all checks passed), `nix develop -c true`
(devShell builds), root `go test ./...` and `pkg/provider go test
./...`, `golangci-lint run ./...` on BOTH modules (0 issues each),
`erraudit lint ./... --type-aware` (rc=0) and `erraudit nolint-audit .`
(14 directives: 14 needed, 0 stale — matching the AGENTS.md count), the
dogfood battery 5×PASS before the docs commit, `docs-freshness
--pre-flight` (SAFE) on the new TODO row, `buildflow -s go-auto-upgrade
--format finding` on BOTH scopes (all 26 findings enumerated and
classified, not summarized from the failing run's truncated paste), and
the consumer-proof `nix build git+file:///home/lars/projects/erraudit?rev=5c841cf2…`
(builds the full derivation chain a flake update would pull).
**NOT verified:** a full default-mode `buildflow` rerun over docs-gates
(still exits 1 by the two documented roundtrip bugs; nothing in this
session touched code those steps compile, and the pasted run already
demonstrated the failure class), erraudit's own full test/buildflow
lanes (pre-existing reds — file-size 49, cue policy, toolchain skew —
partially polluted by this session's own machine load, see §d1; not
this session's repair scope), CI (never runs locally; next push
triggers it), and the erraudit push itself (owner action, §b1).

---

## a) FULLY DONE

1. **The nix breakage root-caused and reproduced.** The pasted buildflow
   failure's `nix-flake-update` repair step had bumped the erraudit
   input to master `c1902fe` (revCount 1887), whose own flake carries a
   stale go-modules vendorHash — proven by building erraudit at that
   exact rev and getting the same hash mismatch the roundtrip showed.
   Upstream cause: erraudit's auto-daemon committed a go.mod/go.sum
   sweep (`c1902fea`, 03:38) without re-pinning its flake hash.
2. **docs-gates' flake.lock restored to the pinned working rev**
   (`3d1f7ed`) — surgical 8-line restore, committed as `ea7f380` BEFORE
   the daemon could wrap it. Tripwire PASS; `nix flake check
   --all-systems --no-build` all-green; `nix develop -c true` builds
   the devShell again.
3. **erraudit's vendorHash fixed in the owning repo** via the
   fleet-blessed `buildflow -s nix-hash-fix --fix` (never hand-pasted):
   `skC3…` → `xoyn…`, one line in erraudit's flake.nix, committed as
   erraudit `5c841cf2` with a full-rationale message. `nix build
   .#default` green there; consumer-proof green from here.
4. **All 26 go-auto-upgrade findings triaged individually** (not from
   the paste — a fresh `--format finding` enumeration of both scopes):
   they reduce to exactly three classes — samber/lo (12), jsonv1tov2
   (4 root + 1 provider boundary), cross-file-wrapper (6) — each a
   deliberate non-fix with a recorded rationale (§e items below).
5. **AGENTS.md now documents the incident class** (buildflow repair
   steps rewrite lockfiles mid-run; the daemon wraps the damage; the
   tripwire is the net; recovery = forward `git restore --source` +
   your own commit) **and the three warning classes** as deliberate
   non-fixes (`4c73c78` + daemon `3abe9f1`).
6. **TODO_LIST gains the cross-repo leg** (§3 row): push erraudit
   `5c841cf2`, then `nix flake update erraudit` + re-pin
   `check-erraudit-rev.sh` expected_rev deliberately. Pre-flown SAFE
   before landing.
7. **The docs rituals held:** pre-flight before the living-doc edit,
   dogfood battery 5×PASS before the commit — the `4cf0e50` lesson
   applied, not just cited.
8. **Full close-out verification green:** both go suites, golangci 0
   findings on both modules, erraudit lint rc=0, nolint-audit 14/14
   fresh, flake eval both-arch green.

## b) PARTIALLY DONE

1. **The erraudit fix is landed but stranded locally.** Pushing is an
   owner action (no standing push authority in erraudit's AGENTS.md,
   unlike docs-gates' "push your own trains"). Until it lands on
   origin, every consumer flake that updates erraudit still pulls the
   broken `c1902fe`; docs-gates stays deliberately pinned at `3d1f7ed`
   and cannot take the fix.
2. **The deliberate bump-and-re-pin loop is prepared, not executed.**
   The TODO row spells it out (run erraudit checks on the new rev
   first), but `nix flake update erraudit` here can only happen after
   the push.
3. **The incident lesson is recorded in the consumer repo only.**
   docs-gates' AGENTS.md carries the full recovery ritual; erraudit's
   AGENTS.md — the repo whose daemon caused the class — does not yet
   know its vendorHash went stale behind its own daemon's sweep, nor
   that its "fast vendorHash drift gate" (flake.nix:153, an eval-time
   `test -d`) structurally cannot catch a fixed-output hash mismatch.
4. **The BuildFlow toolchain skew is diagnosed, not fixed.** The
   binary this session used embeds go1.27.0; erraudit's floor moved to
   1.27.1 (a dependency's directive, riding the daemon sweep), so
   samber-linter and test-compile lanes fail there with "rebuild or
   refresh the BuildFlow binary". docs-gates is unaffected (floors
   1.26/1.27), but the fix (`buildflow upgrade` / rebuild) was out of
   scope and not run.

## c) NOT STARTED

1. Push erraudit `5c841cf2` (owner).
2. `nix flake update erraudit` in docs-gates + run erraudit-lint +
   nolint-audit on the new rev + bump `expected_rev` (blocked on 1).
3. erraudit AGENTS.md incident entry (its own repo, its own session or
   an authorized cross-repo train).
4. BuildFlow binary refresh (`buildflow upgrade`) to clear the 1.27.1
   skew lanes.
5. erraudit's pre-existing reds noticed in its pre-commit hook run:
   cue-vet fleet-policy violations (ci.yml job-name conflict,
   release.yml missing concurrency), 49 file-size findings,
   `flake-pin-drift` (project-discovery-sdk pinned v0.22.0 vs go.mod
   v0.23.1), `go-mod-ignore-check` (dist dir unignored).
6. docs-gates pre-commit arming (`git config --unset core.hooksPath`)
   so the docs-battery commit leg auto-installs instead of being run
   manually (done manually this session, again).
7. HARVEST of this report's §f into TODO_LIST/ROADMAP (the standing
   next ritual; explicitly deferred — user said wait).
8. A buildflow-side or repo-side guard that re-runs the rev tripwire
   automatically after any run whose repair phase rewrote flake.lock
   (this session proved the manual discipline is the only net today).

## d) TOTALLY FUCKED UP

Nothing landed broken this session — every working-tree change was
verified green before its commit, and the two repos are both buildable
at HEAD. The honest entries are near-misses and self-inflicted noise:

1. **This session starved erraudit's pre-commit hook and nearly
   mis-reported the result.** Five heavy verification lanes ran in
   parallel while erraudit's ~2-minute buildflow hook executed; its own
   summary shows "3 lost to starvation (15m load: 144.63)" plus
   test-compile kills. Those failures are partly THIS session's load,
   not purely pre-existing debt. The pre-existing findings (file-size,
   cue, skew) are real, but the honest verdict for erraudit's hook is
   "red, cause mixed" — re-run it on a quiet machine before trusting
   any per-step claim from that output.
2. **The daemon wrapped the first AGENTS.md edit mid-train**
   (`3abe9f1`, +24 lines) between the restore commit and the docs
   commit — expected under the ~2–3 min cadence, and recovery was
   content-verification of HEAD rather than commit-stat arithmetic,
   which is exactly the discipline AGENTS.md prescribes. Cost: one
   confusing minute, zero damage.
3. **Two wrong first impulses were caught before acting:** hand-pasting
   the vendorHash into erraudit's flake (the buildflow skill's named
   anti-pattern — `nix-hash-fix` owns the job), and treating the
   run-log's "rewrote: go.sum" lines as real dependency changes (the
   empty `git diff` range disproved it before anything was claimed or
   committed). Both would have been quiet wrongs; neither landed.

## e) WHAT WE SHOULD IMPROVE

1. **Serialize or stagger cross-repo verification lanes.** Load 144 on
   a 15-minute window pollutes every timing-sensitive verdict in every
   repo running concurrently (see §d1). The verification battery should
   run heavy lanes sequentially, or background jobs should be
   throttled, when a foreign repo's gated commit is in flight.
2. **Make the tripwire reflex automatic post-buildflow.** The rev
   tripwire caught this drift — but only because the session ran it.
   Candidates: a buildflow post-run hook, a flake check that pins the
   lockfile's erraudit rev against the tripwire's expected_rev (the
   check already exists — the gap is that nothing FORCES it after a
   repair-phase rewrite), or an erraudit-rev assertion inside the
   pre-commit battery leg.
3. **Record lessons in the owning repo, not just the consumer.** The
   vendorHash-staleness lesson currently lives only in docs-gates'
   AGENTS.md; erraudit is the repo whose daemon creates the class and
   whose drift gate gives false confidence. Cross-repo lessons need a
   home in BOTH directions (consumer: recovery ritual; owner: cause +
   prevention).
4. **erraudit's vendor-drift gate should be honest or gone.** An
   eval-time `test -d "${...goModules}"` cannot detect a stale
   fixed-output hash; it passing green while the package cannot build
   is worse than no check. Either replace it with a cheap real build
   of the go-modules drv or delete it and rely on CI/nix-hash-fix.
5. **Verify by content, not by commit-stat arithmetic**, whenever the
   daemon interleaves (paid again in §d2 — the discipline held, but it
   should be the DEFAULT mental model, not a recovery move).
6. **Keep battery-before-commit absolute for docs trains.** It ran
   three times this session (pre-flight, battery, gates) and each one
   carried real information. The one time this repo skipped that
   ordering (`4cf0e50`) is still the canonical horror story.
7. **Read `--format finding` outputs fully before triaging summaries.**
   The pasted "+19 more…" hid a whole third finding class
   (cross-file-wrapper); the fresh enumeration turned a two-class
   story into the correct three-class one and caught that six of the
   "warnings" re-litigate an already-ACCEPTED design decision.

## f) Up to 50 things we should get done next

_(A brainstorm, not a commitment list — HARVEST must route rigorously.
Items 1–8 are this session's direct output trail; 9–20 are adjacent
observables from this session's runs; 21+ are carried from the board
and AGENTS.md as read this session, marked [board]. Sorted by
impact/effort within each band.)_

**This session's direct trail (highest impact, mostly S effort):**

1. Push erraudit `5c841cf2` to origin (owner action; unblocks 2–4).
2. Update docs-gates' erraudit input to the fixed rev, run
   erraudit-lint + nolint-audit on it, bump `expected_rev` in
   scripts/check-erraudit-rev.sh, commit the deliberate bump train.
3. Verify the dogfood battery + flake check after (2) — the eval must
   build erraudit `5c841cf` from the lock.
4. Add the vendorHash-staleness incident to erraudit's AGENTS.md
   (cause: daemon sweep without re-pin; detection gap: eval-time
   drift gate; cure: nix-hash-fix) — in the owning repo.
5. Rebuild/refresh the BuildFlow binary (`buildflow upgrade` or
   `nix build` in the BuildFlow checkout) so embedded analyzers run
   go1.27.1; re-run erraudit's samber-linter/test-compile lanes to
   confirm the skew class clears.
6. Re-run erraudit's full buildflow on a QUIET machine and re-derive
   its honest per-step verdict (strip this session's load pollution
   from §d1).
7. Decide the fate of erraudit's eval-time vendor-drift gate: replace
   with a real go-modules build probe or delete (false confidence is
   worse than absence).
8. Investigate a buildflow-side post-run tripwire hook (or a repo
   pre-commit leg) that fires `check-erraudit-rev.sh` whenever the
   repair phase rewrote flake.lock — mechanical net for §e2.

**Adjacent observables from this session's runs (erraudit-side debts,
each S–M effort, need erraudit-repo authority):**
9. erraudit flake-pin-drift: project-discovery-sdk input pins v0.22.0,
go.mod requires v0.23.1 — align the flake pin.
10. erraudit cue-vet: ci.yml has conflicting job names ("Self Analysis"
vs "Self-Analysis Threshold" under the fleet #CIWorkflow schema).
11. erraudit cue-vet: release.yml lacks the required `concurrency:`
field (fleet #ReleaseWorkflow schema).
12. erraudit go-mod-ignore-check: `dist/` exists but is not in go.mod's
ignore list.
13. erraudit file-size-check: 49 oversized files (types.go at 755 lines
vs max 350 is the worst) — split or raise the pin deliberately.
14. <!-- fresh:skip --> (sibling-scoped: the ritual belongs to erraudit's
own sessions; this board's §3 tracks only the cross-repo leg) erraudit
docs-gates findings: its own battery is red
(freshness/audit-all/annotate/cadence) — its docs ritual is due.
15. erraudit doc-files-age-check: AGENTS/FEATURES/TODO_LIST all stale
past the 7-day threshold.
16. erraudit go-version-auto-configure + go-structure-linter: re-check
the go.mod floor story after the daemon sweep (the 1.27→1.27.1
move came from a DEP's directive — identify which dep and decide
pin-down vs floor-up).
17. erraudit: confirm its flake's go_1_27 package satisfies the 1.27.1
floor (nix build green suggests yes — pin the proof in its
AGENTS.md like docs-gates does).

**docs-gates own surface (S effort each, from this session's contact
points):**
18. Arm the pre-commit hooks properly (`git config --unset
    core.hooksPath`) so nolint-audit + docs-battery legs auto-install.
19. HARVEST this report's §f + the §e items into TODO_LIST/ROADMAP with
routing rigor (the standing next ritual; user deferred it).
20. Consider running `check-erraudit-rev.sh` as part of the docs-battery
pre-commit leg (cheap, closes the manual-discipline gap locally
even without a buildflow-side hook).
21. Add a regression test asserting the tripwire FAILS on a drifted
lock (the tripwire has a red control documented for nolint-audit
but the rev script's own red arm is untested this session — it DID
fire live, so evidence exists; pin it as a test if the fleet wants
mechanical proof).
22. package provider: decide whether `failedGates`'s hand-rolled filter
gets a comment naming the no-lo rationale (currently only AGENTS.md
carries it; a future contributor reading provider.go:95 sees
nothing).
23. Root module: same rationale-comment question for the three
Map/Filter/SliceToMap loops go-auto-upgrade names (gatecli.go,
docsrouting.go, docsfreshness/check.go) — AGENTS.md documents the
class; the sites themselves stay silent.
24. Re-verify the parity goldens are untouched by this session (they
were — no Go code changed — but the claim deserves its one-line
`PARITY_UPDATE` sanity note in the next session log if any gate
output shape work starts).

**Carried from the board (read this session, unchanged, marked so they
don't get lost in the §f flood):**
25. [board] Float SystemNix to the hardened flake (P1, blocked on the
v0.2.0 tag + push).
26. [board] Deploy evo-x2 for live-PATH docs-gates + wrappers (P1,
owner window; also fixes the un-fixed-wedge binary staleness).
27. [board] Standalone `nix flake check --no-build --all-systems` on
SystemNix post-integration (P2, PARTIAL).
28. [board] Pilot ratchet drain in crm + go-appkit (P3; crm regression
168→179 needs the cross-repo authority ruling).
29. [board] File/fix the BuildFlow roundtrip walker bug upstream
(diagnosis sharpened 2026-11-05; fix = lane-level --tree-root
forwarding; coordinate with active BuildFlow sessions).
30. [board] Fix BuildFlow#87 upstream (constant result-cache key), then
drop the BUILDFLOW_NO_RESULT_CACHE bypass from AGENTS.md.
31. [board] Submit lib.maintainers.LarsArtmann to nixpkgs, then
reference it in flake meta.maintainers (blocked on the nixpkgs PR).
32. [board] SystemNix switch for the per-process golangci
analysis-cache wrapper (59bb7cc35, inert until activation).
33. [board] SystemNix push decision (4 ahead, mixed-owner trains —
owner ruling).
34. [board] pma queue-storm + FastFlowLM monitoring docs (BLOCKED on
the concurrent session).
35. [board] ANNOTATE sweep over older unarchived reports (P3, scoped
down; cadence pressure only).
36. [board] F94: AGENTS.md compression pass (the non-obvious-facts
section accreted again this session: +2 bullets).

**BuildFlow-side product ideas surfaced by this session (upstream
value; each needs the verify-before-filing gate):**
37. Repair-phase lockfile rewrites should trigger a configurable
post-repair verification hook (today: silent rewrite + daemon
wrap; the consumer's tripwire is the only net).
38. go-auto-upgrade: honor per-repo finding suppressions with
rationale (the three documented classes would collapse to zero
noise without skipping the step wholesale).
39. cross-file-wrapper finder: respect `//art-dupl:accept`-style
acceptance markers so ACCEPTED settings-projection accessors stop
re-surfacing as "no call sites in this file" warnings.
40. nix-hash-fix: optionally run the consumer-side tripwire scripts
(`check-*-rev.sh` convention) after rewriting a lockfile-adjacent
hash, so owner repos learn about rev-adjacent repairs.
41. Starvation accounting: when a lane reports "lost to starvation",
tag the run summary so downstream readers don't misread polluted
steps as real failures (this session's §d1 exactly).

**Hygiene / smaller polish:**
42. docs-gates CHANGELOG: the lockfile-restore + docs train deserve a
line when the next CHANGELOG pass happens (no code change; ritual
timing only).
43. Re-check `git status --ignored` hygiene per the AGENTS.md
.gitignore gotcha if any new testdata/corpus files ever land (none
did this session — reminder only).
44. Keep `pkg/provider/go.sum` churn watched: two repair steps claim
rewrites each build; if content-neutral rewrites keep appearing in
run logs, ask BuildFlow whether "rewrote" should mean
content-changed (logging honesty).
45. erraudit: consider its own equivalent of docs-gates' rev tripwire
pattern for ITS critical inputs (the fleet pattern proved its
worth here; erraudit tracks inputs too).
46. docs-gates AGENTS.md: the erraudit-rev bullet could name the
recovery commit SHA class (`git restore --source` + own commit) —
it already does; verify wording survives the next compression pass
(F94) intact.
47. Sweep this session's temp artifacts (`/tmp/todo-row-candidate.txt`)
— trivial, but the candidate-line file is the only session residue
outside git.
48. When the erraudit bump lands, update the consumer-proof command in
AGENTS.md's triage vocabulary if the fleet adopts
`git+file://...?rev=` proofs as the standard consumer check.
49. Re-run the five-gate battery right before ANY push train (CI's
dogfood job will run it anyway; local pre-run avoids a red CI
surprise — this session's final battery green is the baseline).
50. Decide whether docs-gates should ever auto-update erraudit
(track-latest policy) via a scheduled local job that runs the
checks + tripwire bump mechanically — today it needs a human
session; the policy says deliberate, the cost is staleness risk
exactly like this incident (owner-policy question, see §g2).

## g) Three questions I can NOT figure out myself

1. **Push authority for erraudit `5c841cf2`:** May I push the
   vendorHash fix to erraudit master now (its AGENTS.md has no
   "push your trains" mandate, unlike docs-gates), and should the
   docs-gates bump-and-re-pin loop (§f items 2–3) ride in the same
   window? Until then this repo stays pinned one rev behind master on
   purpose, and every consumer flake update keeps hitting the broken
   `c1902fe`.
2. **Is the root module's "encoding/json v1 on purpose" a forever
   decision?** The floor is go 1.26 and jsonv2 imports need ≥1.27;
   the JSON face is byte-pinned by parity goldens and the provider's
   strict decode IS the shape contract. If the fleet's jsonv2 direction
   (erraudit already runs GOEXPERIMENT=jsonv2) is eventually meant for
   docs-gates too, the migration needs a planned golden re-pin — I
   cannot tell whether v1-here is "until someone asks" or doctrine.
3. **Cross-repo authority for erraudit's debt list** (§f items 9–17:
   cue policy violations, 49 oversized files, pin drift, docs ritual):
   fix-forward inside erraudit's own sessions, or do you want a
   sanctioned cross-repo cleanup train from this side (the same ruling
   class as the crm harvest authority question already on the board)?
