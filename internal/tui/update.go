package tui

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/internal/project"
	"github.com/jvzantvoort/mktodo/internal/todo"
)

// Update handles a message and returns the updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = size.Width, size.Height
		return m, nil
	}

	switch m.mode {
	case modeMenu:
		return m.updateMenu(msg)
	case modeEdit:
		return m.updateEdit(msg)
	case modeConfirm:
		return m.updateConfirm(msg)
	case modeConflict:
		return m.updateConflict(msg)
	case modeHelp:
		if _, ok := msg.(tea.KeyMsg); ok {
			m.mode = modeNormal
		}
		return m, nil
	default:
		return m.updateNormal(msg)
	}
}

// normalKeys maps keys in the main view to their actions (spec FR-7).
var normalKeys = map[string]func(*Model) tea.Cmd{
	"up": moveUp, "k": moveUp,
	"down": moveDown, "j": moveDown,
	"left": focusProjects, "h": focusProjects,
	"right": focusItems, "l": focusItems,
	"tab":    switchPane,
	"enter":  (*Model).activate,
	" ":      (*Model).toggleItem,
	"e":      (*Model).startEdit,
	"d":      (*Model).confirmDelete,
	"a":      (*Model).startAdd,
	"-":      (*Model).moveItemUp,
	"+":      (*Model).moveItemDown,
	"x":      (*Model).cutItem,
	"p":      (*Model).pasteItem,
	"s":      func(m *Model) tea.Cmd { return m.save(false) },
	"?":      func(m *Model) tea.Cmd { m.mode = modeHelp; return nil },
	"esc":    func(m *Model) tea.Cmd { m.mode = modeMenu; return nil },
	"q":      (*Model).quit,
	"ctrl+c": (*Model).quit,
}

func (m Model) updateNormal(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	action, ok := normalKeys[keyMsg.String()]
	if !ok {
		return m, nil
	}
	m.err, m.status = nil, ""
	cmd := action(&m)
	m.clamp()
	return m, cmd
}

func moveUp(m *Model) tea.Cmd   { m.moveCursor(-1); return nil }
func moveDown(m *Model) tea.Cmd { m.moveCursor(1); return nil }

func (m *Model) moveCursor(delta int) {
	if m.focus == paneProjects {
		m.projectCursor = clampIndex(m.projectCursor+delta, len(m.rows()))
		m.itemCursor = 0
		return
	}
	m.itemCursor = clampIndex(m.itemCursor+delta, len(m.items()))
}

func focusProjects(m *Model) tea.Cmd { m.focus = paneProjects; return nil }
func focusItems(m *Model) tea.Cmd    { m.focus = paneItems; return nil }

func switchPane(m *Model) tea.Cmd {
	if m.focus == paneProjects {
		m.focus = paneItems
	} else {
		m.focus = paneProjects
	}
	return nil
}

// activate handles Enter: it expands or collapses a project with children,
// opens a project without children, or edits the selected item.
func (m *Model) activate() tea.Cmd {
	if m.focus == paneItems {
		return m.startEdit()
	}
	row, ok := m.selectedRow()
	if !ok {
		return nil
	}
	if row.hasChildren {
		m.collapsed[row.project.Name] = !m.collapsed[row.project.Name]
		return nil
	}
	m.focus = paneItems
	return nil
}

// itemInFocus returns the selected item when the items pane has focus.
func (m *Model) itemInFocus() (*markdown.Item, bool) {
	if m.focus != paneItems {
		return nil, false
	}
	return m.selectedItem()
}

func (m *Model) toggleItem() tea.Cmd {
	if item, ok := m.itemInFocus(); ok {
		item.MarkDone()
		m.hasChanges = true
	}
	return nil
}

func (m *Model) startEdit() tea.Cmd {
	item, ok := m.itemInFocus()
	if !ok {
		return nil
	}
	m.adding = false
	m.input.SetValue(item.Description)
	m.input.CursorEnd()
	m.mode = modeEdit
	return m.input.Focus()
}

func (m *Model) startAdd() tea.Cmd {
	if m.selectedProject() == nil {
		m.err = fmt.Errorf("select a project to add the item to")
		return nil
	}
	m.adding = true
	m.input.SetValue("")
	m.mode = modeEdit
	return m.input.Focus()
}

func (m *Model) confirmDelete() tea.Cmd {
	item, ok := m.itemInFocus()
	if !ok {
		return nil
	}
	m.confirmMsg = fmt.Sprintf("Delete '%s'?", item.Description)
	m.confirmAction = func(m *Model) tea.Cmd {
		if err := m.doc.RemoveItem(item); err != nil {
			m.err = err
			return nil
		}
		m.hasChanges = true
		m.clamp()
		return nil
	}
	m.mode = modeConfirm
	return nil
}

func (m *Model) moveItemUp() tea.Cmd   { m.moveItem(-1); return nil }
func (m *Model) moveItemDown() tea.Cmd { m.moveItem(1); return nil }

// moveItem swaps the selected item with its neighbour in the same project.
func (m *Model) moveItem(delta int) {
	item, ok := m.itemInFocus()
	if !ok || item.Project == nil {
		return
	}
	items := m.items()
	target := m.itemCursor + delta
	if target < 0 || target >= len(items) {
		return
	}
	if err := m.doc.SwapItems(item, items[target]); err != nil {
		m.err = err
		return
	}
	m.itemCursor = target
	m.hasChanges = true
}

