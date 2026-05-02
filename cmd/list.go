package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List TODO items",
	Long: `List TODO items with optional filtering.

Examples:
  mktodo list
  mktodo ls -o
  mktodo ls -p lego --format json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented - planned for Phase 3.5")
	},
}

func init() {
	listCmd.Flags().StringP("project", "p", "", "filter by project")
	listCmd.Flags().BoolP("open", "o", false, "show only open items")
	listCmd.Flags().BoolP("done", "d", false, "show only completed items")
	listCmd.Flags().String("format", "text", "output format: text|json|yaml")
}
