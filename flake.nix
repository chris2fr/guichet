{
  description = "A simple LDAP web interface for Bottin";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/0244e143dc943bcf661fdaf581f01eb0f5000fcf";
  inputs.gomod2nix.url = "github:tweag/gomod2nix/40d32f82fc60d66402eb0972e6e368aeab3faf58";

  outputs = { self, nixpkgs, gomod2nix }:
  let
    pkgs = import nixpkgs {
      system = "x86_64-linux";
      overlays = [
        (self: super: {
          gomod = super.callPackage "${gomod2nix}/builder/" { };
        })
      ];
    };
    src = ./.;
    bottin = pkgs.gomod.buildGoApplication {
      pname = "guichet";
      version = "0.1.0";
      src = src;
      modules = ./gomod2nix.toml;

      CGO_ENABLED=0;

      ldflags = [
        "-X main.templatePath=${src + "/templates"}"
        "-X main.staticPath=${src + "/static"}"
      ];

      meta = with pkgs.lib; {
        description = "A simple LDAP web interface for Bottin";
        homepage = "https://git.deuxfleurs.fr/Deuxfleurs/guichet";
        license = licenses.gpl3Plus;
        platforms = platforms.linux;
      };
    };
  in
  {
    packages.x86_64-linux.bottin = bottin;
    packages.x86_64-linux.default = self.packages.x86_64-linux.bottin;
  };
}
