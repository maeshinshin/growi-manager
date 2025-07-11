{
  description = "cdktf-oci with golang and helmify";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    devshell.url = "github:numtide/devshell";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { self
    , nixpkgs
    , flake-utils
    , devshell
    ,
    }:
    flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        inherit system;

        overlays = [ devshell.overlays.default ];
        config.allowUnfree = true;
      };

      helmify = pkgs.stdenv.mkDerivation rec {
        pname = "helmify";
        version = "0.4.18";
        src =
          let
            sources = {
              "arm64-darwin" = {
                file = "helmify_Darwin_arm64.tar.gz";
                hash = "e0e621cf19792089c7dc030ef6fe16574d48d1fb524ce231a3ec48b371c07bb7";
              };
              "x86_64-darwin" = {
                file = "helmify_Darwin_x86_64.tar.gz";
                hash = "e2096319e23ee4ada0931b1863491c271fea7ec5830327fb599ea622638c9f1b";
              };
              "arm64-linux" = {
                file = "helmify_Linux_arm64.tar.gz";
                hash = "bf8f72efe836ad9f97f8d02cf6dec2ec56f1c4df901d65c82b33ed9e53b3e575";
              };
              "i386-linux" = {
                file = "helmify_Linux_i386.tar.gz";
                hash = "644100c660c2d5e202d255ca17a489a2b4b04320dc91715183edbf245a9468ca";
              };
              "x86_64-linux" = {
                file = "helmify_Linux_x86_64.tar.gz";
                hash = "367a1c310bdd65efe65075e275bdd2b802a8e19168cc755c3a39e63ca1ff104c";
              };
              "i386-windows" = {
                file = "helmify_Windows_i386.zip";
                hash = "7d4f713b7ae79678d5196f11acc7fe9fff949dba31bca4716c8ce5d6d641dae2";
              };
              "x86_64-windows" = {
                file = "helmify_Windows_x86_64.zip";
                hash = "4bb9635cc042ba6d09b70792be7e105f796985be9a40505802e7323e303711dd";
              };
            };
            source = sources.${pkgs.system} or (throw "helmify: Unsupported system: ${pkgs.system}");
          in
          pkgs.fetchurl {
            url = "https://github.com/arttor/helmify/releases/download/v${version}/${source.file}";
            sha256 = source.hash;
          };

        sourceRoot = ".";

        installPhase = ''
          runHook preInstall
          install -Dm755 helmify $out/bin/helmify
          runHook postInstall
        '';
      };
    in
    {
      devShell = pkgs.devshell.mkShell {
        imports = [ (pkgs.devshell.importTOML ./devshell.toml) ];
        packages = [ helmify ];
      };
    });
}
