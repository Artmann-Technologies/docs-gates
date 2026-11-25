package config

// EditDistance is the Levenshtein distance between two lowercase words.
// The DP rows are append-built: length expressions the makezero lint
// cannot prove non-zero are exactly the indexing class it exists for.
// Shared by the verb did-you-mean (cmd dispatch) and the config
// unknown-key warning — one distance, one implementation.
func EditDistance(a, b string) int {
	prev := make([]int, 0, len(b)+1)

	for j := 0; j <= len(b); j++ {
		prev = append(prev, j)
	}

	for i := 1; i <= len(a); i++ {
		curr := make([]int, 0, len(b)+1)
		curr = append(curr, i)

		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}

			curr = append(curr, min(prev[j]+1, curr[j-1]+1, prev[j-1]+cost))
		}

		prev = curr
	}

	return prev[len(b)]
}

// NearestKey returns the known key nearest to want within max edits,
// with the tie rule: two distinct keys at the same minimal distance
// refuse the suggestion — guessing between `audit_corpus` and
// `annotate_corpus` is how wrong keys get "fixed". ok=false means no
// key sits within max, or the nearest two tie.
func NearestKey(want string, known []string, ceiling int) (string, bool) {
	best := ""
	bestDist := ceiling + 1
	tie := false

	for _, k := range known {
		d := EditDistance(want, k)

		switch {
		case d < bestDist:
			best, bestDist, tie = k, d, false
		case d == bestDist:
			tie = true
		}
	}

	if best == "" || bestDist > ceiling || tie {
		return "", false
	}

	return best, true
}
