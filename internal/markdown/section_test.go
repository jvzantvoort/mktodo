package markdown

import (
"testing"

"github.com/jvzantvoort/mktodo/internal/project"
)

func TestParseSections(t *testing.T) {
tests := []struct {
name          string
content       string
wantSections  int
validateFirst func(*testing.T, *Section)
}{
{
name:         "empty content",
content:      "",
wantSections: 0,
},
{
name:         "single header",
content:      "# Header\n",
wantSections: 1,
validateFirst: func(t *testing.T, s *Section) {
if s.Type != SectionHeader {
t.Errorf("Type = %v, want SectionHeader", s.Type)
}
if s.Level != 1 {
t.Errorf("Level = %d, want 1", s.Level)
}
if s.Title != "Header" {
t.Errorf("Title = %q, want %q", s.Title, "Header")
}
},
},
{
name: "header with TODO items",
content: `# TODO

- [ ] Task 1
- [X] Task 2
`,
wantSections: 2,
validateFirst: func(t *testing.T, s *Section) {
if s.Type != SectionHeader {
t.Errorf("Type = %v, want SectionHeader", s.Type)
}
},
},
{
name: "mixed content",
content: `# Header 1

Some text here.

- [ ] TODO item

More text.

## Header 2

- [ ] Another TODO
`,
wantSections: 6,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
sections, err := ParseSections(tt.content)
if err != nil {
t.Fatalf("ParseSections() error = %v", err)
}

if len(sections) != tt.wantSections {
t.Errorf("got %d sections, want %d", len(sections), tt.wantSections)
}

if tt.validateFirst != nil && len(sections) > 0 {
tt.validateFirst(t, sections[0])
}
})
}
}

func TestParseSections_TODODetection(t *testing.T) {
content := `# Project

- [ ] Open task
- [X] Done task
- [ ] FIXME: broken
* [ ] With asterisk
+ [ ] With plus
`

sections, err := ParseSections(content)
if err != nil {
t.Fatalf("ParseSections() error = %v", err)
}

// Find TODO section
var todoSection *Section
for _, s := range sections {
if s.Type == SectionTODO {
todoSection = s
break
}
}

if todoSection == nil {
t.Fatal("no TODO section found")
}

if len(todoSection.TODOItems) != 5 {
t.Errorf("got %d TODO items, want 5", len(todoSection.TODOItems))
}

// Check FIXME detection
fixmeCount := 0
for _, item := range todoSection.TODOItems {
if item.IsFIXME {
fixmeCount++
}
}
if fixmeCount != 1 {
t.Errorf("got %d FIXME items, want 1", fixmeCount)
}
}

func TestAssociateProjects(t *testing.T) {
content := `# TODO

- [ ] Task 1

## Subtask

- [ ] Task 2
`

sections, err := ParseSections(content)
if err != nil {
t.Fatalf("ParseSections() error = %v", err)
}

// Create mock projects
projects := map[string]*project.Project{
"default": {
Name:  "default",
Title: "TODO",
Level: 1,
},
"subtask": {
Name:  "subtask",
Title: "Subtask",
Level: 2,
},
}

err = AssociateProjects(sections, projects)
if err != nil {
t.Fatalf("AssociateProjects() error = %v", err)
}

// Check that header sections have projects
headerCount := 0
for _, section := range sections {
if section.Type == SectionHeader && section.Project != nil {
headerCount++
}
}

if headerCount != 2 {
t.Errorf("got %d headers with projects, want 2", headerCount)
}
}

func TestSection_String(t *testing.T) {
section := &Section{
Type:  SectionHeader,
Level: 2,
Title: "Test Header",
Lines: []string{
"## Test Header",
"Some content",
},
}

expected := "## Test Header\nSome content"
got := section.String()

if got != expected {
t.Errorf("String() = %q, want %q", got, expected)
}
}
