package markdown

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jvzantvoort/mktodo/internal/config"
)

func strPtr(s string) *string { return &s }

// defaultConfig has a single root project "default" titled "TODO".
var defaultConfig = &config.Config{
	TodoFile: "README.md",
	Projects: []config.ProjectConfig{{Name: "default", Title: "TODO", Parent: strPtr("nil")}},
}

// nestedConfig has lego > technic, lego > city, and a separate "bugs" root.
var nestedConfig = &config.Config{
	TodoFile: "README.md",
	Projects: []config.ProjectConfig{
		{Name: "lego", Title: "Lego", Parent: strPtr("nil")},
		{Name: "technic", Title: "Technic", Parent: strPtr("lego")},
		{Name: "city", Title: "City", Parent: strPtr("lego")},
		{Name: "bugs", Title: "Bugs", Parent: strPtr("nil")},
	},
}

// loadString writes content to a temp file and loads it as a document.
func loadString(t *testing.T, content string, cfg *config.Config) *Document {
	t.Helper()
	path := filepath.Join(t.TempDir(), "README.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := LoadDocument(path, cfg)
	if err != nil {
		t.Fatalf("LoadDocument() error = %v", err)
	}
	return doc
}

// saveAndRead saves the document and returns the file content.
func saveAndRead(t *testing.T, doc *Document) string {
	t.Helper()
	if err := SaveDocument(doc.Path, doc); err != nil {
		t.Fatalf("SaveDocument() error = %v", err)
	}
	b, err := os.ReadFile(doc.Path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestRoundTrip_Unchanged(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"empty file", ""},
		{"no final newline", "# TODO\n\n- [ ] a"},
		{"CRLF line endings", "# TODO\r\n\r\n- [ ] a\r\n- [x] b\r\n"},
		{"trailing blank lines", "# TODO\n\n- [ ] a\n\n\n"},
		{"bullets, indentation and spacing", "# TODO\n\n* [ ] star\n  + [x] nested\n-  [X]   wide  \n"},
		{"code fences", "# TODO\n\n```bash\n# not a header\n- [ ] not an item\n```\n\n~~~~\n```\n- [ ] still code\n~~~~\n"},
		{"non-project content", "Intro\n\n# Other\n\n- [ ] orphan\n\n## TODO\n\ntext\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := loadString(t, tt.content, defaultConfig)
			if got := saveAndRead(t, doc); got != tt.content {
				t.Errorf("round trip changed the file:\ngot:  %q\nwant: %q", got, tt.content)
			}
		})
	}
}

