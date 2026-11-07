<!--
Title: annotate: an archived-corpus scope knob (skip docs/status/archived/** without skipping the gate)

Drafted by Crush (GLM-5.3-flash) from the go-taskqueue docs-gates adoption session, 2026-11-07.
-->

## Problem

The annotate corpus unconditionally includes the archived reports subdir — `internal/docsannotate/docsannotate.go:179-180` at `2278406`:

```go
docGlobs := append([]string{
	filepath.Join(settings.dir(), "*.md"),
	filepath.Join(settings.dir(), settings.archivedSub(), "*.md"),
}, settings.extraGlobs()...)
```

The config carries knobs to EXTEND the corpus (`annotate_extra_globs`, `internal/config/config.go:75`) and to ratchet findings (`annotate_max`, `config.go:80`), but nothing to SHRINK it. Archived reports are history-locked by definition: their struck items are frozen record, not living work. Any repo with a real archive accumulates annotate findings that can never be legitimately fixed — you cannot add a verdict to an archived report without falsifying history, and un-archiving defeats the point of the archive.

## What a consumer does today

go-taskqueue's `.docs-gates.json` carries `"skip_gates": ["annotate"]` — the whole gate is off, not because live-report annotations are unwanted, but because the archived slice alone can never green. The ratchet (`annotate_max`) is a poor fit: archived findings only ever GROW as more resolved reports get archived, so the pin chases the pile forever instead of scoping it.

## Ask

A scope knob, e.g. `annotate_skip_archived: true` (or an `annotate_exclude_globs` mirroring `annotate_extra_globs`) in `.docs-gates.json`: when set, drop the `archivedSub()` glob from the corpus at `docsannotate.go:180`. Default stays as-is (archived scanned), so nothing changes for existing repos.

## Verified

- Source read at `2278406`; no exclusion path exists between `collectDocFiles` and the glob list.
- Dup search: open+closed issues searched for "archived" / "annotate scope" — #3 and #2 are adjacent annotate defects/docs, not this knob.

💘 Generated with Crush
