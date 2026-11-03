package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/internal/project"
)

// Update handles a message and returns the updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = size.Width
		m.height = size.Height
		m.list.SetSize(size.Width, size.Height-4)
		return m, nil
	}

	switch m.mode {
	case modeMenu:
		return m.updateMenu(msg)
	case modeEdit:
		return m.updateEdit(msg)
	case modeConfirm:
		return m.updateConfirm(msg)
	default:
		return m.updateNormal(msg)
	}
}

func (m Model) updateNormal(msg tea.Msg) (tea.Model, tea.Cmd) {
	// While the list's filter input is active, keys belong to the filter.
	if keyMsg, ok := msg.(tea.KeyMsg); ok && !m.list.SettingFilter() {
		switch keyMsg.String() {
		case "ctrl+c", "q":
			if m.hasChanges {
				m.mode = modeMenu
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		case "esc":
			m.mode = modeMenu
			return m, nil
		case " ":
			return m.toggleItem(), nil
		case "e":
			return m.startEdit()
		case "d":
			return m.confirmDelete(), nil
		case "a":
			return m.startAdd()
		case "s":
			m.save()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "esc":
		m.mode = modeNormal
	case "s":
		m.mode = modeNormal
		m.save()
	case "q":
		m.quitting = true
		return m, tea.Quit
	case "x":
		// Only quit once the changes are safely written.
		m.mode = modeNormal
		if m.save() {
			m.quitting = true
			return m, tea.Quit
		}
	case "up", "k":
		if m.menuCursor > 0 {
			m.menuCursor--
		}
	case "down", "j":
		if m.menuCursor < len(m.menuItems)-1 {
			m.menuCursor++
		}
	}

	return m, nil
}

func (m Model) updateEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc", "ctrl+c":
			return m.finishEdit(), nil
		case "enter":
			return m.commitEdit(), nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "y":
		m.mode = modeNormal
		if m.confirmAction != nil {
			m.confirmAction(&m)
			m.confirmAction = nil
		}
	case "n", "esc":
		m.mode = modeNormal
		m.confirmAction = nil
	}

	return m, nil
}

// selected returns the selected TODO item, if any.
func (m Model) selected() (todoItem, bool) {
	item, ok := m.list.SelectedItem().(todoItem)
	return item, ok
}

func (m Model) toggleItem() Model {
	item, ok := m.selected()
	if !ok {
		return m
	}
	item.item.MarkDone()
	m.hasChanges = true
	return m
}

func (m Model) startEdit() (tea.Model, tea.Cmd) {
	item, ok := m.selected()
	if !ok {
		return m, nil
	}
	m.adding = false
	m.input.SetValue(item.item.Description)
	m.input.CursorEnd()
	m.mode = modeEdit
	return m, m.input.Focus()
}

func (m Model) startAdd() (tea.Model, tea.Cmd) {
	m.adding = true
	m.input.SetValue("")
	m.mode = modeEdit
	return m, m.input.Focus()
}

// commitEdit applies the text input: it adds a new item or changes the
// selected item's description. Empty input changes nothing.
func (m Model) commitEdit() Model {
	value := m.input.Value()
	if value == "" {
		return m.finishEdit()
	}

	if m.adding {
		m.addItem(value)
	} else if item, ok := m.selected(); ok && item.item.Description != value {
		item.item.SetDescription(value)
		m.hasChanges = true
	}
	return m.finishEdit()
}

func (m Model) finishEdit() Model {
	m.mode = modeNormal
	m.adding = false
	m.input.Blur()
	m.input.SetValue("")
	return m
}

// addItem adds an item to the selected item's project, or to the default
// project when nothing with a project is selected, and selects it.
func (m *Model) addItem(description string) {
	proj := m.targetProject()
	if proj == nil {
		m.err = fmt.Errorf("no project to add the item to")
		return
	}

	newItem, err := m.doc.AddItem(proj, description)
	if err != nil {
		m.err = err
		return
	}
	m.hasChanges = true
	m.refreshList()
	for i, li := range m.list.Items() {
		if ti, ok := li.(todoItem); ok && ti.item == newItem {
			m.list.Select(i)
			break
		}
	}
}

func (m Model) targetProject() *project.Project {
	if item, ok := m.selected(); ok && item.project != nil {
		return item.project
	}
	if proj, ok := m.doc.Projects["default"]; ok {
		return proj
	}
	if roots := project.GetRootProjects(m.doc.Projects); len(roots) > 0 {
		return roots[0]
	}
	return nil
}

func (m Model) confirmDelete() Model {
	item, ok := m.selected()
	if !ok {
		return m
	}
	m.confirmMsg = fmt.Sprintf("Delete '%s'?", item.item.Description)
	m.confirmAction = func(m *Model) { m.deleteItem(item.item) }
	m.mode = modeConfirm
	return m
}

func (m *Model) deleteItem(item *markdown.Item) {
	if err := m.doc.RemoveItem(item); err != nil {
		m.err = err
		return
	}
	m.hasChanges = true
	index := m.list.Index()
	m.refreshList()
	if n := len(m.list.Items()); index >= n && n > 0 {
		m.list.Select(n - 1)
	}
}

// refreshList rebuilds the list from the document after items are added or removed.
func (m *Model) refreshList() {
	m.list.SetItems(buildItemList(m.doc))
}

// save writes the document atomically and reports whether it succeeded.
// A failure is shown in the status area and the changes are kept.
func (m *Model) save() bool {
	if err := markdown.SaveDocument(m.todoPath, m.doc); err != nil {
		m.err = fmt.Errorf("saving %s: %w", m.todoPath, err)
		return false
	}
	m.err = nil
	m.hasChanges = false
	return true
}
