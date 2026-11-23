# go-codec utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v0.3.4 (exact tag; HEAD == tag: 8d7a27477cb309633690699d3b719895cf82b256, `git describe` → `v0.3.4`), module `github.com/larsartmann/go-codec`, package `codec`.
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 2 (HTTP/serialization) subagent of the multi-library utilization audit fleet.
- Fleet stance addressed: does stdlib-only insulation here avoid re-implementation debt, or leave docs-gates hand-rolling something the fleet solved?

## Verdict

NOT-ADOPTED — go-codec is the event-sourcing stack's payload-serialization contract (CBOR/JSON/Raw codecs, determinism for signing, envelopes, transcoding, pooling, observability), while every JSON touch in docs-gates is a thin, deliberately-shaped stdlib call whose semantics are contract, not implementation: the config loader's unknown-keys-WARN two-pass decode, the battery's byte-pinned `--json` face, and the provider's strict json/v2 decode — none is codec-shaped duplication, and adopting go-codec would remove capability (its JSON path has no unknown-field rejection and forces case-insensitive matching in the v2 build).

## Score

- Inputs N = 14 (inventory rows below).
- A = 2 (the two concept rows that genuinely apply: unknown-field-posture decoding, and one-document JSON serialization for an external machine consumer — both satisfied by direct stdlib calls).
- M = 0.
- Score = round(100 × 0/2) = **0** — band **low**. The verdict carries the meaning; the zero measures that no codec *machinery* is needed, not that a need went unmet.

- Applicable rows are exactly the inventory rows graded **Missed Opportunity** (the consumer has the underlying need; the capability is hand-rolled, solved in-domain, or declined by contract). Every **Not Applicable** row carries a one-line reason. `M` counts rows graded Fully Leveraged.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | `Codec` interface — `Encoding`/`Encode`/`Decode` over CBOR/JSON/Raw | No — no payload format variance; every reader/writer is JSON by contract | N/A | go-codec/codec.go:56-64 |
| 2 | `CBORCodec` — canonical RFC 7049, deterministic, signing-safe | No — nothing signed, no binary format on disk or wire | N/A | go-codec/cbor.go:19; go-codec/README.md:64-69 |
| 3 | `CBORCompactCodec` — RFC 8949 Core Deterministic + unknown-field rejection on decode | No — docs-gates' strict-decode need is JSON-shaped and already strict (see row 13) | N/A | go-codec/README.md:110-122 |
| 4 | `RawCodec` passthrough | No — no pre-encoded payload handoff | N/A | go-codec/raw.go:9 |
| 5 | `BufferEncoder` / `EncodePooled` — zero-allocation hot-path encoding | No — one config read and one report emit per process; no hot loop | N/A | go-codec/codec.go:67; go-codec/pool.go:66 |
| 6 | Streaming encoders/decoders (CBOR stream, NDJSON `NewJSONEncoder`/`NewJSONDecoder`) | No — one JSON document per run by contract, not a stream of values | N/A | go-codec/streaming.go:13-42; go-codec/README.md:281-308 |
| 7 | `WrapEncode`/`UnwrapDecode` self-describing envelopes + `DecodeEnvelopeOrLegacy` | No — no persisted events with a legacy fallback story | N/A | go-codec/envelope.go:23, :48, :85 |
| 8 | `AutoDetect`/`AutoDetectDebug` — explainable format detection for mixed streams | No — exactly one format exists (JSON) | N/A | go-codec/autodetect.go:63, :71 |
| 9 | `TranscodeToJSON` — schema-free CBOR→JSON bridge | No — nothing is stored in CBOR | N/A | go-codec/transcode.go:37 |
| 10 | `Size`/`SizeResult` payload-size diagnostics | No — no format trade-off to make | N/A | go-codec/size.go:17 |
| 11 | `ObservableCodec` metrics/hooks wrapper | No — per-run CLI; nothing to meter | N/A | go-codec/observability.go:35 |
| 12 | COSE Sign1/Encrypt0 structures, base64-JSON helpers | No — no signing/encryption | N/A | go-codec/cose.go:168; go-codec/base64_json.go:34 |
| 13 | Structured decode with an explicit unknown-field posture | Yes in concept — two of docs-gates' three decoders take unknown fields seriously | Missed Opportunity (solved with stdlib) — config: WARN pass via `map[string]json.RawMessage` then struct decode; provider: json/v2 `RejectUnknownMembers(true)`; go-codec's JSON path offers no rejection mode (CBOR-only) | consumer internal/config/config.go:152-170; pkg/provider/provider.go:95-99 |
| 14 | One-document JSON serialization for a machine consumer | Yes in concept — the battery `--json` face | Missed Opportunity (solved with stdlib) — `json.MarshalIndent` + snake_case tags, byte-pinned by the battery-json golden; go-codec's JSONCodec is the same Marshal wrapped (and v2-mismatched, see gap analysis) | consumer internal/battery/battery.go:295; go-codec/json.go:14-20 |

