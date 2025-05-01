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

	footerStyle = lipgloss.NewStyle().
			Foreground(colorBorder).
			Align(lipgloss.Center)

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
	modules []types.ScannedNodeModule
	stats   types.ScanInfo
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
			Background(colorSecondary).
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
		return m, listenForProgress(m.progressChan)

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
				formattedPaths = append(formattedPaths, path)
			}
			paths := strings.Join(formattedPaths, "\n")
			b.WriteString(loadingBoxStyle.Render(paths))
		}

		return baseStyle.Render(b.String())
	}

	var b strings.Builder

	// Title with improved styling
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("📦 NODE_MODULES SCAN RESULTS 📦"))
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
	footerText := "q/Ctrl+C: quit • ↑/↓: navigate • d: delete"
	enhancedFooter := lipgloss.NewStyle().
		Foreground(colorSecondary).
		Align(lipgloss.Center).
		Width(m.width - 4).
		Render(footerText)

	b.WriteString(enhancedFooter)
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
