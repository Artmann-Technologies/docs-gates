package docsvocabulary

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/config"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// featuresFixture is the go-verdict shape the issue's prototype measured
// against: a legend table, a prose extension defining OUT_OF_SCOPE, a
// feature table with a Status column, and a Notes column carrying an
// ALL-CAPS token that must NEVER flag (the measured false-positive
// class: whole-line scanning flagged `JSON`).
const featuresFixture = `# Features

## Status legend

| Status                    | Meaning            |
| ------------------------- | ------------------ |
| 🟢 ` + "`FULLY_FUNCTIONAL`" + `     | Works as intended. |
| 🟡 ` + "`PARTIALLY_FUNCTIONAL`" + ` | Ships with gaps.   |

Features absent by design carry ` + "`OUT_OF_SCOPE`" + ` — no code, ever.

## Inventory

| Feature        | Status                   | Notes                |
| -------------- | ------------------------ | -------------------- |
| Verdict engine | ` + "`FULLY_FUNCTIONAL`" + `        | Speaks ` + "`JSON`" + ` over stdin |
| CLI renderer   | ` + "`PARTIALLY_FUNCTIONAL`" + `    | No color support yet |
| Absent thing   | ` + "`OUT_OF_SCOPE`" + `            | Deliberate           |
`

func writeFeatures(t *testing.T, body string) string {
	t.Helper()

	root := t.TempDir()

	if err := os.WriteFile(filepath.Join(root, "FEATURES.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	return root
}

func TestScanCleanCorpusHasNoFindings(t *testing.T) {
	t.Parallel()

	findings, infos, skips, err := Scan(writeFeatures(t, featuresFixture), Settings{
		Files:   []string{"FEATURES.md"},
		Enforce: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(findings) != 0 {
		t.Fatalf("clean corpus must have no findings: %+v", findings)
	}

	if len(skips) != 0 {
		t.Fatalf("a judged file is not a skip: %v", skips)
	}

	if len(infos) != 0 {
		t.Fatalf("every legend token is used, infos = %+v", infos)
	}
}

// TestScanFlagsUndefinedStatus is the issue's synthetic mutation: one
// row's status flips to IN_PROGRESS and the gate names exactly that.
func TestScanFlagsUndefinedStatus(t *testing.T) {
	t.Parallel()

	mutated := strings.Replace(
		featuresFixture,
		"| CLI renderer   | `PARTIALLY_FUNCTIONAL`    | No color support yet |",
		"| CLI renderer   | `IN_PROGRESS`            | No color support yet |",
		1,
	)

	findings, _, _, err := Scan(writeFeatures(t, mutated), Settings{Files: []string{"FEATURES.md"}})
	if err != nil {
		t.Fatal(err)
	}

	if len(findings) != 1 || findings[0].Token != "IN_PROGRESS" {
		t.Fatalf("exactly the undefined token must flag, got %+v", findings)
	}

	if findings[0].Line == 0 {
		t.Fatal("the finding must name the row's line")
	}
}

// TestScanIgnoresCapsOutsideTheStatusColumn pins the column-aware
// extraction: `JSON` in a Notes cell is legitimate prose, never a
// status token.
func TestScanIgnoresCapsOutsideTheStatusColumn(t *testing.T) {
	t.Parallel()

	findings, _, _, err := Scan(writeFeatures(t, featuresFixture), Settings{Files: []string{"FEATURES.md"}})
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range findings {
		if f.Token == "JSON" {
			t.Fatalf("Notes-column tokens must never flag: %+v", f)
		}
	}
}

// TestScanCountsProseLegendDefinitions: OUT_OF_SCOPE is defined only
// in prose under the legend heading — using it must stay clean, and a
// file whose legend is table-only must flag it.
func TestScanCountsProseLegendDefinitions(t *testing.T) {
	t.Parallel()

	tableOnly := strings.Replace(
		featuresFixture,
		"Features absent by design carry `OUT_OF_SCOPE` — no code, ever.\n",
		"",
		1,
	)

	findings, _, _, err := Scan(writeFeatures(t, tableOnly), Settings{Files: []string{"FEATURES.md"}})
	if err != nil {
		t.Fatal(err)
	}

	if len(findings) != 1 || findings[0].Token != "OUT_OF_SCOPE" {
		t.Fatalf("without the prose definition OUT_OF_SCOPE must flag, got %+v", findings)
	}
}

func TestScanReportsUnusedLegendInformationally(t *testing.T) {
	t.Parallel()

	extended := strings.Replace(
		featuresFixture,
		"| 🟡 `PARTIALLY_FUNCTIONAL` | Ships with gaps.   |",
		"| 🟡 `PARTIALLY_FUNCTIONAL` | Ships with gaps.   |\n| 🔴 `BROKEN`               | Present but failing.|",
		1,
	)

	findings, infos, _, err := Scan(writeFeatures(t, extended), Settings{Files: []string{"FEATURES.md"}})
	if err != nil {
		t.Fatal(err)
	}

	if len(findings) != 0 {
		t.Fatalf("unused legend entries never fail: %+v", findings)
	}

	if len(infos) != 1 || infos[0].Token != "BROKEN" {
		t.Fatalf("the unused legend entry is informational, got %+v", infos)
	}
}

func TestScanSkipsFilesWithoutALegend(t *testing.T) {
	t.Parallel()

	root := writeFeatures(t, "# Features\n\n## DONE\n\n- prose sections, no tables\n")

	findings, _, skips, err := Scan(root, Settings{Files: []string{"FEATURES.md"}})
	if err != nil {
		t.Fatal(err)
	}

	if len(findings) != 0 || len(skips) != 1 {
		t.Fatalf("a legend-less FEATURES.md is a skip, not a finding: findings=%v skips=%v", findings, skips)
	}
}

//nolint:paralleltest // process-global capture swap, serialized on purpose
func TestRunWarnModeThenEnforce(t *testing.T) {
	root := writeFeatures(t, strings.Replace(
		featuresFixture,
		"| CLI renderer   | `PARTIALLY_FUNCTIONAL`    | No color support yet |",
		"| CLI renderer   | `IN_PROGRESS`            | No color support yet |",
		1,
	))

	captured := gatecli.CaptureOutput(func() int {
		return Run([]string{root})
	})

	if captured.Code != gatecli.ExitClean {
		t.Fatalf("warn mode stays green, code = %d:\n%s", captured.Code, captured.Stdout)
	}

	if !strings.Contains(captured.Stdout, "WARN FEATURES.md:") || !strings.Contains(captured.Stdout, "IN_PROGRESS") {
		t.Fatalf("the warn must name file and token:\n%s", captured.Stdout)
	}

	if err := os.WriteFile(
		filepath.Join(root, ".docs-gates.json"),
		[]byte(`{"vocabulary_enforce": true}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	captured = gatecli.CaptureOutput(func() int {
		return Run([]string{root})
	})

	if captured.Code != gatecli.ExitFindings {
		t.Fatalf("enforce mode fails on findings, code = %d", captured.Code)
	}
}

//nolint:paralleltest // process-global capture swap, serialized on purpose
func TestRunDogfoodRepoShapeSkipsClean(t *testing.T) {
	root := writeFeatures(t, "# Features\n\n## DONE\n\nThe inventory lives in prose sections.\n")

	captured := gatecli.CaptureOutput(func() int {
		return Run([]string{root})
	})

	if captured.Code != gatecli.ExitClean {
		t.Fatalf("code = %d:\n%s", captured.Code, captured.Stdout)
	}

	if !strings.Contains(captured.Stdout, "SKIP FEATURES.md") {
		t.Fatalf("a legend-less file must print the skip note:\n%s", captured.Stdout)
	}
}

func TestSettingsFromConfigKeepsDefaults(t *testing.T) {
	t.Parallel()

	settings := SettingsFromConfig(config.Defaults())

	if len(settings.Files) != 1 || settings.Files[0] != config.DefaultFeaturesFile {
		t.Fatalf("default corpus = %v, want [FEATURES.md]", settings.Files)
	}

	if settings.Enforce {
		t.Fatal("the rollout default is warn mode")
	}
}
