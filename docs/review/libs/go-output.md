# go-output utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v0.38.7 (latest local tag, fetched 2026-11-23) + HEAD c0e177ee04a8835bb4dd04dbbd7581139757209d (`git describe` shows `bdd/v0.38.7-11-gc0e177e`; 666 tags total, v0.38.7 the newest v0.38.x). Modules: root `github.com/larsartmann/go-output` plus 16 submodules — bdd, d2, daghtml, delimited, escape, graph, markdown, markup, nom, plantuml, serialization, table, testhelpers, tree, tui (most pinned v0.38.7) and examples/integration (v0.23.2). Root is NOT stdlib-free: go-branded-id v0.7.0 + testhelpers v0.38.7 + golang.org/x/term v0.47.0 (go-output/go.mod:5-10).
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f baseline — root module stdlib-only (scripts/stack-drift-check.sh:9) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 3 (I/O + runtime + output) subagent of the multi-library utilization audit fleet.

## Verdict

NOT-ADOPTED — every rendering surface docs-gates owns (verdict matrix lines, one snake_case JSON document, truncation) is a byte-pinned contract the library would replace with a configurable engine, and the one overlap the fleet ever nominated (`Truncate` vs `nom.TruncateVisible`) is a recorded deliberate non-fix; no Go code in the consumer renders markdown or tables at all.

## Score

- Inputs N = 12 (inventory rows below).
- A = 3 (terminal-safe truncation, table serialization, golden-test helpers — real rendering concerns with an in-repo answer).
- M = 0.
- Score = **0** (round(100·0/3)) — band **low**, as expected for NOT-ADOPTED; the standing decisions carry the signal.

## What the library offers

