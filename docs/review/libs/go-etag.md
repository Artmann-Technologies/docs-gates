# go-etag utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v0.7.1 (latest local tag, fetched 2026-11-23) + HEAD be01b8a028d2085857378dfe1245c1186bb8a0c5 (`git describe` → `v0.7.1-6-gbe01b8a`). Nested modules at the v0.6.2 train: `client` v0.6.2, `entitytag` v0.6.2, `metrics` v0.6.2, `server` v0.6.2. Root module is a deprecated alias shim for `server`, removed at v1.0.0.
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 1 (resilience/error family) subagent of the multi-library utilization audit fleet.
- Fleet stance addressed per library: "if all my projects re-implement everything, I will only fix bugs all day and none benefit from collective improvements" — does stdlib-only insulation here avoid real re-implementation debt, or leave docs-gates hand-rolling something the fleet solved?

## Verdict

NOT-APPLICABLE — go-etag is a two-sided RFC 7232/9111 HTTP conditional-request library (server middleware + client RoundTripper), and docs-gates has no HTTP anywhere by architecture; its one content-hashing use (stub fingerprint digests) is stdlib checksum identity against config-pinned values, not validator negotiation.

## Score

- Inputs N = 15 (inventory rows below).
- A = 1 (the body-content-hashing concept row — a real docs-gates need, satisfied by two stdlib calls).
- M = 0.
- Score = round(100 × 0/1) = **0** — band **low**. As with the other group-1 NOT-APPLICABLE verdicts, the number measures leverage that is correctly zero; the verdict carries the meaning.

Library release shape (context for any future adoption): v0.7.1 was released as one
bottom-up staircase with the nested-module train v0.6.2 (`entitytag/` → `server/` →
`client/` → `metrics/`, per go-etag/CHANGELOG.md `[0.7.1] - 2026-11-18`) — an adopter pins
the root tag and gets the four nested modules at their train versions; the root module
itself is a deprecated alias shim slated for removal at v1.0.0, so any future import
should target `go-etag/server` / `go-etag/client` directly, never the root path.

- Applicable rows are exactly the inventory rows graded **Missed Opportunity** (the consumer has the underlying need; the capability is hand-rolled, solved in-domain, or declined by contract). Every **Not Applicable** row carries a one-line reason. `M` counts rows graded Fully Leveraged.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | `server`: ETag generation middleware — buffers the body, hashes it, sets the header automatically | No — no net/http, no server, no HTTP client in root or provider | N/A | go-etag/server/etag.go |
| 2 | `server`: `If-None-Match` → `304 Not Modified` with RFC 7232 weak comparison | No | N/A | go-etag/server/conditional_test.go (behavior surface), README "Why?" |
| 3 | `server`: correct 304/HEAD details (no body, Content-Length handling) | No | N/A | go-etag/server/etag_bdd_head_test.go, notmodified_test.go |
| 4 | `server`: streaming/hijack degradation past `MaxBufferSize`, `Flush` passthrough | No | N/A | go-etag/server/streaming_test.go, skip_test.go |
| 5 | `server`: `SkipIfPresent` passthrough so handler-set strong validators (DB revisions) win | No | N/A | go-etag/server/config.go |
| 6 | `client`: `http.RoundTripper` conditional-GET cache — replays stored validators as `If-None-Match` | No — the provider's only exec is the local battery binary; nothing fetches over HTTP | N/A | go-etag/client/transport.go; consumer pkg/provider/provider.go:148-171 (local exec only) |
| 7 | `client`: rebuilds 304 as 200 with the cached body, so SDKs see no difference | No | N/A | go-etag/client/rebuild.go |
| 8 | `client`: RFC 9111 cache semantics — header freshening (§4.3.4), unsafe-method invalidation (§4.4), HEAD freshen (§4.3.5), `no-store` (§3, §5.2.2.5) | No | N/A | go-etag/client/cache.go, spec_freshen_test.go, spec_invalidation_test.go |
| 9 | `client`: bounded memory — FIFO eviction (`MaxEntries`), body cap (`MaxBodyBytes`) | No | N/A | go-etag/client/options.go |
| 10 | `entitytag`: `ParseETag`/`ParseETagList` — RFC 7232 §2.3 header-value parsing | No — no `W/"…"`/`"…"`-shaped strings exist anywhere in docs-gates | N/A | go-etag/entitytag/entity_tag.go:111,143 |
| 11 | `entitytag`: `Strength`, `StrongEqual`/`WeakEqual`, `IsValid` comparison domain | No | N/A | go-etag/entitytag/entity_tag.go:18-100 |
| 12 | `metrics`: ready-made atomic counters for the server's observability hooks | No — durations are printed per run (battery matrix), never retained counters | N/A | go-etag/metrics/metrics.go |
| 13 | Root module alias shim (deprecated, removed at v1.0.0) | No | N/A | go-etag/deprecated.go |
| 14 | go-error-family classification of middleware errors (retry-aware observability) | No — see the go-error-family report | N/A | go-etag/server/go.mod (requires go-error-family v0.12.0) |
| 15 | Body-content hashing for change identity (the concept under the middleware) | Yes — docsstub fingerprints file bodies | Missed Opportunity (solved with stdlib) — md5/sha256 over the file bytes compared to config-pinned digests; go-etag exports no non-HTTP hashing API, so there is nothing to leverage even in principle | consumer internal/docsstub/docsstub.go:161-165; internal/config/content_gates.go:30-31,94 |

