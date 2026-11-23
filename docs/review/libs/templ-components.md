# templ-components utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v1.27.2 (latest local tag, fetched 2026-11-23) + HEAD d87221760214cd7bb6ced4c55966f7a651bb3296 per the fleet manifest (the checkout's tip carries two later docs/gitignore-only commits, cf725c9d; no component code moved). Modules: root `github.com/larsartmann/templ-components` plus `charts/echarts`, `datastar`, `errorpage`, `htmx`, `icons`, `utils` (all v1.27.2), `visualtest` + `website` (untagged).
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f baseline — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 5 (UI + CQRS/application frameworks) subagent of the multi-library utilization audit fleet.

## Verdict

NOT-APPLICABLE — templ-components is a server-rendered HTML component library for templ + HTMX + Tailwind v4 (138 components, 102 SVG icons, typed props), and docs-gates renders no HTML at all: its two output surfaces are byte-pinned terminal prose and one snake_case JSON document, so there is no UI for the library to render.

## Score

- Inputs N = 13 (inventory rows below).
- A = 1 (the single concept row: "one shared, versioned presentation layer" — row 13).
- M = 0.
- Score = round(100 × 0/1) = **0** — band **low**. Expected for NOT-APPLICABLE; the number measures correctly-unrealized leverage, the verdict carries the meaning.

- Applicable rows are exactly the inventory rows graded **Missed Opportunity** (the consumer has the underlying need; the capability is hand-rolled, solved in-domain, or declined by contract). Every **Not Applicable** row carries a one-line reason. `M` counts rows graded Fully Leveraged.

## What the library offers

| #  | Capability                                                                                                       | Applies to docs-gates?                                                                                                                | Status                                                                                                        | Evidence                                                               |
| -- | ---------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| 1  | `display` — 53 data-display components (cards, tables, modals, tabs, kanban, native SVG charts)                  | No — no HTML output anywhere                                                                                                          | N/A                                                                                                           | templ-components/README.md:151                                         |
| 2  | `forms` — 26 typed form controls (enums that fail the build on invalid state)                                    | No — no forms, no input round-trips                                                                                                   | N/A                                                                                                           | templ-components/README.md:222                                         |
| 3  | `layout` — 11 page-layout components (Base, ThemeScript, CSP-nonce threading)                                    | No — no pages                                                                                                                         | N/A                                                                                                           | templ-components/README.md:246                                         |
| 4  | `navigation` — breadcrumbs (with hardened JSON-LD), nav structures                                               | No — no navigation surface                                                                                                            | N/A                                                                                                           | templ-components/README.md:242                                         |
| 5  | `icons` module — 102 SVG icons as Go values                                                                      | No — glyphs never rendered                                                                                                            | N/A                                                                                                           | templ-components/README.md:87                                          |
| 6  | `charts/echarts` module — ECharts chart components                                                               | No — the verdict matrix is text tables; no chart exists to draw                                                                       | N/A                                                                                                           | templ-components/charts/echarts/go.mod:1                               |
| 7  | `htmx` module — HTMX response/transport helpers                                                                  | No — no HTTP, no browser                                                                                                              | N/A                                                                                                           | templ-components/htmx/go.mod:1                                         |
| 8  | `datastar` module — opt-in Datastar runtime integration                                                          | No — same                                                                                                                             | N/A                                                                                                           | templ-components/datastar/go.mod:1                                     |
| 9  | `errorpage` module — pretty classified error pages (HTML) + JSON contracts                                       | No — errors are the 0/1/2 exit codes plus stderr prose                                                                                | N/A                                                                                                           | templ-components/errorpage/go.mod:1                                    |
| 10 | `utils` module — `BaseProps`/`ComponentProps` composition, `Class` tailwind-merge, `ScriptComponent` CSP hygiene | No — no components, no CSS, no scripts                                                                                                | N/A                                                                                                           | templ-components/utils/utils.go:14, templ-components/utils/utils.go:48 |
| 11 | `visualtest` module — chromedp pixel-level visual goldens (206)                                                  | No — docs-gates parity is byte goldens over text (internal/battery/parity_test.go), no browser in the loop                            | N/A                                                                                                           | templ-components/README.md:115                                         |
| 12 | `website` + `tc` CLI (docs site, `tc doctor` repo scanner)                                                       | No — docs-gates ships markdown docs; its CLI scans markdown corpora, not templ call sites                                             | N/A                                                                                                           | templ-components/README.md:84                                          |
| 13 | Concept: ONE shared, semver-versioned presentation layer so every fleet UI improves collectively                 | Concept only — docs-gates' presentation contract is the opposite shape: byte-pinned plaintext, re-pinned deliberately per golden bump | Missed Opportunity (declined: opposite shape — byte-pinned plaintext, re-pinned deliberately per golden bump) | internal/battery/parity_test.go; AGENTS.md:51-54                       |

Dependency note for any future adoption: the root module requires
`github.com/a-h/templ v0.3.1100` plus the six sibling modules
(templ-components/go.mod:5-14) — importing any component package into the
docs-gates root breaks the zero-require contract on the first `go get`
(scripts/stack-drift-check.sh:8-13 fails CI).

## What docs-gates does today

Zero usage, zero module-graph presence. Verified:

```console
$ rg -n "larsartmann/templ-components|a-h/templ" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**'
# rc=1, no matches — not in go.mod (root has no requires at all), not in
# pkg/provider/go.mod:5-10 (go-finding, toolsdk, go-error-family only), not in
# any source file
$ rg -n "html/template|text/template" /home/lars/projects/docs-gates --type go
# rc=1 — zero files; no templating engine of any kind is imported
```

