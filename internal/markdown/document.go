// Package markdown parses, modifies and writes TODO markdown documents.
package markdown

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/project"
	"github.com/jvzantvoort/mktodo/internal/todo"
)

// ErrItemNotFound is returned when an item is not part of the document.
var ErrItemNotFound = errors.New("item not found in document")

// Item is an alias for todo.Item for convenience
type Item = todo.Item

// Document represents a parsed markdown document.
//
// Items are shared with the TODO sections that render them, so changing an
// item's Done or Description changes the document text directly.
type Document struct {
	Path     string
	Sections []*Section
	Projects map[string]*project.Project
	Items    []*Item

	eol        string // line ending of the source file
	noFinalEOL bool   // source file did not end with a line ending

	// source is what the file looked like when the document was loaded
	// (or last saved), so a save can tell whether someone else changed it.
	source *sourceState
}

// NewDocument creates an empty document for a file that does not exist yet.
func NewDocument(path string, projects map[string]*project.Project) *Document {
	return &Document{
		Path:     path,
		Projects: projects,
		eol:      "\n",
		source:   &sourceState{},
	}
}

// LoadDocument loads and parses a markdown file
func LoadDocument(path string, cfg *config.Config) (*Document, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	projects, err := project.BuildHierarchy(cfg)
	if err != nil {
		return nil, fmt.Errorf("building hierarchy: %w", err)
	}

	lines, eol, trailingEOL := splitLines(string(content))
	sections := parseLines(lines)

	if err := AssociateProjects(sections, projects); err != nil {
		return nil, fmt.Errorf("associating projects: %w", err)
	}

	var items []*todo.Item
	for _, section := range sections {
		items = append(items, section.TODOItems...)
	}

	return &Document{
		Path:       path,
		Sections:   sections,
		Projects:   projects,
		Items:      items,
		eol:        eol,
		noFinalEOL: !trailingEOL,
		source:     newSourceState(content),
	}, nil
}

// GetItemsByProject returns TODO items for a specific project
func (d *Document) GetItemsByProject(proj *project.Project) []*todo.Item {
	var items []*todo.Item
	for _, item := range d.Items {
		if item.Project == proj {
			items = append(items, item)
		}
	}
	return items
}

// GetOpenItems returns all open (not done) TODO items
func (d *Document) GetOpenItems() []*todo.Item {
	var items []*todo.Item
	for _, item := range d.Items {
		if !item.Done {
			items = append(items, item)
		}
	}
	return items
}

// GetDoneItems returns all completed TODO items
func (d *Document) GetDoneItems() []*todo.Item {
	var items []*todo.Item
	for _, item := range d.Items {
		if item.Done {
			items = append(items, item)
		}
	}
	return items
}

// GetFIXMEItems returns all FIXME items
func (d *Document) GetFIXMEItems() []*todo.Item {
	var items []*todo.Item
	for _, item := range d.Items {
		if item.IsFIXME {
			items = append(items, item)
		}
	}
	return items
}

// FindItem finds TODO items matching a description; see MatchItems.
func (d *Document) FindItem(query string) []*todo.Item {
	return MatchItems(d.Items, query)
}

// MatchItems returns the items whose description equals query (ignoring
// case and surrounding whitespace). If there are none, it returns the items
// whose description contains query. Preferring exact matches means an item
// can always be selected by its full description, even when it is a prefix
// of another item.
func MatchItems(items []*todo.Item, query string) []*todo.Item {
	query = strings.TrimSpace(query)
	var exact, partial []*todo.Item
	for _, item := range items {
		if strings.EqualFold(item.Description, query) {
			exact = append(exact, item)
		} else if item.Matches(query) {
			partial = append(partial, item)
		}
	}
	if len(exact) > 0 {
		return exact
	}
	return partial
}

// HasSection reports whether the document has a header for the project.
func (d *Document) HasSection(proj *project.Project) bool {
	return d.headerIndex(proj) >= 0
}

