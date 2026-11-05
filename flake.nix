{
  description = "docs-gates: the docs-health battery gates as one standalone stdlib-only Go binary";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  # The //nolint:erraudit directives are live only when the erraudit tool
  # is present: this input puts the pinned binary in the devShell and
  # powers the nolint-audit staleness check. The repo is PRIVATE (gh,
  # verified 2026-11-04), so the URL must be git+ssh; flake.lock pins the
  # rev, so rule behavior cannot drift until an explicit
  # `nix flake update` (BuildFlow's track-latest input policy, 2026-10-05).
  # The package consumed is erraudit's own hermetic build (its
  # go-standard FOD pins its private transitive deps) — none of that
  # touches this repo's stdlib-only module contract.
  inputs.erraudit.url = "git+ssh://git@github.com/LarsArtmann/erraudit";

  # Evaluate erraudit against THIS flake's nixpkgs instead of its own
  # lock node (they sit on the same rev today, so behavior is identical;
  # the follows keeps them from diverging and dedupes nixpkgs_2 from the
  # lock). Verified: both erraudit checks still build green after the
  # follows (2026-11-04).
  inputs.erraudit.inputs.nixpkgs.follows = "nixpkgs";

  # The commit-time nolint-audit leg (the bank-sync pre-commit pattern):
  # the flake checks keep directives honest per `nix flake check`; this
  # input lets every `git commit` run the same staleness gate BEFORE the
  # rot lands. Follows this flake's nixpkgs to dedupe the lock.
  inputs.git-hooks = {
    url = "github:cachix/git-hooks.nix";
    inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs =
    {
      self,
      nixpkgs,
      erraudit,
      git-hooks,
    }:
    let
      lib = nixpkgs.lib;

      # Linux only, fleet precedent (the pbx flake): current nixos-unstable
      # no longer evaluates x86_64-darwin, and no consumer rides macOS.
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];

      forAllSystems = lib.genAttrs systems;

      version = self.shortRev or self.dirtyShortRev or "dev";

      # One toolchain line for the package build AND the devShell. 1.27,
      # not 1.26 (the root go.mod floor): the nested pkg/provider module
      # pins go 1.27, and every surface sets GOTOOLCHAIN=local — go_1_26
      # would reject the provider module ("go.mod requires go >= 1.27"),
      # and local is what forbids go's auto mode from silently fetching a
      # newer toolchain from the network.
      goPkgAttr = "go_1_27";

      # One binding for erraudit's hermetic package, shared by the checks
      # and the devShell (was duplicated per-output let). Call per system.
      errauditPkg = system: erraudit.packages.${system}.default;

      # One pre-commit configuration, shared by the sandboxed
      # `pre-commit-check` and the devShell hook install (was duplicated
      # per-output). The nolint-audit hook is the commit-time leg of the
      # directive-honesty gate; the flake `nolint-audit` check stays the
      # push-time backstop. Sandboxed-safe: the /tmp caches mirror the
      # flake check's plumbing, because the hook must run green BOTH in
      # `nix flake check` (read-only, no network, no HOME) and in the
      # devShell.
      preCommitCheckFor =
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        git-hooks.lib.${system}.run {
          src = ./.;
          hooks.nolint-audit = {
            enable = true;
            name = "nolint-audit";
            files = "\\.go$";
            pass_filenames = false;
            language = "unsupported";
            entry = "${pkgs.bash}/bin/bash -c 'PATH=${pkgs.${goPkgAttr}}/bin:$PATH GOTOOLCHAIN=local GOWORK=off GOPROXY=off GOCACHE=/tmp/dg-hooks-gocache GOPATH=/tmp/dg-hooks-gopath exec ${(errauditPkg system)}/bin/erraudit nolint-audit .'";
          };
        };

      # The legacy standalone binary names. Each wrapper execs the one
      # multi-call binary with the matching verb, so batteries and docs
      # written against the old names keep working unchanged; each verb
      # answers --version under its legacy name (see cmd/docs-gates).
      legacyWrappers = [
        {
          name = "docs-freshness";
          verb = "freshness";
        }
        {
          name = "docs-annotate";
          verb = "annotate";
        }
        {
          name = "docs-routing";
          verb = "routing";
        }
        {
          name = "docs-cadence";
          verb = "cadence";
        }
        {
          name = "docs-strike";
          verb = "strike";
        }
      ];

      docsGatesFor =
        pkgs:
        (pkgs.buildGoModule.override { go = pkgs.${goPkgAttr}; }) {
          pname = "docs-gates";
          inherit version;

          src = ./.;

          # stdlib-only by contract (AGENTS.md): no module requires, so
          # the build stays hermetic without a vendor hash.
          vendorHash = null;

          subPackages = [ "cmd/docs-gates" ];

          env.CGO_ENABLED = "0";

          ldflags = [ "-X main.version=${version}" ];

          meta = {
            description = "docs-health battery gates: freshness, annotate, routing, cadence, strike";
            homepage = "https://github.com/Artmann-Technologies/docs-gates";
            license = lib.licenses.mit;
            mainProgram = "docs-gates";
            platforms = lib.platforms.linux;
          };
        };
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          docs-gates = docsGatesFor pkgs;

          wrappers = map (
            entry:
            pkgs.writeShellScriptBin entry.name ''
              exec ${docs-gates}/bin/docs-gates ${entry.verb} "$@"
            ''
          ) legacyWrappers;
        in
        {
          docs-gates = docs-gates;

          default = pkgs.symlinkJoin {
            name = "docs-gates-${version}";

            paths = [ docs-gates ] ++ wrappers;

            meta = docs-gates.meta;
          };
        }
      );

      checks = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          # buildGoModule's default checkPhase runs `go test ./...` — the
          # package build IS the test gate, stdlib-only suites included.
          package = self.packages.${system}.docs-gates;

          # The wrappers + symlinkJoin surface the fleet actually installs.
          default = self.packages.${system}.default;

          # The five legacy names are API (flake wrappers exec'ing the
          # multi-call binary): each must resolve on PATH and answer
          # --version under its legacy name — with the LDFLAGS-STAMPED
          # version string, not just exit 0 (a stale stamp under a legacy
          # name would silently lie about what is installed).
          legacy-wrappers =
            pkgs.runCommand "docs-gates-legacy-wrappers-smoke"
              {
                nativeBuildInputs = [ self.packages.${system}.default ];
              }
              ''
                want="${version}"

                for name in docs-gates docs-freshness docs-annotate docs-routing docs-cadence docs-strike; do
                  got="$($name --version | awk '{print $NF}')"
                  if [ "$got" != "$want" ]; then
                    echo "FAIL: $name --version stamped '$got', want '$want'" >&2
                    exit 1
                  fi
                done

                touch $out
              '';

          # The stale-nolint cadence (the source-fleet lesson: three stale
          # //nolint:erraudit directives shipped green for days before one
          # session ran the audit): staleness-checks every erraudit
          # directive in the ROOT module, so a directive whose finding is
          # gone — or a new suppression with no finding behind it — fails
          # the check. Module-scoped by contract: nolint-audit refuses to
          # cross nested go.mod boundaries, and pkg/provider carries no
          # erraudit directives (verified 2026-11-04).
          nolint-audit =
            pkgs.runCommand "docs-gates-nolint-audit"
              {
                nativeBuildInputs = [
                  (errauditPkg system)
                  pkgs.${goPkgAttr}
                ];
                env = {
                  GOTOOLCHAIN = "local";
                  GOWORK = "off";
                  GOPROXY = "off";
                  GOCACHE = "/tmp/dg-nolint-audit-gocache";
                  GOPATH = "/tmp/dg-nolint-audit-gopath";
                };
              }
              ''
                cd ${self}
                erraudit nolint-audit .
                # Discrimination leg: the check must be able to FAIL. The
                # permanent planted-stale-directive test (internal/battery)
                # runs here, where erraudit is on PATH — a bare `go test`
                # skips it, so the flake gate is what keeps the proof alive.
                go test ./internal/battery -run TestNolintAuditDiscrimination -count=1
                touch $out
              '';

          # The findings axis (nolint-audit only proves directives honest;
          # this proves the CODE is): erraudit's type-aware legacyerrors
          # audit (errors.As/Is → AsType) over the ROOT module must stay
          # green — every migratable site is either migrated or carries a
          # live //nolint:erraudit directive. Same module scope, same
          # hermetic sandbox plumbing. Discrimination-proven 2026-11-04
          # (planted errors.As → check red, naming the site).
          erraudit-lint =
            pkgs.runCommand "docs-gates-erraudit-lint"
              {
                nativeBuildInputs = [
                  (errauditPkg system)
                  pkgs.${goPkgAttr}
                ];
                env = {
                  GOTOOLCHAIN = "local";
                  GOWORK = "off";
                  GOPROXY = "off";
                  GOCACHE = "/tmp/dg-erraudit-lint-gocache";
                  GOPATH = "/tmp/dg-erraudit-lint-gopath";
                };
              }
              ''
                cd ${self}
                erraudit lint ./... --type-aware
                touch $out
              '';

          # The nested provider module's suite (root `./...` never sees
          # it — AGENTS.md gotcha). Its own buildGoModule: public deps
          # fetched by vendorHash, so the checkPhase go test runs
          # hermetically. This does NOT touch the root stdlib-only
          # contract, whose go.mod still carries zero requires.
          provider = (pkgs.buildGoModule.override { go = pkgs.${goPkgAttr}; }) {
            pname = "docs-gates-provider-suite";
            inherit version;

            src = "${self}/pkg/provider";

            vendorHash = "sha256-qOX4CkUkDtAiNSewsMAJN/4Zk1FZe27E4pF0bHonwVU=";

            env = {
              CGO_ENABLED = "0";
              GOTOOLCHAIN = "local";
            };

            # A green suite must MEAN something: the e2e leg skips in this
            # sandbox (only pkg/provider is vendored in), and without this
            # proof the check would pass no matter which arm fired — or if
            # the leg silently vanished. Assert the sentinel logged the
            # SANCTIONED skip (23-06 report §f.19).
            checkPhase = ''
              runHook preCheck
              go test ./... -v > provider-test.log 2>&1
              cat provider-test.log 1>&2
              grep -q "E2E LEG: skipped in this environment" provider-test.log || {
                echo "FATAL: no e2e sentinel line in the provider log — the suite is green but its end-to-end leg is unaccounted"
                exit 1
              }
              grep -q "root module source not adjacent" provider-test.log || {
                echo "FATAL: the e2e skip is not the sanctioned sandbox arm (root module source not adjacent)"
                exit 1
              }
              runHook postCheck
            '';

            meta = {
              description = "docs-gates provider module test suite check";
              platforms = lib.platforms.linux;
            };
          };

          # Rule-behavior determinism: the erraudit input tracks master,
          # so `nix flake update` can silently swap the rule set. The
          # tripwire pins the locked rev in scripts/check-erraudit-rev.sh
          # and fails until that pin is bumped deliberately (precedent:
          # go-codec's check-erraudit-version.sh).
          erraudit-rev =
            pkgs.runCommand "docs-gates-erraudit-rev"
              {
                nativeBuildInputs = [
                  pkgs.bash
                  pkgs.jq
                ];
              }
              ''
                cd ${self}
                bash scripts/check-erraudit-rev.sh
                touch $out
              '';

          # Golangci drift guard: the CI action version must match the
          # flake's nixpkgs-locked golangci-lint (the engine local dev
          # runs at 0 findings). Drift means CI lints a different
          # rule set than the devShell — findings diverge silently in
          # BOTH directions. Eval-time interpolation of
          # pkgs.golangci-lint.version keeps the check hermetic (no
          # network, the lock is the truth).
          golangci-drift =
            pkgs.runCommand "docs-gates-golangci-drift" { nativeBuildInputs = [ pkgs.bash ]; }
              ''
                cd ${self}
                want="${pkgs.golangci-lint.version}"
                got_all="$(grep -h '^          version: ' .github/workflows/ci.yml | awk '{print $2}' | sort -u)"
                count="$(printf '%s\n' "$got_all" | grep -c .)"
                if [ "$count" -ne 2 ]; then
                  echo "FATAL: expected exactly 2 pinned golangci action versions (go + provider jobs), found $count:" >&2
                  printf '%s\n' "$got_all" >&2
                  exit 1
                fi
                if [ "$got_all" != "v$want" ]; then
                  echo "FATAL: golangci drift — CI action pins $got_all but the flake's locked nixpkgs builds v$want." >&2
                  echo "Remedy: align .github/workflows/ci.yml's action version with the flake pin (or bump the nixpkgs input deliberately), in ONE commit." >&2
                  exit 1
                fi
                touch $out
              '';

          # Shell + workflow lint: shellcheck over the repo's scripts
          # (safe-strike and its refusal-matrix harness are the
          # corruption guardrails — they must lint clean themselves)
          # and actionlint over the CI workflows.
          shell-lint =
            pkgs.runCommand "docs-gates-shell-lint"
              {
                nativeBuildInputs = [
                  pkgs.bash
                  pkgs.shellcheck
                  pkgs.actionlint
                ];
              }
              ''
                cd ${self}
                # --severity warning: SC2016 info fires on the harness's
                # deliberately literal backtick verdict strings
                shellcheck --severity warning scripts/*.sh
                actionlint
                touch $out
              '';

          # The commit-time leg of the directive-honesty gate: pre-commit
          # runs the same `erraudit nolint-audit .` staleness check on
          # every commit touching Go files (see preCommitCheckFor).
          pre-commit-check = preCommitCheckFor system;
        }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};

          # The pre-commit machinery: shellHook installs the git hooks and
          # enabledPackages carries the `pre-commit` binary itself.
          inherit (preCommitCheckFor system) shellHook enabledPackages;
        in
        {
          # No C compiler anywhere in this repo: the package builds
          # CGO_ENABLED=0 and the suites are stdlib-only.
          default = pkgs.mkShellNoCC {
            packages = [
              pkgs.${goPkgAttr}
              pkgs.golangci-lint
              pkgs.nixfmt-tree
              pkgs.actionlint
              pkgs.shellcheck
              (errauditPkg system)
            ]
            ++ enabledPackages;

            inherit shellHook;

            GOTOOLCHAIN = "local";
            CGO_ENABLED = "0";
          };
        }
      );

      formatter = forAllSystems (system: nixpkgs.legacyPackages.${system}.nixfmt-tree);
    };
}
