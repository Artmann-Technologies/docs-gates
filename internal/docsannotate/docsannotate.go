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
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

var (
	itemStartRe = regexp.MustCompile(`^\s*(?:\d+[.)]|[-*+])\s+\S`)
	rowStartRe  = regexp.MustCompile(`^\|\s*(?:~~\d+~~|\d+)\s*\|`)
	sepRowRe    = regexp.MustCompile(`^\|[\s:|-]+\|\s*$`)
	headingRe   = regexp.MustCompile(`^#{1,6}\s`)
	fenceRe     = regexp.MustCompile(`^\s*` + "```")
	strikeRe    = regexp.MustCompile(`(?s)~~.*?~~`)
	// A `~~` inside inline code is a literal (e.g. the completeness gate's
	// own grep over archived reports), not a strike. Neutralise the tildes
	// inside code spans before strike detection.
	inlineCodeRe = regexp.MustCompile("`[^`]*`")
	markerRe     = regexp.MustCompile(
		`(?i)\b(?:done|resolved|superseded|fixed|adopted|decided|verified|` +
			`won'?t implement|not-do|duplicate|struck|moot|no longer)\b`,
	)
	openPhrases = regexp.MustCompile(
		`(?i)(?:\bstill (?:open|stale|owed|outstanding)|\bremains open|\bnot yet\b|\bowes\b|\bowed\b|\bnot deployed yet\b)`,
	)
	pathTokenRe = regexp.MustCompile("`([A-Za-z0-9_./-]+)`")
	pathExtRe   = regexp.MustCompile(`\.(?:nix|py|sh|md|json|yml|yaml|html|js|toml|service|timer|conf)$`)
	doneAtRe    = regexp.MustCompile(`(?i)done at\s*` + "`" + `([0-9a-f]{7,40})` + "`")
	spacingRe   = regexp.MustCompile(`[A-Za-z0-9]` + "`" + `[A-Za-z0-9]`)
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

const (
	// D: only a verdict that LEADS with the phrase is a contradiction.
	openWindow = 30
	// F opt-out markers.
	formatAllow     = "<!-- format:allow -->"
	hashFormatAllow = "# format:allow"

	// git's "command not found" status; sandboxes (nix flake check) ship
	// no git, and the hash/tag rules warn-skip on it.
	rcNoGit = 127

	// Display truncations (finding snippets, INFO leads).
	snippetMax = 110
	leadMax    = 60
	spacingMax = 90
)

// Finding is one gate finding (or INFO entry).
type Finding struct {
	Rule   string
	Path   string
	Line   int
	Detail string
	Text   string
}

type gitRunner struct {
	root    string
	missing bool
}

func (g gitRunner) run(args ...string) (string, int) {
	if g.missing {
		return "", rcNoGit
	}

	//nolint:gosec // git -C this repo; every subcommand is a literal from this package
	cmd := exec.CommandContext(context.Background(), "git", append([]string{"-C", g.root}, args...)...)

	out, err := cmd.Output()
	if err != nil {
		if ee, ok := errors.AsType[*exec.ExitError](err); ok {
			return string(out), ee.ExitCode()
		}
		// Sandboxes (nix flake check) ship no git; treat as "not a
		// checkout" so the hash/tag rules warn-skip instead of crashing.
		return "", rcNoGit
	}

	return string(out), 0
}

func isCheckout(root string) (gitRunner, bool) {
	g := gitRunner{root: root}

	out, rc := g.run("rev-parse", "--git-dir")
	if rc == rcNoGit {
		g.missing = true

		return g, false
	}

	return g, rc == 0 && strings.TrimSpace(out) != ""
}

// MaskCode neutralises `~~` inside inline code spans (literal, not a strike).
func MaskCode(text string) string {
	return inlineCodeRe.ReplaceAllStringFunc(text, func(m string) string {
		return strings.ReplaceAll(m, "~~", "--")
	})
}

// VerdictText is the item text with every struck span removed — the
// verdict's own words.
func VerdictText(item string) string {
	return strikeRe.ReplaceAllString(item, " ")
}

type block struct {
	start  int
	text   string
	isItem bool
}

// ExtractItems returns (1-based start line, text, is_item) for every
// checkable block. Numbered/bulleted items and numbered table rows take
// their continuation lines with them (multi-line strikes are the norm
// here). A line holding a strike that no item owns becomes a standalone
// block, so prose corrections are checked too.
func ExtractItems(lines []string) []block {
	var items []block

	covered := map[int]bool{}

	var current []string

	currentStart := 0
	currentIsItem := true
	fenced := false

	flush := func() {
		if len(current) > 0 {
			items = append(items, block{currentStart, strings.TrimSpace(strings.Join(current, " ")), currentIsItem})
			current = nil
		}
	}

	for idx, raw := range lines {
		num := idx + 1

		if fenceRe.MatchString(raw) {
			fenced = !fenced

			flush()

			continue
		}

		if fenced {
			continue
		}

		if headingRe.MatchString(raw) {
			flush()

			continue
		}

		if strings.TrimSpace(raw) == "" {
			flush()

			continue
		}

		if itemStartRe.MatchString(raw) || rowStartRe.MatchString(raw) {
			flush()

			currentStart = num
			currentIsItem = true
			current = []string{MaskCode(strings.TrimSpace(raw))}
			covered[num] = true

			continue
		}

		if sepRowRe.MatchString(raw) {
			flush()

			continue
		}

		if len(current) > 0 {
			current = append(current, MaskCode(strings.TrimSpace(raw)))
			covered[num] = true

			continue
		}
		// Unattached continuation or a bare prose line.
		if !strings.HasPrefix(raw, " ") {
			flush()
		}
	}

	flush()

	items = append(items, strayStrikes(lines, covered)...)

	sort.SliceStable(items, func(i, j int) bool { return items[i].start < items[j].start })

	return items
}

// strayStrikes owns the lines no block claimed that still carry a strike, so
// bare prose corrections are checked too.
func strayStrikes(lines []string, covered map[int]bool) []block {
	var strikes []block

	for idx, raw := range lines {
		num := idx + 1
		if covered[num] || !strings.Contains(MaskCode(raw), "~~") {
			continue
		}

		if headingRe.MatchString(raw) || fenceRe.MatchString(raw) {
			continue
		}

		strikes = append(strikes, block{num, MaskCode(strings.TrimSpace(raw)), false})
	}

	return strikes
}

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

// CheckSpacing: F. No letter/digit may touch a backtick anywhere in the
// docs the config's rule_f_extensions select. Disabled entirely when the
// caller passes rule-F off (Run checks the toggle before calling).
func CheckSpacing(root string, g gitRunner, hasGit bool, extensions []string) []Finding {
	walkFindings, docs := collectDocs(root, g, hasGit, extensions)
	findings := append([]Finding{}, walkFindings...)

	for _, path := range docs {
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			continue //nolint:erraudit // per-file skip; a doc under root always has a relative path
		}

		data, err := os.ReadFile(path)
		if err != nil {
			continue //nolint:erraudit // per-file skip keeps the gate alive; a racing delete must not fail it
		}

		findings = append(findings, spacingFindings(rel, data)...)
	}

	return findings
}

