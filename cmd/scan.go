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

// --- Styles ---

var (
	colorPrimary   = lipgloss.Color("205") // Purple
	colorSecondary = lipgloss.Color("99")  // Pink
	colorBorder    = lipgloss.Color("240") // Gray
	colorSelected  = lipgloss.Color("57")  // Dark blue
	colorError     = lipgloss.Color("9")   // Red
	colorHighlight = lipgloss.Color("229") // Light Yellow

	baseStyle  = lipgloss.NewStyle().Foreground(colorPrimary)
	titleStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true).
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 2)

	statsStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2).
			Align(lipgloss.Center)

	errorStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorError).
			Padding(1, 2).
			Align(lipgloss.Center)

	footerStyle = lipgloss.NewStyle().
			Foreground(colorBorder).
			Align(lipgloss.Center)
)

// --- Model ---

type model struct {
	spinner       spinner.Model
	table         table.Model
	isLoading     bool
	scanComplete  bool
	modules       []scan.ScannedNodeModule
	scanPath      string
	staleness     int64
	noCache       bool
	resetCache    bool
	err           error
	width, height int
	totalSize     int64
	avgStaleness  float64
}

// --- Init Functions ---

func initialModel(scanPath string, staleness int64, noCache bool, resetCache bool) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = baseStyle
	return model{
		spinner:      s,
		isLoading:    true,
		scanComplete: false,
		scanPath:     scanPath,
		staleness:    staleness,
		noCache:      noCache,
		resetCache:   resetCache,
	}
}

type scanResultMsg struct {
	modules []scan.ScannedNodeModule
	stats   scan.ScanInfo
	err     error
}

func startScan(path string, staleness int64, noCache bool, resetCache bool) tea.Cmd {
	return func() tea.Msg {
		modules, stats, err := scan.NodeScan(path, staleness, noCache, resetCache)
		return scanResultMsg{modules: modules, stats: stats, err: err}
	}
}

// --- BubbleTea Handlers ---

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		startScan(m.scanPath, m.staleness, m.noCache, m.resetCache),
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
		}
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.scanComplete {
			m.table.SetWidth(m.width - 4)
			m.table.SetHeight(m.height - 10)
		}
	case scanResultMsg:
		m.isLoading = false
		m.scanComplete = true
		m.modules = msg.modules
		m.err = msg.err
		m.totalSize = msg.stats.TotalSize
		m.avgStaleness = msg.stats.AvgStaleness

		if m.err != nil {
			utils.Log("Error scanning: %v\n", m.err)
			return m, tea.Quit
		}

		columns := []table.Column{
			{Title: "PATH", Width: 50},
			{Title: "SIZE", Width: 15},
			{Title: "STALENESS", Width: 15},
		}

		var rows []table.Row
		for _, module := range m.modules {
			rows = append(rows, table.Row{
				module.Path,
				utils.FormatSize(module.Size),
				utils.ColorCodedStaleness(module.Staleness),
			})
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(m.height-10),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(colorBorder).BorderBottom(true).Bold(true)
		s.Selected = s.Selected.Foreground(colorHighlight).Background(colorSelected).Bold(true)
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

func (m model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf(
			"\nAn error occurred:\n\n%s\n\nPress q to quit.",
			m.err.Error(),
		))
	}

	if m.isLoading {
		return baseStyle.Render(fmt.Sprintf("\n%s Scanning...\n", m.spinner.View()))
	}

	var b strings.Builder

	// Title
	b.WriteString("\n")
	b.WriteString(titleStyle.Render(" SCAN RESULTS "))
	b.WriteString("\n\n")

	// Path + Settings
	b.WriteString(fmt.Sprintf(" Path: %s | Staleness: %d days | Cache: %t\n\n",
		m.scanPath, m.staleness, !m.noCache))

	// Stats with nice border
	stats := fmt.Sprintf(
		"Found %d node_modules directories\nTotal Size: %.2f MB | Avg Staleness: %.2f days",
		len(m.modules),
		float64(m.totalSize)/1024/1024,
		m.avgStaleness,
	)
	b.WriteString(statsStyle.Render(stats))
	b.WriteString("\n\n")

	// Table
	b.WriteString(m.table.View())

	// Footer
	b.WriteString("\n")
	b.WriteString(footerStyle.Render("Press q to quit • Arrow keys to navigate"))
	b.WriteString("\n")

	return b.String()
}

// Function that starts the scan
func scanNode(stalenessFlag string, scanPath string, noCache bool, resetCacheFlag bool) {
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
		initialModel(scanPath, stalenessFlagInt, noCache, resetCacheFlag),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		utils.Log("Error when scanning: %v\n", err)
		os.Exit(1)
	}
}
