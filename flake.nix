{
  description = "docs-gates: the docs-health battery gates as one standalone stdlib-only Go binary";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      # Linux only, fleet precedent (the pbx flake): current nixos-unstable
      # no longer evaluates x86_64-darwin, and no consumer rides macOS.
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];

      forAllSystems = nixpkgs.lib.genAttrs systems;

      version = self.shortRev or self.dirtyShortRev or "dev";

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
        pkgs.buildGoModule {
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
            license = pkgs.lib.licenses.mit;
            mainProgram = "docs-gates";
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

      checks = forAllSystems (system: {
        # buildGoModule's default checkPhase runs `go test ./...` — the
        # package build IS the test gate, stdlib-only suites included.
        package = self.packages.${system}.docs-gates;
      });

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            packages = [
              pkgs.go_1_26
              pkgs.golangci-lint
              pkgs.nixfmt-tree
            ];
          };
        }
      );

      formatter = forAllSystems (system: nixpkgs.legacyPackages.${system}.nixfmt-tree);
    };
}
