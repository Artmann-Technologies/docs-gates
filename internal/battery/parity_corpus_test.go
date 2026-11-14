package battery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// materializeCorpus copies one corpus dir under testdata/parity into a
// fresh temp root, substituting the date token in paths and contents.
func materializeCorpus(t *testing.T, corpus, today string) string {
	t.Helper()

	src := filepath.Join("..", "..", "testdata", "parity", corpus)
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
// battery.Run for the run verb. stdin (when non-empty) is swapped in as
// the process stdin for the duration of the call — the --pre-flight
// candidate lines arrive that way.
