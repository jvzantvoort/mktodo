package markdown

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// Errors returned when saving.
var (
	// ErrModifiedExternally means the file changed on disk after the
	// document was loaded, so saving would overwrite someone else's changes.
	ErrModifiedExternally = errors.New("file was changed by another program since it was loaded")

	// ErrLocked means another mktodo kept the file locked for too long.
	ErrLocked = errors.New("file is locked by another mktodo process")
)

// Locking settings; variables so tests can shorten them.
var (
	lockTimeout    = 3 * time.Second
	lockStaleAfter = 30 * time.Second
	lockRetry      = 50 * time.Millisecond
)

// sourceState records the file's content when it was loaded or last saved.
type sourceState struct {
	exists bool
	hash   [sha256.Size]byte
}

func newSourceState(content []byte) *sourceState {
	return &sourceState{exists: true, hash: sha256.Sum256(content)}
}

// matches reports whether the file at path still has the recorded content.
func (s *sourceState) matches(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return !s.exists, nil
	}
	if err != nil {
		return false, err
	}
	hash := sha256.Sum256(content)
	return s.exists && bytes.Equal(hash[:], s.hash[:]), nil
}

// SaveDocument saves the document atomically. When path is the file the
// document was loaded from and that file has changed since, it returns
// ErrModifiedExternally and writes nothing.
func SaveDocument(path string, doc *Document) error {
	return doc.save(path, true)
}

// OverwriteDocument saves the document even if the file changed since it
// was loaded, replacing those changes.
func OverwriteDocument(path string, doc *Document) error {
	return doc.save(path, false)
}

// save holds a lock on the file while it checks for outside changes and
// writes, so two saves can't interleave between the check and the write.
func (d *Document) save(path string, check bool) error {
	target, err := resolveTarget(path)
	if err != nil {
		return err
	}
	release, err := acquireLock(target)
	if err != nil {
		return err
	}
	defer release()

	// Documents built by hand (no source) and saves to a different path
	// have nothing to compare against.
	if check && d.source != nil && filepath.Clean(path) == filepath.Clean(d.Path) {
		same, err := d.source.matches(path)
		if err != nil {
			return fmt.Errorf("checking for changes: %w", err)
		}
		if !same {
			return ErrModifiedExternally
		}
	}

	content := d.String()
	if err := NewWriter(path).WriteString(content); err != nil {
		return err
	}
	if filepath.Clean(path) == filepath.Clean(d.Path) {
		d.source = newSourceState([]byte(content))
	}
	return nil
}

// acquireLock creates a lock file next to target and returns a function
// that removes it. A lock left behind by a crashed process is removed once
// it is older than lockStaleAfter.
func acquireLock(target string) (func(), error) {
	lockPath := filepath.Join(filepath.Dir(target), "."+filepath.Base(target)+".mktodo-lock")
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return nil, fmt.Errorf("creating directory: %w", err)
	}

	deadline := time.Now().Add(lockTimeout)
	for {
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			_ = f.Close()
			return func() { _ = os.Remove(lockPath) }, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return nil, fmt.Errorf("creating lock file: %w", err)
		}

		if info, statErr := os.Stat(lockPath); statErr == nil && time.Since(info.ModTime()) > lockStaleAfter {
			_ = os.Remove(lockPath)
			continue
		}
		if time.Now().After(deadline) {
			return nil, ErrLocked
		}
		time.Sleep(lockRetry)
	}
}
