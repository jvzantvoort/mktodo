package tui

import (
	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/internal/project"
)

// projectRow is one visible line of the project tree. A row with a nil
// project collects the items that aren't under any configured project.
type projectRow struct {
	project     *project.Project
	depth       int
	hasChildren bool
	count       int // items in this project and its children
}

// rows returns the visible project tree: projects in configuration order,
// children under their parent unless the parent is collapsed.
func (m Model) rows() []projectRow {
	var rows []projectRow
	var walk func(p *project.Project, depth int)
	walk = func(p *project.Project, depth int) {
		rows = append(rows, projectRow{
			project:     p,
			depth:       depth,
			hasChildren: len(p.Children) > 0,
			count:       m.subtreeCount(p),
		})
		if m.collapsed[p.Name] {
			return
		}
		for _, child := range p.Children {
			walk(child, depth+1)
		}
	}
	for _, root := range project.GetRootProjects(m.doc.Projects) {
		walk(root, 0)
	}

	if n := len(m.unassignedItems()); n > 0 {
		rows = append(rows, projectRow{count: n})
	}
	return rows
}

func (m Model) subtreeCount(p *project.Project) int {
	n := len(m.doc.GetItemsByProject(p))
	for _, child := range p.Children {
		n += m.subtreeCount(child)
	}
	return n
}

func (m Model) unassignedItems() []*markdown.Item {
	var items []*markdown.Item
	for _, item := range m.doc.Items {
		if item.Project == nil {
			items = append(items, item)
		}
	}
	return items
}

// selectedRow returns the project row under the cursor.
func (m Model) selectedRow() (projectRow, bool) {
	rows := m.rows()
	if m.projectCursor < 0 || m.projectCursor >= len(rows) {
		return projectRow{}, false
	}
	return rows[m.projectCursor], true
}

// selectedProject returns the project under the cursor, or nil for the
// "no project" row or an empty tree.
func (m Model) selectedProject() *project.Project {
	row, _ := m.selectedRow()
	return row.project
}

// items returns the items shown in the right pane, in file order.
func (m Model) items() []*markdown.Item {
	row, ok := m.selectedRow()
	if !ok {
		return nil
	}
	if row.project == nil {
		return m.unassignedItems()
	}
	return m.doc.GetItemsByProject(row.project)
}

// selectedItem returns the item under the cursor in the right pane.
func (m Model) selectedItem() (*markdown.Item, bool) {
	items := m.items()
	if m.itemCursor < 0 || m.itemCursor >= len(items) {
		return nil, false
	}
	return items[m.itemCursor], true
}

// clamp keeps both cursors inside their lists after the lists change.
func (m *Model) clamp() {
	m.projectCursor = clampIndex(m.projectCursor, len(m.rows()))
	m.itemCursor = clampIndex(m.itemCursor, len(m.items()))
}

func clampIndex(i, n int) int {
	if i >= n {
		i = n - 1
	}
	if i < 0 {
		i = 0
	}
	return i
}

// selectItem moves the item cursor to item, if it's in the right pane.
func (m *Model) selectItem(item *markdown.Item) {
	for i, it := range m.items() {
		if it == item {
			m.itemCursor = i
			return
		}
	}
}
