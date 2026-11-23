# cmdguard utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v5.1.0 (latest local tag of root module `github.com/larsartmann/cmdguard/v5`, fetched 2026-11-23) + HEAD f36c61b1ad01b738db7ccd4f10508634c9b6d9c9. Submodules in the tree: flightrecorder v0.2.0, glamour v0.3.0, lint v0.4.0, prompts v0.3.0, schema v0.1.0, spinner v0.3.0, telemetry v0.3.0.
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 4 (ops/health/auth/CLI-framework) subagent of the multi-library utilization audit fleet.
- Fleet stance addressed: "if all my projects re-implement everything, I will only fix bugs all day and none benefit from collective improvements" — cmdguard is the sharpest test of this argument in the whole run, because the consumer's verb dispatch + flag parser + config prologue is exactly what cmdguard productizes.

## Verdict

NOT-ADOPTED — a genuine, nameable overlap exists (the hand-rolled verb table, boolean flag parser, config prologue, did-you-mean engine, and exit-code ladder are cmdguard's core product, down to one near line-for-line duplicated Levenshtein implementation), but adopting it would break the root module's zero-require contract, force re-pinning every byte-pinned help text and parity golden that defines the argv-compatibility API (`Run([]string) int` matching the historical standalone binaries, legacy wrappers answering `--version` under their own names), and invert a relationship that already works in the right direction: cmdguard is docs-gates' showcase *consumer*, not the other way around.

## Score

- Inputs N = 16 (inventory rows below).
- A = 7 (rows with a live hand-rolled counterpart in the consumer: dispatch, flag parsing, help surface, did-you-mean, version, config prologue, exit codes).
- M = 0.
- Score = round(100 × 0/7) = **0** — band **low**. Honest in both directions: the zero is deliberate insulation under a contract set the library cannot serve, not absence of overlap — the duplication is named below.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | Typed struct-tag flags (13 tags: flag/short/default/help/env/required/validate/values/count/local/hidden/prompt/secret) | Concept yes — docs-gates hand-rolls boolean-only parsing | Not used | cmdguard/README.md:455-483; internal/gatecli/gatecli.go:124-156 |
| 2 | Per-command flag structs, validated at construction (missing handler, duplicate, invalid flag die at `AddCommand`) | Concept yes — the verb tables are static map literals that cannot mis-wire | Not used | cmdguard/README.md:101-121; cmd/docs-gates/verbs.go:77-92 |
| 3 | Command tree/dispatch (cobra underneath, parents + subcommands) | Yes — the 14-verb dispatch table is the same concern | Hand-rolled | cmdguard/README.md:344-359; cmd/docs-gates/verbs.go:77-92 |
| 4 | "Did you mean?" suggestions for commands and flags (Levenshtein) | Yes — **duplicated almost line-for-line** (see gap analysis) | Hand-rolled | cmdguard/pkg/cmdguard/v5/flags_suggest.go:13-42; cmd/docs-gates/main.go:188-247 |
| 5 | Version command + `WithCLIVersion` | Concept yes — docs-gates intercepts `--version` per legacy name | Hand-rolled | cmdguard/README.md:706-719; internal/gatecli/gatecli.go:109-117 |
| 6 | Styled help surface (fang + lipgloss, markdown via glamour submodule) | Concept yes — docs-gates' help is byte-pinned plain text | Hand-rolled | cmdguard/README.md:152-158; cmd/docs-gates/main.go:165-182 |
| 7 | Config files with precedence flag → env → file → default, strict keys ON, unknown-key suggestions | Concept yes — docs-gates' JSON config prologue warns on unknown keys (no suggestions) | Hand-rolled | cmdguard/README.md:613-627; cmdguard/pkg/cmdguard/v5/config_keys.go:456-459; internal/gatecli/gatecli.go:220-234 |
| 8 | Exit-code contract (`ExitCode(err)`, `ExitCoder`, error printed exactly once) | Concept yes — docs-gates' 0/1/2 constants with different semantics (exit 1 is a verdict, not an error) | Hand-rolled | cmdguard/pkg/cmdguard/v5/cli_errors_json.go:73; cmdguard/pkg/cmdguard/v5/errors.go:205; internal/gatecli/gatecli.go:26-30 |
| 9 | DI container (samber/do v2, Provide/Invoke, lazy services, lifecycle) | No — gates are stateless packages; nothing to inject | N/A | cmdguard/README.md:283-305 |
| 10 | Signal handling + graceful reverse-order shutdown (default ON in v5) | No — run-and-exit; no long-lived services | N/A | cmdguard/README.md:559-569 |
| 11 | Middleware (Timing, Recovery, custom; spinner/telemetry/flightrecorder as submodules) | No — handlers are one `Run` call, no chain | N/A | cmdguard/README.md:269; cmdguard/README.md:631-643 |
| 12 | 16 output formats via go-output (table/JSON/CSV/YAML/…) | No — output is byte-pinned prose + exactly one JSON face | N/A | cmdguard/README.md:327-340; internal/battery/battery.go:264-313 |
| 13 | Machine-readable `describe` contract (versioned JSON of the command/flag tree) | Concept adjacent — `run --json` is the machine face, pinned + strict-decoded by the provider | N/A | cmdguard/README.md:133-149; pkg/provider/provider.go:92-102 |
| 14 | `doctor` command over `do.HealthcheckerWithContext` services | No — health is the exit code (go-health report, this group) | N/A | cmdguard/README.md:305 |
| 15 | prompts / spinner / glamour / telemetry / flightrecorder / lint / schema submodules | No — no interactivity, no TTY chrome, no OTel, not a cmdguard CLI to lint | N/A | cmdguard/README.md:631-643 |
| 16 | Cobra escape hatch (`ConfigFromContext`, `WithPostFlagParse`, raw subcommands) | No — no cobra anywhere to bridge from | N/A | cmdguard/README.md:397-423 |

Dependency note: cmdguard is the opposite pole of the consumer's
architecture — cmdguard/go.mod:72-87 directly requires cobra, pflag, koanf
(JSON/YAML/TOML parsers), fang, lipgloss, go-output, samber/do,
samber-do-auditlog, go-branded-id, and three of its own submodules, with
~60 indirect requires on top (cmdguard/go.mod:5-69). Any import anywhere
in the root module fails scripts/stack-drift-check.sh:8-9 on sight.

## What docs-gates does today

Zero module usage. Verified:

```console
$ rg -n "cmdguard" -g '*.go' -g 'go.mod' -g 'go.sum' /home/lars/projects/docs-gates
# rc=1, no matches in either module (root go.mod has no requires at all)
```

The word "cmdguard" appears only as prose (planning docs, status reports,
README's adoption section) and one flake.lock transitive node: the erraudit
master-tracking input depends on cmdguard (flake.lock:3-22) — docs-gates
never imports it.

The hand-rolled scaffolding that occupies cmdguard's product space, with
the census:

- **Verb dispatch** — cmd/docs-gates/verbs.go:77-92 (`verbHandlers` map,
  14 verbs → `Run([]string) int`), :56-71 (`displayName` legacy-name map),
  :97-101 (display order), :107-122 (one-liners), :129-144 (usage wiring);
  cmd/docs-gates/main.go:34-81 is the whole dispatcher: `--version`
  interception at :37-41 and :67-69, help interception at :45-47 and
  :74-78, unknown-verb refusal at :57-63.
- **Flag parsing** — internal/gatecli/gatecli.go:124-156 (`ParseFlags`:
  boolean `--flags` anywhere + sorted unknown-flag rejection),
  :158-175 (`Parse`: flags + repo-root resolution), :188-218 (`CutIntFlag`
  for the three valued `--flag=N` ints), :236-255 (`ResolveRoot`). The
  entire flag vocabulary is small: freshness carries 9 booleans + 2 valued
  flags (internal/docsfreshness/docsfreshness.go:86-102), annotate 2
  booleans (internal/docsannotate/docsannotate.go:131), routing 1
  (internal/docsrouting/docsrouting.go:231), battery `--json`
  (internal/battery/battery.go:44), archive `--apply`
  (internal/docsarchive/docsarchive.go:73); six verbs take no flags.
- **Config prologue** — internal/gatecli/gatecli.go:220-234 (`LoadConfig`)
  over internal/config/known_keys.go:3-6 (unknown-key WARN machinery);
  malformed JSON exits 2.
- **Exit codes** — internal/gatecli/gatecli.go:26-30 (`ExitClean`/
  `ExitFindings`/`ExitUsage`), :96-101 (`FatalUsage`'s one refusal shape).
- **The inverted relationship (the most important fact)** — cmdguard is
  docs-gates' worked adoption example: README.md:76-77 ("its adoption is
  exactly this section, verbatim"), and cmdguard's CI dogfood job runs the
  battery over cmdguard itself
  (cmdguard/.github/workflows/ci.yml:191, pinned `@v0.3.1`). docs-gates
  judges cmdguard's living docs; collective improvement already flows
  between them — in the direction the fleet argument wants, without a
  single import.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed — honestly in both directions.**

*What adoption would genuinely retire:* cmd/docs-gates/main.go (248 lines)
+ verbs.go (145 lines) + gatecli's parse/config prologue (~130 lines) are
hand-rolled CLI plumbing, and cmdguard's own "Replace your hand-rolled"
table (cmdguard/README.md:63-77) reads like a checklist of this repo:
exit-code ladder, version command, did-you-mean, config loader. One item
is flat duplication: cmd/docs-gates/main.go:188-247 (`suggestVerb` +
`editDistance`, prefix-match then Levenshtein ≤2) is the same algorithm as
cmdguard/pkg/cmdguard/v5/flags_suggest.go:13-42 (`SuggestFlag`/
`SuggestCommand` + `editDistance`, distance ≤3) — both hand-rolled because
the root module cannot require cobra's. If the fleet argument is read as
"one Levenshtein per fleet is enough," docs-gates is on the wrong side of
it, and so is cmdguard (which rewrote cobra's instead of exporting it).
Construction-time validation (row 2) is the one capability with no
consumer-side equivalent at all — but the verb tables are package-level
map literals over compiled-in functions; there is no runtime registration
to get wrong, and main_test.go plus per-gate usage tests pin the wiring.

*What adoption would destroy — the bar the brief sets:* root adoption must
justify breaking zero-requires, and it cannot:

1. **The argv contract is the API.** Every gate is a `Run([]string) int`
   whose argv matches the historical standalone binaries (AGENTS.md layout
   table); the battery itself invokes gates through those argv shapes
   (internal/battery/battery.go:58-81). pflag semantics differ from
   `ParseFlags` in observable ways (bundled shorts, `--flag value` spaced
   form, `--` terminator, unknown-flag message shape), and
   internal/battery/parity_test.go:84's table byte-pins gate output over
   fifteen-plus invocations — every one a re-pin, and AGENTS.md calls a
   golden bump a contract change.
2. **Exit 1 is not an error here.** cmdguard's core contract is "the error
   is printed exactly once, exit code mapped" (cmdguard/README.md:152-158).
   docs-gates' exit 1 is a *verdict*: findings print in each gate's own
   byte-pinned format on its own stream, and the battery's job is to
   aggregate them (internal/battery/battery.go:317-340). Mapping that onto
   `NewExitError` + error-print suppression means fighting the framework's
   central metaphor to preserve a stricter one.
3. **The legacy wrapper surface.** The five historical names are API:
   flake wrappers exec `docs-gates <verb>` and each verb answers
   `--version` under its legacy name (cmd/docs-gates/verbs.go:56-71,
   internal/gatecli/gatecli.go:109-117). A cobra-based binary reports
   `argv[0]`/root-command identity; per-verb display names would need
   custom version plumbing anyway.
4. **Help bytes.** Per-gate `Usage` constants are pinned by tests
   (internal/docsfreshness/usage_test.go:15,
   internal/battery/usage_test.go:15); fang/lipgloss-styled help changes
   every byte of every one.
5. **The dependency weight is the point of the exclusion.** Hermetic nix
   build, `go install`-ability, the dogfood doctrine ("drop a binary, run
   the battery") all lean on zero requires
   (scripts/stack-drift-check.sh:8-9). cmdguard is a framework for apps
   whose UX chrome (spinners, prompts, styled errors) is the product;
   docs-gates' product is verdicts.
6. **pkg/provider is not a home either** — its module budget is
   go-finding/toolsdk only, and it has no CLI to framework.

Net: the ~500 lines of dispatch/parse/config scaffolding are not free, but
they are the *price* of the contract set, they are pinned by tests at
three layers (unit, usage pins, parity goldens), and they change rarely —
the TODO/CHANGELOG show no flag-parser bug class. The fleet argument's
correct application here is the one already in force: docs-gates serves
cmdguard (and every other repo) as the shared tool, and inherits fleet
improvements through releases, not through imports.

RECONSIDER-WHEN: (a) this repo grows a second, interactive binary (a
battery TUI, a fleet dashboard host, a daemon) — build it on cmdguard in
its own go module from day one, the pkg/provider pattern, inheriting none
of the root contract; (b) cmdguard ever ships a zero-dependency core
distribution (pure-stdlib typed flags + dispatch, no cobra/fang/lipgloss)
— at that point root adoption becomes reviewable against the golden-re-pin
cost alone; (c) a gate's flag surface grows rich enough (env defaults,
required, enums, validation rules) that `ParseFlags` + `CutIntFlag` start
accumulating per-flag special cases — the moment parsing logic becomes a
maintenance line item rather than a settled one.

## Ranked opportunities

1. **Port cmdguard's unknown-config-key suggestion into the config WARN
   path.** cmdguard turns unknown config keys into
   `did you mean %q?` (cmdguard/pkg/cmdguard/v5/config_keys.go:456-459);
   docs-gates only WARNs on the unknown key
   (internal/config/known_keys.go:3-6). A `--audit-max` → `audit_max`-class
   typo in `.docs-gates.json` today costs a re-read of the README; with the
   suggestion it costs nothing. Zero dependencies: the Levenshtein already
   exists in cmd/docs-gates/main.go:223-247 and would lift into gatecli or
   config as a shared helper. ~20 lines, one test, visible payoff on every
   mis-keyed repo config.
2. (Nothing else adoption-shaped. The `editDistance` duplication itself is
   not worth consolidating cross-repo — a 25-line stdlib function with
   pinned behavior on both sides is below any reasonable sharing threshold;
   the fleet-level fix is cmdguard exporting its suggester someday, not
   docs-gates importing the framework.)

## How to verify

```console
$ git -C /home/lars/projects/cmdguard log -1 --format='%H'
f36c61b1ad01b738db7ccd4f10508634c9b6d9c9
$ git -C /home/lars/projects/cmdguard tag --sort=-v:refname | head -1
v5.1.0
$ sed -n '72,87p' /home/lars/projects/cmdguard/go.mod
require (
	charm.land/fang/v2 v2.0.1
	charm.land/lipgloss/v2 v2.0.6
	github.com/knadh/koanf/parsers/json v1.0.1
	github.com/knadh/koanf/parsers/toml v0.1.0
	github.com/knadh/koanf/parsers/yaml v1.1.1
	github.com/knadh/koanf/v2 v2.4.1
	github.com/larsartmann/cmdguard/flightrecorder v0.2.0
	github.com/larsartmann/cmdguard/glamour v0.3.0
	github.com/larsartmann/cmdguard/spinner v0.3.0
	github.com/larsartmann/go-branded-id v0.7.1
	github.com/larsartmann/go-output v0.38.7
	github.com/larsartmann/samber-do-auditlog v0.12.3
	github.com/samber/do/v2 v2.1.0
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
# (the block's closing paren follows on the next line; above it, the
# cmdguard/go.mod indirect block carries ~60 more requires)

$ rg -n "cmdguard" -g '*.go' -g 'go.mod' -g 'go.sum' /home/lars/projects/docs-gates
# rc=1 — zero code/module-graph references (flake.lock:3-22 carries it only
# as erraudit's transitive input; prose mentions are adoption-history notes)

# the duplicated did-you-mean engines, side by side:
$ rg -n "func suggestVerb|func editDistance" /home/lars/projects/docs-gates/cmd/docs-gates/main.go
188:func suggestVerb(typed string) (string, bool) {
223:func editDistance(a, b string) int {
$ rg -n "func SuggestFlag|func SuggestCommand|func editDistance" /home/lars/projects/cmdguard/pkg/cmdguard/v5/flags_suggest.go
13:func SuggestFlag(validNames []string, input string) (string, bool) {
36:func SuggestCommand(validCommands []string, input string) (string, bool) {
42:func editDistance(a, b string) int {

# the argv-compatibility surface adoption would have to preserve:
$ rg -n "func gates\(" /home/lars/projects/docs-gates/internal/battery/battery.go
58:func gates(root string, cfg config.Config) []gate {
$ rg -n "return \[\]parityGate{" /home/lars/projects/docs-gates/internal/battery/parity_test.go
84:	return []parityGate{
$ rg -n "var displayName = map\[string\]string{" /home/lars/projects/docs-gates/cmd/docs-gates/verbs.go
56:var displayName = map[string]string{

# the inverted relationship (cmdguard consumes docs-gates):
$ rg -n "docs-gates/cmd/docs-gates" /home/lars/projects/cmdguard/.github/workflows/ci.yml
191:        run: go run github.com/Artmann-Technologies/docs-gates/cmd/docs-gates@v0.3.1 run .
```
