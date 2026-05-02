package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jvzantvoort/mktodo/internal/markdown"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
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
			return m.toggleItem()

		case "e":
			return m.startEdit()

		case "d":
			return m.confirmDelete()

		case "a":
			return m.startAdd()

		case "s":
			return m.save()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.mode = modeNormal
			return m, nil

		case "s":
			m.mode = modeNormal
			return m.save()

		case "q":
			m.quitting = true
			return m, tea.Quit

		case "x":
			newM, cmd := m.save()
			if model, ok := newM.(Model); ok {
				model.quitting = true
				return model, tea.Sequence(cmd, tea.Quit)
			}
			return newM, cmd

		case "up", "k":
			if m.menuCursor > 0 {
				m.menuCursor--
			}

		case "down", "j":
			if m.menuCursor < len(m.menuItems)-1 {
				m.menuCursor++
			}
		}
	}

	return m, nil
}

func (m Model) updateEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.mode = modeNormal
			m.editInput = ""
			return m, nil

		case "enter":
			return m.saveEdit()

		case "backspace":
			if len(m.editInput) > 0 {
				m.editInput = m.editInput[:len(m.editInput)-1]
			}

		default:
			if key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))) {
				m.mode = modeNormal
				m.editInput = ""
				return m, nil
			}
			m.editInput += msg.String()
		}
	}

	return m, nil
}

func (m Model) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			m.mode = modeNormal
			if m.confirmAction != nil {
				return m, func() tea.Msg { return m.confirmAction() }
			}
			return m, nil

		case "n", "esc":
			m.mode = modeNormal
			m.confirmAction = nil
			return m, nil
		}
	}

	return m, nil
}

func (m Model) toggleItem() (tea.Model, tea.Cmd) {
	if len(m.list.Items()) == 0 {
		return m, nil
	}

	selected := m.list.SelectedItem()
	if selected == nil {
		return m, nil
	}

	item := selected.(todoItem)
	item.item.Done = !item.item.Done
	m.hasChanges = true

	// Update the item in the list
	m.list.SetItem(m.list.Index(), item)

	return m, nil
}

func (m Model) startEdit() (tea.Model, tea.Cmd) {
	if len(m.list.Items()) == 0 {
		return m, nil
	}

	selected := m.list.SelectedItem()
	if selected == nil {
		return m, nil
	}

	item := selected.(todoItem)
	m.editInput = item.item.Description
	m.mode = modeEdit

	return m, nil
}

func (m Model) saveEdit() (tea.Model, tea.Cmd) {
	if len(m.list.Items()) == 0 {
		m.mode = modeNormal
		return m, nil
	}

	selected := m.list.SelectedItem()
	if selected == nil {
		m.mode = modeNormal
		return m, nil
	}

	item := selected.(todoItem)
	item.item.Description = m.editInput
	m.hasChanges = true
	m.mode = modeNormal
	m.editInput = ""

	// Update the item in the list
	m.list.SetItem(m.list.Index(), item)

	return m, nil
}

func (m Model) confirmDelete() (tea.Model, tea.Cmd) {
	if len(m.list.Items()) == 0 {
		return m, nil
	}

	selected := m.list.SelectedItem()
	if selected == nil {
		return m, nil
	}

	item := selected.(todoItem)
	m.confirmMsg = fmt.Sprintf("Delete '%s'?", item.item.Description)
	m.confirmAction = func() tea.Msg {
		return m.deleteItem()
	}
	m.mode = modeConfirm

	return m, nil
}

func (m Model) deleteItem() tea.Msg {
	if len(m.list.Items()) == 0 {
		return nil
	}

	idx := m.list.Index()
	item := m.list.Items()[idx].(todoItem)

	// Remove from document items
	newItems := make([]*markdown.Item, 0)
	for _, it := range m.doc.Items {
		if it != item.item {
			newItems = append(newItems, it)
		}
	}
	m.doc.Items = newItems

	// Remove from sections
	for _, section := range m.doc.Sections {
		if section.Project == item.project {
			newTODOs := make([]*markdown.Item, 0)
			for _, it := range section.TODOItems {
				if it != item.item {
					newTODOs = append(newTODOs, it)
				}
			}
			section.TODOItems = newTODOs
		}
	}

	// Remove from list
	m.list.RemoveItem(idx)
	m.hasChanges = true

	return nil
}

func (m Model) startAdd() (tea.Model, tea.Cmd) {
	m.editInput = ""
	m.mode = modeEdit
	return m, nil
}

func (m Model) save() (tea.Model, tea.Cmd) {
	content := m.doc.String()
	err := os.WriteFile(m.todoPath, []byte(content), 0644)
	if err != nil {
		m.err = err
		return m, nil
	}

	m.hasChanges = false
	return m, func() tea.Msg {
		return saveMsg{err: nil}
	}
}
