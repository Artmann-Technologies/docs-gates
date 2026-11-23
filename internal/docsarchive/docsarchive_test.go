package docsarchive

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// runGit runs one git subcommand inside the fixture repo; the context
// bounds it so a hung fixture git cannot wedge the suite.
func runGit(t *testing.T, root string, args ...string) error {
	t.Helper()

	return exec.CommandContext(
		context.Background(),
		"git",
		append([]string{"-C", root}, args...)...,
	).Run()
}

// writeReport seeds one status report into the fixture repo's report
// dir and stages it, so git mv can carry it (git mv needs an index
// entry; the docsannotate flake-source lesson, same class).
func writeReport(t *testing.T, root, name, sectionF string) {
	t.Helper()

	dir := filepath.Join(root, "docs", "status")

	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}

	report := "# Status Report — fixture\n\n" + sectionF

	if err := os.WriteFile(filepath.Join(dir, name), []byte(report), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := runGit(t, root, "add", filepath.Join("docs", "status", name)); err != nil {
		t.Fatal(err)
	}
}

// newFixtureRepo builds a git repo with the default report layout.
func newFixtureRepo(t *testing.T) string {
	t.Helper()

	root := t.TempDir()

	if err := runGit(t, root, "init", "-q"); err != nil {
		t.Skipf("no git available for the fixture: %v", err)
	}

	return root
}

const (
	resolvedSectionF = "## f) Next tasks\n\nall harvested\n"
	openSectionF     = "## f) Next tasks\n\n1. adopt the annotate_corpus knob fleet-wide\n"
)

//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunDryRunPlansWithoutWriting(t *testing.T) {
	root := newFixtureRepo(t)

	writeReport(t, root, "2026-01-01_10-00_resolved.md", resolvedSectionF)
	writeReport(t, root, "2026-01-02_10-00_open.md", openSectionF)

	out := gatecli.CaptureOutput(func() int { return Run([]string{root}) })

	if out.Code != gatecli.ExitClean {
		t.Fatalf("dry-run over a plannable corpus exits 0, got %d; stdout: %s", out.Code, out.Stdout)
	}

	if !strings.Contains(out.Stdout, "ARCHIVE docs/status/2026-01-01_10-00_resolved.md") {
		t.Errorf("the resolved report must plan as ARCHIVE, got: %s", out.Stdout)
	}

	if !strings.Contains(out.Stdout, "SKIP    docs/status/2026-01-02_10-00_open.md (1 open") {
		t.Errorf("the open report must plan as SKIP with its count, got: %s", out.Stdout)
	}

	if _, err := os.Stat(filepath.Join(root, "docs", "status", "2026-01-01_10-00_resolved.md")); err != nil {
		t.Fatalf("dry-run must not move the report: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "docs", "status", "archived")); !os.IsNotExist(err) {
		t.Error("dry-run must not create the archive dir")
	}
}

//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunApplyBannersAndMoves(t *testing.T) {
	root := newFixtureRepo(t)

	writeReport(t, root, "2026-01-01_10-00_resolved.md", resolvedSectionF)

	out := gatecli.CaptureOutput(func() int { return Run([]string{root, "--apply"}) })

	if out.Code != gatecli.ExitClean {
		t.Fatalf("apply over a fully-resolved corpus exits 0, got %d; stderr: %s", out.Code, out.Stderr)
	}

	archived := filepath.Join(root, "docs", "status", "archived", "2026-01-01_10-00_resolved.md")

	data, err := os.ReadFile(archived)
	if err != nil {
		t.Fatalf("the report must live in the archive after --apply: %v\nstdout: %s", err, out.Stdout)
	}

	if !strings.HasPrefix(string(data), bannerPrefix) {
		t.Errorf(
			"the archived report must open with the Live-remainder banner, got: %q",
			string(data[:min(80, len(data))]),
		)
	}

	if !strings.Contains(string(data), "# Status Report — fixture") {
		t.Error("the banner must be PREPENDED, not replace the report body")
	}

	if _, err := os.Stat(filepath.Join(root, "docs", "status", "2026-01-01_10-00_resolved.md")); !os.IsNotExist(err) {
		t.Error("the unarchived copy must be gone (git mv, not copy)")
	}
}

//nolint:paralleltest // process-global capture, serialized on purpose
func TestRunApplyNeverMovesOpenReports(t *testing.T) {
	root := newFixtureRepo(t)

	writeReport(t, root, "2026-01-02_10-00_open.md", openSectionF)

	out := gatecli.CaptureOutput(func() int { return Run([]string{root, "--apply"}) })

	if out.Code != gatecli.ExitClean {
		t.Fatalf("an all-skip corpus is exit 0, got %d; stderr: %s", out.Code, out.Stderr)
	}

	if !strings.Contains(out.Stdout, "nothing to archive (1 report(s) skipped") {
		t.Errorf("the skip summary must name the count, got: %s", out.Stdout)
	}

	if _, err := os.Stat(filepath.Join(root, "docs", "status", "2026-01-02_10-00_open.md")); err != nil {
		t.Fatalf("an open report must never move: %v", err)
	}
}

//nolint:paralleltest // process-global capture, serialized on purpose
func TestApplyRefusesDoubleBanner(t *testing.T) {
	root := newFixtureRepo(t)

	dir := filepath.Join(root, "docs", "status")

	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}

	already := banner(time.Now()) + "# Status Report — already archived\n"

	if err := os.WriteFile(filepath.Join(dir, "2026-01-01_10-00_done.md"), []byte(already), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := runGit(t, root, "add", "docs/status/2026-01-01_10-00_done.md"); err != nil {
		t.Fatal(err)
	}

	out := gatecli.CaptureOutput(func() int { return Run([]string{root, "--apply"}) })

	if out.Code != gatecli.ExitUsage {
		t.Fatalf("a second banner is a FATAL refusal (exit 2), got %d; stderr: %s", out.Code, out.Stderr)
	}

	if !strings.Contains(out.Stderr, "refusing to double-archive") {
		t.Errorf("the refusal must name the corruption class, got: %s", out.Stderr)
	}
}
