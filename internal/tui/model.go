package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"mcfn/internal/generator"
	gh "mcfn/internal/github"
)

// Options are filled from CLI flags.
type Options struct {
	Hostname   string
	Username   string
	Output     string
	NoGitHub   bool
	Settings   *generator.Settings // loaded from --config
}

type step int

const (
	stepWelcome step = iota
	stepHostname
	stepUsername
	stepExtraUsers
	stepUseFlake
	stepBootloader
	stepArchitecture
	stepNixosVer
	stepDesktopEnv
	stepDisplayMgr
	stepFilesystem
	stepNetworkMgr
	stepSSH
	stepHomeManager
	stepSopsNix
	stepPackages
	stepOutputChoice
	stepOutputDir
	stepConfirm
	stepPreview
	stepGenerating
	stepGenDone
	stepAskGitHub
	stepGitHubToken
	stepGitHubRepo
	stepPushing
	stepFinished
	stepError
)

type doneMsg struct{ err error }
type pushDoneMsg struct {
	url string
	err error
}

type Model struct {
	opts step // suppress unused warning — use opts field below
	// suppress: re-declared below as Options
	options Options
	step    step

	hostname    string
	username    string
	extraUsers  string
	useFlake    bool
	bootloader  int // 0=systemd-boot, 1=grub-efi, 2=grub-legacy
	arch        int // 0=x86_64-linux, 1=aarch64-linux
	nixosVer    int // 0=25.05, 1=26.05, 2=unstable
	desktopEnv  int // 0=none..10=niri
	displayMgr  int // 0=none, 1=gdm, 2=sddm, 3=lightdm, 4=greetd, 5=ly
	filesystem  int // 0=ext4, 1=btrfs, 2=zfs
	zfsHostId   string
	netMgr      bool
	enableSSH   bool
	homeManager bool
	sopsNix     bool
	packages    string
	outputDir   string

	ghToken string
	ghRepo  string
	repoURL string

	input    textinput.Model
	spin     spinner.Model
	vp       viewport.Model
	cursor   int
	validErr string

	previewFiles    []string
	previewContents []string
	previewTab      int

	errMsg string
	width  int
	height int
}

var (
	archOptions    = []string{"x86_64-linux", "aarch64-linux"}
	nixosVerOpts   = []string{"25.05", "26.05", "unstable"}
	bootloaderOpts = []string{"systemd-boot", "grub-efi", "grub-legacy"}
	deOptions = []string{
		"none", "gnome", "kde", "xfce", "cinnamon",
		"mate", "lxqt", "i3", "sway", "hyprland", "niri",
	}
	deLabels = []string{
		"Нет (минимальная система)",
		"GNOME",
		"KDE Plasma 6",
		"XFCE",
		"Cinnamon",
		"MATE",
		"LXQt",
		"i3 (X11 WM)",
		"Sway (Wayland WM)",
		"Hyprland (Wayland WM)",
		"Niri (Wayland compositor)",
	}
	dmOptions = []string{"none", "gdm", "sddm", "lightdm", "greetd", "ly"}
	dmLabels  = []string{
		"Не использовать (TTY вход)",
		"GDM (GNOME Display Manager)",
		"SDDM (Qt, рекомендуется для KDE)",
		"LightDM (лёгкий, универсальный)",
		"greetd (Wayland-совместимый)",
		"ly (TUI дисплей-менеджер)",
	}

	outputPresets = []string{
		"./nixos-config",
		"/etc/nixos",
		"~/nixos-config",
		"Своя директория...",
	}

	fsOptions = []string{"ext4", "btrfs", "zfs"}
	fsLabels  = []string{
		"ext4 (стандартный, надёжный)",
		"btrfs (снапшоты, сжатие zstd)",
		"zfs (продвинутый, генерирует hostId)",
	}
)

