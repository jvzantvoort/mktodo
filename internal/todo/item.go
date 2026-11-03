// Package todo defines TODO items and their parsing.
package todo

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jvzantvoort/mktodo/internal/project"
)

// Item represents a single TODO item.
//
// An item parsed from a file remembers its original line, so String()
// reproduces it byte for byte while Done and Description are unchanged, and
// only the checkbox or description is rewritten when they change. Indentation,
// the list bullet and the checkbox case are always kept.
type Item struct {
	Description string
	Done        bool
	IsFIXME     bool
	LineNumber  int
	Project     *project.Project

	raw      string // original line, empty for new items
	prefix   string // indentation, bullet and "[" (e.g. "  * [")
	mark     string // original checkbox character
	sep      string // "]" plus following whitespace
	origDone bool
	origDesc string
}

var (
	// Regex patterns for parsing TODO items
	todoPattern = regexp.MustCompile(`^(\s*[-*+]\s*\[)([ xX])(\]\s*)(.*)$`)
)

// New creates a new open TODO item with the given description.
func New(description string) *Item {
	return &Item{Description: description, IsFIXME: isFIXME(description)}
}

func isFIXME(description string) bool {
	return strings.HasPrefix(strings.ToUpper(description), "FIXME")
}

// SetDescription changes the description and updates the FIXME flag.
func (i *Item) SetDescription(description string) {
	i.Description = description
	i.IsFIXME = isFIXME(description)
}

// Parse parses a line and returns a TODO item if it matches the pattern
func Parse(line string) (*Item, error) {
	matches := todoPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil, nil // Not a TODO line
	}

	description := strings.TrimSpace(matches[4])
	done := strings.ToLower(matches[2]) == "x"
	item := &Item{
		Done:        done,
		Description: description,
		IsFIXME:     isFIXME(description),
		raw:         line,
		prefix:      matches[1],
		mark:        matches[2],
		sep:         matches[3],
		origDone:    done,
		origDesc:    description,
	}

	return item, nil
}

// String returns the markdown representation of the item
func (i *Item) String() string {
	if i.raw != "" && i.Done == i.origDone && i.Description == i.origDesc {
		return i.raw
	}

	prefix, sep := i.prefix, i.sep
	if prefix == "" {
		prefix, sep = "- [", "] "
	}
	if !strings.HasSuffix(sep, " ") && !strings.HasSuffix(sep, "\t") && i.Description != "" {
		sep += " "
	}

	checkbox := " "
	if i.Done {
		checkbox = "X"
		if i.mark == "x" {
			checkbox = "x" // keep the file's own style
		}
	}
	return fmt.Sprintf("%s%s%s%s", prefix, checkbox, sep, i.Description)
}

// MarkDone toggles the done status
func (i *Item) MarkDone() {
	i.Done = !i.Done
}

// Matches checks if the item description matches the given query
func (i *Item) Matches(query string) bool {
	return strings.Contains(strings.ToLower(i.Description), strings.ToLower(query))
}
