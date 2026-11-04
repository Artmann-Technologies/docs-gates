# docs-gates

The docs-health battery gates as one standalone, stdlib-only Go binary.
Five gates — freshness, annotate, routing, cadence, strike — plus `run`,
the ordered battery that executes them as one verdict matrix.

Extracted from the pbx-artmann deployment flake's `pbx-tools` package, where
the gates were hand-rolled per repo. Every behavioral constant is now
config-driven per repo (`.docs-gates.json`), the defaults are the proven
values, and the battery shape (independent gates, one matrix, no
first-failure abort) is preserved byte-for-byte at the exit-code level.

## Why

The docs-health doctrine (status reports + TODO_LIST/ROADMAP/CHANGELOG
living docs, struck items carry resolvable verdicts, reports archive at a
cadence) was enforced by tooling that lived inside one project. A second
repo adopting it meant copying Go packages or writing fresh shell. This
repo is the one tool any repo can adopt: drop a binary, optionally seed a
`.docs-gates.json`, run the battery.

## Quick start

```console
nix run github:Artmann-Technologies/docs-gates -- run .
# or
go install github.com/Artmann-Technologies/docs-gates/cmd/docs-gates@latest
docs-gates run .
```

Verbs: `run | freshness | annotate | routing | cadence | strike`. Each gate
verb keeps its historical standalone name (`docs-freshness`,
`docs-annotate`, `docs-routing`, `docs-cadence`, `docs-strike`) as a flake
wrapper, and answers `--version` under that name.

### Exit codes

`0` clean, `1` findings, `2` usage/environment error — one contract for
every verb:

| Verb        | `0`                             | `1`                                                  | `2`                                                        |
| ----------- | ------------------------------- | ---------------------------------------------------- | ---------------------------------------------------------- |
| `run`       | every gate PASS or SKIP         | ≥ 1 gate FAIL (a gate's own rc=2 becomes a FAIL row) | bad repo root / bad config only                            |
| `freshness` | FRESH                           | DRIFT — unharvested §f items                         | usage error, malformed config                              |
| `annotate`  | clean or within the ratchet     | annotation findings past the pin                     | usage error, malformed config                              |
| `routing`   | no routing gaps                 | open rows no living doc represents                   | usage error, malformed config                              |
| `cadence`   | purge/archive thresholds met    | a threshold is overdue                               | usage error, malformed config                              |
| `strike`    | row struck (dry-run is default) | —                                                    | refusal: ambiguous row, double strike, verdict-less target |

## BuildFlow consumers

The battery self-registers as a [BuildFlow](https://github.com/LarsArtmann/BuildFlow)
tool via `pkg/provider` (its own Go module; the root module stays
stdlib-only). Blank-import it and BuildFlow's detect step finds the
battery wherever the `docs-gates` binary is on PATH:

```go
import _ "github.com/Artmann-Technologies/docs-gates/pkg/provider"
```

One finding per red gate row; read-only by contract.

## Configuration

Optional `.docs-gates.json` at the repo root; every key defaults to the
historical constant (see `internal/config.Defaults`):

```json
{
  "report_dir": "docs/status",
  "living_docs": ["TODO_LIST.md", "ROADMAP.md", "CHANGELOG.md"],
  "generic_df": 8,
  "audit_max": -1,
  "annotate_max": -1,
  "cadence": { "struck_rows": 20, "max_unarchived": 15, "oldest_days": 21 },
  "skip_gates": []
}
```

The ratchet keys are strict by default (`-1` omits the flag entirely);
a repo with measured debt pins the count (`audit_max`/`annotate_max` >= 0)
and drains it down — REGRESSION past the pin fails. `skip_gates` entries
must name real battery gates (`freshness`, `audit-all`, `annotate`,
`routing`, `cadence`) — a typo exits 2 rather than silently running the
gate you meant to drop. Unknown keys warn; malformed JSON or invalid
values exit 2 — a gate silently running wrong thresholds is worse than a
loud failure. JSON, not YAML: the module is stdlib-only by contract, and
JSON is jq-able.

## Freshness diagnostics

Three read-only inspection modes on the freshness verb:

```console
docs-gates freshness . --margins
docs-gates freshness . --dump-items docs/status/<report>.md
echo "- the new TODO row text" | docs-gates freshness . --pre-flight
```

- `--margins` — every living-docs token within 2 document-frequency of
  the generic bar, so drafts avoid those words first-try instead of
  discovering the boundary through a failed run. Informational: never
  flips a verdict.
- `--dump-items <report>` — the §f extraction decisions for one report:
  which lines became open items and the tokens each carries. A report
  without a §f section is a usage error (exit 2).
- `--pre-flight` — judge candidate living-doc lines (stdin) BEFORE the
  edit lands: the replay reports items that would newly flip to DRIFT,
  boundary-crossing tokens, and where each matched token lives.
  Rewrites name the old lines with repeatable `--replace=the old line`:
  they are removed before the replay so the rewrite is judged by its new
  words only. `--duplicates` models a deliberate second copy instead of
  refusing one; `--explain-token` adds per-token tables.

## Releases

v0.1.1 is current (see [CHANGELOG.md](CHANGELOG.md)).

## License

MIT.
