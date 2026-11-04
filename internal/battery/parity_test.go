package battery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/docsannotate"
	"github.com/Artmann-Technologies/docs-gates/internal/docscadence"
	"github.com/Artmann-Technologies/docs-gates/internal/docsfreshness"
	"github.com/Artmann-Technologies/docs-gates/internal/docsrouting"
	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// The golden parity corpus lives at the repo root under
// testdata/parity/corpus: a deliberately custom config (renamed report
// dir, TODO file, and living doc; a moved distinctive-token bar; tuned
// cadence thresholds) proving the config plumbing end to end, with one
// report whose single open item harvests cleanly. The {{DATE}} token in
// filenames and contents is substituted with the run date so ages stay
// stable; the goldens carry the same token.
//
// Re-pin the goldens after an INTENTIONAL output change:
//
//	PARITY_UPDATE=1 go test ./internal/battery/ -run TestParityGolden
//
// then diff them by hand — a golden bump is a contract change.

const parityDateToken = "{{DATE}}"

// parityGate is one pinned invocation.
type parityGate struct {
	name string
	args []string
}

func parityGates() []parityGate {
	return []parityGate{
		{name: "freshness", args: []string{"freshness"}},
		{name: "audit-all", args: []string{"freshness", "--audit-all", "--audit-max=0"}},
		{name: "annotate", args: []string{"annotate"}},
		{name: "routing", args: []string{"routing", "--strict"}},
		{name: "cadence", args: []string{"cadence"}},
		{name: "battery", args: []string{"run"}},
	}
}

// materializeCorpus copies testdata/parity/corpus into a fresh temp root,
// substituting the date token in paths and contents.
func materializeCorpus(t *testing.T, today string) string {
	t.Helper()

	src := filepath.Join("..", "..", "testdata", "parity", "corpus")
	dst := t.TempDir()

	return copyTree(t, src, dst, today)
}

func copyTree(t *testing.T, src, dst, today string) string {
	t.Helper()

	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("read corpus: %v", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, strings.ReplaceAll(entry.Name(), parityDateToken, today))

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0o755); err != nil {
				t.Fatal(err)
			}

			copyTree(t, srcPath, dstPath, today)

			continue
		}

		data, err := os.ReadFile(srcPath)
		if err != nil {
			t.Fatalf("read corpus file %s: %v", srcPath, err)
		}

		out := strings.ReplaceAll(string(data), parityDateToken, today)

		if err := os.WriteFile(dstPath, []byte(out), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return dst
}

// runCaptured execs one verb: the gate Run functions directly,
// battery.Run for the run verb.
func runCaptured(args []string) gatecli.CapturedOutput {
	return gatecli.CaptureOutput(func() int {
		if len(args) == 0 {
			return gatecli.ExitUsage
		}

		switch args[0] {
		case "freshness":
			return docsfreshness.Run(args[1:])
		case "annotate":
			return docsannotate.Run(args[1:])
		case "routing":
			return docsrouting.Run(args[1:])
		case "cadence":
			return docscadence.Run(args[1:])
		case "run":
			return Run(args[1:])
		}

		return gatecli.ExitUsage
	})
}

//nolint:paralleltest // process-global capture, serialized on purpose
func TestParityGolden(t *testing.T) {
	update := os.Getenv("PARITY_UPDATE") != ""

	today := time.Now().Format("2006-01-02")
	root := materializeCorpus(t, today)

	for _, gate := range parityGates() {
		full := append(append([]string{}, gate.args...), root) // verb-first argv, root last

		out := runCaptured(full)

		goldenPath := filepath.Join("..", "..", "testdata", "parity", "golden", gate.name+".txt")
		golden := strings.ReplaceAll(readOrEmpty(t, goldenPath), parityDateToken, today)

		if update {
			raw := strings.ReplaceAll(out.Stdout+out.Stderr, today, parityDateToken)

			if err := os.WriteFile(goldenPath, []byte(raw), 0o644); err != nil {
				t.Fatalf("write golden %s: %v", goldenPath, err)
			}

			continue
		}

		if out.Code != gatecli.ExitClean {
			t.Fatalf("%s exited %d over the golden corpus; stderr: %s", gate.name, out.Code, out.Stderr)
		}

		if out.Stdout+out.Stderr != golden {
			t.Fatalf("%s output drifted from the golden:\n--- want ---\n%s\n--- got ---\n%s",
				gate.name, golden, out.Stdout+out.Stderr)
		}
	}
}

func readOrEmpty(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return ""
	}

	if err != nil {
		t.Fatalf("read golden %s: %v", path, err)
	}

	return string(data)
}
