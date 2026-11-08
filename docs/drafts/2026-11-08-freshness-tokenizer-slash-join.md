> [!NOTE]
> This filing was drafted by GLM-5.3-Flash via Crush from an AI-run investigation, not at my request. When the failure traces to an external report, that source is linked in the body.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

## Problem

The freshness tokenizer keeps `/` inside a token (`internal/docsfreshness/tokens.go:41`):

```go
tokenRe  = regexp.MustCompile(`[a-z0-9][a-z0-9._/-]*`)
```

so a report item that slash-joins its keywords becomes ONE token that matches no living line, and the item fails as DRIFT even though every word is harvested. Real case from `LarsArtmann/timesheets` (docs-gates `2278406`):

```
item: 34. F8: cmdguard P4 help grouping/valid-args/positional IDs (#64).
token table (3 tokens):
  cmdguard                        DF=11   matched  generic
  help                            DF=8    matched  generic
  grouping/valid-args/positional  DF=0    missing  DISTINCTIVE
verdict: UNHARVESTED (only generic tokens matched)
```

The tracked TODO row contains "help grouping", "positional IDs", and "`WithValidArgs`" (the squash arm would match `valid-args`) — three distinct hits, all invisible because the tokenizer never split on `/`. Only `cmdguard` and `help` survived, both above GenericDF, so the item reported "accidental overlap, not a harvest".

Repro: `docs-freshness . --audit-all --explain-item "cmdguard P4 help grouping"` in a repo whose report line slash-joins keywords.

## Fix / Proposal

Drop `/` from the token character class (or post-split tokens on `/`) so `grouping/valid-args/positional` yields three tokens. Tradeoff to weigh: path-like tokens (`docs/handoff/x.md`) currently survive whole; after the change they match as segments instead — segment matching is probably closer to how living docs reference them, but it changes DF counts for any corpus that relies on whole-path tokens.

Workaround used meanwhile: the documented `<!-- fresh:skip: reason -->` annotation on the affected item — works, but every report author has to remember not to slash-join.

💘 Generated with Crush