func NewModel(opts Options) Model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 256

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	m := Model{
		options:   opts,
		step:      stepWelcome,
		outputDir: "./nixos-config",
		nixosVer:  1, // default: 26.05
		input:     ti,
		spin:      sp,
	}
	if opts.Output != "" {
		m.outputDir = opts.Output
	}
	if opts.Settings != nil {
		m = applySettings(m, *opts.Settings)
		m.step = stepConfirm
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.step == stepPreview {
			m.vp.Width = msg.Width
			m.vp.Height = vpHeight(msg.Height)
		}
		return m, nil

	case spinner.TickMsg:
		if m.step == stepGenerating || m.step == stepPushing {
			var cmd tea.Cmd
			m.spin, cmd = m.spin.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.step == stepPreview {
			return m.handlePreviewKey(msg)
		}
		return m.handleKey(msg)

	case doneMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.step = stepError
		} else {
			m.step = stepGenDone
		}
		return m, nil

	case pushDoneMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.step = stepError
		} else {
			m.repoURL = msg.url
			m.step = stepFinished
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) handlePreviewKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		m.previewTab = (m.previewTab + 1) % len(m.previewFiles)
		m.vp.SetContent(m.previewContents[m.previewTab])
		m.vp.GotoTop()
		return m, nil
	case "enter":
		m.step = stepGenerating
		return m, tea.Batch(m.spin.Tick, m.generateCmd())
	case "q":
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.step {
	case stepWelcome:
		if msg.String() == "enter" || msg.String() == " " {
			m.step = stepHostname
			m.input.Placeholder = "например: mynixos"
			m.input.SetValue(m.options.Hostname)
			m.input.Focus()
		}

	case stepHostname:
		if msg.String() == "enter" {
			val := strings.TrimSpace(m.input.Value())
			if e := validateHostname(val); e != "" {
				m.validErr = e
			} else {
				m.validErr = ""
				m.hostname = val
				m.step = stepUsername
				m.input.Placeholder = "например: alice"
				m.input.SetValue(m.options.Username)
				m.input.Focus()
			}
		} else {
			m.validErr = ""
		}

	case stepUsername:
		if msg.String() == "enter" {
			val := strings.TrimSpace(m.input.Value())
			if e := validateUsername(val); e != "" {
				m.validErr = e
			} else {
				m.validErr = ""
				m.username = val
				m.step = stepExtraUsers
				m.input.Placeholder = "bob carol (пробел/запятая; Enter — пропустить)"
				m.input.SetValue("")
				m.input.Focus()
			}
		} else {
			m.validErr = ""
		}

	case stepExtraUsers:
		if msg.String() == "enter" {
			m.extraUsers = m.input.Value()
			m.step = stepUseFlake
			m.cursor = 0
		}

	case stepUseFlake:
		m = choiceNav(m, msg, 1)
		if msg.String() == "enter" {
			m.useFlake = m.cursor == 0
			m.step = stepBootloader
			m.cursor = 0
		}

	case stepBootloader:
		m = choiceNav(m, msg, 2)
		if msg.String() == "enter" {
			m.bootloader = m.cursor
			m.step = stepArchitecture
			m.cursor = 0
		}

	case stepArchitecture:
		m = choiceNav(m, msg, 1)
		if msg.String() == "enter" {
			m.arch = m.cursor
			m.step = stepNixosVer
			m.cursor = m.nixosVer
		}

	case stepNixosVer:
		m = choiceNav(m, msg, 2)
		if msg.String() == "enter" {
			m.nixosVer = m.cursor
			m.step = stepDesktopEnv
			m.cursor = 0
		}

	case stepDesktopEnv:
		m = choiceNav(m, msg, len(deOptions)-1)
		if msg.String() == "enter" {
			m.desktopEnv = m.cursor
			// smart default DM based on DE
			switch deOptions[m.cursor] {
			case "gnome":
				m.displayMgr = 1 // gdm
			case "kde":
				m.displayMgr = 2 // sddm
			case "xfce", "cinnamon", "mate", "lxqt", "i3":
				m.displayMgr = 3 // lightdm
			case "sway", "hyprland", "niri":
				m.displayMgr = 4 // greetd
			default:
				m.displayMgr = 0 // none
			}
			m.step = stepDisplayMgr
			m.cursor = m.displayMgr
		}

	case stepDisplayMgr:
		m = choiceNav(m, msg, len(dmOptions)-1)
		if msg.String() == "enter" {
			m.displayMgr = m.cursor
			m.step = stepFilesystem
			m.cursor = 0
		}

	case stepFilesystem:
		m = choiceNav(m, msg, 2)
		if msg.String() == "enter" {
			m.filesystem = m.cursor
			if m.cursor == 2 && m.zfsHostId == "" {
				m.zfsHostId = generator.RandomHostID()
			}
			m.step = stepNetworkMgr
			m.cursor = 0
		}

	case stepNetworkMgr:
		m = choiceNav(m, msg, 1)
		if msg.String() == "enter" {
			m.netMgr = m.cursor == 0
			m.step = stepSSH
			m.cursor = 0
		}

	case stepSSH:
		m = choiceNav(m, msg, 1)
		if msg.String() == "enter" {
			m.enableSSH = m.cursor == 0
			if m.useFlake {
				m.step = stepHomeManager
			} else {
				m.step = stepPackages
				m.input.Placeholder = "git vim firefox (пробел/запятая; Enter — пропустить)"
				m.input.SetValue("")
				m.input.Focus()
			}
			m.cursor = 0
		}

	case stepHomeManager:
		m = choiceNav(m, msg, 1)
		if msg.String() == "enter" {
			m.homeManager = m.cursor == 0
			m.step = stepSopsNix
			m.cursor = 0
		}

	case stepSopsNix:
		m = choiceNav(m, msg, 1)
		if msg.String() == "enter" {
			m.sopsNix = m.cursor == 0
			m.step = stepPackages
			m.input.Placeholder = "git vim firefox (пробел/запятая; Enter — пропустить)"
			m.input.SetValue("")
			m.input.Focus()
		}

	case stepPackages:
		if msg.String() == "enter" {
			m.packages = m.input.Value()
			m.step = stepOutputChoice
			m.cursor = 0
		}

	case stepOutputChoice:
		m = choiceNav(m, msg, len(outputPresets)-1)
		if msg.String() == "enter" {
			if m.cursor == len(outputPresets)-1 {
				// "Своя директория..." → text input
				m.step = stepOutputDir
				m.input.Placeholder = "/path/to/config"
				m.input.SetValue(m.outputDir)
				m.input.Focus()
			} else {
				m.outputDir = outputPresets[m.cursor]
				m.step = stepConfirm
				m.cursor = 0
			}
		}

	case stepOutputDir:
		if msg.String() == "enter" {
			val := strings.TrimSpace(m.input.Value())
			if val == "" {
				val = "./nixos-config"
			}
			m.outputDir = val
			m.step = stepConfirm
			m.cursor = 0
		}

	case stepConfirm:
		m = choiceNav(m, msg, 1)
		if msg.String() == "enter" {
			if m.cursor == 1 {
				return m, tea.Quit
			}
			newM, err := m.buildPreview()
			if err != nil {
				m.errMsg = err.Error()
				m.step = stepError
				return m, nil
			}
			newM.step = stepPreview
			return newM, nil
		}

	case stepGenDone:
		if msg.String() == "enter" || msg.String() == " " {
			if m.options.NoGitHub {
				m.step = stepFinished
			} else {
				m.step = stepAskGitHub
				m.cursor = 0
			}
		}

	case stepAskGitHub:
		m = choiceNav(m, msg, 1)
		if msg.String() == "enter" {
			if m.cursor == 0 {
				m.step = stepGitHubToken
				m.input.Placeholder = "ghp_xxxxxxxxxxxx"
				m.input.EchoMode = textinput.EchoPassword
				m.input.SetValue("")
				m.input.Focus()
			} else {
				m.step = stepFinished
			}
		}

	case stepGitHubToken:
		if msg.String() == "enter" && strings.TrimSpace(m.input.Value()) != "" {
			m.ghToken = strings.TrimSpace(m.input.Value())
			m.step = stepGitHubRepo
			m.input.EchoMode = textinput.EchoNormal
			m.input.Placeholder = "nixos-config"
			m.input.SetValue("nixos-config")
			m.input.Focus()
		}

	case stepGitHubRepo:
		if msg.String() == "enter" && strings.TrimSpace(m.input.Value()) != "" {
			m.ghRepo = strings.TrimSpace(m.input.Value())
			m.step = stepPushing
			return m, tea.Batch(m.spin.Tick, m.pushCmd())
		}

	case stepFinished, stepError:
		if msg.String() == "enter" || msg.String() == "q" {
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// --- Helpers ---

func (m Model) buildPreview() (Model, error) {
	files, err := generator.Preview(m.buildConfig())
	if err != nil {
		return m, err
	}
	var names, contents []string
	for _, f := range files {
		names = append(names, f.Name)
		contents = append(contents, f.Content)
	}
	m.previewFiles = names
	m.previewContents = contents
	m.previewTab = 0

	w := m.width
	if w == 0 {
		w = 80
	}
	m.vp = viewport.New(w, vpHeight(m.height))
	m.vp.SetContent(contents[0])
	return m, nil
}

func (m Model) buildConfig() generator.Config {
	return generator.Config{
		Hostname:     m.hostname,
		Username:     m.username,
		ExtraUsers:   generator.ParsePackages(m.extraUsers),
		UseFlake:     m.useFlake,
		Bootloader:   bootloaderOpts[m.bootloader],
		Architecture: archOptions[m.arch],
		NixosVer:     nixosVerOpts[m.nixosVer],
		DesktopEnv:   deOptions[m.desktopEnv],
		DisplayMgr:   dmOptions[m.displayMgr],
		Filesystem:   fsOptions[m.filesystem],
		ZfsHostId:    m.zfsHostId,
		NetworkMgr:   m.netMgr,
		EnableSSH:    m.enableSSH,
		HomeManager:  m.homeManager,
		SopsNix:      m.sopsNix,
		Packages:     generator.ParsePackages(m.packages),
	}
}

func (m Model) generateCmd() tea.Cmd {
	cfg := m.buildConfig()
	outDir := expandHome(m.outputDir)
	return func() tea.Msg {
		if err := generator.Generate(cfg, outDir); err != nil {
			return doneMsg{err}
		}
		_ = generator.SaveSettings(cfg, outDir)
		return doneMsg{}
	}
}

func expandHome(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}

func (m Model) pushCmd() tea.Cmd {
	token, repo, dir := m.ghToken, m.ghRepo, m.outputDir
	return func() tea.Msg {
		url, err := gh.PushToRepo(token, repo, dir)
		return pushDoneMsg{url: url, err: err}
	}
}

func choiceNav(m Model, msg tea.KeyMsg, max int) Model {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < max {
			m.cursor++
		}
	}
	return m
}

func vpHeight(h int) int {
	if h <= 12 {
		return 10
	}
	return h - 12
}

func applySettings(m Model, s generator.Settings) Model {
	m.hostname = s.Hostname
	m.username = s.Username
	m.extraUsers = strings.Join(s.ExtraUsers, " ")
	m.useFlake = s.UseFlake
	m.bootloader = indexOf(bootloaderOpts, s.Bootloader)
	m.arch = indexOf(archOptions, s.Architecture)
	m.nixosVer = indexOf(nixosVerOpts, s.NixosVer)
	m.desktopEnv = indexOf(deOptions, s.DesktopEnv)
	m.displayMgr = indexOf(dmOptions, s.DisplayMgr)
	m.filesystem = indexOf(fsOptions, s.Filesystem)
	if s.Filesystem == "zfs" {
		m.zfsHostId = generator.RandomHostID()
	}
	m.netMgr = s.NetworkMgr
	m.enableSSH = s.EnableSSH
	m.homeManager = s.HomeManager
	m.sopsNix = s.SopsNix
	m.packages = strings.Join(s.Packages, " ")
	if s.OutputDir != "" {
		m.outputDir = s.OutputDir
	}
	return m
}

func indexOf(arr []string, val string) int {
	for i, v := range arr {
		if v == val {
			return i
		}
	}
	return 0
}

// --- Validation ---

func validateHostname(s string) string {
	if len(s) == 0 {
		return "Hostname не может быть пустым"
	}
	if len(s) > 63 {
		return "Hostname не может быть длиннее 63 символов"
	}
	if s[0] == '-' || s[len(s)-1] == '-' {
		return "Hostname не может начинаться или заканчиваться на дефис"
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-') {
			return "Hostname: только буквы, цифры и дефис"
		}
	}
	return ""
}

