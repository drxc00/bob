/*
Copyright © 2025 Neil Patrick Villanueva npdvillanueva@gmail.com
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/drxc00/bob/internal/scan"
	"github.com/drxc00/bob/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:              "scan [directory] [flags]",
	Short:            "Scan your development environment for clutter",
	Long:             `Scan your development environment for clutter like node_modules folders`,
	Args:             cobra.MaximumNArgs(1),
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		var scanPath string
		stalenessFlag, err := cmd.Flags().GetString("staleness")
		var stalenessFlagInt int64

		fmt.Println("flag", stalenessFlag)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting staleness flag: %v\n", err)
			os.Exit(1)
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

		// Check the args
		if len(args) > 0 {
			scanPath = args[0]
		} else {
			// Set the current directory as the default scan path
			// If no arguments are provided.
			currentDir, err := os.Getwd() // Get the current directory
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
				os.Exit(1)
			}

			// Convert the current directory to a Windows path
			scanPath = filepath.ToSlash(currentDir)
		}

		// Print the path we're scanning
		fmt.Printf("Scanning directory: %s\n", scanPath)

		scannedNodeModules, err := scan.NodeScan(scanPath, stalenessFlagInt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		}

		fmt.Printf("Found %d node_modules directories\n", len(scannedNodeModules))

		// Print

		// Table printer
		t := table.NewWriter()
		t.SetStyle(table.StyleLight)
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Path", "Size", "Staleness"})

		for _, scannedNodeModule := range scannedNodeModules {
			// fmt.Printf("- %s (size: %d bytes) - Staleness: %d days\n", scannedNodeModule.Path, scannedNodeModule.Size, scannedNodeModule.Staleness)
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
	},
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bob",
	Short: "Your Terminal Janitor for Cleaning Up Your Development Environment",
	Long:  `bob is a lightweight, dependency-free CLI tool that helps you keep your development environment clean and clutter-free.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add all commands to the root command
	rootCmd.AddCommand(scanCmd)

	// Flags to scanCmd
	scanCmd.Flags().StringP("staleness", "s", "0", `
	The staleness of the node_modules directory. Accepts the following formats: 1d, 1h, 1m, 1s
	If no units are specified, it defaults to days.
	`)

}
