# Draft: BuildFlow issue — nix-flake-update blind repair

- Target: LarsArtmann/BuildFlow (own repo)
- Genre: bug report, body register, board-directed (no unsolicited banner), AI-drafted (Crush footer)
- Verify-before-filing gates: 5/5 passed (reproduced live, source-level diagnosis, own-code contrast, fleet-wide scope, no duplicates open/closed)

## Title

nix-flake-update repair runs bare `nix flake update`: silently advances every master-tracking input

## Body

## Problem

`buildflow --fix` rewrites flake.lock entries for inputs that have no drift finding against them. In a repo whose flake carries private, master-tracking inputs whose locked rev IS behavior (erraudit's rules), the repair step silently advances them to latest master — and the auto-commit daemon later wraps the rewrite into a commit nobody reviewed.

Six damage events in docs-gates so far (11-06, 11-14, and three on 11-23 — the last one my own deliberate repro): each restored forward from the pinned revs.

## Repro (live, today)

```
$ cd docs-gates            # lock pins erraudit @ dccfd10, tripwire green
$ buildflow --fix
...
05:13:24 INFO  ✔ nix-flake-update 5.0s
...
05:13:24 INFO  nix> warning: updating lock file ".../docs-gates/flake.lock":
05:13:24 INFO  nix> • Updated input 'erraudit':
05:13:24 INFO  nix> • Updated input 'erraudit/go-composable-business-types':
05:13:24 INFO  nix> • Updated input 'erraudit/go-filewatcher':
05:13:24 INFO  nix> • Updated input 'erraudit/go-ndjson':
```

Result: erraudit moved dccfd10 → cb96501 with its whole subtree, zero findings driving it. A plain `buildflow` run (no `--fix`) leaves the lock byte-identical; bare `nix flake check` also leaves it unchanged — the repair phase owns the rewrite.

## Source

`tools/providers/nix_flake_check_tools.go` (e9e3cd3c3), `NewNixFlakeUpdateProvider`:

```go
Repairer: blindRepairer(
    runner,
    name,
    "nix",
    []string{"flake", "update"},   // bare: EVERY input
    []file.FilePattern{flakeLockPattern},
),
```

`Trigger: flakeTrigger("nix")` — it fires for any nix repo, network on, and being a blind repairer it has no finding to answer for. Meanwhile BuildFlow's own `upgrade_cmd.go:158` refuses this exact command shape:

> bare 'nix flake update' would rewrite EVERY input of whatever project the [...]

## Impact

- Private pins advance silently; the locked rev was chosen deliberately (rev-pin as explicit decision — the fetcher repair's own doc comment says so).
- Daemon-wrapped damage commits: the rewrite lands as `chore: auto-commit`, history looks intentional, nothing reviewed it.
- Broken devShells: an advanced input rev can fail to build (11-23: erraudit 8831ce2 broke `nix develop` in docs-gates for a while).

## Proposed change (minimal)

Make the update finding-driven, matching what `flake_input_repair.go` already does per drift:

```go
cmd := exec.CommandContext(ctx, "nix", "flake", "update", drift.InputName)
```

Either restrict the repairer to inputs with an actual drift finding (go.mod vs flake), or drop `nix-flake-update` from the `--fix` DAG entirely and leave whole-lock refreshes to `buildflow update` / `buildflow upgrade`, where the operator asked for input movement.

## Workaround in place (not a fix)

docs-gates now runs `scripts/check-erraudit-rev.sh` as a pre-commit leg whenever flake.lock is staged, so the damage class refuses at commit time instead of surfacing at the next push. That catches the symptom after the fact; the repair step is still the cause.


💘 Generated with Crush

## Filing notes (not part of the body)

- Post with: `gh issue create --repo LarsArtmann/BuildFlow --title "…" --body-file <this-file-extracted>`
- Extract only the Body section (fenced blocks inside — use python split, not sed).
- Verify landed with `gh issue view`.
