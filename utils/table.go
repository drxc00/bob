package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func FormatPath(p string, r string) string {
	relPath, err := filepath.Rel(r, p)
	if err != nil {
		return p
	}

	// Split into parts
	parts := strings.Split(relPath, string(os.PathSeparator))
	if len(parts) < 2 {
		return relPath
	}

	// Get project name (parent of node_modules)
	projectName := parts[len(parts)-2]
	// pathPart := filepath.Dir(relPath)

	// mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	return projectName
}
