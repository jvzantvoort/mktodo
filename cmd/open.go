package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "List open TODO items",
	Long: `List only open (not completed) TODO items.

Equivalent to: mktodo list --open

Examples:
  mktodo open
  mktodo open -p lego.technic`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented - planned for Phase 3.6")
	},
}

func init() {
	openCmd.Flags().StringP("project", "p", "", "filter by project")
}
