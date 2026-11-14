> [!NOTE]
> This filing was drafted by GLM-5.3 via Crush from an AI-run investigation, not at my request.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

**TL;DR:** `pre-commit install` refuses while a global-scope `core.hooksPath` is set, but its hint only shows the repo-local unset — after following it I was still refused, and nothing in the output says where the value lives. Ask: have the refusal name the scope (`git config --show-origin --get-all core.hooksPath`), so the next user sees which config file holds the value.

## Problem

Live run today (pre-commit installed via nix, git 2.51):

1. `pre-commit install` in a clean repo → refused with the hint `git config --unset-all core.hooksPath`.
2. `git config --unset core.hooksPath` (repo-local) → still refused, same hint.
3. `git config --show-origin --get-all core.hooksPath` → the value lived in the global `~/.gitconfig`, pointing at a hooks directory that does not exist.
4. `git config --global --unset-all core.hooksPath` → install proceeds.

The refusal itself is right by design (an install into `.git/hooks` would be ignored under `hooksPath`); the gap is that the hint cannot distinguish "unset it here" from "it is not set here at all".

## Evidence

- Refusal + hint: `pre_commit/commands/install_uninstall.py`, `install()` (master, `Cowardly refusing` block).
- The incomplete-hint class is already tracked: #3571 (global `core.hooksPath`, closed COMPLETED without the hint changing on master) and #3570 (closed as filed to a different project).

## Ask

Name the scope in the refusal: print the origin line (the output of `git config --show-origin --get-all core.hooksPath`) above the hint, or minimally extend the static hint to point at `--show-origin --get-all` first. A dead global value like mine produces a refusal whose hint sends the user to the wrong scope twice.

💘 Generated with Crush
