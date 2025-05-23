/*
Copyright © 2025 Neil Patrick Villanueva npdvillanueva@gmail.com
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/drxc00/sweepy/cmd/tui"
	"github.com/drxc00/sweepy/types"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "sweepy",
	Short:            "Your Terminal Janitor for Cleaning Up Your Development Environment",
	Long:             `Sweepy is a lightweight, dependency-free CLI tool that helps you keep your development environment clean and clutter-free.`,
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		// Vars
		var scanPath string

		// Flags
		stalenessFlag, errStalenessFlag := cmd.Flags().GetString("staleness")
		noCacheFlag, errNoCacheFlag := cmd.Flags().GetBool("no-cache")
		resetCacheFlag, errResetCacheFlag := cmd.Flags().GetBool("reset-cache")

		if errResetCacheFlag != nil {
			fmt.Fprintf(os.Stderr, "Error getting reset-cache flag: %v\n", errResetCacheFlag)
			os.Exit(1)
		}

		if errStalenessFlag != nil {
			fmt.Fprintf(os.Stderr, "Error getting staleness flag: %v\n", errStalenessFlag)
			os.Exit(1)
		}

		if errNoCacheFlag != nil {
			fmt.Fprintf(os.Stderr, "Error getting no-cache flag: %v\n", errNoCacheFlag)
			os.Exit(1)
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

		ctx := types.NewScanContext(scanPath,
			stalenessFlag,
			noCacheFlag,
			resetCacheFlag,
		)

		tui.ScanNode(ctx)

	},
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

	// Flags to scanCmd
	rootCmd.Flags().StringP("staleness", "s", "0", "The staleness of the node_modules directory. Accepts input in days. If no units are specified, it defaults to days.")
	rootCmd.Flags().BoolP("no-cache", "c", false, "Disable caching")
	rootCmd.Flags().BoolP("reset-cache", "r", false, "Reset the cache")
	rootCmd.Flags().BoolP("system", "y", false, "Scan the entire system for node_modules directories. Determines all drives and scans them.")

}
