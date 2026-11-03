// Package git locates the enclosing git repository and resolves paths within it.
package git

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	// ErrNotInGitRepo is returned when not in a git repository
	ErrNotInGitRepo = errors.New("not in a git repository")

	// ErrOutsideRepo is returned by SafePath for a path that leads outside
	// the repository.
	ErrOutsideRepo = errors.New("path is outside the git repository")
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

// SafePath resolves a path relative to the repository root and checks that
// it stays inside the repository, also after following symlinks. It
// returns the joined path; ErrOutsideRepo means the path is absolute, uses
// ".." to leave the repository, or goes through a symlink that points out
// of it. The file itself doesn't need to exist.
func (r *Repository) SafePath(relPath string) (string, error) {
	if filepath.IsAbs(relPath) || filepath.VolumeName(relPath) != "" {
		return "", ErrOutsideRepo
	}

	root := filepath.Clean(r.Root)
	target := filepath.Join(root, relPath)
	if !within(root, target) {
		return "", ErrOutsideRepo
	}

	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	realTarget, err := evalExisting(target)
	if err != nil {
		return "", err
	}
	if !within(realRoot, realTarget) {
		return "", ErrOutsideRepo
	}
	return target, nil
}

// within reports whether path is base or inside it.
func within(base, path string) bool {
	rel, err := filepath.Rel(base, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// evalExisting follows symlinks in the longest existing prefix of path and
// appends the part that doesn't exist yet.
func evalExisting(path string) (string, error) {
	var missing []string
	current := path
	for {
		if _, err := os.Lstat(current); err == nil {
			resolved, err := filepath.EvalSymlinks(current)
			if errors.Is(err, fs.ErrNotExist) {
				// A dangling symlink can't be checked, so it isn't trusted.
				return "", ErrOutsideRepo
			}
			if err != nil {
				return "", err
			}
			for i := len(missing) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, missing[i])
			}
			return resolved, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return path, nil
		}
		missing = append(missing, filepath.Base(current))
		current = parent
	}
}
