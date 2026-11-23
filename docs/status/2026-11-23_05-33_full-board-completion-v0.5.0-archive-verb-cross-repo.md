# Status Report — Full board completion: v0.5.0 shipped, archive verb, cross-repo legs, distill (docs-gates)

> **Date:** 2026-11-23 05:33 CET · **Session:** the 02-39 report's §7
> board executed end to end under the finish-the-whole-board
> directive; the three open owner questions decided on evidence (§g).

## a) What was accomplished

- **v0.5.0 released** (both tags at `34c7c5f`): vocabulary + agents
  gates, annotate_corpus knob, the eight-issue fix wave. Proxy indexed,
  clean-dir `go get` green for root AND pkg/provider; install hint and
  README/FEATURES/CHANGELOG synced. No GitHub Release (house pattern:
  zero releases across eight tags).
- **Enforce flips for this repo**: vocabulary_enforce + agents_enforce
  true, proven exit-0 first (mtime-preserving copy test); battery 10/10
  with teeth. Fleet-wide flip keeps the one-week doctrine (~11-30).
- **The archive verb** (fourteenth): banner + git mv for zero-open-§f
  reports, dry-run default, double-banner FATAL, six tests, dispatch +
  usage pins wired.
- **Pre-commit rev-tripwire leg**: staging flake.lock runs
  check-erraudit-rev.sh — negative-proven (rc=1 on a drifted rev) and
  it CAUGHT the 6th damage event live at commit time.
- **BuildFlow#115 filed and verified**: the lockfile damage's root
  cause is the `nix-flake-update` blindRepairer over bare
  `nix flake update` — live-reproduced (plain runs and bare
  `nix flake check` leave the lock byte-identical; `--fix` rewrites
  erraudit's subtree), source-cited, voice-check 0/0.
- **go-taskqueue**: annotate gate flipped ON via
  `annotate_corpus: "unarchived"` + `annotate_max: 144` (measured on a
  v0.5.0 binary rebuilt from the exact tag into ~/.local/bin); their
  gate script moved 7→8 PASS; doctrine + TODO row updated (commit
  `ea496f69`, local — their release pushes are owner-gated).
- **pbx-artmann**: docs-gates pin v0.3.1→v0.5.0; the sharper v0.5.0
  extraction surfaced 95 unharvested → `audit_max: 95` ratchet; battery
  10/10 (commit `8f7ac739`).
- **BuildFlow**: stale binary rebuilt at HEAD after fixing its own
  stale vendorHash (nix-hash-fix); doctor binary-freshness ✓.
- **go-verdict OUT_OF_SCOPE**: NO-OP by evidence — the prose already
  lives inside the legend; gate green, enforcement changes nothing.
- **erraudit incident row**: done-by-reference (their AGENTS.md:176
  landed it 11-14); docs-gates §6 row struck.
- **Three TODO strikes** with battery-green commits (§9 release row,
  §5 provider-mapping row, §6 erraudit row).
- **Final-loop fixes the fresh buildflow binary unmasked**:
  docsvocabulary split (364→295 + legend.go, file-size finding cured);
  root go floor 1.26→1.27 (the documented D2 recommendation; go.mod,
  both modules now 1.27); AGENTS.md distill 286→223 with the archive
  row, md5 verdict, and fourteen-verb refresh added (see §b).
- LSP ghost source cleared (`~/.cache/golangci-lint-analysis` purged,
  both LSPs restarted — diagnostics went live).

## b) What went wrong and how it was fixed

- **Three more lockfile damage events (4th, 5th, 6th)** — the 6th was
  my own `buildflow --fix` repro, daemon-committed before cleanup. Each
  restored forward per the ritual, tripwire green, own commits. Root
  cause now PROVEN and filed (#115).
- **The daemon raced staging five times** — my edits landed as daemon
  commits ("nothing to commit" while the work sat in `git log`).
  Procedure: check git log, push. Recorded in AGENTS.
- **False moved-tag alarm** — annotated-tag object sha vs commit sha
  misread as a hijacked v0.5.0; `^{commit}` is the honest comparison.
  Recorded.
- **Provider go get false failure** — a pre-propagation first attempt
  poisoned the module cache; fresh-GOMODCACHE retest proved the release.
  Recorded.
- **Edit-tool/fmt races** — the mod-time guard bit four times against
  formatter rewrites; read-verify-edit back-to-back is the discipline.
- **AGENTS.md distill is 3 lines short of done**: 223 vs the 220
  ceiling (286 at peak). One more small trim closes it.

## c) Current state

- docs-gates: battery 10/10 PASS under enforcement; parity goldens
  byte-stable through every refactor; CI green on all pushes (latest
  runs success); the CHANGELOG [Unreleased] carries the archive verb;
  TODO_LIST strikes landed with evidence.
- The findings gate shows ONE standing known-delta: library-policy's
  crypto/md5 ERROR in docsstub — deliberate census-fingerprint design
  (documented in AGENTS; suppression belongs upstream in
  library-policy, whose policy DSL is compiled-in).
- Unpushed: three daemon commits + the AGENTS/go.mod/distill train
  (this report rides the same push).
- go-taskqueue/pbx/BuildFlow: committed locally (or daemon-swept);
  their pushes ride their owner-gated trains.

## d) Lessons (all encoded in AGENTS.md)

- Blind repair steps are damage: a repairer without a finding to answer
  for will rewrite state nobody asked about — the tripwire + commit-leg
  pattern catches the class, the fix is finding-driven updates.
- The stale-binary class now also MASKS findings: three errors appeared
  only after the rebuild. Doctor first, then trust a lint verdict.
- Annotated tags, nested-module go get caching, daemon staging races —
  three small traps, each a two-line AGENTS row away from never being
  paid twice.

## e) Metrics (this session)

- Commits pushed to origin/main: 14 (5 mine-authored, 9 daemon-wrapped
  carrying my work); lockfile restores: 3; releases cut: 1 (2 tags);
  upstream issues filed: 1 (#115, verified); TODO rows struck: 3;
  battery runs green: 9+; lockfile damage events survived: 6 total
  (11-06..11-23); AGENTS.md: 286→223 lines, zero concepts dropped
  (every removal verified against its replacement).

## f) Next up (owner-filtered)

- The last 3 lines of the AGENTS distill (220 ceiling).
- Fleet `agents_enforce` flip ~2026-11-30 (stamps doctrine) + the
  67-repo adoption sweep (§9 rows carry the blast-radius data).
- BuildFlow#115 fix → the tripwire becomes depth, not the only net.
- SystemNix owner window: float to v0.5.0 + evo-x2 deploy + wrapper
  switch (one window, per §3 doctrine).
- Drain pbx's 95-item audit ratchet and go-taskqueue's 144-item
  annotate ratchet as docs-health rituals run.
- library-policy scoped-exception mechanism (upstream feature ask) so
  the md5 census-fingerprint verdict stops tripping the error gate.

## g) Questions for the owner (unanswerable from here)

1. **GitHub Releases channel**: docs-gates has never used one — should
   v0.5.0+ publish release artifacts there, or is tags+CHANGELOG the
   permanent surface?
2. **The evo-x2/SystemNix window**: float+deploy+switch are queued
   behind your quiescent-window doctrine — when?
3. **library-policy md5**: accept the standing known-delta here, or
   sponsor the upstream scoped-exception feature (it affects any repo
   with deliberate non-security md5)?
