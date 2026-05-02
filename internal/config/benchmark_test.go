package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkLoadConfig benchmarks configuration loading
func BenchmarkLoadConfig(b *testing.B) {
	tmpDir := b.TempDir()

	configFile := filepath.Join(tmpDir, ".mktodo.yml")
	configContent := `---
todofile: README.md
projects:
  - name: default
    title: TODO
    parent: nil
  - name: project1
    title: Project 1
    parent: nil
  - name: project2
    title: Project 2
    parent: project1
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Load(configFile)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidateConfig benchmarks configuration validation
func BenchmarkValidateConfig(b *testing.B) {
	parent1 := "nil"
	parent2 := "project1"
	parent3 := "project2"

	cfg := &Config{
		TodoFile: "README.md",
		Projects: []ProjectConfig{
			{Name: "default", Title: "TODO", Parent: &parent1},
			{Name: "project1", Title: "Project 1", Parent: &parent1},
			{Name: "project2", Title: "Project 2", Parent: &parent2},
			{Name: "project3", Title: "Project 3", Parent: &parent3},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := cfg.Validate()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkComplexConfigValidation benchmarks validation of complex configurations
func BenchmarkComplexConfigValidation(b *testing.B) {
	// Create a complex configuration with many projects
	nilVal := "nil"
	defaultVal := "default"

	projects := []ProjectConfig{
		{Name: "default", Title: "TODO", Parent: &nilVal},
	}

	for i := 0; i < 20; i++ {
		parentCopy := defaultVal
		projects = append(projects, ProjectConfig{
			Name:   fmt.Sprintf("project%c", rune('A'+i)),
			Title:  fmt.Sprintf("Project %c", rune('A'+i)),
			Parent: &parentCopy,
		})
	}

	cfg := &Config{
		TodoFile: "README.md",
		Projects: projects,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := cfg.Validate()
		if err != nil {
			b.Fatal(err)
		}
	}
}
