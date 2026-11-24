# docs-gates

The docs-health battery gates as one standalone, stdlib-only Go binary.
The battery runs ten gates — freshness, audit-all, annotate, routing,
cadence, stub, citations, stamps, vocabulary, agents — as one verdict matrix;
`rows` is the read-only report-table gate, and `strike`/`purge` are
guarded mutator verbs outside the battery.

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
nix run github:Artmann-Technologies/docs-gates/v0.5.0 -- run .
# or
go install github.com/Artmann-Technologies/docs-gates/cmd/docs-gates@v0.5.0
docs-gates run .
```

Verbs: `run | freshness | annotate | rows | routing | cadence | stub | citations | stamps | vocabulary | agents | strike | purge`.
The five original gate verbs keep their historical standalone names
(`docs-freshness`, `docs-annotate`, `docs-routing`, `docs-cadence`,
`docs-strike`) as flake wrappers and answer `--version` under those
names. `rows` and the content gates (`stub`, `citations`,
`stamps`, `vocabulary`, `agents`) were born inside this binary and deliberately
carry no legacy wrapper. `docs-gates help` (or `<verb> --help`)
explains every verb and its flags; a typo'd verb gets a did-you-mean
suggestion.

`docs-gates rows <report.md>` judges one report's tables (PARTIAL and
mixed rows; `--bare` for bare rows, `--strict` to also flag wholly
unstruck tables) — the check to run before harvesting a report. It is
read-only end to end; the writer's `--normalize` rewrite stays in
`annotate.go`.

### Exit codes

`0` clean, `1` findings, `2` usage/environment error — one contract for
every verb:

| Verb        | `0`                                 | `1`                                                  | `2`                                                        |
| ----------- | ----------------------------------- | ---------------------------------------------------- | ---------------------------------------------------------- |
| `run`       | every gate PASS or SKIP             | ≥ 1 gate FAIL (a gate's own rc=2 becomes a FAIL row) | bad repo root / bad config only                            |
| `freshness` | FRESH                               | DRIFT — unharvested §f items                         | usage error, malformed config                              |
| `annotate`  | clean or within the ratchet         | annotation findings past the pin                     | usage error, malformed config                              |
| `rows`      | every table complete / no bare rows | PARTIAL or mixed rows (`--bare`: bare rows)          | usage error, unreadable file                               |
| `routing`   | no routing gaps                     | open rows no living doc represents                   | usage error, malformed config                              |
| `cadence`   | purge/archive thresholds met        | a threshold is overdue                               | usage error, malformed config                              |
| `strike`    | row struck (dry-run is default)     | —                                                    | refusal: ambiguous row, double strike, verdict-less target |
| `purge`     | draft printed (dry-run) or purged   | unsafe draft (the gate refuses)                      | usage error, missing [Unreleased]/Changed anchors          |

**Judge vs writer.** `docs-gates` reads and judges a corpus; it never
rewrites one. Annotation (striking, repairing, marking verdicts) belongs to
the docs-health skill's `annotate.go`, which shares this binary's marker
grammar. The boundary, the executable grammar test, and the migration plan
live in [ADR 0001](docs/adr/0001-marker-grammar-ownership.md). Measured
2026-11-15: the writer as `go run` costs ~1.9 s cold / ~25 ms warm, the
judge binary answers in ~1.5 ms — pin the binary in CI, keep the writer
zero-install.

## Adopting docs-gates

A repo adopts in three steps; [cmdguard](https://github.com/Artmann-Technologies/cmdguard)
is the worked example — its adoption is exactly this section, verbatim.

**1. Run the battery once.** No config needed: the defaults are the
fleet-proven constants and a missing `.docs-gates.json` is not a finding.
Pin the release — a floating install can drift from the battery your CI
judged:

```console
go run github.com/Artmann-Technologies/docs-gates/cmd/docs-gates@v0.5.0 run .
```

**2. Wire CI (the dogfood job).** Full history is load-bearing: the
annotate gate resolves `done at <hash>` verdicts against the real commit
graph, and a depth-1 checkout makes every non-HEAD hash unresolvable
(false failures):

```yaml
dogfood:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v7 # with: { fetch-depth: 0 }
    - uses: actions/setup-go@v7 # with: { go-version: "1.27" }
    - run: go run github.com/Artmann-Technologies/docs-gates/cmd/docs-gates@v0.5.0 run .
