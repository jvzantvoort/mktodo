package cmd

import (
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
}

var rootCmd = &cobra.Command{
	Use:   "mktodo",
	Short: "Manage TODO items in markdown files",
	Long: `mktodo is a CLI and TUI tool for managing TODO items within markdown files in git repositories.

It supports hierarchical project organization, FIXME prioritization, and preserves all non-TODO content.`,
	Version: version,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default: .mktodo.yml in git root)")
	rootCmd.SetVersionTemplate("{{.Version}}\n")
}
