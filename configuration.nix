{ config, pkgs, ... }: {
  imports = [ ./hardware-configuration.nix ];
  networking.hostName = "bomb";
  boot.loader.systemd-boot.enable = true;
  boot.loader.efi.canTouchEfiVariables = true;
  users.users.Bomber = {
    isNormalUser = true;
    extraGroups = [ "wheel" "networkmanager" ];
  };
  system.stateVersion = "26.05";
}