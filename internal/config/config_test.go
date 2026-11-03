package config

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

func writeConfig(t *testing.T, dir, body string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(body), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestLoadMissingFileIsDefaultsWithoutWarnings(t *testing.T) {
	root := t.TempDir()

	cfg, warnings, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(warnings) != 0 {
		t.Fatalf("warnings = %v, want none", warnings)
	}

	want := Defaults()
	if !reflect.DeepEqual(cfg, want) {
		t.Fatalf("cfg = %+v, want defaults %+v", cfg, want)
	}
}

func TestLoadPartialFileKeepsDefaultsForTheRest(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `{"report_dir": "notes/status", "generic_df": 5}`)

	cfg, warnings, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(warnings) != 0 {
		t.Fatalf("warnings = %v, want none", warnings)
	}

	if cfg.ReportDir != "notes/status" {
		t.Errorf("ReportDir = %q", cfg.ReportDir)
	}

	if cfg.GenericDF != 5 {
		t.Errorf("GenericDF = %d, want 5", cfg.GenericDF)
	}

	if cfg.ReportGlob != Defaults().ReportGlob {
		t.Errorf("ReportGlob = %q, want the default", cfg.ReportGlob)
	}

	if !reflect.DeepEqual(cfg.Cadence, Defaults().Cadence) {
		t.Errorf("Cadence = %+v, want defaults", cfg.Cadence)
	}
}

func TestLoadUnknownKeyWarnsAndContinues(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `{"report_dir": "x", "generic_df_bar": 9}`)

	_, warnings, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(warnings) != 1 || !contains(warnings, `unknown config key "generic_df_bar" ignored`) {
		t.Fatalf("warnings = %v, want the generic_df_bar notice", warnings)
	}
}

func TestLoadMalformedJSONIsAnError(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `{"report_dir": }`)

	if _, _, err := Load(root); err == nil {
		t.Fatal("Load succeeded on malformed JSON, want an error")
	}
}

func TestLoadInvalidValuesAreErrors(t *testing.T) {
	cases := map[string]string{
		"generic_df zero":   `{"generic_df": 0}`,
		"audit_max too low": `{"audit_max": -2}`,
		"empty todo_file":   `{"todo_file": ""}`,
		"cadence zero":      `{"cadence": {"struck_rows": 0}}`,
		"bad extension":     `{"rule_f_extensions": ["md"]}`,
		"abs report_dir":    `{"report_dir": "/tmp/status"}`,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			writeConfig(t, root, body)

			if _, _, err := Load(root); err == nil {
				t.Fatalf("Load(%s) succeeded, want an error", body)
			}
		})
	}
}

func TestDefaultsValidate(t *testing.T) {
	if err := Defaults().validate(); err != nil {
		t.Fatalf("defaults do not validate: %v", err)
	}
}

func TestPathJoinsRoot(t *testing.T) {
	if got, want := Path("/repo"), "/repo/.docs-gates.json"; got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
}

func contains(list []string, needle string) bool {
	return slices.Contains(list, needle)
}
