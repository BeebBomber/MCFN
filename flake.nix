{
  description = "MCFN: My Configuration for NixOS Tool";
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
        vendorHash = null; # Замените на реальный хэш после go mod tidy

        postInstall = ''
          install -Dm644 man/en/mcfn.1 $out/share/man/man1/mcfn.1
          install -Dm644 man/ru/mcfn.1 $out/share/man/ru/man1/mcfn.1
        '';
      };

      devShells.${system}.default = pkgs.mkShell {
        buildInputs = with pkgs; [ go git nixos-rebuild ];
      };
    };
}
