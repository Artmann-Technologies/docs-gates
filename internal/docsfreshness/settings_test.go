package docsfreshness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// The config-driven contract end to end: a repo whose .docs-gates.json
// moves the report corpus, renames the living docs and tightens the
// distinctive bar must be judged against ITS layout, not the defaults.
func TestSettingsForRootHonorsRepoConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	reports := filepath.Join(root, "notes", "audits")
	if err := os.MkdirAll(reports, 0o755); err != nil {
		t.Fatal(err)
	}

	report := "# Report\n\n## f) Next\n\n1. offset recalibration rework\n"
	if err := os.WriteFile(filepath.Join(reports, "2026-01-01_00-00_AUDIT.md"), []byte(report), 0o644); err != nil {
		t.Fatal("write report")
	}

	// The default layout must stay EMPTY (nothing is read from
	// docs/status), while "offset" and "recalibration" each ride exactly
	// 3 living lines in the CUSTOM corpus: below a bar of 4 they are
	// distinctive and the item harvests on them; under a bar of 2 both
	// genericize and the audit must flip. The custom corpus is what got
	// read — that is the point of the fixture.
	todo := strings.Repeat("- offset recalibration watchrow\n", 3)
	todo += "- unrelated filler\n"

	if err := os.WriteFile(filepath.Join(root, "NOTES.md"), []byte(todo), 0o644); err != nil {
		t.Fatal("write living doc")
	}

	cfg := `{
		"report_dir": "notes/audits",
		"report_glob": "20*.md",
		"living_docs": ["NOTES.md"],
		"generic_df": 4
	}`
	if err := os.WriteFile(filepath.Join(root, ".docs-gates.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal("write config")
	}

	if rc := Run([]string{root, "--audit-all"}); rc != gatecli.ExitClean {
		t.Fatalf("the custom-corpus repo must audit clean, got rc %d", rc)
	}

	// And the negative: under a bar of 2, both matched tokens (3 lines
	// each) are generic, and "rework" is missing — the audit must go red.
	cfg2 := `{"report_dir": "notes/audits", "report_glob": "20*.md", "living_docs": ["NOTES.md"], "generic_df": 2}`
	if err := os.WriteFile(filepath.Join(root, ".docs-gates.json"), []byte(cfg2), 0o644); err != nil {
		t.Fatal("rewrite config")
	}

	if rc := Run([]string{root, "--audit-all"}); rc != gatecli.ExitFindings {
		t.Fatalf("a tightened bar must flip the audit to findings, got rc %d", rc)
	}
}

// A malformed config is a usage error, never silently-ignored defaults.
func TestRunFatalsOnMalformedConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".docs-gates.json"), []byte(`{`), 0o644); err != nil {
		t.Fatal(err)
	}

	if rc := Run([]string{root}); rc != gatecli.ExitUsage {
		t.Fatalf("malformed config must exit usage, got %d", rc)
	}
}

// Zero-value Settings still behave like the defaults (the df() and path
// fallbacks), so tests can construct fragments without the full default
// set.
func TestZeroSettingsFallBackToDefaults(t *testing.T) {
	t.Parallel()

	var s Settings
	if got := s.df(); got != GenericDF {
		t.Fatalf("zero Settings df = %d, want the GenericDF constant %d", got, GenericDF)
	}

	if s.dir() == "" || s.archivedSub() == "" || len(s.livingDocs()) == 0 {
		t.Fatalf("zero Settings path fallbacks missing: %q %q %v", s.dir(), s.archivedSub(), s.livingDocs())
	}
}
