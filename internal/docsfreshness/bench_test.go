package docsfreshness

import (
	"fmt"
	"strings"
	"testing"
)

// synthBenchCorpus builds a deterministic CV-shaped corpus for the
// audit-all cliff benchmark: reportsN reports each carrying itemsPer
// numbered §f items whose tokens draw from a vocabN-word pool, replayed
// against livingN living-doc lines. Every third living line spells one
// of its anchors with spaces instead of hyphens, so the squash arm of
// the presence match carries realistic weight. The vocabulary/placement
// arithmetic puts per-token document frequencies around the GenericDF
// bar, exercising both the distinctive and the generic verdict paths.
func synthBenchCorpus(reportsN, itemsPer, livingN, vocabN int) (reports, living []string) {
	vocab := make([]string, vocabN)
	for i := range vocab {
		vocab[i] = fmt.Sprintf("tok-%04d", i)
	}

	living = make([]string, livingN)
	for i := range living {
		anchorA := vocab[(i*3)%vocabN]
		anchorB := vocab[(i*5+1)%vocabN]
		if i%4 == 0 {
			anchorA = strings.ReplaceAll(anchorA, "-", " ")
		}

		living[i] = "the living anchor " + anchorA + " carries " + anchorB + " evidence"
	}

	reports = make([]string, reportsN)
	for r := range reports {
		var body strings.Builder
		body.WriteString("# Synth\n\n## f) Next\n\n")

		for it := range itemsPer {
			t0 := vocab[(r*31+it*13)%vocabN]
			t1 := vocab[(r*31+it*13+7)%vocabN]
			t2 := vocab[(r*31+it*13+14)%vocabN]
			t3 := vocab[(r*31+it*13+21)%vocabN]

			fmt.Fprintf(&body, "%d. verify the %s relay binds %s and %s evidence for %s\n",
				it+1, t0, t1, t2, t3)
		}

		reports[r] = body.String()
	}

	return reports, living
}

// BenchmarkCheckReportReplay times the audit-all hot loop — one full
// replay of every report's §f items against the living corpus — at the
// fleet's largest observed corpus shape (CV, 2026-11-05: 1,643 reports
// × ~2,000 living lines).
func BenchmarkCheckReportReplay(b *testing.B) {
	reports, living := synthBenchCorpus(1600, 6, 2000, 600)
	squashed := squashLines(living)
	tokenDF := tokenDFOver(living)
	settings := DefaultSettings()

	b.ResetTimer()

	for range b.N {
		for _, report := range reports {
			CheckReport(report, living, tokenDF, squashed, settings)
		}
	}
}
