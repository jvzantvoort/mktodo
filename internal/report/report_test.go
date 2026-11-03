package report

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/markdown"
)

var generated = time.Date(2026, 11, 4, 9, 30, 0, 0, time.UTC)

func strPtr(s string) *string { return &s }

// testConfig has TODO, and Lego with children Technic and City.
var testConfig = &config.Config{Projects: []config.ProjectConfig{
	{Name: "default", Title: "TODO", Parent: strPtr("nil")},
	{Name: "lego", Title: "Lego", Parent: strPtr("nil")},
	{Name: "technic", Title: "Technic", Parent: strPtr("lego")},
	{Name: "city", Title: "City", Parent: strPtr("lego")},
}}

const testDoc = `# TODO

- [ ] FIXME: parser bug
- [x] docs
- [x] tests
- [ ] release

# Lego

## Technic

- [ ] crane
- [X] FIXME: lost pieces

## City

# Notes

- [ ] orphan
`

func loadDoc(t *testing.T, content string) *markdown.Document {
	t.Helper()
	path := filepath.Join(t.TempDir(), "TODO.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := markdown.LoadDocument(path, testConfig)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestBuild_Counts(t *testing.T) {
	r := Build(loadDoc(t, testDoc), generated)

	if want := (Counts{Total: 7, Open: 4, Done: 3, FIXME: 2}); r.Summary != want {
		t.Errorf("Summary = %+v, want %+v", r.Summary, want)
	}
	if want := (Counts{Total: 1, Open: 1}); r.Unassigned != want {
		t.Errorf("Unassigned = %+v, want %+v", r.Unassigned, want)
	}

	want := []struct {
		path  string
		depth int
		Counts
	}{
		{"default", 0, Counts{Total: 4, Open: 2, Done: 2, FIXME: 1}},
		{"lego", 0, Counts{}},
		{"lego.technic", 1, Counts{Total: 2, Open: 1, Done: 1, FIXME: 1}},
		{"lego.city", 1, Counts{}},
	}
	if len(r.Projects) != len(want) {
		t.Fatalf("got %d projects, want %d", len(r.Projects), len(want))
	}
	for i, w := range want {
		p := r.Projects[i]
		if p.Path != w.path || p.Depth != w.depth || p.Counts != w.Counts {
			t.Errorf("Projects[%d] = %s depth %d %+v, want %s depth %d %+v", i, p.Path, p.Depth, p.Counts, w.path, w.depth, w.Counts)
		}
	}

	if len(r.FIXMEItems) != 2 || r.FIXMEItems[1].Description != "FIXME: lost pieces" || !r.FIXMEItems[1].Done {
		t.Errorf("FIXMEItems = %+v", r.FIXMEItems)
	}
}

func TestWriteText(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteText(&buf, Build(loadDoc(t, testDoc), generated)); err != nil {
		t.Fatal(err)
	}

	want := `mktodo Report - 2026-11-04
══════════════════════════

Summary
───────
Total Tasks:     7
Open:            4 (57%)
Completed:       3 (43%)
FIXME:           2 (29%)

By Project
──────────
TODO          █████░░░░░  2/4  (50%)  [1 FIXME]
Lego          ░░░░░░░░░░  no items
  Technic     █████░░░░░  1/2  (50%)  [1 FIXME]
  City        ░░░░░░░░░░  no items
(no project)  ░░░░░░░░░░  0/1  (0%)

FIXME Items
───────────
1. [TODO] FIXME: parser bug
2. [Technic] FIXME: lost pieces (done)
`
	if got := buf.String(); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestWriteText_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteText(&buf, Build(loadDoc(t, "# TODO\n"), generated)); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Total Tasks:     0", "Open:            0 (0%)", "TODO       ░░░░░░░░░░  no items", "No FIXME items."} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("output missing %q:\n%s", want, buf.String())
		}
	}
}

