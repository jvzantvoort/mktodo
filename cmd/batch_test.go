package cmd

import (
	"path/filepath"
	"strings"
	"testing"
)

const batchTodo = "# TODO\n\n- [ ] Build crane\n- [x] Build crane model\n- [ ] Paint wall\n"

func TestDone_All(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()
	withTodo(batchTodo)(t, dir)

	// Run twice: --all marks done and never toggles back.
	for i := 0; i < 2; i++ {
		if code, stderr := run(t, "done", "--all", "crane"); code != ExitOK {
			t.Fatalf("run %d: exit %d: %s", i+1, code, stderr)
		}
	}
	want := "# TODO\n\n- [X] Build crane\n- [x] Build crane model\n- [ ] Paint wall\n"
	if got := readFile(t, filepath.Join(dir, "TODO.md")); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRemove_All(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()
	withTodo(batchTodo)(t, dir)

	if code, stderr := run(t, "rm", "-a", "-y", "crane"); code != ExitOK {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	if got, want := readFile(t, filepath.Join(dir, "TODO.md")), "# TODO\n\n- [ ] Paint wall\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRemove_AllAsksOnce(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()
	withTodo(batchTodo)(t, dir)

	// Without a terminal the prompt is answered "no": nothing is removed.
	if code, _ := run(t, "rm", "--all", "crane"); code != ExitGeneral {
		t.Errorf("exit %d, want %d", code, ExitGeneral)
	}
	if got := readFile(t, filepath.Join(dir, "TODO.md")); got != batchTodo {
		t.Errorf("file changed: %q", got)
	}
}

func TestBatch_Errors(t *testing.T) {
	tests := []struct {
		args       []string
		wantStderr string
	}{
		{[]string{"done", "crane"}, "or use --all to act on all of them"},
		{[]string{"done", "--all", "nothing"}, "TODO item not found: 'nothing'"},
		{[]string{"rm", "--all", "-p", "nope", "crane"}, "Project 'nope' not found"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			dir, cleanup := setupTestRepo(t)
			defer cleanup()
			withTodo(batchTodo)(t, dir)

			code, stderr := run(t, tt.args...)
			if code != ExitUsage || !strings.Contains(stderr, tt.wantStderr) {
				t.Errorf("exit %d, stderr %q; want %d and %q", code, stderr, ExitUsage, tt.wantStderr)
			}
			if got := readFile(t, filepath.Join(dir, "TODO.md")); got != batchTodo {
				t.Errorf("file changed: %q", got)
			}
		})
	}
}

func TestConfig_ErrorFormat(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()
	withConfig("todofile: TODO.md\nprojets:\n  - name: default\n    title: TODO\n")(t, dir)

	code, stderr := run(t, "list")
	want := "Error: Invalid configuration in .mktodo.yml\n- Unknown key 'projets' (expected: todofile, projects)\n"
	if code != ExitConfig || !strings.HasPrefix(stderr, want) {
		t.Errorf("exit %d, stderr:\n%s\nwant it to start with:\n%s", code, stderr, want)
	}
}
