package generator

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Config struct {
	Hostname       string
	Username       string
	ExtraUsers     []string
	UseFlake       bool
	Bootloader     string
	Architecture   string
	NixosVer       string
	NixpkgsURL     string
	HomeManagerURL string
	DesktopEnv     string // "", "gnome", "kde", "xfce", ...
	DisplayMgr     string // "", "gdm", "sddm", "lightdm", "greetd", "ly"
	Filesystem     string // "ext4", "btrfs", "zfs"
	ZfsHostId      string
	NetworkMgr     bool
	EnableSSH      bool
	HomeManager    bool
	SopsNix        bool
	Packages       []string
}

// Settings is the JSON-serialisable snapshot of a generation session.
type Settings struct {
	Hostname     string   `json:"hostname"`
	Username     string   `json:"username"`
	ExtraUsers   []string `json:"extra_users,omitempty"`
	UseFlake     bool     `json:"use_flake"`
	Bootloader   string   `json:"bootloader"`
	Architecture string   `json:"architecture"`
	NixosVer     string   `json:"nixos_ver"`
	DesktopEnv   string   `json:"desktop_env"`
	DisplayMgr   string   `json:"display_mgr"`
	Filesystem   string   `json:"filesystem"`
	NetworkMgr   bool     `json:"network_mgr"`
	EnableSSH    bool     `json:"enable_ssh"`
	HomeManager  bool     `json:"home_manager"`
	SopsNix      bool     `json:"sops_nix"`
	Packages     []string `json:"packages,omitempty"`
	OutputDir    string   `json:"output_dir"`
}

type PreviewFile struct {
	Name    string
	Content string
}

// --- Templates ---

const configTmpl = `{ config, pkgs, ... }: {
  imports = [ ./hardware-configuration.nix ];

  networking.hostName = "{{ .Hostname }}";
{{ if .NetworkMgr }}
  networking.networkmanager.enable = true;
{{ end }}
{{ if eq .Bootloader "systemd-boot" }}
  boot.loader.systemd-boot.enable = true;
  boot.loader.efi.canTouchEfiVariables = true;
{{ else if eq .Bootloader "grub-efi" }}
  boot.loader.grub.enable = true;
  boot.loader.grub.efiSupport = true;
  boot.loader.grub.device = "nodev";
  boot.loader.efi.canTouchEfiVariables = true;
{{ else }}
  boot.loader.grub.enable = true;
  boot.loader.grub.device = "/dev/sda";
{{ end }}
{{ if eq .Filesystem "btrfs" }}
  services.btrfs.autoScrub = {
    enable = true;
    interval = "monthly";
  };
{{ else if eq .Filesystem "zfs" }}
  boot.supportedFilesystems = [ "zfs" ];
  networking.hostId = "{{ .ZfsHostId }}";
  services.zfs.autoScrub.enable = true;
{{ end }}
{{ if eq .DisplayMgr "gdm" }}
  services.displayManager.gdm.enable = true;
{{ else if eq .DisplayMgr "sddm" }}
  services.displayManager.sddm.enable = true;
{{ else if eq .DisplayMgr "lightdm" }}
  services.displayManager.lightdm.enable = true;
{{ else if eq .DisplayMgr "greetd" }}
  services.greetd.enable = true;
{{ else if eq .DisplayMgr "ly" }}
  services.displayManager.ly.enable = true;
{{ end }}
{{ if eq .DesktopEnv "gnome" }}
  services.xserver.enable = true;
  services.xserver.desktopManager.gnome.enable = true;
{{ else if eq .DesktopEnv "kde" }}
  services.xserver.enable = true;
  services.desktopManager.plasma6.enable = true;
{{ else if eq .DesktopEnv "xfce" }}
  services.xserver.enable = true;
  services.xserver.desktopManager.xfce.enable = true;
{{ else if eq .DesktopEnv "cinnamon" }}
  services.xserver.enable = true;
  services.xserver.desktopManager.cinnamon.enable = true;
{{ else if eq .DesktopEnv "mate" }}
  services.xserver.enable = true;
  services.xserver.desktopManager.mate.enable = true;
{{ else if eq .DesktopEnv "lxqt" }}
  services.lxqt.enable = true;
{{ else if eq .DesktopEnv "i3" }}
  services.xserver.enable = true;
  services.xserver.windowManager.i3.enable = true;
  services.displayManager.defaultSession = "none+i3";
{{ else if eq .DesktopEnv "sway" }}
  programs.sway.enable = true;
{{ else if eq .DesktopEnv "hyprland" }}
  programs.hyprland.enable = true;
{{ else if eq .DesktopEnv "niri" }}
  programs.niri.enable = true;
{{ end }}
  users.users.{{ .Username }} = {
    isNormalUser = true;
    extraGroups = [ "wheel"{{ if .NetworkMgr }} "networkmanager"{{ end }} ];
  };
{{ range .ExtraUsers }}
  users.users.{{ . }} = {
    isNormalUser = true;
    extraGroups = [{{ if $.NetworkMgr }} "networkmanager"{{ end }} ];
  };
{{ end }}
{{ if .SopsNix }}
  sops.defaultSopsFile = ./secrets/secrets.yaml;
  sops.age.sshKeyPaths = [ "/etc/ssh/ssh_host_ed25519_key" ];
{{ end }}
{{ if .HomeManager }}
  home-manager.useGlobalPkgs = true;
  home-manager.useUserPackages = true;
  home-manager.users.{{ .Username }} = import ./home.nix;
{{ end }}
{{ if .Packages }}
  environment.systemPackages = with pkgs; [
{{ range .Packages }}    {{ . }}
{{ end }}  ];
{{ end }}
  system.stateVersion = "{{ .NixosVer }}";
}
`

