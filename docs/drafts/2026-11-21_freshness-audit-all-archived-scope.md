> [!NOTE]
> This filing was drafted by DeepSeek-V4.1-Flash via Crush from an AI-run investigation, not at my request. The failure was found while repairing a red master in `go-sse`.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

TL;DR: `docs-freshness --audit-all` replays every archived report as well as the unarchived corpus, even though the verb's own usage text says "the WHOLE unarchived corpus", and its verdict is living-doc token frequency, so the unharvested count moves when unrelated prose is edited (go-sse measured 122 → 123 → 120 across three prose-only edits in one session). Add an archived-scope knob, or make the replay match the usage text.

## Problem

`internal/docsfreshness/corpus.go:173-185` (`statusReports`, at `e852a57`) appends the archived glob unconditionally:

```go
reports, _ := filepath.Glob(settings.reportGlobPath(root))
...
archived, _ := filepath.Glob(settings.archivedGlobPath(root))
...
return append(reports, archived...)
```

`auditAll` iterates that combined list (`internal/docsfreshness/audit.go:27-45`), so `--audit-all` covers both corpora. The usage text disagrees (`internal/docsfreshness/docsfreshness.go:60`):

```
  --audit-all      replay the WHOLE unarchived corpus
```

No flag or `.docs-gates.json` key scopes the replay. `known_keys.go` offers `audit_max` (ratchet pin) and `generic_df` (token bar), neither of which is a scope. The annotate gate has the same shape filed as #4; this is the freshness replay.

## Why it cannot be pinned

The verdict is distinctive-token frequency over the living docs. An item passes when at least one matched token appears on fewer than `GenericDF` (8) living-doc lines (`internal/docsfreshness/check.go:19-24`), so ordinary prose edits move archived items across the bar.

Measured in `go-sse` during one session, with prose-only edits to AGENTS.md / FEATURES.md / README.md / CONTRIBUTING.md and no archived report touched:

- before: 122 unharvested
- after one wording change: 123
- after the doc sync: 120 (`64 report(s) replayed, 120 unharvested open item(s)`)

`audit_max` therefore encodes the current prose, not the backlog. `generic_df` only moves the whole threshold: 111 tokens sat at DF 7 against the bar of 8 in that repo, so lowering it flips a large block at once. The only honest local lever today is `skip_gates: ["audit-all"]`, which turns the gate off instead of scoping it.

## Ask

Either:

1. `audit_skip_archived: true` (or `audit_corpus: "unarchived" | "all"`) in `.docs-gates.json`, defaulting to today's behavior; or
2. make `--audit-all` match its usage text (unarchived only) and add `--audit-archived` for the full replay.

Option 1 keeps existing consumers bit-identical. Either way the usage text and the implementation must agree.

## Verified

- Source read at `e852a57`: `statusReports` appends the archived glob unconditionally; `auditAll` consumes it.
- `docs-gates freshness . --audit-all` in `go-sse` reports 64 reports and 120 unharvested items after the prose-only doc sync.
- Dup search: #4 covers the annotate corpus scope, #5 is a tokenizer defect. Neither is this replay scope.

💘 Generated with Crush