func TestBar(t *testing.T) {
	for percent, want := range map[int]string{
		0: "░░░░░░░░░░", 4: "░░░░░░░░░░", 5: "█░░░░░░░░░", 50: "█████░░░░░", 100: "██████████",
	} {
		if got := bar(percent); got != want {
			t.Errorf("bar(%d) = %q, want %q", percent, got, want)
		}
	}
}

func TestWriteMarkdown(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteMarkdown(&buf, Build(loadDoc(t, testDoc), generated)); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"# mktodo Report - 2026-11-04\n",
		"- **Open:** 4 (57%)\n",
		"| TODO | 2 | 4 | 50% | 1 |\n",
		"| &nbsp;&nbsp;Technic | 1 | 2 | 50% | 1 |\n",
		"| &nbsp;&nbsp;City | 0 | 0 | – | 0 |\n",
		"| (no project) | 0 | 1 | 0% | 0 |\n",
		"2. **Technic:** FIXME: lost pieces _(done)_\n",
	} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("markdown missing %q:\n%s", want, buf.String())
		}
	}
}

func TestWriteMarkdown_EscapesUserText(t *testing.T) {
	var buf bytes.Buffer
	doc := loadDoc(t, "# TODO\n\n- [ ] FIXME: a | b *c* <x>\n")
	if err := WriteMarkdown(&buf, Build(doc, generated)); err != nil {
		t.Fatal(err)
	}
	if want := `FIXME: a \| b \*c\* &lt;x&gt;`; !strings.Contains(buf.String(), want) {
		t.Errorf("markdown missing escaped %q:\n%s", want, buf.String())
	}
}

func TestWriteHTML(t *testing.T) {
	var buf bytes.Buffer
	doc := loadDoc(t, testDoc+"\n# TODO\n\n- [ ] FIXME: <script>alert(1)</script>\n")
	if err := WriteHTML(&buf, Build(doc, generated)); err != nil {
		t.Fatal(err)
	}
	html := buf.String()

	if strings.Contains(html, "<script>") {
		t.Error("user text was not escaped")
	}
	for _, want := range []string{
		"<!doctype html>",
		"<title>mktodo Report - 2026-11-04</title>",
		"prefers-color-scheme: dark",
		`style="width: 50%"`,
		`style="padding-left: 28px">Technic</td>`,
		"&lt;script&gt;",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("HTML missing %q", want)
		}
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, Build(loadDoc(t, testDoc), generated)); err != nil {
		t.Fatal(err)
	}

	var got struct {
		Generated string `json:"generated"`
		Summary   struct {
			Total, Open, Done, FIXME, Completion int
		} `json:"summary"`
		Projects []struct {
			Path  string `json:"path"`
			Depth int    `json:"depth"`
			Done  int    `json:"done"`
		} `json:"projects"`
		Unassigned *struct{ Total int } `json:"unassigned"`
		FIXMEItems []struct {
			Project string `json:"project"`
			Done    bool   `json:"done"`
		} `json:"fixme_items"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if got.Generated != "2026-11-04T09:30:00Z" {
		t.Errorf("generated = %q", got.Generated)
	}
	if got.Summary.Total != 7 || got.Summary.FIXME != 2 || got.Summary.Completion != 43 {
		t.Errorf("summary = %+v", got.Summary)
	}
	if len(got.Projects) != 4 || got.Projects[2].Path != "lego.technic" || got.Projects[2].Depth != 1 {
		t.Errorf("projects = %+v", got.Projects)
	}
	if got.Unassigned == nil || got.Unassigned.Total != 1 {
		t.Errorf("unassigned = %+v", got.Unassigned)
	}
	if len(got.FIXMEItems) != 2 || got.FIXMEItems[1].Project != "lego.technic" || !got.FIXMEItems[1].Done {
		t.Errorf("fixme_items = %+v", got.FIXMEItems)
	}
}

func TestWriteJSON_EmptyListsAreArrays(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, Build(loadDoc(t, ""), generated)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"fixme_items": []`) || strings.Contains(buf.String(), "null") {
		t.Errorf("empty report should use [] and omit unassigned:\n%s", buf.String())
	}
}
