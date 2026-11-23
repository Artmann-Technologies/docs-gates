#!/usr/bin/env python3
"""Citation sweep for the docs/review/libs audit reports.

Extracts every ``path:line`` (and ``path:line-line``) citation from the
report corpus and verifies that the file exists and the line range is in
bounds — against BOTH the docs-gates repo and the audited library
checkouts under ~/projects.

Resolution order for a cited path:
  1. absolute path (must exist)
  2. relative to the docs-gates repo root
  3. <lib-checkout-name>/rest/of/path under ~/projects
  4. bare basename: searched in docs-gates, then each library checkout
     (reported as ambiguous-but-resolved)

Exit 0 = every citation resolves; exit 1 = failures (printed).
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
PROJECTS = Path.home() / "projects"
REPORT_DIRS = [REPO_ROOT / "docs" / "review" / "libs"]

LIB_NAMES = [
    "go-error-family",
    "go-idempotency",
    "go-retry",
    "go-business-rules",
    "go-etag",
    "httputil",
    "go-codec",
    "go-sse",
    "go-datastar",
    "go-atomic-write",
    "go-filewatcher",
    "go-daemon",
    "go-output",
    "go-health",
    "go-health-dashboard",
    "go-credentials-store",
    "cmdguard",
    "templ-components",
    "go-appkit",
    "go-cqrs-lite",
    "cqrs-htmx",
    "go-finding",
    "erraudit",
]

CITE_RE = re.compile(
    r"(?<![\w/.-])"
    r"((?:[\w.-]+/)*[\w.-]+\.(?:go|md|nix|sh|json|mod|sum|ya?ml|toml|templ|ts|html))"
    r":(\d+)(?:[–-](\d+))?"
)


def normalize_dash(text: str) -> str:
    return text.replace("–", "-").replace("—", "-")


def iter_reports() -> list[Path]:
    files: list[Path] = []
    for directory in REPORT_DIRS:
        if directory.is_dir():
            files.extend(sorted(directory.glob("*.md")))
    return files


def consumer_basenames() -> dict[str, Path]:
    index: dict[str, Path] = {}
    for path in REPO_ROOT.rglob("*"):
        if path.is_file() and ".git/" not in str(path):
            index.setdefault(path.name, path)
    return index


def resolve(path_text: str, basenames: dict[str, Path]) -> tuple[Path | None, str]:
    """Return (resolved path, how) or (None, '')."""
    candidate = Path(path_text)
    if candidate.is_absolute():
        return (candidate, "absolute") if candidate.exists() else (None, "")
    rel = REPO_ROOT / candidate
    if rel.exists():
        return rel, "consumer"
    first, _, rest = path_text.partition("/")
    if first in LIB_NAMES and rest:
        lib = PROJECTS / first / rest
        if lib.exists():
            return lib, "lib-qualified"
    if "/" not in path_text:
        consumer_hit = basenames.get(path_text)
        if consumer_hit is not None:
            return consumer_hit, "consumer-basename"
        for lib in LIB_NAMES:
            lib_file = PROJECTS / lib / path_text
            if lib_file.exists():
                return lib_file, f"basename-in-{lib}"
    return None, ""


def main() -> int:
    reports = iter_reports()
    if not reports:
        print("no reports found under docs/review/libs/")
        return 1
    total = 0
    failures: list[str] = []
    ambiguous: list[str] = []
    basenames = consumer_basenames()
    for report in reports:
        text = normalize_dash(report.read_text(encoding="utf-8"))
        for lineno, line in enumerate(text.splitlines(), start=1):
            for match in CITE_RE.finditer(line):
                path_text, start, end = match.group(1), match.group(2), match.group(3)
                if path_text.startswith("http") or "/" in path_text and path_text.split("/")[0] in ("http:", "https:"):
                    continue
                total += 1
                resolved, how = resolve(path_text, basenames)
                where = f"{report.name}:{lineno}"
                if resolved is None:
                    failures.append(f"{where}: `{match.group(0)}` — file not found")
                    continue
                if how.startswith("basename"):
                    ambiguous.append(f"{where}: `{match.group(0)}` -> {resolved} ({how})")
                line_count = sum(1 for _ in resolved.open(encoding="utf-8", errors="replace"))
                first_line = int(start)
                last_line = int(end) if end else first_line
                if first_line > line_count or last_line > line_count:
                    failures.append(
                        f"{where}: `{match.group(0)}` — out of range "
                        f"(file has {line_count} lines)"
                    )
    print(f"reports: {len(reports)}  citations: {total}")
    for note in ambiguous:
        print(f"  ambiguous {note}")
    if failures:
        print(f"FAILURES: {len(failures)}")
        for failure in failures:
            print(f"  {failure}")
        return 1
    print("all citations resolve; all line ranges in bounds")
    return 0


if __name__ == "__main__":
    sys.exit(main())
