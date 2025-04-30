package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/drxc00/sweepy/internal/scan"
	"github.com/drxc00/sweepy/types"
	"github.com/drxc00/sweepy/utils"
)

// --- Styles ---

var (
	colorPrimary   = lipgloss.Color("205") // Purple
	colorSecondary = lipgloss.Color("99")  // Pink
	colorBorder    = lipgloss.Color("240") // Gray
	colorSelected  = lipgloss.Color("25")  // Green
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

	loadingBoxStyle = lipgloss.NewStyle().
			BorderForeground(colorBorder).
			Height(12).
			Width(80).
			MaxHeight(12).
			MaxWidth(80).
			Align(lipgloss.Left)

	// Add a style for the progress text
	progressTextStyle = lipgloss.NewStyle().
				Width(76) // Slightly less than box width to account for padding

	// Add a style for truncating long paths
	pathStyle = lipgloss.NewStyle().
			Width(76).
			Foreground(colorPrimary)
)

// --- Model ---

type model struct {
	spinner      spinner.Model
	table        table.Model
	isLoading    bool
	scanComplete bool
	modules      []scan.ScannedNodeModule

	// Config
	ctx types.ScanContext

	// Verbose
	progressChan  chan string
	scanningPaths []string

	err           error
	width, height int
	totalSize     int64
	avgStaleness  float64
	scanDuration  string
}

// --- Init Functions ---

func initialModel(ctx types.ScanContext) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = baseStyle
	return model{
		spinner:      s,
		isLoading:    true,
		scanComplete: false,
		ctx:          ctx,
		progressChan: make(chan string, 1000),
	}
}

type scanResultMsg struct {
	modules []scan.ScannedNodeModule
	stats   scan.ScanInfo
	err     error
}

type scanProgressMsg struct {
	path string
}

func startScan(ctx types.ScanContext, progressChan chan string) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			modules, stats, err := scan.NodeScan(ctx, progressChan)
			return scanResultMsg{modules: modules, stats: stats, err: err}
		},
		listenForProgress(progressChan),
	)
}

func listenForProgress(progressChan chan string) tea.Cmd {
	return func() tea.Msg {
		if p, ok := <-progressChan; ok {
			return scanProgressMsg{path: p}
		}
		return nil
	}
}

// --- BubbleTea Handlers ---

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		startScan(m.ctx, m.progressChan),
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
		m.scanDuration = msg.stats.ScanDuration.String()

		if m.err != nil {
			utils.Log("Error scanning: %v\n", m.err)
			return m, tea.Quit
		}

		columns := []table.Column{
			{Title: "PROJECT", Width: 20},
			{Title: "PATH", Width: 50},
			{Title: "SIZE", Width: 15},
			{Title: "LAST MODIFIED", Width: 20},
			{Title: "STALENESS", Width: 15},
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
			table.WithHeight(m.height-8),
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

	case scanProgressMsg:
		m.scanningPaths = append(m.scanningPaths, msg.path)
		return m, listenForProgress(m.progressChan)

	}

	return m, nil
}

// Update the View method's loading section
func (m model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf(
			"\nAn error occurred:\n\n%s\n\nPress q to quit.",
			m.err.Error(),
		))
	}

	if m.isLoading {
		var b strings.Builder

		// Scanning status
		status := fmt.Sprintf("%s Scanning for node_modules...", m.spinner.View())
		b.WriteString(status)
		b.WriteString("\n\n")

		// Progress box
		if m.ctx.Verbose && len(m.scanningPaths) > 0 {
			start := 0
			if len(m.scanningPaths) > 8 {
				start = len(m.scanningPaths) - 8
			}
			var formattedPaths []string
			for _, path := range m.scanningPaths[start:] {
				// Truncate long paths with ellipsis
				if len(path) > 76 {
					path = path[:73] + "..."
				}
				formattedPaths = append(formattedPaths, pathStyle.Render(path))
			}
			paths := strings.Join(formattedPaths, "\n")
			b.WriteString(loadingBoxStyle.Render(paths))
		}

		return baseStyle.Render(b.String())
	}

	var b strings.Builder

	// Title
	b.WriteString("\n")
	b.WriteString(titleStyle.Render(" SCAN RESULTS "))
	b.WriteString("\n\n")

	// Path + Settings
	b.WriteString(fmt.Sprintf(" Path: %s | Staleness Flag: %d days | Cache: %t\n\n",
		m.ctx.Path, m.ctx.Staleness, !m.ctx.NoCache))

	// Stats with nice border
	stats := fmt.Sprintf(
		"Found %d node_modules directories\nTotal Size: %.2f MB | Avg Staleness: %.2f days\nScan Duration: %s\n",
		len(m.modules),
		float64(m.totalSize)/1024/1024,
		m.avgStaleness,
		m.scanDuration,
	)
	b.WriteString(statsStyle.Render(stats))
	b.WriteString("\n\n")

	// Table
	b.WriteString(m.table.View())

	// Footer
	b.WriteString("\n")
	b.WriteString(footerStyle.Render("Press q or Ctl+C to quit • Arrow keys to navigate • Press d to delete"))
	b.WriteString("\n")

	return b.String()
}

// Function that starts the scan
func scanNode(ctx types.ScanContext) {
	p := tea.NewProgram(
		initialModel(ctx),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		utils.Log("Error when scanning: %v\n", err)
		os.Exit(1)
	}
}
