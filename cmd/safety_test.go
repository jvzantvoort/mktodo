package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdd_RejectsUnsafeDescriptions(t *testing.T) {
	const todo = "# TODO\n\n- [ ] one\n"
	tests := []struct {
		name       string
		desc       string
		wantStderr string
	}{
		{"newline injection", "evil\n# Injected heading\n- [x] fake done", "line break or control character"},
		{"empty", "", "Description is required"},
		{"only spaces", "   ", "Description is required"},
		{"too long", strings.Repeat("a", 501), "longer than 500 characters"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, cleanup := setupTestRepo(t)
			defer cleanup()
			withTodo(todo)(t, dir)

			code, stderr := run(t, "add", "-y", tt.desc)
			if code != ExitUsage || !strings.Contains(stderr, tt.wantStderr) {
				t.Errorf("exit %d, stderr %q; want %d and %q", code, stderr, ExitUsage, tt.wantStderr)
			}
			if got := readFile(t, filepath.Join(dir, "TODO.md")); got != todo {
				t.Errorf("TODO.md changed: %q", got)
			}
		})
	}
}

func TestTodoFile_OutsideRepoRefused(t *testing.T) {
	for _, todofile := range []string{"../outside/STOLEN.md", "docs/../../STOLEN.md"} {
		t.Run(todofile, func(t *testing.T) {
			dir, cleanup := setupTestRepo(t)
			defer cleanup()
			withConfig("todofile: "+todofile+"\nprojects:\n  - name: default\n    title: TODO\n    parent: nil\n")(t, dir)

			for _, args := range [][]string{{"add", "-y", "x"}, {"list"}, {"report"}} {
				code, stderr := run(t, args...)
				if code != ExitConfig || !strings.Contains(stderr, "is outside the git repository") {
					t.Errorf("%v: exit %d, stderr %q; want %d", args, code, stderr, ExitConfig)
				}
			}
			if _, err := os.Stat(filepath.Join(dir, "..", "STOLEN.md")); !os.IsNotExist(err) {
				t.Error("a file was written outside the repository")
			}
		})
	}
}

func TestTodoFile_SymlinkOutsideRepoRefused(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	outside := filepath.Join(t.TempDir(), "elsewhere.md")
	writeFile(t, outside, "# TODO\n")
	if err := os.Symlink(outside, filepath.Join(dir, "TODO.md")); err != nil {
		t.Skipf("symlinks not available: %v", err)
	}

	code, stderr := run(t, "add", "-y", "x")
	if code != ExitConfig || !strings.Contains(stderr, "is outside the git repository") {
		t.Errorf("exit %d, stderr %q; want %d", code, stderr, ExitConfig)
	}
	if got := readFile(t, outside); got != "# TODO\n" {
		t.Errorf("file outside the repository was changed: %q", got)
	}
}
