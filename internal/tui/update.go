package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/internal/project"
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

		case "-":
			return m.moveItemUp()

		case "+":
			return m.moveItemDown()

		case "x":
			return m.cutItem()

		case "p":
			return m.pasteItem()
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
			if m.hasChanges {
				m.confirmMsg = "Discard unsaved changes and quit? (y/n)"
				m.confirmQuit = true
				m.mode = modeConfirm
				return m, nil
			}
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
		case "esc", "ctrl+c":
			m.mode = modeNormal
			m.editInput = ""
			m.isAdding = false
			return m, nil

		case "enter":
			return m.saveEdit()

		case "backspace":
			if len(m.editInput) > 0 {
				m.editInput = m.editInput[:len(m.editInput)-1]
			}

		default:
			if len(msg.String()) == 1 {
				m.editInput += msg.String()
			}
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
			if m.confirmQuit {
				m.confirmQuit = false
				m.quitting = true
				return m, tea.Quit
			}
			return m.executeDelete()

		case "n", "esc":
			m.mode = modeNormal
			m.confirmQuit = false
			m.confirmIdx = -1
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

	// Sync the change to section.Lines so save writes the correct state
	if err := m.doc.UpdateItem(item.item); err != nil {
		m.err = err
		return m, nil
	}

	m.hasChanges = true
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
	m.isAdding = false
	m.mode = modeEdit
	return m, nil
}

func (m Model) startAdd() (tea.Model, tea.Cmd) {
	m.editInput = ""
	m.isAdding = true
	m.mode = modeEdit
	return m, nil
}

func (m Model) saveEdit() (tea.Model, tea.Cmd) {
	if m.editInput == "" {
		m.mode = modeNormal
		m.isAdding = false
		return m, nil
	}

	if m.isAdding {
		// Determine target project: use selected item's project, or "default", or first available
		var targetProj *project.Project
		if selected := m.list.SelectedItem(); selected != nil {
			ti := selected.(todoItem)
			if ti.project != nil {
				targetProj = ti.project
			}
		}
		if targetProj == nil {
			if p, ok := m.doc.Projects["default"]; ok {
				targetProj = p
			}
		}
		if targetProj == nil {
			for _, p := range m.doc.Projects {
				targetProj = p
				break
			}
		}
		if targetProj == nil {
			m.err = fmt.Errorf("no project available")
			m.mode = modeNormal
			m.editInput = ""
			m.isAdding = false
			return m, nil
		}

		_, err := m.doc.AddItem(targetProj, m.editInput)
		if err != nil {
			m.err = err
			m.mode = modeNormal
			m.editInput = ""
			m.isAdding = false
			return m, nil
		}

		// Rebuild list to include the new item
		newItems := buildItemList(m.doc)
		m.items = newItems
		m.list.SetItems(newItems)
		m.hasChanges = true
	} else {
		// Edit existing item
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

		// Sync to section.Lines so save writes the correct text
		if err := m.doc.UpdateItem(item.item); err != nil {
			m.err = err
			m.mode = modeNormal
			m.editInput = ""
			return m, nil
		}

		m.hasChanges = true
		m.list.SetItem(m.list.Index(), item)
	}

	m.mode = modeNormal
	m.editInput = ""
	m.isAdding = false
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

	ti := selected.(todoItem)
	m.confirmIdx = m.list.Index()
	m.confirmQuit = false
	m.confirmMsg = fmt.Sprintf("Delete '%s'? (y/n)", ti.item.Description)
	m.mode = modeConfirm
	return m, nil
}

// executeDelete is called from updateConfirm when the user presses "y".
// It operates on the current model state (no stale closure captures).
func (m Model) executeDelete() (tea.Model, tea.Cmd) {
	idx := m.confirmIdx
	if idx < 0 || idx >= len(m.list.Items()) {
		return m, nil
	}

	ti := m.list.Items()[idx].(todoItem)

	if err := m.doc.RemoveItem(ti.item); err != nil {
		m.err = err
		return m, nil
	}

	m.list.RemoveItem(idx)
	m.hasChanges = true
	m.confirmIdx = -1
	return m, nil
}

