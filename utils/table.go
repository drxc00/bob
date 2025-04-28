package utils

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func ColorCodedStaleness(staleness int64) string {
	var style lipgloss.Style

	// Apply different colors based on the staleness value
	if staleness > 365 {
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Red
	} else if staleness > 180 {
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
	} else if staleness > 90 {
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("14")) // Cyan
	} else {
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Green
	}

	return style.Render(fmt.Sprintf("%d days", staleness))
}
