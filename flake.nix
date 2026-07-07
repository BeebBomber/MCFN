{
  description = "MCFN — NixOS Config Generator";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forAll = f: nixpkgs.lib.genAttrs systems f;
    in {
      packages = forAll (system: {
        default = nixpkgs.legacyPackages.${system}.buildGoModule {
          pname = "mcfn";
          version = "0.2.0";
          src = ./.;
          vendorHash = null;
          meta.mainProgram = "mcfn";
        };
      });

      apps = forAll (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/mcfn";
        };
      });

      devShells = forAll (system: {
        default = nixpkgs.legacyPackages.${system}.mkShell {
          buildInputs = [ nixpkgs.legacyPackages.${system}.go ];
        };
      });
    };
}
