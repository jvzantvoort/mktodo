package project

import (
	"fmt"
	"testing"

	"github.com/jvzantvoort/mktodo/internal/config"
)

// BenchmarkBuildHierarchy benchmarks project hierarchy building
func BenchmarkBuildHierarchy(b *testing.B) {
	nilVal := "nil"
	parent1 := "project1"
	parent2 := "project2"
	
	cfg := &config.Config{
		TodoFile: "README.md",
		Projects: []config.ProjectConfig{
			{Name: "default", Title: "TODO", Parent: &nilVal},
			{Name: "project1", Title: "Project 1", Parent: &nilVal},
			{Name: "project2", Title: "Project 2", Parent: &parent1},
			{Name: "project3", Title: "Project 3", Parent: &parent2},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := BuildHierarchy(cfg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFindByPath benchmarks path-based project lookup
func BenchmarkFindByPath(b *testing.B) {
	nilVal := "nil"
	parent1 := "project1"
	parent2 := "project2"
	
	cfg := &config.Config{
		TodoFile: "README.md",
		Projects: []config.ProjectConfig{
			{Name: "default", Title: "TODO", Parent: &nilVal},
			{Name: "project1", Title: "Project 1", Parent: &nilVal},
			{Name: "project2", Title: "Project 2", Parent: &parent1},
			{Name: "project3", Title: "Project 3", Parent: &parent2},
		},
	}

	hierarchy, err := BuildHierarchy(cfg)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := FindByPath(hierarchy, "project1.project2.project3")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkComplexHierarchy benchmarks building complex hierarchies
func BenchmarkComplexHierarchy(b *testing.B) {
	// Create a complex hierarchy with many projects
	nilVal := "nil"
	defaultVal := "default"
	
	projects := []config.ProjectConfig{
		{Name: "default", Title: "TODO", Parent: &nilVal},
	}
	
	// Add 20 top-level projects
	for i := 0; i < 20; i++ {
		name := fmt.Sprintf("project%c", rune('A'+i))
		parentCopy := defaultVal
		projects = append(projects, config.ProjectConfig{
			Name:   name,
			Title:  fmt.Sprintf("Project %c", rune('A'+i)),
			Parent: &parentCopy,
		})
	}
	
	cfg := &config.Config{
		TodoFile: "README.md",
		Projects: projects,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := BuildHierarchy(cfg)
		if err != nil {
			b.Fatal(err)
		}
	}
}
