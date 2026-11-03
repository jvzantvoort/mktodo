package markdown

import (
	"regexp"
	"strings"

	"github.com/jvzantvoort/mktodo/internal/project"
	"github.com/jvzantvoort/mktodo/internal/todo"
)

// SectionType represents the type of content section
type SectionType int

const (
	// SectionHeader is a markdown header line plus the non-TODO content after it
	SectionHeader SectionType = iota
	// SectionTODO is a run of consecutive TODO item lines
	SectionTODO
	// SectionOther is any other content
	SectionOther
)

// Section represents a section of markdown content.
//
// Header and other sections keep their original lines in Lines. A TODO
// section is rendered from TODOItems, so changing an item is enough to
// change the output; its Lines field is not used.
type Section struct {
	Type      SectionType
	Level     int              // Header level (1-6), 0 for non-headers
	Title     string           // Header title
	Project   *project.Project // Associated project (for headers)
	Lines     []string         // Original lines (header and other sections)
	TODOItems []*todo.Item     // Parsed TODO items (if SectionTODO)
	StartLine int              // Line number where section starts (1-indexed)
}

var (
	headerPattern = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	fencePattern  = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})")
)

// fence tracks whether the parser is inside a fenced code block.
type fence struct {
	marker string // opening run of ` or ~, empty when outside a block
}

// update reports whether line is part of a fenced code block (including the
// opening and closing fence lines) and advances the fence state.
func (f *fence) update(line string) bool {
	m := fencePattern.FindStringSubmatch(line)
	if f.marker == "" {
		if m != nil {
			f.marker = m[1]
			return true
		}
		return false
	}
	// A closing fence uses the same character, is at least as long, and has
	// nothing but whitespace after it.
	if m != nil && m[1][0] == f.marker[0] && len(m[1]) >= len(f.marker) {
		rest := strings.TrimLeft(line, " ")[len(m[1]):]
		if strings.TrimSpace(rest) == "" {
			f.marker = ""
		}
	}
	return true
}

// ParseSections parses markdown content into sections
func ParseSections(content string) ([]*Section, error) {
	lines, _, _ := splitLines(content)
	return parseLines(lines), nil
}

// splitLines splits content into lines without line terminators. It reports
// the line ending used by the file and whether the content ended with one,
// so the document can be written back unchanged.
func splitLines(content string) (lines []string, eol string, trailingEOL bool) {
	eol = "\n"
	if idx := strings.Index(content, "\n"); idx > 0 && content[idx-1] == '\r' {
		eol = "\r\n"
	}
	if content == "" {
		return nil, eol, true
	}

	trailingEOL = strings.HasSuffix(content, "\n")
	content = strings.TrimSuffix(content, "\n")
	lines = strings.Split(content, "\n")
	if eol == "\r\n" {
		for i, line := range lines {
			lines[i] = strings.TrimSuffix(line, "\r")
		}
	}
	return lines, eol, trailingEOL
}

// sectionBuilder accumulates lines into sections.
type sectionBuilder struct {
	sections []*Section
	current  *Section
}

func (b *sectionBuilder) flush() {
	if b.current != nil && (len(b.current.Lines) > 0 || len(b.current.TODOItems) > 0) {
		b.sections = append(b.sections, b.current)
	}
	b.current = nil
}

func (b *sectionBuilder) addHeader(line string, level int, title string, lineNum int) {
	b.flush()
	b.current = &Section{
		Type:      SectionHeader,
		Level:     level,
		Title:     title,
		Lines:     []string{line},
		StartLine: lineNum,
	}
}

func (b *sectionBuilder) addItem(item *todo.Item, lineNum int) {
	if b.current == nil || b.current.Type != SectionTODO {
		b.flush()
		b.current = &Section{Type: SectionTODO, StartLine: lineNum}
	}
	item.LineNumber = lineNum
	b.current.TODOItems = append(b.current.TODOItems, item)
}

func (b *sectionBuilder) addOther(line string, lineNum int) {
	// Content after a header belongs to the header section, so the header and
	// its introduction stay together.
	if b.current == nil || b.current.Type == SectionTODO {
		b.flush()
		b.current = &Section{Type: SectionOther, StartLine: lineNum}
	}
	b.current.Lines = append(b.current.Lines, line)
}

// parseLines groups lines into header, TODO and other sections. Lines inside
// fenced code blocks are always treated as plain content, so headers and
// checkboxes in code examples are left alone.
func parseLines(lines []string) []*Section {
	var b sectionBuilder
	var f fence

	for i, line := range lines {
		lineNum := i + 1

		if f.update(line) {
			b.addOther(line, lineNum)
			continue
		}

		if matches := headerPattern.FindStringSubmatch(line); matches != nil {
			b.addHeader(line, len(matches[1]), strings.TrimSpace(matches[2]), lineNum)
			continue
		}

		// todo.Parse never returns an error; a non-matching line is nil.
		if item, _ := todo.Parse(line); item != nil {
			b.addItem(item, lineNum)
			continue
		}

		b.addOther(line, lineNum)
	}

	b.flush()
	return b.sections
}

// AssociateProjects links header sections to projects and TODO items to the
// project whose header they appear under.
//
// A header matches a project when the level and title match and, for a child
// project, the nearest enclosing header is the parent project's header. This
// keeps projects with the same title under different parents apart. Projects
// are tried in configuration order so the result is deterministic.
func AssociateProjects(sections []*Section, projects map[string]*project.Project) error {
	ordered := project.Sorted(projects)
	var open [7]*project.Project // matched project per header level (1-6)
	var currentProject *project.Project

	for _, section := range sections {
		switch section.Type {
		case SectionHeader:
			level := section.Level
			for l := level; l < len(open); l++ {
				open[l] = nil
			}
			section.Project = matchHeader(ordered, section, open[level-1])
			open[level] = section.Project
			currentProject = section.Project
		case SectionTODO:
			for _, item := range section.TODOItems {
				item.Project = currentProject
			}
		}
	}

	return nil
}

// matchHeader returns the project for a header section, or nil. enclosing is
// the project matched by the nearest header one level up, if any.
func matchHeader(ordered []*project.Project, section *Section, enclosing *project.Project) *project.Project {
	for _, proj := range ordered {
		if proj.Level != section.Level || proj.Title != section.Title {
			continue
		}
		if proj.Parent == nil || proj.Parent == enclosing {
			return proj
		}
	}
	return nil
}

// lines returns the section's lines; TODO sections are rendered from their items.
func (s *Section) lines() []string {
	if s.Type != SectionTODO {
		return s.Lines
	}
	out := make([]string, len(s.TODOItems))
	for i, item := range s.TODOItems {
		out[i] = item.String()
	}
	return out
}

// String returns the section as markdown text
func (s *Section) String() string {
	return strings.Join(s.lines(), "\n")
}
