# httputil/endpoints utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: endpoints/v0.1.0 (latest — and only — local tag, fetched 2026-11-23; first release 2026-11-06 per the httputil CHANGELOG); the submodule sits in the httputil checkout at HEAD 34dbf05f670d28830e61170a087f552070279b74, module `github.com/larsartmann/httputil/endpoints`.
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 2 (HTTP/serialization) subagent of the multi-library utilization audit fleet.
- Fleet stance addressed: does stdlib-only insulation here avoid re-implementation debt, or leave docs-gates hand-rolling something the fleet solved?

## Verdict

NOT-APPLICABLE — endpoints is a route registry that mounts handlers on `http.ServeMux` and serves a content-negotiated `GET /endpoints` self-listing, and docs-gates has no routes, no mux, and no server; its self-describing surface is the argv verb table plus `help`/did-you-mean, an exit-2 CLI contract with nothing HTTP-negotiated about it.

## Score

- Inputs N = 7 (inventory rows below).
- A = 1 (the self-describing-surface concept row — docs-gates does publish a machine-derivable listing of its own command surface; solved natively by the verb tables and pinned help).
- M = 0.
- Score = round(100 × 0/1) = **0** — band **low**. Expected for NOT-APPLICABLE; the verdict carries the meaning.

- Applicable rows are exactly the inventory rows graded **Missed Opportunity** (the consumer has the underlying need; the capability is hand-rolled, solved in-domain, or declined by contract). Every **Not Applicable** row carries a one-line reason. `M` counts rows graded Fully Leveraged.

## What the library offers

| # | Capability                                                                                                                              | Applies to docs-gates?                                                                                                 | Status                                                                                                 | Evidence                                                                       |
| - | --------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------ |
| 1 | `Registry` — one declaration both mounts a route on `*http.ServeMux` and records it, so the listing cannot drift from what is served    | No — no `http.ServeMux`, no routes to mount                                                                            | N/A                                                                                                    | httputil/endpoints/endpoints.go:55, :93; httputil/endpoints/README.md:7-9      |
| 2 | `GET /endpoints` listing handler, mounted via `Register` (lists itself)                                                                 | No — no GET, no endpoint                                                                                               | N/A                                                                                                    | httputil/endpoints/endpoints.go:16, :191                                       |
| 3 | RFC 9110 proactive content negotiation between JSON and an HTML table on `Accept` (multi-line Accept combined; never 406s; ties → JSON) | No — the battery's dual face is chosen by `--json` flag, not negotiated: one JSON document on stdout, humans on stderr | N/A                                                                                                    | httputil/endpoints/README.md:74-79; consumer internal/battery/battery.go:97-98 |
| 4 | CSP-clean HTML table rendering with escaped fields, plus `Vary: Accept` / `Cache-Control: no-store`                                     | No — no HTML output, no cacheable surface                                                                              | N/A                                                                                                    | httputil/endpoints/README.md:88                                                |
| 5 | UTF-8 validation and wiring errors as values at registration (nil mux, duplicate/malformed pattern)                                     | No — verb dispatch rejects unknown verbs with exit 2 + did-you-mean, argv-shaped, not pattern-shaped                   | N/A                                                                                                    | httputil/endpoints/endpoints.go:62; consumer cmd/docs-gates/verbs.go:77        |
| 6 | Concurrency-safe registry with immutable atomic snapshots for reads                                                                     | No — the verb table is fixed at compile time; nothing registers at runtime                                             | N/A                                                                                                    | httputil/endpoints/README.md:83                                                |
| 7 | Self-describing surface listing (the concept under the route registry)                                                                  | Yes in concept — docs-gates' help/usage surface                                                                        | Missed Opportunity (solved natively) — verb tables + per-verb usage pinned by usage_test.go; see below | consumer cmd/docs-gates/verbs.go:77, :107, :129                                |

Module shape worth noting: like server_timing, this submodule is stdlib-only
with zero dependencies (httputil/endpoints/go.mod:1-3, httputil/endpoints/README.md:11),
so the hermeticity objection reduces to "any require breaks the contract",
not "a dependency tree would land".

### Release shape (context for any future adoption)

`endpoints/v0.1.0` is the module's first and only tag (2026-11-06), cut under
an explicit owner ruling that it sits in the pre-1.0 iteration tier — 0.x
semantics govern "while real consumers validate it" (httputil/CHANGELOG.md:61)
— unlike `server_timing`'s day-one-v1 posture. The first real consumer is CV,
which routes its entire surface through `Registry.Handle` so the listing
derives from the actual registration site (httputil/endpoints/README.md:97);
the `[Unreleased]` section above v0.1.0 already hardens negotiation (all
`Accept` header lines combined per RFC 9110 §5.3) and CI now runs the
submodule tests on every push. An adopter today gets a small, stdlib-only,
single-purpose module — exactly the right shape to slot into a future
docs-gates HTTP sibling, and exactly irrelevant to the current CLI.


## What docs-gates does today

Zero usage, zero presence. Verified:

