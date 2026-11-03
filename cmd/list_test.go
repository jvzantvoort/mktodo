package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const listTodo = `# TODO

- [ ] first
- [x] FIXME: done bug
- [ ] second
- [ ] fixme: open bug
`

// runOut executes mktodo with args and returns the exit code, stdout and stderr.
func runOut(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	t.Cleanup(func() { rootCmd.SetOut(nil) })
	code, stderr := run(t, args...)
	return code, stdout.String(), stderr
}

func TestList_FIXMEFirstThenFileOrder(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()
	withTodo(listTodo)(t, dir)

	code, out, stderr := runOut(t, "list")
	if code != ExitOK {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	want := `[X] FIXME: done bug [FIXME] (default)
[ ] fixme: open bug [FIXME] (default)
[ ] first (default)
[ ] second (default)
`
	if out != want {
		t.Errorf("got:\n%s\nwant:\n%s", out, want)
	}
}

func TestList_Filters(t *testing.T) {
	tests := []struct {
		args []string
		want []string // descriptions, in order
	}{
		{[]string{"list", "--fixme"}, []string{"FIXME: done bug", "fixme: open bug"}},
		{[]string{"list", "-f", "-o"}, []string{"fixme: open bug"}},
		{[]string{"list", "-d"}, []string{"FIXME: done bug"}},
		{[]string{"open"}, []string{"fixme: open bug", "first", "second"}},
		{[]string{"open", "--fixme"}, []string{"fixme: open bug"}},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			dir, cleanup := setupTestRepo(t)
			defer cleanup()
			withTodo(listTodo)(t, dir)

			code, out, stderr := runOut(t, append(tt.args, "--format", "json")...)
			if code != ExitOK {
				t.Fatalf("exit %d: %s", code, stderr)
			}
			var got []string
			for _, line := range strings.Split(out, "\n") {
				if strings.Contains(line, `"description"`) {
					got = append(got, strings.Trim(strings.SplitN(line, ": ", 2)[1], `",`))
				}
			}
			if strings.Join(got, "|") != strings.Join(tt.want, "|") {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOpen_SameAsListOpen(t *testing.T) {
	for _, format := range []string{"text", "json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			dir, cleanup := setupTestRepo(t)
			defer cleanup()
			withTodo(listTodo)(t, dir)

			_, openOut, _ := runOut(t, "open", "--format", format)
			_, listOut, _ := runOut(t, "list", "--open", "--format", format)
			if openOut != listOut || openOut == "" {
				t.Errorf("open and list --open differ:\nopen:\n%s\nlist --open:\n%s", openOut, listOut)
			}
		})
	}
}

func TestList_EmptyMessages(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()
	withTodo("# TODO\n\n- [x] done\n")(t, dir)

	if _, out, _ := runOut(t, "open"); !strings.Contains(out, "No open TODO items") {
		t.Errorf("open with nothing open printed %q", out)
	}
}

func TestReport_HTMLToFile(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()
	withTodo(listTodo)(t, dir)
	reportNow = func() time.Time { return time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { reportNow = time.Now })

	outFile := filepath.Join(dir, "out", "report.html")
	code, out, stderr := runOut(t, "report", "--format", "html", "-o", outFile)
	if code != ExitOK {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	if !strings.Contains(out, "Report written to "+outFile) {
		t.Errorf("stdout = %q", out)
	}
	b, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "<title>mktodo Report - 2026-11-04</title>") {
		t.Errorf("report.html doesn't look like the report:\n%s", b)
	}
}
