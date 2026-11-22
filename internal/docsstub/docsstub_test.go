package docsstub

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// bareTemplate is the fleet's unfilled DDD scaffold byte-for-byte (the
// 38-copy census group, 2026-11-21). A copy of it under the corpus
// path must be the gate's canonical finding.
func bareTemplate(t *testing.T) []byte {
	t.Helper()

	body, err := os.ReadFile(filepath.Join("testdata", "bare-template.md"))
	if err != nil {
		t.Fatal(err)
	}

	return body
}

func writeCorpus(t *testing.T, body []byte) string {
	t.Helper()

	root := t.TempDir()

	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(root, "docs", "DOMAIN_LANGUAGE.md"), body, 0o644); err != nil {
		t.Fatal(err)
	}

	return root
}

func writeConfig(t *testing.T, root, json string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(root, config.FileName), []byte(json), 0o644); err != nil {
		t.Fatal(err)
	}
}

// The 38-repo census class: the bare template under the corpus path is
// THE finding. Warn mode (the shipped default) must keep the battery
// green while making the debt visible.
func TestBareTemplateWarnsByDefault(t *testing.T) {
	t.Parallel()

	if rc := Run([]string{writeCorpus(t, bareTemplate(t))}); rc != gatecli.ExitClean {
		t.Fatalf("bare template must WARN-exit-clean in default warn mode, got rc %d", rc)
	}
}

func TestBareTemplateFailsWhenEnforced(t *testing.T) {
	t.Parallel()

	root := writeCorpus(t, bareTemplate(t))
	writeConfig(t, root, `{"stub_enforce": true}`)

	if rc := Run([]string{root}); rc != gatecli.ExitFindings {
		t.Fatalf("enforced bare template must fail, got rc %d", rc)
	}
}

func TestRealGlossaryIsClean(t *testing.T) {
	t.Parallel()

	glossary := "# Domain Language\n\n| Term | Definition | Context |\n| --- | --- | --- |\n| Verdict | an honest answer | api/verdict.go |\n"

	if rc := Run([]string{writeCorpus(t, []byte(glossary))}); rc != gatecli.ExitClean {
		t.Fatalf("a filled glossary must be clean, got rc %d", rc)
	}
}

func TestExemptFileIsSkip(t *testing.T) {
	t.Parallel()

	root := writeCorpus(t, bareTemplate(t))
	writeConfig(t, root, `{"stub_enforce": true, "stub_exempts": ["docs/DOMAIN_LANGUAGE.md"]}`)

	if rc := Run([]string{root}); rc != gatecli.ExitClean {
		t.Fatalf("an exempt corpus file must stay clean even enforced, got rc %d", rc)
	}
}

func TestMissingCorpusFileIsSkipNotFinding(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	if rc := Run([]string{root}); rc != gatecli.ExitClean {
		t.Fatalf("a repo without the corpus file must be clean (freshness owns absence), got rc %d", rc)
	}
}

func TestDefaultsPinTheCensusDigests(t *testing.T) {
	t.Parallel()

	defaults := config.Defaults()

	if len(defaults.StubChecksums) != 2 {
		t.Fatalf("defaults must pin both digests of the bare template, got %v", defaults.StubChecksums)
	}

	if defaults.StubEnforce {
		t.Fatal("the shipped rollout is warn mode; stub_enforce defaults to false")
	}

	if len(defaults.StubFiles) != 1 || defaults.StubFiles[0] != config.DefaultDomainLanguageFile {
		t.Fatalf("default corpus must be the domain-language glossary, got %v", defaults.StubFiles)
	}
}

// The fixture IS the pinned template: if the fleet scaffold ever
// changes, this test forces the defaults (and the census evidence) to
// be re-derived instead of silently matching nothing.
func TestFixtureMatchesPinnedChecksums(t *testing.T) {
	t.Parallel()

	findings, _, err := Scan("", Settings{
		Files:     []string{filepath.Join("testdata", "bare-template.md")},
		Checksums: config.Defaults().StubChecksums,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(findings) != 1 {
		t.Fatalf("the tracked bare-template fixture must match the pinned digests, got %d findings", len(findings))
	}
}
