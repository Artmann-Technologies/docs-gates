package docsrouting

import (
	"cmp"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/docsfreshness"
)

// Settings is routing's own projection of the repo config: the report
// corpus, the living docs, and the distinctive bar (a zero GenericDF
// falls back to docsfreshness.GenericDF).
type Settings struct {
	ReportDir  string
	ReportGlob string
	LivingDocs []string
	GenericDF  int
}

func settingsFromConfig(c config.Config) Settings {
	return Settings{
		ReportDir:  c.ReportDir,
		ReportGlob: c.ReportGlob,
		LivingDocs: c.LivingDocs,
		GenericDF:  c.GenericDF,
	}
}

func (s Settings) df() int {
	if s.GenericDF > 0 {
		return s.GenericDF
	}

	return docsfreshness.GenericDF
}

func (s Settings) livingDocs() []string {
	return config.OrSlice(s.LivingDocs, config.Defaults().LivingDocs)
}

func (s Settings) reportDir() string {
	return cmp.Or(s.ReportDir, config.Defaults().ReportDir)
}

func (s Settings) reportGlob() string {
	return cmp.Or(s.ReportGlob, config.Defaults().ReportGlob)
}
