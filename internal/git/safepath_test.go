package git

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// symlink creates a symlink or skips the test where that isn't allowed
// (Windows without developer mode).
func symlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks not available: %v", err)
	}
}

func TestSafePath(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "repo")
	outside := filepath.Join(base, "outside")
	for _, dir := range []string{filepath.Join(root, "docs"), outside} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	repo := &Repository{Root: root}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"file in root", "TODO.md", false},
		{"nested file", "docs/TASKS.md", false},
		{"missing directories", "new/dir/TODO.md", false},
		{"dot-dot that stays inside", "docs/../TODO.md", false},
		{"parent directory", "../outside/TODO.md", true},
		{"deep escape", "docs/../../outside/TODO.md", true},
		{"absolute path", filepath.Join(outside, "TODO.md"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.SafePath(tt.path)
			if tt.wantErr {
				if !errors.Is(err, ErrOutsideRepo) {
					t.Errorf("SafePath(%q) = %q, %v; want ErrOutsideRepo", tt.path, got, err)
				}
				return
			}
			if err != nil || got != filepath.Join(root, tt.path) {
				t.Errorf("SafePath(%q) = %q, %v; want %q", tt.path, got, err, filepath.Join(root, tt.path))
			}
		})
	}
}

func TestSafePath_Symlinks(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "repo")
	outside := filepath.Join(base, "outside")
	for _, dir := range []string{filepath.Join(root, "docs"), outside} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "real.md"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	symlink(t, filepath.Join(outside, "x.md"), filepath.Join(root, "out-file.md"))
	symlink(t, outside, filepath.Join(root, "out-dir"))
	symlink(t, filepath.Join(root, "docs", "real.md"), filepath.Join(root, "in-file.md"))
	symlink(t, filepath.Join(root, "missing.md"), filepath.Join(root, "dangling.md"))

	// The repository root itself may be reached through a symlink (as /var
	// is on macOS); that must not count as leaving it.
	linkedRoot := filepath.Join(base, "linked-repo")
	symlink(t, root, linkedRoot)

	tests := []struct {
		name    string
		root    string
		path    string
		wantErr bool
	}{
		{"link to a file outside", root, "out-file.md", true},
		{"file under a linked dir outside", root, "out-dir/TODO.md", true},
		{"link to a file inside", root, "in-file.md", false},
		{"dangling link", root, "dangling.md", true},
		{"root reached through a link", linkedRoot, "docs/real.md", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := (&Repository{Root: tt.root}).SafePath(tt.path)
			if tt.wantErr != errors.Is(err, ErrOutsideRepo) || (!tt.wantErr && err != nil) {
				t.Errorf("SafePath(%q) error = %v, want outside-repo error: %v", tt.path, err, tt.wantErr)
			}
		})
	}
}
