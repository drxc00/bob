package cmd

import (
	"fmt"
	"os"

	"github.com/drxc00/bob/internal/scan"
	"github.com/drxc00/bob/utils"
	"github.com/jedib0t/go-pretty/v6/table"
)

func scanNode(stalenessFlag string, scanPath string) {
	// Vars
	var (
		stalenessFlagInt   int64
		err                error
		scannedNodeModules []scan.ScannedNodeModule
	)

	// Cache handler
	cache := utils.NewCache[scan.ScannedNodeModule](scanPath)
	ok, err := cache.Load() // Returns true if the cache is loaded, false if it's not

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading cache: %v\n", err)
	}

	if stalenessFlag == "0" {
		fmt.Println("Staleness flag not set, defaulting to 0")
		stalenessFlagInt = 0
	} else {
		stalenessFlagInt, err = utils.ParseStalenessFlagValue(stalenessFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing staleness flag: %v\n", err)
			os.Exit(1)
		}
	}

	if !ok || cache.IsExpired() {
		scannedNodeModules, err = scan.NodeScan(scanPath, stalenessFlagInt)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		}
	} else {
		scannedNodeModules = cache.GetAll() // Get all the cached data
	}

	fmt.Printf("Found %d node_modules directories\n", len(scannedNodeModules))

	// Table printer
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Path", "Size", "Staleness"})

	for _, scannedNodeModule := range scannedNodeModules {

		staleness := fmt.Sprintf("%d days", scannedNodeModule.Staleness)

		var sizeStr string
		if scannedNodeModule.Size > 1024 {
			sizeStr = fmt.Sprintf("%.2f MB", float64(scannedNodeModule.Size)/1024/1024)
		} else {
			sizeStr = fmt.Sprintf("%d bytes", scannedNodeModule.Size)
		}

		t.AppendRow([]interface{}{scannedNodeModule.Path, sizeStr, staleness})
		t.AppendSeparator()
	}

	// Render after loop
	t.Render()
}
