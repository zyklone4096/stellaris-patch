package patcher

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFileChangeChecker_HasChanged(t *testing.T) {
	t.Parallel()

	t.Run("saved state unchanged", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		metadata := filepath.Join(root, "metadata")
		file := filepath.Join("subdir", "file.txt")
		full := filepath.Join(root, file)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir source: %v", err)
		}
		if err := os.WriteFile(full, []byte("hello"), 0o644); err != nil {
			t.Fatalf("write source: %v", err)
		}

		checker := NewChangeChecker(root, metadata)
		if err := checker.SaveCurrentState(file); err != nil {
			t.Fatalf("save current state: %v", err)
		}

		changed, err := checker.HasChanged(file)
		if err != nil {
			t.Fatalf("has changed: %v", err)
		}
		if changed {
			t.Fatalf("expected unchanged")
		}
	})

	t.Run("missing metadata changed nil error", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		metadata := filepath.Join(root, "metadata")
		file := "file.txt"
		full := filepath.Join(root, file)
		if err := os.WriteFile(full, []byte("hello"), 0o644); err != nil {
			t.Fatalf("write source: %v", err)
		}

		checker := NewChangeChecker(root, metadata)
		changed, err := checker.HasChanged(file)
		if err != nil {
			t.Fatalf("has changed: %v", err)
		}
		if !changed {
			t.Fatalf("expected changed")
		}
	})

	t.Run("corrupt metadata changed nil error", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		metadata := filepath.Join(root, "metadata")
		file := "file.txt"
		full := filepath.Join(root, file)
		if err := os.WriteFile(full, []byte("hello"), 0o644); err != nil {
			t.Fatalf("write source: %v", err)
		}
		if err := os.MkdirAll(metadata, 0o755); err != nil {
			t.Fatalf("mkdir metadata: %v", err)
		}
		if err := os.WriteFile(filepath.Join(metadata, file), []byte("{"), 0o644); err != nil {
			t.Fatalf("write metadata: %v", err)
		}

		checker := NewChangeChecker(root, metadata)
		changed, err := checker.HasChanged(file)
		if err != nil {
			t.Fatalf("has changed: %v", err)
		}
		if !changed {
			t.Fatalf("expected changed")
		}
	})

	t.Run("missing source changed with error", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		metadata := filepath.Join(root, "metadata")
		checker := NewChangeChecker(root, metadata)

		changed, err := checker.HasChanged("missing.txt")
		if err == nil {
			t.Fatalf("expected error")
		}
		if !changed {
			t.Fatalf("expected changed")
		}
	})
}

func TestFileChangeChecker_ListChanges(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	metadata := filepath.Join(root, "metadata")
	checker := NewChangeChecker(root, metadata)

	changedFile := filepath.Join(root, "subdir", "z.txt")
	corruptMetadataFile := filepath.Join(root, "subdir", "c.txt")
	unchangedFile := filepath.Join(root, "subdir", "a.txt")
	if err := os.MkdirAll(filepath.Dir(changedFile), 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.WriteFile(changedFile, []byte("changed"), 0o644); err != nil {
		t.Fatalf("write changed source: %v", err)
	}
	if err := os.WriteFile(corruptMetadataFile, []byte("corrupt metadata"), 0o644); err != nil {
		t.Fatalf("write corrupt metadata source: %v", err)
	}
	if err := os.WriteFile(unchangedFile, []byte("same"), 0o644); err != nil {
		t.Fatalf("write unchanged source: %v", err)
	}
	if err := checker.SaveCurrentState(filepath.Join("subdir", "a.txt")); err != nil {
		t.Fatalf("save unchanged state: %v", err)
	}
	if err := os.WriteFile(filepath.Join(metadata, "subdir", "c.txt"), []byte("{"), 0o644); err != nil {
		t.Fatalf("write corrupt metadata: %v", err)
	}

	got, err := checker.ListChanges("subdir")
	if err != nil {
		t.Fatalf("list changes: %v", err)
	}
	want := []string{filepath.Join("subdir", "c.txt"), filepath.Join("subdir", "z.txt")}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("list changes = %v, want %v", got, want)
	}

	got, err = checker.ListChanges("")
	if err != nil {
		t.Fatalf("list root changes: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("list root changes = %v, want %v", got, want)
	}
}
