package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary   = lipgloss.Color("205") // Purple
	colorSecondary = lipgloss.Color("99")  // Pink
	colorBorder    = lipgloss.Color("240") // Gray
	colorSelected  = lipgloss.Color("25")  // Green
	colorError     = lipgloss.Color("9")   // Red
	colorHeader    = lipgloss.Color("231") // White

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

	// Add styles for stats text
	statsLabelStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true)

	statsValueStyle = lipgloss.NewStyle().
			Foreground(colorPrimary)
)

// UI Styles for loading state
var (
	loadingTitleStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true).
				MarginBottom(1).
				Align(lipgloss.Center)

	loadingBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorSecondary).
			Padding(1, 2).
			Width(78)

	loadingPathStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true)

	scanningLabelStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)

	scanningCountStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true)

	scanningStatusStyle = lipgloss.NewStyle().
				Align(lipgloss.Center).
				MarginTop(1).
				MarginBottom(1)

	loadingFooterStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Align(lipgloss.Center).
				MarginTop(1)
)

var (
	// ... existing colors ...
	colorDeleting = lipgloss.Color("#FFA500") // Orange
	colorDeleted  = lipgloss.Color("#00FF00") // Green
	colorFailed   = lipgloss.Color("#FF0000") // Red
)
