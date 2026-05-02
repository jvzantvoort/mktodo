package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:     "done <description>",
	Aliases: []string{"complete"},
	Short:   "Mark a TODO item as done",
	Long: `Mark a TODO item as done (toggle checkbox state).

Examples:
  mktodo done "Complete documentation"
  mktodo complete -p lego.technic "Build crane"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented - planned for Phase 3.4")
	},
}

func init() {
	doneCmd.Flags().StringP("project", "p", "default", "project path")
}
