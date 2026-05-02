package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	// Use git rev-parse to find the repository root
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = absPath

	output, err := cmd.Output()
	if err != nil {
		// Not in a git repository or git command failed
		return nil, ErrNotInGitRepo
	}

	root := strings.TrimSpace(string(output))
	if root == "" {
		return nil, ErrNotInGitRepo
	}

	return &Repository{Root: root}, nil
}

// ConfigPath returns the path to the .mktodo.yml configuration file
func (r *Repository) ConfigPath() string {
	return filepath.Join(r.Root, ".mktodo.yml")
}

// ResolvePath resolves a relative path from the repository root
func (r *Repository) ResolvePath(relPath string) string {
	return filepath.Join(r.Root, relPath)
}
