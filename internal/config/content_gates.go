package config

// The three content gates (docs-stub, docs-citations, docs-stamps) and
// their config surface: the shared glossary corpus, the census-pinned
// bare-template digests, the citation extension list, and the stamp age
// bars. One file so the content-gate contract reads as one table.

// DefaultDomainLanguageFile is the three content gates' shared default
// corpus: the domain-language glossary every LarsArtmann repo carries
// under docs/. One name so the stub/citations/stamps Settings
// projections cannot drift apart.
const DefaultDomainLanguageFile = "docs/DOMAIN_LANGUAGE.md"

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

// validateContentGates enforces the content-gate value contract: stamp
// bars are ordered, stub digests are md5/sha256 hex, citation
// extensions are dot-prefixed.
func (c Config) validateContentGates() error {
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

	for _, sum := range c.StubChecksums {
		if len(sum) != 32 && len(sum) != 64 {
			return invalidf("stub_checksums entries must be md5 (32 hex) or sha256 (64 hex) digests, got %q", sum)
		}
	}

	for _, ext := range c.CitationExts {
		if len(ext) < 2 || ext[0] != '.' {
			return invalidf("citation_exts entries must be dot-prefixed extensions, got %q", ext)
		}
	}

	return nil
}
