package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/drxc00/sweepy/internal/clean"
	"github.com/drxc00/sweepy/internal/scan"
	"github.com/drxc00/sweepy/types"
	"github.com/drxc00/sweepy/utils"
)

// Function that starts the scan
func ScanNode(ctx types.ScanContext) {
	p := tea.NewProgram(
		initialModel(ctx),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		utils.Log("Error when scanning: %v\n", err)
		os.Exit(1)
	}
}

func DeleteNode(n types.ScannedNodeModule, idx int) tea.Msg {
	err := clean.CleanNodeModule(n.Path)
	if err != nil {
		utils.Log("Error deleting node_module: %v\n", err)
		return deleteErrMsg{err: err, index: idx, path: n.Path}
	}
	return deleteSuccessMsg{path: n.Path, index: idx, size: n.Size}
}

func StartScan(ctx types.ScanContext, progressChan chan string) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			modules, stats, err := scan.NodeScan(ctx, progressChan)
			return scanResultMsg{modules: modules, stats: stats, err: err}
		},
		ListenForProgress(progressChan),
	)
}

func ListenForProgress(progressChan chan string) tea.Cmd {
	return func() tea.Msg {
		if p, ok := <-progressChan; ok {
			return scanProgressMsg{path: p}
		}
		return nil
	}
}
