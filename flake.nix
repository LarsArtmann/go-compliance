{
  description = "go-compliance — Fleet compliance hub: state model, v1 wire contract, realtime dashboard";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    go-nix-helpers = {
      url = "github:LarsArtmann/go-nix-helpers/master";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    # Private Go dep (flake = false; pinned to the exact commit hub/go.mod
    # pseudo-versions, so the nix build compiles what go.sum pins).
    go-daemon = {
      url = "git+ssh://git@github.com/LarsArtmann/go-daemon?rev=e8e1a1fff5028d1e7fc292d0eeb87f8d65e2cf36";
      flake = false;
    };
  };

  outputs =
    inputs@{
      self,
      flake-parts,
      ...
    }:
    let
      # One map, two consumers (go-standard shell env + the manual package
      # build below). Sub-modules are auto-discovered; one entry per repo.
      privateDeps = {
        "github.com/LarsArtmann/go-daemon" = inputs.go-daemon;
      };

      # LarsArtmann deps matching the private glob that are PUBLIC
      # (proxy.golang.org serves them). Base paths cover sub-modules and
      # /vN suffixes.
      publicDeps = [
        "github.com/larsartmann/go-compliance"
        "github.com/larsartmann/go-branded-id"
        "github.com/larsartmann/go-error-family"
        "github.com/larsartmann/cmdguard"
        "github.com/larsartmann/go-datastar"
        "github.com/larsartmann/go-output"
        "github.com/larsartmann/go-sqlitestore"
        "github.com/larsartmann/go-sse"
        "github.com/larsartmann/templ-components"
      ];
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ inputs.go-nix-helpers.flakeModules.go-standard ];

      go-standard = {

        enableCompletions = true;

        completionStyle = "subcommand";
        pname = "compliance-hub";
        description = "Fleet compliance hub: realtime dashboard over BuildFlow evidence";
        goPkgAttr = "go_1_27";

        # deps here drives GOPRIVATE auto-injection into devShells (local
        # `go mod tidy` fetches go-daemon over git). The PACKAGE is built
        # manually below: go-standard's FOD glue double-cds into modRoot
        # (hub) when private deps are set — nixpkgs' own phases already cd —
        # which breaks for sub-module binaries (go-nix-helpers upstream fix
        # pending). Manual buildGoModule uses the default FOD phases.
        enableDefaultPackage = false;
        deps = privateDeps;
        inherit publicDeps;

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
            # Tailwind v4 CLI for the dashboard's compiled CSS pipeline
            # (scripts/regen-ui.sh; CI regenerates and diffs byte-fresh).
            ps.tailwindcss_4
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
          lib,
          ...
        }:
        let
          version = self.rev or self.dirtyRev or "dev";

          mkPreparedSource = import "${inputs.go-nix-helpers}/mkPreparedSource.nix" {
            inherit pkgs lib;
            goPkg = pkgs.go_1_27;
          };

          # Private-dep injection for the sandboxed build: copies go-daemon
          # into _local_deps/ and injects replace directives (incl. the hub
          # sub-module, auto-discovered).
          preparedSrc = mkPreparedSource {
            name = "compliance-hub";
            inherit version;
            src = self;
            deps = privateDeps;
            inherit publicDeps;
            # mkPreparedSource injects replace directives into the ROOT go.mod
            # only; the hub sub-module needs its own. Replace paths resolve
            # relative to the go.mod they live in, hence ../_local_deps.
            postPatchExtra = ''
              echo 'replace github.com/LarsArtmann/go-daemon => ../_local_deps/go-daemon' >> hub/go.mod

              # `go mod vendor` copies a directory-replaced module verbatim
              # (flake.nix included) into the FOD output. flake.nix carries
              # this derivation's own vendorHash, so the FOD hashed its own
              # placeholder — an unfixable flip-flop (2026-11-18: two builds
              # at different dirtyRevs diffed exactly by that one line, via
              # the two stored go-modules outputs). Repoint the core replace
              # at a Go-files-only copy: no self-reference, and vendorHash
              # no longer churns on doc-only edits.
              mkdir -p _core_only
              cp go.mod go.sum _core_only/
              cp *.go _core_only/
              sed -i 's|=> \.\./$|=> ../_core_only|' hub/go.mod
            '';
          };

          compliance-hub = (pkgs.buildGoModule.override { go = pkgs.go_1_27; }) {
            pname = "compliance-hub";
            inherit version;
            src = preparedSrc;
            # go.work spans core + hub; module-mode build of the hub module
            # (its replace => ../ resolves the core module in-tree).
            postPatch = "rm -f go.work";
            modRoot = "hub";
            subPackages = [ "./cmd/compliance-hub" ];
            vendorHash = "sha256-kJhOLXuAV/KNQBpzaEQEb8Ggj26JoGJyatKojkeVHeg=";
            ldflags = [
              "-s"
              "-w"
            ];
            meta = {
              description = "Fleet compliance hub: realtime dashboard over BuildFlow evidence";
              mainProgram = "compliance-hub";
              license = lib.licenses.mit;
              platforms = lib.platforms.unix;
              maintainers = [
                {
                  name = "Lars Artmann";
                  github = "LarsArtmann";
                }
              ];
            };
          };

          # Core module (root, library — public deps only): compile + test in
          # the sandbox so `nix flake check` covers BOTH modules.
          core-build = (pkgs.buildGoModule.override { go = pkgs.go_1_27; }) {
            pname = "go-compliance-core";
            inherit version;
            src = self;
            postPatch = "rm -f go.work";
            vendorHash = "sha256-KKGkAmStPL8De3sxGZUl6q54bk+v/Vw1zrWAvxxPFYo=";
            subPackages = [ "." ];
            doCheck = true;
          };
        in
        {
          packages = {
            default = compliance-hub;
            inherit compliance-hub;
          };

          checks = {
            hub-build = compliance-hub;
            inherit core-build;
          };
        };
    };
}
