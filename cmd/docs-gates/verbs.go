package main

// The verb tables: vocabulary, legacy display names, dispatch, display
// order, one-liners, and per-verb Usage wiring — one concern, one file.
// Deliberately cmd-local: config.KnownGates names the BATTERY rows (the
// skip_gates vocabulary — it has audit-all and lacks run/strike).
// Overlapping values, distinct concepts.

import (
	"github.com/Artmann-Technologies/docs-gates/internal/battery"
	"github.com/Artmann-Technologies/docs-gates/internal/docsannotate"
	"github.com/Artmann-Technologies/docs-gates/internal/docscadence"
	"github.com/Artmann-Technologies/docs-gates/internal/docscitations"
	"github.com/Artmann-Technologies/docs-gates/internal/docsfreshness"
	"github.com/Artmann-Technologies/docs-gates/internal/docspurge"
	"github.com/Artmann-Technologies/docs-gates/internal/docsrouting"
	"github.com/Artmann-Technologies/docs-gates/internal/docsrows"
	"github.com/Artmann-Technologies/docs-gates/internal/docsstamps"
	"github.com/Artmann-Technologies/docs-gates/internal/docsstrike"
	"github.com/Artmann-Technologies/docs-gates/internal/docsstub"
)

const (
	verbRun       = "run"
	verbFreshness = "freshness"
	verbAnnotate  = "annotate"
	verbRows      = "rows"
	verbRouting   = "routing"
	verbCadence   = "cadence"
	verbStub      = "stub"
	verbCitations = "citations"
	verbStamps    = "stamps"
	verbStrike    = "strike"
	verbPurge     = "purge"
)

// suggestMaxEditDistance is the typo tolerance of suggestVerb: any verb
// within this Levenshtein distance is offered as a did-you-mean.
const suggestMaxEditDistance = 2

// displayName maps each verb to the program name its --version line
// carries: the battery verb answers as docs-gates, the five legacy gate
// verbs answer under their legacy standalone names (the wrappers exec
// this binary, so `docs-freshness --version` must still read
// "docs-freshness <version>"). The three content gates (stub,
// citations, stamps) have no legacy wrapper — they answer under their
// own docs-* names, the same convention the legacy five set.
//
//nolint:gochecknoglobals // the fixed verb-to-legacy-name map, same class as the contract tables
var displayName = map[string]string{
	verbRun:       "docs-gates",
	verbFreshness: "docs-freshness",
	verbAnnotate:  "docs-annotate",
	verbRows:      "docs-gates",
	verbRouting:   "docs-routing",
	verbCadence:   "docs-cadence",
	verbStub:      "docs-stub",
	verbCitations: "docs-citations",
	verbStamps:    "docs-stamps",
	verbStrike:    "docs-strike",
	verbPurge:     "docs-gates",
}

// verbHandlers is the dispatch table: every verb to its gate's Run. The
// table IS the vocabulary — unknownVerb answers for anything not in it.
//
//nolint:gochecknoglobals // the fixed dispatch table, same class as the contract tables
var verbHandlers = map[string]func([]string) int{
	verbRun:       battery.Run,
	verbFreshness: docsfreshness.Run,
	verbAnnotate:  docsannotate.Run,
	verbRows:      docsrows.Run,
	verbRouting:   docsrouting.Run,
	verbCadence:   docscadence.Run,
	verbStub:      docsstub.Run,
	verbCitations: docscitations.Run,
	verbStamps:    docsstamps.Run,
	verbStrike:    docsstrike.Run,
	verbPurge:     docspurge.Run,
}

// verbOrder is the display order of the verbs in the root usage.
//
//nolint:gochecknoglobals // the fixed display order, same class as the contract tables
var verbOrder = []string{
	verbRun, verbFreshness, verbAnnotate, verbRows,
	verbRouting, verbCadence, verbStub, verbCitations, verbStamps,
	verbStrike, verbPurge,
}

// verbOverview is each verb's one-line purpose for the root usage; the
// detailed flag help lives in the gate packages' Usage.
//
//nolint:gochecknoglobals // the fixed verb one-liners, same class as the contract tables
var verbOverview = map[string]string{
	verbRun:       "all eight gates as one ordered verdict matrix (the battery)",
	verbFreshness: "the newest report's §f items must be harvested into the living docs",
	verbAnnotate:  "struck items must carry honest, verifiable verdicts",
	verbRows:      "table-row completeness and bare-row gates over explicit files",
	verbRouting:   "open actionable rows outside §f must be routed to a living doc",
	verbCadence:   "TODO_LIST purge and report archive thresholds",
	verbStub:      "a domain-language doc matching the bare template checksum is an unfilled stub",
	verbCitations: "backticked code citations in the glossary must resolve to real files",
	verbStamps:    "verification stamps age; oldest per file flagged past warn/fail bars",
	verbStrike:    "strike one TODO_LIST row (dry-run by default)",
	verbPurge:     "collect struck TODO rows into a CHANGELOG purge entry (dry-run by default)",
}

// verbUsage wires each verb to its package's Usage text — the same
// packages the dispatch table calls, so the help and the behavior
// cannot fork.
//
//nolint:gochecknoglobals // the fixed verb-to-usage map, same class as the contract tables
var verbUsage = map[string]string{
	verbRun:       battery.Usage,
	verbFreshness: docsfreshness.Usage,
	verbAnnotate:  docsannotate.Usage,
	verbRows:      docsrows.Usage,
	verbRouting:   docsrouting.Usage,
	verbCadence:   docscadence.Usage,
	verbStub:      docsstub.Usage,
	verbCitations: docscitations.Usage,
	verbStamps:    docsstamps.Usage,
	verbStrike:    docsstrike.Usage,
	verbPurge:     docspurge.Usage,
}
