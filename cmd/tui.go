package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive terminal UI",
	Long: `Launch the interactive terminal UI for managing TODO items.

The TUI provides:
  - Project tree navigation
  - TODO item editing
  - Keyboard shortcuts for all operations
  - Real-time markdown file updates

Examples:
  mktodo tui`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented - planned for Phase 5.9")
	},
}
