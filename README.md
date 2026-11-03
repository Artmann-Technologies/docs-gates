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

Exit codes: `0` clean, `1` findings, `2` usage/environment error.

## Configuration

Optional `.docs-gates.json` at the repo root; every key defaults to the
historical constant (see `internal/config.Defaults`):

```json
{
  "report_dir": "docs/status",
  "living_docs": ["TODO_LIST.md", "ROADMAP.md", "CHANGELOG.md"],
  "generic_df": 8,
  "audit_max": 0,
  "cadence": { "struck_rows": 20, "max_unarchived": 15, "oldest_days": 21 },
  "skip_gates": []
}
```

Unknown keys warn; malformed JSON or invalid values exit 2 — a gate silently
running wrong thresholds is worse than a loud failure. JSON, not YAML: the
module is stdlib-only by contract, and JSON is jq-able.

## License

MIT.
