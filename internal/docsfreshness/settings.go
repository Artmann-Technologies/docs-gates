package docsfreshness

import (
	"path/filepath"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// Settings carries the config-driven knobs of the freshness gate. The
// zero value is valid everywhere: path fields fall back to the shipped
// layout and df() falls back to the GenericDF constant, so tests can
// pass DefaultSettings() (or literally Settings{} without list fields)
// and get the historical behavior.
type Settings struct {
	ReportDir      string   // unarchived status-report directory under the repo root
	ReportGlob     string   // report filename pattern
	ArchivedSubdir string   // archive subdir UNDER ReportDir
	LivingDocs     []string // the harvest corpus file names, under the repo root
	GenericDF      int      // per-repo distinctive-token bar override (<= 0 = the GenericDF constant)
}

// DefaultSettings mirrors config.Defaults() — the exact constants this
// gate carried before the config file existed.
func DefaultSettings() Settings {
	c := config.Defaults()

	return Settings{
		ReportDir:      c.ReportDir,
		ReportGlob:     c.ReportGlob,
		ArchivedSubdir: c.ArchivedSubdir,
		LivingDocs:     c.LivingDocs,
		GenericDF:      c.GenericDF,
	}
}

// SettingsFromConfig projects the whole repo config onto the fields this
// gate consumes.
func SettingsFromConfig(c config.Config) Settings {
	return Settings{
		ReportDir:      c.ReportDir,
		ReportGlob:     c.ReportGlob,
		ArchivedSubdir: c.ArchivedSubdir,
		LivingDocs:     c.LivingDocs,
		GenericDF:      c.GenericDF,
	}
}

// df is the effective distinctive-token bar; a zero/negative override
// (the zero value of Settings) selects the GenericDF constant.
func (s Settings) df() int {
	if s.GenericDF > 0 {
		return s.GenericDF
	}

	return GenericDF
}

func (s Settings) reportGlobPath(root string) string {
	return filepath.Join(root, s.dir(), s.ReportGlob)
}

func (s Settings) archivedGlobPath(root string) string {
	return filepath.Join(root, s.dir(), s.archivedSub(), s.ReportGlob)
}

func (s Settings) dir() string {
	if s.ReportDir != "" {
		return s.ReportDir
	}

	return config.Defaults().ReportDir
}

func (s Settings) archivedSub() string {
	if s.ArchivedSubdir != "" {
		return s.ArchivedSubdir
	}

	return config.Defaults().ArchivedSubdir
}

func (s Settings) livingDocs() []string {
	if len(s.LivingDocs) > 0 {
		return s.LivingDocs
	}

	return config.Defaults().LivingDocs
}

// loadSettings resolves the repo config for a Run-shaped entry point and
// reports it ready-to-use: config warnings go to stderr (non-fatal), a
// malformed config is an error (the caller FATALs with exit 2).
func loadSettings(root string) (Settings, error) {
	cfg, warnings, err := config.Load(root)
	for _, w := range warnings {
		gatecli.Errf("config: %s\n", w)
	}

	if err != nil {
		return DefaultSettings(), err
	}

	return SettingsFromConfig(cfg), nil
}
