// Package docsstub fails a repo whose domain-language doc is still the
// bare DDD template.
//
// The 2026-11-21 fleet census (~/projects/docs research/
// 2026-11-21_domain-language-commonalities.html) found 38 repos whose
// docs/DOMAIN_LANGUAGE.md is byte-identical to the unfilled scaffold —
// a doc that looks present but documents nothing. The class rebuilds
// every time a scaffolded repo ships without its glossary pass, so this
// gate pins the bare-template digests and fails on a checksum match.
//
// Rollout contract: findings WARN and exit 0 until the repo flips
// stub_enforce (the fleet-wide flip rides a TODO row one week after
// ship, the samber-linter ratchet doctrine). Exempt files
// (stub_exempts) cover honest stubs — the template host repo and
// utility repos whose docs deliberately say "no domain to document"
// in their own words rather than the template's.
package docsstub

import (
	"crypto/md5" //nolint:gosec // checksum identity, not security
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// Settings is the gate's config-driven knobs. Every field falls back to
// the shipped default when the repo's .docs-gates.json is absent or
// silent, so zero-value Settings never reach the gate logic.
type Settings struct {
	Files     []string // repo-relative corpus to checksum
	Checksums []string // bare-template digests (md5 or sha256 hex); a match is a stub
	Exempts   []string // repo-relative files exempt from the finding
	Enforce   bool     // false = findings WARN and exit 0; true = FAIL
}

// SettingsFromConfig projects the shared config onto the gate's knobs,
// keeping the shipped defaults wherever the config is silent.
func SettingsFromConfig(cfg config.Config) Settings {
	return Settings{
		Files:     config.OrSlice(cfg.StubFiles, config.Defaults().StubFiles),
		Checksums: config.OrSlice(cfg.StubChecksums, config.Defaults().StubChecksums),
		Exempts:   cfg.StubExempts,
		Enforce:   cfg.StubEnforce,
	}
}

// Usage is the `docs-gates stub --help` text — the verb's contract in
// one place, printed by the dispatcher before the parser runs.
const Usage = `docs-stub — fails a repo whose domain-language doc is the bare template

usage: docs-stub [repo-root]

No flags: the corpus (stub_files), the pinned bare-template digests
(stub_checksums), exemptions (stub_exempts) and the rollout switch
(stub_enforce) live in the repo's .docs-gates.json. A file whose md5 or
sha256 matches a pinned digest is an unfilled stub: WARN + exit 0 while
stub_enforce is false, FAIL + exit 1 once flipped. A missing corpus file
is a SKIP note, not a finding (repos without the doc are the freshness
gate's problem, not this one's).

Exit codes: 0 clean or warn-mode findings, 1 enforced stub findings, 2 usage error.
`

// Run is the CLI entry (repo-root positional).
func Run(args []string) int {
	flags, root, err := gatecli.Parse(args)
	if err != nil {
		return gatecli.FatalUsage("stub", err)
	}

	_ = flags // no flags; the positional root is the only input

	cfg, err := gatecli.LoadConfig(root)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	settings := SettingsFromConfig(cfg)

	findings, skips, err := Scan(root, settings)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	for _, skip := range skips {
		gatecli.Outf("SKIP %s\n", skip)
	}

	for _, finding := range findings {
		if settings.Enforce {
			gatecli.Outf("FAIL %s is the bare template (%s) — write the glossary or an honest stub\n", finding.Path, finding.Digest)
		} else {
			gatecli.Outf("WARN %s is the bare template (%s) — write the glossary or an honest stub (stub_enforce=false, warn mode)\n", finding.Path, finding.Digest)
		}
	}

	if len(findings) == 0 {
		gatecli.Outf("OK no bare-template stubs in %d file(s)\n", len(settings.Files))

		return gatecli.ExitClean
	}

	if settings.Enforce {
		return gatecli.ExitFindings
	}

	return gatecli.ExitClean
}

// Finding is one stub match: the repo-relative path and the digest that
// pinned it (short form for the verdict line).
type Finding struct {
	Path   string
	Digest string
}

// Scan checksums the corpus and returns the stub findings plus the
// human-readable skip notes. Missing corpus files are skips; exempt
// files are skips; both keep the gate honest about what it saw.
func Scan(root string, settings Settings) ([]Finding, []string, error) {
	var findings []Finding

	var skips []string

	for _, rel := range settings.Files {
		if slices.Contains(settings.Exempts, rel) {
			skips = append(skips, rel+" (exempt)")

			continue
		}

		body, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			if os.IsNotExist(err) {
				skips = append(skips, rel+" (missing)")

				continue
			}

			return nil, nil, fmt.Errorf("reading %s: %w", rel, err)
		}

		md5Sum := md5.Sum(body) //nolint:gosec // checksum identity, not security
		shaSum := sha256.Sum256(body)
		digests := map[string]string{
			hex.EncodeToString(md5Sum[:]): "md5",
			hex.EncodeToString(shaSum[:]): "sha256",
		}

		for digest, kind := range digests {
			if slices.Contains(settings.Checksums, digest) {
				findings = append(findings, Finding{Path: rel, Digest: kind})

				break
			}
		}
	}

	return findings, skips, nil
}
