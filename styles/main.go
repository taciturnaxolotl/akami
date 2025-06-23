package styles

import "github.com/charmbracelet/lipgloss/v2"

var Fancy = lipgloss.NewStyle().Foreground(lipgloss.Magenta).Bold(true).Italic(true)
var Muted = lipgloss.NewStyle().Foreground(lipgloss.BrightBlue).Italic(true)
var Bad = lipgloss.NewStyle().Foreground(lipgloss.BrightRed).Bold(true)
