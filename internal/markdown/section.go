package markdown

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/jvzantvoort/mktodo/internal/project"
	"github.com/jvzantvoort/mktodo/internal/todo"
)

// SectionType represents the type of content section
type SectionType int

const (
	SectionHeader SectionType = iota
	SectionTODO
	SectionOther
)

// Section represents a section of markdown content
type Section struct {
	Type      SectionType
	Level     int              // Header level (1-6), 0 for non-headers
	Title     string           // Header title
	Project   *project.Project // Associated project (for headers)
	Lines     []string         // Original lines (preserved exactly)
	TODOItems []*todo.Item     // Parsed TODO items (if SectionTODO)
	StartLine int              // Line number where section starts (1-indexed)
}

var (
	headerPattern = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
)

// ParseSections parses markdown content into sections
func ParseSections(content string) ([]*Section, error) {
	var sections []*Section
	var currentSection *Section

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Check if this is a header
		if matches := headerPattern.FindStringSubmatch(line); matches != nil {
			// Save current section if exists
			if currentSection != nil && len(currentSection.Lines) > 0 {
				sections = append(sections, currentSection)
			}

			// Create new header section
			currentSection = &Section{
				Type:      SectionHeader,
				Level:     len(matches[1]),
				Title:     strings.TrimSpace(matches[2]),
				Lines:     []string{line},
				StartLine: lineNum,
			}
			continue
		}

		// Check if this is a TODO item
		item, err := todo.Parse(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		if item != nil {
			// If current section is not a TODO section, save it and start new one
			if currentSection != nil && currentSection.Type != SectionTODO {
				sections = append(sections, currentSection)
				currentSection = nil
			}

			// Create or append to TODO section
			if currentSection == nil {
				currentSection = &Section{
					Type:      SectionTODO,
					Lines:     []string{},
					TODOItems: []*todo.Item{},
					StartLine: lineNum,
				}
			}

			item.LineNumber = lineNum
			currentSection.Lines = append(currentSection.Lines, line)
			currentSection.TODOItems = append(currentSection.TODOItems, item)
			continue
		}

		// Regular content
		if currentSection == nil {
			currentSection = &Section{
				Type:      SectionOther,
				Lines:     []string{line},
				StartLine: lineNum,
			}
		} else if currentSection.Type == SectionOther || currentSection.Type == SectionHeader {
			// Append to current section
			currentSection.Lines = append(currentSection.Lines, line)
		} else {
			// Previous section was TODO, start new other section
			if currentSection != nil {
				sections = append(sections, currentSection)
			}
			currentSection = &Section{
				Type:      SectionOther,
				Lines:     []string{line},
				StartLine: lineNum,
			}
		}
	}

	// Save last section
	if currentSection != nil && len(currentSection.Lines) > 0 {
		sections = append(sections, currentSection)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning error: %w", err)
	}

	return sections, nil
}

// AssociateProjects links sections to projects based on headers
func AssociateProjects(sections []*Section, projects map[string]*project.Project) error {
	var currentProject *project.Project

	for _, section := range sections {
		if section.Type == SectionHeader {
			// Try to find matching project
			found := false
			for _, proj := range projects {
				if proj.Level == section.Level && proj.Title == section.Title {
					section.Project = proj
					currentProject = proj
					found = true
					break
				}
			}
			// If no matching project found, reset current project
			if !found {
				currentProject = nil
			}
		} else if section.Type == SectionTODO && currentProject != nil {
			// Associate TODO items with current project
			for _, item := range section.TODOItems {
				item.Project = currentProject
			}
		}
	}

	return nil
}

// String returns the section as markdown text
func (s *Section) String() string {
	return strings.Join(s.Lines, "\n")
}
