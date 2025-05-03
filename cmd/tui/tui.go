package tui

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/drxc00/sweepy/types"
	"github.com/drxc00/sweepy/utils"
)

// --- Model ---

type model struct {
	spinner      spinner.Model
	table        table.Model
	isLoading    bool
	scanComplete bool
	modules      []types.ScannedNodeModule

	// Config
	ctx types.ScanContext

	// Verbose
	progressChan  chan string
	scanningPaths []string
	lastUpdated   time.Time

	err           error
	width, height int
	totalSize     int64
	avgStaleness  float64
	scanDuration  string

	// deleted
	deletedPaths []string
	beingDeleted []string
}

// --- Init Functions ---

func initialModel(ctx types.ScanContext) model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	return model{
		spinner:      s,
		isLoading:    true,
		scanComplete: false,
		ctx:          ctx,
		progressChan: make(chan string, 1000),
		lastUpdated:  time.Now(),
	}
}

type scanResultMsg struct {
	modules []types.ScannedNodeModule
	stats   types.ScanInfo
	err     error
}

type scanProgressMsg struct {
	path string
}

type deleteSuccessMsg struct {
	path  string
	index int
	size  int64
}

type deleteErrMsg struct {
	err   error
	path  string
	index int
}

// --- BubbleTea Handlers ---

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		StartScan(m.ctx, m.progressChan),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up":
			m.table.MoveUp(1)
		case "down":
			m.table.MoveDown(1)
		case " ":
			if m.scanComplete && len(m.table.Rows()) > 0 {
				selectedIndex := m.table.Cursor()
				if selectedIndex >= 0 && selectedIndex < len(m.table.Rows()) {
					currentRows := m.table.Rows()
					selectedPath := currentRows[selectedIndex][1] // Assuming the second column contains the paths
					if len(currentRows[selectedIndex]) > 0 && selectedPath != "" && !slices.Contains(m.deletedPaths, selectedPath) && !slices.Contains(m.beingDeleted, selectedPath) {
						// Add to the list of paths being deleted
						m.beingDeleted = append(m.beingDeleted, selectedPath)

						selectedModule := m.modules[selectedIndex]

						// Immediately update the UI to show "Deleting..."
						updatedRow := make(table.Row, len(currentRows[selectedIndex]))
						copy(updatedRow, currentRows[selectedIndex])
						updatedRow[0] = "[DELETING...] " + strings.Replace(updatedRow[0], "[FAILED] ", "", 1) // Modify the first column

						newRows := make([]table.Row, len(currentRows))
						copy(newRows, currentRows)
						newRows[selectedIndex] = updatedRow
						m.table.SetRows(newRows)

						// Perform the deletion as a command to avoid blocking the UI
						cmd := func() tea.Msg {
							defer func() {
								// Safely remove from beingDeleted slice only if the path exists
								idx := slices.Index(m.beingDeleted, selectedPath)
								if idx >= 0 {
									m.beingDeleted = slices.Delete(m.beingDeleted, idx, idx+1)
								}
							}()
							return DeleteNode(selectedModule, selectedIndex)
						}
						return m, cmd
					}
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.scanComplete {
			// Allow the table to use most of the available width
			m.table.SetWidth(m.width - 4)
			m.table.SetHeight(m.height - 12) // Give more space for stats and footer
		}
	case scanResultMsg:
		m.isLoading = false
		m.scanComplete = true
		m.modules = msg.modules
		m.err = msg.err
		m.totalSize = msg.stats.TotalSize
		m.avgStaleness = msg.stats.AvgStaleness
		m.scanDuration = msg.stats.ScanDuration.String()

		if m.err != nil {
			utils.Log("Error scanning: %v\n", m.err)
			return m, tea.Quit
		}

		// Calculate proportional column widths based on content
		availableWidth := m.width - 8 // Allow for some padding and borders

		// Define column ratios (proportions of total width)
		projectRatio := 0.15
		pathRatio := 0.40
		sizeRatio := 0.10
		modifiedRatio := 0.15
		stalenessRatio := 0.15

		// Apply ratios to calculate actual column widths
		projectWidth := int(float64(availableWidth) * projectRatio)
		pathWidth := int(float64(availableWidth) * pathRatio)
		sizeWidth := int(float64(availableWidth) * sizeRatio)
		modifiedWidth := int(float64(availableWidth) * modifiedRatio)
		stalenessWidth := int(float64(availableWidth) * stalenessRatio)

		columns := []table.Column{
			{Title: "PROJECT", Width: projectWidth},
			{Title: "PATH", Width: pathWidth},
			{Title: "SIZE", Width: sizeWidth},
			{Title: "LAST MODIFIED", Width: modifiedWidth},
			{Title: "STALENESS", Width: stalenessWidth},
		}

		var rows []table.Row
		for _, module := range m.modules {
			rows = append(rows, table.Row{
				utils.FormatPath(module.Path, m.ctx.Path),
				module.Path,
				utils.FormatSize(module.Size),
				module.LastModified.Format("2006-01-02 15:04:05"),
				utils.ColorCodedStaleness(module.Staleness),
			})
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(m.height-20),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorSecondary).
			BorderBottom(true).
			Bold(true).
			Foreground(colorHeader).
			Background(colorBorder).
			Padding(0, 1)

		s.Selected = s.Selected.
			Bold(true).
			Background(colorSelected).
			Foreground(colorPrimary)

		t.SetStyles(s)

		// Adjust table dimensions to account for borders and padding
		t.SetHeight(m.height - 14)
		t.SetWidth(m.width - 6)

		m.table = t
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case scanProgressMsg:
		m.scanningPaths = append(m.scanningPaths, msg.path)
		m.lastUpdated = time.Now()
		return m, ListenForProgress(m.progressChan)
	case deleteSuccessMsg:
		// Update the UI to show "DELETED" and remove the module from the model
		if msg.index >= 0 && msg.index < len(m.table.Rows()) {
			currentRows := m.table.Rows()
			if len(currentRows[msg.index]) > 0 {
				updatedRow := make(table.Row, len(currentRows[msg.index]))
				copy(updatedRow, currentRows[msg.index])
				updatedRow[0] = "[DELETED] " + strings.Replace(updatedRow[0], "[DELETING...] ", "", 1) // Modify the first column to indicate deletion
				newRows := make([]table.Row, len(currentRows))
				copy(newRows, currentRows)
				newRows[msg.index] = updatedRow
				m.table.SetRows(newRows)

				// Update the model
				if msg.index < len(m.modules) {
					m.totalSize -= m.modules[msg.index].Size
					// m.modules = append(m.modules[:msg.index], m.modules[msg.index+1:]...)
					m.modules = slices.Delete(m.modules, msg.index, msg.index+1)

					// Remove the deleted row from the table (optional, but might be desired)
					currentTableRows := m.table.Rows()
					if msg.index < len(currentTableRows) {
						r := slices.Delete(currentTableRows, msg.index, msg.index+1)
						// m.table.SetRows(append(currentTableRows[:msg.index], currentTableRows[msg.index+1:]...))
						m.table.SetRows(r)

						// Adjust cursor if necessary
						if m.table.Cursor() >= len(m.table.Rows()) && len(m.table.Rows()) > 0 {
							m.table.SetCursor(len(m.table.Rows()) - 1)
						}
					}
				}

				m.deletedPaths = append(m.deletedPaths, msg.path)
			}
		}
		return m, nil

	case deleteErrMsg:
		// Handle the error message.
		// Indicate that the deletion failed with [FAILED]

		if msg.index >= 0 && msg.index < len(m.table.Rows()) {
			currentRows := m.table.Rows()
			if len(currentRows[msg.index]) > 0 {
				updatedRow := make(table.Row, len(currentRows[msg.index]))
				// Remove the "[DELETING...]" prefix

				copy(updatedRow, currentRows[msg.index])
				updatedRow[0] = "[FAILED] " + strings.Replace(updatedRow[0], "[DELETING...] ", "", 1)
				newRows := make([]table.Row, len(currentRows))
				copy(newRows, currentRows)
				newRows[msg.index] = updatedRow
				m.table.SetRows(newRows)
			}
		}
	}

	return m, nil
}

