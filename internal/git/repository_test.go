package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRepository(t *testing.T) {
	// Test finding repository from current directory
	repo, err := FindRepository()
	if err != nil {
		t.Fatalf("FindRepository() error = %v", err)
	}

	if repo == nil {
		t.Fatal("FindRepository() returned nil repository")
	}

	if repo.Root == "" {
		t.Error("Repository root is empty")
	}

	// Verify .git directory exists
	gitDir := filepath.Join(repo.Root, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Errorf(".git directory not found at %s", gitDir)
	}
}

func TestFindRepositoryFrom(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "current directory",
			path:    ".",
			wantErr: false,
		},
		{
			name:    "temp directory (not in git)",
			path:    os.TempDir(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := FindRepositoryFrom(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Error("FindRepositoryFrom() expected error, got nil")
				}
				if err != ErrNotInGitRepo {
					t.Errorf("FindRepositoryFrom() expected ErrNotInGitRepo, got %v", err)
				}
				if repo != nil {
					t.Error("FindRepositoryFrom() expected nil repository on error")
				}
			} else {
				if err != nil {
					t.Errorf("FindRepositoryFrom() unexpected error: %v", err)
				}
				if repo == nil {
					t.Error("FindRepositoryFrom() returned nil repository")
				}
			}
		})
	}
}

func TestRepository_ConfigPath(t *testing.T) {
	repo := &Repository{Root: "/path/to/repo"}
	expected := filepath.Join("/path/to/repo", ".mktodo.yml")
	
	if repo.ConfigPath() != expected {
		t.Errorf("ConfigPath() = %s, want %s", repo.ConfigPath(), expected)
	}
}

func TestFindRepositoryNestedPath(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	nestedDir := filepath.Join(tmpDir, "nested", "path")

	// Create .git directory
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create nested directories
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Find repository from nested path
	repo, err := FindRepositoryFrom(nestedDir)
	if err != nil {
		t.Fatalf("FindRepositoryFrom() error = %v", err)
	}

	// Verify it found the correct root
	if repo.Root != tmpDir {
		t.Errorf("FindRepositoryFrom() root = %s, want %s", repo.Root, tmpDir)
	}
}
