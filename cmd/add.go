package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add <description>",
	Aliases: []string{"create"},
	Short:   "Add a new TODO item",
	Long: `Add a new TODO item to the specified project.

Examples:
  mktodo add "Complete documentation"
  mktodo add -p lego.technic "Build crane model"
  mktodo add "FIXME: Security vulnerability"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented - planned for Phase 3.2")
	},
}

func init() {
	addCmd.Flags().StringP("project", "p", "default", "project path (e.g., lego.technic)")
	addCmd.Flags().BoolP("yes", "y", false, "skip confirmations")
}
