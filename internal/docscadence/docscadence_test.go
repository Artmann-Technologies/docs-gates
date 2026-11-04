package docscadence

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

func writeTodo(t *testing.T, rows string) string {
	t.Helper()

	root := t.TempDir()

	body := "# TODO\n\n" + rows + "\n"
	if err := os.WriteFile(filepath.Join(root, "TODO_LIST.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	return root
}

func row(strike string) string {
	return "| " + strike + " some task | ✅ `DONE` | P1 | 5m | evidence |\n"
}

// The 2026-10-08 (111 rows) and 2026-10-12 (123 rows) purges are the two
// data points behind the threshold: the backlog class silently rebuilds,
// so the gate must FAIL once struck rows cross it — a purge at ~20 rows
// is a session chore, at ~120 a CHANGELOG-preservation event.
func TestOverdueBacklogFails(t *testing.T) {
	t.Parallel()

	var rows strings.Builder
	for range Threshold + 1 {
		rows.WriteString(row("~~task~~ done (evidence)"))
	}

	if rc := Run([]string{writeTodo(t, rows.String())}); rc != gatecli.ExitFindings {
		t.Fatalf("%d struck rows must fail, got rc %d", Threshold+1, rc)
	}
}

func TestThresholdExactlyIsHealthy(t *testing.T) {
	t.Parallel()

	var rows strings.Builder
	for range Threshold {
		rows.WriteString(row("~~task~~ done (evidence)"))
	}

	if rc := Run([]string{writeTodo(t, rows.String())}); rc != gatecli.ExitClean {
		t.Fatalf("exactly %d struck rows must stay clean, got rc %d", Threshold, rc)
	}
}

// A struck row without the DONE status (verdict present, status flip
// pending, e.g. mid-strike) and a DONE status without the wrap
// (docs-strike misuse) must not silently dodge the counter: the pair is
// the completed-row signature, the halves are reported.
func TestHalfStrikesAreCountedAsWarnings(t *testing.T) {
	t.Parallel()

	rows := row("~~task~~ done (evidence)") + "| plain task | ✅ `DONE` | P1 | 5m | ev |\n"
	root := writeTodo(t, rows)

	if rc := Run([]string{root}); rc != gatecli.ExitClean {
		t.Fatalf("one struck + one bare DONE must stay clean, got rc %d", rc)
	}

	struck, bare := classify("# TODO\n\n" + rows)
	if struck != 1 || bare != 1 {
		t.Fatalf("want 1 struck + 1 bare DONE, got %d/%d", struck, bare)
	}
}

func TestMissingTodoIsUsage(t *testing.T) {
	t.Parallel()

	if rc := Run([]string{t.TempDir()}); rc != gatecli.ExitUsage {
		t.Fatalf("a missing TODO_LIST.md must exit usage, got rc %d", rc)
	}
}

// The live 2026-10-13 tree carries five rows whose TASK text contains the
// word DONE ("the SWITCH IS DONE AND VERIFIED", "absorb-sibling row DONE")
// — a line-based substring check miscounted all five as bare DONEs. The
// status check must be cell-scoped.
func TestProseDoneIsNotAStatusCell(t *testing.T) {
	t.Parallel()

	body := "# TODO\n\n" +
		"| USER: the switch IS DONE AND VERIFIED LIVE (prose) | 🔵 `BLOCKED` | High | 5min | ev |\n" +
		"| ~~task~~ done (evidence) | ✅ `DONE` | P1 | 5m | ev |\n"

	struck, bare := classify(body)
	if struck != 1 || bare != 0 {
		t.Fatalf("want 1 struck + 0 bare DONE (prose DONE is not a status), got %d/%d", struck, bare)
	}
}

// The strike ritual also lands DATED verdicts in the status cell
// ("✅ DONE 2026-11-04", "DONE — evidence"): the 2026-11-04 execution
// sessions struck 21 rows that shape and the counter read 0, so the
// purge tripwire could never fire (found by the 19:15 docs-health pass).
// A cell STARTING with a DONE marker counts; prose DONE still never
// occupies a whole cell.
func TestDatedDoneCellsAreStruckRows(t *testing.T) {
	t.Parallel()

	body := "# TODO\n\n" +
		"| ~~dated task~~ | ✅ DONE 2026-11-04 | P1 | 5m | ev |\n" +
		"| ~~dash task~~ | DONE — landed in ab12cd3 | P1 | 5m | ev |\n" +
		"| open task | 🔴 `TODO` | P1 | 5m | ev |\n"

	struck, bare := classify(body)
	if struck != 2 || bare != 0 {
		t.Fatalf("want 2 struck (dated DONE cells), got %d/%d", struck, bare)
	}
}

// writeReports stages dated report filenames under root's docs/status/.
func writeReports(t *testing.T, root string, names ...string) {
	t.Helper()

	dir := filepath.Join(root, "docs", "status")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, name := range names {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("# report\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func reportNames(count int, date string) []string {
	names := make([]string, 0, count)
	for i := range count {
		names = append(names, date+fmt.Sprintf("_10-00_report-%02d.md", i))
	}

	return names
}

// The boundary contract: exactly ArchiveThreshold unarchived reports is
// clean, one more fails — the purge arm's exact-threshold semantics,
// applied to the inventory.
func TestArchiveCountBoundary(t *testing.T) {
	t.Parallel()

	today := time.Now().Format("2006-01-02")

	atLimit := writeTodo(t, "")
	writeReports(t, atLimit, reportNames(ArchiveThreshold, today)...)

	if rc := Run([]string{atLimit}); rc != gatecli.ExitClean {
		t.Fatalf("exactly %d unarchived reports must stay clean, got rc %d", ArchiveThreshold, rc)
	}

	overLimit := writeTodo(t, "")
	writeReports(t, overLimit, reportNames(ArchiveThreshold+1, today)...)

	if rc := Run([]string{overLimit}); rc != gatecli.ExitFindings {
		t.Fatalf("%d unarchived reports must fail, got rc %d", ArchiveThreshold+1, rc)
	}
}

// An undated stray .md in docs/status/ is not a report and must not count.
func TestArchiveStraysAreSkipped(t *testing.T) {
	t.Parallel()

	today := time.Now().Format("2006-01-02")
	root := writeTodo(t, "")
	writeReports(t, root, today+"_10-00_real.md", "not-dated.md")

	count, _, oldest := summarize([]string{today + "_10-00_real.md", "not-dated.md"}, time.Now())
	if count != 1 || oldest != today+"_10-00_real.md" {
		t.Fatalf("want 1 counted report with the dated name oldest, got %d/%q", count, oldest)
	}
}

// The age boundary: an oldest report exactly ArchiveAgeThresholdDays old
// reads at the limit, one day older crosses it. The pinned clock keeps
// the boundary deterministic (the Run-level twin below uses an ancient
// hardcoded date so it never depends on wall time).
func TestArchiveAgeBoundary(t *testing.T) {
	t.Parallel()

	today := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	atLimit := today.AddDate(0, 0, -ArchiveAgeThresholdDays)
	overLimit := today.AddDate(0, 0, -ArchiveAgeThresholdDays-1)

	count, age, oldest := summarize(
		[]string{
			"2026-10-24_09-00_fresh.md",
			atLimit.Format("2006-01-02") + "_08-00_at-limit.md",
		},
		today,
	)
	if count != 2 || age != ArchiveAgeThresholdDays {
		t.Fatalf("want 2 reports aged exactly %d, got %d/%d", ArchiveAgeThresholdDays, count, age)
	}

	if !strings.HasPrefix(oldest, atLimit.Format("2006-01-02")) {
		t.Fatalf("oldest must be the at-limit report, got %q", oldest)
	}

	_, age, _ = summarize([]string{overLimit.Format("2006-01-02") + "_00-00_old.md"}, today)
	if age != ArchiveAgeThresholdDays+1 {
		t.Fatalf("one day past the limit must read %d days, got %d", ArchiveAgeThresholdDays+1, age)
	}
}

// The Run-level age proof: one ancient report (hardcoded 2020 date) is
// years past the threshold on any wall clock the test can run at.
func TestArchiveAncientReportFails(t *testing.T) {
	t.Parallel()

	root := writeTodo(t, "")
	writeReports(t, root, "2020-01-01_00-00_ancient.md")

	if rc := Run([]string{root}); rc != gatecli.ExitFindings {
		t.Fatalf("a report years old must fail the archive arm, got rc %d", rc)
	}
}

// The config contract: a repo's .docs-gates.json moves every knob — the
// thresholds, the TODO file, and the report dir — away from the shipped
// defaults, and the gate enforces the configured values exactly: three
// struck rows fail under a configured bar of two where the default twenty
// would pass, and the default TODO_LIST.md being absent is not a finding
// once the config renames the file.
func TestConfigDrivenKnobs(t *testing.T) {
	t.Parallel()

	cfg := `{"todo_file": "TODOS.md", "report_dir": "meta/reports",
		"cadence": {"struck_rows": 2, "max_unarchived": 1, "oldest_days": 5}}`

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, config.FileName), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	var rows strings.Builder
	for range 2 {
		rows.WriteString(row("~~task~~ done (evidence)"))
	}

	rows.WriteString(row("~~widget~~ done (verified)"))

	if err := os.WriteFile(
		filepath.Join(root, "TODOS.md"),
		[]byte("# TODO\n\n"+rows.String()+"\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	if rc := Run([]string{root}); rc != gatecli.ExitFindings {
		t.Fatalf("3 struck rows under struck_rows=2 must fail, got rc %d", rc)
	}

	today := time.Now().Format("2006-01-02")

	repdir := filepath.Join(root, "meta", "reports")
	if err := os.MkdirAll(repdir, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{today + "_10-00_a.md", today + "_10-00_b.md"} {
		if err := os.WriteFile(filepath.Join(repdir, name), []byte("# r\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if rc := Run([]string{root}); rc != gatecli.ExitFindings {
		t.Fatalf("2 unarchived reports under max_unarchived=1 must fail, got rc %d", rc)
	}
}
