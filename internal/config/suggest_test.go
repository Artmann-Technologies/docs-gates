package config

import (
	"strings"
	"testing"
)

// TestNearestKeySuggestsWithinTwo pins the did-you-mean bar: a typo
// within edit distance 2 of exactly one known key suggests it.
func TestNearestKeySuggestsWithinTwo(t *testing.T) {
	t.Parallel()

	known := []string{"audit_corpus", "annotate_corpus", "audit_max", "generic_df"}

	got, ok := NearestKey("audit_corp", known, 2)
	if !ok || got != "audit_corpus" {
		t.Fatalf("audit_corp must suggest audit_corpus, got %q ok=%v", got, ok)
	}

	if _, ok := NearestKey("zzzzzzzz", known, 2); ok {
		t.Fatal("a word within distance 2 of nothing must refuse")
	}
}

// TestNearestKeyRefusesTies pins the tie rule: two known keys at the
// same minimal distance suggest NOTHING — guessing between two keys is
// how wrong keys get "fixed".
func TestNearestKeyRefusesTies(t *testing.T) {
	t.Parallel()

	known := make([]string, 0, 3)
	known = append(known, "cat", "hat")

	if got, ok := NearestKey("bat", known, 2); ok {
		t.Fatalf("bat sits at distance 1 from cat AND hat — must refuse, suggested %q", got)
	}

	// A third, closer key breaks the tie and suggests itself.
	known = append(known, "bat")

	if got, ok := NearestKey("bat", known, 2); !ok || got != "bat" {
		t.Fatalf("the exact key must win the tie-break, got %q ok=%v", got, ok)
	}
}

// TestLoadWarnsUnknownKeyWithSuggestion walks a real config file with a
// typo'd key through Load: the warning names the nearest real key.
func TestLoadWarnsUnknownKeyWithSuggestion(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	writeConfig(t, dir, `{"audit_corp": "all"}`)

	_, warnings, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(warnings) != 1 {
		t.Fatalf("one unknown key must yield one warning, got %v", warnings)
	}

	if !strings.Contains(warnings[0], `did you mean "audit_corpus"?`) {
		t.Fatalf("the warning must suggest the nearest key, got %q", warnings[0])
	}
}
