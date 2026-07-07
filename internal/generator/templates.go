package generator

import (
	"os"
	"text/template"
)

type ConfigData struct {
	Hostname string
	Username string
}

const FlakeTemplate = `{
  description = "MCFN Generated System";
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";
  outputs = { self, nixpkgs, ... }@inputs: {
    nixosConfigurations.{{.Hostname}} = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [ ./configuration.nix ];
    };
  };
}`

const ConfigTemplate = `{ config, pkgs, ... }: {
  imports = [ ./hardware-configuration.nix ];
  networking.hostName = "{{.Hostname}}";
  boot.loader.systemd-boot.enable = true;
  boot.loader.efi.canTouchEfiVariables = true;
  users.users.{{.Username}} = {
    isNormalUser = true;
    extraGroups = [ "wheel" "networkmanager" ];
  };
  system.stateVersion = "26.05";
}`

// Функция, которая реально создает файлы
func CreateConfigs(hostname, username string) error {
	data := ConfigData{Hostname: hostname, Username: username}

	// Создаем flake.nix
	f, _ := os.Create("flake.nix")
	tmplFlake, _ := template.New("flake").Parse(FlakeTemplate)
	tmplFlake.Execute(f, data)
	f.Close()

	// Создаем configuration.nix
	c, _ := os.Create("configuration.nix")
	tmplConfig, _ := template.New("config").Parse(ConfigTemplate)
	tmplConfig.Execute(c, data)
	c.Close()

	return nil
}
