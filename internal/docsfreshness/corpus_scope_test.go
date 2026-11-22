package docsfreshness

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// corpusScopeFixture builds a repo whose ARCHIVED report carries an
// unharvested open item (distinctive tokens absent from every living
// doc) while the unarchived report's item harvests on its living-doc
// rows. The archived item is the only finding the default (all) corpus
// can produce, which makes the scope knob's effect observable through
// the audit verdict alone.
func corpusScopeFixture(t *testing.T) string {
	t.Helper()

	root := t.TempDir()

	status := filepath.Join(root, "docs", "status")
	if err := os.MkdirAll(filepath.Join(status, "archived"), 0o755); err != nil {
		t.Fatal(err)
	}

	unarchived := "# Report\n\n## f) Next\n\n1. routine row alpha\n"
	if err := os.WriteFile(filepath.Join(status, "2026-01-01_00-00_NEW.md"), []byte(unarchived), 0o644); err != nil {
		t.Fatal("write unarchived report")
	}

	archived := "# Archived\n\n## f) Next\n\n1. quantumzeta flux recalibration\n"

	archivedPath := filepath.Join(status, "archived", "2026-01-02_00-00_OLD.md")
	if err := os.WriteFile(archivedPath, []byte(archived), 0o644); err != nil {
		t.Fatal("write archived report")
	}

	todo := "- routine row alpha\n- filler one\n- filler two\n"
	if err := os.WriteFile(filepath.Join(root, "TODO_LIST.md"), []byte(todo), 0o644); err != nil {
		t.Fatal("write TODO_LIST")
	}

	if err := os.WriteFile(filepath.Join(root, "ROADMAP.md"), []byte("- roadmap filler\n"), 0o644); err != nil {
		t.Fatal("write ROADMAP")
	}

	changelog := "## Changed\n\n- changelog filler\n"
	if err := os.WriteFile(filepath.Join(root, "CHANGELOG.md"), []byte(changelog), 0o644); err != nil {
		t.Fatal("write CHANGELOG")
	}

	return root
}

// The default corpus replays BOTH globs — the pre-knob behavior, pinned
// here so the knob can never silently narrow it.
func TestStatusReportsDefaultIncludesArchived(t *testing.T) {
	t.Parallel()

	root := corpusScopeFixture(t)

	reports := statusReports(root, DefaultSettings())
	if len(reports) != 2 {
		t.Fatalf("default corpus = %d report(s), want 2 (unarchived + archived): %v", len(reports), reports)
	}

	if !slices.ContainsFunc(reports, func(p string) bool { return filepath.Base(p) == "2026-01-02_00-00_OLD.md" }) {
		t.Fatalf("default corpus must contain the archived report: %v", reports)
	}

	// And the end to end: the archived item fires under the default.
	if rc := Run([]string{root, "--audit-all"}); rc != gatecli.ExitFindings {
		t.Fatalf("default corpus must report the archived item, got rc %d", rc)
	}
}

// audit_corpus "unarchived" drops the archived subdir from the replay:
// the archived backlog stops moving the unharvested count with unrelated
// living-doc prose edits (the go-sse 122 → 123 → 120 drift class).
func TestAuditCorpusUnarchivedExcludesArchived(t *testing.T) {
	t.Parallel()

	root := corpusScopeFixture(t)

	cfg := `{"audit_corpus": "unarchived"}`
	if err := os.WriteFile(filepath.Join(root, ".docs-gates.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal("write config")
	}

	reports := statusReports(root, mustSettings(t, root))
	if len(reports) != 1 || filepath.Base(reports[0]) != "2026-01-01_00-00_NEW.md" {
		t.Fatalf("unarchived corpus = %v, want only the unarchived report", reports)
	}

	if rc := Run([]string{root, "--audit-all"}); rc != gatecli.ExitClean {
		t.Fatalf("unarchived scope must skip the archived item, got rc %d", rc)
	}
}

// An unset audit_corpus (zero-value Settings, empty JSON string) is the
// shipped all-corpus behavior.
func TestAuditCorpusZeroValueDefaultsToAll(t *testing.T) {
	t.Parallel()

	var zero Settings
	if got := zero.auditCorpus(); got != "all" {
		t.Fatalf("zero Settings auditCorpus = %q, want \"all\"", got)
	}

	if got := DefaultSettings().auditCorpus(); got != "all" {
		t.Fatalf("DefaultSettings auditCorpus = %q, want \"all\"", got)
	}
}

// mustSettings resolves the fixture's repo config, failing the test on
// a load error instead of silently auditing the defaults.
func mustSettings(t *testing.T, root string) Settings {
	t.Helper()

	settings, err := loadSettings(root)
	if err != nil {
		t.Fatalf("loadSettings: %v", err)
	}

	return settings
}
