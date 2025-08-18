{
  description = "A tiny cli for Linux that quickly connects and disconnects paired bluetooth devices";
  inputs = {
    devshell.url = "github:numtide/devshell";
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    systems.url = "github:nix-systems/default";
  };
  outputs = inputs @ {
    self,
    flake-parts,
    ...
  }:
    inputs.flake-parts.lib.mkFlake {inherit inputs;} (top @ {
      config,
      withSystem,
      moduleWithSystem,
      ...
    }: {
      systems = import inputs.systems;
      imports = [
        inputs.devshell.flakeModule
      ];
      perSystem = {
        self,
        pkgs,
        config,
        lib,
        system,
        ...
      }: {
        packages.default = pkgs.callPackage ./. {
          inherit pkgs;
          self = top.self;
        };

        devshells.default = {
          env = [];
          packages = with pkgs; [
            go_1_24
          ];
        };
      };

      # Add overlays to the root output set
      flake.overlays = {
        default = final: prev: {
          btsw = final.callPackage ./. {
            pkgs = final;
            self = self;
          };
        };
      };
    });
}
