package markdown

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jvzantvoort/mktodo/internal/config"
)

func TestLoadDocument(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		TodoFile: "README.md",
		Projects: []config.ProjectConfig{
			{Name: "default", Title: "TODO"},
		},
	}

	// Load simple test file
	doc, err := LoadDocument("../../testdata/markdown/simple.md", cfg)
	if err != nil {
		t.Fatalf("LoadDocument() error = %v", err)
	}

	if doc == nil {
		t.Fatal("LoadDocument() returned nil")
	}

	if len(doc.Sections) == 0 {
		t.Error("no sections parsed")
	}

	if len(doc.Items) == 0 {
		t.Error("no TODO items found")
	}
}

func TestDocument_GetOpenItems(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "default", Title: "TODO"},
		},
	}

	doc, err := LoadDocument("../../testdata/markdown/simple.md", cfg)
	if err != nil {
		t.Fatalf("LoadDocument() error = %v", err)
	}

	openItems := doc.GetOpenItems()
	if len(openItems) == 0 {
		t.Error("expected open items")
	}

	// All open items should not be done
	for _, item := range openItems {
		if item.Done {
			t.Error("GetOpenItems() returned done item")
		}
	}
}

func TestDocument_GetDoneItems(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "default", Title: "TODO"},
		},
	}

	doc, err := LoadDocument("../../testdata/markdown/simple.md", cfg)
	if err != nil {
		t.Fatalf("LoadDocument() error = %v", err)
	}

	doneItems := doc.GetDoneItems()
	
	// All done items should be done
	for _, item := range doneItems {
		if !item.Done {
			t.Error("GetDoneItems() returned non-done item")
		}
	}
}

func TestDocument_GetFIXMEItems(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "default", Title: "TODO"},
		},
	}

	doc, err := LoadDocument("../../testdata/markdown/simple.md", cfg)
	if err != nil {
		t.Fatalf("LoadDocument() error = %v", err)
	}

	fixmeItems := doc.GetFIXMEItems()
	if len(fixmeItems) == 0 {
		t.Error("expected FIXME items")
	}

	// All should be FIXME
	for _, item := range fixmeItems {
		if !item.IsFIXME {
			t.Error("GetFIXMEItems() returned non-FIXME item")
		}
	}
}

func TestDocument_FindItem(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "default", Title: "TODO"},
		},
	}

	doc, err := LoadDocument("../../testdata/markdown/simple.md", cfg)
	if err != nil {
		t.Fatalf("LoadDocument() error = %v", err)
	}

	tests := []struct {
		query string
		want  int
	}{
		{"task", 2},    // Should match "Open task 1" and "Done task 1"
		{"FIXME", 1},   // Should match FIXME item
		{"broken", 1},  // Should match "broken feature"
		{"xyz", 0},     // Should match nothing
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			items := doc.FindItem(tt.query)
			if len(items) != tt.want {
				t.Errorf("FindItem(%q) found %d items, want %d", tt.query, len(items), tt.want)
			}
		})
	}
}

func TestDocument_String(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "default", Title: "TODO"},
		},
	}

	doc, err := LoadDocument("../../testdata/markdown/simple.md", cfg)
	if err != nil {
		t.Fatalf("LoadDocument() error = %v", err)
	}

	// Get original content
	original, err := os.ReadFile("../../testdata/markdown/simple.md")
	if err != nil {
		t.Fatalf("reading original file: %v", err)
	}

	// Compare
	docStr := doc.String()
	if docStr != string(original) {
		t.Errorf("Document.String() doesn't match original:\ngot:\n%s\nwant:\n%s", docStr, string(original))
	}
}

func TestDocument_UpdateItem(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "default", Title: "TODO"},
		},
	}

	doc, err := LoadDocument("../../testdata/markdown/simple.md", cfg)
	if err != nil {
		t.Fatalf("LoadDocument() error = %v", err)
	}

	// Find an open item
	openItems := doc.GetOpenItems()
	if len(openItems) == 0 {
		t.Fatal("no open items to test")
	}

	item := openItems[0]
	originalDone := item.Done

	// Toggle it
	item.MarkDone()

	// Update document
	if err := doc.UpdateItem(item); err != nil {
		t.Fatalf("UpdateItem() error = %v", err)
	}

	// Verify the line was updated
	found := false
	for _, section := range doc.Sections {
		for _, line := range section.Lines {
			if line == item.String() {
				found = true
				break
			}
		}
	}

	if !found {
		t.Error("item line not updated in document")
	}

	// Toggle back
	item.Done = originalDone
}

func TestDocument_RemoveItem(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "default", Title: "TODO"},
		},
	}

	doc, err := LoadDocument("../../testdata/markdown/simple.md", cfg)
	if err != nil {
		t.Fatalf("LoadDocument() error = %v", err)
	}

	originalCount := len(doc.Items)
	if originalCount == 0 {
		t.Fatal("no items to remove")
	}

	// Remove first item
	item := doc.Items[0]
	if err := doc.RemoveItem(item); err != nil {
		t.Fatalf("RemoveItem() error = %v", err)
	}

	if len(doc.Items) != originalCount-1 {
		t.Errorf("item count = %d, want %d", len(doc.Items), originalCount-1)
	}
}

func TestSaveDocument(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "default", Title: "TODO"},
		},
	}

	doc, err := LoadDocument("../../testdata/markdown/simple.md", cfg)
	if err != nil {
		t.Fatalf("LoadDocument() error = %v", err)
	}

	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")

	// Save document
	if err := SaveDocument(tmpFile, doc); err != nil {
		t.Fatalf("SaveDocument() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("file was not created")
	}

	// Load it back
	doc2, err := LoadDocument(tmpFile, cfg)
	if err != nil {
		t.Fatalf("loading saved document: %v", err)
	}

	// Compare
	if len(doc2.Items) != len(doc.Items) {
		t.Errorf("saved document has %d items, original had %d", len(doc2.Items), len(doc.Items))
	}
}
