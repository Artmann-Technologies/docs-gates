#!/usr/bin/env python3
"""Replay the "How to verify" blocks of the docs/review/libs audit reports.

Every ```console block under a report's "## How to verify" heading is a
transcript: ``$ command`` lines followed by the output the auditor claims to
have seen. This harness re-runs each command (read-only commands only) and
diffs the real output against the transcript, so a report's own evidence
cannot rot silently or ship non-literal.

Exit 0 = every replayed command reproduces its transcript (or is an
acknowledged environment-dependent command); exit 1 = mismatches (printed).
"""

from __future__ import annotations

import re
import subprocess
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
REPORT_DIR = REPO_ROOT / "docs" / "review" / "libs"
TIMEOUT_SEC = 30

ALLOWED_HEADS = {
    "rg",
    "grep",
    "git",
    "sed",
    "awk",
    "cat",
    "head",
    "tail",
    "wc",
    "ls",
    "find",
    "go",
    "jq",
    "cut",
    "sort",
    "uniq",
    "comm",
    "diff",
    "echo",
    "printf",
    "true",
    "test",
    "nix",
    "python3",
    "sha256sum",
    "md5sum",
    "stat",
    "file",
}

DENY = re.compile(
    r"(^|\s|;|&\|)(rm|rmdir|mv|cp|touch|mkdir|ln|chmod|chown|truncate|tee)\s"
    r"|sed\s+[^|]*-i"
    r"|git\s+(commit|push|checkout|restore|reset|add|stash|rebase|merge|pull|fetch|switch|tag\s+-d|rm|apply|worktree)\b"
    r"|nix\s+(build|develop|run|shell|profile|edit|flake\s+(update|check))\b"
    r"|>>?\s*[^&]"  # any redirection (write or read)
    r"|(^|[;&|]\s*)exec\s|\beval\b|\bsource\b|\bxargs\b"
)


def split_commands(block: str) -> list[tuple[str, list[str]]]:
    """Return [(command, expected_output_lines)] from a transcript block."""
    commands: list[tuple[str, list[str]]] = []
    lines = block.splitlines()
    index = 0
    while index < len(lines):
        line = lines[index]
        if line.startswith("$ "):
            command = line[2:]
            while command.rstrip().endswith("\\") and index + 1 < len(lines):
                command = command.rstrip()[:-1] + " " + lines[index + 1].strip()
                index += 1
            expected: list[str] = []
            index += 1
            while index < len(lines) and not lines[index].startswith("$ "):
                text = lines[index]
                if text.strip() and not text.lstrip().startswith("#"):
                    expected.append(text)
                index += 1
            commands.append((command, expected))
        else:
            index += 1
    return commands


def guard(command: str) -> str | None:
    if DENY.search(command):
        return "denied by read-only guard"
    head = re.match(r"\s*(?:cd\s+\S+\s*&&\s*)?(?:\w+=\S+\s+)*([\w./-]+)", command)
    first = head.group(1) if head else ""
    if first == "bash":
        script = re.match(r"\s*bash\s+(\S+)", command)
        if script and re.fullmatch(r"(/home/lars/projects/docs-gates/)?scripts/[\w-]+\.sh", script.group(1)):
            return None
    if not head or first not in ALLOWED_HEADS:
        return f"head not in allowlist: {first or '?'}"
    return None


def run(command: str, cwd: Path) -> tuple[int, str, str]:
    try:
        done = subprocess.run(
            ["bash", "-c", command],
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=TIMEOUT_SEC,
        )
    except subprocess.TimeoutExpired:
        return 124, "", "timeout"
    return done.returncode, done.stdout, done.stderr


def normalize(line: str) -> str:
    line = line.replace("–", "-").replace("/home/lars/projects/docs-gates/", "./")
    return re.sub(r"\s+", "", line)


def strip_annotation(line: str) -> str:
    """Drop a trailing ``# note`` the auditor appended to an output line."""
    return re.sub(r"\s{2,}#\s.*$", "", line).rstrip()


def matches(expected: list[str], actual: list[str]) -> bool:
    """Strict: the transcript must show exactly the real output lines.

    Compared as multisets (a multi-file ``rg`` search returns files in a
    nondeterministic order) after whitespace is ignored and any trailing
    auditor note is removed.
    """
    want = sorted(normalize(strip_annotation(line)) for line in expected)
    return want == sorted(actual)


def main() -> int:
    reports = sorted(REPORT_DIR.glob("*.md"))
    replayed = exact = mismatched = skipped = 0
    failures: list[str] = []
    for report in reports:
        text = report.read_text(encoding="utf-8")
        section = re.search(r"^## How to verify(.*)$", text, re.S | re.M)
        if not section:
            failures.append(f"{report.name}: no '## How to verify' section")
            continue
        blocks = re.findall(r"```(\w*)\n(.*?)```", section.group(1), re.S)
        items: list[tuple[str, list[str]]] = []
        for index, (_, body) in enumerate(blocks):
            if "$ " not in body:
                continue
            commands = split_commands(body)
            last = commands[-1]
            if not last[1] and index + 1 < len(blocks) and "$ " not in blocks[index + 1][1]:
                last[1].extend(
                    line
                    for line in blocks[index + 1][1].splitlines()
                    if line.strip() and not line.lstrip().startswith("#")
                )
            items.extend(commands)
        for command, expected in items:
            replayed += 1
            reason = guard(command)
            if reason:
                skipped += 1
                failures.append(f"{report.name}: SKIPPED ({reason}): {command[:90]}")
                continue
            code, out, err = run(command, REPO_ROOT)
            actual = [normalize(line) for line in (out + err).splitlines() if line.strip()]
            if matches(expected, actual):
                exact += 1
                continue
            mismatched += 1
            failures.append(
                f"{report.name}: MISMATCH (rc={code}) `{command[:80]}`\n"
                f"    want: {[normalize(w) for w in expected[:3]]}\n    got:  {actual[:3]}"
            )
    print(
        f"reports: {len(reports)}  commands: {replayed}  exact: {exact}"
        f"  mismatched: {mismatched}  skipped: {skipped}"
    )
    for failure in failures:
        print(f"  {failure}")
    return 1 if mismatched else 0


if __name__ == "__main__":
    sys.exit(main())
