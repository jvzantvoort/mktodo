package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/jvzantvoort/mktodo/internal/config"
)

func strPtr(s string) *string { return &s }

// nestedConfig has TODO, and Lego with children Technic and City.
var nestedConfig = &config.Config{TodoFile: "TODO.md", Projects: []config.ProjectConfig{
	{Name: "default", Title: "TODO", Parent: strPtr("nil")},
	{Name: "lego", Title: "Lego", Parent: strPtr("nil")},
	{Name: "technic", Title: "Technic", Parent: strPtr("lego")},
	{Name: "city", Title: "City", Parent: strPtr("lego")},
}}

const nestedDoc = `# TODO

- [ ] one
- [ ] two
- [ ] three

# Lego

## Technic

- [ ] crane

## City

- [ ] station
`

func TestNewModel(t *testing.T) {
	m, _ := newTestModel(t, "# TODO\n\n- [ ] a\n- [x] b\n")
	model := asModel(t, m)
	if model.focus != paneItems || model.mode != modeNormal {
		t.Errorf("focus = %v, mode = %v; want items pane in normal mode", model.focus, model.mode)
	}
	if n := len(model.items()); n != 2 {
		t.Errorf("items() has %d items, want 2", n)
	}
}

func TestTree_RowsAndCounts(t *testing.T) {
	m, _ := newTestModelWith(t, nestedDoc, nestedConfig)
	rows := asModel(t, m).rows()

	want := []struct {
		title string
		depth int
		count int
	}{{"TODO", 0, 3}, {"Lego", 0, 2}, {"Technic", 1, 1}, {"City", 1, 1}}
	if len(rows) != len(want) {
		t.Fatalf("got %d rows, want %d", len(rows), len(want))
	}
	for i, w := range want {
		r := rows[i]
		if r.project.Title != w.title || r.depth != w.depth || r.count != w.count {
			t.Errorf("row %d = %s depth %d count %d, want %s depth %d count %d",
				i, r.project.Title, r.depth, r.count, w.title, w.depth, w.count)
		}
	}
}

func TestTree_NavigateAndCollapse(t *testing.T) {
	m, _ := newTestModelWith(t, nestedDoc, nestedConfig)

	// Move to Technic in the projects pane: its items are shown.
	model := asModel(t, press(m, "left", "down", "down"))
	if p := model.selectedProject(); p == nil || p.Name != "technic" {
		t.Fatalf("selected project = %v, want technic", p)
	}
	if items := model.items(); len(items) != 1 || items[0].Description != "crane" {
		t.Errorf("items = %v, want [crane]", items)
	}

	// Collapse Lego: Technic and City are hidden.
	model = asModel(t, press(model, "up", "enter"))
	if n := len(model.rows()); n != 2 {
		t.Errorf("after collapsing Lego, %d rows are visible, want 2", n)
	}
	if !strings.Contains(model.View(), "▸ Lego (2)") {
		t.Error("collapsed project should show ▸")
	}

	// Enter again expands it.
	model = asModel(t, press(model, "enter"))
	if n := len(model.rows()); n != 4 {
		t.Errorf("after expanding Lego, %d rows are visible, want 4", n)
	}
}

func TestTree_EnterOnLeafFocusesItems(t *testing.T) {
	m, _ := newTestModelWith(t, nestedDoc, nestedConfig)
	model := asModel(t, press(m, "left", "enter"))
	if model.focus != paneItems {
		t.Error("enter on a project without children should focus its items")
	}
}

func TestTUI_MoveItem(t *testing.T) {
	m, path := newTestModel(t, "# TODO\n\n- [ ] one\n- [ ] two\n- [ ] three\n")

	// "+" moves "one" down twice, "-" moves it back up once.
	model := asModel(t, press(m, "+", "+", "-"))
	if item, _ := model.selectedItem(); item.Description != "one" {
		t.Errorf("cursor should follow the moved item, it's on %q", item.Description)
	}
	press(model, "s")
	if got, want := readFile(t, path), "# TODO\n\n- [ ] two\n- [ ] one\n- [ ] three\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTUI_MoveItemAtEdgeDoesNothing(t *testing.T) {
	m, _ := newTestModel(t, "# TODO\n\n- [ ] one\n- [ ] two\n")
	if asModel(t, press(m, "-")).hasChanges {
		t.Error("moving the first item up should change nothing")
	}
}

