package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/markdown"
)

// Helper function to create a test repository
func setupTestRepo(t *testing.T) (string, func()) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Initialize as git repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("creating .git dir: %v", err)
	}

	// Create config
	cfgPath := filepath.Join(tmpDir, ".mktodo.yml")
	cfgContent := `---
todofile: TODO.md
projects:
  - name: default
    title: TODO
    parent: nil
`
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
		t.Fatalf("creating config: %v", err)
	}

	// Change to temp directory
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)

	cleanup := func() {
		os.Chdir(oldWd)
	}

	return tmpDir, cleanup
}

func TestIntegration_AddCommand(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Test adding an item
	addCmd.Flags().Set("yes", "true")
	addCmd.Flags().Set("project", "default")

	err := runAdd(addCmd, []string{"Test", "task"})
	if err != nil {
		t.Fatalf("runAdd() error = %v", err)
	}

	// Verify file was created and contains the item
	todoPath := filepath.Join(tmpDir, "TODO.md")
	content, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("reading TODO file: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "Test task") {
		t.Error("TODO file doesn't contain added item")
	}
	if !contains(contentStr, "- [ ]") {
		t.Error("TODO file doesn't contain checkbox")
	}
}

func TestIntegration_DoneCommand(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create TODO file with an item
	todoPath := filepath.Join(tmpDir, "TODO.md")
	content := `# TODO

- [ ] Complete documentation
- [ ] Write tests
`
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		t.Fatalf("creating TODO file: %v", err)
	}

	// Mark item as done
	err := runDone(doneCmd, []string{"documentation"})
	if err != nil {
		t.Fatalf("runDone() error = %v", err)
	}

	// Verify item was marked as done
	newContent, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("reading TODO file: %v", err)
	}

	if !contains(string(newContent), "[X]") && !contains(string(newContent), "[x]") {
		t.Error("item not marked as done")
	}
}

func TestIntegration_RemoveCommand(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create TODO file with items
	todoPath := filepath.Join(tmpDir, "TODO.md")
	content := `# TODO

- [ ] Task to remove
- [ ] Keep this task
`
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		t.Fatalf("creating TODO file: %v", err)
	}

	// Remove item with --yes flag
	removeCmd.Flags().Set("yes", "true")
	err := runRemove(removeCmd, []string{"remove"})
	if err != nil {
		t.Fatalf("runRemove() error = %v", err)
	}

	// Verify item was removed
	newContent, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("reading TODO file: %v", err)
	}

	if contains(string(newContent), "Task to remove") {
		t.Error("item not removed from file")
	}
	if !contains(string(newContent), "Keep this task") {
		t.Error("other item was incorrectly removed")
	}
}

func TestIntegration_ListCommand(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create TODO file
	todoPath := filepath.Join(tmpDir, "TODO.md")
	content := `# TODO

- [ ] Open task
- [X] Done task
- [ ] FIXME: broken
`
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		t.Fatalf("creating TODO file: %v", err)
	}

	// Test list command (basic execution test)
	err := runList(listCmd, []string{})
	if err != nil {
		t.Fatalf("runList() error = %v", err)
	}

	// Test open filter
	listCmd.Flags().Set("open", "true")
	err = runList(listCmd, []string{})
	if err != nil {
		t.Fatalf("runList() with --open error = %v", err)
	}
}

func TestIntegration_OpenCommand(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create TODO file
	todoPath := filepath.Join(tmpDir, "TODO.md")
	content := `# TODO

- [ ] Open task 1
- [X] Done task
- [ ] FIXME: broken feature
`
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		t.Fatalf("creating TODO file: %v", err)
	}

	// Test open command
	err := runOpen(openCmd, []string{})
	if err != nil {
		t.Fatalf("runOpen() error = %v", err)
	}
}

func TestIntegration_ReportCommand(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create TODO file
	todoPath := filepath.Join(tmpDir, "TODO.md")
	content := `# TODO

- [ ] Task 1
- [X] Task 2
- [ ] FIXME: Task 3
`
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		t.Fatalf("creating TODO file: %v", err)
	}

	// Test report command
	err := runReport(reportCmd, []string{})
	if err != nil {
		t.Fatalf("runReport() error = %v", err)
	}

	// Test with different formats
	reportCmd.Flags().Set("format", "json")
	err = runReport(reportCmd, []string{})
	if err != nil {
		t.Fatalf("runReport() with json format error = %v", err)
	}

	reportCmd.Flags().Set("format", "markdown")
	err = runReport(reportCmd, []string{})
	if err != nil {
		t.Fatalf("runReport() with markdown format error = %v", err)
	}
}

func TestIntegration_FullWorkflow(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	// 1. Add a task
	addCmd.Flags().Set("yes", "true")
	if err := runAdd(addCmd, []string{"Implement", "feature"}); err != nil {
		t.Fatalf("adding task: %v", err)
	}

	// 2. List tasks
	if err := runList(listCmd, []string{}); err != nil {
		t.Fatalf("listing tasks: %v", err)
	}

	// 3. Mark as done
	if err := runDone(doneCmd, []string{"feature"}); err != nil {
		t.Fatalf("marking done: %v", err)
	}

	// 4. Generate report
	if err := runReport(reportCmd, []string{}); err != nil {
		t.Fatalf("generating report: %v", err)
	}

	// 5. Verify the task was completed
	todoPath := filepath.Join(tmpDir, "TODO.md")
	cfg, _ := config.Load(filepath.Join(tmpDir, ".mktodo.yml"))
	doc, err := markdown.LoadDocument(todoPath, cfg)
	if err != nil {
		t.Fatalf("loading document: %v", err)
	}

	doneItems := doc.GetDoneItems()
	if len(doneItems) == 0 {
		t.Error("expected at least one done item")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)
}

func containsHelper(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
