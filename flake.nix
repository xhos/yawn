{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    git-hooks = {
      url = "github:cachix/git-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    self,
    nixpkgs,
    git-hooks,
    ...
  }: let
    forAllSystems = function:
      nixpkgs.lib.genAttrs nixpkgs.lib.systems.flakeExposed (
        system: function nixpkgs.legacyPackages.${system}
      );
  in {
    # core package
    packages = forAllSystems (pkgs: {
      yawn = pkgs.buildGoModule {
        pname = "yawn";
        version = "0.1.2";
        src = ./.;
        vendorHash = "sha256-RNbS40G+8rtwlSJgYLN1puTCytGfXdagQTEs6sIXwnM=";
        ldflags = ["-s" "-w"];
        subPackages = ["cmd/yawn"];
      };

      default = self.packages.${pkgs.stdenv.hostPlatform.system}.yawn;
    });

    # dev stuff
    checks = forAllSystems (pkgs: {
      pre-commit = git-hooks.lib.${pkgs.stdenv.hostPlatform.system}.run {
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

          (pkgs.writeShellScriptBin "stub" ''
            go run ./cmd/greetd-stub "$@"
          '')
        ];
        shellHook = self.checks.${pkgs.stdenv.hostPlatform.system}.pre-commit.shellHook;
      };
    });

    # test VM
    nixosConfigurations.test-vm = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        ({modulesPath, ...}: {
          imports = [(modulesPath + "/virtualisation/qemu-vm.nix")];

          virtualisation.graphics = true;

          services.greetd = {
            enable = true;
            settings.default_session = {
              command = "${self.packages.x86_64-linux.yawn}/bin/yawn -cmd bash -user test";
              user = "greeter";
            };
          };

          users.users.test = {
            isNormalUser = true;
            initialPassword = "test";
          };

          networking.hostName = "yawn";
          system.stateVersion = "24.11";
        })
      ];
    };

    apps.x86_64-linux.test-vm = {
      type = "app";
      program = "${self.nixosConfigurations.test-vm.config.system.build.vm}/bin/run-yawn-vm";
    };
  };
}
