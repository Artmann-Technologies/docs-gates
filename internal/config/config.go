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
	"slices"
	"strings"
)

// FileName is the per-repo config file the gates look for under the repo
// root.
const FileName = ".docs-gates.json"

// The shipped default values (see Defaults). Named so the defaults block
// reads as a table of decisions, not a pile of magic numbers.
const (
	DefaultGenericDF      = 8
	DefaultAuditMax       = -1
	DefaultAnnotateMax    = -1
	DefaultStruckRows     = 20
	DefaultMaxUnarchived  = 15
	DefaultOldestDays     = 21
	DefaultGateTimeoutSec = 300
)

// The battery's gate names — the whole valid skip_gates vocabulary.
// Single source of truth: battery.go names its matrix entries with these
// constants, and validate refuses anything outside the list, so a typo'd
// skip_gates entry is a loud exit 2 instead of a silent no-op.
const (
	GateFreshness = "freshness"
	GateAuditAll  = "audit-all"
	GateAnnotate  = "annotate"
	GateRouting   = "routing"
	GateCadence   = "cadence"
)

// KnownGates lists every battery gate name skip_gates may name. A
// function, not a var: the vocabulary must not be mutable state.
func KnownGates() []string {
	return []string{GateFreshness, GateAuditAll, GateAnnotate, GateRouting, GateCadence}
}

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
	ReportDir          string   `json:"report_dir"`           // the unarchived status-report directory
	ReportGlob         string   `json:"report_glob"`          // filename pattern inside ReportDir and its archived subdir
	ArchivedSubdir     string   `json:"archived_subdir"`      // archive subdir UNDER ReportDir
	LivingDocs         []string `json:"living_docs"`          // the harvest corpus (freshness/routing)
	AnnotateExtraGlobs []string `json:"annotate_extra_globs"` // extra annotate corpus beyond reports (planning dirs)
	RuleF              bool     `json:"rule_f"`               // docs-annotate rule F (inline-code spacing) on/off
	RuleFExtensions    []string `json:"rule_f_extensions"`    // file extensions rule F scans (dot-prefixed)
	GenericDF          int      `json:"generic_df"`           // the distinctive-token bar (a token on >= N living lines is generic)
	AuditMax           int      `json:"audit_max"`            // docs-freshness --audit-max ratchet pin (-1 = strict)
	AnnotateMax        int      `json:"annotate_max"`         // docs-annotate adoption ratchet pin (-1 = strict)
	TodoFile           string   `json:"todo_file"`            // the TODO_LIST file (cadence reads it, docs-strike edits it)
	Cadence            Cadence  `json:"cadence"`
	GateTimeoutSec     int      `json:"gate_timeout_sec"` // battery per-gate wall-clock cap in seconds (0 = no timeout)
	SkipGates          []string `json:"skip_gates"`       // battery verbs to skip (docs-gates run)
}

// Defaults returns the shipped configuration — the exact constants the
// gates carried before the config file existed (pbx-artmann's proven
// values). A repo without .docs-gates.json gets precisely the historical
// behavior.
func Defaults() Config {
	return Config{
		ReportDir:          "docs/status",
		ReportGlob:         "2*.md",
		ArchivedSubdir:     "archived",
		LivingDocs:         []string{"TODO_LIST.md", "ROADMAP.md", "CHANGELOG.md"},
		AnnotateExtraGlobs: []string{"docs/planning/*.md", "docs/planning/archived/*.md"},
		RuleF:              true,
		RuleFExtensions:    []string{".md", ".sh", ".nix"},
		GenericDF:          DefaultGenericDF,
		AuditMax:           DefaultAuditMax,
		AnnotateMax:        DefaultAnnotateMax,
		TodoFile:           "TODO_LIST.md",
		Cadence: Cadence{
			StruckRows:    DefaultStruckRows,
			MaxUnarchived: DefaultMaxUnarchived,
			OldestDays:    DefaultOldestDays,
		},
		GateTimeoutSec: DefaultGateTimeoutSec,
		SkipGates:      nil,
	}
}

// OrSlice is cmp.Or for slices (which are not comparable): the first
// argument carrying elements wins. Every Settings projection's list
// fallback rides this, so "a zero-value list falls back to the shipped
// Defaults" has one spelling across the gates.
func OrSlice[T any](override, def []T) []T {
	if len(override) > 0 {
		return override
	}

	return def
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

		return cfg, nil, newLoadError(loadRead, Path(root), err)
	}

	// Unknown-key pass: decode into a key map first so a typo warns
	// instead of silently doing nothing.

	var raw map[string]json.RawMessage

	if err := json.Unmarshal(data, &raw); err != nil {
		return cfg, nil, newLoadError(loadParse, Path(root), err)
	}

	known := knownKeys()

	var warnings []string

	for _, key := range sortedKeys(raw) {
		if !known[key] {
			warnings = append(warnings, fmt.Sprintf("unknown config key %q ignored", key))
		}
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, warnings, newLoadError(loadParse, Path(root), err)
	}

	if err := cfg.validate(); err != nil {
		return cfg, warnings, newLoadError(loadValidate, Path(root), err)
	}

	return cfg, warnings, nil
}