func validateUsername(s string) string {
	if len(s) == 0 {
		return "Username не может быть пустым"
	}
	if len(s) > 32 {
		return "Username не может быть длиннее 32 символов"
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			return "Username: только буквы, цифры, _ и -"
		}
	}
	return ""
}

// --- Views ---

func (m Model) View() string {
	switch m.step {
	case stepWelcome:
		return m.viewWelcome()
	case stepHostname:
		return m.viewInput("Hostname", "Имя вашей машины:", m.input.View())
	case stepUsername:
		return m.viewInput("Username", "Имя основного пользователя:", m.input.View())
	case stepExtraUsers:
		return m.viewInput("Доп. пользователи", "Дополнительные пользователи (Enter — пропустить):", m.input.View())
	case stepUseFlake:
		return m.viewChoice("Использовать flake.nix?",
			[]string{"Да — использовать flake.nix", "Нет — только configuration.nix"})
	case stepBootloader:
		return m.viewChoice("Загрузчик:",
			[]string{"systemd-boot (EFI — рекомендуется)", "GRUB EFI", "GRUB Legacy (BIOS)"})
	case stepArchitecture:
		return m.viewChoice("Архитектура процессора:",
			[]string{"x86_64-linux (Intel / AMD)", "aarch64-linux (ARM / Apple Silicon)"})
	case stepNixosVer:
		return m.viewChoice("Версия NixOS:",
			[]string{"25.05 (stable)", "26.05 (stable, latest)", "unstable"})
	case stepDesktopEnv:
		return m.viewChoice("Рабочий стол:", deLabels)
	case stepDisplayMgr:
		return m.viewChoice("Дисплей-менеджер:", dmLabels)
	case stepFilesystem:
		return m.viewChoice("Файловая система:", fsLabels)
	case stepNetworkMgr:
		return m.viewChoice("Включить NetworkManager?", []string{"Да", "Нет"})
	case stepSSH:
		return m.viewChoice("Включить SSH (openssh)?", []string{"Да", "Нет"})
	case stepHomeManager:
		return m.viewChoice("Включить Home Manager?",
			[]string{"Да — добавить home-manager модуль", "Нет"})
	case stepSopsNix:
		return m.viewChoice("Включить sops-nix (секреты)?",
			[]string{"Да — добавить sops-nix модуль", "Нет"})
	case stepPackages:
		return m.viewInput("Пакеты", "Дополнительные пакеты (Enter — пропустить):", m.input.View())
	case stepOutputChoice:
		return m.viewChoice("Куда сохранить конфиг?", outputPresets)
	case stepOutputDir:
		return m.viewInput("Своя директория", "Введите путь:", m.input.View())
	case stepConfirm:
		return m.viewConfirm()
	case stepPreview:
		return m.viewPreview()
	case stepGenerating:
		return m.viewSpinner("Генерирую файлы конфигурации...")
	case stepGenDone:
		return m.viewGenDone()
	case stepAskGitHub:
		return m.viewChoice("Загрузить конфиг в приватный репозиторий GitHub?",
			[]string{"Да — создать приватный репо", "Нет — оставить локально"})
	case stepGitHubToken:
		return m.viewInput("GitHub Token", "Введите GitHub Personal Access Token:", m.input.View())
	case stepGitHubRepo:
		return m.viewInput("Репозиторий", "Название нового репозитория:", m.input.View())
	case stepPushing:
		return m.viewSpinner("Создаю репозиторий и загружаю файлы...")
	case stepFinished:
		return m.viewFinished()
	case stepError:
		return m.viewError()
	}
	return ""
}

