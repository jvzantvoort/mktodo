// Package tui implements the interactive terminal UI for managing TODO items.
package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/internal/project"
)

type mode int

const (
	modeNormal mode = iota
	modeEdit
	modeConfirm
	modeMenu
)

// Model is the Bubbletea model for the TUI.
type Model struct {
	config   *config.Config
	doc      *markdown.Document
	projects []*project.Project
	list     list.Model
	mode     mode

	// Edit state: input holds the text being edited; adding is true when
	// the input creates a new item instead of changing the selected one.
	input  textinput.Model
	adding bool

	// Confirm dialog state
	confirmMsg    string
	confirmAction func(*Model)

	// Menu state
	menuItems  []string
	menuCursor int

	// File state
	todoPath   string
	hasChanges bool

	// UI state
	width    int
	height   int
	err      error
	quitting bool
}

type todoItem struct {
	item    *markdown.Item
	project *project.Project
}

func (t todoItem) Title() string {
	status := "[ ]"
	if t.item.Done {
		status = "[✓]"
	}

	prefix := ""
	if t.item.IsFIXME {
		prefix = "🔴 FIXME: "
	}

	return prefix + status + " " + t.item.Description
}

func (t todoItem) Description() string {
	if t.project == nil {
		return "Project: (none)"
	}
	return "Project: " + t.project.Title
}

func (t todoItem) FilterValue() string {
	return t.item.Description
}