```console
$ rg -n "httputil/endpoints|endpoints\." /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**' --type go
# rc=1, no matches
$ rg -n "github.com/larsartmann/httputil/endpoints" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 — not in any module graph
```

The consumer's actual command-surface machinery, checked for the
route-registry-shaped duplication the brief asks about:

- **Dispatch**: `cmd/docs-gates` keeps one verb-to-handler map
  (`verbHandlers`, cmd/docs-gates/verbs.go:77), an ordered verb list for help
  (:97), an overview map (:107), and per-verb usage text (:129). Unknown verbs
  exit 2 with a did-you-mean suggestion (README.md:39-41 documents it). This is
  argv dispatch — string keys to `func([]string) int` — with no mux, no
  patterns, no method routing, and no runtime registration.
- **Self-description**: `docs-gates help`, `<verb> --help`, and `--version`
  are intercepted before any gate parser (AGENTS.md:27-30), and usage is
  pinned per gate by usage_test.go — the anti-drift mechanism is a test
  golden, whereas endpoints' anti-drift mechanism is single-declaration
  registration. Same goal (surface cannot drift from reality), entirely
  different substrate (compile-time map + golden vs runtime registry +
  handler).
- **The machine face**: `run --json` emits one document on stdout; the
  provider consumes it (pkg/provider/provider.go:92-102). There is no
  content negotiation anywhere — the flag, not the `Accept` header, selects
  the representation, and that is byte-pinned by the battery-json golden.

No handler scaffolding, route table, or negotiation code exists to name as
duplicated — the honest duplication answer is none.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** endpoints exists because `*http.ServeMux`
cannot enumerate its own patterns — a real, shared fleet problem for services
(CV adopted it for exactly that, httputil/endpoints/README.md:97). docs-gates'
dispatch substrate is a plain Go map the binary itself owns end to end; it can
already enumerate its surface (`verbOrder` is the listing), and the help
golden pins it. The problem the library solves structurally cannot arise
here: there is no external router whose opaqueness needs a shadow registry.

**Root-contract bar.** Adoption would add a require and fail
scripts/stack-drift-check.sh:8-9, importing `net/http`-typed machinery
(`http.ServeMux`, `http.Handler`, ResponseWriter negotiation) that a
network-free exit-code CLI can never execute. Dead weight, contract broken.

**pkg/provider check.** No target. If the provider ever wanted to advertise
the battery's _verbs_ to BuildFlow, the natural shape is findings/Spec
metadata via toolsdk (already registered, pkg/provider/provider.go:187) — not
an HTTP route listing.

**Negotiation vs flag — the sharpest contrast.** endpoints' headline
mechanic is serving two representations of one URL picked by `Accept`
(httputil/endpoints/README.md:74-79). docs-gates has the same two-audience
problem (humans and the provider) and deliberately resolves it by _stream
separation under a flag_: `--json` reserves stdout for the machine document
and detours every human line to stderr, so a consumer never parses mixed
content (internal/battery/battery.go:97-98, :171-173). For an exec'd CLI
this is the strictly safer design — no heuristic can mis-pick the
representation — and it is byte-pinned by the battery-json golden. The
fleet's two solutions answer two different substrates; neither subsumes the
other.

**One honest parallel worth stating (not adoption-shaped):** the design
instinct behind endpoints — "the listing must be derived from the registration
site, never maintained by hand" — is already honored in docs-gates: the
battery's gate rows, `KnownGates`/`skip_gates` vocabulary, and the help
surface all derive from the same tables and are pinned by tests
(internal/config/config.go:48-80 names the single-source-of-truth rule for
gate names). The fleet lesson is absorbed; the library itself has no seam
here.

RECONSIDER-WHEN: a docs-gates sibling module grows an HTTP service (status
endpoint, dashboard backend) whose routes are worth self-describing — then
endpoints is the immediate right answer, zero-dependency and composable. For
the CLI as it stands, the insulation is correct.

## Ranked opportunities

None.

## How to verify

```console
$ git -C /home/lars/projects/httputil tag --list 'endpoints/*'
endpoints/v0.1.0
$ cat /home/lars/projects/httputil/endpoints/go.mod
module github.com/larsartmann/httputil/endpoints
go 1.27
# (no requires — stdlib-only submodule)

$ rg -n "httputil/endpoints" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**'
# rc=1 — zero hits anywhere (code, go.mod, go.sum)
$ rg -c "net/http" /home/lars/projects/docs-gates --type go
# rc=1 — no HTTP substrate for a route registry to mount on

# the consumer's self-describing surface (argv tables, not routes):
$ rg -n "var verbHandlers|var verbOrder|var verbUsage" /home/lars/projects/docs-gates/cmd/docs-gates/verbs.go
77:var verbHandlers = map[string]func([]string) int{
97:var verbOrder = []string{
129:var verbUsage = map[string]string{
$ rg -n "KnownGates lists every battery gate name" /home/lars/projects/docs-gates/internal/config/config.go
65:// KnownGates lists every battery gate name skip_gates may name. A
```
