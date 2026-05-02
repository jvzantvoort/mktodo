package config

import (
	"fmt"
	"strings"
)

// Validate validates the configuration
func (c *Config) Validate() error {
	var errors []string

	// Check if we have at least one project
	if len(c.Projects) == 0 {
		errors = append(errors, "at least one project must be defined")
	}

	// Track project names for duplicate detection
	projectNames := make(map[string]bool)

	// Track parent references for validation
	projectMap := make(map[string]*ProjectConfig)

	for i, proj := range c.Projects {
		// Check required fields
		if proj.Name == "" {
			errors = append(errors, fmt.Sprintf("project at index %d: name is required", i))
		}
		if proj.Title == "" {
			errors = append(errors, fmt.Sprintf("project '%s': title is required", proj.Name))
		}

		// Check for duplicate names
		if proj.Name != "" {
			if projectNames[proj.Name] {
				errors = append(errors, fmt.Sprintf("duplicate project name: '%s'", proj.Name))
			}
			projectNames[proj.Name] = true
			projectMap[proj.Name] = &c.Projects[i]
		}
	}

	// Validate parent references
	for _, proj := range c.Projects {
		if proj.Parent != nil && *proj.Parent != "nil" {
			// Check if parent exists
			if !projectNames[*proj.Parent] {
				errors = append(errors, fmt.Sprintf("project '%s': parent '%s' does not exist", proj.Name, *proj.Parent))
			}
		}
	}

	// Check for circular dependencies
	for name := range projectMap {
		if err := c.hasCircularDependency(name, projectMap); err != nil {
			errors = append(errors, err.Error())
			break // Only report first circular dependency found
		}
	}

	// Validate nesting depth
	for _, proj := range c.Projects {
		depth := c.calculateDepth(proj.Name, projectMap)
		if depth > 6 {
			errors = append(errors, fmt.Sprintf("project '%s': nesting depth %d exceeds maximum of 6", proj.Name, depth))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// hasCircularDependency checks if a project has circular parent references
func (c *Config) hasCircularDependency(name string, projectMap map[string]*ProjectConfig) error {
	visited := make(map[string]bool)
	current := name
	maxIterations := len(projectMap) + 1 // Safety limit

	// Follow parent chain
	for i := 0; i < maxIterations; i++ {
		if visited[current] {
			return fmt.Errorf("circular dependency detected involving project '%s'", name)
		}
		
		visited[current] = true
		
		proj, ok := projectMap[current]
		if !ok || proj.Parent == nil || *proj.Parent == "nil" {
			// Reached root or invalid reference
			return nil
		}
		
		current = *proj.Parent
	}
	
	// If we hit max iterations, there must be a cycle
	return fmt.Errorf("circular dependency detected involving project '%s'", name)
}

// calculateDepth calculates the nesting depth of a project
func (c *Config) calculateDepth(name string, projectMap map[string]*ProjectConfig) int {
	depth := 1
	current := name
	visited := make(map[string]bool)
	maxIterations := len(projectMap) + 1

	for i := 0; i < maxIterations; i++ {
		if visited[current] {
			// Circular dependency - return large depth to trigger error
			return 999
		}
		
		visited[current] = true
		
		proj, ok := projectMap[current]
		if !ok || proj.Parent == nil || *proj.Parent == "nil" {
			break
		}
		depth++
		current = *proj.Parent
	}

	return depth
}
