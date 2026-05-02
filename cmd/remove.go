package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <description>",
	Aliases: []string{"rm", "destroy"},
	Short:   "Remove a TODO item",
	Long: `Remove a TODO item from the specified project.

Examples:
  mktodo remove "Complete documentation"
  mktodo rm -p lego "Old task"
  mktodo rm -p lego.technic "Build crane" --yes`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented - planned for Phase 3.3")
	},
}

func init() {
	removeCmd.Flags().StringP("project", "p", "default", "project path")
	removeCmd.Flags().BoolP("yes", "y", false, "skip confirmation")
}
