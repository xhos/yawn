{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    git-hooks.url = "github:cachix/git-hooks.nix";
    git-hooks.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = {
    nixpkgs,
    git-hooks,
    ...
  }: let
    forAllSystems = function:
      nixpkgs.lib.genAttrs nixpkgs.lib.systems.flakeExposed (
        system: function nixpkgs.legacyPackages.${system}
      );
  in {
    checks = forAllSystems (pkgs: {
      pre-commit = git-hooks.lib.${pkgs.system}.run {
        src = ./.;
        hooks = {
          govet.enable = true;
          gofmt.enable = true;
          alejandra.enable = true;
        };
      };
    });

    devShells = forAllSystems (pkgs: {
      default = pkgs.mkShell {
        packages = with pkgs; [
          go
        ];
      };
    });
  };
}