func header() string {
	return titleStyle.Render("╔══ MCFN — NixOS Config Generator ══╗") + "\n\n"
}

func (m Model) viewWelcome() string {
	return header() + boxStyle.Render(
		labelStyle.Render("Добро пожаловать!")+"\n\n"+
			"Этот инструмент поможет вам сгенерировать\n"+
			"configuration.nix (и опционально flake.nix)\n"+
			"для вашей системы NixOS.\n\n"+
			dimStyle.Render("Enter — начать  •  Ctrl+C — выход"),
	)
}

func (m Model) viewInput(label, prompt, inputView string) string {
	s := header()
	s += labelStyle.Render(label) + "\n" + prompt + "\n\n"
	s += activeInputStyle.Render(inputView) + "\n"
	if m.validErr != "" {
		s += "\n" + errorStyle.Render("✗ "+m.validErr) + "\n"
	} else {
		s += "\n"
	}
	s += "\n" + dimStyle.Render("Enter — продолжить")
	return s
}

func (m Model) viewChoice(prompt string, options []string) string {
	s := header() + labelStyle.Render(prompt) + "\n\n"
	for i, opt := range options {
		if i == m.cursor {
			s += selectedStyle.Render("▶  "+opt) + "\n"
		} else {
			s += unselectedStyle.Render("   "+opt) + "\n"
		}
	}
	s += "\n" + dimStyle.Render("↑/↓ — выбор  •  Enter — подтвердить")
	return s
}

