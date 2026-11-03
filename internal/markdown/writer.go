package markdown

import (
	"errors"
	"fmt"
	"io/fs"
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
func (w *Writer) Write(doc *Document) error {
	return w.WriteString(doc.String())
}

// WriteString atomically writes a string to the file. It writes a temporary
// file in the same directory and renames it over the target, so the target
// is either fully replaced or left unchanged. The target's permissions are
// kept, missing parent directories are created, and if the target is a
// symlink the file it points to is replaced rather than the link.
func (w *Writer) WriteString(content string) (err error) {
	target, err := resolveTarget(w.path)
	if err != nil {
		return err
	}

	dir := filepath.Dir(target)
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, ".mktodo-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		if err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err = tmpFile.WriteString(content); err != nil {
		return fmt.Errorf("writing content: %w", err)
	}
	if err = tmpFile.Sync(); err != nil {
		return fmt.Errorf("syncing file: %w", err)
	}
	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	perm := os.FileMode(0o644)
	if info, statErr := os.Stat(target); statErr == nil {
		perm = info.Mode().Perm()
	}
	if err = os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("setting permissions: %w", err)
	}

	if err = os.Rename(tmpPath, target); err != nil {
		return fmt.Errorf("renaming file: %w", err)
	}
	return nil
}

// resolveTarget returns the absolute path to write, following a symlink at
// path if there is one.
func resolveTarget(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("getting absolute path: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(absPath)
	if errors.Is(err, fs.ErrNotExist) {
		return absPath, nil
	}
	if err != nil {
		return "", fmt.Errorf("resolving path: %w", err)
	}
	return resolved, nil
}

// SaveDocument is a convenience function to save a document atomically
func SaveDocument(path string, doc *Document) error {
	writer := NewWriter(path)
	return writer.Write(doc)
}
