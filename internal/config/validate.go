package config

import (
	"fmt"
	"sort"
	"strings"
)

// MaxDepth is the deepest nesting allowed: markdown has six header levels.
const MaxDepth = 6

// ValidationError lists everything wrong with a configuration, one
// problem per line, in the format of the specification (FR-2):
//
//   - Project 'technic' references non-existent parent 'lego'
//   - Circular reference detected: project1 -> project2 -> project1
type ValidationError struct {
	Problems []string
}

func (e *ValidationError) Error() string {
	return "- " + strings.Join(e.Problems, "\n- ")
}

func (e *ValidationError) add(format string, args ...interface{}) {
	e.Problems = append(e.Problems, fmt.Sprintf(format, args...))
}

func (e *ValidationError) orNil() error {
	if len(e.Problems) == 0 {
		return nil
	}
	return e
}

// hasParent reports whether a project has a parent ("nil" means none).
func (p ProjectConfig) hasParent() bool {
	return p.Parent != nil && *p.Parent != "nil"
}

// Validate checks the configuration and returns a *ValidationError listing
// every problem found.
func (c *Config) Validate() error {
	problems := &ValidationError{}
	if len(c.Projects) == 0 {
		problems.add("At least one project must be defined")
		return problems
	}

	projectMap := c.checkNames(problems)

	for _, proj := range c.Projects {
		if proj.hasParent() && projectMap[*proj.Parent] == nil {
			problems.add("Project '%s' references non-existent parent '%s'", proj.Name, *proj.Parent)
		}
	}

	cycles := c.findCycles(projectMap)
	for _, cycle := range cycles {
		problems.add("Circular reference detected: %s", strings.Join(cycle, " -> "))
	}

	if len(cycles) == 0 {
		for _, proj := range c.Projects {
			if depth := c.calculateDepth(proj.Name, projectMap); depth > MaxDepth {
				problems.add("Project '%s' is nested %d levels deep; markdown headers allow at most %d", proj.Name, depth, MaxDepth)
			}
		}
	}

	return problems.orNil()
}

// checkNames reports missing names and titles and duplicate names, and
// returns the projects by name.
func (c *Config) checkNames(problems *ValidationError) map[string]*ProjectConfig {
	projectMap := make(map[string]*ProjectConfig)
	for i, proj := range c.Projects {
		if proj.Name == "" {
			problems.add("Project at position %d has no name", i+1)
			continue
		}
		if proj.Title == "" {
			problems.add("Project '%s' has no title", proj.Name)
		}
		if projectMap[proj.Name] != nil {
			problems.add("Duplicate project name '%s'", proj.Name)
			continue
		}
		projectMap[proj.Name] = &c.Projects[i]
	}
	return projectMap
}

// findCycles returns each parent cycle once, as the path around it,
// starting and ending at its alphabetically first project, e.g.
// ["a", "b", "a"].
func (c *Config) findCycles(projectMap map[string]*ProjectConfig) [][]string {
	seen := make(map[string]bool)
	var cycles [][]string

	for _, proj := range c.Projects {
		var path []string
		index := make(map[string]int)
		for name := proj.Name; name != ""; {
			if i, ok := index[name]; ok {
				if cycle := rotateCycle(path[i:]); !seen[cycle[0]] {
					seen[cycle[0]] = true
					cycles = append(cycles, append(cycle, cycle[0]))
				}
				break
			}
			index[name] = len(path)
			path = append(path, name)

			p := projectMap[name]
			if p == nil || !p.hasParent() {
				break
			}
			name = *p.Parent
		}
	}
	return cycles
}

// rotateCycle starts a cycle at its alphabetically first member, so the
// same cycle always reads the same way.
func rotateCycle(cycle []string) []string {
	sorted := append([]string(nil), cycle...)
	sort.Strings(sorted)
	for i, name := range cycle {
		if name == sorted[0] {
			return append(append([]string(nil), cycle[i:]...), cycle[:i]...)
		}
	}
	return cycle
}

// calculateDepth calculates the nesting depth of a project
func (c *Config) calculateDepth(name string, projectMap map[string]*ProjectConfig) int {
	depth := 1
	current := name
	visited := make(map[string]bool)

	for !visited[current] {
		visited[current] = true
		proj, ok := projectMap[current]
		if !ok || !proj.hasParent() {
			return depth
		}
		depth++
		current = *proj.Parent
	}
	// Circular dependency: report a depth that is always too deep.
	return 999
}
