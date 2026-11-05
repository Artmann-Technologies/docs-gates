#!/usr/bin/env bash
# safe-strike-test: the refusal-matrix harness for safe-strike.sh (the
# 2026-11-04 corruption class — every rail must refuse loudly, never
# guess). Nine cases on a scratch git repo + board fixture:
#
#   1. anchor matching ZERO lines        -> refuse, exit 1, file intact
#   2. anchor matching TWO lines         -> refuse, exit 1, full inventory, file intact
#   3. ordinal-only anchor (<4 letters)  -> refuse, exit 1, file intact
#   4. missing file                      -> usage, exit 2
#   5. whole-line strike without verdict -> usage, exit 2
#   6. --cell with a verdict             -> usage, exit 2
#   7. --cell + --set-cell same column   -> usage, exit 2
#   8. dry-run default                   -> exit 0, file UNCHANGED
#   9. --apply on a clean repo           -> exit 0, line struck, neighbors intact;
#      plus the dirty-file refusal (10) and the beyond-columns cell (11)
#
# Exit codes: 0 all cases behaved, 1 a rail failed to refuse (or struck
# wrongly), 2 harness setup failure.
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
strike="$here/safe-strike.sh"

scratch="$(mktemp -d)"
trap 'rm -rf "$scratch"' EXIT

git init -q "$scratch/board"
cd "$scratch/board"

cat >TODOS.md <<'EOF'
# Board

| # | Task | Status |
| --- | --- | --- |
| 1 | wire the alpha relay harness | TODO |
| 2 | wire the alpha relay harness twice | TODO |
| 3 | ship the beta lantern fixture | TODO |
| 4 | quux thing | TODO |
EOF

git add TODOS.md
git -c user.name=test -c user.email=test@test commit -qm "fixture"

fail=0

expect_rc() {
	# expect_rc <want-rc> <case-name> <args...>
	local want="$1" name="$2"
	shift 2

	local got=0
	"$strike" "$@" >/dev/null 2>&1 || got=$?

	if [ "$got" -ne "$want" ]; then
		echo "FAIL [$name]: exit $got, want $want" >&2
		fail=1
	else
		echo "ok   [$name] exit $got"
	fi
}

expect_untouched() {
	if ! git diff --quiet; then
		echo "FAIL [$1]: the file changed under a refusal" >&2
		fail=1
	else
		echo "ok   [$1] file untouched"
	fi
}

expect_rc 1 "zero-match anchor" TODOS.md 'nonexistent-zephyr-anchor' 'done'
expect_untouched "zero-match anchor"

expect_rc 1 "two-match anchor" TODOS.md 'wire the alpha relay harness' 'done'
expect_untouched "two-match anchor"

expect_rc 1 "ordinal-only anchor" TODOS.md '^\| 3 \|' 'done'
expect_untouched "ordinal-only anchor"

expect_rc 2 "missing file" TODOS.md.absent 'alpha relay' 'done'
expect_rc 2 "no verdict" TODOS.md 'alpha relay harness \|'
expect_rc 2 "cell with verdict" --cell 2 TODOS.md 'alpha relay harness \|' 'done'
expect_rc 2 "same cell+set-cell" --cell 2 --set-cell 2=x TODOS.md 'alpha relay harness \|'

# 8: dry-run default leaves the file byte-identical.
expect_rc 0 "dry-run default" TODOS.md 'beta lantern fixture' 'done at `deadbee`'
expect_untouched "dry-run default"

# 9: apply strikes exactly the anchored line and nothing else.
expect_rc 0 "apply whole line" --apply TODOS.md 'beta lantern fixture' 'done at `deadbee`'
if [ "$(grep -c 'beta lantern' TODOS.md)" -ne 1 ] || ! grep -q '~~| 3 | ship the beta lantern fixture | TODO |~~ done at `deadbee`' TODOS.md; then
	echo "FAIL [apply whole line]: struck shape wrong:" >&2
	cat TODOS.md >&2
	fail=1
else
	echo "ok   [apply whole line] struck shape"
fi
git add TODOS.md
git -c user.name=test -c user.email=test@test commit -qm strike

# 10: a dirty file refuses without --allow-dirty.
echo "draft line" >>TODOS.md
expect_rc 1 "dirty refusal" --apply TODOS.md 'alpha relay harness \| TODO \|' 'done'
expect_rc 0 "dirty override" --allow-dirty --apply TODOS.md 'alpha relay harness \| TODO \|' 'done'
git restore --source=HEAD -- TODOS.md
# (restore the fixture to the committed struck state for cleanliness)

# 11: a cell beyond the row's column count refuses and leaves the row intact.
expect_rc 1 "beyond-column cell" --apply --cell 9 TODOS.md 'quux thing'

if [ "$fail" -eq 0 ]; then
	echo "safe-strike-test: ALL cases behaved"
	exit 0
fi

echo "safe-strike-test: FAILURES above" >&2
exit 1
