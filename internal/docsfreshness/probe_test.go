package docsfreshness

import (
	"fmt"
	"testing"
)

func TestProbeSynth(t *testing.T) {
	reports, living := synthBenchCorpus(4, 6, 2000, 600)
	section := ExtractSectionF(reports[0])
	items := ExtractItems(section)
	fmt.Printf("section=%q items=%d\n", section[:min(40, len(section))], len(items))
	squashed := squashLines(living)
	tokenDF := tokenDFOver(living)
	findings := CheckReport(reports[0], living, tokenDF, squashed, DefaultSettings())
	fmt.Printf("findings=%d first=%q\n", len(findings), func() string {
		if len(findings) == 0 {
			return ""
		}
		return findings[0][0]
	}())
	if len(items) == 0 || len(findings) == 0 {
		t.Fatal("synthetic corpus produced no work")
	}
}
