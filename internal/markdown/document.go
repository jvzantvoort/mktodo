package markdown

import (
	"fmt"
	"os"
	"strings"

	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/project"
	"github.com/jvzantvoort/mktodo/internal/todo"
)

// Item is an alias for todo.Item for convenience
type Item = todo.Item

// Document represents a parsed markdown document
type Document struct {
	Path     string
	Sections []*Section
	Projects map[string]*project.Project
	Items    []*Item
}

// LoadDocument loads and parses a markdown file
func LoadDocument(path string, cfg *config.Config) (*Document, error) {
	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Build project hierarchy
	projects, err := project.BuildHierarchy(cfg)
	if err != nil {
		return nil, fmt.Errorf("building hierarchy: %w", err)
	}

	// Parse sections
	sections, err := ParseSections(string(content))
	if err != nil {
		return nil, fmt.Errorf("parsing sections: %w", err)
	}

	// Associate projects with sections
	if err := AssociateProjects(sections, projects); err != nil {
		return nil, fmt.Errorf("associating projects: %w", err)
	}

	// Collect all TODO items
	var items []*todo.Item
	for _, section := range sections {
		items = append(items, section.TODOItems...)
	}

	doc := &Document{
		Path:     path,
		Sections: sections,
		Projects: projects,
		Items:    items,
	}

	return doc, nil
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

// FindItem finds a TODO item by description (case-insensitive partial match)
func (d *Document) FindItem(query string) []*todo.Item {
	var items []*todo.Item
	for _, item := range d.Items {
		if item.Matches(query) {
			items = append(items, item)
		}
	}
	return items
}

// AddItem adds a new TODO item to a project
func (d *Document) AddItem(proj *project.Project, description string) (*todo.Item, error) {
	item := &todo.Item{
		Description: description,
		Done:        false,
		Project:     proj,
	}

	// Check for FIXME
	if strings.HasPrefix(strings.ToUpper(description), "FIXME") {
		item.IsFIXME = true
	}

	// Find the section for this project
	var targetSection *Section
	for _, section := range d.Sections {
		if section.Type == SectionHeader && section.Project == proj {
			// Look for TODO section after this header
			for i, s := range d.Sections {
				if s == section {
					// Check next section
					if i+1 < len(d.Sections) && d.Sections[i+1].Type == SectionTODO {
						targetSection = d.Sections[i+1]
					}
					break
				}
			}
			break
		}
	}

	// If no TODO section exists, create one
	if targetSection == nil {
		// Find the header section
		for i, section := range d.Sections {
			if section.Type == SectionHeader && section.Project == proj {
				// Check if header section has content beyond the header line
				// Header sections may contain: header line, blank line, then description
				headerEndIdx := 1 // At minimum, keep the header line

				// Skip any blank lines immediately after header
				for headerEndIdx < len(section.Lines) && section.Lines[headerEndIdx] == "" {
					headerEndIdx++
				}

				// If there's additional content after header, split the section
				var newOtherSection *Section
				if headerEndIdx < len(section.Lines) {
					// Create new section for the remaining content
					newOtherSection = &Section{
						Type:      SectionOther,
						Lines:     section.Lines[headerEndIdx:],
						StartLine: section.StartLine + headerEndIdx,
					}
					// Trim the header section to just the header (and blank lines)
					section.Lines = section.Lines[:headerEndIdx]
				}

				// Create new TODO section after header
				newSection := &Section{
					Type:      SectionTODO,
					Lines:     []string{},
					TODOItems: []*todo.Item{},
					StartLine: section.StartLine + len(section.Lines),
				}

				// Create blank line section for proper formatting if there's content after
				var blankSection *Section
				if newOtherSection != nil || i+1 < len(d.Sections) {
					blankSection = &Section{
						Type:      SectionOther,
						Lines:     []string{""},
						StartLine: newSection.StartLine,
					}
				}

				// Insert new sections into the document
				if newOtherSection != nil {
					// Insert: header, TODO, blank, other content
					d.Sections = append(d.Sections[:i+1], append([]*Section{newSection, blankSection, newOtherSection}, d.Sections[i+1:]...)...)
				} else if blankSection != nil {
					// Insert: header, TODO, blank, (existing next sections)
					d.Sections = append(d.Sections[:i+1], append([]*Section{newSection, blankSection}, d.Sections[i+1:]...)...)
				} else {
					// Insert: header, TODO (nothing after)
					d.Sections = append(d.Sections[:i+1], append([]*Section{newSection}, d.Sections[i+1:]...)...)
				}
				targetSection = newSection
				break
			}
		}
	}

	if targetSection == nil {
		return nil, fmt.Errorf("could not find or create section for project %s", proj.Name)
	}

	// Add item to section
	targetSection.Lines = append(targetSection.Lines, item.String())
	targetSection.TODOItems = append(targetSection.TODOItems, item)
	d.Items = append(d.Items, item)

	return item, nil
}

// RemoveItem removes a TODO item from the document
func (d *Document) RemoveItem(item *todo.Item) error {
	// Find and remove from section
	for _, section := range d.Sections {
		if section.Type != SectionTODO {
			continue
		}

		for i, sectionItem := range section.TODOItems {
			if sectionItem == item {
				// Remove from TODOItems
				section.TODOItems = append(section.TODOItems[:i], section.TODOItems[i+1:]...)

				// Remove from Lines
				if i < len(section.Lines) {
					section.Lines = append(section.Lines[:i], section.Lines[i+1:]...)
				}

				// Remove from document items
				for j, docItem := range d.Items {
					if docItem == item {
						d.Items = append(d.Items[:j], d.Items[j+1:]...)
						break
					}
				}

				return nil
			}
		}
	}

	return fmt.Errorf("item not found in document")
}

// UpdateItem updates a TODO item's done status
func (d *Document) UpdateItem(item *todo.Item) error {
	// Find item in sections and update the line
	for _, section := range d.Sections {
		if section.Type != SectionTODO {
			continue
		}

		for i, sectionItem := range section.TODOItems {
			if sectionItem == item {
				// Update the line
				if i < len(section.Lines) {
					section.Lines[i] = item.String()
				}
				return nil
			}
		}
	}

	return fmt.Errorf("item not found in document")
}

// String returns the document as markdown text
func (d *Document) String() string {
	var parts []string
	for _, section := range d.Sections {
		parts = append(parts, section.String())
	}
	result := strings.Join(parts, "\n")
	// Ensure file ends with newline
	if len(result) > 0 && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result
}
