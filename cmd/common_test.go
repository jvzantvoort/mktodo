package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestLoadConfig_InvalidConfigRejected(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	writeFile(t, filepath.Join(tmpDir, ".mktodo.yml"), `todofile: TODO.md
projects:
  - name: technic
    title: Technic
    parent: lego
`)

	err := runList(listCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "Project 'technic' references non-existent parent 'lego'") {
		t.Fatalf("runList() error = %v, want a validation error about the missing parent", err)
	}
}

func TestLoadConfig_ConfigFlag(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	altConfig := filepath.Join(tmpDir, "alt.yml")
	writeFile(t, altConfig, `todofile: ALT.md
projects:
  - name: default
    title: TODO
    parent: nil
`)
	if err := rootCmd.PersistentFlags().Set("config", altConfig); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("config", "") })

	setFlag(t, addCmd, "yes", "true")
	setFlag(t, addCmd, "project", "default")
	if err := runAdd(addCmd, []string{"via alt config"}); err != nil {
		t.Fatalf("runAdd() error = %v", err)
	}

	if got := readFile(t, filepath.Join(tmpDir, "ALT.md")); !strings.Contains(got, "- [ ] via alt config") {
		t.Errorf("ALT.md = %q, want the new item", got)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "TODO.md")); !os.IsNotExist(err) {
		t.Error("TODO.md from the default config was written")
	}
}

func TestAdd_CreatesMissingSectionInExistingFile(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	todoPath := filepath.Join(tmpDir, "TODO.md")
	writeFile(t, todoPath, "# My Project\n\nSome text.\n")

	setFlag(t, addCmd, "yes", "true")
	setFlag(t, addCmd, "project", "default")
	if err := runAdd(addCmd, []string{"first"}); err != nil {
		t.Fatalf("runAdd() error = %v", err)
	}

	want := "# My Project\n\nSome text.\n\n# TODO\n\n- [ ] first\n"
	if got := readFile(t, todoPath); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDone_ExactMatchWins(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	todoPath := filepath.Join(tmpDir, "TODO.md")
	writeFile(t, todoPath, "# TODO\n\n- [ ] Build crane\n- [ ] Build crane model\n")

	setFlag(t, doneCmd, "project", "")
	if err := runDone(doneCmd, []string{"Build", "crane"}); err != nil {
		t.Fatalf("runDone() error = %v", err)
	}

	want := "# TODO\n\n- [X] Build crane\n- [ ] Build crane model\n"
	if got := readFile(t, todoPath); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRemove_PreservesSurroundingContent(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	todoPath := filepath.Join(tmpDir, "TODO.md")
	writeFile(t, todoPath, "# TODO\r\n\r\n* [ ] keep\r\n* [ ] drop\r\n\r\n```\r\n- [ ] drop\r\n```\r\n")

	setFlag(t, removeCmd, "yes", "true")
	setFlag(t, removeCmd, "project", "")
	if err := runRemove(removeCmd, []string{"drop"}); err != nil {
		t.Fatalf("runRemove() error = %v", err)
	}

	want := "# TODO\r\n\r\n* [ ] keep\r\n\r\n```\r\n- [ ] drop\r\n```\r\n"
	if got := readFile(t, todoPath); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
