{
  description = "Hyprlaptop";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.gomod2nix.url = "github:nix-community/gomod2nix";
  inputs.gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
  inputs.gomod2nix.inputs.flake-utils.follows = "flake-utils";

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      gomod2nix,
    }:
    let
      homeManagerModule =
        {
          config,
          lib,
          pkgs,
          ...
        }:
        let
          cfg = config.services.hyprlaptop;
        in
        {
          options.services.hyprlaptop = {
            enable = lib.mkEnableOption "Hyprlaptop Listener";

            package = lib.mkOption {
              type = lib.types.package;
              default = self.packages.${pkgs.system}.default;
              description = "The hyprlaptop package to use";
            };
          };

          config = lib.mkIf cfg.enable {
            home.packages = [ cfg.package ];
            systemd.user.services.hyprlaptop = {
              Unit = {
                Description = "Hyprlaptop Listener";
                After = [ "graphical-session.target" ];
              };

              Service = {
                ExecStart = "${cfg.package}/bin/hyprlaptop listen";
                Restart = "on-failure";
                RestartSec = 2;
              };

              Install = {
                WantedBy = [ "graphical-session.target" ];
              };
            };
          };
        };
    in
    {
      homeManagerModules.default = homeManagerModule;
      homeManagerModules.hyprlaptop = homeManagerModule;
    }
    // (flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        callPackage = pkgs.callPackage;
        go-lint = pkgs.stdenvNoCC.mkDerivation {
          name = "go-lint";
          dontBuild = true;
          src = ./.;
          doCheck = true;
          nativeBuildInputs = with pkgs; [
            golangci-lint
            go
            writableTmpDirAsHomeHook
          ];
          checkPhase = ''
            golangci-lint run
          '';
          installPhase = ''
            mkdir "$out"
          '';
        };
      in
      {
        checks = {
          inherit go-lint;
        };
        packages.default = callPackage ./. {
          inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
        };
        devShells.default = callPackage ./shell.nix {
          inherit (gomod2nix.legacyPackages.${system}) mkGoEnv gomod2nix;
        };
      }
    ));
}
