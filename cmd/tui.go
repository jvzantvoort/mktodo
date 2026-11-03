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

The left pane shows the project tree, the right pane the items of the
selected project. Changes are kept in memory until you save.

Keys:
  ↑/↓ k/j      move up/down
  ←/→ h/l tab  switch between projects and items
  enter        expand/collapse project, or edit item
  space        toggle done
  e / d / a    edit / delete / add
  - / +        move item up / down
  x / p        cut / paste item
  s            save
  esc          menu (save, save and quit, quit)
  ?            help
  q            quit (asks if there are unsaved changes)

Colours are turned off when NO_COLOR is set or TERM=dumb.

Examples:
  mktodo tui`,
	Args: usageArgs(cobra.NoArgs),
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
			return configError(messages.Errorf("ERR_CONFIG_INVALID", buildErr))
		}
		doc = markdown.NewDocument(todoPath, projects)
	} else if err != nil {
		return fileError(messages.Errorf("ERR_FILE_READ_FAILED", todoPath, err))
	}

	// Create TUI model
	model, err := tui.NewModel(cfg, doc, todoPath)
	if err != nil {
		return configError(messages.Errorf("ERR_CONFIG_INVALID", err))
	}
	model = model.WithVersion(version)

	// Run TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return messages.Errorf("ERR_OPERATION_FAILED", err)
	}

	return nil
}