```

Bump the pin manually when a docs-gates release lands (dependabot does
not see `go run` pins).

**3. Ignore the land-parity stamp.** A SAFE `--pre-flight` records its
batch in `.docs-gates-preflight.json` (local editor state, never
shared). One `.gitignore` line keeps an auto-commit daemon from sweeping
it into history:

```gitignore
.docs-gates-preflight.json
```

Then adopt the rituals the gates enforce: strike landed TODO rows
(`docs-gates strike`, or `scripts/safe-strike.sh` copied from this repo
for the guarded mechanics), archive resolved reports, purge struck rows
into CHANGELOG (`docs-gates purge`). A red gate is the docs ritual
due — fix the docs, never loosen the gate.

**Known consumers** (the full install-coverage survey with per-repo
pins lives in `docs/planning/` and is refreshed after each release
install): crm, CV, go-appkit, webphone, and pbx-artmann run the
docs-health writer (`annotate.go`) via `go run` from the skill fan-out;
the docs-health SKILL.md pins the writer/judge subcommand map; this
repo's `internal/docsrows` parity tests execute the live writer as
their oracle.

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
  "audit_corpus": "all",
  "annotate_max": -1,
  "annotate_corpus": "all",
  "cadence": { "struck_rows": 20, "max_unarchived": 15, "oldest_days": 21 },
  "gate_timeout_sec": 300,
  "skip_gates": []
}
```

The ratchet keys are strict by default (`-1` omits the flag entirely);
a repo with measured debt pins the count (`audit_max`/`annotate_max` >= 0)
and drains it down — REGRESSION past the pin fails. `audit_corpus`
scopes the freshness replay (`--audit-all` and `--pre-flight`): `"all"`
(default) replays the report dir plus its archived subdir;
`"unarchived"` drops the archived subdir, so an archived backlog stops
moving the unharvested count with unrelated living-doc prose edits.
`annotate_corpus` scopes the annotate corpus the same way (`"all"`
default; `"unarchived"` drops the archived subdir, whose history-locked
strikes can never be legitimately repaired — repos that only want the
live slice annotated no longer need to skip the whole gate).
`vocabulary_files` (default `FEATURES.md`) and `vocabulary_enforce`
drive the vocabulary gate: every table Status-column token must be
defined by the file's own `## Status legend` section; findings WARN
until `vocabulary_enforce` flips.
`agents_file` (default `AGENTS.md`), the age bars `agents_warn_days`
(default 7) and `agents_max_days` (default 14), and `agents_enforce`
drive the agents gate: the file's mtime is flagged WARN past the warn
bar, FAIL past the max bar (a missing file is always fail severity);
findings WARN until `agents_enforce` flips.
`gate_timeout_sec`
caps each battery gate's wall clock (default 300s; `0` disables) — an
expired gate becomes a FAIL row (`rc=timeout`) and the battery moves
on, so a wedged gate can never hang the run. `skip_gates` entries
must name real battery gates (`freshness`, `audit-all`, `annotate`,
`routing`, `cadence`, `stub`, `citations`, `stamps`, `vocabulary`,
`agents`) — a
typo exits 2 rather than silently running the
gate you meant to drop. Unknown keys warn; malformed JSON or invalid
values exit 2 — a gate silently running wrong thresholds is worse than a
loud failure. JSON, not YAML: the module is stdlib-only by contract, and
JSON is jq-able.

## Annotate verdict rules

The annotate gate judges every struck (`~~…~~`) item in the report
corpus. The rules in source order (`internal/docsannotate`):

| Rule                                  | Verdict | What it requires                                                                                                                                       |
| ------------------------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `A/evidence`                          | FAIL    | A struck item's verdict text (everything outside the `~~…~~`) must carry a marker word — see the vocabulary below.                                     |
| `B/hash`                              | FAIL    | `` `done at <sha>` `` citations must resolve to a real commit or tag via git.                                                                          |
| `C/path`                              | FAIL    | Backticked repo-relative paths in the verdict must exist on disk; paths inside the strike are exempt.                                                  |
| `D/verdict-leads-open`                | INFO    | A verdict marker followed by an unfinished phrase ("still open", "owed", …) is reported, never failed.                                                 |
| `E/ref`                               | FAIL    | Tag-existence and removed-path claims are verified against the repo.                                                                                   |
| `F/spacing`, `F/walk`                 | FAIL    | Rule F: no letter/digit may touch a backtick; `<!-- format:allow -->` exempts a line (dprint strips spaces around inline code with escaped backticks). |
| `G/uniform`                           | FAIL    | Every table cell needs an even `~~` count — strike per cell, never across the pipe (a cross-cell strike renders as a broken row).                      |
| `P/verdict-inside`, `P/number-inside` | FAIL    | The verdict word and the list-item number must sit OUTSIDE the `~~` spans — inside, the renderer deletes the verdict and renumbers the list tail.      |