// EnsureSection makes sure the document has a header for the project,
// creating it (and any missing parent headers) if needed. A root project's
// header is appended to the end of the document; a child project's header is
// placed at the end of its parent's section, so the hierarchy is kept.
func (d *Document) EnsureSection(proj *project.Project) error {
	if d.HasSection(proj) {
		return nil
	}
	if proj.Level < 1 || proj.Level > 6 {
		return fmt.Errorf("project %s: header level %d is outside 1-6", proj.Name, proj.Level)
	}

	pos := len(d.Sections)
	if proj.Parent != nil {
		if err := d.EnsureSection(proj.Parent); err != nil {
			return err
		}
		pos = d.subtreeEnd(d.headerIndex(proj.Parent))
	}

	header := &Section{
		Type:    SectionHeader,
		Level:   proj.Level,
		Title:   proj.Title,
		Project: proj,
		Lines:   []string{strings.Repeat("#", proj.Level) + " " + proj.Title},
	}
	if pos < len(d.Sections) {
		header.Lines = append(header.Lines, "")
	}

	if pos > 0 && lastLine(d.Sections[pos-1]) != "" {
		d.insertSection(pos, blankSection(1))
		pos++
	}
	d.insertSection(pos, header)
	return nil
}

// AddItem adds a new TODO item at the end of the project's TODO list,
// creating the project's header if the document doesn't have one.
//
// The description is checked with todo.CleanDescription first.
func (d *Document) AddItem(proj *project.Project, description string) (*todo.Item, error) {
	description, err := todo.CleanDescription(description)
	if err != nil {
		return nil, err
	}
	item := todo.New(description)
	if err := d.InsertItem(proj, nil, item); err != nil {
		return nil, err
	}
	return item, nil
}

// InsertItem puts an existing item into the project's TODO list: directly
// after the item `after`, or at the end of the list when after is nil. The
// item keeps its text and checkbox state, so this also pastes an item that
// was cut with RemoveItem. The project's header is created if needed.
func (d *Document) InsertItem(proj *project.Project, after, item *todo.Item) error {
	// Check `after` first so a rejected call leaves the document unchanged.
	if after != nil {
		if section, _ := d.locate(after); section == nil || after.Project != proj {
			return ErrItemNotFound
		}
	}
	if err := d.EnsureSection(proj); err != nil {
		return fmt.Errorf("creating section for project %s: %w", proj.Name, err)
	}
	item.Project = proj

	if after != nil {
		section, i := d.locate(after)
		section.TODOItems = append(section.TODOItems[:i+1], append([]*todo.Item{item}, section.TODOItems[i+1:]...)...)
		d.reindex()
		return nil
	}

	target := d.lastTODOSection(proj)
	if target == nil {
		target = d.insertTODOSection(d.headerIndex(proj))
	}
	target.TODOItems = append(target.TODOItems, item)
	d.reindex()
	return nil
}

// SwapItems exchanges the positions of two items in the same project. Each
// item keeps its own line, so this is how an item is moved up or down.
func (d *Document) SwapItems(a, b *todo.Item) error {
	if a.Project != b.Project {
		return fmt.Errorf("items belong to different projects")
	}
	sa, ia := d.locate(a)
	sb, ib := d.locate(b)
	if sa == nil || sb == nil {
		return ErrItemNotFound
	}
	sa.TODOItems[ia], sb.TODOItems[ib] = b, a
	d.reindex()
	return nil
}

// locate returns the TODO section holding item and its index there.
func (d *Document) locate(item *todo.Item) (*Section, int) {
	for _, section := range d.Sections {
		if section.Type != SectionTODO {
			continue
		}
		for i, it := range section.TODOItems {
			if it == item {
				return section, i
			}
		}
	}
	return nil, -1
}

// reindex rebuilds Items in document order after items move.
func (d *Document) reindex() {
	d.Items = nil
	for _, section := range d.Sections {
		d.Items = append(d.Items, section.TODOItems...)
	}
}

// RemoveItem removes a TODO item from the document
func (d *Document) RemoveItem(item *todo.Item) error {
	for si, section := range d.Sections {
		if section.Type != SectionTODO {
			continue
		}
		for i, sectionItem := range section.TODOItems {
			if sectionItem != item {
				continue
			}
			section.TODOItems = append(section.TODOItems[:i], section.TODOItems[i+1:]...)
			d.removeFromItems(item)
			if len(section.TODOItems) == 0 {
				d.removeEmptyTODOSection(si)
			}
			return nil
		}
	}
	return ErrItemNotFound
}

// UpdateItem checks that the item belongs to the document. Changes to an
// item are rendered automatically, so no other work is needed; it is kept so
// callers can confirm an item is still part of the document.
func (d *Document) UpdateItem(item *todo.Item) error {
	for _, docItem := range d.Items {
		if docItem == item {
			return nil
		}
	}
	return ErrItemNotFound
}