func (m Model) moveItemUp() (tea.Model, tea.Cmd) {
	if len(m.list.Items()) == 0 {
		return m, nil
	}

	selected := m.list.SelectedItem()
	if selected == nil {
		return m, nil
	}

	ti := selected.(todoItem)
	section, sectionIdx := findItemSection(m.doc, ti.item)
	if section == nil || sectionIdx <= 0 {
		return m, nil
	}

	// Swap within the section
	section.TODOItems[sectionIdx], section.TODOItems[sectionIdx-1] = section.TODOItems[sectionIdx-1], section.TODOItems[sectionIdx]
	section.Lines[sectionIdx], section.Lines[sectionIdx-1] = section.Lines[sectionIdx-1], section.Lines[sectionIdx]

	// Rebuild doc.Items and list
	m.doc.Items = rebuildDocItems(m.doc)
	newItems := buildItemList(m.doc)
	m.items = newItems
	m.list.SetItems(newItems)

	listIdx := m.list.Index()
	if listIdx > 0 {
		m.list.Select(listIdx - 1)
	}

	m.hasChanges = true
	return m, nil
}

func (m Model) moveItemDown() (tea.Model, tea.Cmd) {
	if len(m.list.Items()) == 0 {
		return m, nil
	}

	selected := m.list.SelectedItem()
	if selected == nil {
		return m, nil
	}

	ti := selected.(todoItem)
	section, sectionIdx := findItemSection(m.doc, ti.item)
	if section == nil || sectionIdx >= len(section.TODOItems)-1 {
		return m, nil
	}

	// Swap within the section
	section.TODOItems[sectionIdx], section.TODOItems[sectionIdx+1] = section.TODOItems[sectionIdx+1], section.TODOItems[sectionIdx]
	section.Lines[sectionIdx], section.Lines[sectionIdx+1] = section.Lines[sectionIdx+1], section.Lines[sectionIdx]

	// Rebuild doc.Items and list
	m.doc.Items = rebuildDocItems(m.doc)
	newItems := buildItemList(m.doc)
	m.items = newItems
	m.list.SetItems(newItems)

	listIdx := m.list.Index()
	if listIdx < len(newItems)-1 {
		m.list.Select(listIdx + 1)
	}

	m.hasChanges = true
	return m, nil
}

func (m Model) cutItem() (tea.Model, tea.Cmd) {
	if len(m.list.Items()) == 0 {
		return m, nil
	}

	selected := m.list.SelectedItem()
	if selected == nil {
		return m, nil
	}

	ti := selected.(todoItem)
	m.clipboard = ti.item
	m.clipboardProject = ti.project

	idx := m.list.Index()
	if err := m.doc.RemoveItem(ti.item); err != nil {
		m.err = err
		m.clipboard = nil
		m.clipboardProject = nil
		return m, nil
	}

	m.list.RemoveItem(idx)
	m.hasChanges = true
	return m, nil
}

func (m Model) pasteItem() (tea.Model, tea.Cmd) {
	if m.clipboard == nil {
		return m, nil
	}

	// Paste to the selected item's project, or the clipboard item's original project
	targetProj := m.clipboardProject
	if selected := m.list.SelectedItem(); selected != nil {
		ti := selected.(todoItem)
		if ti.project != nil {
			targetProj = ti.project
		}
	}
	if targetProj == nil {
		return m, nil
	}

	item, err := m.doc.AddItem(targetProj, m.clipboard.Description)
	if err != nil {
		m.err = err
		return m, nil
	}
	// Preserve done status from clipboard
	item.Done = m.clipboard.Done
	if err := m.doc.UpdateItem(item); err != nil {
		m.err = err
		return m, nil
	}

	m.clipboard = nil
	m.clipboardProject = nil

	// Rebuild list to include pasted item
	newItems := buildItemList(m.doc)
	m.items = newItems
	m.list.SetItems(newItems)
	m.hasChanges = true
	return m, nil
}

func (m Model) save() (tea.Model, tea.Cmd) {
	if err := markdown.SaveDocument(m.todoPath, m.doc); err != nil {
		m.err = err
		return m, nil
	}

	m.hasChanges = false
	m.err = nil
	return m, func() tea.Msg {
		return saveMsg{err: nil}
	}
}

// findItemSection returns the SectionTODO containing the item and the item's index within it.
func findItemSection(doc *markdown.Document, item *markdown.Item) (*markdown.Section, int) {
	for _, section := range doc.Sections {
		if section.Type != markdown.SectionTODO {
			continue
		}
		for i, it := range section.TODOItems {
			if it == item {
				return section, i
			}
		}
	}
	return nil, -1
}

// rebuildDocItems reconstructs doc.Items from all sections in document order.
func rebuildDocItems(doc *markdown.Document) []*markdown.Item {
	var items []*markdown.Item
	for _, section := range doc.Sections {
		items = append(items, section.TODOItems...)
	}
	return items
}
