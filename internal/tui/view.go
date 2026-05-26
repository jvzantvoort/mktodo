package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginLeft(2)

	paginationStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1)

	menuStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2).
			Margin(1, 2)

	editStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Padding(1, 2).
			Width(60)

	confirmStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("196")).
			Padding(1, 2).
			Margin(1, 2)
)

func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var content string

	switch m.mode {
	case modeMenu:
		content = m.viewMenu()
	case modeEdit:
		content = m.viewEdit()
	case modeConfirm:
		content = m.viewConfirm()
	default:
		content = m.viewNormal()
	}

	return content
}

func (m Model) viewNormal() string {
	var b strings.Builder

	// Main list
	b.WriteString(m.list.View())
	b.WriteString("\n")

	// Status bar
	status := m.buildStatusBar()
	b.WriteString(statusBarStyle.Render(status))
	b.WriteString("\n")

	// Help text
	help := "space:toggle • e:edit • a:add • d:del • s:save • -/+:move • x:cut • p:paste • esc:menu"
	b.WriteString(helpStyle.Render(help))

	// Error message if any
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render(fmt.Sprintf("Error: %v", m.err)))
	}

	return b.String()
}

func (m Model) viewMenu() string {
	var b strings.Builder

	b.WriteString(menuStyle.Render(
		titleStyle.Render("Menu") + "\n\n" +
			strings.Join(m.menuItems, "\n"),
	))

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		b.String(),
	)
}

func (m Model) viewEdit() string {
	var b strings.Builder

	title := "Edit TODO:"
	if m.isAdding {
		title = "New TODO:"
	}
	b.WriteString(title + "\n\n")
	b.WriteString(m.editInput + "█\n\n")
	b.WriteString(helpStyle.Render("enter:save • esc:cancel"))

	content := editStyle.Render(b.String())

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (m Model) viewConfirm() string {
	var b strings.Builder

	b.WriteString(m.confirmMsg + "\n\n")
	b.WriteString(helpStyle.Render("y:yes • n:no • esc:cancel"))

	content := confirmStyle.Render(b.String())

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (m Model) buildStatusBar() string {
	total := len(m.list.Items())
	done := 0
	fixme := 0

	for _, item := range m.list.Items() {
		ti := item.(todoItem)
		if ti.item.Done {
			done++
		}
		if ti.item.IsFIXME {
			fixme++
		}
	}

	status := fmt.Sprintf("Total: %d | Done: %d | Open: %d", total, done, total-done)
	if fixme > 0 {
		status += fmt.Sprintf(" | FIXME: %d", fixme)
	}

	if m.hasChanges {
		status += " | [UNSAVED]"
	}
	if m.clipboard != nil {
		status += " | [CLIPBOARD]"
	}

	return status
}
