{
  description = "ssm - a TUI for managing ssh connections";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "ssm";
          version = "0.0.1";
          src = ./.;
          vendorHash = null;

          meta = with pkgs.lib; {
            description = "ssm - a TUI for managing ssh connections";
            license = licenses.mit;
            maintainers = [ ];
            mainProgram = "ssm";
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            go-tools
          ];
        };
      });
}