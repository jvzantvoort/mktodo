// Package cmd implements the mktodo command-line interface.
package cmd

import (
	"fmt"
	"io"
	"os"

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
	// Plain "mktodo" shows help; an unknown subcommand is a usage error.
	// The root needs a RunE for cobra to check its arguments at all.
	Args: usageArgs(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	// Errors are printed once by Execute; usage is only shown for usage errors.
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute runs the root command, prints any error to stderr, and returns
// the process exit code.
func Execute() int {
	return execute(os.Stderr, os.Args[1:])
}

func execute(stderr io.Writer, args []string) int {
	rootCmd.SetArgs(args)
	cmd, err := rootCmd.ExecuteC()
	if err == nil {
		return ExitOK
	}

	code := exitCode(err)
	_, _ = fmt.Fprintln(stderr, formatError(err))
	if code == ExitUsage {
		_, _ = fmt.Fprintf(stderr, "Run '%s --help' for usage.\n", cmd.CommandPath())
	}
	return code
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default: .mktodo.yml in git root)")
	rootCmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return usageError(err)
	})

	// Add subcommands
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(tuiCmd)
}
