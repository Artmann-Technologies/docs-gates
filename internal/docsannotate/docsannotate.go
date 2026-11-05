// Package docsannotate is the annotation-quality gate for historical docs
// (docs-health hardening, 2026-09-23).
//
// Historical reports are annotated non-destructively: a resolved item is
// struck through and carries a verdict marker (`done (evidence)`,
// `done at `sha“, `Won't implement — reason`). The 2026-09-23 audit found
// 6 of ~1078 markers were false — a strike claiming the release was
// republished while it was still STALE, one claiming a rollback tag existed
// while the repo had none — and nothing caught them. This check is that
// missing gate.
//
// Rules (each item = a numbered/bulleted list item, a numbered table row,
// or a standalone paragraph, continuation lines included):
//
//	A. EVIDENCE  - a struck ITEM must carry a verdict marker. A strike with
//	   no verdict is indistinguishable from a deletion (the #1 annotation
//	   failure). Narrative/status-line corrections are exempt: they are
//	   prose, not items, and their correction text IS the verdict.
//	B. HASH      - a hash cited as `done at `sha`` must resolve in this repo
//	   (git cat-file). Catches fabricated hashes. Skipped outside a checkout.
//	C. PATH      - a repo-relative path cited in a verdict's own text must
//	   exist on disk. Paths inside the struck text (the original claim) are
//	   ignored on purpose - the whole point is that they may have been wrong.
//	D. OPEN      - (INFO, never a failure) a verdict that itself leads with
//	   `still open`/`owed`/`not yet` is flagged for triage.
//	E. REF CLAIM - a declarative claim about a repo artifact that is
//	   mechanically checkable: a tag must exist in git tag; something the
//	   verdict says was `trashed`/`removed` must be gone.
//	F. SPACING  - inline code must keep the spaces around it. dprint's
//	   markdown formatter strips them when a code span contains an escaped
//	   backtick. A line that legitimately joins text to code opts out with
//	   an inline `<!-- format:allow -->` (or `# format:allow` in the
//	   non-markdown files rule F scans).
//
// Exit 0 = clean, 1 = findings, 2 = usage/environment error.
package docsannotate

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// Finding is one gate finding (or INFO entry).
type Finding struct {
	Rule   string
	Path   string
	Line   int
	Detail string
	Text   string
}

// Display truncations (finding snippets).
const snippetMax = 110

func checkFile(root, path string, g gitRunner, hasGit bool, infos *[]Finding) ([]Finding, int, error) {
	var findings []Finding

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}

	lines := strings.Split(string(data), "\n")

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return nil, 0, err
	}

	struck := 0

	for _, b := range ExtractItems(lines) {
		if !strings.Contains(b.text, "~~") {
			continue
		}

		struck++
		tail := VerdictText(b.text)
		snippet := gatecli.Truncate(b.text, snippetMax)

		if b.isItem && !markerRe.MatchString(tail) {
			findings = append(
				findings,
				Finding{"A/evidence", rel, b.start, "struck item with no verdict marker", snippet},
			)

			continue
		}

		findings = verifyHashClaims(findings, tail, rel, b, snippet, g, hasGit)
		findings = verifyPathTokens(findings, tail, rel, b, snippet, root)
		findings = verifyRefClaims(findings, tail, rel, b, snippet, root, g, hasGit)

		if info, open := flagOpenVerdict(tail, rel, b, snippet); open {
			*infos = append(*infos, info)
		}
	}

	if strings.EqualFold(filepath.Ext(path), ".md") {
		findings = append(findings, crossCellFindings(rel, lines)...)
	}

	return findings, struck, nil
}

// Usage is the `docs-annotate --help` text — the verb's contract in one
// place, printed by the dispatcher before the parser runs. usage_test.go
// pins it against the parser's flag vocabulary.
const Usage = `docs-annotate — struck items must carry honest, verifiable verdicts

usage: docs-annotate [repo-root] [--annotate-max=N] [--check] [--verbose]

  --annotate-max=N  ratchet pin: tolerated findings, red only past N
  --check           accepted no-op — checking is the gate's only mode
  --verbose         print the struck item text under each finding

The annotate corpus (report dir + archived subdir + extra globs) and the
rule-F toggle come from the repo's .docs-gates.json.

Exit codes: 0 clean, 1 findings, 2 usage error.
`

// allowedFlags is the boolean flag vocabulary Run accepts. Usage
// documents every entry; usage_test.go pins the pair together.
//
//nolint:gochecknoglobals // the fixed flag vocabulary, same class as the contract tables
var allowedFlags = []string{"--check", "--verbose"}

