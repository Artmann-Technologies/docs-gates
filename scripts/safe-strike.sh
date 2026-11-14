#!/usr/bin/env bash
# safe-strike: the strike-sweep safe mechanics as a tool (the 2026-11-04
# corruption class, AGENTS.md + the 23-06 self-review §d/§e):
#
#   - the anchor must match EXACTLY ONE line — 0 or >1 matches is a
#     refusal that prints every match in full (inventory is never
#     truncated)
#   - cell captures are [^|]*-bounded — a strike can never swallow a
#     neighboring cell (the greedy `(.*) (\\|)$` collapse)
#   - ordinal-shaped anchors are refused: dump-item ordinals and table
#     row numbers are different numbering spaces; anchor on task TEXT
#   - the file must be git-clean (one strike batch = one commit
#     boundary; corrupt intermediates must never mix into uncommitted
#     work) unless --allow-dirty
#   - --dry-run is the default; only --apply writes
#   - --gate runs the annotate gate after an apply and refuses non-zero
#   - --battery runs the WHOLE battery after an apply instead (superset:
#     a strike can flip freshness/cadence too, not just annotate)
#   - --batch strikes MULTIPLE rows in one invocation: each positional
#     is ANCHOR=VERDICT; every anchor must uniquely match first (the
#     batch is all-or-nothing — one bad anchor refuses everything), the
#     write is one pass, and the gate/battery leg runs ONCE for the batch
#
# Usage:
#   safe-strike.sh [--apply] [--gate] [--battery] [--allow-dirty]
#                  [--cell N] [--set-cell M=TEXT] [--no-verify]
#                  <file> <anchor-regex> [verdict]
#   safe-strike.sh [--apply] [--gate|--battery] [--allow-dirty] [--batch]
#                  [--no-verify] <file> ANCHOR=VERDICT [ANCHOR=VERDICT ...]
#
# Modes:
#   default      strike the whole line: `~~<line>~~ <verdict>`
#                (verdict required, e.g. "done at `abc1234`")
#   --cell N     strike only table cell N, keep the rest of the row
#                (pair with --set-cell to flip a status cell)
#   --set-cell M=TEXT
#                replace table cell M's content with TEXT verbatim
#
# Exit codes: 0 struck (or clean dry-run), 1 refusal (anchor/shape/dirty/
# gate findings), 2 usage.
set -euo pipefail

usage() {
	echo "usage: safe-strike.sh [--apply] [--gate] [--battery] [--allow-dirty] [--cell N] \\" >&2
	echo "                      [--set-cell M=TEXT] [--no-verify] <file> <anchor-regex> [verdict]" >&2
	echo "       safe-strike.sh [--apply] [--gate|--battery] [--batch] [--no-verify] \\" >&2
	echo "                      <file> ANCHOR=VERDICT [ANCHOR=VERDICT ...]" >&2
	echo "got: safe-strike.sh $*" >&2
	exit 2
}

apply=0
gate=0
battery=0
allow_dirty=0
cell=0
set_cell=""
no_verify=0
batch=0

while [ $# -gt 0 ]; do
	case "$1" in
	--apply) apply=1 ;;
	--gate) gate=1 ;;
	--battery) battery=1 ;;
	--allow-dirty) allow_dirty=1 ;;
	--no-verify) no_verify=1 ;;
	--batch) batch=1 ;;
	--cell)
		[ $# -ge 2 ] || usage "$@"
		cell="$2"
		shift
		;;
	--set-cell)
		[ $# -ge 2 ] || usage "$@"
		set_cell="$2"
		shift
		;;
	-h | --help)
		sed -n '2,52p' "$0" | sed 's/^# \{0,1\}//'
		exit 0
		;;
	-*)
		echo "FATAL: unknown flag: $1" >&2
		usage "$@"
		;;
	*)
		break
		;;
	esac
	shift
done

