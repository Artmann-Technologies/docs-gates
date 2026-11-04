package docsannotate

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

var spacingRe = regexp.MustCompile(`[A-Za-z0-9]` + "`" + `[A-Za-z0-9]`)

const (
	// F opt-out markers.
	formatAllow     = "<!-- format:allow -->"
	hashFormatAllow = "# format:allow"

	// Display truncations (F leads).
	spacingMax = 90
)

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
