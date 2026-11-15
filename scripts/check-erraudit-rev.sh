#!/usr/bin/env bash
# Master-tracking flake-input rev tripwires.
#
# Some inputs track a moving ref (erraudit: default branch; go-filewatcher,
# go-nix-helpers, gogenfilter: master) with the rev locked in flake.lock —
# but `nix flake update <input>` (and BuildFlow's repair phase) bumps those
# revs silently, and for erraudit the rules ARE behavior: a new rev can flag
# new sites or stop flagging old ones. This script fails until every new rev
# is recorded below, making each bump deliberate: run the checks on the new
# rev first, then update that ONE expected pin (one input at a time — a
# multi-input bump entangles rule changes with dependency noise and hides
# which update changed behavior).
#
# Precedent: go-codec scripts/check-erraudit-version.sh (the same
# two-place agreement, for an install-line pin).
#
# Usage: scripts/check-erraudit-rev.sh (run from anywhere; cd's to root)

set -euo pipefail

cd "$(dirname "$0")/.."

# input-name=expected locked rev (the deliberate-bump pins)
declare -A expected=(
	[erraudit]="85bc703f8556655b1b427bb427c709faee592431"
	[go-filewatcher]="a965e6b8be500594e117eb52e4269fefa90cfcfb"
	[go-nix-helpers]="2c5d1d75aeacaaf6ddf558fd7d7bd5a4f4f7687a"
	[gogenfilter]="ebdac855fea9ed24ed20e4e0590ab80fd24d826d"
)

failed=0

for name in erraudit go-filewatcher go-nix-helpers gogenfilter; do
	locked_rev=$(jq -r --arg n "$name" '.nodes[$n].locked.rev // empty' flake.lock)
	if [ -z "$locked_rev" ]; then
		echo "FAIL: flake.lock has no locked rev for the $name input" >&2
		failed=1
		continue
	fi
	if [ "$locked_rev" != "${expected[$name]}" ]; then
		echo "FAIL: $name rev drifted: flake.lock=$locked_rev expected=${expected[$name]}" >&2
		echo "      nix flake update $name changed locked behavior — run the checks on the" >&2
		echo "      new rev, then bump ONLY its expected pin in scripts/check-erraudit-rev.sh" >&2
		echo "      (one input at a time; entangled bumps hide which update changed behavior)" >&2
		failed=1
		continue
	fi
	echo "rev-tripwire: PASS $name (${locked_rev:0:12})"
done

exit "$failed"
