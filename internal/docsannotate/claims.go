package docsannotate

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

var (
	pathTokenRe = regexp.MustCompile("`([A-Za-z0-9_./-]+)`")
	pathExtRe   = regexp.MustCompile(`\.(?:nix|py|sh|md|json|yml|yaml|html|js|toml|service|timer|conf)$`)
	doneAtRe    = regexp.MustCompile(`(?i)done at\s*` + "`" + `([0-9a-f]{7,40})` + "`")
)

// refClaim is one E-rule: a declarative, mechanically checkable artifact
// claim. Add a rule here when a doc claim about an artifact turns out false.
type refClaim struct {
	re   *regexp.Regexp
	kind string
}

//nolint:gochecknoglobals // compiled-once static config; immutable data tables, not mutable state
var (
	repoDirs = []string{
		"tests/",
		"hosts/",
		"docs/",
		"installer/",
		"cloud-init/",
		"telnyx/",
		"backups/",
	}

	refClaims = []refClaim{
		{regexp.MustCompile(`\btags?\s+` + "`" + `([^` + "`" + `]+)` + "`"), "tag-exists"},
		{regexp.MustCompile(`(?:rollback|release|git) tags?\s+` + "`" + `([^` + "`" + `]+)` + "`"), "tag-exists"},
		{
			regexp.MustCompile(`(?:trashed|removes?|removed|deleted|cleans? up)\s+` + "`" + `([^` + "`" + `]+)` + "`"),
			"path-gone",
		},
	}
)

// D: only a verdict that LEADS with the phrase is a contradiction.
const openWindow = 30

// Display truncations (INFO leads).
const leadMax = 60

// verifyHashClaims checks B: every `done at <sha>` must resolve to a commit
// in this repo (skipped outside a checkout).
func verifyHashClaims(
	findings []Finding, tail, rel string, b block, snippet string, g gitRunner, hasGit bool,
) []Finding {
	if !hasGit {
		return findings
	}

	for _, m := range doneAtRe.FindAllStringSubmatch(tail, -1) {
		sha := m[1]

		if _, rc := g.run("cat-file", "-e", sha+"^{commit}"); rc != 0 {
			findings = append(findings, Finding{
				"B/hash", rel, b.start,
				"`done at " + sha + "` does not resolve to a commit here", snippet,
			})
		}
	}

	return findings
}

// verifyPathTokens checks C: a repo-relative path cited in the verdict's own
// text must exist on disk. Paths inside the struck text (the original claim)
// are ignored on purpose - the whole point is that they may have been wrong.
func verifyPathTokens(
	findings []Finding, tail, rel string, b block, snippet, root string,
) []Finding {
	for _, m := range pathTokenRe.FindAllStringSubmatch(tail, -1) {
		token := m[1]
		if !pathExtRe.MatchString(token) && !startsWithAny(token, repoDirs) {
			continue
		}

		if !strings.Contains(token, "/") || strings.HasPrefix(token, "/") ||
			strings.HasPrefix(token, "~") || strings.HasPrefix(token, "http") {
			continue
		}

		if strings.ContainsAny(token, "*{") {
			continue
		}

		if _, err := os.Lstat(filepath.Join(root, token)); err != nil {
			findings = append(findings, Finding{
				"C/path", rel, b.start,
				"verdict cites missing `" + token + "`", snippet,
			})
		}
	}

	return findings
}

// flagOpenVerdict reports D (INFO, never a failure): a verdict that itself
// leads with an unfinished phrase.
func flagOpenVerdict(tail, rel string, b block, snippet string) (Finding, bool) {
	marker := markerRe.FindStringIndex(tail)
	if marker == nil {
		return Finding{}, false
	}

	lead := tail[marker[1]:min(marker[1]+openWindow, len(tail))]
	if openPhrases.MatchString(lead) {
		return Finding{
			"D/verdict-leads-open",
			rel,
			b.start,
			"verdict leads with an unfinished phrase: " + gatecli.Truncate(
				strings.TrimSpace(lead),
				leadMax,
			),
			snippet,
		}, true
	}

	return Finding{}, false
}

// verifyRefClaims checks E: declarative artifact claims — a cited tag must
// exist, something claimed trashed/removed must be gone.
func verifyRefClaims(
	findings []Finding, tail, rel string, b block, snippet, root string, g gitRunner, hasGit bool,
) []Finding {
	claimed := map[[2]string]bool{}

	for _, claim := range refClaims {
		for _, m := range claim.re.FindAllStringSubmatch(tail, -1) {
			target := strings.TrimSpace(m[1])
			if target == "" || strings.Contains(target, " ") ||
				strings.HasPrefix(target, "/") || strings.HasPrefix(target, "~") ||
				strings.HasPrefix(target, "http") {
				continue
			}

			key := [2]string{claim.kind, target}
			if claimed[key] {
				continue
			}

			claimed[key] = true

			switch claim.kind {
			case "tag-exists":
				findings = checkTagClaim(findings, target, rel, b, snippet, g, hasGit)
			case "path-gone":
				findings = checkPathGoneClaim(findings, target, rel, b, snippet, root)
			}
		}
	}

	return findings
}

func checkTagClaim(
	findings []Finding, target, rel string, b block, snippet string, g gitRunner, hasGit bool,
) []Finding {
	if !hasGit {
		return findings
	}

	out, rc := g.run("tag", "-l", target)
	if rc != 0 {
		return findings // cannot verify here; do not claim a missing tag from a failed git
	}

	if strings.TrimSpace(out) == "" {
		findings = append(findings, Finding{
			"E/ref", rel, b.start,
			"claims tag `" + target + "`, which does not exist", snippet,
		})
	}

	return findings
}

func checkPathGoneClaim(
	findings []Finding, target, rel string, b block, snippet, root string,
) []Finding {
	_, statErr := os.Lstat(filepath.Join(root, target))
	if statErr != nil && !strings.Contains(target, "/") && !pathExtRe.MatchString(target) {
		return findings // relative basename that is also gone: nothing checkable
	}

	if statErr == nil {
		findings = append(findings, Finding{
			"E/ref", rel, b.start,
			"claims `" + target + "` was removed, but it still exists", snippet,
		})
	}

	return findings
}

func startsWithAny(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}

	return false
}
