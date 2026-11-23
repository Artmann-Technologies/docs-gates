package config

// The content gates (docs-stub, docs-citations, docs-stamps,
// docs-agents) and their config surface: the shared glossary corpus,
// the census-pinned bare-template digests, the citation extension
// list, the stamp age bars, and the agents freshness bars. One file so
// the content-gate contract reads as one table.

// DefaultDomainLanguageFile is the three content gates' shared default
// corpus: the domain-language glossary every LarsArtmann repo carries
// under docs/. One name so the stub/citations/stamps Settings
// projections cannot drift apart.
const DefaultDomainLanguageFile = "docs/DOMAIN_LANGUAGE.md"

// DefaultFeaturesFile is the vocabulary gate's default corpus: the
// feature inventory whose status columns the legend contract judges.
const DefaultFeaturesFile = "FEATURES.md"

// DefaultAgentsFile is the agents gate's default corpus: the
// AI-collaboration contract every fleet repo carries at its root.
const DefaultAgentsFile = "AGENTS.md"

// defaultStubChecksums pins the bare DDD template fleet-wide: 38 repos
// shipped byte-identical copies of it on census day (2026-11-21,
// ~/projects/docs status/evidence/2026-11-21_domain-language-identical-file-groups.txt).
// Both digests are of the same bytes; a match on either is the finding.
// A function, not a package-level slice: every caller gets fresh storage.
func defaultStubChecksums() []string {
	return []string{
		"5b32f8c3a96af80ad148696b539fb977",                                 // md5
		"0cd7d0c00ba6170e7ccff67c922a29d3211d28b3525a18f5f25b83dfcc9fe60d", // sha256
	}
}

// defaultCitationExts are the file extensions the citations gate counts
// as code citations (dot-prefixed, repo-relative paths only).
func defaultCitationExts() []string {
	return []string{".go", ".ts", ".tsx", ".rs", ".py", ".sh", ".sql", ".templ", ".nix"}
}

// validateContentGates enforces the content-gate value contract: one
// validator per gate, first refusal wins.
func (c Config) validateContentGates() error {
	for _, err := range []error{
		c.validateStampBars(),
		c.validateAgentsBars(),
		c.validateStubChecksums(),
		c.validateCitationExts(),
	} {
		if err != nil {
			return err
		}
	}

	return nil
}

// validateStampBars keeps both stamp age bars positive and the FAIL bar
// never below the WARN bar.
func (c Config) validateStampBars() error {
	if c.StampWarnDays < 1 || c.StampFailDays < 1 {
		return invalidf("stamp thresholds must be >= 1, got warn=%d fail=%d", c.StampWarnDays, c.StampFailDays)
	}

	if c.StampFailDays < c.StampWarnDays {
		return invalidf(
			"stamp_fail_days (%d) must be >= stamp_warn_days (%d) — the FAIL bar cannot sit below the WARN bar",
			c.StampFailDays,
			c.StampWarnDays,
		)
	}

	return nil
}

// validateAgentsBars keeps both agents age bars positive and the max
// bar never below the warn bar.
func (c Config) validateAgentsBars() error {
	if c.AgentsWarnDays < 1 || c.AgentsMaxDays < 1 {
		return invalidf("agents thresholds must be >= 1, got warn=%d max=%d", c.AgentsWarnDays, c.AgentsMaxDays)
	}

	if c.AgentsMaxDays < c.AgentsWarnDays {
		return invalidf(
			"agents_max_days (%d) must be >= agents_warn_days (%d) — the FAIL bar cannot sit below the WARN bar",
			c.AgentsMaxDays,
			c.AgentsWarnDays,
		)
	}

	return nil
}

// validateStubChecksums keeps every digest an md5 (32 hex) or sha256
// (64 hex) string.
func (c Config) validateStubChecksums() error {
	for _, sum := range c.StubChecksums {
		if len(sum) != 32 && len(sum) != 64 {
			return invalidf("stub_checksums entries must be md5 (32 hex) or sha256 (64 hex) digests, got %q", sum)
		}
	}

	return nil
}

// validateCitationExts keeps every extension dot-prefixed.
func (c Config) validateCitationExts() error {
	for _, ext := range c.CitationExts {
		if len(ext) < 2 || ext[0] != '.' {
			return invalidf("citation_exts entries must be dot-prefixed extensions, got %q", ext)
		}
	}

	return nil
}