const flakeTmpl = `{
  description = "NixOS configuration for {{ .Hostname }}";

  inputs = {
    nixpkgs.url = "{{ .NixpkgsURL }}";
{{ if .HomeManager }}
    home-manager = {
      url = "{{ .HomeManagerURL }}";
      inputs.nixpkgs.follows = "nixpkgs";
    };
{{ end }}
{{ if .SopsNix }}
    sops-nix = {
      url = "github:Mic92/sops-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
{{ end }}
  };

  outputs = { self, nixpkgs{{ if .HomeManager }}, home-manager{{ end }}{{ if .SopsNix }}, sops-nix{{ end }}, ... }: {
    nixosConfigurations.{{ .Hostname }} = nixpkgs.lib.nixosSystem {
      system = "{{ .Architecture }}";
      modules = [
        ./configuration.nix
{{ if .HomeManager }}
        home-manager.nixosModules.home-manager
{{ end }}
{{ if .SopsNix }}
        sops-nix.nixosModules.sops
{{ end }}
      ];
    };
  };
}
`

const homeTmpl = `{ config, pkgs, ... }: {
  home.username = "{{ .Username }}";
  home.homeDirectory = "/home/{{ .Username }}";
  home.stateVersion = "{{ .NixosVer }}";

  programs.home-manager.enable = true;
}
`

const secretsYaml = `# sops-nix secrets placeholder
#
# Зашифруйте этот файл перед коммитом:
#   sops --encrypt --in-place secrets/secrets.yaml
#
# Пример:
# myPassword: super_secret_value
# dbUrl: postgres://user:pass@localhost/db
`

const sopsYaml = `# sops конфигурация
#
# Сгенерируйте age-ключ: age-keygen -o ~/.config/sops/age/keys.txt
# Замените REPLACE_WITH_YOUR_AGE_PUBLIC_KEY на публичный ключ из вывода команды.
keys:
  - &admin age1REPLACE_WITH_YOUR_AGE_PUBLIC_KEY
creation_rules:
  - path_regex: secrets/.*\.yaml$
    key_groups:
      - age:
          - *admin
`

// --- Public API ---

