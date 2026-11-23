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
- Why the other nine rows are N/A, grouped: the eight format renderers (markdown, table, tree, graph, d2, plantuml, daghtml, markup, delimited, tui) target outputs the consumer does not produce — no format flag, no column math, no diagram, no TUI exists; `nom`'s live progress targets long-running builds, not a battery that finishes in seconds or fails a row on timeout; the `bdd` suite is a test-paradigm offer to a repo whose testing contract forbids it.

- Applicable rows are exactly the inventory rows graded **Missed Opportunity** (the consumer has the underlying need; the capability is hand-rolled, solved in-domain, or declined by contract). Every **Not Applicable** row carries a one-line reason. `M` counts rows graded Fully Leveraged.

## What the library offers

| #  | Capability                                                                                   | Applies?                                                                                                                  | Status                                                                                 | Evidence                                                                                       |
| -- | -------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| 1  | `nom.TruncateVisible` — ANSI/width-aware truncation with `…` inside the width budget         | Yes — the exact canonical candidate once triaged for `gatecli.Truncate`                                                   | Missed Opportunity (declined: standing decision)                                       | go-output/nom/terminal.go:51; internal/gatecli/gatecli_output.go:14; docs/agent-notes.md:64-68 |
| 2  | `markdown.MarkdownTable` — pipe tables with auto column widths, alignment, footers, escaping | No — no Go code in docs-gates composes a markdown table (verified below)                                                  | N/A — different problem (parsing, not rendering)                                       | go-output/markdown/markdown.go:49-116; internal/docsrows/scan.go:82-127                        |
| 3  | `serialization.NewJSONTableRenderer` + JSON/YAML/TOML/JSONL writers                          | Adjacent — the battery has a JSON face, but it is one fixed document schema, not a table-of-rows                          | Missed Opportunity (solved in-domain) — typed struct + `encoding/json` already pins it | go-output/serialization/json.go:80; internal/battery/battery.go:295                            |
| 4  | Root `Table`/`TreeNode`/`Graph` shapes + `RenderTable` dispatch + `Format` enum (16 formats) | No — the battery matrix is three fixed line shapes, not user-selected formats                                             | N/A                                                                                    | go-output/tabledata.go:30; go-output/render_tabledata.go:45; go-output/format.go:48            |
| 5  | `table` submodule — lipgloss terminal tables                                                 | No — no terminal table anywhere in the consumer                                                                           | N/A                                                                                    | go-output/table/go.mod:5 (lipgloss dep)                                                        |
| 6  | `tree` / `graph` / `d2` / `plantuml` / `daghtml` renderers                                   | No — no tree/diagram output exists                                                                                        | N/A                                                                                    | go-output/tree/cqrs.go:39; go-output/graph/cqrs.go:83                                          |
| 7  | `markup` — XML/HTML/AsciiDoc tables                                                          | No — no markup output; the docs/status/*.md reports are authored by the status-report SKILL, not by this binary's Go code | N/A                                                                                    | go-output/markup/go.mod                                                                        |
| 8  | `delimited` — streaming CSV/TSV writers                                                      | No — no delimited output verb                                                                                             | N/A                                                                                    | go-output/delimited/go.mod                                                                     |
| 9  | `bdd` submodule — ginkgo/v2 + gomega spec suite over the renderers                           | No — consumer testing contract is stdlib-`testing`-only by AGENTS.md                                                      | N/A                                                                                    | go-output/bdd/go.mod:8-13; AGENTS.md:196-200                                                   |
| 10 | `testhelpers` — `AssertContains`/`AssertEqual`/golden-flavored assertion helpers             | Adjacent — parity goldens are the consumer's equivalent, already byte-exact and self-owned                                | Missed Opportunity (solved in-domain) — weaker fit than the in-repo ritual             | go-output/testhelpers/helpers.go:26-122; internal/battery/parity_test.go:188-189               |
| 11 | `tui` — bubbletea interactive views                                                          | No — no TUI (deliberate; CLI exit-code contract)                                                                          | N/A                                                                                    | go-output/tui/go.mod:5 (bubbletea dep)                                                         |
| 12 | `nom` — NOM-style real-time progress visualization (dependency trees, ETA, inline terminal)  | No — the battery's progress surface is one stderr heartbeat + a half-budget watchdog line                                 | N/A                                                                                    | go-output/nom/go.mod:5-13; internal/battery/battery_watchdog.go                                |

### Reading the three "Adjacent/Yes" rows fairly

- Row 1 (truncation) is the only overlap ever triaged, and it lost on semantics BEFORE the contract argument: `TruncateVisible` budgets terminal cell width and appends `…` (go-output/nom/terminal.go:48-52); `gatecli.Truncate` budgets BYTES and never appends — finding snippets must stay byte-honest for a corpus whose lines are compared mechanically (docs/agent-notes.md:64-68).
- Row 3 (JSON serialization) is adjacent only in noun: `NewJSONTableRenderer` renders a `Table` as an array of row objects (go-output/serialization/json.go:80); the battery renders one FIXED document (root/pass/gates with embedded streams) whose snake_case tags are a lint-pinned contract consumed by pkg/provider — a schema, not a serialization choice.
- Row 10 (test helpers) offers assertions; the consumer's parity harness offers PROOFS (byte equality over fifteen invocations). Substring asserts cannot pin a protocol; goldens can. The consumer already owns the stronger tool.

### Why the library exists vs why the consumer does not need it

go-output's pitch is "write your data once, render it anywhere": one `Table`/`Tree`/`Graph` value, sixteen format renderers, runtime format selection via the `Format` enum (go-output/format.go:48), unified dispatch through `RenderTable` (go-output/render_tabledata.go:45). That value proposition begins at the SECOND output format. docs-gates' rendering obligations run in the opposite direction: the human face and the JSON face are not two renders of one dataset — they are two different contracts (progress + verdict lines vs one machine document), each byte-pinned independently by the parity goldens (internal/battery/parity_test.go:188), with `--json` deliberately reserving stdout for the document alone (internal/battery/battery.go:97, the render at :260). A format matrix would be a liability: every format would be an unpinned surface the parity ritual has to grow to cover, or an unpinned surface it silently does not.

### The dependency shape, submodule by submodule

The workspace is honest about weight — formats are modules so importers pay only for what they use — but the formats a table-centric consumer would want are exactly the heavy ones: `table` requires lipgloss v2 (go-output/table/go.mod:5), `tui` requires bubbletea + lipgloss + charmbracelet/x (go-output/tui/go.mod:5-10), `nom` (the TruncateVisible home) rides the full charmbracelet ANSI stack (go-output/nom/go.mod:5-13), and `bdd` requires ginkgo/v2 + gomega plus a 20-line indirect tree (go-output/bdd/go.mod:8-13). Even the root module — the zero-heavy-deps selling point — carries three non-stdlib requires (go-output/go.mod:5-10). For a repo whose CI fails on ANY root require (scripts/stack-drift-check.sh:9), there is no light corner of this library to stand on.

### Testing philosophy: bdd/testhelpers vs the consumer's parity ritual

go-output validates itself two ways the consumer was never going to copy: a ginkgo/gomega BDD suite as its own module (go-output/bdd/bdd_suite_test.go:12 runs `ginkgo.RunSpecs`; gomega assertions throughout the sibling modules' tests) and `testhelpers` — substring/equality assertion helpers like `AssertContains`/`AssertAllContained`/`AssertLineCount` (go-output/testhelpers/helpers.go:26-122) plus error/writer mocks (go-output/testhelpers/writers.go:14-26). docs-gates' testing contract is stdlib-`testing`-only, gates exercised through their real `Run` functions with `gatecli.CaptureOutput`, never mocked output (AGENTS.md:196-200), with corpus fixtures keeping ages stable via a `{{DATE}}` token. Note the asymmetry: substring asserts are WEAKER than what the consumer already does — byte-exact golden comparison over fifteen invocations (internal/battery/parity_test.go:188). Adopting the library's test style would trade a stronger guarantee for a dependency and a paradigm shift; there is nothing here the fleet argument can claim as shared debt.

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
- **JSON machine face** — `renderJSON` marshals two typed structs with snake_case tags via stdlib `encoding/json` (internal/battery/battery.go:264-315, `json.MarshalIndent` at :295); the shape is pinned by the battery-json golden and parsed strictly by `pkg/provider` (`DisallowUnknownFields` is the row-shape contract, AGENTS.md:123-127). A generic table serializer would not produce this fixed schema.
- **Version/help text** — `--version` (ldflags-stamped) and the per-verb usage surfaces are likewise pinned by usage tests — more byte-protocol surfaces a rendering engine cannot own.
- **Truncation** — `gatecli.Truncate`, a 14-line byte-budget UTF-8-boundary cut (internal/gatecli/gatecli_output.go:14-29), consumed by 10+ gate call sites (e.g. internal/docsannotate/docsannotate.go:83, internal/docsrouting/docsrouting.go:258, internal/docsfreshness/nearmiss.go:70). The fleet candidate `nom.TruncateVisible` differs semantically (visible-width budget, `…` suffix) and lives in a lipgloss-dependent submodule; the goldens pin the current bytes. (There is also one private fixed-width `truncate` in internal/docsrows/rows.go:220 for row previews — intra-repo variety, not fleet duplication, and also golden-covered.)
- **Markdown** — the consumer PARSES markdown tables (internal/docsrows/scan.go:82-127, internal/docsstrike/strike.go:9-10, internal/docsannotate/items.go:12) but never renders one: `rg -n "text/tabwriter|\\| %s \\|" internal/ cmd/ pkg/ --type go` → rc=1 for any renderer hit. The reports under docs/status/ are written by the status-report skill (agent workflow), not by this Go code — confirmed by the zero renderer hits above.

  This parse-only posture deserves one more sentence, because it is where a hurried auditor would see duplication that is not there: go-output/markdown RENDERS tables from cell data (go-output/markdown/markdown.go:116-150 — escape cells, measure widths, write header/separator/rows/footer); the consumer's docsrows does the INVERSE (internal/docsrows/scan.go walks fenced blocks, finds separator rows, extracts cells) to JUDGE existing tables. A renderer and a judge share vocabulary (`|`, `---`) and nothing else; docs-gates could not reuse the renderer without first inventing a renderer-shaped problem.
- **Output policy** — every human line routes through `gatecli.Outf/Outln/Errf/Errln/Livef` (internal/gatecli/gatecli_output.go:32, :78), the one home for the "a stream write failure is not a gate finding" policy; raw `fmt.Print*` is forbidden by forbidigo. go-output's `WriteRendered` (go-output/write_rendered.go:15) solves a different problem (error-wrapped writer tails for renderers), not this policy.

### The two golden layers, and why a library cannot own them

The rendering correctness story is not "code that formats" — it is two pinned invariants: (1) `TestParityGolden` byte-compares fifteen invocations across five corpora (internal/battery/parity_test.go:188, re-pin only via `PARITY_UPDATE=1` at :189), and (2) the defaults-are-historical invariant — zero-value Settings fall back to the pre-extraction constants so a config-less repo is byte-for-byte the old battery (AGENTS.md:40-42). Both bind the OUTPUT bytes, not the rendering mechanism. Swapping the mechanism for go-output would obligate a full golden re-pin with no behavioral gain — the textbook "golden bump is a contract change" the repo warns about (AGENTS.md:51-54). The goldens are the reason there is no rendering debt to share with the fleet: the shared thing IS the pinned bytes, and those already travel (the provider consumes them cross-repo).

### The consumer's full output inventory (nothing hides here)

Enumerating every output shape the binary produces — this is the completeness argument that no go-output capability is being hand-rolled under a different name:

- Verdict/heartbeat lines: `PASS/FAIL/SKIP` matrix rows, per-gate `gate X ...` progress, timeout/FATAL notices — all `gatecli.Outf`-family one-liners (internal/battery/battery.go:317-339; internal/gatecli/gatecli_output.go:32).
- Finding lines: each gate names its findings as prefixed lines with file:line anchors and truncated snippets — e.g. `UNROUTED [report:line] text` (internal/docsrouting/docsrouting.go:258), `table at line N ...` (internal/docsrows/rows.go:157), rule-F spacing findings (internal/docsannotate/spacing.go:64). No columns, no alignment, no footers — a line protocol, golden-pinned.
- One JSON document: the `--json` face (internal/battery/battery.go:264-315).
- Files: the four guarded living-doc writes covered by the go-atomic-write report; the reports themselves (docs/status/*.md) are authored by the status-report SKILL, outside this Go code.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** The argument lands only if docs-gates hand-rolls something go-output solved. It does not: the battery's total rendering surface is ~30 lines of pinned verdict rows plus one JSON document — there is no format matrix, no column-width math, no escaping, no footer rendering, no multi-format dispatch. The one place the library's name came up (Truncate) was triaged to a deliberate non-fix on the record: "the canonical candidate (go-output/nom.TruncateVisible) breaks the stdlib-only contract and its semantics differ anyway — and the parity goldens pin Truncate's byte-budget cut byte-for-byte" (docs/agent-notes.md:64-68; CHANGELOG.md:331; AGENTS.md:123-124 lists it among deliberate non-fixes). This audit cites that decision rather than re-litigating it: swapping a byte-budget cut for a visible-width cut would re-pin fifteen goldens and change the provider-visible stderr for zero user-visible gain in a non-TTY CI context.

**Root-contract bar.** Barred outright: the root module tolerates zero requires, and even go-output's root (its lightest module) carries three, with the relevant truncation living in a submodule that transitively pulls charmbracelet/x/ansi, lipgloss, and a dozen more (go-output/nom/go.mod:5-33). The hermetic nix build and the byte-pinned goldens are the product; a configurable rendering engine is the opposite of a pinned protocol.

**pkg/provider check.** The provider's job is strict JSON decode of an already-rendered document (pkg/provider parses `run --json` with `DisallowUnknownFields`); it produces findings objects for BuildFlow, not human rendering. No format multiplexing exists there to delegate.

**Testing-style note.** go-output's bdd submodule demonstrates ginkgo/gomega specs, and testhelpers offers substring assertions — but the consumer's testing contract is stdlib-`testing`-only, gates exercised through their real `Run` functions with `gatecli.CaptureOutput`, never mocked (AGENTS.md:196-200). Its parity harness already IS a golden framework, purpose-built around `PARITY_UPDATE=1` re-pinning and a `{{DATE}}` token; testhelpers' looser assertions would be a downgrade, not a dedup.

**What root adoption would actually cost, concretely.** (1) Break zero-requires — CI red on the next stack-drift run (scripts/stack-drift-check.sh:9). (2) A STACK.md row and a separate-module carve-out even for the lightest import, or a golden re-pin train for any face that touched library-rendered bytes. (3) The `Format`-enum mental model imported into a repo whose two faces are deliberately NOT interchangeable renders of one dataset. Against that: the ability to call `TruncateVisible` — a function whose semantics the repo explicitly does not want (docs/agent-notes.md:64-68). The ledger does not balance in any row.

RECONSIDER-WHEN: docs-gates grows a verb whose OUTPUT FORMAT is user-selected or whose output is a real data table — e.g. a `docs-gates report --format markdown|html|csv` verb that renders the verdict matrix into docs/status dashboards, or an age-bars/DF-margins report that benefits from column alignment. That verb should live in its own nested module (the pkg/provider pattern) with go-output/markdown or go-output/serialization behind it, while `run`'s byte-pinned faces stay stdlib. Until a second consumer asks for machine-rendered tables, the pinned plain lines are the cheaper contract.

**Fleet-argument scorecard, per prong.** (a) "Re-implementation debt": none found — the matrix, the JSON face, and Truncate are pinned contracts, not format engines; no markdown/table composition exists in Go (the skills write the reports). The one historical candidate (Truncate) was triaged ON THE RECORD as a deliberate non-fix, which is the fleet argument working correctly: the decision is written down, costed (golden re-pin), and revisitable, not silently re-derived per project. (b) "Collective improvement": if go-output's markdown/serialization renderers improve, docs-gates inherits nothing under NOT-ADOPTED — acceptable because docs-gates ships zero rendered artifacts that a renderer could improve; its artifacts are verdict lines and a JSON schema, both frozen on purpose.

## Ranked opportunities

None. (The Truncate overlap is adjudicated; the matrix/JSON faces are contracts, not rendering debt; no markdown is rendered from Go.)

For the record, the three things that WOULD have been opportunities if the architecture were different:

1. A `--format` flag on `run` — the day more than one machine consumer exists.
2. Rendered dashboard artifacts — the day a human-facing web view of gate verdicts is wanted.
3. Terminal tables — the day the matrix grows columns worth aligning.

Each is a product decision, not a rendering gap; none is on the roadmap.

## How to verify

```console
$ git -C /home/lars/projects/go-output rev-parse c0e177ee04a8835bb4dd04dbbd7581139757209d^{commit}
c0e177ee04a8835bb4dd04dbbd7581139757209d
# — the reviewed commit, pinned: the repository tip drifts under the fleet's auto-commit daemon
$ git -C /home/lars/projects/go-output describe --tags c0e177ee04a8835bb4dd04dbbd7581139757209d
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
14:func Truncate(s string, n int) string {
32:func Outf(format string, args ...any) {
78:func Livef(format string, args ...any) {
$ rg -n "func TestParityGolden|PARITY_UPDATE" /home/lars/projects/docs-gates/internal/battery/parity_test.go
28://	PARITY_UPDATE=1 go test ./internal/battery/ -run TestParityGolden
188:func TestParityGolden(t *testing.T) {
189:	update := os.Getenv("PARITY_UPDATE") != ""

$ sed -n '64,68p' /home/lars/projects/docs-gates/docs/agent-notes.md
- **`gatecli.Truncate` stays a hand-rolled 10-liner** (deliberate
  non-fix): the canonical candidate (go-output/nom.TruncateVisible)
  breaks the stdlib-only contract and its semantics differ anyway — and
  the parity goldens pin Truncate's byte-budget cut byte-for-byte.
- **`nix fmt` needs the tracked treefmt.toml** — without it the formatter
# — the standing non-fix: TruncateVisible breaks stdlib-only, semantics differ, goldens pin the cut
$ git -C /home/lars/projects/go-output tag --list | wc -l
666
$ rg -n "func truncate" /home/lars/projects/docs-gates/internal --type go
/home/lars/projects/docs-gates/internal/docsrows/rows.go:220:func truncate(s string) string {
# (a private fixed-width preview variant) — the consumer's complete truncation surface
$ sed -n '325,334p' /home/lars/projects/docs-gates/CHANGELOG.md
  test-file similarity, the intentional class — and a `--structural`
  spot-check found no hidden actionable clones (the one non-test group is
  const-declaration shape). The audit's `-t 1` self-scan also caught the
  fresh docsannotate settings projection mirroring docsfreshness's: both
  now carry in-source accepts with rationale (baseline re-proven at 0
  shown groups). The `gatecli.Truncate` canonical-helper triage landed as
  a deliberate non-fix — the width-budgeted go-output helper would break
  the stdlib-only contract and the byte-precise golden shapes.
- Shared-helper round 2 (zero behavior change, goldens byte-identical):
  docsfreshness.Corpus is now the ONE prepared living-docs corpus (raw +
# — the CHANGELOG twin of the same decision
```