// Update the View method's loading section
func (m model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf(
			"\nError: %s\n\nPress q to quit.",
			m.err.Error(),
		))
	}

	if m.isLoading {
		var b strings.Builder

		// Title
		b.WriteString("\n")
		b.WriteString(loadingTitleStyle.Render("📦 SWEEPY 📦"))
		b.WriteString("\n\n")

		// Scanning status centered with count
		dirCount := len(m.scanningPaths)
		status := scanningStatusStyle.Render(fmt.Sprintf(
			"%s %s %s",
			m.spinner.View(),
			scanningLabelStyle.Render("Scanning for node_modules..."),
			scanningCountStyle.Render(fmt.Sprintf("(%d found)", dirCount)),
		))
		b.WriteString(status)
		b.WriteString("\n\n")

		// Progress box with enhanced styling for paths
		if m.ctx.Verbose && len(m.scanningPaths) > 0 {
			var pathsContent strings.Builder

			// Show latest paths, most recent at bottom
			start := 0
			if len(m.scanningPaths) > 6 {
				start = len(m.scanningPaths) - 6
			}

			// Add a header
			pathsContent.WriteString(lipgloss.NewStyle().
				Foreground(colorHeader).
				Underline(true).
				Render("Recently scanned paths:") + "\n\n")

			for _, path := range m.scanningPaths[start:] {
				// Truncate long paths with ellipsis
				if len(path) > 70 {
					path = "..." + path[len(path)-67:]
				}
				pathsContent.WriteString(loadingPathStyle.Render(path))
				pathsContent.WriteString("\n")
			}

			// Add the progress box to the view
			b.WriteString(loadingBoxStyle.Render(pathsContent.String()))
			b.WriteString("\n")
		}

		// Footer with keybindings
		b.WriteString(loadingFooterStyle.Render("Press q or Ctrl+C to quit"))

		return b.String()
	}

	var b strings.Builder

	// Title with improved styling
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("📦 SWEEPY 📦"))
	b.WriteString("\n\n")

	// Stats with improved formatting
	stats := fmt.Sprintf(
		"%s %s\n%s %s\n%s %s\n%s %s\n",
		statsLabelStyle.Render("Found:"),
		statsValueStyle.Render(fmt.Sprintf("%d node_modules directories", len(m.modules))),
		statsLabelStyle.Render("Total Size:"),
		statsValueStyle.Render(fmt.Sprintf("%.2f MB", float64(m.totalSize)/1024/1024)),
		statsLabelStyle.Render("Avg Staleness:"),
		statsValueStyle.Render(fmt.Sprintf("%.2f days", m.avgStaleness)),
		statsLabelStyle.Render("Scan Duration:"),
		statsValueStyle.Render(m.scanDuration),
	)
	b.WriteString(statsStyle.Render(stats))
	b.WriteString("\n")

	// Table with border
	tableBorder := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(1, 1)

	b.WriteString(tableBorder.Render(m.table.View()))

	// Footer with improved styling
	b.WriteString("\n")
	footerText := "q/Ctrl+C: quit • ↑/↓: navigate • space: delete"
	enhancedFooter := lipgloss.NewStyle().
		Foreground(colorSecondary).
		Align(lipgloss.Center).
		Width(m.width - 4).
		Render(footerText)

	b.WriteString(enhancedFooter)
	b.WriteString("\n")

	return b.String()
}
