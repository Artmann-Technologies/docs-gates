# cmdguard docs-gates adoption + land-parity fixes — session status

**Scope line.** This report VERIFIES, with commands run in this session and
their exit codes: the docs-gates battery over both repos (system binary, the
pinned `go run ...@v0.3.0` CI invocation, and the BuildFlow provider step —
all PASS), the full docs-gates root suite (`go build ./...`, `go vet ./...`,
`go test ./...` green incl. parity goldens), golangci-lint 0 issues, erraudit
lint rc=0, nolint-audit 16/16 needed 0 stale, dprint clean on every edited
file, actionlint clean on ci.yml, and the live land-parity DRIFT→OK flip over
cmdguard's real stamp. It does NOT verify: the new cmdguard dogfood job on a
real GitHub runner (never pushed), `nix flake check` after commit `4572726`
(another session's flake fix, arrived mid-session and not reviewed here), the
pkg/provider nested-module suite (not re-run; no JSON-face change was made),
and provider behavior on a host WITHOUT the binary on PATH (untested).

## Session narrative

The task was "adopt /home/lars/projects/cmdguard fully to the max" for
docs-gates. Research showed the adoption was ~half done: config existed, the
battery was green, and BuildFlow's compiled-in provider already rode the
default pipeline and pre-commit mode — but the ROADMAP's "docs-health CI
gate" row was still open, no CI dogfood job existed, and nothing documented
the adoption contract. That surface is now closed (cmdguard `4c47703`),
and closing it surfaced two real docs-gates tool bugs (land-parity case
bug; stamp without a gitignore guard) which are fixed and regression-tested
on docs-gates master. Mid-session, a concurrent BuildFlow go-fix pass and
the auto-commit daemon re-landed parts of my in-flight trains in both repos;
net content was reconciled forward and verified on HEAD — nothing was lost,
but see §d for the honest accounting.

## a) FULLY DONE

- **cmdguard CI dogfood job** — `dogfood` job in `.github/workflows/ci.yml`:
  pinned `go run github.com/Artmann-Technologies/docs-gates/cmd/docs-gates@v0.3.0 run .`,
  `fetch-depth: 0` (annotate resolves done-at hashes against the real commit
  graph), same SHA-pinned actions as the existing jobs. actionlint clean; the
  EXACT CI invocation tested locally: PASS, exit 0.
- **cmdguard AGENTS.md adoption section** — surfaces (CI job, BuildFlow
  provider incl. pre-commit mode, fleet binary from PATH), the deliberately-NO-flake-input
  rationale (docs-gates' flake carries a PRIVATE erraudit git+ssh input; consuming
  its outputs would force a private fetch on CI and public clones), the
  near-bar cadence expectation, and the living-doc edit ritual (pre-flight →
  battery BEFORE commit). Plus a Quick Start battery line. dprint clean.
- **ROADMAP strike** — "docs-health CI gate" struck as LANDED 2026-11-14 with
  evidence; pre-flight SAFE before the edit; land-parity OK after the fix;
  battery green after; the pre-existing dprint table drift in the same file
  re-padded by the formatter (28 lines, padding-only churn, verified).
- **cmdguard `.gitignore` stamp guard** — `.docs-gates-preflight.json` entry
  added (outside the buildflow-managed block, docs-gates convention) and the
  stamp the daemon had committed is untracked (`git rm --cached`, file kept
  on disk; land-parity still verifies it).
- **docs-gates land-parity case bug (root-caused + fixed + tested)** — the
  corpus lens is lowercased (`LivingDocsLines`) while the stamp keeps typed
  case, so `PreFlightParity` false-DRIFTed every mixed-case landing; the
  first parity run over cmdguard refused a byte-identical line (proven with
  `cat -A`). Fix: compare through the same trim+lower lens as the --replace
  matching. Three new regression tests (case lens, trailing whitespace, and
  the existing retyped/no-stamp/malformed set still green). The old fixtures
  were all-lowercase — exactly why the bug survived since 2026-11-11.
- **docs-gates stamp-gitignore warning** — `writePreflightStamp` now probes
  `git check-ignore -q` after a SAFE write and notes (stderr, best-effort)
  when the stamp is not ignored; a TRACKED stamp still reports not-ignored,
  which is the worst case, so the plain probe is correct. Two-arm git-fixture
  test (no .gitignore → note fires; entry present → silent); probe failure
  (no git/not a checkout) skips silently. `//nolint:gosec` counted LIVE by
  nolint-audit (16/16, 0 stale).
- **docs-gates AGENTS.md** — two paid-for Non-obvious-facts entries (stamp
  rode commits; corpus-lens parity), with the discrimination story.
- **End-to-end verification** — batteries green over cmdguard (system binary,
  pinned @v0.3.0, BuildFlow provider) and docs-gates (twice, incl. after the
  AGENTS.md edit); full root suite green; lint/erraudit/audit gates green;
  parity goldens intact (land-parity is not among the six pinned
  invocations).
- **Commits** — cmdguard `4c47703` (gitignore + ROADMAP reformat + stamp
  untrack; the daemon's `7500ff9` carries the earlier partial train: ci.yml,
  AGENTS.md, strike); docs-gates `e5eaa07` (the concurrent go-fix pass landed
  my in-flight preflight changes — diff read, content equals my verified
  intent) + daemon `a1da77e` (AGENTS.md). Nothing pushed (per repo rules).

## b) PARTIALLY DONE

- **cmdguard adoption, CI pin freshness** — the dogfood job pins `@v0.3.0`,
  which predates today's land-parity/stamp fixes. Fine for CI (the battery
  matrix never runs land-parity), but the pin should bump to the next
  docs-gates release. Blocked on that release existing (§c).
