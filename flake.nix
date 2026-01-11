{
  description = "Hyprdocked";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

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
          cfg = config.services.hyprdocked;
        in
        {
          options.services.hyprdocked = {
            enable = lib.mkEnableOption "Hyprdocked Listener";

            package = lib.mkOption {
              type = lib.types.package;
              inherit (self.packages.${pkgs.stdenv.hostPlatform.system}) default;
              description = "The hyprdocked package to use";
            };
          };

          config = lib.mkIf cfg.enable {
            home.packages = [ cfg.package ];
            systemd.user.services.hyprdocked = {
              Unit = {
                Description = "Hyprdocked Listener";
                After = [ "graphical-session.target" ];
              };

              Service = {
                ExecStart = "${cfg.package}/bin/hyprdocked";
                Restart = "on-failure";
                RestartSec = 2;
              };

              Install = {
                WantedBy = [ "wayland-session@Hyprland.target" ];
              };
            };
          };
        };
    in
    {
      homeManagerModules.default = homeManagerModule;
      homeManagerModules.hyprdocked = homeManagerModule;
    }
    // (flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        inherit (pkgs) callPackage;
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