func TestToggle_PreservesFormatting(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{"star bullet and indent", "# TODO\n\n  * [ ] star\n", "# TODO\n\n  * [X] star\n"},
		{"lowercase x is kept", "# TODO\n\n- [x] done\n", "# TODO\n\n- [ ] done\n"},
		{"CRLF is kept", "# TODO\r\n\r\n- [ ] a\r\n", "# TODO\r\n\r\n- [X] a\r\n"},
		{"no final newline is kept", "# TODO\n\n- [ ] a", "# TODO\n\n- [X] a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := loadString(t, tt.content, defaultConfig)
			doc.Items[0].MarkDone()
			if got := saveAndRead(t, doc); got != tt.want {
				t.Errorf("got  %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestToggle_TwiceRestoresOriginalLine(t *testing.T) {
	content := "# TODO\n\n- [x]   done  \n"
	doc := loadString(t, content, defaultConfig)
	doc.Items[0].MarkDone()
	doc.Items[0].MarkDone()
	if got := saveAndRead(t, doc); got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestEdit_KeepsBulletAndIndent(t *testing.T) {
	doc := loadString(t, "# TODO\n\n  * [ ] old\n", defaultConfig)
	doc.Items[0].SetDescription("FIXME: new")
	if !doc.Items[0].IsFIXME {
		t.Error("SetDescription should update IsFIXME")
	}
	want := "# TODO\n\n  * [ ] FIXME: new\n"
	if got := saveAndRead(t, doc); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAddItem(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		content string
		project string
		want    string
	}{
		{
			name:    "appends to end of list split by blank lines",
			cfg:     defaultConfig,
			content: "# TODO\n\n- [ ] a\n\n- [ ] b\n\n# Other\n",
			project: "default",
			want:    "# TODO\n\n- [ ] a\n\n- [ ] b\n- [ ] new\n\n# Other\n",
		},
		{
			name:    "creates section in existing file",
			cfg:     defaultConfig,
			content: "# My Project\n\nSome text.\n",
			project: "default",
			want:    "# My Project\n\nSome text.\n\n# TODO\n\n- [ ] new\n",
		},
		{
			name:    "creates list under header with intro text",
			cfg:     defaultConfig,
			content: "# TODO\n\nIntro.\n\n# Next\n",
			project: "default",
			want:    "# TODO\n\nIntro.\n\n- [ ] new\n\n# Next\n",
		},
		{
			name:    "creates list under header followed directly by another",
			cfg:     defaultConfig,
			content: "# TODO\n# Next\n",
			project: "default",
			want:    "# TODO\n\n- [ ] new\n\n# Next\n",
		},
		{
			name:    "creates child and parent headers",
			cfg:     nestedConfig,
			content: "# Notes\n",
			project: "technic",
			want:    "# Notes\n\n# Lego\n\n## Technic\n\n- [ ] new\n",
		},
		{
			name:    "creates child header inside parent section",
			cfg:     nestedConfig,
			content: "# Lego\n\n## Technic\n\n- [ ] t\n\n# Bugs\n\n- [ ] b\n",
			project: "city",
			want:    "# Lego\n\n## Technic\n\n- [ ] t\n\n## City\n\n- [ ] new\n\n# Bugs\n\n- [ ] b\n",
		},
		{
			name:    "skips headers and items in code fences",
			cfg:     defaultConfig,
			content: "```\n# TODO\n- [ ] example\n```\n\n# TODO\n\n- [ ] real\n",
			project: "default",
			want:    "```\n# TODO\n- [ ] example\n```\n\n# TODO\n\n- [ ] real\n- [ ] new\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := loadString(t, tt.content, tt.cfg)
			proj := doc.Projects[tt.project]
			item, err := doc.AddItem(proj, "new")
			if err != nil {
				t.Fatalf("AddItem() error = %v", err)
			}
			if item.Project != proj {
				t.Errorf("item.Project = %v, want %v", item.Project, proj)
			}
			if got := saveAndRead(t, doc); got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
			// The added item must be found again under its project.
			reloaded := loadString(t, tt.want, tt.cfg)
			found := MatchItems(reloaded.GetItemsByProject(reloaded.Projects[tt.project]), "new")
			if len(found) != 1 {
				t.Errorf("after reload, found %d items named 'new' in %s, want 1", len(found), tt.project)
			}
		})
	}
}

func TestRemoveItem_LastItemLeavesNoExtraBlank(t *testing.T) {
	doc := loadString(t, "# TODO\n\n- [ ] only\n\n# Next\n", defaultConfig)
	if err := doc.RemoveItem(doc.Items[0]); err != nil {
		t.Fatal(err)
	}
	want := "# TODO\n\n# Next\n"
	if got := saveAndRead(t, doc); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRemoveItem_KeepsOtherLines(t *testing.T) {
	doc := loadString(t, "# TODO\n\n* [ ] a\n  - [x] b\n- [ ] c\n", defaultConfig)
	if err := doc.RemoveItem(doc.Items[1]); err != nil {
		t.Fatal(err)
	}
	want := "# TODO\n\n* [ ] a\n- [ ] c\n"
	if got := saveAndRead(t, doc); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCodeFence_ItemsAndHeadersIgnored(t *testing.T) {
	doc := loadString(t, "# TODO\n\n- [ ] real\n\n```md\n# TODO\n- [ ] fake\n```\n", defaultConfig)
	if len(doc.Items) != 1 || doc.Items[0].Description != "real" {
		t.Errorf("items = %v, want only 'real'", doc.Items)
	}
}

func TestAssociateProjects_SameTitleDifferentParents(t *testing.T) {
	cfg := &config.Config{Projects: []config.ProjectConfig{
		{Name: "web", Title: "Web", Parent: strPtr("nil")},
		{Name: "webbugs", Title: "Bugs", Parent: strPtr("web")},
		{Name: "cli", Title: "CLI", Parent: strPtr("nil")},
		{Name: "clibugs", Title: "Bugs", Parent: strPtr("cli")},
	}}
	doc := loadString(t, "# Web\n\n## Bugs\n\n- [ ] w\n\n# CLI\n\n## Bugs\n\n- [ ] c\n", cfg)

	for name, want := range map[string]string{"webbugs": "w", "clibugs": "c"} {
		items := doc.GetItemsByProject(doc.Projects[name])
		if len(items) != 1 || items[0].Description != want {
			t.Errorf("%s items = %v, want [%s]", name, items, want)
		}
	}
}

func TestAssociateProjects_ChildOutsideParentNotMatched(t *testing.T) {
	doc := loadString(t, "# Bugs\n\n## Technic\n\n- [ ] x\n", nestedConfig)
	if items := doc.GetItemsByProject(doc.Projects["technic"]); len(items) != 0 {
		t.Errorf("technic matched a header outside Lego: %v", items)
	}
}

func TestMatchItems_PrefersExactMatch(t *testing.T) {
	doc := loadString(t, "# TODO\n\n- [ ] Build crane\n- [ ] Build crane model\n", defaultConfig)

	if got := MatchItems(doc.Items, "build CRANE"); len(got) != 1 || got[0].Description != "Build crane" {
		t.Errorf("exact match: got %v, want only 'Build crane'", got)
	}
	if got := MatchItems(doc.Items, "crane"); len(got) != 2 {
		t.Errorf("partial match: got %d items, want 2", len(got))
	}
}

func TestNewDocument_EnsureSectionsInOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "TODO.md")
	doc := loadString(t, "", nestedConfig)
	doc = NewDocument(path, doc.Projects)
	for _, name := range []string{"lego", "technic", "city", "bugs"} {
		if err := doc.EnsureSection(doc.Projects[name]); err != nil {
			t.Fatal(err)
		}
	}
	want := "# Lego\n\n## Technic\n\n## City\n\n# Bugs\n"
	if got := saveAndRead(t, doc); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
