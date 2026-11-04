package docsannotate

import (
	"cmp"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// Settings carries the config-driven knobs of the annotate gate. Gate
// logic only ever sees Settings produced by loadSettings (DefaultSettings
// or SettingsFromConfig), so the historical behavior survives both the
// config-less repo and the silent-key config; do not construct Settings{}
// by hand in production paths.
type Settings struct {
	ReportDir          string   // unarchived status-report directory under the repo root
	ArchivedSubdir     string   // archive subdir UNDER ReportDir
	AnnotateExtraGlobs []string // extra annotate corpus beyond reports (planning dirs)
	RuleF              bool     // docs-annotate rule F (inline-code spacing) on/off
	RuleFExtensions    []string // file extensions rule F scans (dot-prefixed)
	AnnotateMax        int      // adoption ratchet pin (-1 = strict)
}

// DefaultSettings mirrors config.Defaults() — the exact knobs this gate
// carried before the config file existed.
func DefaultSettings() Settings {
	c := config.Defaults()

	return Settings{
		ReportDir:       c.ReportDir,
		ArchivedSubdir:  c.ArchivedSubdir,
		RuleF:           c.RuleF,
		RuleFExtensions: c.RuleFExtensions,
		AnnotateMax:     c.AnnotateMax,
	}
}

// SettingsFromConfig projects the whole repo config onto the fields this
// gate consumes.
func SettingsFromConfig(c config.Config) Settings {
	return Settings{
		ReportDir:          c.ReportDir,
		ArchivedSubdir:     c.ArchivedSubdir,
		AnnotateExtraGlobs: c.AnnotateExtraGlobs,
		RuleF:              c.RuleF,
		RuleFExtensions:    c.RuleFExtensions,
		AnnotateMax:        c.AnnotateMax,
	}
}

func (s Settings) dir() string {
	return cmp.Or(s.ReportDir, config.Defaults().ReportDir)
}

func (s Settings) archivedSub() string {
	return cmp.Or(s.ArchivedSubdir, config.Defaults().ArchivedSubdir)
}

func (s Settings) extraGlobs() []string {
	return config.OrSlice(s.AnnotateExtraGlobs, config.Defaults().AnnotateExtraGlobs)
}

func (s Settings) ruleFExtensions() []string {
	return config.OrSlice(s.RuleFExtensions, config.Defaults().RuleFExtensions)
}

// loadSettings resolves the repo config for a Run-shaped entry point and
// reports it ready-to-use: config warnings go to stderr (non-fatal), a
// malformed config is an error (the caller FATALs with exit 2).
func loadSettings(root string) (Settings, error) {
	cfg, err := gatecli.LoadConfig(root)
	if err != nil {
		return DefaultSettings(), err
	}

	return SettingsFromConfig(cfg), nil
}