// loadStep is the Load stage whose refusal produced a loadError: reading
// the file, parsing its JSON, or validating the decoded values. The zero
// value is no stage — the default String spelling makes an unset step
// loud instead of silently reading as the first stage.
type loadStep int

const (
	loadRead loadStep = iota + 1
	loadParse
	loadValidate
)

func (s loadStep) String() string {
	switch s {
	case loadRead:
		return "reading"
	case loadParse:
		return "parsing"
	case loadValidate:
		return "invalid"
	default:
		return fmt.Sprintf("loadStep(%d)", int(s))
	}
}

// loadError is the package's static error type for Load failures: the
// refusing step, the file it refused on, and the underlying cause. The
// message is the caller's FATAL line verbatim (historical
// "reading/parsing/invalid <path>: <cause>" spelling); Unwrap keeps
// errors.Is/errors.AsType classification of the cause reachable.
type loadError struct {
	step  loadStep
	path  string
	cause error
}

func (e *loadError) Error() string {
	return fmt.Sprintf("%s %s: %v", e.step, e.path, e.cause)
}

func (e *loadError) Unwrap() error { return e.cause }

// newLoadError builds one loadError with the FATAL-line message.
func newLoadError(step loadStep, path string, cause error) *loadError {
	return &loadError{step: step, path: path, cause: cause}
}

// invalidError is the package's static error type: every validate
// refusal is the same concrete type (user-facing CLI text; no errors.Is
// consumers by design, so no Unwrap).
type invalidError struct {
	msg string
}

func (e *invalidError) Error() string { return e.msg }

// invalidf builds one invalidError with the refusal message.
func invalidf(format string, args ...any) *invalidError {
	return &invalidError{msg: fmt.Sprintf(format, args...)}
}

// validate enforces the value contract beyond what JSON typing gives.
func (c Config) validate() error {
	if err := c.validatePaths(); err != nil {
		return err
	}

	if err := c.validateNumbers(); err != nil {
		return err
	}

	if c.TodoFile == "" {
		return invalidf("todo_file must not be empty")
	}

	if err := c.validateSkipGates(); err != nil {
		return err
	}

	return c.validateExtensions()
}

// validatePaths checks the directory/glob/list shape fields.
func (c Config) validatePaths() error {
	const Template = "%s must be a relative directory, got %q"

	if c.ReportDir == "" || filepath.IsAbs(c.ReportDir) {
		return invalidf(Template, "report_dir", c.ReportDir)
	}

	if c.ReportGlob == "" {
		return invalidf("report_glob must not be empty")
	}

	if c.ArchivedSubdir == "" || filepath.IsAbs(c.ArchivedSubdir) {
		return invalidf(Template, "archived_subdir", c.ArchivedSubdir)
	}

	if len(c.LivingDocs) == 0 {
		return invalidf("living_docs must list at least one file")
	}

	return nil
}

// validateNumbers checks the numeric threshold fields.
func (c Config) validateNumbers() error {
	if c.GenericDF < 1 {
		return invalidf("generic_df must be >= 1, got %d", c.GenericDF)
	}

	if c.AuditMax < -1 {
		return invalidf("audit_max must be >= -1 (-1 = strict), got %d", c.AuditMax)
	}

	if c.AnnotateMax < -1 {
		return invalidf("annotate_max must be >= -1 (-1 = strict), got %d", c.AnnotateMax)
	}

	if c.Cadence.StruckRows < 1 || c.Cadence.MaxUnarchived < 1 || c.Cadence.OldestDays < 1 {
		return invalidf("cadence thresholds must all be >= 1, got %+v", c.Cadence)
	}

	if c.GateTimeoutSec < 0 {
		return invalidf("gate_timeout_sec must be >= 0 (0 = no timeout), got %d", c.GateTimeoutSec)
	}

	return nil
}

// validateSkipGates refuses names outside the battery's gate set: the
// battery matches skip entries with slices.Contains and would never
// notice a typo, silently running a gate the config meant to drop.
func (c Config) validateSkipGates() error {
	for _, name := range c.SkipGates {
		if name == "" {
			return invalidf("skip_gates entries must not be empty; valid gates: %s", strings.Join(KnownGates(), ", "))
		}

		if !slices.Contains(KnownGates(), name) {
			return invalidf("skip_gates: unknown gate %q; valid gates: %s", name, strings.Join(KnownGates(), ", "))
		}
	}

	return nil
}

// validateExtensions checks the rule-F file-extension list.
func (c Config) validateExtensions() error {
	for _, ext := range c.RuleFExtensions {
		if len(ext) < 2 || ext[0] != '.' {
			return invalidf("rule_f_extensions entries must be dot-prefixed extensions, got %q", ext)
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
		"annotate_max":         true,
		"todo_file":            true,
		"cadence":              true,
		"gate_timeout_sec":     true,
		"skip_gates":           true,
	}
}

func sortedKeys(m map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	return keys
}
