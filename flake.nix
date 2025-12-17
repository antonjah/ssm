{
  description = "ssm - a TUI for managing ssh connections";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let
      mkPackage = pkgs: pkgs.buildGoModule {
        pname = "ssm";
        version = "1.0.3";
        src = ./.;
        vendorHash = "sha256-WukLNCsgqAck5LMmY//kFfYst2U1JmKt73BT4H4QcVQ=";

        meta = with pkgs.lib; {
          description = "ssm - a TUI for managing ssh connections";
          homepage = "https://github.com/antonjah/ssm";
          license = licenses.mit;
          maintainers = [ ];
          mainProgram = "ssm";
        };
      };

      homeManagerModule = { config, lib, pkgs, ... }:
        let
          cfg = config.programs.ssm;
        in
        {
          options.programs.ssm = {
            enable = lib.mkEnableOption "ssm - a TUI for managing ssh connections";

            package = lib.mkOption {
              type = lib.types.package;
              default = mkPackage pkgs;
              defaultText = lib.literalExpression "pkgs.ssm";
              description = "The ssm package to use.";
            };
          };

          config = lib.mkIf cfg.enable {
            home.packages = [ cfg.package ];
          };
        };
    in
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages = {
          default = mkPackage pkgs;
          ssm = mkPackage pkgs;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            go-tools
          ];
        };
      }) // {
      overlays.default = final: prev: {
        ssm = mkPackage final;
      };

      homeModules.default = homeManagerModule;
      homeModules.ssm = homeManagerModule;
    };
}
