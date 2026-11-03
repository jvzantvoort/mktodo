package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// resetFlags restores every flag to its default, since the command tree is
// shared between test runs.
func resetFlags(t *testing.T) {
	t.Helper()
	var reset func(c *cobra.Command)
	reset = func(c *cobra.Command) {
		for _, fs := range []*pflag.FlagSet{c.Flags(), c.PersistentFlags()} {
			fs.VisitAll(func(f *pflag.Flag) {
				_ = f.Value.Set(f.DefValue)
				f.Changed = false
			})
		}
		for _, child := range c.Commands() {
			reset(child)
		}
	}
	reset(rootCmd)
	t.Cleanup(func() { reset(rootCmd) })
}

// run executes mktodo with args and returns the exit code and stderr.
func run(t *testing.T, args ...string) (int, string) {
	t.Helper()
	resetFlags(t)
	var stderr bytes.Buffer
	code := execute(&stderr, args)
	return code, stderr.String()
}

func TestExecute_ExitCodes(t *testing.T) {
	const todo = "# TODO\n\n- [ ] Build crane\n- [ ] Build crane model\n"

	tests := []struct {
		name       string
		setup      func(t *testing.T, dir string)
		args       []string
		wantCode   int
		wantStderr string // substring of stderr
		wantHint   bool   // whether the --help hint is printed
	}{
		{"success", withTodo(todo), []string{"list"}, ExitOK, "", false},
		{"unknown subcommand", nil, []string{"frobnicate"}, ExitUsage, `unknown command "frobnicate"`, true},
		{"unknown flag", nil, []string{"list", "--bogus"}, ExitUsage, "unknown flag: --bogus", true},
		{"missing argument", nil, []string{"done"}, ExitUsage, "requires at least 1 arg", true},
		{"unexpected argument", nil, []string{"list", "extra"}, ExitUsage, `unknown command "extra"`, true},
		{"conflicting flags", nil, []string{"list", "-o", "-d"}, ExitUsage, "--open and --done cannot be used together", true},
		{"invalid list format", nil, []string{"list", "--format", "xml"}, ExitUsage, "Invalid format 'xml'", true},
		{"invalid report format", nil, []string{"report", "--format", "html"}, ExitUsage, "Valid formats: text, markdown, json", true},
		{"unknown project", withTodo(todo), []string{"list", "-p", "nope"}, ExitUsage, "Available projects: default", true},
		{"item not found", withTodo(todo), []string{"done", "paint"}, ExitUsage, "TODO item not found: 'paint'", true},
		{"ambiguous item", withTodo(todo), []string{"done", "crane"}, ExitUsage, "Multiple TODO items match 'crane'", true},
		{"missing config", removeConfig, []string{"list"}, ExitConfig, "Configuration file not found", false},
		{"invalid config", withConfig("todofile: TODO.md\nprojects:\n  - name: a\n    title: A\n    parent: missing\n"), []string{"list"}, ExitConfig, "Invalid configuration", false},
		{"missing todo file", nil, []string{"list"}, ExitFile, "File 'TODO.md' not found", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, cleanup := setupTestRepo(t)
			defer cleanup()
			if tt.setup != nil {
				tt.setup(t, dir)
			}

			code, stderr := run(t, tt.args...)
			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d (stderr: %q)", code, tt.wantCode, stderr)
			}
			if !strings.Contains(stderr, tt.wantStderr) {
				t.Errorf("stderr = %q, want it to contain %q", stderr, tt.wantStderr)
			}
			if got := strings.Contains(stderr, "--help' for usage"); got != tt.wantHint {
				t.Errorf("help hint shown = %v, want %v (stderr: %q)", got, tt.wantHint, stderr)
			}
			if strings.Contains(stderr, "Usage:") {
				t.Errorf("full usage text printed on error: %q", stderr)
			}
			if n := strings.Count(stderr, "Error:"); code != ExitOK && n != 1 {
				t.Errorf("error printed %d times, want once: %q", n, stderr)
			}
		})
	}
}

func TestExecute_NotInGitRepo(t *testing.T) {
	t.Chdir(t.TempDir())
	code, stderr := run(t, "list")
	if code != ExitGit {
		t.Errorf("exit code = %d, want %d", code, ExitGit)
	}
	if !strings.Contains(stderr, "Run 'git init' to initialize a repository") {
		t.Errorf("stderr = %q, want the FR-1 message", stderr)
	}
}

func TestExecute_CancelledPrompt(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()
	withTodo("# Notes\n")(t, dir)

	// With no terminal input the prompt is answered "no".
	code, stderr := run(t, "add", "task")
	if code != ExitGeneral {
		t.Errorf("exit code = %d, want %d", code, ExitGeneral)
	}
	if strings.TrimSpace(stderr) != "Operation cancelled." {
		t.Errorf("stderr = %q, want only the cancellation message", stderr)
	}
}

func withTodo(content string) func(t *testing.T, dir string) {
	return func(t *testing.T, dir string) {
		writeFile(t, filepath.Join(dir, "TODO.md"), content)
	}
}

func withConfig(content string) func(t *testing.T, dir string) {
	return func(t *testing.T, dir string) {
		writeFile(t, filepath.Join(dir, ".mktodo.yml"), content)
	}
}

func removeConfig(t *testing.T, dir string) {
	t.Helper()
	if err := os.Remove(filepath.Join(dir, ".mktodo.yml")); err != nil {
		t.Fatal(err)
	}
}
