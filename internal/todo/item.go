package todo

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jvzantvoort/mktodo/internal/project"
)

// Item represents a single TODO item
type Item struct {
	Description string
	Done        bool
	IsFIXME     bool
	LineNumber  int
	Project     *project.Project
}

var (
	// Regex patterns for parsing TODO items
	todoPattern = regexp.MustCompile(`^[\s]*[-*+]\s*\[([ xX])\]\s*(.*)$`)
)

// Parse parses a line and returns a TODO item if it matches the pattern
func Parse(line string) (*Item, error) {
	matches := todoPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil, nil // Not a TODO line
	}

	item := &Item{
		Done:        strings.ToLower(matches[1]) == "x",
		Description: strings.TrimSpace(matches[2]),
	}

	// Check for FIXME prefix (case-insensitive)
	if strings.HasPrefix(strings.ToUpper(item.Description), "FIXME") {
		item.IsFIXME = true
	}

	return item, nil
}

// String returns the markdown representation of the item
func (i *Item) String() string {
	checkbox := " "
	if i.Done {
		checkbox = "X"
	}
	return fmt.Sprintf("- [%s] %s", checkbox, i.Description)
}

// MarkDone toggles the done status
func (i *Item) MarkDone() {
	i.Done = !i.Done
}

// Matches checks if the item description matches the given query
func (i *Item) Matches(query string) bool {
	return strings.Contains(strings.ToLower(i.Description), strings.ToLower(query))
}
