package docscitations

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

const glossaryWithCitations = `# Domain Language

| Term | Definition | Context |
| --- | --- | --- |
| Schedule | A cron-based rule | ` + "`internal/models/schedule.go:27`" + ` |
| Helper | shell plumbing (` + "`scripts/lib.sh:7`" + `) | elsewhere |

Prose backticks that are NOT citations: ` + "`https://example.com/x.go`" + `,
a glob ` + "`cmd/*/main.go`" + `, an env token ` + "`$HOME/x.go`" + `, a doc
link ` + "`README.md`" + ` (extension not in the list), and an absolute
path ` + "`/etc/config.go`" + `.
`

func writeGlossaryRepo(t *testing.T, withTargets bool) string {
	t.Helper()

	root := t.TempDir()

	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(root, "docs", "DOMAIN_LANGUAGE.md"),
		[]byte(glossaryWithCitations),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	if withTargets {
		if err := os.MkdirAll(filepath.Join(root, "internal", "models"), 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.MkdirAll(filepath.Join(root, "scripts"), 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(
			filepath.Join(root, "internal", "models", "schedule.go"),
			[]byte("package models\n"),
			0o644,
		); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(root, "scripts", "lib.sh"), []byte("# lib\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return root
}

func TestResolvingGlossaryIsClean(t *testing.T) {
	t.Parallel()

	if rc := Run([]string{writeGlossaryRepo(t, true)}); rc != gatecli.ExitClean {
		t.Fatalf("a glossary whose citations resolve must be clean, got rc %d", rc)
	}
}

func TestDanglingCitationWarnsByDefault(t *testing.T) {
	t.Parallel()

	if rc := Run([]string{writeGlossaryRepo(t, false)}); rc != gatecli.ExitClean {
		t.Fatalf("dangling citations must WARN-exit-clean in default warn mode, got rc %d", rc)
	}
}

func TestDanglingCitationFailsWhenEnforced(t *testing.T) {
	t.Parallel()

	root := writeGlossaryRepo(t, false)

	if err := os.WriteFile(
		filepath.Join(root, config.FileName),
		[]byte(`{"citation_enforce": true}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	if rc := Run([]string{root}); rc != gatecli.ExitFindings {
		t.Fatalf("enforced dangling citations must fail, got rc %d", rc)
	}
}

// Classification is the gate's honesty: prose shapes in backticks must
// never become findings — the census glossaries carry URLs, globs and
// env tokens in code spans that are documentation, not pointers.
func TestCitationClassification(t *testing.T) {
	t.Parallel()

	exts := []string{".go", ".sh"}

	got := Citation(glossaryWithCitations, exts)

	want := []string{"internal/models/schedule.go", "scripts/lib.sh"}
	if len(got) != len(want) {
		t.Fatalf("classification: got %v, want %v", got, want)
	}

	for i, path := range want {
		if got[i] != path {
			t.Fatalf("classification: got %v, want %v", got, want)
		}
	}
}

func TestClassifyRefusesProseShapes(t *testing.T) {
	t.Parallel()

	exts := []string{".go"}

	for _, token := range []string{
		"https://example.com/x.go", // URL
		"cmd/*/main.go",            // glob
		"$HOME/x.go",               // env token
		"/etc/config.go",           // absolute
		"./local.go",               // relative-dot
		"README.md",                // extension not listed
		"main.go.orig",             // suffix not listed (.orig)
		"two words.go",             // space
	} {
		if _, ok := classify(token, exts); ok {
			t.Errorf("classify(%q) must refuse the prose shape", token)
		}
	}

	if path, ok := classify("internal/api/server.go:112-140", exts); !ok || path != "internal/api/server.go" {
		t.Errorf("classify line-range suffix: got %q ok=%v, want internal/api/server.go", path, ok)
	}
}

func TestDefaults(t *testing.T) {
	t.Parallel()

	defaults := config.Defaults()

	if defaults.CitationEnforce {
		t.Fatal("the shipped rollout is warn mode; citation_enforce defaults to false")
	}

	if len(defaults.CitationFiles) != 1 || defaults.CitationFiles[0] != config.DefaultDomainLanguageFile {
		t.Fatalf("default corpus must be the domain-language glossary, got %v", defaults.CitationFiles)
	}
}
