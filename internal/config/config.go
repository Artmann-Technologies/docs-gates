// Package config loads the per-repo `.docs-gates.json` file and holds the
// documented defaults every gate falls back to when the file is absent.
//
// Format decision (2026-11-03): JSON, not YAML — the module is stdlib-only
// by contract (hermetic nix builds, vendorHash-free), and JSON is jq-able,
// which matters in a fleet whose tooling speaks jq. YAML support is a
// deliberate exclusion; revisit only if a consumer cannot seed JSON.
//
// Load never merges half a file: the whole document decodes over the
// defaults, so a repo config only names what it changes. Unknown keys are
// WARNINGS, not errors (config sprawl guard: a typo'd key must announce
// itself, but a new optional key in a newer binary must not break older
// configs the other way). Malformed JSON and invalid values ARE errors —
// a gate silently running the wrong thresholds is worse than a loud exit 2.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FileName is the per-repo config file the gates look for under the repo
// root.
const FileName = ".docs-gates.json"

// Cadence holds the purge/archive thresholds (docs-cadence gate).
type Cadence struct {
	StruckRows    int `json:"struck_rows"`    // max struck DONE rows before the purge ritual is owed
	MaxUnarchived int `json:"max_unarchived"` // max unarchived reports before the archive ritual is owed
	OldestDays    int `json:"oldest_days"`    // max tolerated age (days) of the oldest unarchived report
}

// Config is the whole per-repo contract. Zero-value fields are never
// consumed directly — Defaults() fills everything, and Load starts from
// Defaults().
type Config struct {
	ReportDir         string   `json:"report_dir"`          // the unarchived status-report directory
	ReportGlob        string   `json:"report_glob"`         // filename pattern inside ReportDir and its archived subdir
	ArchivedSubdir    string   `json:"archived_subdir"`     // archive subdir UNDER ReportDir
	LivingDocs        []string `json:"living_docs"`         // the harvest corpus (freshness/routing)
	AnnotateExtraGlobs []string `json:"annotate_extra_globs"` // extra annotate corpus beyond reports (planning dirs)
	RuleF             bool     `json:"rule_f"`              // docs-annotate rule F (inline-code spacing) on/off
	RuleFExtensions   []string `json:"rule_f_extensions"`   // file extensions rule F scans (dot-prefixed)
	GenericDF         int      `json:"generic_df"`          // the distinctive-token bar (a token on >= N living lines is generic)
	AuditMax          int      `json:"audit_max"`           // docs-freshness --audit-max ratchet pin (-1 = strict)
	TodoFile          string   `json:"todo_file"`           // the TODO_LIST file (cadence reads it, docs-strike edits it)
	Cadence           Cadence  `json:"cadence"`
	SkipGates         []string `json:"skip_gates"` // battery verbs to skip (docs-gates run)
}

// Defaults returns the shipped configuration — the exact constants the
// gates carried before the config file existed (pbx-artmann's proven
// values). A repo without .docs-gates.json gets precisely the historical
// behavior.
func Defaults() Config {
	return Config{
		ReportDir:         "docs/status",
		ReportGlob:        "2*.md",
		ArchivedSubdir:    "archived",
		LivingDocs:        []string{"TODO_LIST.md", "ROADMAP.md", "CHANGELOG.md"},
		AnnotateExtraGlobs: []string{"docs/planning/*.md", "docs/planning/archived/*.md"},
		RuleF:             true,
		RuleFExtensions:   []string{".md", ".sh", ".nix"},
		GenericDF:         8,
		AuditMax:          -1,
		TodoFile:          "TODO_LIST.md",
		Cadence:           Cadence{StruckRows: 20, MaxUnarchived: 15, OldestDays: 21},
		SkipGates:         nil,
	}
}

// Path returns the config file location under a repo root.
func Path(root string) string {
	return filepath.Join(root, FileName)
}

// Load reads root's config over the defaults: (config, unknown-key
// warnings, error). A missing file is defaults with no warnings — the
// common case, not a finding. A malformed file or an invalid value is an
// error whose message is ready for the caller's FATAL line.
func Load(root string) (Config, []string, error) {
	cfg := Defaults()

	data, err := os.ReadFile(Path(root))
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil, nil
		}

		return cfg, nil, fmt.Errorf("reading %s: %w", Path(root), err)
	}

	// Unknown-key pass: decode into a key map first so a typo warns
	// instead of silently doing nothing.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return cfg, nil, fmt.Errorf("parsing %s: %w", Path(root), err)
	}

	known := knownKeys()
	var warnings []string

	for _, key := range sortedKeys(raw) {
		if !known[key] {
			warnings = append(warnings, fmt.Sprintf("unknown config key %q ignored", key))
		}
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, warnings, fmt.Errorf("parsing %s: %w", Path(root), err)
	}

	if err := cfg.validate(); err != nil {
		return cfg, warnings, fmt.Errorf("invalid %s: %w", Path(root), err)
	}

	return cfg, warnings, nil
}

// validate enforces the value contract beyond what JSON typing gives.
func (c Config) validate() error {
	const Template = "%s must be a relative directory, got %q"

	if c.ReportDir == "" || filepath.IsAbs(c.ReportDir) {
		return fmt.Errorf(Template, "report_dir", c.ReportDir)
	}

	if c.ReportGlob == "" {
		return fmt.Errorf("report_glob must not be empty")
	}

	if c.ArchivedSubdir == "" || filepath.IsAbs(c.ArchivedSubdir) {
		return fmt.Errorf(Template, "archived_subdir", c.ArchivedSubdir)
	}

	if len(c.LivingDocs) == 0 {
		return fmt.Errorf("living_docs must list at least one file")
	}

	if c.GenericDF < 1 {
		return fmt.Errorf("generic_df must be >= 1, got %d", c.GenericDF)
	}

	if c.AuditMax < -1 {
		return fmt.Errorf("audit_max must be >= -1 (-1 = strict), got %d", c.AuditMax)
	}

	if c.TodoFile == "" {
		return fmt.Errorf("todo_file must not be empty")
	}

	if c.Cadence.StruckRows < 1 || c.Cadence.MaxUnarchived < 1 || c.Cadence.OldestDays < 1 {
		return fmt.Errorf("cadence thresholds must all be >= 1, got %+v", c.Cadence)
	}

	for _, ext := range c.RuleFExtensions {
		if len(ext) < 2 || ext[0] != '.' {
			return fmt.Errorf("rule_f_extensions entries must be dot-prefixed extensions, got %q", ext)
		}
	}

	return nil
}

func knownKeys() map[string]bool {
	return map[string]bool{
		"report_dir":           true,
		"report_glob":          true,
		"archived_subdir":      true,
		"living_docs":          true,
		"annotate_extra_globs": true,
		"rule_f":               true,
		"rule_f_extensions":    true,
		"generic_df":           true,
		"audit_max":            true,
		"todo_file":            true,
		"cadence":              true,
		"skip_gates":           true,
	}
}

func sortedKeys(m map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	// Insertion sort over a tiny slice; no need for the sort package's
	// interface dance when the whole key set fits on one screen.
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}

	return keys
}
