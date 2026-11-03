package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/jvzantvoort/mktodo/internal/markdown"
)

// Colours adapt to light and dark terminals. Every state is also shown
// without colour (cursor marker, [✓], "!" for FIXME, [UNSAVED]), so the UI
// works with NO_COLOR or TERM=dumb.
var (
	accent = lipgloss.AdaptiveColor{Light: "#8839ef", Dark: "#cba6f7"}
	muted  = lipgloss.AdaptiveColor{Light: "#6c6f85", Dark: "#9399b2"}
	danger = lipgloss.AdaptiveColor{Light: "#d20f39", Dark: "#f38ba8"}

	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(accent)
	mutedStyle    = lipgloss.NewStyle().Foreground(muted)
	titleStyle    = lipgloss.NewStyle().Bold(true)
	focusStyle    = lipgloss.NewStyle().Bold(true).Foreground(accent)
	selectedStyle = lipgloss.NewStyle().Bold(true).Reverse(true)
	doneStyle     = lipgloss.NewStyle().Faint(true).Strikethrough(true)
	fixmeStyle    = lipgloss.NewStyle().Bold(true).Foreground(danger)
	errorStyle    = lipgloss.NewStyle().Foreground(danger)
	boxStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(1, 2)
)

const keyHints = "space toggle  e edit  d delete  a add  -/+ move  x cut  p paste  ? all keys"

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	switch m.mode {
	case modeMenu:
		return m.overlay(m.viewMenu())
	case modeEdit:
		return m.overlay(m.viewEdit())
	case modeConfirm:
		return m.overlay(m.confirmMsg + "\n\n" + mutedStyle.Render("y yes · n no"))
	case modeHelp:
		return m.overlay(viewHelp())
	case modeConflict:
		return m.overlay(m.viewConflict())
	default:
		return m.viewMain()
	}
}

func (m Model) viewMain() string {
	width, height := max(m.width, 40), max(m.height, 10)
	bodyHeight := height - 5 // header, two rules, hints, status
	leftWidth := min(max(width/3, 18), 36)
	rightWidth := width - leftWidth - 3

	left := m.projectLines(leftWidth, bodyHeight)
	right := m.itemLines(rightWidth, bodyHeight)

	var b strings.Builder
	b.WriteString(m.header(width) + "\n")
	b.WriteString(mutedStyle.Render(strings.Repeat("─", width)) + "\n")
	for i := 0; i < bodyHeight; i++ {
		b.WriteString(left[i] + mutedStyle.Render(" │ ") + right[i] + "\n")
	}
	b.WriteString(mutedStyle.Render(strings.Repeat("─", width)) + "\n")
	b.WriteString(mutedStyle.Render(fit(keyHints, width)) + "\n")
	b.WriteString(m.statusLine(width))
	return b.String()
}

func (m Model) header(width int) string {
	left := headerStyle.Render("mktodo " + m.version)
	right := mutedStyle.Render("? help · esc menu")
	gap := max(width-ansi.StringWidth(left)-ansi.StringWidth(right), 1)
	return left + strings.Repeat(" ", gap) + right
}

// paneTitle marks the focused pane with "»" as well as colour.
func paneTitle(title string, focused bool, width int) string {
	if focused {
		return focusStyle.Render(fit("» "+title, width))
	}
	return titleStyle.Render(fit("  "+title, width))
}

func (m Model) projectLines(width, height int) []string {
	rows := m.rows()
	focused := m.focus == paneProjects
	lines := []string{paneTitle("Projects", focused, width)}

	start := scrollStart(m.projectCursor, len(rows), height-1)
	for i := start; i < len(rows) && len(lines) < height; i++ {
		row := rows[i]
		icon, name := "•", "(no project)"
		if row.project != nil {
			name = row.project.Title
			if row.hasChildren {
				icon = "▾"
				if m.collapsed[row.project.Name] {
					icon = "▸"
				}
			}
		}
		text := fmt.Sprintf("%s%s %s (%d)", strings.Repeat("  ", row.depth), icon, name, row.count)
		lines = append(lines, renderRow(text, i == m.projectCursor, focused, lipgloss.NewStyle(), width))
	}
	return padLines(lines, width, height)
}

