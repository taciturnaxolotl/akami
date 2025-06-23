package styles

import "github.com/charmbracelet/lipgloss/v2"

var Fancy = lipgloss.NewStyle().Foreground(lipgloss.Magenta).Bold(true).Italic(true)
var Muted = lipgloss.NewStyle().Foreground(lipgloss.BrightBlue).Italic(true)
var Bad = lipgloss.NewStyle().Foreground(lipgloss.BrightRed).Bold(true)
var Success = lipgloss.NewStyle().Foreground(lipgloss.Green).Bold(true)
var Warn = lipgloss.NewStyle().Foreground(lipgloss.Yellow).Bold(true)
