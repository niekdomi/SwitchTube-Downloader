// Package styles provides shared lipgloss styles for terminal output.
package styles

import "github.com/charmbracelet/lipgloss"

// Terminal color palette (ANSI basic 16, adapts to terminal theme).
const (
	Red    = lipgloss.Color("1")
	Green  = lipgloss.Color("2")
	Yellow = lipgloss.Color("3")
	Cyan   = lipgloss.Color("6")
)

// Semantic badge styles.
var (
	Success = lipgloss.NewStyle().Foreground(Green).Bold(true)
	Error   = lipgloss.NewStyle().Foreground(Red).Bold(true)
	Warning = lipgloss.NewStyle().Foreground(Yellow).Bold(true)
	Info    = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
)
