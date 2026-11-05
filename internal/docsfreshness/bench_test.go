package docsfreshness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// synthBenchCorpus writes a deterministic CV-shaped corpus for the
// audit-all cliff benchmark: reportsN report files each carrying
// itemsPer numbered §f items whose tokens draw from a vocabN-word pool,
// replayed against livingN living-doc lines. Every fourth living line
// spells one of its anchors with spaces instead of hyphens, so the
// squash arm of the presence match carries realistic weight, and the
// vocabulary/placement arithmetic puts per-token document frequencies
// around the GenericDF bar, exercising both verdict paths.
func synthBenchCorpus(tb testing.TB, reportsN, itemsPer, livingN, vocabN int) ([]string, []string) {
	tb.Helper()

	vocab := make([]string, 0, vocabN)
	for i := range vocabN {
		vocab = append(vocab, fmt.Sprintf("tok-%04d", i))
	}

	living := make([]string, 0, livingN)
	for i := range livingN {
		anchorA := vocab[(i*3)%vocabN]
		anchorB := vocab[(i*5+1)%vocabN]

		if i%4 == 0 {
			anchorA = strings.ReplaceAll(anchorA, "-", " ")
		}

		living = append(living, "the living anchor "+anchorA+" carries "+anchorB+" evidence")
	}

	reportPaths := make([]string, 0, reportsN)
	for repIdx := range reportsN {
		var body strings.Builder
		body.WriteString("# Synth\n\n## f) Next\n\n")

		for it := range itemsPer {
			t0 := vocab[(repIdx*31+it*13)%vocabN]
			t1 := vocab[(repIdx*31+it*13+7)%vocabN]
			t2 := vocab[(repIdx*31+it*13+14)%vocabN]
			t3 := vocab[(repIdx*31+it*13+21)%vocabN]

			fmt.Fprintf(&body, "%d. verify the %s relay binds %s and %s evidence for %s\n",
				it+1, t0, t1, t2, t3)
		}

		path := filepath.Join(tb.TempDir(), fmt.Sprintf("2026-01-%02d_00-00_SYNTH%04d.md", 1+repIdx%28, repIdx))
		if err := os.WriteFile(path, []byte(body.String()), 0o644); err != nil {
			tb.Fatal(err)
		}

		reportPaths = append(reportPaths, path)
	}

	return reportPaths, living
}

// TestSynthBenchCorpusProducesWork guards the benchmark's shape: a
// corpus whose §f sections extract zero items (or whose replay finds
// zero unharvested findings) would time a no-op — exactly the trap the
// first version of this benchmark fell into (ExtractSectionF takes a
// file PATH, and in-memory bodies read as unreadable files).
func TestSynthBenchCorpusProducesWork(t *testing.T) {
	t.Parallel()

	reports, living := synthBenchCorpus(t, 4, 6, 200, 60)

	items := ExtractItems(ExtractSectionF(reports[0]))
	if len(items) != 6 {
		t.Fatalf("want 6 §f items in the synthetic report, got %d", len(items))
	}

	findings := CheckReport(reports[0], NewCorpus(living), DefaultSettings())
	if len(findings) == 0 {
		t.Fatal("the synthetic corpus must produce unharvested findings, got none")
	}
}

// BenchmarkCheckReportReplay times the audit-all hot loop — one full
// replay of every report's §f items against the living corpus — at the
// fleet's largest observed corpus shape (CV, 2026-11-05: 1,643 reports
// × ~2,000 living lines).
func BenchmarkCheckReportReplay(b *testing.B) {
	reports, living := synthBenchCorpus(b, 1600, 6, 2000, 600)
	corpus := NewCorpus(living)
	settings := DefaultSettings()

	b.ResetTimer()

	for range b.N {
		for _, report := range reports {
			CheckReport(report, corpus, settings)
		}
	}
}