func TestTUI_CutAndPasteToOtherProject(t *testing.T) {
	m, path := newTestModelWith(t, nestedDoc, nestedConfig)

	// Cut "two" from TODO, select City, paste after "station".
	model := asModel(t, press(m, "down", "x"))
	if model.clipboard == nil || model.clipboard.Description != "two" {
		t.Fatalf("clipboard = %v, want 'two'", model.clipboard)
	}
	model = asModel(t, press(model, "left", "down", "down", "down", "right", "p"))
	if model.clipboard != nil {
		t.Error("clipboard should be empty after paste")
	}
	if item, _ := model.selectedItem(); item.Description != "two" {
		t.Errorf("pasted item should be selected, got %q", item.Description)
	}

	press(model, "s")
	want := "# TODO\n\n- [ ] one\n- [ ] three\n\n# Lego\n\n## Technic\n\n- [ ] crane\n\n## City\n\n- [ ] station\n- [ ] two\n"
	if got := readFile(t, path); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestTUI_SecondCutIsRefused(t *testing.T) {
	m, _ := newTestModel(t, "# TODO\n\n- [ ] one\n- [ ] two\n")
	model := asModel(t, press(m, "x", "x"))
	if model.err == nil || model.clipboard.Description != "one" {
		t.Error("a second cut should be refused so the first item isn't lost")
	}
	if n := len(model.items()); n != 1 {
		t.Errorf("%d items left, want 1", n)
	}
}

func TestTUI_QuitAsksWhenUnsaved(t *testing.T) {
	m, _ := newTestModel(t, "# TODO\n\n- [ ] one\n")

	// No changes: q quits straight away.
	if _, cmd := press(m).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}); cmd == nil {
		t.Error("q without changes should quit")
	}

	// With changes: q asks, n cancels, q then y quits.
	model := asModel(t, press(m, "space", "q"))
	if model.mode != modeConfirm || model.quitting {
		t.Fatal("q with unsaved changes should ask first")
	}
	model = asModel(t, press(model, "n"))
	if model.mode != modeNormal || model.quitting || !model.hasChanges {
		t.Error("n should cancel quitting and keep the changes")
	}
	next, cmd := asModel(t, press(model, "q")).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if !asModel(t, next).quitting || cmd == nil {
		t.Error("y should quit")
	}
}

func TestTUI_MenuQuitAsksWhenUnsaved(t *testing.T) {
	m, _ := newTestModel(t, "# TODO\n\n- [ ] one\n")
	model := asModel(t, press(m, "space", "esc", "q"))
	if model.mode != modeConfirm || model.quitting {
		t.Error("menu q with unsaved changes should ask first")
	}
}

func TestTUI_HelpOverlay(t *testing.T) {
	m, _ := newTestModel(t, "# TODO\n\n- [ ] one\n")
	model := asModel(t, press(m, "?"))
	if model.mode != modeHelp || !strings.Contains(model.View(), "cut / paste") || !strings.Contains(model.View(), "this help") {
		t.Fatal("? should show the help overlay")
	}
	if asModel(t, press(model, "j")).mode != modeNormal {
		t.Error("any key should close the help overlay")
	}
}

func TestView_StatesVisibleWithoutColour(t *testing.T) {
	m, _ := newTestModel(t, "# TODO\n\n- [x] done item\n- [ ] FIXME: urgent\n")
	model := asModel(t, press(m, "down", "space")).WithVersion("v1.2.3")
	view := ansi.Strip(model.View())

	for _, want := range []string{
		"mktodo v1.2.3",         // header with version
		"» TODO",                // focused pane marker
		"  [✓]   done item",     // done marker
		"› [✓] ! FIXME: urgent", // cursor marker and FIXME flag
		"[UNSAVED]",             // dirty indicator
	} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q:\n%s", want, view)
		}
	}
}

func TestView_FitsTerminalSize(t *testing.T) {
	m, _ := newTestModelWith(t, nestedDoc+"\n- [ ] "+strings.Repeat("very long description ", 10)+"\n", nestedConfig)
	screens := map[string][]string{"main": nil, "help": {"?"}, "menu": {"esc"}, "edit": {"e"}, "confirm": {"d"}}
	for _, size := range []tea.WindowSizeMsg{{Width: 80, Height: 24}, {Width: 50, Height: 12}, {Width: 120, Height: 40}} {
		for name, keys := range screens {
			next, _ := m.Update(size)
			lines := strings.Split(press(next, keys...).View(), "\n")
			if len(lines) > size.Height {
				t.Errorf("%s at %dx%d: %d lines", name, size.Width, size.Height, len(lines))
			}
			for i, line := range lines {
				if w := ansi.StringWidth(line); w > size.Width {
					t.Errorf("%s at %dx%d: line %d is %d cells wide", name, size.Width, size.Height, i, w)
				}
			}
		}
	}
}

func TestView_ScrollsToCursor(t *testing.T) {
	var b strings.Builder
	b.WriteString("# TODO\n\n")
	for i := 0; i < 30; i++ {
		b.WriteString("- [ ] item " + strings.Repeat("x", i%3) + string(rune('A'+i%26)) + "\n")
	}
	m, _ := newTestModel(t, b.String())
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 12})
	model := asModel(t, next)
	for i := 0; i < 29; i++ {
		model = asModel(t, press(model, "down"))
	}
	if !strings.Contains(ansi.Strip(model.View()), "› [ ]   item xxD") {
		t.Error("the selected last item should be visible after scrolling")
	}
}
