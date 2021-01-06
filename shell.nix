{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [ go gopls ];
  CGO_ENABLED = "0";
}
