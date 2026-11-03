#!/usr/bin/env bash
# stack-drift-check: fail when the root go.mod gains a require/replace
# without a STACK.md row (the stdlib-only contract, AGENTS.md).
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

if grep -qE '^\s*(require|replace|exclude)\s' go.mod; then
	echo "FATAL: root go.mod carries require/replace/exclude — the stdlib-only" >&2
	echo "contract forbids it. A third-party dependency needs a STACK.md row and a" >&2
	echo "separate go module (see AGENTS.md); the root module stays vendor-free." >&2
	exit 1
fi

echo "stack-drift: root go.mod is dependency-free — OK"
