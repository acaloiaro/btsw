{
  self,
  pkgs ? import <nixpkgs> {},
  ...
}:
pkgs.buildGoModule rec {
  pname = "btsw";
  version = pkgs.lib.strings.removeSuffix "\n" (builtins.readFile ./version.txt);
  src = ./.;
  vendorHash = null;
  ldflags = [
    "-X 'main.version=${version}-nix'"
    "-X 'main.commit=${self.rev or "dev"}'"
  ];
  env.CGO_ENABLED = 0;

  meta = {
    description = "A tiny cli for Linux that quickly connects and disconnects paired bluetooth devices";
    homepage = "https://github.com/acaloiaro/btsw";
    license = pkgs.lib.licenses.bsd2;
  };
}
