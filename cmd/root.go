/*
Copyright © 2025 Neil Patrick Villanueva npdvillanueva@gmail.com
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/drxc00/bob/internal/scan"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [directory]",
	Short: "Scan your development environment for clutter",
	Long:  `Scan your development environment for clutter like node_modules folders`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scanPath := "." // Default to current directory
		if len(args) > 0 {
			scanPath = args[0]
		}

		// Print the path we're scanning
		fmt.Printf("Scanning directory: %s\n", scanPath)

		paths, err := scan.FindNodeModules(scanPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
			os.Exit(1)
		}

		if len(paths) > 0 {
			fmt.Printf("Found node_modules directories:\n")
			for _, path := range paths {
				fmt.Printf("- %s\n", path)
			}
		} else {
			fmt.Println("No node_modules directories found")
		}
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

	// Add any command-specific flags here
	// scanCmd.Flags().BoolP("recursive", "r", false, "Recursively scan subdirectories")
}