Dependency note: go-codec is decidedly not stdlib-only — go-codec/go.mod:5-11
requires `fxamacker/cbor/v2`, `go-snaps`, `go-error-family`, `gomega`, and
`rapid`. Root-module adoption is dead on arrival against
scripts/stack-drift-check.sh:8-9; only pkg/provider is even arithmetically
eligible.

## What docs-gates does today

Zero usage, zero presence. Verified:

```console
$ rg -n "github.com/larsartmann/go-codec" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**'
# rc=1 — zero hits in code or module graphs (the prose mentions in CHANGELOG/
# flake.nix/scripts cite go-codec's *erraudit tripwire precedent*, not the module)
$ rg -n "go-codec" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 — not required anywhere, not even indirect
```

The consumer's complete JSON surface, examined for codec-shape (the brief's
question: are these wrappers, or just stdlib calls?):

1. **Config loader** (internal/config/config.go:137-177): read file, first
   `json.Unmarshal` into `map[string]json.RawMessage` to WARN on unknown keys
   (config.go:152-166), second `json.Unmarshal` over the pre-filled defaults
   struct (config.go:168), then domain validation. This is a *policy*
   (unknown keys must warn, malformed must exit 2) expressed in ~20 lines over
   two stdlib calls — there is no codec abstraction to factor out, and the
   two-pass shape is itself the documented decision (config.go:9-14).
2. **Battery JSON face** (internal/battery/battery.go:295): one
   `json.MarshalIndent` of the report struct; snake_case tags enforced by the
   tagliatelle lint rule; byte shape pinned by the battery-json golden and
   mirrored by the provider's strict decoder (AGENTS.md:73-76). Again a
   single stdlib call wrapped in contract, not a codec.
3. **Pre-flight stamp** (internal/docsfreshness/preflight_land.go:49-58 write,
   :110-124 read): `MarshalIndent`/`Unmarshal` of a tiny local ledger file —
   the smallest possible codec-shaped site, and already deliberate.
4. **Provider strict decode** (pkg/provider/provider.go:92-102): stdlib
   `encoding/json/v2` — `jsontext.NewDecoder` + `json.UnmarshalDecode` with
   `json.RejectUnknownMembers(true)` (pkg/provider/provider.go:34-35 imports).
   This is the row-shape contract: an unknown field fails the parse and
   surfaces as its own finding instead of silently reading zero values.

**Are these codec-shaped?** No — by the fleet's own taxonomy they are
contract-bearing stdlib calls: each site is one or two functions, the whole
`encoding/json` import surface is eight files (two of them tests; the
provider's two are the `encoding/json/v2`/`jsontext` subpackage paths), and the semantics (WARN-vs-reject posture, byte-pinned output, exit-code
mapping) are the product, not the encoding mechanics. go-codec's value
columns — format variance, determinism, envelopes, pooling, telemetry —
address problems docs-gates does not have.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** go-codec is the shared serialization
contract for the event-sourcing stack (go-cqrs-lite's storage/event/signing/
encryption modules, go-codec/README.md:14-46) — exactly the "stop
re-implementing per project" shape the owner names, for projects that
serialize many payload types across stores. docs-gates serializes exactly two
shapes (a config file it defines, a report document it defines) at
process-per-run frequency, and the standing decisions already adjudicated the
adjacent questions: "JSON, not YAML: the module is stdlib-only by contract"
(internal/config/config.go:4-7) and jsonv1tov2 declined for the battery face
because the strict row-shape decode is the byte-pinned contract
(AGENTS.md:123-127). Handing these to a codec abstraction would not retire
duplicated machinery — there is none to name, honestly — it would relocate
contract semantics into a library whose defaults differ.