// cutItem removes the selected item and keeps it for pasting. Only one item
// is held at a time, so a cut item is never dropped by a second cut.
func (m *Model) cutItem() tea.Cmd {
	item, ok := m.itemInFocus()
	if !ok {
		return nil
	}
	if m.clipboard != nil {
		m.err = fmt.Errorf("paste '%s' before cutting another item", m.clipboard.Description)
		return nil
	}
	if err := m.doc.RemoveItem(item); err != nil {
		m.err = err
		return nil
	}
	m.clipboard = item
	m.hasChanges = true
	m.status = fmt.Sprintf("Cut '%s'. Select a project and press p to paste.", item.Description)
	return nil
}

// pasteItem puts the cut item into the selected project: after the selected
// item when the items pane has focus, otherwise at the end of the list.
func (m *Model) pasteItem() tea.Cmd {
	if m.clipboard == nil {
		return nil
	}
	proj := m.selectedProject()
	if proj == nil {
		m.err = fmt.Errorf("select a project to paste into")
		return nil
	}

	after, _ := m.itemInFocus()
	item := m.clipboard
	if err := m.doc.InsertItem(proj, after, item); err != nil {
		m.err = err
		return nil
	}
	m.clipboard = nil
	m.hasChanges = true
	m.focus = paneItems
	m.selectItem(item)
	m.status = fmt.Sprintf("Pasted '%s' into %s.", item.Description, proj.Title)
	return nil
}

// quit exits, asking first when there are unsaved changes.
func (m *Model) quit() tea.Cmd {
	if !m.hasChanges {
		m.quitting = true
		return tea.Quit
	}
	m.confirmMsg = "Quit without saving your changes?"
	m.confirmAction = func(m *Model) tea.Cmd {
		m.quitting = true
		return tea.Quit
	}
	m.mode = modeConfirm
	return nil
}

func (m Model) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	m.mode = modeNormal
	switch keyMsg.String() {
	case "s":
		return m, m.save(false)
	case "q":
		return m, m.quit()
	case "x":
		// Only quit once the changes are safely written.
		return m, m.save(true)
	case "esc":
	default:
		m.mode = modeMenu // ignore other keys while the menu is open
	}
	return m, nil
}

func (m Model) updateEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc", "ctrl+c":
			m.finishEdit()
			return m, nil
		case "enter":
			m.commitEdit()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// commitEdit applies the text input: it adds a new item to the selected
// project or changes the selected item. Empty input changes nothing. An
// invalid description keeps the dialog open with the error shown, so the
// text isn't lost.
func (m *Model) commitEdit() {
	if strings.TrimSpace(m.input.Value()) == "" {
		m.finishEdit()
		return
	}
	value, err := todo.CleanDescription(m.input.Value())
	if err != nil {
		m.err = err
		return
	}

	if m.adding {
		m.addItem(value)
	} else if item, ok := m.selectedItem(); ok && item.Description != value {
		item.SetDescription(value)
		m.hasChanges = true
	}
	m.finishEdit()
}

func (m *Model) addItem(description string) {
	proj := m.selectedProject()
	if proj == nil {
		return
	}
	item, err := m.doc.AddItem(proj, description)
	if err != nil {
		m.err = err
		return
	}
	m.hasChanges = true
	m.focus = paneItems
	m.selectItem(item)
}

func (m *Model) finishEdit() {
	m.err = nil
	m.mode = modeNormal
	m.adding = false
	m.input.Blur()
	m.input.SetValue("")
}

func (m Model) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "y":
		m.mode = modeNormal
		action := m.confirmAction
		m.confirmAction = nil
		if action != nil {
			return m, action(&m)
		}
	case "n", "esc":
		m.mode = modeNormal
		m.confirmAction = nil
	}
	return m, nil
}

// save writes the document atomically. If the file was changed outside
// the TUI since it was loaded, it opens the conflict dialog instead of
// overwriting those changes. With quitAfter, the TUI quits once the
// changes are written. A failure is shown in the status line and the
// changes are kept.
func (m *Model) save(quitAfter bool) tea.Cmd {
	err := markdown.SaveDocument(m.todoPath, m.doc)
	if errors.Is(err, markdown.ErrModifiedExternally) {
		m.mode = modeConflict
		m.quitAfterSave = quitAfter
		return nil
	}
	return m.afterSave(err, quitAfter)
}

func (m *Model) afterSave(err error, quitAfter bool) tea.Cmd {
	if err != nil {
		m.err = fmt.Errorf("saving %s: %w", m.todoPath, err)
		return nil
	}
	m.err = nil
	m.hasChanges = false
	m.status = "Saved " + m.todoPath
	if quitAfter {
		m.quitting = true
		return tea.Quit
	}
	return nil
}

// updateConflict handles the dialog shown when the file changed on disk:
// reload it (dropping the unsaved changes), overwrite it, or cancel.
func (m Model) updateConflict(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	quitAfter := m.quitAfterSave
	switch keyMsg.String() {
	case "o":
		m.mode, m.quitAfterSave = modeNormal, false
		return m, m.afterSave(markdown.OverwriteDocument(m.todoPath, m.doc), quitAfter)
	case "r":
		m.mode, m.quitAfterSave = modeNormal, false
		m.reload()
	case "esc", "n":
		m.mode, m.quitAfterSave = modeNormal, false
		m.status = "Not saved: the file changed on disk."
	}
	return m, nil
}

// reload replaces the document with the file on disk.
func (m *Model) reload() {
	doc, err := markdown.LoadDocument(m.todoPath, m.config)
	if errors.Is(err, fs.ErrNotExist) {
		var projects map[string]*project.Project
		if projects, err = project.BuildHierarchy(m.config); err == nil {
			doc = markdown.NewDocument(m.todoPath, projects)
		}
	}
	if err != nil {
		m.err = fmt.Errorf("reloading %s: %w", m.todoPath, err)
		return
	}
	m.doc = doc
	m.clipboard = nil
	m.hasChanges = false
	m.clamp()
	m.status = "Reloaded " + m.todoPath
}
