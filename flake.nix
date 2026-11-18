{
  description = "go-compliance — Fleet compliance hub: state model, v1 wire contract, realtime dashboard";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    go-nix-helpers = {
      url = "git+ssh://git@github.com/LarsArtmann/go-nix-helpers?ref=master";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{
      self,
      flake-parts,
      ...
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ inputs.go-nix-helpers.flakeModules.go-standard ];

      go-standard = {
        pname = "compliance-hub";
        description = "Fleet compliance hub: realtime dashboard over BuildFlow evidence";
        goPkgAttr = "go_1_27";

        # Placeholder until the first `nix build` prints the real hash (F084).
        vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";

        # The binary lives in the hub sub-module (own go.mod at ./hub).
        subPackages = [ "./cmd/compliance-hub" ];
        extraBuildAttrs = {
          modRoot = "hub";
          # go.work spans core + hub; module-mode build of the hub module only.
          postPatch = "rm -f go.work";
        };

        # Go formatting keeps a dedicated gate (dprint for md/json/yaml;
        # gofmt via the Go toolchain) — mirrors the bimi-lens decision to
        # not double-source Go formatting in treefmt. Nix files DO format.
        enableGofumpt = false;
        enableGoimports = false;
        enableNixfmt = true;

        devShellExtraPackages =
          ps:
          let
            # templ GENERATOR pinned to the go.mod tool directive (v0.3.1100):
            # pkgs.templ drifts behind the pin and its older parser can
            # mangle generated output (templ-components precedent). The tag
            # ships a stale vendor/ dir — remove and vendor from go.mod.
            templGen = (ps.buildGoModule.override { go = ps.go_1_27; }) {
              pname = "templ-gen";
              version = "v0.3.1100";
              src = ps.fetchFromGitHub {
                owner = "a-h";
                repo = "templ";
                rev = "v0.3.1100";
                hash = "sha256-QViQAMtCqaXS4GodlSCW/nE4Sox+EF49sbb7h9RBIS0=";
              };
              vendorHash = "sha256-MMAoHofeZzGNFsjxY4nrPRyaiKOk4fLYY+kIVmdAL3g=";
              postPatch = "rm -rf vendor";
              subPackages = [ "cmd/templ" ];
            };
          in
          [
            templGen
            ps.dprint
            ps.actionlint
          ];

        devShellHook = ''
          # Two-module workspace (go.work): go-standard forces GOWORK=off for
          # hermetic single-module builds; the workspace IS the dev surface
          # here (go test ./... spans core + hub). GOWORK=off stays available
          # per-command for module isolation checks.
          unset GOWORK
          echo "go-compliance dev shell — $(go version)"
          echo "Workspace: core + hub (go.work)"
        '';
      };

      perSystem =
        {
          pkgs,
          ...
        }:
        {
          # Core module (root, library): compile + test in the sandbox so
          # `nix flake check` covers BOTH modules, not just the hub binary.
          checks.core-build = (pkgs.buildGoModule.override { go = pkgs.go_1_27; }) {
            pname = "go-compliance-core";
            version = self.rev or self.dirtyRev or "dev";
            src = self;
            postPatch = "rm -f go.work";
            # Placeholder until the first evaluation prints the real hash.
            vendorHash = "sha256-BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=";
            # Library root package: compiles, installs nothing.
            subPackages = [ "." ];
            doCheck = true;
          };
        };
    };
}
