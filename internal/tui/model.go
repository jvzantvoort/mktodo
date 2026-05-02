package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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
	editInput  string
	editCursor int

	// Confirm dialog state
	confirmMsg    string
	confirmAction func() tea.Msg

	// Menu state
	menuVisible bool
	menuItems   []string
	menuCursor  int

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
	return "Project: " + t.project.Title
}

func (t todoItem) FilterValue() string {
	return t.item.Description
}

type confirmMsg struct {
	confirmed bool
}

type saveMsg struct {
	err error
}