| # | Capability | Applies? | Status | Evidence |
|---|---|---|---|---|
| 1 | `nom.TruncateVisible` — ANSI/width-aware truncation with `…` inside the width budget | Yes — the exact canonical candidate once triaged for `gatecli.Truncate` | Declined — standing decision | go-output/nom/terminal.go:51; internal/gatecli/gatecli_output.go:14; docs/agent-notes.md:64-68 |
| 2 | `markdown.MarkdownTable` — pipe tables with auto column widths, alignment, footers, escaping | No — no Go code in docs-gates composes a markdown table (verified below) | N/A — different problem (parsing, not rendering) | go-output/markdown/markdown.go:49-116; internal/docsrows/scan.go:82-127 |
| 3 | `serialization.NewJSONTableRenderer` + JSON/YAML/TOML/JSONL writers | Adjacent — the battery has a JSON face, but it is one fixed document schema, not a table-of-rows | N/A — typed struct + `encoding/json` already pins it | go-output/serialization/json.go:80; internal/battery/battery.go:295 |
| 4 | Root `Table`/`TreeNode`/`Graph` shapes + `RenderTable` dispatch + `Format` enum (16 formats) | No — the battery matrix is three fixed line shapes, not user-selected formats | N/A | go-output/tabledata.go:30; go-output/render_tabledata.go:45; go-output/format.go:48 |
| 5 | `table` submodule — lipgloss terminal tables | No — no terminal table anywhere in the consumer | N/A | go-output/table/go.mod:5 (lipgloss dep) |
| 6 | `tree` / `graph` / `d2` / `plantuml` / `daghtml` renderers | No — no tree/diagram output exists | N/A | go-output/tree/cqrs.go:39; go-output/graph/cqrs.go:83 |
| 7 | `markup` — XML/HTML/AsciiDoc tables | No — no markup output; the docs/status/*.md reports are authored by the status-report SKILL, not by this binary's Go code | N/A | go-output/markup/go.mod |
| 8 | `delimited` — streaming CSV/TSV writers | No — no delimited output verb | N/A | go-output/delimited/go.mod |
| 9 | `bdd` submodule — ginkgo/v2 + gomega spec suite over the renderers | No — consumer testing contract is stdlib-`testing`-only by AGENTS.md | N/A | go-output/bdd/go.mod:8-13; AGENTS.md:191-195 |
| 10 | `testhelpers` — `AssertContains`/`AssertEqual`/golden-flavored assertion helpers | Adjacent — parity goldens are the consumer's equivalent, already byte-exact and self-owned | N/A — weaker fit than the in-repo ritual | go-output/testhelpers/helpers.go:26-122; internal/battery/parity_test.go:188-189 |
| 11 | `tui` — bubbletea interactive views | No — no TUI (deliberate; CLI exit-code contract) | N/A | go-output/tui/go.mod:5 (bubbletea dep) |
| 12 | `nom` — NOM-style real-time progress visualization (dependency trees, ETA, inline terminal) | No — the battery's progress surface is one stderr heartbeat + a half-budget watchdog line | N/A | go-output/nom/go.mod:5-13; internal/battery/battery_watchdog.go |

## What docs-gates does today

Zero usage, zero presence in either module graph. Verified:

```console
$ rg -n "go-output|larsartmann/go-output" /home/lars/projects/docs-gates -g '!docs/review/**' --type go
# rc=1 (zero Go-file hits)
$ rg -n "go-output" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 (not in any module graph)
```

Textual mentions outside code are exactly: the standing non-fix in CHANGELOG.md:331 and docs/agent-notes.md:64-68, flake.lock (transitive under the private erraudit input — never consumed by docs-gates code), one archived status report, and this fleet's scaffolding.

What the consumer does instead, per surface:

- **Verdict matrix** — `renderMatrix` prints three fixed line shapes per gate (`PASS/FAIL/SKIP name (in Xs, rc=…)`), internal/battery/battery.go:317-339. It is not a table data structure; it is a pinned text protocol, byte-checked by `TestParityGolden` over five corpora (internal/battery/parity_test.go:188, re-pin ritual `PARITY_UPDATE=1` at :189 and documented at :28).
- **JSON machine face** — `renderJSON` marshals two typed structs with snake_case tags via stdlib `encoding/json` (internal/battery/battery.go:264-315, `json.MarshalIndent` at :295); the shape is pinned by the battery-json golden and parsed strictly by `pkg/provider` (`DisallowUnknownFields` is the row-shape contract, AGENTS.md:123-126). A generic table serializer would not produce this fixed schema.
- **Truncation** — `gatecli.Truncate`, a 14-line byte-budget UTF-8-boundary cut (internal/gatecli/gatecli_output.go:14-29), consumed by 10+ gate call sites (e.g. internal/docsannotate/docsannotate.go:83, internal/docsrouting/docsrouting.go:258, internal/docsfreshness/nearmiss.go:70). The fleet candidate `nom.TruncateVisible` differs semantically (visible-width budget, `…` suffix) and lives in a lipgloss-dependent submodule; the goldens pin the current bytes.
- **Markdown** — the consumer PARSES markdown tables (internal/docsrows/scan.go:82-127, internal/docsstrike/strike.go:9-10, internal/docsannotate/items.go:12) but never renders one: `rg -n "text/tabwriter|\\| %s \\|" internal/ cmd/ pkg/ --type go` → rc=1 for any renderer hit. The reports under docs/status/ are written by the status-report skill (agent workflow), not by this Go code — confirmed by the zero renderer hits above.
- **Output policy** — every human line routes through `gatecli.Outf/Outln/Errf/Errln/Livef` (internal/gatecli/gatecli_output.go:32, :78), the one home for the "a stream write failure is not a gate finding" policy; raw `fmt.Print*` is forbidden by forbidigo. go-output's `WriteRendered` (go-output/write_rendered.go:15) solves a different problem (error-wrapped writer tails for renderers), not this policy.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** The argument lands only if docs-gates hand-rolls something go-output solved. It does not: the battery's total rendering surface is ~30 lines of pinned verdict rows plus one JSON document — there is no format matrix, no column-width math, no escaping, no footer rendering, no multi-format dispatch. The one place the library's name came up (Truncate) was triaged to a deliberate non-fix on the record: "the canonical candidate (go-output/nom.TruncateVisible) breaks the stdlib-only contract and its semantics differ anyway — and the parity goldens pin Truncate's byte-budget cut byte-for-byte" (docs/agent-notes.md:64-68; CHANGELOG.md:331; AGENTS.md:123-124 lists it among deliberate non-fixes). This audit cites that decision rather than re-litigating it: swapping a byte-budget cut for a visible-width cut would re-pin fifteen goldens and change the provider-visible stderr for zero user-visible gain in a non-TTY CI context.

**Root-contract bar.** Barred outright: the root module tolerates zero requires, and even go-output's root (its lightest module) carries three, with the relevant truncation living in a submodule that transitively pulls charmbracelet/x/ansi, lipgloss, and a dozen more (go-output/nom/go.mod:5-33). The hermetic nix build and the byte-pinned goldens are the product; a configurable rendering engine is the opposite of a pinned protocol.

**pkg/provider check.** The provider's job is strict JSON decode of an already-rendered document (pkg/provider parses `run --json` with `DisallowUnknownFields`); it produces findings objects for BuildFlow, not human rendering. No format multiplexing exists there to delegate.

**Testing-style note.** go-output's bdd submodule demonstrates ginkgo/gomega specs, and testhelpers offers substring assertions — but the consumer's testing contract is stdlib-`testing`-only, gates exercised through their real `Run` functions with `gatecli.CaptureOutput`, never mocked (AGENTS.md:191-195). Its parity harness already IS a golden framework, purpose-built around `PARITY_UPDATE=1` re-pinning and a `{{DATE}}` token; testhelpers' looser assertions would be a downgrade, not a dedup.

RECONSIDER-WHEN: docs-gates grows a verb whose OUTPUT FORMAT is user-selected or whose output is a real data table — e.g. a `docs-gates report --format markdown|html|csv` verb that renders the verdict matrix into docs/status dashboards, or an age-bars/DF-margins report that benefits from column alignment. That verb should live in its own nested module (the pkg/provider pattern) with go-output/markdown or go-output/serialization behind it, while `run`'s byte-pinned faces stay stdlib. Until a second consumer asks for machine-rendered tables, the pinned plain lines are the cheaper contract.

## Ranked opportunities

None. (The Truncate overlap is adjudicated; the matrix/JSON faces are contracts, not rendering debt; no markdown is rendered from Go.)

## How to verify

```console
$ git -C /home/lars/projects/go-output log -1 --format='%H'
c0e177ee04a8835bb4dd04dbbd7581139757209d
$ git -C /home/lars/projects/go-output describe --tags
bdd/v0.38.7-11-gc0e177e
$ git -C /home/lars/projects/go-output tag --list 'v0.38*' | tail -1
v0.38.7

$ rg -n "go-output|larsartmann/go-output" /home/lars/projects/docs-gates -g '!docs/review/**' --type go
# rc=1 — zero Go-file hits
$ rg -n "go-output" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 — not in any module graph

$ rg -n "text/tabwriter|\| %s \|" /home/lars/projects/docs-gates/internal /home/lars/projects/docs-gates/cmd /home/lars/projects/docs-gates/pkg --type go | rg -v _test
# rc=1 — no table renderer anywhere in Go code (markdown is PARSED: docsrows/scan.go, docsstrike/strike.go, docsannotate/items.go)

$ rg -n "func Truncate|func Outf|func Livef" /home/lars/projects/docs-gates/internal/gatecli/gatecli_output.go
# 14:func Truncate …  32:func Outf …  78:func Livef
$ rg -n "func TestParityGolden|PARITY_UPDATE" /home/lars/projects/docs-gates/internal/battery/parity_test.go
# 28: re-pin ritual documented … 188:func TestParityGolden … 189: update := os.Getenv("PARITY_UPDATE") != ""

$ sed -n '64,68p' /home/lars/projects/docs-gates/docs/agent-notes.md
# — the standing non-fix: TruncateVisible breaks stdlib-only, semantics differ, goldens pin the cut
$ sed -n '325,334p' /home/lars/projects/docs-gates/CHANGELOG.md
# — the CHANGELOG twin of the same decision
```
