package docsfreshness

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// candidatesError is the package's static error type: readCandidates
// reports stdin failures as *candidatesError with the cause reachable
// through Unwrap.
type candidatesError struct {
	msg string
	err error
}

func (e *candidatesError) Error() string { return e.msg + ": " + e.err.Error() }

func (e *candidatesError) Unwrap() error { return e.err }

// readCandidates reads the --pre-flight candidate lines from stdin,
// dropping blank lines.
func readCandidates() ([]string, *candidatesError) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, &candidatesError{msg: "reading candidate lines from stdin", err: err}
	}

	var candidate []string

	for line := range strings.SplitSeq(string(data), "\n") {
		if strings.TrimSpace(line) != "" {
			candidate = append(candidate, line)
		}
	}

	return candidate, nil
}

// LivingDocsLines reads the named living docs under root into the
// raw-line corpus: every line lowercased. The freshness harvest and the
// routing triage share this corpus shape — both token machineries (the
// GenericDF bar) are tuned against RAW lowercase lines, so there must be
// exactly one reader.
func LivingDocsLines(root string, names []string) []string {
	var lines []string

	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			continue //nolint:erraudit // a not-yet-existing living doc simply contributes no lines
		}

		for line := range strings.SplitSeq(string(data), "\n") {
			lines = append(lines, strings.ToLower(line))
		}
	}

	return lines
}

// Corpus is the prepared living-docs corpus: raw lowercased lines plus
// their squashed keys. The freshness harvest and the routing triage share
// this one preparation so the two gates cannot disagree about what
// "present in the living docs" means (the 2026-11-04 dedup: docsrouting
// grew its own tokenIn/squashContains/livingLineCount with a subtly
// different squash-arm length condition; both now ride this type).
type Corpus struct {
	lines    []string
	squashed []string
}

// NewCorpus prepares raw (already lowercased) lines for token matching.
func NewCorpus(lines []string) Corpus {
	return Corpus{lines: lines, squashed: squashLines(lines)}
}

// LivingDocsCorpus is LivingDocsLines prepared for token matching.
func LivingDocsCorpus(root string, names []string) Corpus {
	return NewCorpus(LivingDocsLines(root, names))
}

// Contains reports whether the token is present in the corpus: raw
// substring match first, then the hyphen/space-collapsed key when the KEY
// is long enough to carry it — the harvest's reference semantics.
func (c Corpus) Contains(token string) bool {
	return tokenPresent(token, c.lines, c.squashed)
}

// Count is the raw-line document frequency: how many lines contain the
// token as a substring. Genericity counts stay on RAW lines: GenericDF
// was tuned against raw substring counts, and squashed counting inflates
// them (words join across boundaries).
func (c Corpus) Count(token string) int {
	return tokenDFOver(c.lines)(token)
}

// loadLivingDocs reads the configured living docs into the comparison
// corpus: lowercased raw lines, their squashed keys, and the raw-line token
// document frequency.
func loadLivingDocs(root string, settings Settings) ([]string, []string, func(string) int) {
	livingLines := LivingDocsLines(root, settings.livingDocs())

	return livingLines, squashLines(livingLines), tokenDFOver(livingLines)
}

// tokenDFOver is the raw-line document-frequency counter: how many
// living-doc lines contain the token as a substring. Genericity counts
// stay on RAW lines: GenericDF = 8 was tuned 2026-09-22 against raw
// substring counts, and squashed counting inflates them (words join
// across boundaries) — measured as a false DRIFT on a resolved item the
// same day this landed.
func tokenDFOver(lines []string) func(string) int {
	return func(token string) int {
		n := 0

		for _, line := range lines {
			if strings.Contains(line, token) {
				n++
			}
		}

		return n
	}
}

// statusReports lists every replayable report: the configured report dir
// plus its archived subdir, oldest first — the corpus --audit-all and
// --pre-flight replay.
func statusReports(root string, settings Settings) []string {
	reports, _ := filepath.Glob(
		settings.reportGlobPath(root),
	) //nolint:erraudit // constant pattern cannot be malformed
	sort.Strings(reports)

	archived, _ := filepath.Glob(
		settings.archivedGlobPath(root),
	) //nolint:erraudit // constant pattern cannot be malformed
	sort.Strings(archived)

	return append(reports, archived...)
}
