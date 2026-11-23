# go-credentials-store utilization audit — docs-gates

- Audited: 2026-11-23.
- Library version reviewed: v0.4.3 (latest local tag, fetched 2026-11-23) + HEAD 6d7caad7abd013019c232eb1b489d144c5048469 — note: per the orchestrator's checkout manifest this HEAD sits 3 commits AHEAD of upstream (local, unpushed work); the repo's auto-commit daemon has since advanced the checkout past that sha (verified `git log -1` → f9c11ce, format/report chores only), so the audited numbers pin the orchestrator baseline. Single module `github.com/larsartmann/go-credentials-store`.
- Consumer: docs-gates @ adfa9c80a74bdba3ea1bd1cc37fdb1b016109f9f — root module stdlib-only (zero non-stdlib requires) + `pkg/provider` as the one third-party-dep module.
- Provenance: group 4 (ops/health/auth/CLI-framework) subagent of the multi-library utilization audit fleet.
- Fleet stance addressed: "if all my projects re-implement everything, I will only fix bugs all day and none benefit from collective improvements" — does stdlib-only insulation here avoid real re-implementation debt, or leave docs-gates hand-rolling something the fleet solved?

## Verdict

NOT-APPLICABLE — go-credentials-store is an event-sourced, encrypted-at-rest secrets vault with a SQLite event store and an admin HTTP UI, while docs-gates holds no secrets, tokens, or credentials of any kind (its only privileged act is reading the repo tree it judges, and git-over-ssh happens outside the binary in the user's environment), so there is no credential concern in the consumer for the library to own.

## Score

- Inputs N = 10 (inventory rows below).
- A = 0 (no capability has a call site: the consumer's threat surface contains no secret material to store, resolve, rotate, or audit).
- M = 0.
- Score = round(100 × 0/0) → defined as **0** by convention when A = 0 — band **low**. Expected for NOT-APPLICABLE.

- Applicable rows are exactly the inventory rows graded **Missed Opportunity** (the consumer has the underlying need; the capability is hand-rolled, solved in-domain, or declined by contract). Every **Not Applicable** row carries a one-line reason. `M` counts rows graded Fully Leveraged.

## What the library offers

| # | Capability | Applies to docs-gates? | Status | Evidence |
|---|---|---|---|---|
| 1 | Encrypted-at-rest credential store (XChaCha20-Poly1305, KeyID rotation, sealed-before-event) | No — docs-gates never persists a secret | N/A | go-credentials-store/README.md:3-12 |
| 2 | Library SDK: `credentials.Build(ctx, opts)` + `Resolve(ctx, platform, kind)` for API keys / passwords / session blobs | No — nothing to resolve; the binary's inputs are argv, JSON config, and repo files | N/A | go-credentials-store/README.md:45-64 |
| 3 | Lifecycle audit trail (created/rotated/probed/used/revealed events) | No — no secret lifecycle to audit | N/A | go-credentials-store/README.md:4-6 |
| 4 | Loopback admin UI (`adminui.Mount`) with audited one-time reveal, SSE live updates | No — no admin surface, no server | N/A | go-credentials-store/README.md:66-81 |
| 5 | API-key / Basic auth gate for non-loopback HTTP binds | No — no HTTP listener | N/A | go-credentials-store/README.md:26-28 |
| 6 | Env-var resolution fallback for credentials | No — docs-gates reads no environment variables in non-test code (verified below) | N/A | go-credentials-store/resolve_env.go:20 |
| 7 | Health/status endpoints + Prometheus metrics with build info | No — covered in the go-health / go-health-dashboard reports; no server here | N/A | go-credentials-store/health.go |
| 8 | Master-key management (key files, rotation by KeyID) | No — no master key, nothing to encrypt | N/A | go-credentials-store/README.md:41; go-credentials-store/README.md:54 |
| 9 | go-cqrs-lite event-sourced composition root (system + metaengine/sqlite) | No — no event store, no DB by contract | N/A | go-credentials-store/README.md:13-14 |
| 10 | Operator drill / recovery posture (restart persistence, at-rest scans) | No — nothing at rest | N/A | go-credentials-store/README.md:18-24 |

Dependency note: the store is one of the heaviest modules in the fleet —
32 direct requires including the whole go-cqrs-lite family, templ,
cmdguard/v5, httputil, prometheus/client_golang, spf13/cobra, and
modernc.org/sqlite (go-credentials-store/go.mod:5-36). Even the "thin"
path is a DB plus an event engine. Nothing here can land in a stdlib-only
root module, and nothing should land in pkg/provider either.

## What docs-gates does today

Zero usage, zero presence. Verified:

```console
$ rg -n "go-credentials-store" -g '*.go' -g 'go.mod' -g 'go.sum' /home/lars/projects/docs-gates
# rc=1, no matches in either module
```

The secret/token/env hunt across the whole consumer found no surface to
name:

- No credential or token handling: `rg -n "token|secret|password|credential" -g '*.go'`
  over the repo returns only domain vocabulary — gate *verdict tokens*
  (FEATURES.md status legend, internal/docsvocabulary), the tokenizer's
  word "token" (docsfreshness distinctive-token doctrine), and test
  fixtures. No keychain, no OAuth, no bearer anything.
- No environment reads in production code: the only `os.Getenv`/`LookupEnv`
  sites are test harness PATH manipulation
  (pkg/provider/e2e_test.go:111, pkg/provider/provider_test.go:29) and the
  golden re-pin opt-in (internal/battery/parity_test.go:189 —
  `PARITY_UPDATE`). Nothing resolves secrets from env.
- The only privileged-adjacent operations are local and literal:
  `exec.CommandContext` appears at exactly four sites — git mv for the
  archive verb (internal/docsarchive/docsarchive.go:203), git for verdict
  resolution (internal/docsannotate/git.go:25), git check-ignore for the
  pre-flight stamp guard (internal/docsfreshness/preflight_land.go:74), and
  the provider exec'ing the docs-gates binary itself
  (pkg/provider/provider.go:149). Authentication for any remote git
  operations is the user environment's ssh agent — outside the binary, by
  design.
- Writes are plain `os.WriteFile` under dry-run-default/`--apply` guards
  (e.g. internal/docsarchive/docsarchive.go for the archive verb); no file
  is ever written with secret-class content.
- The consumer's entire security surface is a policy file: SECURITY.md:3-7
  is a supported-versions table plus a private-reporting pointer — there
  is no attack surface section to write because there is no attack
  surface: a stdlib judge with no network, no DB, no env reads, and no
  credentials cannot mishandle what it never holds.

## Gap analysis — should docs-gates adopt it?

**The fleet argument, addressed.** The argument bites when a project
hand-rolls a solved problem. docs-gates does not hand-roll credential
storage, resolution, rotation, or audit — it has no credentials. The one
place the fleet's credential story touches this repo is operational and
deliberately external: git push auth rides the user's ssh environment, and
the auto-commit daemon runs as the user. Pulling an event-sourced vault
into a judge binary would add a SQLite database and 32 requires to a tool
whose entire value proposition is drop-in, zero-state, hermetic (AGENTS.md:
stdlib-only keeps the nix build hermetic; no DB, no network, no secrets).
The correct insulation here is not just defensible, it IS the security
posture: a docs judge that cannot hold secrets cannot leak them, and its
SECURITY.md surface stays a two-question policy file.

**Root-contract bar.** Trivially failed: go-credentials-store/go.mod:5-36
against the zero-require contract (scripts/stack-drift-check.sh:8-9 fails
CI on the first `require`). pkg/provider equally out: its module graph is
go-finding/toolsdk by design, and its HealthCheck is a PATH probe.

RECONSIDER-WHEN: never for the gate/judge surface. The flip condition
would be docs-gates growing an authenticated serving surface (a fleet
dashboard backend, a remote report-corpus fetcher) that must hold
non-interactive credentials — at that point the serving module should
resolve secrets via go-credentials-store (in its own go module, like
pkg/provider), never via env vars or config files.

## Ranked opportunities

None.

## How to verify

```console
$ git -C /home/lars/projects/go-credentials-store log -1 --format='%H'
f9c11ce694762fa66aba2410a1672a8d01a3b7ff   # daemon moved past the audited 6d7caad baseline
$ git -C /home/lars/projects/go-credentials-store tag --sort=-v:refname | head -1
v0.4.3
$ git -C /home/lars/projects/go-credentials-store log --format='%h %s' v0.4.3..HEAD | head -6
f9c11ce chore: auto-commit 1 changed file(s) (heuristic)
48459d1 chore: auto-commit 2 changed file(s) (heuristic)
6d7caad chore: auto-commit 2 changed file(s) (heuristic)   # ← audited baseline
a224d9e chore: auto-commit 2 changed file(s) (heuristic)
e0cbb82 chore: auto-commit 2 changed file(s) (heuristic)
0ad3c6b chore: auto-commit 2 changed file(s) (heuristic)
$ sed -n '5,36p' /home/lars/projects/go-credentials-store/go.mod
# 32 direct requires (go-cqrs-lite family, templ, cmdguard/v5, httputil,
# prometheus, cobra, modernc.org/sqlite, ...)

$ rg -n "go-credentials-store" -g '*.go' -g 'go.mod' -g 'go.sum' /home/lars/projects/docs-gates
# rc=1 — zero references in either module
$ rg -n "os.Getenv|os.LookupEnv" -g '*.go' /home/lars/projects/docs-gates
pkg/provider/e2e_test.go:111:	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
pkg/provider/provider_test.go:29:	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
internal/battery/parity_test.go:189:	update := os.Getenv("PARITY_UPDATE") != ""
# test-only — zero env reads in production code

$ rg -in "token|secret|password|credential" -g '*.go' --glob '!*_test.go' /home/lars/projects/docs-gates | head -4
cmd/docs-gates/verbs.go:117:      verbVocabulary: "FEATURES status tokens must come from the file's own legend",
internal/docsvocabulary/docsvocabulary.go:2:   // status tokens the file's own legend never defines.
internal/docsvocabulary/docsvocabulary.go:9:  // column must only contain tokens the file's own "## Status legend"
internal/docsvocabulary/docsvocabulary.go:53:const Usage = `docs-vocabulary — FEATURES status tokens must come from the file's own legend
# every hit is domain vocabulary (status tokens / tokenizer words), zero secrets

# every exec site (all literal, local, no credentials):
$ rg -n "exec.CommandContext" -g '*.go' /home/lars/projects/docs-gates | rg -v _test.go
internal/docsarchive/docsarchive.go:203:		cmd := exec.CommandContext(context.Background(), "git", "-C", root, "mv", "--", report.rel, destRel)
internal/docsannotate/git.go:25:	cmd := exec.CommandContext(context.Background(), "git", append([]string{"-C", g.root}, args...)...)
internal/docsfreshness/preflight_land.go:74:	probe := exec.CommandContext(
pkg/provider/provider.go:149:	cmd := exec.CommandContext( //nolint:gosec // constant binary, not user input
```
