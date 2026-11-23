package config

// Defaults returns the shipped configuration — the exact constants the
// gates carried before the config file existed (pbx-artmann's proven
// values). A repo without .docs-gates.json gets precisely the historical
// behavior.
func Defaults() Config {
	return Config{
		ReportDir:          "docs/status",
		ReportGlob:         "2*.md",
		ArchivedSubdir:     "archived",
		LivingDocs:         []string{"TODO_LIST.md", "ROADMAP.md", "CHANGELOG.md"},
		AnnotateExtraGlobs: []string{"docs/planning/*.md", "docs/planning/archived/*.md"},
		RuleF:              true,
		RuleFExtensions:    []string{".md", ".sh", ".nix"},
		GenericDF:          DefaultGenericDF,
		AuditMax:           DefaultAuditMax,
		AuditCorpus:        DefaultAuditCorpus,
		AnnotateMax:        DefaultAnnotateMax,
		AnnotateCorpus:     DefaultAnnotateCorpus,
		TodoFile:           "TODO_LIST.md",
		Cadence: Cadence{
			StruckRows:    DefaultStruckRows,
			MaxUnarchived: DefaultMaxUnarchived,
			OldestDays:    DefaultOldestDays,
		},
		GateTimeoutSec:  DefaultGateTimeoutSec,
		SkipGates:       nil,
		StubFiles:       []string{DefaultDomainLanguageFile},
		StubChecksums:   defaultStubChecksums(),
		StubExempts:     nil,
		StubEnforce:     false,
		CitationFiles:   []string{DefaultDomainLanguageFile},
		CitationExts:    defaultCitationExts(),
		CitationEnforce: false,
		StampFiles:         []string{DefaultDomainLanguageFile},
		StampWarnDays:      DefaultStampWarnDays,
		StampFailDays:      DefaultStampFailDays,
		StampEnforce:       false,
		VocabularyFiles:    []string{DefaultFeaturesFile},
		VocabularyEnforce:  false,
		AgentsFile:         DefaultAgentsFile,
		AgentsWarnDays:     DefaultAgentsWarnDays,
		AgentsMaxDays:      DefaultAgentsMaxDays,
		AgentsEnforce:      false,
	}
}

// Or is cmp.Or for comparable values (strings): the first non-zero
// argument wins. The Settings projections' scalar fallback rides this,
// the sibling of OrSlice.
func Or[T comparable](override, def T) T {
	var zero T
	if override != zero {
		return override
	}

	return def
}

// OrSlice is cmp.Or for slices (which are not comparable): the first
// argument carrying elements wins. Every Settings projection's list
// fallback rides this, so "a zero-value list falls back to the shipped
// Defaults" has one spelling across the gates.
func OrSlice[T any](override, def []T) []T {
	if len(override) > 0 {
		return override
	}

	return def
}
