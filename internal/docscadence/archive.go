package docscadence

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// archiveInventory summarizes the unarchived reports under root's
// reportDir (the archived/ subdir below it is filtered as a directory; a
// missing report dir is an empty inventory — test roots legitimately
// carry only the TODO file). Any other read error is fatal. The bare err
// return below is deliberate: the sole caller frames it as
// `FATAL: reading <reportDir>/: %v`, so wrapping here would duplicate
// the context (erraudit --max flags this site as context_loss; judged a
// false positive at this layer, 2026-11-04).
func archiveInventory(root string, today time.Time, reportDir string) (int, int, string, error) {
	entries, err := os.ReadDir(filepath.Join(root, reportDir))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, "", nil
		}

		return 0, 0, "", err
	}

	names := make([]string, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}

	count, age, oldest := summarize(names, today)

	return count, age, oldest, nil
}

// reportDateLen is the length of the YYYY-MM-DD prefix reportDate
// requires — derived from the layout so the two can never drift apart.
const reportDateLen = len("2006-01-02")

// hoursPerDay converts a time.Duration to whole days of age.
const hoursPerDay = 24

// summarize counts report filenames (a parseable YYYY-MM-DD prefix —
// anything else in docs/status/ is a stray and skipped) and returns the
// count, the age in days of the oldest (measured against today), and the
// oldest filename. A future date (clock skew) clamps to 0 so it can
// never read as overdue.
func summarize(names []string, today time.Time) (int, int, string) {
	count := 0
	oldestName := ""

	var oldest time.Time

	for _, name := range names {
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		date, ok := reportDate(name)
		if !ok {
			continue
		}

		count++

		if oldestName == "" || date.Before(oldest) {
			oldest = date
			oldestName = name
		}
	}

	if count == 0 {
		return 0, 0, ""
	}

	age := max(int(today.Sub(oldest).Hours()/hoursPerDay), 0)

	return count, age, oldestName
}

// reportDate parses the leading YYYY-MM-DD of a report filename
// (2026-10-24_11-35_name.md). A shorter or undated name is not a report.
func reportDate(name string) (time.Time, bool) {
	if len(name) < reportDateLen {
		return time.Time{}, false
	}

	date, err := time.Parse("2006-01-02", name[:reportDateLen])
	if err != nil {
		return time.Time{}, false
	}

	return date, true
}
