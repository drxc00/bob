package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/drxc00/bob/internal/scan"
	"github.com/drxc00/bob/utils"
)

// Define styles
var (
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true)
	// focusedStyle = lipgloss.NewStyle().
	// 		BorderStyle(lipgloss.RoundedBorder()).
	// 		BorderForeground(lipgloss.Color("62"))
)

// Model represents the application state
type model struct {
	spinner       spinner.Model
	table         table.Model
	isLoading     bool
	scanComplete  bool
	modules       []scan.ScannedNodeModule
	scanPath      string
	staleness     int64
	noCache       bool
	err           error
	width, height int
	totalSize     int64
	avgStaleness  float64
}

// Initialize the model
func initialModel(scanPath string, staleness int64, noCache bool) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		spinner:      s,
		isLoading:    true,
		scanComplete: false,
		scanPath:     scanPath,
		staleness:    staleness,
		noCache:      noCache,
	}
}

// scanResultMsg is returned when the scan is complete
type scanResultMsg struct {
	modules []scan.ScannedNodeModule
	stats   scan.ScanInfo
	err     error
}

// scanCmd starts the scan in a goroutine and returns the result
func startScan(path string, staleness int64, noCache bool) tea.Cmd {
	return func() tea.Msg {
		modules, stats, err := scan.NodeScan(path, staleness, noCache)
		return scanResultMsg{modules: modules, stats: stats, err: err}
	}
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		startScan(m.scanPath, m.staleness, m.noCache),
	)
}

// Update handles messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Adjust table width
		if m.scanComplete {
			m.table.SetWidth(msg.Width - 4)
			m.table.SetHeight(msg.Height - 10)
		}

	case scanResultMsg:
		m.isLoading = false
		m.scanComplete = true
		m.modules = msg.modules
		m.err = msg.err

		m.totalSize = msg.stats.TotalSize
		m.avgStaleness = msg.stats.AvgStaleness

		if m.err != nil || msg.err != nil {
			utils.Log("Error when scanning: %v\n", m.err)
			return m, tea.Quit
		}

		// Create the table
		columns := []table.Column{
			{Title: "PATH", Width: 50},
			{Title: "SIZE", Width: 15},
			{Title: "STALENESS", Width: 15},
		}

		var rows []table.Row
		for _, module := range m.modules {
			staleness := fmt.Sprintf("%d days", module.Staleness)

			// Format size
			var sizeStr string
			if module.Size > 1024*1024 {
				sizeStr = fmt.Sprintf("%.2f MB", float64(module.Size)/1024/1024)
			} else if module.Size > 1024 {
				sizeStr = fmt.Sprintf("%.2f KB", float64(module.Size)/1024)
			} else {
				sizeStr = fmt.Sprintf("%d bytes", module.Size)
			}

			// Add row
			rows = append(rows, table.Row{module.Path, sizeStr, staleness})
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(m.height-10),
		)

		// Style the table
		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(true)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(true)
		t.SetStyles(s)

		m.table = t
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the current model
func (m model) View() string {
	if m.err != nil {
		errStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("9")). // red
			Padding(1, 2).
			Width(m.width - 4)

		return errStyle.Render(fmt.Sprintf(
			"An error occurred:\n\n%s\n\nPress q to quit.",
			m.err.Error(),
		))
	}

	if m.isLoading {
		return baseStyle.Render(fmt.Sprintf("\n %s Scanning %s...\n\n",
			m.spinner.View(), m.scanPath))
	}

	// Display results
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render(" bob scan [node_modules] "))
	b.WriteString("\n\n")

	// Stats
	b.WriteString(fmt.Sprintf(" Found %d node_modules directories\n", len(m.modules)))
	b.WriteString(fmt.Sprintf(" Total size: %.2f MB | Avg staleness: %.2f days\n",
		float64(m.totalSize)/1024/1024, m.avgStaleness))
	b.WriteString(fmt.Sprintf(" Path: %s | Staleness: %d days | Cache: %t\n\n",
		m.scanPath, m.staleness, !m.noCache))

	// Table
	b.WriteString(m.table.View())

	// Footer
	b.WriteString("\n\n Press q to quit • Arrow keys to navigate\n")

	return b.String()
}

// Function that starts the scan
func scanNode(stalenessFlag string, scanPath string, noCache bool) {
	var stalenessFlagInt int64
	var err error

	// Parse staleness flag
	if stalenessFlag == "0" {
		stalenessFlagInt = 0
	} else {
		stalenessFlagInt, err = utils.ParseStalenessFlagValue(stalenessFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing staleness flag: %v\n", err)
			os.Exit(1)
		}
	}

	// Create and start the BubbleTea program
	p := tea.NewProgram(
		initialModel(scanPath, stalenessFlagInt, noCache),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		utils.Log("Error when scanning: %v\n", err)
		os.Exit(1)
	}
}
