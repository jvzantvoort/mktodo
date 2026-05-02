package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/internal/project"
)

func NewModel(cfg *config.Config, doc *markdown.Document, todoPath string) Model {
	projectMap, _ := project.BuildHierarchy(cfg)
	projects := make([]*project.Project, 0, len(projectMap))
	for _, p := range projectMap {
		projects = append(projects, p)
	}

	items := buildItemList(doc)

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "TODO Items"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return Model{
		config:   cfg,
		doc:      doc,
		projects: projects,
		list:     l,
		items:    items,
		mode:     modeNormal,
		todoPath: todoPath,
		menuItems: []string{
			"s - Save changes",
			"q - Quit (lose changes)",
			"x - Save and quit",
			"Esc - Cancel",
		},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func buildItemList(doc *markdown.Document) []list.Item {
	items := make([]list.Item, 0)

	for _, item := range doc.Items {
		items = append(items, todoItem{
			item:    item,
			project: item.Project,
		})
	}

	return items
}