- **SU22 archive batch** — deliberately NOT executed (its own ~65-verdict
  session per the TODO row). Consequence, recorded in cmdguard's AGENTS.md:
  cadence sits at 23/23 and the oldest unarchived report (2026-10-04) hits
  the 45-day bar around 2026-11-18 — the CI dogfood will go RED as the
  ritual reminder unless SU22 lands first.
- **This report's §f harvest** — the 50-item list is table-form brainstorm
  (status-report skill: larger N is ROADMAP fuel, not a commitment ledger),
  so the freshness gate reads no open §f items; the genuinely committed
  items among them are already TODO rows (SU22) or owner questions (§g). A
  proper HARVEST pass into TODO_LIST/ROADMAP is owed next session if the
  user promotes any §f rows.
- **Cross-repo documentation symmetry** — docs-gates gained adoption
  knowledge in its AGENTS.md, but a consumer-facing "adopting docs-gates"
  README section (worked example: this session) is not written.

## c) NOT STARTED

- Mirrored status report in cmdguard's own corpus — blocked by cadence
  headroom 0/23 until SU22 frees a slot (adding a report today would trip
  the gate red on arrival).
- docs-gates v0.3.1 (tag + CHANGELOG cut) carrying today's two fixes.
- cmdguard CI pin bump `@v0.3.0` → `@v0.3.1` (depends on the release).
- Automation for the `go run @version` consumer pin (dependabot cannot see
  it; needs a renovate custom manager, a drift-check script, or an accepted
  manual ritual row).
- `nix flake check` re-run after the other session's `4572726` flake fix,
  and `scripts/check-erraudit-rev.sh` (the AGENTS.md rule: re-run the
  tripwire after any flake-touching commit).
