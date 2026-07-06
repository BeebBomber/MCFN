{
  description = "MCFN System Tool";
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";
  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      packages.${system}.default = pkgs.buildGoModule {
        pname = "mcfn";
        version = "1.0.0";
        src = ./.;
        vendorHash = null;
      };
      devShells.${system}.default = pkgs.mkShell {
        buildInputs = with pkgs; [ go git nixos-rebuild ];
      };
    };
}
