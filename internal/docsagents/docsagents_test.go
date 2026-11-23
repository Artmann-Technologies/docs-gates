package docsagents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// writeAgentsMd writes the corpus file with its mtime backdated ageDays
// into the past and returns the mtime, so a Scan test can pin Now =
// mtime + ageDays and the age in days is exact (no wall-clock drift
// through the integer-day truncation).
func writeAgentsMd(t *testing.T, root string, ageDays int) time.Time {
	t.Helper()

	path := filepath.Join(root, config.DefaultAgentsFile)

	if err := os.WriteFile(path, []byte("# Agent Configuration\n\ncontext, conventions, key patterns\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mtime := time.Now().AddDate(0, 0, -ageDays)

	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatal(err)
	}

	return mtime
}

func writeConfig(t *testing.T, root, json string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(root, config.FileName), []byte(json), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFreshAgentsMdIsClean(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeAgentsMd(t, root, 1)

	if rc := Run([]string{root}); rc != gatecli.ExitClean {
		t.Fatalf("a one-day-old AGENTS.md must be clean, got rc %d", rc)
	}
}

func TestMissingAgentsMdWarnsByDefault(t *testing.T) {
	t.Parallel()

	if rc := Run([]string{t.TempDir()}); rc != gatecli.ExitClean {
		t.Fatalf("a missing AGENTS.md must WARN-exit-clean in default warn mode, got rc %d", rc)
	}
}

func TestMissingAgentsMdFailsWhenEnforced(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeConfig(t, root, `{"agents_enforce": true}`)

	if rc := Run([]string{root}); rc != gatecli.ExitFindings {
		t.Fatalf("an enforced missing AGENTS.md must fail, got rc %d", rc)
	}
}

func TestMaxBarAgeWarnsByDefault(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeAgentsMd(t, root, config.DefaultAgentsMaxDays+5)

	if rc := Run([]string{root}); rc != gatecli.ExitClean {
		t.Fatalf("a max-bar age must WARN-exit-clean in default warn mode, got rc %d", rc)
	}
}

func TestMaxBarAgeFailsWhenEnforced(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeAgentsMd(t, root, config.DefaultAgentsMaxDays+5)
	writeConfig(t, root, `{"agents_enforce": true}`)

	if rc := Run([]string{root}); rc != gatecli.ExitFindings {
		t.Fatalf("an enforced max-bar age must fail, got rc %d", rc)
	}
}

func TestWarnBarAgeStaysCleanEvenEnforced(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeAgentsMd(t, root, config.DefaultAgentsWarnDays+5)
	writeConfig(t, root, `{"agents_enforce": true}`)

	if rc := Run([]string{root}); rc != gatecli.ExitClean {
		t.Fatalf("a warn-bar age is informational at every enforcement level, got rc %d", rc)
	}
}

// TestRunPrintsVerdictLines pins the human output: warn mode spells
// WARN with the rollout suffix, enforcement spells FAIL on the same
// finding. Serialized: the capture swap is process-global.
//
//nolint:paralleltest // process-global capture swap, serialized on purpose
func TestRunPrintsVerdictLines(t *testing.T) {
	stale := t.TempDir()
	writeAgentsMd(t, stale, config.DefaultAgentsMaxDays+5)

	warnOut := gatecli.CaptureOutput(func() int {
		return Run([]string{stale})
	})

	if warnOut.Code != gatecli.ExitClean {
		t.Fatalf("warn mode must stay exit-clean, got rc %d:\n%s", warnOut.Code, warnOut.Stdout)
	}

	if !strings.Contains(warnOut.Stdout, "WARN "+config.DefaultAgentsFile) ||
		!strings.Contains(warnOut.Stdout, "agents_enforce=false, warn mode") {
		t.Fatalf("warn mode must print the WARN line with the rollout suffix, got:\n%s", warnOut.Stdout)
	}

	writeConfig(t, stale, `{"agents_enforce": true}`)

	failOut := gatecli.CaptureOutput(func() int {
		return Run([]string{stale})
	})

	if failOut.Code != gatecli.ExitFindings {
		t.Fatalf("enforced mode must exit findings, got rc %d:\n%s", failOut.Code, failOut.Stdout)
	}

	if !strings.Contains(failOut.Stdout, "FAIL "+config.DefaultAgentsFile) {
		t.Fatalf("enforced mode must print the FAIL line, got:\n%s", failOut.Stdout)
	}
}

// TestScanVerdicts drives the pure scan core table-driven: the bar
// boundaries and the missing-file verdict.
func TestScanVerdicts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ageDays  int
		missing  bool
		severity string
	}{
		{name: "fresh stays silent", ageDays: config.DefaultAgentsWarnDays},
		{name: "warn bar flags warn", ageDays: config.DefaultAgentsWarnDays + 1, severity: "warn"},
		{name: "max bar flags fail", ageDays: config.DefaultAgentsMaxDays + 1, severity: "fail"},
		{name: "missing flags fail", missing: true, severity: "fail"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			now := time.Now()
			if !tt.missing {
				now = writeAgentsMd(t, root, tt.ageDays).AddDate(0, 0, tt.ageDays)
			}

			findings, err := Scan(root, Settings{
				File:     config.DefaultAgentsFile,
				WarnDays: config.DefaultAgentsWarnDays,
				MaxDays:  config.DefaultAgentsMaxDays,
				Now:      now,
			})
			if err != nil {
				t.Fatal(err)
			}

			if tt.severity == "" {
				if len(findings) != 0 {
					t.Fatalf("fresh file must be silent, got %v", findings)
				}

				return
			}

			if len(findings) != 1 || findings[0].Severity != tt.severity || findings[0].Missing != tt.missing {
				t.Fatalf("want one %s finding (missing=%v), got %v", tt.severity, tt.missing, findings)
			}
		})
	}
}

// TestSettingsFromConfigKeepsDefaults pins the zero-config projection:
// a repo without .docs-gates.json gets the shipped defaults, never
// zero-value knobs. Load starts from Defaults() and decodes over it, so
// Defaults() IS what a silent config produces; the File field
// additionally survives a fully zero Config via config.Or.
func TestSettingsFromConfigKeepsDefaults(t *testing.T) {
	t.Parallel()

	settings := SettingsFromConfig(config.Defaults())

	if settings.File != config.DefaultAgentsFile {
		t.Fatalf("silent config must fall back to %s, got %q", config.DefaultAgentsFile, settings.File)
	}

	if settings.WarnDays != config.DefaultAgentsWarnDays || settings.MaxDays != config.DefaultAgentsMaxDays {
		t.Fatalf("silent config must keep the shipped bars, got warn=%d max=%d", settings.WarnDays, settings.MaxDays)
	}

	if settings.Enforce {
		t.Fatal("silent config must ship warn mode")
	}

	if zero := SettingsFromConfig(config.Config{}); zero.File != config.DefaultAgentsFile {
		t.Fatalf("zero-value config must still fall back to %s, got %q", config.DefaultAgentsFile, zero.File)
	}
}