- CI `concurrency:` groups and `timeout-minutes` guards on the new job
  (fleet-wide nicety, not present in either repo's workflows).

## d) TOTALLY FUCKED UP (honest near-miss ledger — no data loss, but real)

- **Daemon/concurrent-process train races (twice).** cmdguard: the daemon
  committed my mid-train state (strike pre-reformat, stamp swept into
  history) so the adoption is split across a heuristic commit and mine.
  docs-gates: a concurrent go-fix pass re-landed my in-flight edits under
  ITS commit message while I was editing the same lines. Both reconciled
  forward with full diff reads, but the history is messier than the work:
  lesson — commit trains at every logical completion point instead of
  batching.
- **The tool bug I hit was shipped.** Land-parity (the L1-23 fix, 2026-11-11)
  had a case bug from day one; my first parity run failed on byte-identical
  content. Cost: one ~15-minute debugging detour, which became the fix —
  but it means every mixed-case pre-flight landing since 2026-11-11 got a
  false DRIFT (or, worse, taught operators to ignore the guard).
- **Pipeline exit-code blindness (mine).** `dprint check ... | tail -5 &&
  echo "DPRINT CLEAN"` printed CLEAN while dprint had failed — `&&` bound
  to `tail`, not dprint. Caught by re-reading output; never put a success
  echo after a piped check.
- **The `rg -rn` trap (mine, third strike per the host hook).** Blocked
  twice by the display-replace guard in one session. Use `rg -n`.

## e) WHAT WE SHOULD IMPROVE

- **Commit trains immediately at logical completion points** — both §d races
  trace to batching too much work before the first commit.
- **cmdguard's pre-commit lane is over its budget** (1m42s vs 60s; it runs
  go-auto-upgrade, erraudit, and other repo-wide tools in pre-commit mode).
  Trim via `.buildflow.yml` skip_steps with rationale, or accept and record.
- **docs-gates flake ergonomics for consumers** — the private erraudit input
  makes the flake un-consumable publicly; a packages-only split would let
  adopters add a flake input safely instead of the system-binary +
  `go run @pin` pattern.
- **Consumer-pin automation** — `go run @vX.Y.Z` in CI is invisible to
  dependabot; pick a bump mechanism before the pin silently rots.
- **AGENTS.md growth discipline** — cmdguard's is 438 lines vs the 220
  advisory (D13 accepts it deliberately; trims come from prose only). My
  addition added ~12 load-bearing lines; the next session should look for
  prose to trim, not gotchas to move.
- **Status-report placement** — with cmdguard's cadence headroom at zero,
  sessions working on cmdguard either execute SU22 first or report in the
  other repo; decide the convention explicitly instead of tripping the gate
  by accident.

## f) Up to 50 things we should get done next (brainstorm, sorted by impact)

