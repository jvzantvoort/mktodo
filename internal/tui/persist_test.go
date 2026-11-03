package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/markdown"
)

// defaultConfig has a single project "default" titled "TODO".
var defaultConfig = &config.Config{TodoFile: "TODO.md", Projects: []config.ProjectConfig{{Name: "default", Title: "TODO"}}}

// newTestModel writes content to a temp TODO file and returns a model for it.
func newTestModel(t *testing.T, content string) (tea.Model, string) {
	t.Helper()
	return newTestModelWith(t, content, defaultConfig)
}

// newTestModelWith is newTestModel with a custom configuration.
func newTestModelWith(t *testing.T, content string, cfg *config.Config) (tea.Model, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "TODO.md")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	doc, err := markdown.LoadDocument(path, cfg)
	if err != nil {
		t.Fatal(err)
	}
	m, err := NewModel(cfg, doc, path)
	if err != nil {
		t.Fatal(err)
	}
	var model tea.Model = m
	model, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return model, path
}

// press sends keys to the model: "space", "enter", "esc", "backspace",
// arrow keys ("up", "down", "left", "right"), "tab", or literal text typed
// one rune at a time.
func press(m tea.Model, keys ...string) tea.Model {
	for _, k := range keys {
		var msgs []tea.KeyMsg
		switch k {
		case "space":
			msgs = []tea.KeyMsg{{Type: tea.KeySpace, Runes: []rune{' '}}}
		case "enter":
			msgs = []tea.KeyMsg{{Type: tea.KeyEnter}}
		case "esc":
			msgs = []tea.KeyMsg{{Type: tea.KeyEsc}}
		case "backspace":
			msgs = []tea.KeyMsg{{Type: tea.KeyBackspace}}
		case "down":
			msgs = []tea.KeyMsg{{Type: tea.KeyDown}}
		case "up":
			msgs = []tea.KeyMsg{{Type: tea.KeyUp}}
		case "left":
			msgs = []tea.KeyMsg{{Type: tea.KeyLeft}}
		case "right":
			msgs = []tea.KeyMsg{{Type: tea.KeyRight}}
		case "tab":
			msgs = []tea.KeyMsg{{Type: tea.KeyTab}}
		default:
			for _, r := range k {
				msgs = append(msgs, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			}
		}
		for _, msg := range msgs {
			m, _ = m.Update(msg)
		}
	}
	return m
}

// asModel converts a tea.Model back to Model.
func asModel(t *testing.T, m tea.Model) Model {
	t.Helper()
	model, ok := m.(Model)
	if !ok {
		t.Fatalf("model has type %T", m)
	}
	return model
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestTUI_ToggleIsSaved(t *testing.T) {
	m, path := newTestModel(t, "# TODO\n\n- [ ] one\n- [ ] two\n")
	press(m, "space", "s")
	if got, want := readFile(t, path), "# TODO\n\n- [X] one\n- [ ] two\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTUI_EditIsSaved(t *testing.T) {
	m, path := newTestModel(t, "# TODO\n\n- [ ] one\n")
	// Replace "one" with "über" to check multi-byte input and backspace.
	press(m, "e", "backspace", "backspace", "backspace", "über", "enter", "s")
	if got, want := readFile(t, path), "# TODO\n\n- [ ] über\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTUI_AddCreatesNewItem(t *testing.T) {
	m, path := newTestModel(t, "# TODO\n\n- [ ] one\n")
	press(m, "a", "new task", "enter", "s")
	if got, want := readFile(t, path), "# TODO\n\n- [ ] one\n- [ ] new task\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTUI_AddToEmptyFileCreatesSection(t *testing.T) {
	m, path := newTestModel(t, "")
	press(m, "a", "first", "enter", "s")
	if got, want := readFile(t, path), "# TODO\n\n- [ ] first\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTUI_DeleteIsSaved(t *testing.T) {
	m, path := newTestModel(t, "# TODO\n\n- [ ] one\n- [ ] two\n")
	press(m, "down", "d", "y", "s")
	if got, want := readFile(t, path), "# TODO\n\n- [ ] one\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTUI_CancelledEditChangesNothing(t *testing.T) {
	content := "# TODO\n\n- [ ] one\n"
	m, path := newTestModel(t, content)
	m = press(m, "e", "xyz", "esc")
	if asModel(t, m).hasChanges {
		t.Error("cancelled edit marked the model as changed")
	}
	press(m, "s")
	if got := readFile(t, path); got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestTUI_SaveKeepsPermissions(t *testing.T) {
	m, path := newTestModel(t, "# TODO\n\n- [ ] one\n")
	press(m, "space", "s")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 && os.PathSeparator == '/' {
		t.Errorf("permissions = %o, want 600", perm)
	}
}

func TestTUI_SaveAndQuitStaysOpenOnError(t *testing.T) {
	m, path := newTestModel(t, "# TODO\n\n- [ ] one\n")
	model := asModel(t, press(m, "space"))
	// Point the model at a path that can't be written: its parent is a file.
	model.todoPath = filepath.Join(path, "not-a-dir", "TODO.md")

	next, cmd := press(model, "esc").Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	got := asModel(t, next)
	if got.quitting || cmd != nil {
		t.Error("save-and-quit quit even though the save failed")
	}
	if got.err == nil {
		t.Error("save error was not reported")
	}
	if !got.hasChanges {
		t.Error("unsaved changes were dropped")
	}
}

func TestTUI_ItemsOutsideProjectsAreShown(t *testing.T) {
	m, _ := newTestModel(t, "# Notes\n\n- [ ] orphan\n")
	model := asModel(t, m)

	rows := model.rows()
	last := rows[len(rows)-1]
	if last.project != nil || last.count != 1 {
		t.Fatalf("last row = %+v, want the (no project) row with 1 item", last)
	}

	// Select the (no project) row: its item is listed, but adding is refused.
	model = asModel(t, press(model, "left", "down", "down"))
	if items := model.items(); len(items) != 1 || items[0].Description != "orphan" {
		t.Errorf("items = %v, want [orphan]", items)
	}
	model = asModel(t, press(model, "a"))
	if model.mode != modeNormal || model.err == nil {
		t.Error("adding to the (no project) row should be refused with an error")
	}
	if !strings.Contains(model.View(), "(no project) (1)") {
		t.Error("view doesn't show the (no project) row")
	}
}