// spacingFindings scans one file's lines for the rule-F adjacency signature.
func spacingFindings(rel string, data []byte) []Finding {
	var findings []Finding

	for i, line := range strings.Split(string(data), "\n") {
		if spacingRe.MatchString(line) &&
			!strings.Contains(line, formatAllow) &&
			!strings.Contains(line, hashFormatAllow) {
			findings = append(findings, Finding{
				"F/spacing", rel, i + 1,
				"space stripped around inline code", gatecli.Truncate(strings.TrimSpace(line), spacingMax),
			})
		}
	}

	return findings
}

// crossCellFindings enforces row uniformity (rule G): in a
// markdown TABLE row, every strike must open AND close inside one cell.
// A strike spanning cells (the 2026-09-30-era ID-cell-only class:
// `| ~~C18   | Assistant | text~~ done (…)`) renders as a strike over the
// ID column and defies the per-cell convention every other row follows.
// Cell parity is the detector: an odd ~~ count in a cell (after masking
// inline-code literals) means the pair straddles a boundary.
func crossCellFindings(rel string, lines []string) []Finding {
	var findings []Finding

	for i, line := range lines {
		if !strings.HasPrefix(line, "|") {
			continue
		}

		for cell := range strings.SplitSeq(MaskCode(line), "|") {
			if strings.Count(cell, "~~")%2 == 1 {
				findings = append(findings, Finding{
					"G/uniform",
					rel,
					i + 1,
					"strike crosses a cell boundary — wrap each cell's text independently (the ID-cell-only strike class)",
					gatecli.Truncate(strings.TrimSpace(line), snippetMax),
				})

				break // one finding per row is enough to fix the row
			}
		}
	}

	return findings
}

// collectDocs assembles the file list rule F scans: tracked files via git
// when this is a checkout, a recursive walk otherwise (plus a finding when
// the walk itself fails — an empty list must never read as a false clean).
// The extension list comes from the config (default .md/.sh/.nix).
func collectDocs(root string, g gitRunner, hasGit bool, extensions []string) ([]Finding, []string) {
	if hasGit {
		return nil, docsViaGit(root, g, extensions)
	}

	return docsByWalk(root, extensions)
}

