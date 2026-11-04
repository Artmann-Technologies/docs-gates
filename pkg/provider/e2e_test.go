package provider

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
)

// The e2e leg's skip arms are process-environment facts (nix provider
// sandbox, bare CI runner), so the test would otherwise vanish into a
// green "ok" line and coverage loss would be invisible. The leg records
// what happened in a ledger, and the sentinel below independently probes
// the environment and FAILS on any unaccounted state — a skip that is
// not one of the two sanctioned reasons turns CI red instead of
// silently dropping the only test that exercises the real binary.
var (
	e2eRan      bool
	e2eSkipNote string
)

// e2eLegVerdict is the sentinel's decision table: "ran" when the
// environment allows the leg and it ran; "skipped" when the environment
// forbids it and the leg recorded exactly that; "UNACCOUNTED" for every
// other combination — the CI-visible hole (a skipped-for-unknown-reason
// leg, a leg that ran where it could not, or a leg deleted outright).
func e2eLegVerdict(rootAdjacent, goOnPath, ran bool, skipNote string) string {
	canRun := rootAdjacent && goOnPath

	switch {
	case canRun && ran:
		return "ran"
	case !canRun && skipNote != "":
		return "skipped"
	default:
		return "UNACCOUNTED"
	}
}

// TestDetectEndToEndOverRealBinary is the provider-path battery run: it
// builds the REAL cmd/docs-gates from the root module, points Detect at a
// fixture repo whose battery is red in exactly one gate (freshness alone
// runs, over a report whose open item harvests nowhere), and asserts
// Detect yields exactly one finding naming that gate — the unit tests
// only ever exercise canned matrix text.
//
// Skips when the root module source is not adjacent (the nix provider
// check sandbox contains only pkg/provider) or no go toolchain is on
// PATH. Not parallel: t.Setenv is process-global.
func TestDetectEndToEndOverRealBinary(t *testing.T) {
	rootCmd := "../../cmd/docs-gates"

	if _, err := os.Stat(rootCmd); err != nil {
		e2eSkipNote = "root module source not adjacent (" + rootCmd + " missing) — nix check sandbox"
		t.Skipf("%s", e2eSkipNote)
	}

	if _, err := exec.LookPath("go"); err != nil {
		e2eSkipNote = "go toolchain not on PATH"
		t.Skipf("%s", e2eSkipNote)
	}

	e2eRan = true

	binDir := t.TempDir()
	bin := filepath.Join(binDir, toolName)

	build := exec.CommandContext(context.Background(), "go", "build", "-o", bin, "./cmd/docs-gates")
	build.Dir = "../.."

	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/docs-gates: %v\n%s", err, out)
	}

	repo := t.TempDir()

	config := `{"report_dir": "docs/status", "living_docs": ["NOTES.md"], "skip_gates": ["audit-all", "annotate", "routing", "cadence"]}`

	if err := os.WriteFile(filepath.Join(repo, ".docs-gates.json"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(repo, "docs", "status"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Harvest red, date-independent: the §f item's distinctive token
	// (qxvorlag) appears in no living doc, so freshness fails regardless
	// of report age. Zero reports would be a vacuous PASS.
	report := "# red fixture\n\n## f) open items\n\n1. the qxvorlag spindle reconciliation needs a pass\n"

	if err := os.WriteFile(
		filepath.Join(repo, "docs", "status", "2026-11-04_00-00_red.md"),
		[]byte(report),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(repo, "NOTES.md"),
		[]byte("# notes\n\nnothing about qxvorlag here\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	ctx := finding.WithWorkingDir(context.Background(), repo)

	findings, err := Provider.Detect.Detect(ctx)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("findings = %d (%v); want exactly 1 for the single red gate", len(findings), findings)
	}

	if findings[0].Rule != "freshness" {
		t.Fatalf("rule = %q; want freshness", findings[0].Rule)
	}
}

// The sentinel: probe the environment independently, cross-check the
// leg's ledger, and make an unaccounted state a FAILURE — the CI-visible
// form of "the e2e leg vanished". Verbose mode prints which of the two
// happened; the failure mode prints the full probe state.
//
//nolint:paralleltest // reads the e2e leg's ledger, serialized on purpose
func TestE2ELegNeverVanishesSilently(t *testing.T) {
	_, rootAdjacentErr := os.Stat("../../cmd/docs-gates")
	_, goErr := exec.LookPath("go")

	rootAdjacent := rootAdjacentErr == nil
	goOnPath := goErr == nil

	verdict := e2eLegVerdict(rootAdjacent, goOnPath, e2eRan, e2eSkipNote)

	switch verdict {
	case "ran":
		t.Logf("E2E LEG: ran against the real binary")
	case "skipped":
		t.Logf(
			"E2E LEG: skipped in this environment (%s) — prove it locally: (cd pkg/provider && go test -v -run TestDetectEndToEnd)",
			e2eSkipNote,
		)
	default:
		t.Fatalf("E2E LEG UNACCOUNTED: rootAdjacent=%v goOnPath=%v ran=%v skipNote=%q — "+
			"the end-to-end coverage silently vanished", rootAdjacent, goOnPath, e2eRan, e2eSkipNote)
	}
}

// The decision table's negative control: every combination except the
// two sanctioned states is UNACCOUNTED, so a new skip arm, a deleted
// leg, or a probe/ledger disagreement cannot stay invisible.
func TestE2ELegVerdictTable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		rootAdjacent bool
		goOnPath     bool
		ran          bool
		skipNote     string
		want         string
	}{
		{"env allows + leg ran", true, true, true, "", "ran"},
		{"env forbids (no root) + sanctioned skip", false, true, false, "nix sandbox", "skipped"},
		{"env forbids (no go) + sanctioned skip", true, false, false, "no go on PATH", "skipped"},
		{"env allows + leg skipped for unknown reason", true, true, false, "new guard", "UNACCOUNTED"},
		{"env allows + neither ran nor skipped", true, true, false, "", "UNACCOUNTED"},
		{"env forbids + leg claims ran anyway", false, false, true, "", "UNACCOUNTED"},
		{"env forbids + skip without a note", true, false, false, "", "UNACCOUNTED"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := e2eLegVerdict(tc.rootAdjacent, tc.goOnPath, tc.ran, tc.skipNote); got != tc.want {
				t.Fatalf("e2eLegVerdict(%v, %v, %v, %q) = %q, want %q",
					tc.rootAdjacent, tc.goOnPath, tc.ran, tc.skipNote, got, tc.want)
			}
		})
	}
}
