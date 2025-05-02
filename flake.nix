{
  description = "cdktf-oci with golang";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    devshell.url = "github:numtide/devshell";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    devshell,
  }:
    flake-utils.lib.eachDefaultSystem (system: {
      devShell = let
        pkgs = import nixpkgs {
          inherit system;

          overlays = [devshell.overlays.default];
          config.allowUnfree = true;
        };
      in
        pkgs.devshell.mkShell {imports = [(pkgs.devshell.importTOML ./devshell.toml)];};
    });
}
