package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		configFile  string
		wantErr     bool
		errContains []string
	}{
		{
			name:       "valid simple config",
			configFile: "simple.yml",
			wantErr:    false,
		},
		{
			name:       "valid nested config",
			configFile: "nested.yml",
			wantErr:    false,
		},
		{
			name:        "circular dependency",
			configFile:  "circular.yml",
			wantErr:     true,
			errContains: []string{"circular dependency"},
		},
		{
			name:        "duplicate names",
			configFile:  "duplicate.yml",
			wantErr:     true,
			errContains: []string{"duplicate project name"},
		},
		{
			name:        "deep nesting",
			configFile:  "deep_nesting.yml",
			wantErr:     true,
			errContains: []string{"nesting depth", "exceeds maximum"},
		},
		{
			name:        "missing fields",
			configFile:  "missing_fields.yml",
			wantErr:     true,
			errContains: []string{"required"},
		},
		{
			name:        "invalid parent reference",
			configFile:  "invalid_parent.yml",
			wantErr:     true,
			errContains: []string{"parent", "does not exist"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("../../testdata/configs", tt.configFile)
			cfg, err := Load(path)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("Load() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			err = cfg.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("Validate() expected error, got nil")
					return
				}

				// Check if error contains expected strings
				errStr := err.Error()
				for _, expected := range tt.errContains {
					if !strings.Contains(errStr, expected) {
						t.Errorf("Validate() error = %v, should contain %q", err, expected)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidate_EmptyProjects(t *testing.T) {
	cfg := &Config{
		TodoFile: "README.md",
		Projects: []ProjectConfig{},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() expected error for empty projects, got nil")
	}
	if !strings.Contains(err.Error(), "at least one project") {
		t.Errorf("Validate() error = %v, should mention at least one project", err)
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg := &Config{
		TodoFile: "README.md",
		Projects: []ProjectConfig{
			{Name: "", Title: "No Name", Parent: nil},
			{Name: "dup", Title: "Duplicate", Parent: nil},
			{Name: "dup", Title: "Duplicate", Parent: nil},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}

	errStr := err.Error()
	// Should report multiple errors
	if !strings.Contains(errStr, "name is required") {
		t.Error("Should report missing name")
	}
	if !strings.Contains(errStr, "duplicate") {
		t.Error("Should report duplicate name")
	}
}

func TestCheckCircularDependency(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "no circular dependency",
			config: &Config{
				Projects: []ProjectConfig{
					{Name: "parent", Title: "Parent", Parent: nil},
					{Name: "child", Title: "Child", Parent: stringPtr("parent")},
				},
			},
			wantErr: false,
		},
		{
			name: "direct circular dependency",
			config: &Config{
				Projects: []ProjectConfig{
					{Name: "a", Title: "A", Parent: stringPtr("b")},
					{Name: "b", Title: "B", Parent: stringPtr("a")},
				},
			},
			wantErr: true,
		},
		{
			name: "indirect circular dependency",
			config: &Config{
				Projects: []ProjectConfig{
					{Name: "a", Title: "A", Parent: stringPtr("b")},
					{Name: "b", Title: "B", Parent: stringPtr("c")},
					{Name: "c", Title: "C", Parent: stringPtr("a")},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("Validate() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCalculateDepth(t *testing.T) {
	cfg := &Config{
		Projects: []ProjectConfig{
			{Name: "level1", Title: "Level 1", Parent: nil},
			{Name: "level2", Title: "Level 2", Parent: stringPtr("level1")},
			{Name: "level3", Title: "Level 3", Parent: stringPtr("level2")},
		},
	}

	projectMap := make(map[string]*ProjectConfig)
	for i := range cfg.Projects {
		projectMap[cfg.Projects[i].Name] = &cfg.Projects[i]
	}

	tests := []struct {
		name          string
		projectName   string
		expectedDepth int
	}{
		{"level 1 project", "level1", 1},
		{"level 2 project", "level2", 2},
		{"level 3 project", "level3", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			depth := cfg.calculateDepth(tt.projectName, projectMap)
			if depth != tt.expectedDepth {
				t.Errorf("calculateDepth() = %d, want %d", depth, tt.expectedDepth)
			}
		})
	}
}

func TestValidate_NilParentHandling(t *testing.T) {
	nilStr := "nil"
	cfg := &Config{
		TodoFile: "README.md",
		Projects: []ProjectConfig{
			{Name: "root1", Title: "Root 1", Parent: nil},
			{Name: "root2", Title: "Root 2", Parent: &nilStr},
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() unexpected error for nil parent: %v", err)
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