// parseAnnotateMax splits --annotate-max=N off the argument list BEFORE
// gatecli.Parse (a valued flag would die there as an unknown boolean). The
// parsed pin rides out; the remaining argv is the Parse input. Strict (-1)
// when absent; a malformed value FATALs with exit 2 via the badExit
// sentinel (nonzero only on the error path).
func parseAnnotateMax(args []string) (int, []string, int) {
	annotateMax, rest, err := gatecli.CutIntFlag(args, "annotate-max")
	if err != nil {
		return -1, nil, gatecli.FatalUsage("annotate", err)
	}

	return annotateMax, rest, 0
}

// Run is the CLI entry: [--check] [--verbose] [repo-root]. The annotate
// corpus (report dir + its archived subdir + the configured extra globs)
// and the rule-F toggle/extensions come from the repo's .docs-gates.json
// when present.
func Run(argv []string) int {
	// The adoption ratchet (fleet friction, 2026-11-03 pilots): repos whose
	// historical strikes predate the verdict-vocabulary rules seed
	// annotate_max with the measured finding count and drain down — red
	// only PAST the pin (a regression), mirroring the freshness audit
	// ratchet. Strict (-1, the default) is the pre-extraction behavior.
	annotateMax, argv, badExit := parseAnnotateMax(argv)
	if badExit != 0 {
		return badExit
	}

	flags, root, err := gatecli.Parse(argv, allowedFlags...)
	if err != nil {
		return gatecli.FatalUsage("annotate", err)
	}

	settings, err := loadSettings(root)
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	if settings.AnnotateMax >= 0 && annotateMax < 0 {
		annotateMax = settings.AnnotateMax
	}

	docGlobs := append([]string{
		filepath.Join(settings.dir(), "*.md"),
		filepath.Join(settings.dir(), settings.archivedSub(), "*.md"),
	}, settings.extraGlobs()...)

	files := collectDocFiles(root, docGlobs)

	g, hasGit := isCheckout(root)
	if !hasGit {
		gatecli.Errln("docs-annotate: no git checkout here - hash/tag rules skipped")
	}

	var (
		findings []Finding
		infos    []Finding
	)

	struckTotal := 0

	for _, path := range files {
		fileFindings, struck, err := checkFile(root, path, g, hasGit, &infos)
		if err != nil {
			continue //nolint:erraudit // an unreadable report is skipped; the rest of the corpus still gets gated
		}

		findings = append(findings, fileFindings...)
		struckTotal += struck
	}

	if settings.RuleF {
		findings = append(findings, CheckSpacing(root, g, hasGit, settings.ruleFExtensions())...)
	}

	printFindings(flags, infos, findings, len(files), struckTotal)

	return annotateVerdict(len(findings), annotateMax)
}

// annotateVerdict turns the finding count into the exit code under the
// adoption ratchet: tolerated pins print RATCHET and stay green, counts
// past a pin print REGRESSION, strict counts print the FAIL remedy line.
func annotateVerdict(count, annotateMax int) int {
	switch {
	case count == 0:
		gatecli.Outln("docs-annotate: CLEAN")

		return gatecli.ExitClean
	case annotateMax >= 0 && count <= annotateMax:
		gatecli.Outf(
			"docs-annotate: RATCHET — %d finding(s) tolerated (pinned --annotate-max=%d); red only past %d\n",
			count, annotateMax, annotateMax,
		)
		gatecli.Outln("drain them with the docs-health ANNOTATE ritual, then lower the pin")

		return gatecli.ExitClean
	case annotateMax >= 0:
		gatecli.Errf(
			"docs-annotate: REGRESSION — %d finding(s) past the pinned --annotate-max=%d. ",
			count, annotateMax,
		)
		gatecli.Errln("Drain them (docs-health ANNOTATE) or annotate them resolved, then re-pin.")

		return gatecli.ExitFindings
	default:
		gatecli.Errln("docs-annotate: FAIL - repair the verdict (docs-health ANNOTATE rules: " +
			"strike only on evidence, cite a real hash/path, never strike an open item).")

		return gatecli.ExitFindings
	}
}

func collectDocFiles(root string, patterns []string) []string {
	var files []string

	for _, pattern := range patterns {
		matches, _ := filepath.Glob(
			filepath.Join(root, pattern),
		) //nolint:erraudit // constant pattern cannot be malformed
		sort.Strings(matches)

		for _, m := range matches {
			if info, err := os.Stat(m); err == nil && info.Mode().IsRegular() {
				files = append(files, m)
			}
		}
	}

	return files
}

func printFindings(flags map[string]bool, infos, findings []Finding, files, struckTotal int) {
	for _, info := range infos {
		gatecli.Outf("INFO %s: %s:%d: %s\n", info.Rule, info.Path, info.Line, info.Detail)
	}

	for _, finding := range findings {
		gatecli.Outf("%s: %s:%d: %s\n", finding.Rule, finding.Path, finding.Line, finding.Detail)

		if flags["--verbose"] {
			gatecli.Outf("    %s\n", finding.Text)
		}
	}

	gatecli.Outf("docs-annotate: %d file(s), %d struck item(s), %d finding(s)\n",
		files, struckTotal, len(findings))
}
