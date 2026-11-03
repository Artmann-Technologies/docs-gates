package docsfreshness

// These tests cover the token matching (ex tests/test_docs_freshness.py,
// item 8.4 of the docs-health list, 2026-09-25): a report's `real-cax` token
// must match a living-docs line saying "real CAX" — hyphen/space-insensitive.
// Before the fix, two evidence cells had to be hand-phrased to carry the
// hyphenated spelling, a manual workaround the gate was supposed to remove.

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

func TestSquashCollapsesHyphensAndSpaces(t *testing.T) {
	t.Parallel()

	if got := Squash("real-cax"); got != "realcax" {
		t.Fatalf("Squash(real-cax) = %q", got)
	}

	if got := Squash("real  CAX"); got != "realcax" {
		t.Fatalf("Squash(real  CAX) = %q", got)
	}

	if got := Squash("a - b -- c"); got != "abc" {
		t.Fatalf("Squash(a - b -- c) = %q", got)
	}
}

func writeReport(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()

	report := filepath.Join(dir, "2026-09-25_00-00_TEST.md")
	if err := os.WriteFile(report, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	return report
}

func TestHyphenTokenMatchesSpacedLine(t *testing.T) {
	t.Parallel()

	report := writeReport(t,
		"# Report\n\n## f) Next\n\n1. close the real-cax gap after the switch\n")
	living := []string{"after the switch, build on the real cax evidence line"}
	livingSquashed := []string{Squash(living[0])}

	// Mirrors Run(): matching is squash-augmented, genericity counts stay
	// on raw lines (the tuned basis).
	tokenDF := func(token string) int {
		n := 0

		for _, line := range living {
			if strings.Contains(line, token) {
				n++
			}
		}

		return n
	}

	findings := CheckReport(report, living, tokenDF, livingSquashed, DefaultSettings())
	if len(findings) != 0 {
		t.Fatalf("hyphen/space mismatch must not drift, got %v", findings)
	}
}

func TestFindHtmlBlindSpot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	status := filepath.Join(root, "docs", "status")
	if err := os.MkdirAll(status, 0o755); err != nil {
		t.Fatal(err)
	}

	if s := FindHtmlBlindSpot(root, DefaultSettings()); s.Count != 0 || s.Newest != "" || s.Outranks {
		t.Fatalf("empty status dir must yield a zero spot, got %+v", s)
	}

	write := func(name string) {
		t.Helper()

		if err := os.WriteFile(filepath.Join(status, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// The exact production shape (2026-10-07): an .md report plus an
	// OLDER .html companion, then a NEWER .html the .md scan skipped.
	write("2026-10-07_13-10_md-report.md")

	if s := FindHtmlBlindSpot(root, DefaultSettings()); s.Count != 0 {
		t.Fatalf("no html yet, got %+v", s)
	}

	write("2026-10-07_10-00_older.html")

	if s := FindHtmlBlindSpot(root, DefaultSettings()); s.Count != 1 || s.Outranks {
		t.Fatalf("older html must count but not outrank, got %+v", s)
	}

	write("2026-10-07_16-25_mail-stack-status.html")

	s := FindHtmlBlindSpot(root, DefaultSettings())
	if s.Count != 2 || !s.Outranks || s.Newest != "2026-10-07_16-25_mail-stack-status.html" {
		t.Fatalf("newer html must outrank the newest .md, got %+v", s)
	}
}

func TestUnmatchedItemStillDrifts(t *testing.T) {
	t.Parallel()

	report := writeReport(t,
		"# Report\n\n## f) Next\n\n1. wire the zaphod-beeblebrox relay\n")
	living := []string{"something entirely different"}
	livingSquashed := []string{Squash(living[0])}

	tokenDF := func(token string) int {
		n := 0

		for _, line := range living {
			if strings.Contains(line, token) {
				n++
			}
		}

		return n
	}

	findings := CheckReport(report, living, tokenDF, livingSquashed, DefaultSettings())
	if len(findings) != 1 {
		t.Fatalf("want 1 finding, got %v", findings)
	}

	if !strings.Contains(findings[0][1], "missing tokens") {
		t.Fatalf("reason missing 'missing tokens': %q", findings[0][1])
	}
}

// --explain-token's table builder: the boundary-token class needs the DF
// number and the distinctive/generic verdict per token, visible at a
// glance (the 2026-10-07 flip: "batch" DF 7→8 silently genericized an
// item's last distinctive token; the diagnosis needed a hand-rolled grep
// probe before this existed).
func TestTokenTableShowsDFAndDistinctiveness(t *testing.T) {
	t.Parallel()

	living := make([]string, 0, 8)
	for range 7 {
		living = append(living, "some line mentioning batch")
	}

	living = append(living, "a line with the zephyr-unique token")

	tokenDF := func(tok string) int {
		n := 0

		for _, line := range living {
			if strings.Contains(line, tok) {
				n++
			}
		}

		return n
	}

	table := TokenTable("Re-run the ritual for the batch", living, squashLines(living), tokenDF, DefaultSettings().df())
	joined := strings.Join(table, "\n")

	for _, want := range []string{
		"batch", "DF=7", "DISTINCTIVE",
		"ritual", "missing",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("token table missing %q:\n%s", want, joined)
		}
	}

	if strings.Contains(joined, "zephyr") {
		t.Errorf("table must only carry the item's own tokens:\n%s", joined)
	}
}

// --pre-flight's delta contract: the exact burn class of 2026-10-07
// (five boundary flips in one session, all from the author's own fresh
// lines) must be caught BEFORE the line lands, not by the next ritual
// run. The scenario pins the mechanics: an item whose only distinctive
// token sits at DF 7, plus a candidate line that pushes it to 8.
func preFlightRoot(t *testing.T) string {
	t.Helper()

	root := t.TempDir()

	status := filepath.Join(root, "docs", "status")
	if err := os.MkdirAll(status, 0o755); err != nil {
		t.Fatal(err)
	}

	report := "# Report\n\n## f) Next\n\n1. webphone alpha-probe calibration recheck\n"
	if err := os.WriteFile(filepath.Join(status, "2026-09-01_00-00_BASE.md"), []byte(report), 0o644); err != nil {
		t.Fatal(err)
	}

	// "webphone" rides 8 lines (generic by definition), "alpha-probe"
	// rides 7 (the item's only distinctive token).
	todo := make([]string, 0, 8)
	for i := range 8 {
		todo = append(todo, fmt.Sprintf("- webphone note number %d", i))
	}

	for i := range 7 {
		todo = append(todo, fmt.Sprintf("- alpha-probe evidence number %d", i))
	}

	if err := os.WriteFile(
		filepath.Join(root, "TODO_LIST.md"),
		[]byte("# TODO\n\n"+strings.Join(todo, "\n")+"\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	return root
}

func TestPreFlightFlagsWouldBeDrift(t *testing.T) {
	t.Parallel()

	root := preFlightRoot(t)

	if rc := PreFlight(root, []string{"changelog: alpha-probe summary row"}, false, false, DefaultSettings()); rc != gatecli.ExitFindings {
		t.Fatalf("a line genericizing the last distinctive token must exit findings, got %d", rc)
	}

	unrelated := []string{"changelog: a wholly unrelated qwertyzz row"}
	if rc := PreFlight(root, unrelated, false, false, DefaultSettings()); rc != gatecli.ExitClean {
		t.Fatalf("an unrelated line must exit clean, got %d", rc)
	}
}

// The culprit attribution (TODO preflight-culprit, 2026-10-16): a
// WOULD-DRIFT must name the candidate LINE carrying the generic token —
// the 2026-10-13 filename chase burned two guess rounds because the
// boundary-token report names tokens, never lines, and a hyphenated
// spelling matches without moving the raw DF count at all.
//
// Deliberately NOT parallel: the output-capturing tests swap the global
// os.Stdout, which would race every other parallel test's stdout writes.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestPreFlightNamesCulpritCandidateToken(t *testing.T) {
	root := preFlightRoot(t)

	captured := gatecli.CaptureOutput(func() int {
		return PreFlight(root, []string{"changelog: alpha-probe summary row"}, false, false, DefaultSettings())
	})

	if captured.Code != gatecli.ExitFindings {
		t.Fatalf("the genericizing line must exit findings, got %d", captured.Code)
	}

	got := captured.Stdout

	if !strings.Contains(got, `culprit: candidate #1 token "alpha-probe"`) {
		t.Fatalf("the WOULD-DRIFT must attribute the carrying candidate line, got output:\n%s", got)
	}
}

// The silent class the attribution exists for: a hyphen/space spelling of
// the token matches via the squash key WITHOUT raising the raw DF count —
// the boundary report never fires for it, so the WOULD-DRIFT must mark it
// explicitly.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestPreFlightMarksSquashSpellingCulprit(t *testing.T) {
	root := preFlightRoot(t)

	captured := gatecli.CaptureOutput(func() int {
		return PreFlight(
			root,
			[]string{"changelog: the alpha probe summary row", "changelog: alpha-probe landing row"},
			false,
			false,
			DefaultSettings(),
		)
	})

	if captured.Code != gatecli.ExitFindings {
		t.Fatalf("the batch must exit findings, got %d", captured.Code)
	}

	got := captured.Stdout

	if !strings.Contains(got, `"alpha-probe" (hyphen/space spelling`) {
		t.Fatalf("the squash-only spelling must be marked as the silent culprit, got output:\n%s", got)
	}
}

func TestBoundaryCrossingsReportsOnlyThresholdFlips(t *testing.T) {
	t.Parallel()

	base := make([]string, 0, 15)
	for i := range 8 {
		base = append(base, fmt.Sprintf("webphone line %d", i))
	}

	for i := range 7 {
		base = append(base, fmt.Sprintf("alpha-probe line %d", i))
	}

	mod := append(slices.Clone(base), "alpha-probe mention")

	crossings := BoundaryCrossings([]string{"alpha-probe mention"}, base, mod, DefaultSettings().df())
	if len(crossings) != 1 || crossings[0].Token != "alpha-probe" || crossings[0].Before != 7 ||
		crossings[0].After != 8 {
		t.Fatalf("want exactly alpha-probe 7→8, got %+v", crossings)
	}

	// Already-generic tokens do not cross (8→9 stays above the bar) and
	// fresh tokens do not either (0→1 stays below it).
	crossings = BoundaryCrossings([]string{"webphone again and brandnew-zz"}, base, base, DefaultSettings().df())
	if len(crossings) != 0 {
		t.Fatalf("no crossing expected for already-generic or fresh tokens, got %+v", crossings)
	}
}

func TestPreFlightEmptyCandidateIsUsage(t *testing.T) {
	t.Parallel()

	root := preFlightRoot(t)

	if rc := PreFlight(root, nil, false, false, DefaultSettings()); rc != gatecli.ExitUsage {
		t.Fatalf("empty candidate must exit usage, got %d", rc)
	}
}

// bindFixture builds a root whose ONE report item fails in base (its
// tokens appear nowhere in the living docs) so the pre-flight bind leg
// has a real unharvested item to harvest or miss.
func bindFixture(t *testing.T, item string) string {
	t.Helper()

	root := t.TempDir()

	status := filepath.Join(root, "docs", "status")
	if err := os.MkdirAll(status, 0o755); err != nil {
		t.Fatal(err)
	}

	report := "# Report\n\n## f) Next\n\n1. " + item + "\n"
	if err := os.WriteFile(filepath.Join(status, "2026-09-02_00-00_BIND.md"), []byte(report), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(root, "TODO_LIST.md"),
		[]byte("# TODO\n\n- unrelated filler row for the living docs\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	return root
}

// The bind leg (TODO §77 preflight-bind-check): a candidate row that
// harvests the open item must SAY so, attributed to the exact line.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestPreFlightReportsBoundItemWithCandidate(t *testing.T) {
	root := bindFixture(t, "quota-xyz replenish window followup")

	captured := gatecli.CaptureOutput(func() int {
		return PreFlight(root, []string{"todo: quota-xyz replenish window owner row"}, false, false, DefaultSettings())
	})

	if captured.Code != gatecli.ExitClean {
		t.Fatalf("the harvesting line must exit clean, got %d (output:\n%s)", captured.Code, captured.Stdout)
	}

	got := captured.Stdout

	if !strings.Contains(got, "binds: [2026-09-02_00-00_BIND.md]") ||
		!strings.Contains(got, "quota-xyz replenish window followup") {
		t.Fatalf("the newly-harvested item must be listed with its report, got:\n%s", got)
	}

	if !strings.Contains(got, "via candidate: todo: quota-xyz replenish window owner row") {
		t.Fatalf("the binding must attribute the candidate line, got:\n%s", got)
	}
}

// The 2026-10-18 miss class made visible BEFORE the write: the item's
// distinctive token is the slash-joined "metric/counter"; a row that
// spells it "metric and counter" binds NOTHING, and the bind leg must
// say so instead of letting a SAFE verdict imply a harvest.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestPreFlightBindsNothingWhenSlashTokenIsSplit(t *testing.T) {
	// The item's only harvestable tokens are the slash-joined
	// "metric/counter" plus a pair of fresh words the candidate must NOT
	// carry — otherwise THOSE bind and the slash miss hides again (the
	// first draft of this test failed exactly that way).
	root := bindFixture(t, "the metric/counter qvadrant alarm needs a drain row")

	captured := gatecli.CaptureOutput(func() int {
		return PreFlight(
			root,
			[]string{"todo: one unrelated row spelling metric and counter separately"},
			false,
			false,
			DefaultSettings(),
		)
	})

	if captured.Code != gatecli.ExitClean {
		t.Fatalf("the non-binding line must still exit clean, got %d (output:\n%s)",
			captured.Code, captured.Stdout)
	}

	got := captured.Stdout

	if strings.Contains(got, "via candidate:") {
		t.Fatalf("a split spelling of a slash-joined token must not bind, got:\n%s", got)
	}

	if !strings.Contains(got, "binds: no open report item is harvested") {
		t.Fatalf("the empty-bind summary must be present, got:\n%s", got)
	}
}

func TestPreFlightRefusesCandidatesAlreadyOnDisk(t *testing.T) {
	t.Parallel()

	root := preFlightRoot(t)

	// A candidate that is already a TODO_LIST row: a post-write
	// pre-flight carries its own additions inside the baseline, so the
	// delta cannot see a flip — refuse instead of printing a
	// misleading SAFE (the 2026-10-10 burn).
	if rc := PreFlight(root, []string{"- alpha-probe evidence number 3"}, false, false, DefaultSettings()); rc != gatecli.ExitUsage {
		t.Fatalf("a candidate already on disk must exit usage, got %d", rc)
	}

	// Whitespace-only candidates are vacuous, not "on disk".
	if rc := PreFlight(root, []string{"   "}, false, false, DefaultSettings()); rc != gatecli.ExitClean {
		t.Fatalf("a whitespace-only candidate must stay clean, got %d", rc)
	}
}

// The 2026-10-12 ledger-block class: a candidate line that is ALREADY on
// disk is normally refused (the post-write-run guard), which forced
// operator-side dedup (grep -Fxvf) — silently DROPPING the +1 DF push of
// the second copy that was about to land. --duplicates is the explicit
// "another copy is intentional" assertion: the copy stays modeled and its
// push must surface as findings when it genericizes an item.
func TestPreFlightDuplicatesModelsSecondCopyPush(t *testing.T) {
	t.Parallel()

	root := preFlightRoot(t)

	// Without the flag the second copy is still refused, not answered.
	if rc := PreFlight(root, []string{"- alpha-probe evidence number 3"}, false, false, DefaultSettings()); rc != gatecli.ExitUsage {
		t.Fatalf("second copy without --duplicates must stay refused, got %d", rc)
	}

	// With it, the +1 push genericizes the item's only distinctive token.
	if rc := PreFlight(root, []string{"- alpha-probe evidence number 3"}, false, true, DefaultSettings()); rc != gatecli.ExitFindings {
		t.Fatalf("second copy with --duplicates must model the DF push and exit findings, got %d", rc)
	}

	// A second copy of an already-generic line pushes nothing that matters.
	if rc := PreFlight(root, []string{"- webphone note number 3"}, false, true, DefaultSettings()); rc != gatecli.ExitClean {
		t.Fatalf("second copy of a generic line must stay clean, got %d", rc)
	}
}

// The on-disk refusal must NAME the offending line (the message contract
// the 2026-10-11 report asked for): an operator who cannot see WHICH
// candidate is already present cannot tell a stale re-run from a
// deliberate ledger second copy. Not parallel: CaptureOutput swaps the
// process-global streams.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel()
func TestPreFlightRefusalNamesOffendingLine(t *testing.T) {
	root := preFlightRoot(t)

	captured := gatecli.CaptureOutput(func() int {
		return PreFlight(root, []string{"- alpha-probe evidence number 3"}, false, false, DefaultSettings())
	})

	if captured.Code != gatecli.ExitUsage {
		t.Fatalf("on-disk candidate must exit usage, got %d", captured.Code)
	}

	if !strings.Contains(captured.Stderr, "alpha-probe evidence number 3") {
		t.Fatalf("refusal must name the offending line in its Present list, stderr:\n%s", captured.Stderr)
	}
}

// --duplicates outside pre-flight mode is a usage error: the assertion it
// encodes ("this on-disk line lands as another copy") only has meaning
// for the additions-only delta.
func TestDuplicatesFlagNeedsPreFlightMode(t *testing.T) {
	t.Parallel()

	if rc := Run([]string{".", "--duplicates"}); rc != gatecli.ExitUsage {
		t.Fatalf("--duplicates without --pre-flight must exit usage, got %d", rc)
	}
}

func TestStatusReportsIncludesArchived(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	archived := filepath.Join(root, "docs", "status", "archived")
	if err := os.MkdirAll(archived, 0o755); err != nil {
		t.Fatal(err)
	}

	live := filepath.Join(root, "docs", "status", "2026-09-01_00-00_LIVE.md")
	old := filepath.Join(archived, "2026-08-01_00-00_OLD.md")

	for _, p := range []string{live, old} {
		if err := os.WriteFile(p, []byte("# Report\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	reports := statusReports(root, DefaultSettings())
	if len(reports) != 2 {
		t.Fatalf("want the live and the archived report, got %d: %v", len(reports), reports)
	}

	if !strings.HasSuffix(reports[0], filepath.Join("docs", "status", "2026-09-01_00-00_LIVE.md")) {
		t.Fatalf("live report must come first, got %s", reports[0])
	}

	if !strings.Contains(reports[1], "archived") {
		t.Fatalf("the archived report must be part of the replay corpus, got %s", reports[1])
	}
}

// The 2026-10-10 report feared pre-flight skips docs/status/archived/
// while --audit-all replays it (m144). The corpus lister has included
// archived reports since 2026-09-30 (63af9cf), so this pins the
// behavior end-to-end: a candidate line genericizing the last
// distinctive token of an ARCHIVED report's §f item must flip the
// verdict to findings — the archived corpus is inside the delta.
func TestPreFlightReplaysArchivedCorpus(t *testing.T) {
	t.Parallel()

	root := preFlightRoot(t)

	status := filepath.Join(root, "docs", "status", "archived")
	if err := os.MkdirAll(status, 0o755); err != nil {
		t.Fatal(err)
	}

	report := "# Report\n\n## f) Next\n\n1. webphone alpha-probe calibration recheck\n"
	if err := os.WriteFile(
		filepath.Join(status, "2026-08-01_00-00_ARCHIVED.md"),
		[]byte(report),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	if rc := PreFlight(root, []string{"changelog: alpha-probe summary row"}, false, false, DefaultSettings()); rc != gatecli.ExitFindings {
		t.Fatalf("an archived item's last distinctive token must still flip the verdict, got %d", rc)
	}

	unrelated := []string{"changelog: a wholly unrelated qwertyzz row"}
	if rc := PreFlight(root, unrelated, false, false, DefaultSettings()); rc != gatecli.ExitClean {
		t.Fatalf("an unrelated line must exit clean, got %d", rc)
	}
}

// --explain-item prints the token table for ANY matching open item —
// harvested ones included — because --explain-token tables ride
// UNHARVESTED findings only and a FRESH corpus therefore prints none
// (the §43 row: AGENTS claimed a table that never came).
func TestExplainItemTablesHarvestedItems(t *testing.T) {
	t.Parallel()

	root := preFlightRoot(t)
	livingLines, livingSquashed, tokenDF := loadLivingDocs(root, DefaultSettings())

	rc := ExplainItem(root, "alpha-probe", livingLines, livingSquashed, tokenDF, DefaultSettings())
	if rc != gatecli.ExitClean {
		t.Fatalf("diagnostic mode always exits clean, got %d", rc)
	}
}

func TestExplainItemRunEndToEnd(t *testing.T) {
	t.Parallel()

	root := preFlightRoot(t)

	if rc := Run([]string{root, "--explain-item", "alpha-probe"}); rc != gatecli.ExitClean {
		t.Fatalf("matching needle must exit clean, got %d", rc)
	}

	if rc := Run([]string{root, "--explain-item", "no-such-needle"}); rc != gatecli.ExitClean {
		t.Fatalf("unmatched needle must also exit clean (report-only), got %d", rc)
	}

	if rc := Run([]string{root, "--explain-item"}); rc != gatecli.ExitUsage {
		t.Fatalf("missing needle must exit usage, got %d", rc)
	}
}

// The audit ratchet (TODO §91 coin audit-ratchet): --audit-max=N turns the
// audit red only on a REGRESSION past the pin — a legacy backlog nobody is
// draining in one sitting stays green but named, zero stays green, and a
// count past the pin (or a non-integer pin) still fails. The gate was
// permanently red 2026-10-08..23 while the backlog sat at 78 because any
// nonzero count failed.
func TestAuditMaxRatchet(t *testing.T) {
	t.Parallel()

	// One unharvested item: every token missing from the living docs.
	root := preFlightRoot(t)

	report := "# Report\n\n## f) Next\n\n1. wire the zaphod-beeblebrox relay\n"
	if err := os.WriteFile(
		filepath.Join(root, "docs", "status", "2026-09-02_00-00_GAP.md"),
		[]byte(report), 0o644,
	); err != nil {
		t.Fatal(err)
	}

	if rc := Run([]string{root, "--audit-all"}); rc != gatecli.ExitFindings {
		t.Fatalf("strict audit (no pin) must fail on 1 unharvested, got %d", rc)
	}

	if rc := Run([]string{root, "--audit-all", "--audit-max=1"}); rc != gatecli.ExitClean {
		t.Fatalf("1 unharvested under a --audit-max=1 pin must stay green, got %d", rc)
	}

	if rc := Run([]string{root, "--audit-all", "--audit-max=0"}); rc != gatecli.ExitFindings {
		t.Fatalf("1 unharvested past a --audit-max=0 pin is a REGRESSION, got %d", rc)
	}

	if rc := Run([]string{root, "--audit-all", "--audit-max=bogus"}); rc != gatecli.ExitUsage {
		t.Fatalf("a non-integer pin must exit usage, got %d", rc)
	}

	// Legacy-zero stays green under any pin.
	fresh := preFlightRoot(t)
	if rc := Run([]string{fresh, "--audit-all", "--audit-max=0"}); rc != gatecli.ExitClean {
		t.Fatalf("a fresh corpus stays green under --audit-max=0, got %d", rc)
	}
}

// Struck (~~) and skip-marked items carry no harvest margin and must not
// match in --explain-item — same open-item filter as CheckReport.
// Deliberately NOT parallel: the only output-capturing test in the suite,
// it swaps the global os.Stdout, which would race every other parallel
// test's stdout writes under test-race.
//
//nolint:paralleltest // stdout swap is unsafe under t.Parallel(); all other tests only assert exit codes
func TestExplainItemSkipsStruckAndExemptItems(t *testing.T) {
	root := preFlightRoot(t)
	struck := filepath.Join(root, "docs", "status", "2026-09-02_00-00_STRUCK.md")

	body := "# Report\n\n## f) Next\n\n1. ~~alpha-probe thing~~ done\n2. alpha-probe exempt <!-- fresh:skip -->\n"
	if err := os.WriteFile(struck, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	livingLines, livingSquashed, tokenDF := loadLivingDocs(root, DefaultSettings())

	captured := gatecli.CaptureOutput(func() int {
		return ExplainItem(root, "exempt", livingLines, livingSquashed, tokenDF, DefaultSettings())
	})

	if captured.Code != gatecli.ExitClean {
		t.Fatalf("diagnostic mode always exits clean, got %d", captured.Code)
	}

	if got := captured.Stdout; strings.Contains(got, "exempt") && strings.Contains(got, "verdict:") {
		t.Fatalf("struck/skip-marked items must not be tabled, got output:\n%s", got)
	}
}

// TestSkipMarkerAcceptsSuffixedForm pins the 2026-10-25 parser gap: reports
// carry the documented suffixed markers (e.g. "<!-- fresh:skip: already
// routed in TODO_LIST -->") while the byte-exact substring match silently
// ignored them, resurfacing routed items as unharvested on --audit-all.
func TestSkipMarkerAcceptsSuffixedForm(t *testing.T) {
	t.Parallel()

	root := preFlightRoot(t)

	report := "# Report\n\n## f) Next\n\n1. beta-probe routed elsewhere <!-- fresh:skip: already routed in TODO_LIST -->\n2. beta-probe next-session <!-- fresh:skip: next-session new work -->\n"
	if err := os.WriteFile(
		filepath.Join(root, "docs", "status", "2026-09-02_00-00_SUFFIXED.md"),
		[]byte(report), 0o644,
	); err != nil {
		t.Fatal(err)
	}

	if rc := Run([]string{root, "--audit-all", "--audit-max=0"}); rc != gatecli.ExitClean {
		t.Fatalf("suffixed fresh:skip markers must exempt their items (audit-max=0), got %d", rc)
	}
}