func (m Model) itemLines(width, height int) []string {
	title := "(no project)"
	if p := m.selectedProject(); p != nil {
		title = p.Title
	}
	focused := m.focus == paneItems
	lines := []string{paneTitle(title, focused, width)}

	items := m.items()
	if len(items) == 0 {
		lines = append(lines, mutedStyle.Render(fit("  No items. Press a to add one.", width)))
	}
	start := scrollStart(m.itemCursor, len(items), height-1)
	for i := start; i < len(items) && len(lines) < height; i++ {
		lines = append(lines, renderRow(itemText(items[i]), i == m.itemCursor, focused, itemStyle(items[i]), width))
	}
	return padLines(lines, width, height)
}

func itemText(item *markdown.Item) string {
	box := "[ ]"
	if item.Done {
		box = "[✓]"
	}
	flag := " "
	if item.IsFIXME {
		flag = "!"
	}
	return fmt.Sprintf("%s %s %s", box, flag, item.Description)
}

func itemStyle(item *markdown.Item) lipgloss.Style {
	switch {
	case item.Done:
		return doneStyle
	case item.IsFIXME:
		return fixmeStyle
	default:
		return lipgloss.NewStyle()
	}
}

// renderRow draws one list row. The selected row has a "›" marker; in the
// focused pane it's also shown in reverse video.
func renderRow(text string, selected, focused bool, style lipgloss.Style, width int) string {
	marker := "  "
	if selected {
		marker = "› "
	}
	line := fit(marker+text, width)
	if selected && focused {
		return selectedStyle.Render(line)
	}
	return style.Render(line)
}

func (m Model) statusLine(width int) string {
	var done, fixme int
	for _, item := range m.doc.Items {
		if item.Done {
			done++
		}
		if item.IsFIXME {
			fixme++
		}
	}
	total := len(m.doc.Items)
	parts := []string{fmt.Sprintf("Total %d · Done %d · Open %d · FIXME %d", total, done, total-done, fixme)}
	if m.hasChanges {
		parts = append(parts, "[UNSAVED]")
	}
	if m.clipboard != nil {
		parts = append(parts, "Clipboard: "+m.clipboard.Description)
	}
	line := fit(strings.Join(parts, " · "), width)

	switch {
	case m.err != nil:
		return errorStyle.Render(fit("Error: "+m.err.Error(), width))
	case m.status != "":
		return fit(m.status, width)
	default:
		return mutedStyle.Render(line)
	}
}

func (m Model) viewMenu() string {
	return titleStyle.Render("Menu") + "\n\n" +
		"s    Save changes\n" +
		"x    Save and quit\n" +
		"q    Quit\n" +
		"esc  Close menu"
}

func (m Model) viewEdit() string {
	title := "Edit item"
	if m.adding {
		title = "Add item"
		if p := m.selectedProject(); p != nil {
			title += " to " + p.Title
		}
	}
	body := titleStyle.Render(title) + "\n\n" + m.input.View() + "\n\n"
	if m.err != nil {
		body += errorStyle.Render("Error: "+m.err.Error()) + "\n\n"
	}
	return body + mutedStyle.Render("enter save · esc cancel")
}

func (m Model) viewConflict() string {
	return titleStyle.Render("File changed on disk") + "\n\n" +
		m.todoPath + "\nwas changed outside mktodo since it was opened.\n\n" +
		"r    reload it (your unsaved changes are lost)\n" +
		"o    overwrite it with your version\n" +
		"esc  cancel, don't save"
}

func viewHelp() string {
	return titleStyle.Render("Keys") + `

↑/↓ k/j  move              space  toggle done
←/→ h/l  projects / items  e      edit
tab      switch pane       d      delete
enter    expand / edit     a      add to project
s        save              - +    move item
esc      menu              x p    cut / paste
q        quit              ?      this help

` + mutedStyle.Render("press any key to close")
}

// overlay centres a boxed dialog on the screen, clipped to the terminal.
func (m Model) overlay(content string) string {
	width, height := max(m.width, 40), max(m.height, 10)
	placed := lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, boxStyle.Render(content))
	lines := strings.Split(placed, "\n")
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, width, "")
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}

// scrollStart returns the first row to show so the cursor stays visible.
func scrollStart(cursor, total, visible int) int {
	if visible <= 0 || total <= visible {
		return 0
	}
	return min(max(cursor-visible+1, 0), total-visible)
}

// fit truncates s to width cells and pads it with spaces to exactly width.
func fit(s string, width int) string {
	s = ansi.Truncate(s, width, "…")
	return s + strings.Repeat(" ", max(width-ansi.StringWidth(s), 0))
}

func padLines(lines []string, width, height int) []string {
	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", width))
	}
	return lines[:height]
}
