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

func TestTUI_PastedLineBreaksNeverReachTheFile(t *testing.T) {
	m, path := newTestModel(t, "# TODO\n\n- [ ] one\n")
	// A paste can deliver several lines at once.
	paste := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("evil\n# Injected\n- [x] fake"), Paste: true}
	m = press(m, "a")
	m, _ = m.Update(paste)
	m = press(m, "enter")
	if model := asModel(t, m); model.mode == modeEdit {
		m = press(m, "esc") // refused: the dialog stayed open with an error
	}
	press(m, "s")

	for _, line := range strings.Split(readFile(t, path), "\n") {
		if line == "# Injected" || strings.HasPrefix(line, "- [x] fake") {
			t.Fatalf("pasted text added a line to the file: %q", readFile(t, path))
		}
	}
}

func TestTUI_ControlCharactersNeverReachTheFile(t *testing.T) {
	m, path := newTestModel(t, "# TODO\n\n- [ ] one\n")
	model := asModel(t, press(m, "e"))
	model.input.SetValue("a\x07b\x1b[31m")
	press(model, "enter", "s")

	if got := readFile(t, path); strings.ContainsAny(got, "\x07\x1b") {
		t.Errorf("control characters were written: %q", got)
	}
}

func TestTUI_ExternalChange(t *testing.T) {
	const external = "# TODO\n\n- [ ] one\n- [ ] added elsewhere\n"
	tests := []struct {
		name     string
		keys     []string // after the conflict dialog opens
		want     string   // file content afterwards
		wantQuit bool
	}{
		{"overwrite", []string{"o"}, "# TODO\n\n- [X] one\n", false},
		{"reload", []string{"r"}, external, false},
		{"cancel", []string{"esc"}, external, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, path := newTestModel(t, "# TODO\n\n- [ ] one\n")
			m = press(m, "space")
			if err := os.WriteFile(path, []byte(external), 0o600); err != nil {
				t.Fatal(err)
			}

			model := asModel(t, press(m, "s"))
			if model.mode != modeConflict || !strings.Contains(model.View(), "changed outside mktodo") {
				t.Fatal("saving over an outside change should open the conflict dialog")
			}
			model = asModel(t, press(model, tt.keys...))

			if got := readFile(t, path); got != tt.want {
				t.Errorf("file = %q, want %q", got, tt.want)
			}
			if tt.name == "reload" {
				if model.hasChanges || len(model.items()) != 2 {
					t.Errorf("after reload: hasChanges %v, %d items; want the file's 2 items", model.hasChanges, len(model.items()))
				}
			}
			if tt.name == "cancel" && !model.hasChanges {
				t.Error("cancel should keep the unsaved changes")
			}
		})
	}
}

func TestTUI_SaveAndQuitWithExternalChange(t *testing.T) {
	m, path := newTestModel(t, "# TODO\n\n- [ ] one\n")
	m = press(m, "space")
	if err := os.WriteFile(path, []byte("changed\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	model := asModel(t, press(m, "esc"))
	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if asModel(t, next).mode != modeConflict || cmd != nil {
		t.Fatal("save-and-quit over an outside change should ask, not quit")
	}
	next, cmd = next.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	if !asModel(t, next).quitting || cmd == nil {
		t.Error("overwriting after save-and-quit should quit")
	}
	if got := readFile(t, path); got != "# TODO\n\n- [X] one\n" {
		t.Errorf("file = %q", got)
	}
}
