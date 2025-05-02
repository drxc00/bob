package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary   = lipgloss.Color("205") // Purple
	colorSecondary = lipgloss.Color("99")  // Pink
	colorBorder    = lipgloss.Color("240") // Gray
	colorSelected  = lipgloss.Color("25")  // Green
	colorError     = lipgloss.Color("9")   // Red
	colorHeader    = lipgloss.Color("231") // White

	baseStyle  = lipgloss.NewStyle().Foreground(colorPrimary)
	titleStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true).
			Align(lipgloss.Center).
			Padding(0, 2)

	statsStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Align(lipgloss.Left)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Padding(1, 2)

	loadingBoxStyle = lipgloss.NewStyle().
			Height(12).
			Width(80).
			MaxHeight(12).
			MaxWidth(80).
			Align(lipgloss.Left)

	// Add styles for stats text
	statsLabelStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true)

	statsValueStyle = lipgloss.NewStyle().
			Foreground(colorPrimary)
)
