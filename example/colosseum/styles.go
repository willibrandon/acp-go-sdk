package main

import "github.com/charmbracelet/lipgloss"

// Colors - per DESIGN.md specification
var (
	ClaudeColor = lipgloss.Color("#A855F7") // Purple
	GeminiColor = lipgloss.Color("#3B82F6") // Blue
	UserColor   = lipgloss.Color("#22C55E") // Green
	SystemColor = lipgloss.Color("#6B7280") // Gray
	BorderColor = lipgloss.Color("#374151")
	ErrorColor  = lipgloss.Color("#EF4444") // Red
)

// noColor disables all styling when true
var noColor bool

// SetNoColor enables or disables color output
func SetNoColor(disabled bool) {
	noColor = disabled
}

// Role name styles
var (
	ClaudeNameStyle = lipgloss.NewStyle().
			Foreground(ClaudeColor).
			Bold(true)

	GeminiNameStyle = lipgloss.NewStyle().
			Foreground(GeminiColor).
			Bold(true)

	UserNameStyle = lipgloss.NewStyle().
			Foreground(UserColor).
			Bold(true)

	SystemNameStyle = lipgloss.NewStyle().
			Foreground(SystemColor).
			Bold(true)
)

// Status indicators
var (
	ConnectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22C55E"))

	DisconnectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444"))
)

// Layout styles
var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(BorderColor)

	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(BorderColor).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(SystemColor)
)

// RenderStatus returns a styled connection indicator
func RenderStatus(connected bool) string {
	if noColor {
		if connected {
			return "[+]"
		}
		return "[-]"
	}
	if connected {
		return ConnectedStyle.Render("●")
	}
	return DisconnectedStyle.Render("○")
}

// RenderRoleName returns a styled role label
func RenderRoleName(role Role) string {
	if noColor {
		return string(role) + ":"
	}
	switch role {
	case RoleUser:
		return UserNameStyle.Render("You:")
	case RoleClaude:
		return ClaudeNameStyle.Render("Claude:")
	case RoleGemini:
		return GeminiNameStyle.Render("Gemini:")
	case RoleSystem:
		return SystemNameStyle.Render("System:")
	default:
		return string(role) + ":"
	}
}
