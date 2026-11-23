package docsannotate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// The config-driven corpus: a repo whose report dir lives elsewhere gets
// its reports annotated from THERE, and rule_f: false silences the F
// rules entirely (the toggle the dprint-less repos need).
func TestRunHonorsConfiguredReportDirAndRuleFToggle(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	reports := filepath.Join(root, "notes", "audits")
	if err := os.MkdirAll(reports, 0o755); err != nil {
		t.Fatal(err)
	}

	// A struck item with NO verdict (rule A finding) inside the custom
	// report dir, plus a rule-F violation (letter touching a backtick)
	// in a .md the default walk WOULD scan.
	report := "# Report\n\n## f) Next\n\n1. ~~finished thing~~\n"
	if err := os.WriteFile(filepath.Join(reports, "2026-01-01_00-00_A.md"), []byte(report), 0o644); err != nil {
		t.Fatal("write report")
	}

	ruleF := "# Notes\n\nsee`config.go`here\n"
	if err := os.WriteFile(filepath.Join(root, "NOTES.md"), []byte(ruleF), 0o644); err != nil {
		t.Fatal("write notes")
	}

	cfg := `{"report_dir": "notes/audits", "report_glob": "20*.md", "rule_f": false}`
	if err := os.WriteFile(filepath.Join(root, ".docs-gates.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal("write config")
	}

	captured := gatecli.CaptureOutput(func() int {
		return Run([]string{root})
	})

	if captured.Code != gatecli.ExitFindings {
		t.Fatalf("the struck-without-verdict item must fail, got rc %d (out:\n%s)", captured.Code, captured.Stdout)
	}

	if containsSubstr(captured.Stdout, "F/spacing") {
		t.Fatalf("rule_f=false must silence the F rule, got:\n%s", captured.Stdout)
	}

	if !containsSubstr(captured.Stdout, "A/evidence") {
		t.Fatalf("the custom report dir must be the annotate corpus, got:\n%s", captured.Stdout)
	}
}

func containsSubstr(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

// annotateScopeFixture builds a repo whose ARCHIVED report carries the
// only possible finding (a struck item with no verdict marker) while the
// unarchived corpus is clean: the annotate_corpus knob's effect is
// observable through the verdict alone — the default (all) corpus FAILs,
// "unarchived" stays green without skipping the gate (#4).
func annotateScopeFixture(t *testing.T) string {
	t.Helper()

	root := t.TempDir()

	status := filepath.Join(root, "docs", "status")
	if err := os.MkdirAll(filepath.Join(status, "archived"), 0o755); err != nil {
		t.Fatal(err)
	}

	unarchived := "# Live\n\n## f) Next\n\n1. ~~closed cleanly~~ done at `a1b2c3d`\n"
	if err := os.WriteFile(filepath.Join(status, "2026-01-01_00-00_NEW.md"), []byte(unarchived), 0o644); err != nil {
		t.Fatal(err)
	}

	archived := "# Old\n\n## f) Next\n\n1. ~~a history-locked bare strike~~\n"
	if err := os.WriteFile(
		filepath.Join(status, "archived", "2026-01-02_00-00_OLD.md"),
		[]byte(archived),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	return root
}

//nolint:paralleltest // process-global capture swap, serialized on purpose
func TestRunAnnotateCorpusUnarchivedExcludesArchive(t *testing.T) {
	root := annotateScopeFixture(t)

	if err := os.WriteFile(
		filepath.Join(root, ".docs-gates.json"),
		[]byte(`{"annotate_corpus": "unarchived"}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	captured := gatecli.CaptureOutput(func() int {
		return Run([]string{root})
	})

	if captured.Code != gatecli.ExitClean {
		t.Fatalf(
			"unarchived scope must stay green without findings, code = %d:\n%s%s",
			captured.Code,
			captured.Stdout,
			captured.Stderr,
		)
	}

	if !strings.Contains(captured.Stdout, "CLEAN") {
		t.Fatalf("verdict line missing:\n%s", captured.Stdout)
	}
}

//nolint:paralleltest // process-global capture swap, serialized on purpose
func TestRunAnnotateCorpusDefaultScansArchive(t *testing.T) {
	root := annotateScopeFixture(t)

	captured := gatecli.CaptureOutput(func() int {
		return Run([]string{root})
	})

	if captured.Code != gatecli.ExitFindings {
		t.Fatalf("the default (all) corpus must keep scanning the archive, code = %d", captured.Code)
	}

	if !strings.Contains(captured.Stdout, "A/evidence") {
		t.Fatalf("the archived bare strike must be the finding:\n%s", captured.Stdout)
	}
}