The only templ-components strings in the tree are non-code: an incident-note
comment in the annotate gate about a fan-out symlink whose content lived in the
templ-components checkout (internal/docsannotate/spacing.go:113-114 — the
2026-11-14 crush-config incident), and the audit-fleet scaffolding list in
scripts/citation-sweep.py:47. Every other "templ" match in Go sources is the
phrase "bare template" — the stub gate's detection of an unfilled
domain-language doc by pinned checksum (internal/docsstub/docsstub.go:9, :37),
which is text hashing, not templating. No `.templ` file exists in the repo.

Checked for hand-rolled UI utilities (the fleet-argument duplication question)
and found none to name:

- No component/props/render abstraction — gates are `Run([]string) int`
  packages dispatched from a plain map (cmd/docs-gates/verbs.go:77).
- No CSS/class merging, no icon set, no chart code — the battery's matrix is
  fixed-width text via `gatecli.Outf`/`Errf` (internal/gatecli/gatecli_output.go).
- The machine face is one JSON document emitted by the battery
  (internal/battery/battery.go:128-130), decoded strictly by the provider
  (pkg/provider/provider.go:92-102) — a data contract, not a view layer.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** templ-components exists to stop every fleet
web app from hand-rolling buttons, modals, CSP-safe scripts, and dark mode — a
real shared cost across the owner's GOTH-stack services. docs-gates is not in
that set and cannot be without stopping being what it is: an exit-code markdown
battery whose human output is deliberately terminal prose (AGENTS.md:19 makes
`gatecli` "the ONE home" for output) and whose machine output is deliberately
one jq-able JSON document. There is no re-implementation to retire and no
file:line of duplicated UI to name — no duplication exists because no overlap
exists. The nearest rhyme — the stub gate's "bare template" checksums
(internal/docsstub/docsstub.go:102-108) — is content fingerprinting, not
templating; templ-components solves rendered-output problems docs-gates never
has.

**Root-contract bar.** Any import breaks the stdlib-only root module (the templ
runtime requirement at templ-components/go.mod:6) and drags a UI toolchain's
generated code into a binary whose value is a ~1.5 ms judge (README.md:69-72).
Dead weight on every hermetic nix build.

**pkg/provider check.** Also no target: the provider's contract is one local
exec plus a strict JSON decode (pkg/provider/provider.go:148-171); it renders
nothing and would only gain binary size.

**What docs-gates would have to become.** A browser-facing surface — e.g. a
fleet dashboard visualizing battery verdicts across repos over time. That is a
different deliverable from the battery itself, and it is exactly the shape
templ-components solves fleet-wide; if it ever exists, it should be its own
module (the pkg/provider precedent for sanctioned dep homes) built on
templ-components from day one rather than a hand-rolled HTML layer. Note the
GOTH-stack pairing: the library is designed to ride with go-cqrs-lite +
cqrs-htmx (see those reports), so a docs-gates HTML face would naturally
arrive as part of that stack decision, not as a lone component import — one
flip trigger, three libraries, judged together in this group's reports.

RECONSIDER-WHEN: docs-gates (or a sibling module in this repo) grows an HTML
face — a battery-verdict dashboard, a report-corpus browser, or any rendered
documentation site for the tool. Until then the insulation is correct, not
neglect: the presentation layer docs-gates standardizes for the fleet is the
JSON row shape plus the exit codes, which is already shared outward through
pkg/provider.

## Ranked opportunities

None. (The one adjacent observation — visualtest's pixel goldens versus
docs-gates' byte goldens — is a testing-doctrine rhyme, not an adoption: the
consumer's parity suite already pins its output contract byte-for-byte, and a
chromedp dependency would be absurd inside the root contract.)

## How to verify

```console
$ git -C /home/lars/projects/templ-components rev-parse cf725c9decb30c8663f1cc429b84adef4af7fe3d^{commit}
cf725c9decb30c8663f1cc429b84adef4af7fe3d
# — the reviewed commit, pinned: the repository tip drifts under the fleet's auto-commit daemon
$ git -C /home/lars/projects/templ-components rev-parse d87221760214cd7bb6ced4c55966f7a651bb3296
d87221760214cd7bb6ced4c55966f7a651bb3296
$ git -C /home/lars/projects/templ-components tag --list 'v1.27.*'
v1.27.0
v1.27.1
v1.27.2

$ rg -n "larsartmann/templ-components|a-h/templ" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**'
# rc=1 — zero hits in code and module graphs
$ rg -n "html/template|text/template" /home/lars/projects/docs-gates --type go
# rc=1 — no templating imports in root or pkg/provider
$ rg -n "templ-components" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**'
/home/lars/projects/docs-gates/scripts/citation-sweep.py:47:    "templ-components",
/home/lars/projects/docs-gates/internal/docsannotate/spacing.go:113:// incident: a fixed templ-components SKILL.md was overwritten by the
$ cat /home/lars/projects/docs-gates/pkg/provider/go.mod
module github.com/Artmann-Technologies/docs-gates/pkg/provider
go 1.27
require (
	github.com/larsartmann/go-finding v1.15.1
	github.com/larsartmann/go-finding/toolsdk v1.16.0
)
require github.com/larsartmann/go-error-family v0.12.0 // indirect
# (go-error-family v0.12.0 indirect) — no templ-components in the graph

$ rg -n "bare template|Checksums" /home/lars/projects/docs-gates/internal/docsstub/docsstub.go | head -4
37:	Checksums []string // bare-template digests (md5 or sha256 hex); a match is a stub
47:		Checksums: config.OrSlice(cfg.StubChecksums, config.Defaults().StubChecksums),
55:const Usage = `docs-stub — fails a repo whose domain-language doc is the bare template
102:				"FAIL %s is the bare template (%s) — write the glossary or an honest stub\n",
```
