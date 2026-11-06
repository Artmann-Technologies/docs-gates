package battery

import (
	"os"
	"path/filepath"
	"regexp"
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

// parityGate is one pinned invocation; corpus names the corpus dir under
// testdata/parity (default "corpus") the args run over, stdin feeds the
// verb's standard input (the --pre-flight candidate lines), wantCode is
// the exit the gate owes over the corpus (default clean; the rewrite-mode
// and refusal arms are SUPPOSED to be non-clean — the bytes are the
// contract), argsAfterRoot rides AFTER the root positional (the
// --dump-items report path is positional[1], behind the root).
type parityGate struct {
	name          string
	args          []string
	corpus        string
	stdin         string
	wantCode      int
	argsAfterRoot []string
}

// parityDurationRe collapses the battery's wall-time renderings (the
// per-gate heartbeat `done in 830ms` lines and the matrix `(in 2.4s)`
// columns) to a literal D token: the goldens pin where a duration
// appears, never how long a gate took — that value is machine-local.
// The two prefixes are deliberately exact (`done in ` / `(in `): a bare
// `in ` would also swallow gate counts like "finding(s) in 17 file(s)".
// The µs/ms renderings all start with a digit, so the class catches the
// whole sub-second FormatDuration range.
var parityDurationRe = regexp.MustCompile(`(done in |\(in )[0-9][0-9a-zµ.]*`)

// parityJSONDurationRe collapses the --json face's millisecond columns
// the same way: the JSON document's `duration_ms` value is machine-local.
var parityJSONDurationRe = regexp.MustCompile(`"duration_ms": \d+`)

// parityRootToken replaces the materialized corpus temp root in golden
// bytes, so path-carrying outputs (--dump-items echoes its argument) stay
// byte-stable across runs.
const parityRootToken = "<corpus>"

// preflightSafeCandidate is the SAFE-arm stdin fixture: brand-new tokens
// only, nothing already on disk, nothing pushed to the DF boundary.
const preflightSafeCandidate = "1. the zeppelin hull rivet audit needs scheduling\n"

// preflightReplaceOld / preflightReplaceCandidate are the rewrite-mode
// pair: the candidate rewords the living line the --replace target names,
// so the replay judges the new words alone — the item anchored by the old
// line's tokens loses its anchors and flips (wantCode findings), the
// discrimination the additions-only model cannot see.
const (
	preflightReplaceOld       = "the lighthouse lens calibration watchrow ships this cycle"
	preflightReplaceCandidate = "the beacon prism alignment matrix ships this cycle\n"
)

func parityGates() []parityGate {
	return []parityGate{
		{name: "freshness", args: []string{"freshness"}},
		{name: "audit-all", args: []string{"freshness", "--audit-all", "--audit-max=0"}},
		{name: "annotate", args: []string{"annotate"}},
		{name: "annotate-ratchet", args: []string{"annotate"}, corpus: "corpus-ratchet"},
		{name: "routing", args: []string{"routing", "--strict"}},
		{name: "cadence", args: []string{"cadence"}},
		{name: "cadence-warn", args: []string{"cadence"}, corpus: "corpus-warn"},
		{name: "margins", args: []string{"freshness", "--margins"}},
		{
			// The exclusion pin lives here, not over the main corpus: bar 2
			// (corpus config) puts the one-line living doc's every token in
			// the near-bar window, so the four item-claimed tokens
			// (lighthouse lens calibration watchrow) would headline the
			// report without claimedItemTokens — the golden names only the
			// unclaimed pair (cycle, ships).
			//
			// generic_df: 2 in this fixture's .docs-gates.json is
			// load-bearing for BOTH arms it serves: the rewrite-mode
			// --replace delta below (a one-line living corpus makes the
			// anchor-flip maximally visible) AND this margins arm. Touching
			// it re-pins preflight-replace AND margins-replace together.
			name:   "margins-replace",
			args:   []string{"freshness", "--margins"},
			corpus: "corpus-replace",
		},
		{
			name:          "dump-items",
			args:          []string{"freshness", "--dump-items"},
			argsAfterRoot: []string{"<corpus>/reports/{{DATE}}_10-00_ship-it.md"},
		},
		{
			name:          "dump-items-refusal",
			args:          []string{"freshness", "--dump-items"},
			argsAfterRoot: []string{"<corpus>/TODOS.md"},
			wantCode:      gatecli.ExitUsage,
		},
		{name: "preflight-safe", args: []string{"freshness", "--pre-flight"}, stdin: preflightSafeCandidate},
		{
			name:     "preflight-replace",
			args:     []string{"freshness", "--pre-flight", "--replace=" + preflightReplaceOld},
			stdin:    preflightReplaceCandidate,
			corpus:   "corpus-replace",
			wantCode: gatecli.ExitFindings,
		},
		{name: "battery", args: []string{"run"}, wantCode: gatecli.ExitClean},
		{name: "battery-json", args: []string{"run", "--json"}, wantCode: gatecli.ExitClean},
	}
}

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
func runCaptured(args []string, stdin string) gatecli.CapturedOutput {
	return gatecli.CaptureOutput(func() int {
		if stdin != "" {
			r, w, _ := os.Pipe()
			_, _ = w.WriteString(stdin)
			_ = w.Close()

			orig := os.Stdin
			os.Stdin = r

			defer func() { os.Stdin = orig; _ = r.Close() }()
		}

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

	roots := map[string]string{}

	for _, gate := range parityGates() {
		corpus := gate.corpus
		if corpus == "" {
			corpus = "corpus"
		}

		if roots[corpus] == "" {
			roots[corpus] = materializeCorpus(t, corpus, today)
		}

		root := roots[corpus]

		full := append(append([]string{}, gate.args...), root) // verb-first argv, root last

		for _, extra := range gate.argsAfterRoot {
			extra = strings.ReplaceAll(extra, parityRootToken, root)
			full = append(full, strings.ReplaceAll(extra, parityDateToken, today))
		}

		out := runCaptured(full, gate.stdin)

		// normalize puts a captured run into golden form: dates collapse to
		// the token, the ephemeral corpus root to <corpus>, and wall-time
		// renderings to D, so path- and timing-carrying outputs stay
		// byte-stable.
		normalize := func(s string) string {
			s = strings.ReplaceAll(s, today, parityDateToken)
			s = strings.ReplaceAll(s, root, parityRootToken)
			s = parityDurationRe.ReplaceAllString(s, `${1}D`)

			return parityJSONDurationRe.ReplaceAllString(s, `"duration_ms": D`)
		}

		goldenPath := filepath.Join("..", "..", "testdata", "parity", "golden", gate.name+".txt")

		if update {
			if err := os.WriteFile(goldenPath, []byte(normalize(out.Stdout+out.Stderr)), 0o644); err != nil {
				t.Fatalf("write golden %s: %v", goldenPath, err)
			}

			continue
		}

		if out.Code != gate.wantCode {
			t.Fatalf(
				"%s exited %d (want %d) over the golden corpus; stderr: %s",
				gate.name,
				out.Code,
				gate.wantCode,
				out.Stderr,
			)
		}

		if got, want := normalize(out.Stdout+out.Stderr), readOrEmpty(t, goldenPath); got != want {
			t.Fatalf("%s output drifted from the golden:\n--- want ---\n%s\n--- got ---\n%s",
				gate.name, want, got)
		}
	}

	// Discrimination arm: the pin must bite. corpus-ratchet carries two
	// unannotated strikes with annotate_max pinned at 2 in its config; a
	// stricter pin (1) is a REGRESSION exit 1, proving the ratchet arm
	// fails past the pin instead of always-green.
	out := runCaptured([]string{"annotate", "--annotate-max=1", roots["corpus-ratchet"]}, "")
	if out.Code != gatecli.ExitFindings || !strings.Contains(out.Stderr, "REGRESSION") {
		t.Fatalf("annotate-max=1 over the 2-finding corpus: code=%d stderr=%s, want REGRESSION exit %d",
			out.Code, out.Stderr, gatecli.ExitFindings)
	}

	// Discrimination arm, pre-flight: candidate lines already on disk are
	// REFUSED (exit 2) — a post-write replay would otherwise answer its
	// own additions and hide a flip as SAFE.
	out = runCaptured(
		[]string{"freshness", "--pre-flight", roots["corpus"]},
		"the lighthouse lens calibration watchrow ships this cycle\n",
	)
	if out.Code != gatecli.ExitUsage || !strings.Contains(out.Stderr, "ALREADY on disk") {
		t.Fatalf("pre-flight over an on-disk line: code=%d stderr=%s, want refusal exit %d",
			out.Code, out.Stderr, gatecli.ExitUsage)
	}

	// Discrimination arm, rewrite-mode: the SAME candidate that flips the
	// item under --replace (the pinned preflight-replace golden above) is
	// SAFE without the flag — the old line stays on disk and keeps its
	// anchors, proving the flip is the rewrite's delta and not the
	// candidate's own words.
	out = runCaptured(
		[]string{"freshness", "--pre-flight", roots["corpus-replace"]},
		preflightReplaceCandidate,
	)
	if out.Code != gatecli.ExitClean || !strings.Contains(out.Stdout, "SAFE") {
		t.Fatalf("the replace candidate without the flag must be SAFE: code=%d stderr=%s stdout=%s",
			out.Code, out.Stderr, out.Stdout)
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
