#!/usr/bin/env bash
# erraudit flake-input rev tripwire.
#
# The erraudit input tracks master (BuildFlow 2026-10-05 track-latest
# policy) with the rev locked in flake.lock — but `nix flake update
# erraudit` bumps that rev silently, and erraudit's rules ARE behavior: a
# new rev can flag new sites (fails the erraudit-lint check, fine) or
# stop flagging old ones (silently). This script fails until the new rev
# is recorded below, making every rule change deliberate: run the
# erraudit checks on the new rev first, then bump expected_rev.
#
# Precedent: go-codec scripts/check-erraudit-version.sh (the same
# two-place agreement, for an install-line pin).
#
# Usage: scripts/check-erraudit-rev.sh (run from anywhere; cd's to root)

set -euo pipefail

cd "$(dirname "$0")/.."

expected_rev="3d1f7ed37a1c52a7cc8b9c3e1749a2c8a423256f"

locked_rev=$(jq -r '.nodes.erraudit.locked.rev // empty' flake.lock)
if [ -z "$locked_rev" ]; then
	echo "FAIL: flake.lock has no locked rev for the erraudit input" >&2
	exit 1
fi
if [ "$locked_rev" != "$expected_rev" ]; then
	echo "FAIL: erraudit rev drifted: flake.lock=$locked_rev expected=$expected_rev" >&2
	echo "      nix flake update erraudit changed rule behavior — run the erraudit checks," >&2
	echo "      then bump expected_rev in scripts/check-erraudit-rev.sh" >&2
	exit 1
fi
echo "erraudit-rev: PASS (flake.lock matches pinned rev ${expected_rev:0:12})"
