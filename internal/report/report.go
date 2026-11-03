// Package report computes TODO statistics per project and renders them as
// text, markdown, HTML or JSON (spec FR-10).
package report

import (
	"math"
	"time"

	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/internal/project"
)

// Counts holds item counts for a project or the whole document.
type Counts struct {
	Total int `json:"total"`
	Open  int `json:"open"`
	Done  int `json:"done"`
	FIXME int `json:"fixme"`
}

func (c *Counts) add(item *markdown.Item) {
	c.Total++
	if item.Done {
		c.Done++
	} else {
		c.Open++
	}
	if item.IsFIXME {
		c.FIXME++
	}
}

// Percent returns n as a whole-number percentage of the total (0 when empty).
func (c Counts) Percent(n int) int {
	if c.Total == 0 {
		return 0
	}
	return int(math.Round(float64(n) * 100 / float64(c.Total)))
}

// Completion returns the share of done items as a whole-number percentage.
func (c Counts) Completion() int {
	return c.Percent(c.Done)
}

// ProjectRow is one project in the report, in hierarchy order. Counts cover
// the items directly under the project's header, not its children.
type ProjectRow struct {
	Name  string
	Path  string // dotted path, as used with -p
	Title string
	Depth int // 0 for root projects
	Counts
}

// FIXMEItem is a FIXME item and the project it belongs to.
type FIXMEItem struct {
	Project     *project.Project // nil when the item is outside any project
	Description string
	Done        bool
}

// Report is the data shared by all output formats.
type Report struct {
	Generated  time.Time
	Summary    Counts
	Projects   []ProjectRow
	Unassigned Counts // items under headers that aren't configured projects
	FIXMEItems []FIXMEItem
}

// Build computes the report for a document.
func Build(doc *markdown.Document, generated time.Time) *Report {
	r := &Report{Generated: generated}

	byProject := make(map[*project.Project]*Counts)
	for _, item := range doc.Items {
		r.Summary.add(item)

		counts := &r.Unassigned
		if item.Project != nil {
			if byProject[item.Project] == nil {
				byProject[item.Project] = &Counts{}
			}
			counts = byProject[item.Project]
		}
		counts.add(item)

		if item.IsFIXME {
			r.FIXMEItems = append(r.FIXMEItems, FIXMEItem{
				Project:     item.Project,
				Description: item.Description,
				Done:        item.Done,
			})
		}
	}

	var walk func(p *project.Project, depth int)
	walk = func(p *project.Project, depth int) {
		row := ProjectRow{Name: p.Name, Path: p.FullPath(), Title: p.Title, Depth: depth}
		if c := byProject[p]; c != nil {
			row.Counts = *c
		}
		r.Projects = append(r.Projects, row)
		for _, child := range p.Children {
			walk(child, depth+1)
		}
	}
	for _, root := range project.GetRootProjects(doc.Projects) {
		walk(root, 0)
	}

	return r
}
