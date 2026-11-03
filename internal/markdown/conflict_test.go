package markdown

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func writeString(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readString(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func lockPath(path string) string {
	return filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+".mktodo-lock")
}

func TestSave_RefusesExternalChange(t *testing.T) {
	doc := loadString(t, "# TODO\n\n- [ ] a\n", defaultConfig)
	writeString(t, doc.Path, "# TODO\n\n- [ ] a\n- [ ] added elsewhere\n")

	doc.Items[0].MarkDone()
	if err := SaveDocument(doc.Path, doc); !errors.Is(err, ErrModifiedExternally) {
		t.Fatalf("SaveDocument() error = %v, want ErrModifiedExternally", err)
	}
	if got := readString(t, doc.Path); !strings.Contains(got, "added elsewhere") {
		t.Errorf("the other change was overwritten: %q", got)
	}
}

func TestSave_ExternalDeleteIsAChange(t *testing.T) {
	doc := loadString(t, "# TODO\n\n- [ ] a\n", defaultConfig)
	if err := os.Remove(doc.Path); err != nil {
		t.Fatal(err)
	}
	if err := SaveDocument(doc.Path, doc); !errors.Is(err, ErrModifiedExternally) {
		t.Errorf("SaveDocument() error = %v, want ErrModifiedExternally", err)
	}
}

func TestSave_RepeatedSavesWork(t *testing.T) {
	doc := loadString(t, "# TODO\n\n- [ ] a\n", defaultConfig)
	for i := 0; i < 3; i++ {
		doc.Items[0].MarkDone()
		if err := SaveDocument(doc.Path, doc); err != nil {
			t.Fatalf("save %d: %v", i+1, err)
		}
	}
	if got := readString(t, doc.Path); got != "# TODO\n\n- [X] a\n" {
		t.Errorf("got %q", got)
	}
}

func TestOverwriteDocument_ReplacesExternalChange(t *testing.T) {
	doc := loadString(t, "# TODO\n\n- [ ] a\n", defaultConfig)
	writeString(t, doc.Path, "changed elsewhere\n")
	if err := OverwriteDocument(doc.Path, doc); err != nil {
		t.Fatal(err)
	}
	if got := readString(t, doc.Path); got != "# TODO\n\n- [ ] a\n" {
		t.Errorf("got %q", got)
	}
	// After overwriting, normal saves work again.
	if err := SaveDocument(doc.Path, doc); err != nil {
		t.Errorf("save after overwrite: %v", err)
	}
}

func TestSave_NewDocumentRefusesIfFileAppeared(t *testing.T) {
	path := filepath.Join(t.TempDir(), "TODO.md")
	doc := NewDocument(path, loadString(t, "", defaultConfig).Projects)
	if _, err := doc.AddItem(doc.Projects["default"], "x"); err != nil {
		t.Fatal(err)
	}
	writeString(t, path, "created elsewhere\n")
	if err := SaveDocument(path, doc); !errors.Is(err, ErrModifiedExternally) {
		t.Errorf("SaveDocument() error = %v, want ErrModifiedExternally", err)
	}
}

func TestSave_LockHeldByAnotherProcess(t *testing.T) {
	defer func(d time.Duration) { lockTimeout = d }(lockTimeout)
	lockTimeout = 100 * time.Millisecond

	doc := loadString(t, "# TODO\n\n- [ ] a\n", defaultConfig)
	writeString(t, lockPath(doc.Path), "")

	doc.Items[0].MarkDone()
	if err := SaveDocument(doc.Path, doc); !errors.Is(err, ErrLocked) {
		t.Fatalf("SaveDocument() error = %v, want ErrLocked", err)
	}
	if got := readString(t, doc.Path); got != "# TODO\n\n- [ ] a\n" {
		t.Errorf("file changed while locked: %q", got)
	}
}

func TestSave_StaleLockIsRemoved(t *testing.T) {
	doc := loadString(t, "# TODO\n\n- [ ] a\n", defaultConfig)
	lock := lockPath(doc.Path)
	writeString(t, lock, "")
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(lock, old, old); err != nil {
		t.Fatal(err)
	}

	if err := SaveDocument(doc.Path, doc); err != nil {
		t.Fatalf("SaveDocument() with a stale lock: %v", err)
	}
	if _, err := os.Stat(lock); !os.IsNotExist(err) {
		t.Error("lock file left behind")
	}
}

func TestSave_ConcurrentWritersLoseNothingSilently(t *testing.T) {
	doc := loadString(t, "# TODO\n\n- [ ] a\n", defaultConfig)
	path := doc.Path

	const writers = 8
	var wg sync.WaitGroup
	results := make(chan error, writers)
	for i := 0; i < writers; i++ {
		d, err := LoadDocument(path, defaultConfig)
		if err != nil {
			t.Fatal(err)
		}
		wg.Add(1)
		go func(d *Document, n int) {
			defer wg.Done()
			if _, err := d.AddItem(d.Projects["default"], "from writer "+string(rune('A'+n))); err != nil {
				results <- err
				return
			}
			results <- SaveDocument(path, d)
		}(d, i)
	}
	wg.Wait()
	close(results)

	saved := 0
	for err := range results {
		switch {
		case err == nil:
			saved++
		case !errors.Is(err, ErrModifiedExternally):
			t.Errorf("unexpected error: %v", err)
		}
	}
	// All writers loaded the same version, so exactly one may save; the
	// others must be told, not silently overwritten.
	if saved != 1 {
		t.Errorf("%d writers saved, want exactly 1", saved)
	}
	if n := strings.Count(readString(t, path), "from writer"); n != 1 {
		t.Errorf("file has %d added items, want 1:\n%s", n, readString(t, path))
	}
}