// String returns the document as markdown text, using the line ending of the
// source file and keeping whether it ended with a final line ending.
func (d *Document) String() string {
	var lines []string
	for _, section := range d.Sections {
		lines = append(lines, section.lines()...)
	}
	if len(lines) == 0 {
		return ""
	}
	eol := d.eol
	if eol == "" {
		eol = "\n"
	}
	result := strings.Join(lines, eol)
	if !d.noFinalEOL {
		result += eol
	}
	return result
}

func (d *Document) headerIndex(proj *project.Project) int {
	for i, section := range d.Sections {
		if section.Type == SectionHeader && section.Project == proj {
			return i
		}
	}
	return -1
}

// bodyEnd returns the index of the first header after index i, or the
// number of sections. The sections in between belong to the header at i.
func (d *Document) bodyEnd(i int) int {
	for j := i + 1; j < len(d.Sections); j++ {
		if d.Sections[j].Type == SectionHeader {
			return j
		}
	}
	return len(d.Sections)
}

// subtreeEnd returns the index of the first header after i at the same or a
// higher level, so the sections in between include all child sections.
func (d *Document) subtreeEnd(i int) int {
	level := d.Sections[i].Level
	for j := i + 1; j < len(d.Sections); j++ {
		if s := d.Sections[j]; s.Type == SectionHeader && s.Level <= level {
			return j
		}
	}
	return len(d.Sections)
}

func (d *Document) lastTODOSection(proj *project.Project) *Section {
	hi := d.headerIndex(proj)
	if hi < 0 {
		return nil
	}
	var last *Section
	for _, s := range d.Sections[hi+1 : d.bodyEnd(hi)] {
		if s.Type == SectionTODO {
			last = s
		}
	}
	return last
}

// insertTODOSection creates an empty TODO list after the header at index hi.
// The list is separated from the header text by one blank line, and the
// blank lines that ended the header section are moved after the list so the
// next header stays separated.
func (d *Document) insertTODOSection(hi int) *Section {
	header := d.Sections[hi]
	trailing := 0
	for trailing < len(header.Lines)-1 && header.Lines[len(header.Lines)-1-trailing] == "" {
		trailing++
	}
	header.Lines = append(header.Lines[:len(header.Lines)-trailing:len(header.Lines)-trailing], "")

	list := &Section{Type: SectionTODO, StartLine: header.StartLine + len(header.Lines)}
	d.insertSection(hi+1, list)

	next := hi + 2
	switch {
	case trailing > 0:
		d.insertSection(next, blankSection(trailing))
	case next < len(d.Sections) && firstLine(d.Sections[next]) != "":
		d.insertSection(next, blankSection(1))
	}
	return list
}

// removeEmptyTODOSection drops an emptied TODO section at index i and, if
// that leaves two blank lines in a row, one of them.
func (d *Document) removeEmptyTODOSection(i int) {
	d.Sections = append(d.Sections[:i], d.Sections[i+1:]...)
	if i == 0 || i >= len(d.Sections) {
		return
	}
	prev, next := d.Sections[i-1], d.Sections[i]
	if next.Type == SectionOther && lastLine(prev) == "" && firstLine(next) == "" {
		next.Lines = next.Lines[1:]
		if len(next.Lines) == 0 {
			d.Sections = append(d.Sections[:i], d.Sections[i+1:]...)
		}
	}
}

func (d *Document) removeFromItems(item *todo.Item) {
	for j, docItem := range d.Items {
		if docItem == item {
			d.Items = append(d.Items[:j], d.Items[j+1:]...)
			return
		}
	}
}

func (d *Document) insertSection(i int, s *Section) {
	d.Sections = append(d.Sections, nil)
	copy(d.Sections[i+1:], d.Sections[i:])
	d.Sections[i] = s
}

func blankSection(n int) *Section {
	return &Section{Type: SectionOther, Lines: make([]string, n)}
}

func firstLine(s *Section) string {
	if lines := s.lines(); len(lines) > 0 {
		return lines[0]
	}
	return ""
}

func lastLine(s *Section) string {
	if lines := s.lines(); len(lines) > 0 {
		return lines[len(lines)-1]
	}
	return ""
}
