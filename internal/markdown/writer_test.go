package markdown

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriter_Write(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	// Create a simple document
	doc := &Document{
		Path: testFile,
		Sections: []*Section{
			{
				Type:  SectionHeader,
				Level: 1,
				Title: "Test",
				Lines: []string{"# Test"},
			},
			{
				Type:  SectionOther,
				Lines: []string{"Some content"},
			},
		},
	}

	// Write document
	writer := NewWriter(testFile)
	if err := writer.Write(doc); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("file was not created")
	}

	// Read back and verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	expected := "# Test\nSome content"
	if string(content) != expected {
		t.Errorf("content = %q, want %q", string(content), expected)
	}
}

func TestWriter_WriteString(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	content := "# Header\n\nSome content\n"

	writer := NewWriter(testFile)
	if err := writer.WriteString(content); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}

	// Read back
	readContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	if string(readContent) != content {
		t.Errorf("content not preserved")
	}
}

func TestWriter_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	// Write initial content
	initialContent := "Initial content"
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("writing initial file: %v", err)
	}

	// Create document
	doc := &Document{
		Sections: []*Section{
			{Lines: []string{"New content"}},
		},
	}

	// Write atomically
	writer := NewWriter(testFile)
	if err := writer.Write(doc); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Verify new content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	if string(content) != "New content\n" {
		t.Errorf("content = %q, want %q", string(content), "New content\n")
	}
}

func TestWriter_PreservesPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	// Create file with specific permissions
	initialPerm := os.FileMode(0600)
	if err := os.WriteFile(testFile, []byte("test"), initialPerm); err != nil {
		t.Fatalf("creating initial file: %v", err)
	}

	// Write new content
	doc := &Document{
		Sections: []*Section{
			{Lines: []string{"New content"}},
		},
	}

	writer := NewWriter(testFile)
	if err := writer.Write(doc); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Check permissions
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("stat file: %v", err)
	}

	if info.Mode() != initialPerm {
		t.Errorf("permissions = %v, want %v", info.Mode(), initialPerm)
	}
}

func TestWriter_NoTempFilesLeft(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	doc := &Document{
		Sections: []*Section{
			{Lines: []string{"Content"}},
		},
	}

	writer := NewWriter(testFile)
	if err := writer.Write(doc); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Check for temp files
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("reading dir: %v", err)
	}

	for _, entry := range entries {
		if entry.Name() != "test.md" {
			t.Errorf("unexpected file left: %s", entry.Name())
		}
	}
}

func TestSaveDocument_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	doc := &Document{
		Path: testFile,
		Sections: []*Section{
			{
				Type:  SectionHeader,
				Level: 1,
				Title: "TODO",
				Lines: []string{"# TODO"},
			},
			{
				Type: SectionTODO,
				Lines: []string{
					"- [ ] Task 1",
					"- [X] Task 2",
				},
			},
		},
	}

	// Save
	if err := SaveDocument(testFile, doc); err != nil {
		t.Fatalf("SaveDocument() error = %v", err)
	}

	// Read back
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	expected := "# TODO\n- [ ] Task 1\n- [X] Task 2"
	if string(content) != expected {
		t.Errorf("content = %q, want %q", string(content), expected)
	}
}
