package docspurge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// TestSplitStrikeToleratesSwallowedPipe pins the T4 collector fix: a
// strike whose wrap swallowed the leading pipe (`~~| task | cells |~~
// verdict`) must draft a bullet from the TASK TEXT, not from the raw
// pipe-laden span — the pre-fix draft read the whole row as one pipe
// string and a CHANGELOG bullet rendered as a broken table row.
func TestSplitStrikeToleratesSwallowedPipe(t *testing.T) {
	t.Parallel()

	line := "~~| Write the note: the sweep broke the hash, the remedy is documented | 🔴 `TODO` | P3 | S | evidence cell |~~ done by reference: the note landed"

	row, ok := splitStrike(line, 7)
	if !ok {
		t.Fatal("a pipe-swallowed strike must collect")
	}

	if strings.Contains(row.task, "|") {
		t.Errorf("the task must carry no stray pipe: %q", row.task)
	}

	if !strings.Contains(row.task, "Write the note") {
		t.Errorf("the task must be the task text: %q", row.task)
	}

	if row.verdict != "done by reference: the note landed" {
		t.Errorf("the verdict must come from the wrap tail: %q", row.verdict)
	}

	if row.lineNo != 7 {
		t.Errorf("line number must survive, got %d", row.lineNo)
	}
}

// TestDraftBulletsStayPipeFree walks a whole corrupted-row corpus through
// the collector: no drafted bullet may contain a table pipe.
func TestDraftBulletsStayPipeFree(t *testing.T) {
	t.Parallel()

	body := "## 1) Board\n\n" +
		"~~| Struck before the pipe | 🔴 `TODO` | P3 | S | ev |~~ verdict one\n" +
		"| ~~healthy strike~~ | ✅ DONE | P2 | M | ev two |\n"

	rows, guards := collectStruck(body)
	if len(rows) != 2 {
		t.Fatalf("both strikes must collect (guards: %v)", guards)
	}

	for i, b := range draftBullets(rows) {
		if strings.Contains(b, "|") {
			t.Errorf("bullet %d carries a stray pipe: %q", i, b)
		}
	}
}

// TestApplyPurgeWritesChangelogBeforeTodo pins the write order: the
// bullets land in the CHANGELOG FIRST, the TODO rows leave SECOND — a
// crash between the writes must never leave the board purged with the
// verdicts lost (the two-write window keeps the evidence side).
func TestApplyPurgeWritesChangelogBeforeTodo(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	todo := filepath.Join(dir, "TODO_LIST.md")
	changelog := filepath.Join(dir, "CHANGELOG.md")

	body := "# TODO\n\n| Task | Status |\n| ---- | ------ |\n| ~~done thing~~ | ✅ DONE |\n"
	if err := os.WriteFile(todo, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(changelog, []byte("# CHANGELOG\n\n## [Unreleased]\n\n### Changed\n\n- existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := Options{Root: dir, File: "TODO_LIST.md", Changelog: "CHANGELOG.md", Apply: true}

	rows, _ := collectStruck(body)

	var order []string

	applyPurgeWith(func(path, _ string) error {
		order = append(order, filepath.Base(path))

		if len(order) == 2 {
			return os.ErrPermission // the SECOND write fails
		}

		return nil
	}, opts, todo, body, rows, draftBullets(rows))

	if len(order) == 0 || order[0] != "CHANGELOG.md" {
		t.Fatalf("the CHANGELOG must be written first, got %v", order)
	}

	if len(order) < 2 || order[1] != "TODO_LIST.md" {
		t.Fatalf("the TODO file must be written second, got %v", order)
	}

	onDisk, err := os.ReadFile(todo)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(onDisk), "~~done thing~~") {
		t.Errorf("a failed second write must leave the board unpurged:\n%s", onDisk)
	}
}

// TestApplyPurgeStillPurgesWhenBothWritesSucceed pins the happy path
// through the same seam: rows leave the TODO, bullets join the
// CHANGELOG, exit clean.
func TestApplyPurgeStillPurgesWhenBothWritesSucceed(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	todo := filepath.Join(dir, "TODO_LIST.md")
	changelog := filepath.Join(dir, "CHANGELOG.md")

	body := "# TODO\n\n| Task | Status |\n| ---- | ------ |\n| ~~done thing~~ | ✅ DONE |\n"
	if err := os.WriteFile(todo, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(changelog, []byte("# CHANGELOG\n\n## [Unreleased]\n\n### Changed\n\n- existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := Options{Root: dir, File: "TODO_LIST.md", Changelog: "CHANGELOG.md", Apply: true}

	rows, _ := collectStruck(body)

	captured := gatecli.CaptureOutput(func() int {
		return applyPurge(opts, todo, body, rows, draftBullets(rows))
	})

	if captured.Code != gatecli.ExitClean {
		t.Fatalf("happy path must exit clean, got %d: %s", captured.Code, captured.Stderr)
	}

	onDisk, err := os.ReadFile(todo)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(onDisk), "done thing") {
		t.Errorf("the struck row must leave the TODO file:\n%s", onDisk)
	}
}
