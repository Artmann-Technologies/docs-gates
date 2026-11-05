package docsfreshness

import (
	"strings"
	"testing"
)

// naivePresent and naiveDF are the pre-memo reference implementations:
// full rescans per call, no caching. The equivalence test holds the
// memoized Corpus against them — the parity goldens pin whole-gate
// output, this pins the cache itself.
func naivePresent(token string, lines, squashed []string) bool {
	for _, line := range lines {
		if strings.Contains(line, token) {
			return true
		}
	}

	key := Squash(token)
	if len(key) >= MinTokenLen {
		for _, line := range squashed {
			if strings.Contains(line, key) {
				return true
			}
		}
	}

	return false
}

func naiveDF(token string, lines []string) int {
	n := 0

	for _, line := range lines {
		if strings.Contains(line, token) {
			n++
		}
	}

	return n
}

// TestCorpusMemoMatchesNaiveScans: every verdict the memoized Corpus
// returns must equal the naive rescan over the same lines — across
// present, absent, squash-arm-only, and generic-DF tokens, and across
// repeated calls (the cache-hit path). A memo that ever diverges would
// silently flip harvest verdicts while the output SHAPE stays valid.
func TestCorpusMemoMatchesNaiveScans(t *testing.T) {
	t.Parallel()

	_, living := synthBenchCorpus(t, 8, 4, 120, 30)
	living = append(living,
		"the webphone alpha-probe calibration line",
		"a spaced anchor for the real cax evidence",
	)

	squashed := squashLines(living)
	corpus := NewCorpus(living)

	probes := []string{
		"webphone", "alpha-probe", "calibration", "real-cax", "evidence",
		"verify", "relay", "binds", "absent-zephyr-xyz", "tok-0000", "tok 0001", "anchor",
	}

	for _, line := range living {
		probes = append(probes, Tokens(line)...)
	}

	seen := map[string]bool{}

	for _, tok := range probes {
		if seen[tok] || tok == "" {
			continue
		}

		seen[tok] = true

		if got, want := corpus.Contains(tok), naivePresent(tok, living, squashed); got != want {
			t.Fatalf("Contains(%q) = %v, naive scan says %v", tok, got, want)
		}

		if got, want := corpus.Count(tok), naiveDF(tok, living); got != want {
			t.Fatalf("Count(%q) = %d, naive scan says %d", tok, got, want)
		}

		// The cache-hit path must return the identical verdict.
		if got, want := corpus.Contains(tok), naivePresent(tok, living, squashed); got != want {
			t.Fatalf("second Contains(%q) = %v diverged to %v", tok, got, want)
		}
	}
}

// TestCorpusMemoPopulatesAndSurvivesValueCopies: the caches fill as a
// side effect of queries, and because Corpus flows by value through the
// gate helpers, a copy must share the warmed cache (maps are reference
// types) rather than silently starting cold or diverging.
func TestCorpusMemoPopulatesAndSurvivesValueCopies(t *testing.T) {
	t.Parallel()

	corpus := NewCorpus([]string{"the webphone alpha-probe line"})

	if _, cached := corpus.present["webphone"]; cached {
		t.Fatal("a fresh corpus must start with an empty present cache")
	}

	corpus.Contains("webphone")

	if _, cached := corpus.present["webphone"]; !cached {
		t.Fatal("Contains must populate the present cache")
	}

	corpus.Count("webphone")

	if _, cached := corpus.df["webphone"]; !cached {
		t.Fatal("Count must populate the df cache")
	}

	copied := corpus // value copy: same map backing
	if !copied.Contains("webphone") {
		t.Fatal("a value copy must see the warmed cache")
	}

	if copied.Count("webphone") != corpus.Count("webphone") {
		t.Fatal("a value copy must agree on cached DF")
	}
}

// TestCorpusZeroValueLazyInitsCaches: a zero-value Corpus (no NewCorpus)
// must still answer Contains/Count without a nil-map write panic —
// defensive laziness, since the memo fields are unexported.
func TestCorpusZeroValueLazyInitsCaches(t *testing.T) {
	t.Parallel()

	var corpus Corpus

	if corpus.Contains("anything") {
		t.Fatal("a zero corpus carries no lines, nothing can be present")
	}

	if corpus.Count("anything") != 0 {
		t.Fatal("a zero corpus carries no lines, DF must be 0")
	}
}
