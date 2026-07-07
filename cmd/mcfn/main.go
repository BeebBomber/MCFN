package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"mcfn/internal/generator"
	"mcfn/internal/tui"
)

const helpText = `MCFN — NixOS Config Generator

Использование:
  mcfn [флаги]

Флаги:
  -h, --help              Показать эту справку
      --hostname <name>   Предзаполнить hostname
      --username <name>   Предзаполнить имя пользователя
      --output <dir>      Директория для сохранения конфига (default: ./nixos-config)
      --no-github         Пропустить шаг загрузки в GitHub
      --config <file>     Загрузить настройки из mcfn-config.json

Примеры:
  mcfn
  mcfn --hostname mypc --username alice
  mcfn --output /etc/nixos --no-github
  mcfn --config ./nixos-config/mcfn-config.json
  mcfn --config old-config.json --output ./new-config
`

func main() {
	var opts tui.Options
	var configFile string
	help := false

	flag.BoolVar(&help, "help", false, "")
	flag.BoolVar(&help, "h", false, "")
	flag.StringVar(&opts.Hostname, "hostname", "", "")
	flag.StringVar(&opts.Username, "username", "", "")
	flag.StringVar(&opts.Output, "output", "", "")
	flag.BoolVar(&opts.NoGitHub, "no-github", false, "")
	flag.StringVar(&configFile, "config", "", "")
	flag.Usage = func() { fmt.Print(helpText) }
	flag.Parse()

	if help {
		fmt.Print(helpText)
		os.Exit(0)
	}

	if configFile != "" {
		s, err := generator.LoadSettings(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка загрузки конфига: %v\n", err)
			os.Exit(1)
		}
		opts.Settings = &s
		// CLI --output overrides saved outputDir
		if opts.Output != "" {
			s.OutputDir = opts.Output
			opts.Settings = &s
		}
	}

	p := tea.NewProgram(tui.NewModel(opts), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка запуска: %v\n", err)
		os.Exit(1)
	}
}
