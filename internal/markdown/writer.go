package markdown

import (
	"fmt"
	"os"
	"path/filepath"
)

// Writer provides atomic writing of markdown documents
type Writer struct {
	path string
}

// NewWriter creates a new Writer for the given file path
func NewWriter(path string) *Writer {
	return &Writer{path: path}
}

// Write atomically writes the document to disk
// It writes to a temporary file first, then atomically renames it
func (w *Writer) Write(doc *Document) error {
	// Get absolute path
	absPath, err := filepath.Abs(w.path)
	if err != nil {
		return fmt.Errorf("getting absolute path: %w", err)
	}

	// Create temp file in same directory (for atomic rename)
	dir := filepath.Dir(absPath)
	tmpFile, err := os.CreateTemp(dir, ".mktodo-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup on error
	defer func() {
		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	// Write content
	content := doc.String()
	if _, err := tmpFile.WriteString(content); err != nil {
		return fmt.Errorf("writing content: %w", err)
	}

	// Ensure content is flushed
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("syncing file: %w", err)
	}

	// Close temp file
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}
	tmpFile = nil // Prevent defer from closing again

	// Get original file permissions (if exists)
	var perm os.FileMode = 0644
	if info, err := os.Stat(absPath); err == nil {
		perm = info.Mode()
	}

	// Set permissions on temp file
	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("setting permissions: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, absPath); err != nil {
		return fmt.Errorf("renaming file: %w", err)
	}

	return nil
}

// WriteString atomically writes a string to the file
func (w *Writer) WriteString(content string) error {
	// Get absolute path
	absPath, err := filepath.Abs(w.path)
	if err != nil {
		return fmt.Errorf("getting absolute path: %w", err)
	}

	// Create temp file in same directory
	dir := filepath.Dir(absPath)
	tmpFile, err := os.CreateTemp(dir, ".mktodo-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup on error
	defer func() {
		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	// Write content
	if _, err := tmpFile.WriteString(content); err != nil {
		return fmt.Errorf("writing content: %w", err)
	}

	// Sync and close
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("syncing file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}
	tmpFile = nil

	// Get original permissions
	var perm os.FileMode = 0644
	if info, err := os.Stat(absPath); err == nil {
		perm = info.Mode()
	}

	// Set permissions
	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("setting permissions: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, absPath); err != nil {
		return fmt.Errorf("renaming file: %w", err)
	}

	return nil
}

// SaveDocument is a convenience function to save a document atomically
func SaveDocument(path string, doc *Document) error {
	writer := NewWriter(path)
	return writer.Write(doc)
}