func Generate(cfg Config, outDir string) error {
	cfg.NixpkgsURL = nixpkgsURL(cfg.NixosVer)
	cfg.HomeManagerURL = homeManagerURL(cfg.NixosVer)
	if cfg.Filesystem == "zfs" && cfg.ZfsHostId == "" {
		cfg.ZfsHostId = RandomHostID()
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию: %w", err)
	}
	if err := writeTemplate(filepath.Join(outDir, "configuration.nix"), configTmpl, cfg); err != nil {
		return fmt.Errorf("configuration.nix: %w", err)
	}
	if cfg.UseFlake {
		if err := writeTemplate(filepath.Join(outDir, "flake.nix"), flakeTmpl, cfg); err != nil {
			return fmt.Errorf("flake.nix: %w", err)
		}
	}
	if cfg.HomeManager {
		if err := writeTemplate(filepath.Join(outDir, "home.nix"), homeTmpl, cfg); err != nil {
			return fmt.Errorf("home.nix: %w", err)
		}
	}
	if cfg.SopsNix {
		secretsDir := filepath.Join(outDir, "secrets")
		if err := os.MkdirAll(secretsDir, 0700); err != nil {
			return fmt.Errorf("secrets/: %w", err)
		}
		if err := os.WriteFile(filepath.Join(secretsDir, "secrets.yaml"), []byte(secretsYaml), 0600); err != nil {
			return fmt.Errorf("secrets/secrets.yaml: %w", err)
		}
		if err := os.WriteFile(filepath.Join(outDir, ".sops.yaml"), []byte(sopsYaml), 0644); err != nil {
			return fmt.Errorf(".sops.yaml: %w", err)
		}
	}
	return nil
}

func Preview(cfg Config) ([]PreviewFile, error) {
	cfg.NixpkgsURL = nixpkgsURL(cfg.NixosVer)
	cfg.HomeManagerURL = homeManagerURL(cfg.NixosVer)
	if cfg.Filesystem == "zfs" && cfg.ZfsHostId == "" {
		cfg.ZfsHostId = "<auto-generated>"
	}

	var files []PreviewFile

	content, err := renderToString(configTmpl, cfg)
	if err != nil {
		return nil, fmt.Errorf("configuration.nix: %w", err)
	}
	files = append(files, PreviewFile{"configuration.nix", content})

	if cfg.UseFlake {
		content, err = renderToString(flakeTmpl, cfg)
		if err != nil {
			return nil, fmt.Errorf("flake.nix: %w", err)
		}
		files = append(files, PreviewFile{"flake.nix", content})
	}
	if cfg.HomeManager {
		content, err = renderToString(homeTmpl, cfg)
		if err != nil {
			return nil, fmt.Errorf("home.nix: %w", err)
		}
		files = append(files, PreviewFile{"home.nix", content})
	}
	if cfg.SopsNix {
		files = append(files, PreviewFile{"secrets/secrets.yaml", secretsYaml})
		files = append(files, PreviewFile{".sops.yaml", sopsYaml})
	}
	return files, nil
}

func SaveSettings(cfg Config, outDir string) error {
	s := Settings{
		Hostname:     cfg.Hostname,
		Username:     cfg.Username,
		ExtraUsers:   cfg.ExtraUsers,
		UseFlake:     cfg.UseFlake,
		Bootloader:   cfg.Bootloader,
		Architecture: cfg.Architecture,
		NixosVer:     cfg.NixosVer,
		DesktopEnv:   cfg.DesktopEnv,
		DisplayMgr:   cfg.DisplayMgr,
		Filesystem:   cfg.Filesystem,
		NetworkMgr:   cfg.NetworkMgr,
		EnableSSH:    cfg.EnableSSH,
		HomeManager:  cfg.HomeManager,
		SopsNix:      cfg.SopsNix,
		Packages:     cfg.Packages,
		OutputDir:    outDir,
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "mcfn-config.json"), data, 0644)
}

func LoadSettings(path string) (Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Settings{}, err
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return Settings{}, err
	}
	return s, nil
}

func RandomHostID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// --- Helpers ---

func ParsePackages(raw string) []string {
	var pkgs []string
	for _, p := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' '
	}) {
		p = strings.TrimSpace(p)
		if p != "" {
			pkgs = append(pkgs, p)
		}
	}
	return pkgs
}

func writeTemplate(path, tmplStr string, data Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return err
	}
	return tmpl.Execute(f, data)
}

func renderToString(tmplStr string, data Config) (string, error) {
	var buf bytes.Buffer
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func nixpkgsURL(ver string) string {
	return "github:nixos/nixpkgs/nixos-" + ver
}

func homeManagerURL(ver string) string {
	if ver == "unstable" {
		return "github:nix-community/home-manager"
	}
	return "github:nix-community/home-manager/release-" + ver
}