| #  | Item                                                                                                          | Repo         | Note                                                         |
| -- | ------------------------------------------------------------------------------------------------------------- | ------------ | ------------------------------------------------------------ |
| 1  | Land SU22: the 2026-11-08_02-29 archive batch (~65 table verdicts), return `max_unarchived` 23→20             | cmdguard     | Already TODO SU22; the 45-day cadence breach fires ~11-18    |
| 2  | Cut docs-gates v0.3.1 (CHANGELOG + annotated tag + push), carrying the land-parity + stamp fixes              | docs-gates   | Unblocks the consumer pin bump                               |
| 3  | Bump cmdguard's CI pin `@v0.3.0` → `@v0.3.1` after the release                                                | cmdguard     | One-line ci.yml change + battery                             |
| 4  | Re-run `nix flake check` + `scripts/check-erraudit-rev.sh` after `4572726`                                    | docs-gates   | The AGENTS.md rule after any flake-touching commit           |
| 5  | Run `pkg/provider` suite (`cd pkg/provider && go test ./...`) as its own verification step                    | docs-gates   | Not re-run this session                                      |
| 6  | Watch the dogfood job's first real GitHub run (push cmdguard, check module-download latency + green)          | cmdguard     | Never executed on a real runner                              |
| 7  | Trim cmdguard's pre-commit BuildFlow lane to its budget (1m42s vs 60s)                                        | cmdguard     | skip_steps with rationale, or accept-and-record              |
| 8  | Add the `uses:` reusable-job shape to the cue-policy schema (kills 2 pre-existing cue-vet error findings)     | cue-policy   | Owned upstream; cmdguard's hook is advisory meanwhile        |
| 9  | Decide the `go run @version` bump mechanism (renovate custom manager / drift script / manual ritual row)      | both         | Dependabot cannot see run-command pins                       |
| 10 | Mirror this adoption report into cmdguard's corpus after SU22 frees a slot                                    | cmdguard     | Cadence headroom 0/23 today                                  |
| 11 | Write a consumer-facing "adopting docs-gates" README section (this session as the worked example)             | docs-gates   | Config + CI job + pre-commit leg + stamp gitignore           |
| 12 | Split a private-input-free packages-only flake output for public consumers                                    | docs-gates   | Defuses the no-flake-input limitation documented in cmdguard |
| 13 | Add `concurrency:` groups + `timeout-minutes` to workflows (both repos)                                       | both         | Cancel superseded runs; bound the new job                    |
| 14 | Reinstall buildflow (binary predates HEAD by ~9h; doctor freshness warn)                                      | host         | `nix build . && nix run .#reinstall` per the warn text       |
| 15 | Run `buildflow format` once so cached verdicts re-verify under the drifted formatter identities               | cmdguard     | Preflight warn; ROADMAP reformat was one instance            |
| 16 | Point dprint.json at local .wasm plugins (nix store) or accept the cold-fetch hang risk explicitly            | cmdguard     | Preflight warn with exact fix text                           |
| 17 | Add lychee excludes for the four point-in-time archived doc dirs                                              | cmdguard     | Preflight warn names them                                    |
| 18 | Add pytest/vulture to a devShell or blocklist the python steps (no python in this repo)                       | cmdguard     | "not in project devShell" warns every pre-commit             |
| 19 | Silence/absorb module-inventory go.work infos (13 findings from the LOCAL-ONLY go.work)                       | cmdguard     | Baseline accept or step config                               |
| 20 | Rebuild the fleet system docs-gates package so PATH picks up today's fixes                                    | host         | System binary is at eb2fa6b (pre-fix)                        |
| 21 | Consider `--clear-stamp` flag instead of hand-deleting `.docs-gates-preflight.json`                           | docs-gates   | The DRIFT message currently says "strike the stamp file"     |
| 22 | Land-parity stamp: record per-batch instead of last-write-wins (split sessions overwrite)                     | docs-gates   | Design pass; documented behavior, lossy for multi-batch      |
| 23 | Add a parity golden for `--land-parity` output (currently unpinned message shape)                             | docs-gates   | A DRIFT-message wording change would otherwise sail          |
| 24 | Land-parity stamp: record the living-docs git SHA for stronger post-hoc-edit detection                        | docs-gates   | Design idea; guard-not-gate philosophy applies               |
| 25 | Surface land-parity as an opt-in battery gate row (skip_gates vocabulary)                                     | docs-gates   | Needs a design pass; guard vs gate                           |
| 26 | Provider HealthCheck hint: suggest `@vX.Y.Z` instead of `@latest`                                             | docs-gates   | Version-drift-prone install advice                           |
| 27 | Document the `go run ...@vX.Y.Z` consumer pattern in docs-gates README                                        | docs-gates   | Install-hint parity for non-fleet hosts                      |
| 28 | Upload the battery `--json` verdict document as a CI artifact on red                                          | cmdguard     | Debuggability for the dogfood job                            |
| 29 | Write per-gate timings to `$GITHUB_STEP_SUMMARY` in the dogfood job                                           | cmdguard     | Cheap observability win                                      |
| 30 | Add a skip-if-no-git guard to the stamp-warning test for git-less sandboxes                                   | docs-gates   | CI has git; robustness only                                  |
| 31 | Fix pingen-spike's docs-gates red (TODO SU9)                                                                  | pingen-spike | Third repo where adoption is pending                         |
| 32 | Evaluate the docs-gates BuildFlow provider on a host WITHOUT the binary (install-hint path)                   | docs-gates   | Untested failure mode                                        |
| 33 | Verify TODO_LIST §77 preflight-bind-check is struck (referenced by test comments)                             | docs-gates   | Verify-only                                                  |
| 34 | Verify `nix run .#check-all` still passes across all 8 modules post-v5.1.0                                    | cmdguard     | Verify-only                                                  |
| 35 | Refine ROADMAP "Stale-doc-ref CI guard" row — partially covered by the battery now                            | cmdguard     | Possibly strike or narrow                                    |
| 36 | Add "run the battery after formatter sweeps" to the post-rewrite ritual bullet                                | cmdguard     | The dprint table reformat was an instance                    |
| 37 | AGENTS.md prose trim per D13 (438 lines; prose only, never gotchas)                                           | cmdguard     | Advisory warning accepted but growing                        |
| 38 | After SU22: re-tighten `.docs-gates.json` to 20 and re-prove the battery                                      | cmdguard     | The pin comment promises the return path                     |
| 39 | Decide whether cmdguard-lint's provider should match by default in Go repos (explicit `-s` today)             | cmdguard     | Provider-matching policy question                            |
| 40 | Add the pinned `go run` fallback line next to the fleet-binary line in cmdguard Quick Start                   | cmdguard     | Non-fleet-host contributors                                  |
| 41 | Dogfood CI: pin the runner Go version via `go-version-file` on docs-gates' go.mod instead of "1.27" literal?  | cmdguard     | Consistency question with the test job's matrix              |
| 42 | Watch the `docs-health` token DF crossing (7→8) for future cmdguard doc edits                                 | cmdguard     | Pre-flight flagged it; monitor, no code                      |
| 43 | Reconcile `.docs-gates-preflight.json` semantics docs (AGENTS vs report vs message text) in one place         | docs-gates   | Doc-consolidation nit                                        |
| 44 | Consider stamping the battery's estimate line into CI logs (audit-all replay shape)                           | docs-gates   | Observability nit                                            |
| 45 | Fleet convention: should every adopted repo add the stamp gitignore entry via BuildFlow auto-configure?       | BuildFlow    | The warning could become a managed-block entry               |
| 46 | Add cmdguard's dogfood job to submodule-smoke/website workflows? (probably not — one battery per repo)        | cmdguard     | Decide + record if declined                                  |
| 47 | Sweep other fleet repos for committed `.docs-gates-preflight.json` files (the daemon class may have hit more) | fleet        | `git log --all --full-history -- '*.docs-gates-preflight*'`  |
| 48 | Document in cmdguard AGENTS.md that `docs-gates run .` needs the fleet binary (or the go run fallback)        | cmdguard     | One-line clarification                                       |
| 49 | Re-verify erraudit's `nolint-audit` count claim in any doc that cites it (16 today; never cite from memory)   | docs-gates   | Standing rule                                                |
| 50 | Consider a CHANGELOG entry convention for repo-infra changes (CI job) vs library releases (cmdguard has none) | cmdguard     | This session's CI change has no CHANGELOG home               |

## g) Questions I can NOT figure out myself

1. **Report placement convention:** with cmdguard's cadence headroom at 0/23,
   do you want cmdguard sessions to execute SU22 (the archive batch) BEFORE
   writing their status reports, or should cmdguard-side sessions keep
   reporting in the tooling repo (docs-gates) until SU22 lands? I chose
   docs-gates for this report; a mirrored cmdguard report trips the gate red
   on arrival today.
2. **Release cadence:** should I cut docs-gates v0.3.1 now so cmdguard's CI
   pin moves to a bug-fixed release, or do you prefer accumulating fixes and
   releasing less often (leaving the pin at v0.3.0, which is sufficient for
   CI since the battery never runs land-parity)?
3. **Adoption path going forward:** is the blessed consumer surface
   system-binary + pinned `go run @version` (what I wired), or should
   docs-gates grow a private-input-free packages-only flake so adopters can
   add a real flake input safely? This decides item #12 and how the next
   five repos adopt.
