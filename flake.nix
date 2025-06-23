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
            (pkgs.writeShellScriptBin "akami-build" ''
              echo "Building akami binaries for all platforms..."
              mkdir -p $PWD/bin

              # Build for Linux (64-bit)
              echo "Building for Linux (x86_64)..."
              GOOS=linux GOARCH=amd64 go build -o $PWD/bin/akami-linux-amd64 ./main.go

              # Build for Linux ARM (64-bit)
              echo "Building for Linux (aarch64)..."
              GOOS=linux GOARCH=arm64 go build -o $PWD/bin/akami-linux-arm64 ./main.go

              # Build for macOS (64-bit Intel)
              echo "Building for macOS (x86_64)..."
              GOOS=darwin GOARCH=amd64 go build -o $PWD/bin/akami-darwin-amd64 ./main.go

              # Build for macOS ARM (64-bit)
              echo "Building for macOS (aarch64)..."
              GOOS=darwin GOARCH=arm64 go build -o $PWD/bin/akami-darwin-arm64 ./main.go

              # Build for Windows (64-bit)
              echo "Building for Windows (x86_64)..."
              GOOS=windows GOARCH=amd64 go build -o $PWD/bin/akami-windows-amd64.exe ./main.go

              echo "All binaries built successfully in $PWD/bin/"
              ls -la $PWD/bin/
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
        akami-build = {
          type = "app";
          program = "${self.devShells.${pkgs.system}.default.inputDerivation}/bin/akami-build";
        };
      });
    };
}
