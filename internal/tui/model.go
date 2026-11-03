// Package tui implements the interactive terminal UI for managing TODO items.
//
// The screen has two panes: a project tree on the left and the items of the
// selected project on the right (spec FR-7). All changes go through the
// markdown.Document, so saving writes exactly what is shown.
package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/internal/todo"
)

type mode int

const (
	modeNormal mode = iota
	modeEdit
	modeConfirm
	modeMenu
	modeHelp
	modeConflict
)

type pane int

const (
	paneProjects pane = iota
	paneItems
)

// Default size until the terminal reports its real size.
const (
	defaultWidth  = 80
	defaultHeight = 24
)

// Model is the Bubbletea model for the TUI.
type Model struct {
	config   *config.Config
	doc      *markdown.Document
	todoPath string
	version  string

	mode  mode
	focus pane

	// Project tree state. collapsed is keyed by project name.
	collapsed     map[string]bool
	projectCursor int
	itemCursor    int

	// Edit state: input holds the text being edited; adding is true when
	// the input creates a new item instead of changing the selected one.
	input  textinput.Model
	adding bool

	// Confirm dialog state. The action runs when the user answers "y".
	confirmMsg    string
	confirmAction func(*Model) tea.Cmd

	// clipboard holds an item that was cut and not yet pasted.
	clipboard *markdown.Item

	// quitAfterSave remembers a save-and-quit while the conflict dialog is
	// open, so overwriting also quits.
	quitAfterSave bool

	hasChanges bool
	status     string // feedback for the last action
	err        error

	width    int
	height   int
	quitting bool
}

// NewModel builds the TUI model for the given configuration and document.
func NewModel(cfg *config.Config, doc *markdown.Document, todoPath string) (Model, error) {
	input := textinput.New()
	input.Placeholder = "Description"
	input.CharLimit = todo.MaxDescriptionLength

	return Model{
		config:    cfg,
		doc:       doc,
		todoPath:  todoPath,
		version:   "dev",
		focus:     paneItems,
		collapsed: make(map[string]bool),
		input:     input,
		width:     defaultWidth,
		height:    defaultHeight,
	}, nil
}

// WithVersion sets the version shown in the header.
func (m Model) WithVersion(version string) Model {
	m.version = version
	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}
