package docsfreshness

// The pre-flight refusal machinery: the on-disk collision gate (structural
// headings model as a second copy, content refuses), and the missing
// --replace-target refusals that name WHERE a line actually lives.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// headingLineRe matches a markdown ATX heading candidate line
// (`# Example` .. `###### Example`).
var headingLineRe = regexp.MustCompile(`^#{1,6}\s`)

// onDiskMatch pairs one candidate line found verbatim in the corpus
// with the disk line that matched — the WHERE beside the WHAT.
type onDiskMatch struct {
	candidate string
	diskLine  string
}

// onDiskMatches pairs each candidate line already present in the
// living-docs corpus (lowercased substring match, blank lines never
// count) with the first disk line containing it. Shared by the refusal
// and the --duplicates accounting.
func onDiskMatches(candidate, baseLines []string) []onDiskMatch {
	var matches []onDiskMatch

	for _, line := range candidate {
		trimmed := strings.ToLower(strings.TrimSpace(line))
		if trimmed == "" {
			continue
		}

		for _, disk := range baseLines {
			if strings.Contains(strings.ToLower(disk), trimmed) {
				matches = append(matches, onDiskMatch{candidate: line, diskLine: disk})

				break
			}
		}
	}

	return matches
}

// handleCollisions classifies the on-disk collisions and answers the
// invocation: a repeated structural heading models as a second copy
// (the implicit --duplicates — a CHANGELOG landing batch legitimately
// re-includes `### Changed`, and refusing it would block every
// heading-led batch), a duplicated CONTENT line still refuses (the
// additions-only model needs the PRE-edit tree), and an explicit
// --duplicates models everything. Returns true when refused.
func handleCollisions(candidate, remaining []string, allowDuplicates bool) bool {
	collisions := onDiskMatches(candidate, remaining)
	if len(collisions) == 0 {
		return false
	}

	var content []onDiskMatch

	for _, m := range collisions {
		if allowDuplicates || structuralRepeat(m) {
			// The operator-side dedup workaround (grep -Fxvf) silently DROPS
			// the +1 DF push of a line arriving as a second copy — exactly the
			// push a tail-ledger block makes. --duplicates (explicit or
			// structural) asserts "another copy is about to land", so the
			// delta models the copy instead of refusing the question.
			gatecli.Outf("  second copy: %-52s stays modeled — every token on it gains +1 DF\n",
				gatecli.Truncate(m.candidate, reasonMax))

			continue
		}

		content = append(content, m)
	}

	if len(content) == 0 {
		return false
	}

	printAlreadyOnDisk(content)

	return true
}

// structuralRepeat reports whether an on-disk collision is a repeated
// structural heading (`### Changed`, `## [Unreleased]`, …) rather than
// a duplicated content line: sections repeat across landings, content
// must not.
func structuralRepeat(m onDiskMatch) bool {
	return headingLineRe.MatchString(m.candidate)
}

// printAlreadyOnDisk explains the content-line refusal.
func printAlreadyOnDisk(content []onDiskMatch) {
	gatecli.Errln("docs-freshness --pre-flight: candidate line(s) ALREADY on disk — the additions-only model")
	gatecli.Errln("  needs the PRE-edit tree; once landed, baseline and modified corpus coincide and a would-be")
	gatecli.Errln("  flip hides as SAFE. Run pre-flight BEFORE writing (or against a stashed/HEAD tree).")
	gatecli.Errln("  If the line is about to land as a deliberate SECOND copy (a tail-ledger block), rerun with")
	gatecli.Errln("  --duplicates to model its +1 DF push instead of refusing. Collisions (candidate ← disk):")

	for _, m := range content {
		gatecli.Errf("    %s\n", gatecli.Truncate(m.candidate, reasonMax))
		gatecli.Errf("      ← %s\n", gatecli.Truncate(m.diskLine, reasonMax))
	}
}

// outsideHit is a missing --replace target located in a repo file the
// corpus does not cover (AGENTS.md, README.md, …).
type outsideHit struct {
	line string
	file string
}

// refuseMissingReplaces reports the missing --replace targets on both
// classes: targets found OUTSIDE the corpus are named with their file
// (the honest WHERE — a rewrite landing there moves no corpus DF), the
// genuinely-not-anywhere remainder keeps the bare not-found list.
func refuseMissingReplaces(root string, missing []string, settings Settings) {
	named := replaceTargetsOutsideCorpus(root, missing, settings)

	if len(named) > 0 {
		gatecli.Errln("docs-freshness --pre-flight: --replace target(s) found OUTSIDE the living-docs corpus:")

		for _, h := range named {
			gatecli.Errf("    %s\n", gatecli.Truncate(h.line, reasonMax))
			gatecli.Errf(
				"      → lives in %s, which the corpus (%s) does not cover — a rewrite landing there moves "+
					"no corpus DF; drop --replace for it (skip pre-flight entirely when the whole batch targets %s)\n",
				h.file,
				strings.Join(settings.livingDocs(), ", "),
				h.file,
			)
		}
	}

	rest := missingOutside(missing, named)
	if len(rest) == 0 {
		return
	}

	gatecli.Errln("docs-freshness --pre-flight: --replace target(s) not found in the living docs:")

	for _, m := range rest {
		gatecli.Errf("    %s\n", gatecli.Truncate(m, reasonMax))
	}
}

// replaceTargetsOutsideCorpus locates missing replace targets in
// root-level markdown files outside the corpus — the honest refusal
// names WHERE the line lives instead of a bare not-found. Bounded to
// the repo root's own *.md files: the corpus is root-level by contract.
func replaceTargetsOutsideCorpus(root string, missing []string, settings Settings) []outsideHit {
	outside, err := outsideMarkdownFiles(root, settings)
	if err != nil {
		return nil
	}

	var hits []outsideHit

	for _, target := range missing {
		if file := findTargetInFiles(root, outside, target); file != "" {
			hits = append(hits, outsideHit{line: target, file: file})
		}
	}

	return hits
}

// outsideMarkdownFiles lists the root-level *.md files the corpus does
// not cover (the candidates an out-of-universe edit targets).
func outsideMarkdownFiles(root string, settings Settings) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	corpus := make(map[string]bool, len(settings.livingDocs()))
	for _, name := range settings.livingDocs() {
		corpus[strings.ToLower(name)] = true
	}

	var outside []string

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(name), ".md") || corpus[strings.ToLower(name)] {
			continue
		}

		outside = append(outside, name)
	}

	return outside, nil
}

// findTargetInFiles returns the first file whose content contains the
// target line (lowercased substring match); "" when none does.
func findTargetInFiles(root string, files []string, target string) string {
	trimmed := strings.ToLower(strings.TrimSpace(target))

	for _, name := range files {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			continue
		}

		for line := range strings.SplitSeq(string(data), "\n") {
			if strings.Contains(strings.ToLower(line), trimmed) {
				return name
			}
		}
	}

	return ""
}

// missingOutside keeps the missing targets the outside scan did NOT
// name — the genuinely-not-anywhere class for the bare not-found list.
func missingOutside(missing []string, hits []outsideHit) []string {
	named := make(map[string]bool, len(hits))
	for _, h := range hits {
		named[h.line] = true
	}

	out := make([]string, 0, len(missing))

	for _, m := range missing {
		if !named[m] {
			out = append(out, m)
		}
	}

	return out
}