func (m Model) viewConfirm() string {
	s := header() + labelStyle.Render("Подтвердите настройки:") + "\n\n"

	yn := func(b bool) string {
		if b {
			return "Да"
		}
		return "Нет"
	}
	extra := m.extraUsers
	if extra == "" {
		extra = "(нет)"
	}
	pkgs := m.packages
	if pkgs == "" {
		pkgs = "(нет)"
	}
	fsLabel := fsLabels[m.filesystem]
	if m.filesystem == 2 {
		fsLabel += " [hostId: " + m.zfsHostId + "]"
	}

	info := fmt.Sprintf(
		"  Hostname:        %s\n"+
			"  Основной user:   %s\n"+
			"  Доп. users:      %s\n"+
			"  flake.nix:       %s\n"+
			"  Загрузчик:       %s\n"+
			"  Архитектура:     %s\n"+
			"  Версия NixOS:    %s\n"+
			"  Рабочий стол:    %s\n"+
			"  Дисплей-менеджер: %s\n"+
			"  Файловая с-ма:   %s\n"+
			"  NetworkMgr:      %s\n"+
			"  SSH:             %s\n"+
			"  Home Manager:    %s\n"+
			"  sops-nix:        %s\n"+
			"  Пакеты:          %s\n"+
			"  Директория:      %s",
		m.hostname, m.username, extra,
		yn(m.useFlake),
		bootloaderOpts[m.bootloader],
		archOptions[m.arch],
		nixosVerOpts[m.nixosVer],
		deLabels[m.desktopEnv],
		dmLabels[m.displayMgr],
		fsLabel,
		yn(m.netMgr), yn(m.enableSSH),
		yn(m.homeManager), yn(m.sopsNix),
		pkgs, m.outputDir,
	)
	s += boxStyle.Render(info) + "\n\n"

	for i, opt := range []string{"Предпросмотр и генерация", "Отмена"} {
		if i == m.cursor {
			s += selectedStyle.Render("▶  "+opt) + "\n"
		} else {
			s += unselectedStyle.Render("   "+opt) + "\n"
		}
	}
	s += "\n" + dimStyle.Render("↑/↓ — выбор  •  Enter — подтвердить")
	return s
}

