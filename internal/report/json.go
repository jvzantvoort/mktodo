package report

import (
	"encoding/json"
	"io"
	"time"
)

type jsonCounts struct {
	Counts
	Completion int `json:"completion"` // percentage of items done
}

type jsonProject struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	Title string `json:"title"`
	Depth int    `json:"depth"`
	jsonCounts
}

type jsonFIXME struct {
	Project     string `json:"project,omitempty"` // dotted path; empty outside any project
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

type jsonReport struct {
	Generated  string        `json:"generated"`
	Summary    jsonCounts    `json:"summary"`
	Projects   []jsonProject `json:"projects"`
	Unassigned *jsonCounts   `json:"unassigned,omitempty"`
	FIXMEItems []jsonFIXME   `json:"fixme_items"`
}

func toJSONCounts(c Counts) jsonCounts {
	return jsonCounts{Counts: c, Completion: c.Completion()}
}

// WriteJSON renders the report as indented JSON, for scripts.
func WriteJSON(w io.Writer, r *Report) error {
	out := jsonReport{
		Generated:  r.Generated.Format(time.RFC3339),
		Summary:    toJSONCounts(r.Summary),
		Projects:   []jsonProject{},
		FIXMEItems: []jsonFIXME{},
	}
	for _, p := range r.Projects {
		out.Projects = append(out.Projects, jsonProject{
			Name: p.Name, Path: p.Path, Title: p.Title, Depth: p.Depth, jsonCounts: toJSONCounts(p.Counts),
		})
	}
	if r.Unassigned.Total > 0 {
		c := toJSONCounts(r.Unassigned)
		out.Unassigned = &c
	}
	for _, item := range r.FIXMEItems {
		f := jsonFIXME{Description: item.Description, Done: item.Done}
		if item.Project != nil {
			f.Project = item.Project.FullPath()
		}
		out.FIXMEItems = append(out.FIXMEItems, f)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
