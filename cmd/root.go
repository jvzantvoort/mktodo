// Package cmd implements the mktodo command-line interface.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version string
	commit  string
	date    string
)

// SetVersion sets the version information
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
	rootCmd.Version = fmt.Sprintf("%s (commit %s, built %s)", v, c, d)
}

var rootCmd = &cobra.Command{
	Use:   "mktodo",
	Short: "Manage TODO items in markdown files",
	Long: `mktodo is a CLI and TUI tool for managing TODO items within markdown files in git repositories.

It supports hierarchical project organization, FIXME prioritization, and preserves all non-TODO content.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default: .mktodo.yml in git root)")

	// Add subcommands
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(tuiCmd)
}