## What docs-gates does today

Zero usage, zero presence in either module graph. Verified:

```console
$ rg -n "go-etag|etag|If-None-Match|ETag" /home/lars/projects/docs-gates -g '!docs/review/**' --type go
# rc=1, no matches

$ rg -n "go-etag" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1, no matches
```

The nearest neighbors inside docs-gates, checked and cleared as non-duplications:

### Content fingerprinting — the only hashing in the consumer

`internal/docsstub/docsstub.go:20-21` imports `crypto/md5` + `crypto/sha256`; `:161-165`
computes both digests per corpus file and matches them against pinned values:

```go
md5Sum := md5.Sum(body) //nolint:gosec // checksum identity, not security
shaSum := sha256.Sum256(body)
// hex-encoded digests compared against config's stub_checksums
```

Pinned digests live in config (`internal/config/content_gates.go:30-31` defaults; `:94`
`validateStubChecksums` — every digest must be md5-32-hex or sha256-64-hex). This is
census-fingerprint identity over file bytes — the standing AGENTS.md decision ("crypto/md5 in
docsstub is census-fingerprint matching, not security"). It shares with ETag only the idea
"hash a body"; there is no validator exchange, no stored-representation comparison, no
conditional anything — and go-etag's hashing is internal to HTTP negotiation, not an exported
primitive.

### Change detection across runs is deliberately absent

Every battery run re-judges the corpus from bytes — "Gate verdicts are point-in-time under
foreign daemons — timestamp the run" (AGENTS.md). Caching verdicts behind content validators
would directly contradict the doctrine that a verdict is a fresh judgment. (The `result` cache
that exists in this ecosystem is BuildFlow's, addressed there as BuildFlow#87 — not
docs-gates'.)

### `.docs-gates-preflight.json` — a ledger, not a validator

The SAFE pre-flight batch stamp (`internal/docsfreshness/preflight_land.go:58`): a land-parity
record of which candidate lines landed, for editor-local attribution. It records content lines
and a date, not a digest for conditional requests — an anti-duplication ledger, the opposite
job of a cache validator.

No `file:line` in docs-gates hand-rolls If-None-Match parsing, entity-tag comparison,
conditional caching, or 304 semantics — the library's entire exported surface.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** This is the cleanest case in group 1 for the insulation half
of the argument: go-etag exists to stop every HTTP handler and HTTP client from hand-rolling
RFC 7232/9111 negotiation — a genuinely shared, subtle, spec-bound problem across the owner's
service projects. docs-gates has no HTTP surface at all ("No net/http, no network calls" is
contract architecture, reinforced by the hermetic nix build and the exec-only provider
boundary). There is no re-implementation to retire: the consumer's hashing is two stdlib digest
calls whose semantics (md5-or-sha256 hex match) are pinned by the config contract and goldens,
and the library offers no transport-agnostic hashing API that could replace even those.

**Root-contract bar.** Triply barred: the zero-require contract (`scripts/stack-drift-check.sh`
fails CI), the server/client modules pull `go-error-family` transitively, and net/http-based
middleware in a CLI that makes no network calls is pure dead weight.

**pkg/provider check.** No target either — its single external interaction is a local process
exec whose result is parsed strictly, once, per Detect (`provider.go:148-171`).

RECONSIDER-WHEN: docs-gates — much more likely a sibling fleet service — grows an HTTP surface:
a fleet status endpoint serving battery verdicts, a dashboard feeding go-health-dashboard, or a
remote report-corpus fetcher needing conditional caching. Then go-etag server/client slot into
that new module immediately (it is precisely the "stop re-implementing per project" case the
owner names), while the root stays network-free.

## Ranked opportunities

None.

## How to verify

```console
$ git -C /home/lars/projects/go-etag log -1 --format='%H'
be01b8a028d2085857378dfe1245c1186bb8a0c5
$ git -C /home/lars/projects/go-etag describe --tags
v0.7.1-6-gbe01b8a
$ git -C /home/lars/projects/go-etag tag --list | tail -2
v0.7.0
v0.7.1
$ git -C /home/lars/projects/go-etag tag --list "client/*" "entitytag/*" "server/*" "metrics/*" | sort | tail -4
client/v0.6.2
entitytag/v0.6.2
metrics/v0.6.2
server/v0.6.2

$ rg -n "go-etag|etag|If-None-Match|ETag" /home/lars/projects/docs-gates -g '!docs/review/**' --type go
# rc=1 (zero Go-file hits)
$ rg -n "go-etag" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 (not in any module graph)

# the consumer's only hashing (fingerprinting, not validators):
$ rg -n "crypto/md5|crypto/sha256" /home/lars/projects/docs-gates/internal --type go | rg -v _test
internal/docsstub/docsstub.go:20: "crypto/md5" //nolint:gosec // checksum identity, not security
internal/docsstub/docsstub.go:21: "crypto/sha256"
$ rg -n '"go-etag"' /home/lars/projects/docs-gates/scripts/citation-sweep.py
34:    "go-etag",          # audit-fleet scaffolding mention only, same as the other four libs
```
