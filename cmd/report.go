package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a TODO report",
	Long: `Generate a graphical report of TODO items per project.

Examples:
  mktodo report
  mktodo report --format html -o report.html`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented - planned for Phase 3.7")
	},
}

func init() {
	reportCmd.Flags().String("format", "text", "output format: text|html|markdown|json")
	reportCmd.Flags().StringP("output", "o", "", "output file (default: stdout)")
}