if [ "$batch" -eq 1 ]; then
	if [ "$cell" != "0" ] || [ -n "$set_cell" ]; then
		echo "FATAL: --batch strikes whole lines — --cell/--set-cell are single-strike options" >&2
		usage "$@"
	fi

	[ $# -ge 2 ] || {
		[ $# -ge 1 ] || echo "FATAL: missing <file> — the TODO file the batch strikes (e.g. TODO_LIST.md)" >&2
		echo "FATAL: --batch needs at least one ANCHOR=VERDICT positional after the file" >&2
		usage "$@"
	}
else
	if [ $# -lt 2 ] || [ $# -gt 3 ]; then
		[ $# -ge 1 ] || echo "FATAL: missing <file> — the TODO file to strike (e.g. TODO_LIST.md)" >&2
		[ $# -ge 2 ] || echo "FATAL: missing <anchor-regex> — a unique task-TEXT substring of the row" >&2
		[ $# -le 3 ] || echo "FATAL: $# positionals is too many — <file> <anchor-regex> [verdict] takes at most 3" >&2
		usage "$@"
	fi
fi
if [ "$batch" -eq 1 ]; then
	file="$1"
else
	file="$1"
	anchor="$2"
	verdict="${3:-}"

	if [ "$cell" = "0" ] && [ -z "$set_cell" ] && [ -z "$verdict" ]; then
		echo "FATAL: a whole-line strike needs a verdict (e.g. \"done at \`abc1234\`\");" >&2
		echo "cell-only operations may omit it" >&2
		usage "$@"
	fi

	if [ "$cell" != "0" ] && [ -n "$set_cell" ] && [ "$cell" = "${set_cell%%=*}" ]; then
		echo "FATAL: --cell and --set-cell name the same column" >&2
		usage "$@"
	fi

	if [ "$cell" != "0" ] && [ -n "$verdict" ]; then
		echo "FATAL: --cell mode takes no verdict — a verdict appended after a table" >&2
		echo "row would become a phantom extra column; use --set-cell M=TEXT to flip a" >&2
		echo "status cell instead" >&2
		usage "$@"
	fi
fi
shift || true

[ -f "$file" ] || {
	echo "FATAL: not a file: $file" >&2
	exit 2
}

# The git-clean rail. A dirty file means uncommitted work the daemon has
# not yet swept — striking on top of it entangles a corrupt intermediate
# with someone else's edit. One batch, one commit boundary.
if [ "$allow_dirty" -eq 0 ] && [ "$apply" -eq 1 ]; then
	if [ -n "$(git status --porcelain -- "$file")" ]; then
		echo "FATAL: $file has uncommitted changes — commit (or sweep) first," >&2
		echo "so this strike lands as its own reviewable boundary (--allow-dirty overrides)" >&2
		exit 1
	fi
fi

if [ "$batch" -eq 1 ]; then
	# Batch mode: every positional is ANCHOR=VERDICT. Verify EVERY anchor
	# (unique match + task-text rail) before any write; one bad anchor
	# refuses the whole batch. Strikes are in-place line rewrites, so the
	# resolved line numbers stay valid across the whole batch.
	entries=""
	seen_lines=""

	for pair in "$@"; do
		case "$pair" in
		*=*) anchor="${pair%%=*}" verdict="${pair#*=}" ;;
		*)
			echo "FATAL: batch entry [$pair] carries no =VERDICT — batch positionals are ANCHOR=VERDICT" >&2
			exit 2
			;;
		esac

		if [ -z "$verdict" ]; then
			echo "FATAL: batch entry [$pair] has an empty verdict — a whole-line strike needs evidence" >&2
			exit 2
		fi

		matches="$(grep -nE -- "$anchor" "$file" || true)"
		count="$(printf '%s\n' "$matches" | grep -c . || true)"

		if [ "$count" -ne 1 ]; then
			echo "FATAL: batch anchor [$anchor] matches $count line(s) (want exactly 1) — full inventory:" >&2
			[ -n "$matches" ] && printf '%s\n' "$matches" >&2
			echo "refusing the whole batch; anchor on longer unique task TEXT" >&2
			exit 1
		fi

		if ! printf '%s' "$anchor" | tr -cd '[:alpha:]' | grep -q '[a-zA-Z][a-zA-Z][a-zA-Z][a-zA-Z]'; then
			echo "FATAL: batch anchor [$anchor] carries no task text — digits and scaffolding" >&2
			echo "can match any numbering space (the 2026-11-04 corruption)" >&2
			exit 1
		fi

		line_no="${matches%%:*}"
		line_content="${matches#*:}"

		if printf '%s' "$seen_lines" | grep -qx "$line_no"; then
			echo "FATAL: two batch anchors hit the same line $line_no — a line strikes once per batch" >&2
			exit 1
		fi

		seen_lines="${seen_lines}${line_no}\n"
		echo "safe-strike: target [$file:$line_no]"
		echo "  before: $line_content"
		entries="${entries}${line_no}$(printf '\t%s\t%s' "$(printf '%s' "$line_content" | base64 -w0)" "$(printf '%s' "$verdict" | base64 -w0)")"$'\n'
	done

	if [ "$apply" -eq 0 ]; then
		echo "safe-strike: dry-run only — pass --apply to write the whole batch (gate once after)"
		exit 0
	fi

	STRIKE_BATCH="$entries" perl -e '	use strict; use warnings;
	use MIME::Base64 qw(decode_base64);
	my ($file) = @ARGV;
	open(my $in, "<", $file) or die "FATAL: read: $!";
	binmode $in; my @lines = <$in>; close $in;
	for my $row (split /\n/, $ENV{STRIKE_BATCH}) {
		my ($n, $want, $verdict) = split /\t/, $row;
		die "FATAL: line $n vanished under us\n" if !defined $lines[$n-1];
		(my $have = $lines[$n-1]) =~ s/\r?\n\z//;
		die "FATAL: line $n changed between match and write; re-run\n"
			if $have ne decode_base64($want);
		$lines[$n-1] = "~~$have~~ " . decode_base64($verdict) . "\n";
	}
	open(my $fh, ">", $file) or die "FATAL: write: $!";
	binmode $fh; print {$fh} @lines; close $fh;
' "$file" || {
		echo "safe-strike: refused — the file is untouched" >&2
		exit 1
	}

	printf '%s' "$seen_lines" | while read -r n; do
		[ -n "$n" ] || continue
		echo "   after: $(sed -n "${n}p" "$file")"
	done

	if [ "$no_verify" -eq 0 ] && [ "$gate" -eq 1 ]; then
		echo "safe-strike: gate per batch — running the annotate gate ONCE for the batch"
		if command -v docs-gates >/dev/null 2>&1; then
			docs-gates annotate .
		else
			go run ./cmd/docs-gates annotate .
		fi
	fi

	if [ "$no_verify" -eq 0 ] && [ "$battery" -eq 1 ]; then
		echo "safe-strike: battery per batch — running all five gates ONCE for the batch"
		if command -v docs-gates >/dev/null 2>&1; then
			docs-gates run .
		else
			go run ./cmd/docs-gates run .
		fi
	fi

	echo "safe-strike: struck $(( $(printf '%s' "$seen_lines" | grep -c .) )) row(s) — review the diff, run the battery, commit"
	exit 0
fi

# The uniqueness rail: EXACTLY one matching line, shown in full.
matches="$(grep -nE -- "$anchor" "$file" || true)"
count="$(printf '%s\n' "$matches" | grep -c . || true)"

if [ "$count" -ne 1 ]; then
	echo "FATAL: anchor matches $count line(s) (want exactly 1) — full inventory:" >&2
	if [ -n "$matches" ]; then
		printf '%s\n' "$matches" >&2
	fi
	echo "refusing to guess; anchor on longer unique task TEXT" >&2
	exit 1
fi

line_no="${matches%%:*}"
line_content="${matches#*:}"

# The ordinal rail: a bare row-number anchor is the 2026-11-04 burn —
# dump ordinals and table row numbers are different numbering spaces.
# Anchors carry task TEXT: fewer than 4 letters means the anchor is
# only digits and regex scaffolding (`3`, `^\| 3 \|`), which can match
# any numbering space and must never select a row.
if ! printf '%s' "$anchor" | tr -cd '[:alpha:]' | grep -q '[a-zA-Z][a-zA-Z][a-zA-Z][a-zA-Z]'; then
	echo "FATAL: the anchor carries no task text ($anchor) — an anchor of only" >&2
	echo "digits and regex scaffolding can match any numbering space, and dump" >&2
	echo "ordinals with table row numbers are different spaces (the 2026-11-04" >&2
	echo "corruption). Anchor on the row's unique task words instead." >&2
	exit 1
fi

echo "safe-strike: target [$file:$line_no]"
echo "  before: $line_content"

if [ "$apply" -eq 0 ]; then
	echo "safe-strike: dry-run only — pass --apply to write (gate per batch: apply one row, run --gate, commit)"
	exit 0
fi

export STRIKE_LINE="$line_no"
export STRIKE_CELL="$cell"
export STRIKE_SET="$set_cell"
verdict_b64="$(printf '%s' "$verdict" | base64 -w0)"
line_want_b64="$(printf '%s' "$line_content" | base64 -w0)"
export STRIKE_VERDICT="$verdict_b64"
export STRIKE_FILE_WANT="$line_want_b64"

# perl does the single-line rewrite with bounded captures; STRIKE_FILE_WANT
# re-verifies the line byte-for-byte between match and write, so a file
# changed under us cannot be struck at the wrong row.
if ! perl -e '	use strict; use warnings;
	my ($file) = @ARGV;
	open(my $in, "<", $file) or die "FATAL: read: $!";
	binmode $in; my @lines = <$in>; close $in;
	my $n = $ENV{STRIKE_LINE};
	die "FATAL: line $n vanished under us\n" if !defined $lines[$n-1];
	use MIME::Base64 qw(decode_base64);
	my $expect = decode_base64($ENV{STRIKE_FILE_WANT});
	(my $have = $lines[$n-1]) =~ s/\r?\n\z//;
	die "FATAL: line $n changed between match and write; re-run\n"
		if $have ne $expect;
	my $out;
	if (my $sc = $ENV{STRIKE_SET}) {
		my ($m, $text) = $sc =~ /^(\d+)=(.*)\z/s or die "FATAL: bad --set-cell\n";
		$out = cell($lines[$n-1], $m, sub { " $text " });
	} else {
		$out = $lines[$n-1];
	}
	if ($ENV{STRIKE_CELL} ne "0") {
		$out = cell($out, $ENV{STRIKE_CELL}, sub {
			my ($lead, $body, $trail) = @_;
			return $body =~ /^\s*\z/ ? "${lead}~~ ~~${trail}" : "${lead}~~${body}~~${trail}";
		});
	} elsif (!length $ENV{STRIKE_SET}) {
		my $verdict = decode_base64($ENV{STRIKE_VERDICT});
		(my $line = $out) =~ s/\r?\n\z//;
		$out = "~~${line}~~ $verdict\n";
	}
	$lines[$n-1] = $out;
	open(my $fh, ">", $file) or die "FATAL: write: $!";
	binmode $fh; print {$fh} @lines; close $fh;

	sub cell {
		my ($line, $k, $fn) = @_;
		my @parts = split /(\|)/, $line, -1;
		my $idx = 2 * $k;
		die "FATAL: cell $k is beyond the row column count\n" if $idx > $#parts;
		$parts[$idx] =~ s/^(\s*)(.*?)(\s*)\z/$fn->($1, $2, $3)/se;
		return join "", @parts;
	}
' "$file"; then
	# die exits 255; the refusal contract is 1. The write happens only
	# after every check, so a refusal here left the file untouched.
	echo "safe-strike: refused — the file is untouched" >&2
	exit 1
fi

after="$(sed -n "${line_no}p" "$file")"
echo "   after: $after"

if [ "$no_verify" -eq 0 ] && [ "$gate" -eq 1 ]; then
	echo "safe-strike: gate per batch — running the annotate gate"
	if command -v docs-gates >/dev/null 2>&1; then
		docs-gates annotate .
	else
		go run ./cmd/docs-gates annotate .
	fi
fi

if [ "$no_verify" -eq 0 ] && [ "$battery" -eq 1 ]; then
	echo "safe-strike: battery per batch — running all five gates"
	if command -v docs-gates >/dev/null 2>&1; then
		docs-gates run .
	else
		go run ./cmd/docs-gates run .
	fi
fi

echo "safe-strike: struck — review the diff, run the battery over the repo, commit"