func docsViaGit(root string, g gitRunner, extensions []string) []string {
	var docs []string

	// Rule F's home is .md (dprint's markdown plugin owns the corruption
	// class). The same adjacency signature is cheap to scan in the files
	// OTHER formatters own — shell scripts (no formatter today) and
	// .nix comments (nixfmt; widened from bare flake.nix to every
	// tracked .nix when flake.nix was split into flake/, 2026-10-04) —
	// so the class resurfacing there fails loudly too. Opt-out marker in
	// those files is a comment: `# format:allow` (# is their comment
	// syntax).
	for _, ext := range extensions {
		out, _ := g.run("ls-files", "*"+ext)
		for n := range strings.SplitSeq(out, "\n") {
			if strings.TrimSpace(n) != "" {
				docs = append(docs, filepath.Join(root, n))
			}
		}
	}

	return docs
}

func docsByWalk(root string, extensions []string) ([]Finding, []string) {
	var (
		findings []Finding
		docs     []string
	)

	// Recursive walk, not glob: Go's filepath.Glob has no real
	// double-star — a `**` segment matches exactly ONE path level —
	// so glob patterns silently missed hosts/pbx/*.nix and every
	// deeper file in this branch (found 2026-10-05). Walking yields
	// the same file set the git branch's ls-files produces.
	exts := map[string]bool{}
	for _, ext := range extensions {
		exts[ext] = true
	}

	if walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// An unreadable entry ABORTS the walk: skipping it would
			// shrink the file list silently — the false-clean class
			// this repo's gates exist to kill. Returning the error is
			// the only way filepath.WalkDir reports anything (it
			// swallows errors the closure answers with nil — the
			// F/walk branch below was unreachable dead code before
			// this return, proven by TestDocsByWalkFailureIsAFinding).
			return err
		}

		if d.IsDir() || !exts[filepath.Ext(path)] {
			return nil
		}

		docs = append(docs, path)

		return nil
	}); walkErr != nil {
		// A failed walk yields an EMPTY or PARTIAL docs list, which
		// would read as a false CLEAN from the outside — surface it as
		// a finding instead. Text stays non-empty so --verbose renders
		// a remedy line, not a bare indent (the walk-failure fix's own remedy rule).
		findings = append(findings, Finding{
			Rule:   "F/walk",
			Path:   ".",
			Detail: "walk failed: " + walkErr.Error(),
			Text:   "the file list is incomplete — fix the unreadable entry and re-run",
		})
	}

	return findings, docs
}

// Run is the CLI entry: [--check] [--verbose] [repo-root]. The annotate
// corpus (report dir + its archived subdir + the configured extra globs)
// and the rule-F toggle/extensions come from the repo's .docs-gates.json
// when present.
// parseAnnotateMax splits --annotate-max=N off the argument list BEFORE
// gatecli.Parse (a valued flag would die there as an unknown boolean). The
// parsed pin rides out; the remaining argv is the Parse input. Strict (-1)
// when absent; a malformed value FATALs with exit 2 via the ok=false +
// fatal=true pair.
func parseAnnotateMax(args []string) (int, []string) {
	annotateMax := -1

	rest := make([]string, 0, len(args))

	for _, a := range args {
		if v, ok := strings.CutPrefix(a, "--annotate-max="); ok {
			if n, parseErr := strconv.Atoi(v); parseErr == nil && n >= 0 {
				annotateMax = n

				continue
			}

			gatecli.Errf("FATAL: --annotate-max wants a non-negative integer, got %q\n", v)

			return -1, nil
		}

		rest = append(rest, a)
	}

	return annotateMax, rest
}

func Run(argv []string) int {
	// The adoption ratchet (fleet friction, 2026-11-03 pilots): repos whose
	// historical strikes predate the verdict-vocabulary rules seed
	// annotate_max with the measured finding count and drain down — red
	// only PAST the pin (a regression), mirroring the freshness audit
	// ratchet. Strict (-1, the default) is the pre-extraction behavior.
	annotateMax, argv := parseAnnotateMax(argv)
	if argv == nil {
		return gatecli.ExitUsage
	}

	flags, root, err := gatecli.Parse(argv, "--check", "--verbose")
	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	cfg, warnings, err := config.Load(root)
	for _, w := range warnings {
		gatecli.Errf("config: %s\n", w)
	}

	if err != nil {
		gatecli.Errf("FATAL: %v\n", err)

		return gatecli.ExitUsage
	}

	if cfg.AnnotateMax >= 0 && annotateMax < 0 {
		annotateMax = cfg.AnnotateMax
	}

	docGlobs := append([]string{
		filepath.Join(cfg.ReportDir, "*.md"),
		filepath.Join(cfg.ReportDir, cfg.ArchivedSubdir, "*.md"),
	}, cfg.AnnotateExtraGlobs...)

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

	if cfg.RuleF {
		findings = append(findings, CheckSpacing(root, g, hasGit, cfg.RuleFExtensions)...)
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