Marker vocabulary (case-insensitive, word-boundary matched): `done`,
`resolved`, `routed`, `superseded`, `fixed`, `adopted`, `decided`,
`verified`, `won't implement`, `not-do`, `duplicate`, `struck`,
`moot`, `no longer`. `routed` counts so striking a routed row —
`~~item~~ **→ routed to TODO_LIST #32**` — closes it the same way the
routing gate does; the unstruck routed marker stays the rows gate's
jurisdiction. Annotate never strikes an open item: the D rule reports
the smell, the A rule refuses the bare strike. Exempt one §f item from
harvest with `<!-- fresh:skip: reason -->`; the ratchet
(`annotate_max`) tolerates a pinned finding count green-but-named
while a corpus drains.

## Scripts

| Script                          | What it does                                                                                                                                                                                                              |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `scripts/safe-strike.sh`        | Strike TODO/report rows by unique task-text anchor: exactly-one-match, cell-bounded rewrites, dry-run default, `--gate`/`--battery` verify per batch — the mechanical rails from the 2026-11-04 strike-corruption lesson. |
| `scripts/safe-strike-test.sh`   | The refusal-matrix harness for safe-strike (11 cases on a scratch git repo; all must behave).                                                                                                                             |
| `scripts/stack-drift-check.sh`  | Fails the moment the root `go.mod` grows a require/replace/exclude — the stdlib-only contract, wired into CI.                                                                                                             |
| `scripts/check-erraudit-rev.sh` | Tripwire pinning the erraudit flake input's locked rev (also a flake check).                                                                                                                                              |

## Freshness diagnostics

How an open §f item binds to the living docs (2026-11-24 semantics): an
item passes only when ONE living-doc line co-carries its matched tokens
— two matched tokens by default, one for single-token items — with at
least one token distinctive (document frequency below `generic_df`).
Tokens scattered across separate lines are coincidence, not a harvest,
and the item fails naming the near-miss. Two WARNs guard report grammar
without flipping verdicts: a report with no `## f)` heading (freshness
passes by omission — that blind spot should be loud), and a §f-section
bullet line that reads like a next step but is not a harvestable item
(only numbered lines and table rows are). The routing gate judges
open-row binding with the same per-line bar, so the two gates cannot
disagree about whether a row is represented.

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
- `--land-parity` — after a SAFE pre-flight batch, verify the landed
  bytes: the recorded stamp (`.docs-gates-preflight.json`, gitignored)
  is compared against the living docs, and a retyped edit that sailed
  through on a verdict earned by different bytes is refused with every
  missing line named. No stamp means OK (opt-in state); a malformed
  stamp is a usage error with the recovery.

## Releases

v0.5.0 is current (see [CHANGELOG.md](CHANGELOG.md)); the BuildFlow
provider module is versioned in lockstep as `pkg/provider/v0.5.0`.
`--version` prints the flake's shortRev, so between tags it names a
commit, not a release.

Release checklist — push and install are part of the release, not an
afterthought:

1. Cut the CHANGELOG section; bump the version pins in README and
   FEATURES.
2. `buildflow --build-mode full` green, plus the battery on this repo
   and one adopting repo.
3. Annotated tag `vX.Y.Z`; push the branch and the tag; GitHub release
   from the CHANGELOG section.
4. Install over the previous fleet binary and VERIFY `docs-gates
   --version` prints the new rev before any caller is switched to new
   verbs.

Contributing: [CONTRIBUTING.md](CONTRIBUTING.md); security reporting:
[SECURITY.md](SECURITY.md); the paid-for incident narratives behind
AGENTS.md live in [docs/agent-notes.md](docs/agent-notes.md).

## License

MIT.
