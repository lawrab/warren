{
  description = "Warren - A Hyprland-optimized file manager";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in
      {
        # Development shell
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go development
            go             # Latest Go version
            gopls          # Go language server
            gotools        # goimports, godoc, etc.
            go-tools       # staticcheck, etc.
            golangci-lint  # Linting
            delve          # Debugger

            # GTK4 and dependencies
            gtk4
            gtk4-layer-shell
            pkg-config
            glib
            gobject-introspection

            # Build tools
            gcc
            gnumake

            # Development tools
            git

            # Nice-to-have tools
            air            # Live reload for Go

            # For testing and debugging
            strace
            ltrace
            gdb
          ];

          shellHook = ''
            echo "🐰 Warren development environment"
            echo ""
            echo "Available tools:"
            echo "  go version: $(go version)"
            echo "  gopls: Go language server"
            echo "  golangci-lint: Linting"
            echo "  air: Live reload"
            echo ""
            echo "GTK4 development libraries loaded"
            echo ""
            echo "Quick start:"
            echo "  go mod init github.com/lawrab/warren"
            echo "  go get github.com/diamondburned/gotk4/pkg/gtk/v4"
            echo "  go run cmd/warren/main.go"
            echo ""
            echo "Documentation: ./docs/"
            echo ""

            # Set up Go environment
            export GOPATH="$HOME/go"
            export PATH="$GOPATH/bin:$PATH"

            # GTK environment
            export PKG_CONFIG_PATH="${pkgs.gtk4.dev}/lib/pkgconfig:${pkgs.glib.dev}/lib/pkgconfig:$PKG_CONFIG_PATH"

            # Enable CGO (required for GTK4 bindings)
            export CGO_ENABLED=1
          '';

          # Environment variables
          env = {
            # Development mode
            WARREN_DEV = "1";

            # GTK settings for development
            GTK_DEBUG = "interactive";  # Enable GTK inspector with Ctrl+Shift+I
          };
        };

        # Package definition (for when we want to build)
        packages.default = pkgs.buildGoModule {
          pname = "warren";
          version = "0.1.0-dev";
          src = ./.;

          vendorHash = null; # Will need to be updated after go mod vendor

          nativeBuildInputs = with pkgs; [
            pkg-config
            gobject-introspection
          ];

          buildInputs = with pkgs; [
            gtk4
            glib
          ];

          # Build flags
          ldflags = [
            "-s"
            "-w"
            "-X main.Version=${self.version}"
          ];

          meta = with pkgs.lib; {
            description = "A Hyprland-optimized file manager";
            homepage = "https://github.com/lawrab/warren";
            license = licenses.mit; # Update as needed
            platforms = platforms.linux;
            mainProgram = "warren";
          };
        };

        # Apps definition for `nix run`
        apps.default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/warren";
        };
      }
    );
}
