package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/git"
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
	// Find git repository
	repo, err := git.FindRepository()
	if err != nil {
		return messages.Errorf("ERR_NOT_IN_GIT_REPO")
	}

	// Load configuration
	cfg, err := config.Load(repo.ConfigPath())
	if err != nil {
		return messages.Errorf("ERR_CONFIG_LOAD_FAILED", err)
	}

	// Get todo file path
	todoPath := repo.ResolvePath(cfg.TodoFile)

	// Parse or create document
	doc, err := markdown.LoadDocument(todoPath, cfg)
	if err != nil {
		// If file doesn't exist, create minimal document
		projectMap, _ := project.BuildHierarchy(cfg)
		doc = &markdown.Document{
			Path:     todoPath,
			Sections: []*markdown.Section{},
			Projects: projectMap,
			Items:    []*markdown.Item{},
		}
	}

	// Create TUI model
	model := tui.NewModel(cfg, doc, todoPath)

	// Run TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return messages.Errorf("ERR_OPERATION_FAILED", err)
	}

	return nil
}