func (m Model) viewPreview() string {
	s := header()
	tabs := ""
	for i, name := range m.previewFiles {
		if i == m.previewTab {
			tabs += selectedStyle.Render(" ["+name+"] ")
		} else {
			tabs += unselectedStyle.Render(" ["+name+"] ")
		}
	}
	s += tabs + "\n\n"
	s += m.vp.View() + "\n\n"
	pct := fmt.Sprintf("%d%%", int(m.vp.ScrollPercent()*100))
	s += dimStyle.Render("Tab — след. файл  •  ↑/↓ — прокрутка  •  "+pct+"  •  Enter — сгенерировать  •  q — выход")
	return s
}

func (m Model) viewSpinner(msg string) string {
	return header() +
		m.spin.View() + " " + labelStyle.Render(msg) + "\n\n" +
		dimStyle.Render("Подождите...")
}

func (m Model) viewGenDone() string {
	s := header()
	s += successStyle.Render("✓ Конфигурация успешно сгенерирована!") + "\n\n"
	s += "Файлы сохранены в: " + labelStyle.Render(m.outputDir) + "\n"
	s += "  • configuration.nix\n"
	if m.useFlake {
		s += "  • flake.nix\n"
	}
	if m.homeManager {
		s += "  • home.nix\n"
	}
	if m.sopsNix {
		s += "  • secrets/secrets.yaml\n"
		s += "  • .sops.yaml\n"
	}
	s += "  • mcfn-config.json " + dimStyle.Render("(настройки для повторного использования)") + "\n\n"
	if m.useFlake {
		s += dimStyle.Render("Применить: sudo nixos-rebuild switch --flake " + m.outputDir + "#" + m.hostname)
	} else {
		s += dimStyle.Render("Применить: sudo nixos-rebuild switch -I nixos-config=" + m.outputDir + "/configuration.nix")
	}
	s += "\n\n" + dimStyle.Render("Нажмите Enter чтобы продолжить")
	return s
}

func (m Model) viewFinished() string {
	s := header() + successStyle.Render("✓ Всё готово!") + "\n\n"
	if m.repoURL != "" {
		s += "Репозиторий создан: " + labelStyle.Render(m.repoURL) + "\n\n"
	}
	s += dimStyle.Render("Нажмите Enter или q для выхода")
	return s
}

func (m Model) viewError() string {
	return header() +
		errorStyle.Render("✗ Ошибка:") + "\n\n" +
		m.errMsg + "\n\n" +
		dimStyle.Render("Нажмите Enter для выхода")
}
