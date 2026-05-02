package tui

import (
	"testing"

	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/internal/project"
	"github.com/jvzantvoort/mktodo/internal/todo"
)

func TestNewModel(t *testing.T) {
	cfg := &config.Config{
		TodoFile: "TODO.md",
		Projects: []config.ProjectConfig{
			{Name: "default", Title: "TODO"},
		},
	}

	projects, _ := project.BuildHierarchy(cfg)
	item1 := &todo.Item{Description: "Test 1", Done: false}
	item2 := &todo.Item{Description: "Test 2", Done: true}

	doc := &markdown.Document{
		Path:     "TODO.md",
		Sections: []*markdown.Section{},
		Projects: projects,
		Items:    []*markdown.Item{item1, item2},
	}

	model := NewModel(cfg, doc, "TODO.md")

	if model.config == nil {
		t.Error("config should not be nil")
	}

	if model.doc == nil {
		t.Error("doc should not be nil")
	}

	if len(model.items) != 2 {
		t.Errorf("expected 2 items, got %d", len(model.items))
	}

	if model.mode != modeNormal {
		t.Errorf("expected mode modeNormal, got %v", model.mode)
	}
}

func TestTodoItem(t *testing.T) {
	proj := &project.Project{
		Name:  "test",
		Title: "Test Project",
		Level: 1,
	}

	item := &todo.Item{
		Description: "Test item",
		Done:        false,
		IsFIXME:     false,
	}

	ti := todoItem{
		item:    item,
		project: proj,
	}

	title := ti.Title()
	if title != "[ ] Test item" {
		t.Errorf("expected '[ ] Test item', got '%s'", title)
	}

	desc := ti.Description()
	if desc != "Project: Test Project" {
		t.Errorf("expected 'Project: Test Project', got '%s'", desc)
	}

	filter := ti.FilterValue()
	if filter != "Test item" {
		t.Errorf("expected 'Test item', got '%s'", filter)
	}
}

func TestTodoItem_Done(t *testing.T) {
	proj := &project.Project{
		Name:  "test",
		Title: "Test Project",
	}

	item := &todo.Item{
		Description: "Done item",
		Done:        true,
	}

	ti := todoItem{
		item:    item,
		project: proj,
	}

	title := ti.Title()
	if title != "[✓] Done item" {
		t.Errorf("expected '[✓] Done item', got '%s'", title)
	}
}

func TestTodoItem_FIXME(t *testing.T) {
	proj := &project.Project{
		Name:  "test",
		Title: "Test Project",
	}

	item := &todo.Item{
		Description: "Fix this",
		Done:        false,
		IsFIXME:     true,
	}

	ti := todoItem{
		item:    item,
		project: proj,
	}

	title := ti.Title()
	if title != "🔴 FIXME: [ ] Fix this" {
		t.Errorf("expected '🔴 FIXME: [ ] Fix this', got '%s'", title)
	}
}
