{
  description = "🌷 the cutsie hackatime helper";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      allSystems = [
        "x86_64-linux" # 64-bit Intel/AMD Linux
        "aarch64-linux" # 64-bit ARM Linux
        "x86_64-darwin" # 64-bit Intel macOS
        "aarch64-darwin" # 64-bit ARM macOS
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {
        pkgs = import nixpkgs { inherit system; };
      });
    in
    {
      packages = forAllSystems ({ pkgs }: {
        default = pkgs.buildGoModule {
          pname = "akami";
          version = "0.0.1";
          subPackages = [ "." ];  # Build from root directory
          src = ./.;
          vendorHash = "sha256-9gO00c3D846SJl5dbtfj0qasmONLNxU/7V1TG6QEaxM=";
        };
      });

      devShells = forAllSystems ({ pkgs }: {
        default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
            (pkgs.writeShellScriptBin "akami-dev" ''
              go build -o ./bin/akami ./main.go
              ./bin/akami "$@" || true
            '')
          ];

          shellHook = ''
            export PATH=$PATH:$PWD/bin
            mkdir -p $PWD/bin
          '';
        };
      });

      apps = forAllSystems ({ pkgs }: {
        default = {
          type = "app";
          program = "${self.packages.${pkgs.system}.default}/bin/akami";
        };
        akami-dev = {
          type = "app";
          program = toString (pkgs.writeShellScript "akami-dev" ''
            go build -o ./bin/akami ./main.go
            ./bin/akami $* || true
          '');
        };
      });
    };
}
