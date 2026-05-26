package tui

import (
	"github.com/charmbracelet/bubbles/list"
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

type Model struct {
	config   *config.Config
	doc      *markdown.Document
	projects []*project.Project
	list     list.Model
	items    []list.Item
	mode     mode

	// Edit state
	editInput string
	isAdding  bool // true when adding a new item, false when editing existing

	// Confirm dialog state
	confirmMsg  string
	confirmIdx  int  // list index of item pending deletion (-1 if none)
	confirmQuit bool // true = confirming quit-with-changes, false = confirming delete

	// Clipboard for cut/paste
	clipboard        *markdown.Item
	clipboardProject *project.Project

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
		return "(no project)"
	}
	return "Project: " + t.project.Title
}

func (t todoItem) FilterValue() string {
	return t.item.Description
}

type saveMsg struct {
	err error
}
