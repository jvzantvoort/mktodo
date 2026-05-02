package project

import (
	"fmt"
	"strings"

	"github.com/jvzantvoort/mktodo/internal/config"
)

// Project represents a project in the hierarchy
type Project struct {
	Name     string
	Title    string
	Parent   *Project
	Children []*Project
	Level    int // Header level (1-6)
}

// BuildHierarchy builds the project hierarchy from configuration
func BuildHierarchy(cfg *config.Config) (map[string]*Project, error) {
	projects := make(map[string]*Project)

	// First pass: create all projects
	for _, projCfg := range cfg.Projects {
		projects[projCfg.Name] = &Project{
			Name:     projCfg.Name,
			Title:    projCfg.Title,
			Children: make([]*Project, 0),
		}
	}

	// Second pass: link parents and children
	for _, projCfg := range cfg.Projects {
		proj := projects[projCfg.Name]
		
		if projCfg.Parent != nil && *projCfg.Parent != "nil" {
			parent, ok := projects[*projCfg.Parent]
			if !ok {
				return nil, fmt.Errorf("parent '%s' not found for project '%s'", *projCfg.Parent, projCfg.Name)
			}
			proj.Parent = parent
			parent.Children = append(parent.Children, proj)
		}
	}

	// Third pass: calculate levels
	for _, proj := range projects {
		proj.Level = calculateLevel(proj)
	}

	return projects, nil
}

// calculateLevel calculates the header level for a project (1-6)
func calculateLevel(proj *Project) int {
	level := 1
	current := proj

	for current.Parent != nil {
		level++
		current = current.Parent
	}

	return level
}

// FindByPath finds a project by dotted path notation (e.g., "lego.technic")
func FindByPath(projects map[string]*Project, path string) (*Project, error) {
	if path == "" {
		return nil, fmt.Errorf("empty project path")
	}

	// Split path by dots
	parts := strings.Split(path, ".")
	
	// Find the first part
	current, ok := projects[parts[0]]
	if !ok {
		return nil, fmt.Errorf("project '%s' not found", parts[0])
	}

	// Navigate through children
	for i := 1; i < len(parts); i++ {
		found := false
		for _, child := range current.Children {
			if child.Name == parts[i] {
				current = child
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("project '%s' not found in path '%s'", parts[i], path)
		}
	}

	return current, nil
}

// GetRootProjects returns all projects without parents
func GetRootProjects(projects map[string]*Project) []*Project {
	roots := make([]*Project, 0)
	for _, proj := range projects {
		if proj.Parent == nil {
			roots = append(roots, proj)
		}
	}
	return roots
}

// FullPath returns the full dotted path to this project
func (p *Project) FullPath() string {
	if p.Parent == nil {
		return p.Name
	}
	return p.Parent.FullPath() + "." + p.Name
}
