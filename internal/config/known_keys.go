package config

// The unknown-key machinery: the known-key vocabulary of
// .docs-gates.json and the sorted-key helper the decode warnings
// print. A key added to Config MUST be listed here — the config sprawl
// guard warns on exactly the difference between this map and the file.

import (
	"encoding/json"
	"slices"
)

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
		"audit_corpus":         true,
		"annotate_max":         true,
		"annotate_corpus":      true,
		"todo_file":            true,
		"cadence":              true,
		"gate_timeout_sec":     true,
		"skip_gates":           true,
		"stub_files":           true,
		"stub_checksums":       true,
		"stub_exempts":         true,
		"stub_enforce":         true,
		"citation_files":       true,
		"citation_exts":        true,
		"citation_enforce":     true,
		"stamp_files":          true,
		"stamp_warn_days":      true,
		"stamp_fail_days":      true,
		"stamp_enforce":        true,
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
