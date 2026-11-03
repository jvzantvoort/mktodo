package cmd

import (
	"errors"
	"io/fs"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/internal/project"
	"github.com/jvzantvoort/mktodo/internal/tui"
	"github.com/jvzantvoort/mktodo/messages"
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

Keyboard shortcuts:
  space    - Toggle TODO done/undone
  e        - Edit selected item
  a        - Add new item
  d        - Delete selected item
  s        - Save changes
  esc      - Open menu
  q        - Quit

Examples:
  mktodo tui`,
	RunE: runTUI,
}

func runTUI(cmd *cobra.Command, args []string) error {
	repo, cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	// Get todo file path
	todoPath := repo.ResolvePath(cfg.TodoFile)

	// Load the document, or start an empty one if the file doesn't exist yet.
	// Any other error is reported rather than risking a save over the file.
	doc, err := markdown.LoadDocument(todoPath, cfg)
	if errors.Is(err, fs.ErrNotExist) {
		projects, buildErr := project.BuildHierarchy(cfg)
		if buildErr != nil {
			return messages.Errorf("ERR_CONFIG_INVALID", buildErr)
		}
		doc = markdown.NewDocument(todoPath, projects)
	} else if err != nil {
		return messages.Errorf("ERR_FILE_READ_FAILED", todoPath, err)
	}

	// Create TUI model
	model, err := tui.NewModel(cfg, doc, todoPath)
	if err != nil {
		return messages.Errorf("ERR_CONFIG_INVALID", err)
	}

	// Run TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return messages.Errorf("ERR_OPERATION_FAILED", err)
	}

	return nil
}