**pkg/provider check (the one eligible home).** Adopting go-codec's
`JSONCodec` there would be a net regression, verified against source: (a) its
`Decode` is a bare `json.Unmarshal` wrapper with no unknown-field rejection
(go-codec/json.go:18-20) — the provider's strictness would have to be
re-implemented *around* the library; (b) in the v2 build it forces
`json.MatchCaseInsensitiveNames(true)` (go-codec/json_compat_v2.go:24-26),
which is precisely the wrong posture for a snake_case-pinned row contract;
(c) it would add the cbor/v2 tree to the provider's go.sum for zero used
capability. The provider's three-line stdlib decoder is smaller than the
adapter would be.

**Root-contract bar.** Doubly barred: the require block (go-codec/go.mod:5-11)
fails stack-drift-check, and the battery face's byte-pinning means any
encoding change is a golden bump by definition — the CHANGELOG ritual costs
more than the codec could ever save.

RECONSIDER-WHEN: docs-gates grows a persistence or interchange layer with
format variance — e.g. caching battery verdicts as signed/deterministic
payloads, or exporting findings in a pluggable CBOR/JSON wire format for a
fleet event bus. That layer should live beside pkg/provider (its own module)
and adopt go-codec there, where `DeterministicCodec` and the envelope helpers
are load-bearing. For the current two-shape JSON contract, stdlib is the
correct weight class.

## Ranked opportunities

None.

## How to verify

```console
$ git -C /home/lars/projects/go-codec log -1 --format='%H'
8d7a27477cb309633690699d3b719895cf82b256
$ git -C /home/lars/projects/go-codec describe --tags
v0.3.4

$ rg -n "github.com/larsartmann/go-codec" /home/lars/projects/docs-gates --hidden -g '!.git' -g '!docs/review/**'
# rc=1 — zero hits (prose mentions of the erraudit precedent are
# "go-codec scripts/check-erraudit-version.sh", not the module path)
$ rg -n "go-codec" /home/lars/projects/docs-gates/go.mod /home/lars/projects/docs-gates/pkg/provider/go.mod /home/lars/projects/docs-gates/pkg/provider/go.sum
# rc=1 — not in any module graph

# the consumer's whole JSON surface:
$ rg -l "encoding/json" /home/lars/projects/docs-gates --type go
./pkg/provider/provider.go
./internal/config/known_keys.go
./internal/config/errors.go
./internal/config/config_carrier_test.go
./internal/config/config.go
./internal/docsfreshness/preflight_land.go
./internal/battery/battery_json_test.go
./internal/battery/battery.go

# go-codec's JSON path, checked for the provider-fit claim:
$ rg -n "Decode unmarshals|MatchCaseInsensitiveNames|RejectUnknown" /home/lars/projects/go-codec/json.go /home/lars/projects/go-codec/json_compat_v2.go
/home/lars/projects/go-codec/json_compat_v2.go:25:	return json.Unmarshal(data, v, json.MatchCaseInsensitiveNames(true)) //nolint:wrapcheck
/home/lars/projects/go-codec/json_compat_v2.go:89:	return json.UnmarshalDecode(d.dec, v, json.MatchCaseInsensitiveNames(true)) //nolint:wrapcheck
/home/lars/projects/go-codec/json.go:17:// Decode unmarshals JSON bytes into a value.
# (no RejectUnknown* in the JSON path — unknown-field rejection exists only for CBORCompact)
```
