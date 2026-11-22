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
	DefaultAuditCorpus    = "all"
	DefaultAnnotateMax    = -1
	DefaultAnnotateCorpus = "all"
	DefaultStruckRows     = 20
	DefaultMaxUnarchived  = 15
	DefaultOldestDays     = 21
	DefaultGateTimeoutSec = 300
	DefaultStampWarnDays  = 90
	DefaultStampFailDays  = 180
)

// The battery's gate names — the whole valid skip_gates vocabulary.
// Single source of truth: battery.go names its matrix entries with these
// constants, and validate refuses anything outside the list, so a typo'd
// skip_gates entry is a loud exit 2 instead of a silent no-op.
const (
	GateFreshness  = "freshness"
	GateAuditAll   = "audit-all"
	GateAnnotate   = "annotate"
	GateRouting    = "routing"
	GateCadence    = "cadence"
	GateStub       = "stub"
	GateCitations  = "citations"
	GateStamps     = "stamps"
	GateVocabulary = "vocabulary"
)

// KnownGates lists every battery gate name skip_gates may name. A
// function, not a var: the vocabulary must not be mutable state.
func KnownGates() []string {
	return []string{
		GateFreshness,
		GateAuditAll,
		GateAnnotate,
		GateRouting,
		GateCadence,
		GateStub,
		GateCitations,
		GateStamps,
		GateVocabulary,
	}
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
	AuditCorpus        string   `json:"audit_corpus"`         // freshness replay scope: "all" (default) or "unarchived"
	AnnotateMax        int      `json:"annotate_max"`         // docs-annotate adoption ratchet pin (-1 = strict)
	AnnotateCorpus     string   `json:"annotate_corpus"`      // annotate corpus scope: "all" (default) or "unarchived"
	TodoFile           string   `json:"todo_file"`            // the TODO_LIST file (cadence reads it, docs-strike edits it)
	Cadence            Cadence  `json:"cadence"`
	GateTimeoutSec     int      `json:"gate_timeout_sec"` // battery per-gate wall-clock cap in seconds (0 = no timeout)
	SkipGates          []string `json:"skip_gates"`       // battery verbs to skip (docs-gates run)
	StubFiles          []string `json:"stub_files"`       // corpus the stub gate checksums (docs-stub)
	StubChecksums      []string `json:"stub_checksums"`   // bare-template digests; a match is an unfilled stub (md5 or sha256 hex)
	StubExempts        []string `json:"stub_exempts"`     // repo-relative files exempt from the stub gate (the template host, honest stubs)
	StubEnforce        bool     `json:"stub_enforce"`     // stub gate rollout: false = findings WARN and exit 0; true = FAIL
	CitationFiles      []string `json:"citation_files"`   // corpus the citations gate resolves (docs-citations)
	CitationExts       []string `json:"citation_exts"`    // file extensions counted as code citations (dot-prefixed)
	CitationEnforce    bool     `json:"citation_enforce"` // citations rollout: false = WARN, true = FAIL on missing targets
	StampFiles         []string `json:"stamp_files"`      // corpus the stamps gate dates (docs-stamps)
	StampWarnDays      int     `json:"stamp_warn_days"`  // verification stamps older than this WARN
	StampFailDays      int     `json:"stamp_fail_days"`  // verification stamps older than this FAIL
	StampEnforce       bool    `json:"stamp_enforce"`    // stamps rollout: false = findings WARN and exit 0; true = FAIL rows fail
	VocabularyFiles    []string `json:"vocabulary_files"`   // corpus the vocabulary gate judges (docs-vocabulary)
	VocabularyEnforce  bool    `json:"vocabulary_enforce"` // vocabulary rollout: false = findings WARN and exit 0; true = FAIL
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
		return cfg, nil, newLoadError(loadParse, Path(root), withSyntaxLocation(data, err))
	}

	known := knownKeys()

	var warnings []string

	for _, key := range sortedKeys(raw) {
		if !known[key] {
			warnings = append(warnings, fmt.Sprintf("unknown config key %q ignored", key))
		}
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, warnings, newLoadError(loadParse, Path(root), withSyntaxLocation(data, err))
	}

	if err := cfg.validate(); err != nil {
		return cfg, warnings, newLoadError(loadValidate, Path(root), err)
	}

	return cfg, warnings, nil
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

	if err := c.validateCorpus(); err != nil {
		return err
	}

	if err := c.validateExtensions(); err != nil {
		return err
	}

	return c.validateContentGates()
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

// validateCorpus refuses an audit_corpus or annotate_corpus value
// outside the scope vocabulary: the corpus builders string-compare the
// setting and would silently treat a typo ("unarchved") as the default
// all-corpus scan — exactly the silent-wrong-threshold class the
// validate contract exists for.
func (c Config) validateCorpus() error {
	switch c.AuditCorpus {
	case "", "all", "unarchived":
	default:
		return invalidf("audit_corpus must be \"all\" or \"unarchived\", got %q", c.AuditCorpus)
	}

	switch c.AnnotateCorpus {
	case "", "all", "unarchived":
		return nil
	default:
		return invalidf("annotate_corpus must be \"all\" or \"unarchived\", got %q", c.AnnotateCorpus)
	}
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
