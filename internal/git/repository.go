package git

import (
	"errors"
	"os"
	"path/filepath"
)

var (
	// ErrNotInGitRepo is returned when not in a git repository
	ErrNotInGitRepo = errors.New("not in a git repository")
)

// Repository represents a git repository
type Repository struct {
	Root string // Absolute path to git repository root
}

// FindRepository finds the git repository root from the current directory
func FindRepository() (*Repository, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return FindRepositoryFrom(cwd)
}

// FindRepositoryFrom finds the git repository root from the specified directory
func FindRepositoryFrom(startPath string) (*Repository, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return nil, err
	}

	current := absPath
	for {
		gitDir := filepath.Join(current, ".git")
		if info, err := os.Stat(gitDir); err == nil {
			// Check if .git is a directory (not a file for submodules)
			if info.IsDir() {
				return &Repository{Root: current}, nil
			}
		}

		// Move to parent directory
		parent := filepath.Dir(current)
		
		// Stop if we've reached the root
		if parent == current {
			break
		}
		
		current = parent
	}

	return nil, ErrNotInGitRepo
}

// ConfigPath returns the path to the .mktodo.yml configuration file
func (r *Repository) ConfigPath() string {
	return filepath.Join(r.Root, ".mktodo.yml")
}

// ResolvePath resolves a relative path from the repository root
func (r *Repository) ResolvePath(relPath string) string {
	return filepath.Join(r.Root, relPath)
}
