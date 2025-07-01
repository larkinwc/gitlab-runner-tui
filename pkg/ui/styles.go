package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	ColorPrimary   = lipgloss.Color("#00b4d8")
	ColorSecondary = lipgloss.Color("#90e0ef")
	ColorSuccess   = lipgloss.Color("#52b788")
	ColorWarning   = lipgloss.Color("#f77f00")
	ColorError     = lipgloss.Color("#d62828")
	ColorMuted     = lipgloss.Color("#6c757d")
	ColorBg        = lipgloss.Color("#0a0e27")
	ColorFg        = lipgloss.Color("#ffffff")

	BaseStyle = lipgloss.NewStyle().
		Background(ColorBg).
		Foreground(ColorFg)

	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Background(ColorBg).
		Padding(0, 1)

	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorFg).
		Background(ColorPrimary).
		Padding(0, 1).
		Width(100)

	TabStyle = lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.RoundedBorder(), true, true, false, true).
		BorderForeground(ColorMuted)

	ActiveTabStyle = TabStyle.Copy().
		Bold(true).
		BorderForeground(ColorPrimary).
		Foreground(ColorPrimary)

	ListItemStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		PaddingRight(2)

	SelectedItemStyle = ListItemStyle.Copy().
		Foreground(ColorBg).
		Background(ColorPrimary)

	StatusActiveStyle = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true)

	StatusInactiveStyle = lipgloss.NewStyle().
		Foreground(ColorError).
		Bold(true)

	StatusUnknownStyle = lipgloss.NewStyle().
		Foreground(ColorMuted)

	InfoBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSecondary).
		Padding(1, 2).
		Margin(1, 0)

	ErrorBoxStyle = InfoBoxStyle.Copy().
		BorderForeground(ColorError)

	SuccessBoxStyle = InfoBoxStyle.Copy().
		BorderForeground(ColorSuccess)

	LogStyle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		PaddingLeft(1)

	HelpStyle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Padding(1, 0)

	InputStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSecondary).
		Padding(0, 1)

	FocusedInputStyle = InputStyle.Copy().
		BorderForeground(ColorPrimary)

	ButtonStyle = lipgloss.NewStyle().
		Foreground(ColorFg).
		Background(ColorSecondary).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSecondary)

	FocusedButtonStyle = ButtonStyle.Copy().
		Background(ColorPrimary).
		BorderForeground(ColorPrimary).
		Bold(true)

	SpinnerStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary)
)

func RenderStatus(status string) string {
	switch status {
	case "active", "running", "online":
		return StatusActiveStyle.Render("● " + status)
	case "inactive", "stopped", "offline":
		return StatusInactiveStyle.Render("● " + status)
	default:
		return StatusUnknownStyle.Render("● " + status)
	}
}

func TruncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}