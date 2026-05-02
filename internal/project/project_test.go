package project

import (
	"testing"

	"github.com/jvzantvoort/mktodo/internal/config"
)

func stringPtr(s string) *string {
	return &s
}

func TestBuildHierarchy(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
		checks  func(*testing.T, map[string]*Project)
	}{
		{
			name: "simple flat hierarchy",
			config: &config.Config{
				Projects: []config.ProjectConfig{
					{Name: "default", Title: "TODO", Parent: nil},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, projects map[string]*Project) {
				if len(projects) != 1 {
					t.Errorf("expected 1 project, got %d", len(projects))
				}
				proj := projects["default"]
				if proj == nil {
					t.Fatal("project 'default' not found")
				}
				if proj.Level != 1 {
					t.Errorf("expected level 1, got %d", proj.Level)
				}
				if proj.Parent != nil {
					t.Error("expected no parent")
				}
			},
		},
		{
			name: "nested hierarchy",
			config: &config.Config{
				Projects: []config.ProjectConfig{
					{Name: "lego", Title: "Lego", Parent: nil},
					{Name: "technic", Title: "Technic", Parent: stringPtr("lego")},
					{Name: "city", Title: "City", Parent: stringPtr("lego")},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, projects map[string]*Project) {
				if len(projects) != 3 {
					t.Errorf("expected 3 projects, got %d", len(projects))
				}

				lego := projects["lego"]
				if lego.Level != 1 {
					t.Errorf("lego: expected level 1, got %d", lego.Level)
				}
				if len(lego.Children) != 2 {
					t.Errorf("lego: expected 2 children, got %d", len(lego.Children))
				}

				technic := projects["technic"]
				if technic.Level != 2 {
					t.Errorf("technic: expected level 2, got %d", technic.Level)
				}
				if technic.Parent != lego {
					t.Error("technic: parent should be lego")
				}

				city := projects["city"]
				if city.Level != 2 {
					t.Errorf("city: expected level 2, got %d", city.Level)
				}
				if city.Parent != lego {
					t.Error("city: parent should be lego")
				}
			},
		},
		{
			name: "deep hierarchy",
			config: &config.Config{
				Projects: []config.ProjectConfig{
					{Name: "l1", Title: "Level 1", Parent: nil},
					{Name: "l2", Title: "Level 2", Parent: stringPtr("l1")},
					{Name: "l3", Title: "Level 3", Parent: stringPtr("l2")},
					{Name: "l4", Title: "Level 4", Parent: stringPtr("l3")},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, projects map[string]*Project) {
				levels := map[string]int{"l1": 1, "l2": 2, "l3": 3, "l4": 4}
				for name, expectedLevel := range levels {
					proj := projects[name]
					if proj.Level != expectedLevel {
						t.Errorf("%s: expected level %d, got %d", name, expectedLevel, proj.Level)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projects, err := BuildHierarchy(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.checks != nil {
				tt.checks(t, projects)
			}
		})
	}
}

func TestFindByPath(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "lego", Title: "Lego", Parent: nil},
			{Name: "technic", Title: "Technic", Parent: stringPtr("lego")},
			{Name: "vehicles", Title: "Vehicles", Parent: stringPtr("technic")},
			{Name: "city", Title: "City", Parent: stringPtr("lego")},
		},
	}

	projects, err := BuildHierarchy(cfg)
	if err != nil {
		t.Fatalf("BuildHierarchy failed: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
		check   func(*testing.T, *Project)
	}{
		{
			name:    "find root project",
			path:    "lego",
			wantErr: false,
			check: func(t *testing.T, proj *Project) {
				if proj.Name != "lego" {
					t.Errorf("expected 'lego', got '%s'", proj.Name)
				}
			},
		},
		{
			name:    "find nested project",
			path:    "lego.technic",
			wantErr: false,
			check: func(t *testing.T, proj *Project) {
				if proj.Name != "technic" {
					t.Errorf("expected 'technic', got '%s'", proj.Name)
				}
			},
		},
		{
			name:    "find deeply nested project",
			path:    "lego.technic.vehicles",
			wantErr: false,
			check: func(t *testing.T, proj *Project) {
				if proj.Name != "vehicles" {
					t.Errorf("expected 'vehicles', got '%s'", proj.Name)
				}
			},
		},
		{
			name:    "nonexistent root",
			path:    "nonexistent",
			wantErr: true,
		},
		{
			name:    "nonexistent child",
			path:    "lego.nonexistent",
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj, err := FindByPath(projects, tt.path)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, proj)
			}
		})
	}
}

func TestGetRootProjects(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "root1", Title: "Root 1", Parent: nil},
			{Name: "root2", Title: "Root 2", Parent: nil},
			{Name: "child1", Title: "Child 1", Parent: stringPtr("root1")},
			{Name: "child2", Title: "Child 2", Parent: stringPtr("root2")},
		},
	}

	projects, err := BuildHierarchy(cfg)
	if err != nil {
		t.Fatalf("BuildHierarchy failed: %v", err)
	}

	roots := GetRootProjects(projects)
	if len(roots) != 2 {
		t.Errorf("expected 2 root projects, got %d", len(roots))
	}

	// Check that all roots have no parent
	for _, root := range roots {
		if root.Parent != nil {
			t.Errorf("root project '%s' has parent", root.Name)
		}
	}
}

func TestProject_FullPath(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "lego", Title: "Lego", Parent: nil},
			{Name: "technic", Title: "Technic", Parent: stringPtr("lego")},
			{Name: "vehicles", Title: "Vehicles", Parent: stringPtr("technic")},
		},
	}

	projects, err := BuildHierarchy(cfg)
	if err != nil {
		t.Fatalf("BuildHierarchy failed: %v", err)
	}

	tests := []struct {
		name         string
		projectName  string
		expectedPath string
	}{
		{"root project", "lego", "lego"},
		{"child project", "technic", "lego.technic"},
		{"grandchild project", "vehicles", "lego.technic.vehicles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj := projects[tt.projectName]
			if proj == nil {
				t.Fatalf("project '%s' not found", tt.projectName)
			}
			path := proj.FullPath()
			if path != tt.expectedPath {
				t.Errorf("expected path '%s', got '%s'", tt.expectedPath, path)
			}
		})
	}
}

func TestCalculateLevel(t *testing.T) {
	// Create a simple hierarchy manually
	root := &Project{Name: "root", Title: "Root", Level: 0}
	child := &Project{Name: "child", Title: "Child", Parent: root, Level: 0}
	grandchild := &Project{Name: "grandchild", Title: "Grandchild", Parent: child, Level: 0}

	tests := []struct {
		name          string
		project       *Project
		expectedLevel int
	}{
		{"root level", root, 1},
		{"child level", child, 2},
		{"grandchild level", grandchild, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := calculateLevel(tt.project)
			if level != tt.expectedLevel {
				t.Errorf("expected level %d, got %d", tt.expectedLevel, level)
			}
		})
	}
}
