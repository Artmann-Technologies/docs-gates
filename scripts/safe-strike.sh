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
#
# Usage:
#   safe-strike.sh [--apply] [--gate] [--allow-dirty]
#                  [--cell N] [--set-cell M=TEXT] [--no-verify]
#                  <file> <anchor-regex> [verdict]
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
	echo "usage: safe-strike.sh [--apply] [--gate] [--allow-dirty] [--cell N] \\" >&2
	echo "                      [--set-cell M=TEXT] [--no-verify] <file> <anchor-regex> [verdict]" >&2
	exit 2
}

apply=0
gate=0
allow_dirty=0
cell=0
set_cell=""
no_verify=0

while [ $# -gt 0 ]; do
	case "$1" in
	--apply) apply=1 ;;
	--gate) gate=1 ;;
	--allow-dirty) allow_dirty=1 ;;
	--no-verify) no_verify=1 ;;
	--cell)
		[ $# -ge 2 ] || usage
		cell="$2"
		shift
		;;
	--set-cell)
		[ $# -ge 2 ] || usage
		set_cell="$2"
		shift
		;;
	-h | --help)
		sed -n '2,40p' "$0" | sed 's/^# \{0,1\}//'
		exit 0
		;;
	-*)
		echo "FATAL: unknown flag: $1" >&2
		usage
		;;
	*)
		break
		;;
	esac
	shift
done

[ $# -eq 2 ] || [ $# -eq 3 ] || usage
file="$1"
anchor="$2"
verdict="${3:-}"

if [ "$cell" = "0" ] && [ -z "$set_cell" ] && [ -z "$verdict" ]; then
	echo "FATAL: a whole-line strike needs a verdict (e.g. \"done at \`abc1234\`\");" >&2
	echo "cell-only operations may omit it" >&2
	usage
fi

if [ "$cell" != "0" ] && [ -n "$set_cell" ] && [ "$cell" = "${set_cell%%=*}" ]; then
	echo "FATAL: --cell and --set-cell name the same column" >&2
	usage
fi

if [ "$cell" != "0" ] && [ -n "$verdict" ]; then
	echo "FATAL: --cell mode takes no verdict — a verdict appended after a table" >&2
	echo "row would become a phantom extra column; use --set-cell M=TEXT to flip a" >&2
	echo "status cell instead" >&2
	usage
fi

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
export STRIKE_VERDICT="$(printf '%s' "$verdict" | base64 -w0)"
export STRIKE_FILE_WANT="$(printf '%s' "$line_content" | base64 -w0)"

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

echo "safe-strike: struck — review the diff, run the battery over the repo, commit"
