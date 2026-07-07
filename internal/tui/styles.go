package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7dcfff"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e0af68")).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9ece6a")).
			Bold(true)

	unselectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565f89"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f7768e"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9ece6a"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565f89"))

	boxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3b4261")).
			Padding(1, 3)

	activeInputStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7aa2f7")).
				Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#bb9af7")).
			MarginBottom(1)
)
