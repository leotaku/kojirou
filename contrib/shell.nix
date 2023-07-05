{ pkgs ? import <nixpkgs> { } }:

with pkgs;
let python = python3.withPackages (p: with p; [ lxml httpx fs pydantic ]);
in stdenvNoCC.mkDerivation {
  name = "dev-shell";
  buildInputs = [ python pyright ];
}
