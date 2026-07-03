package patcher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPatcher_Apply_copiesVanilla_whenPatchMissing(t *testing.T) {
	t.Parallel()

	// Given
	repo := t.TempDir()
	vanilla := t.TempDir()
	file := filepath.Join("common", "name.txt")
	writeTestFile(t, filepath.Join(vanilla, file), "vanilla\n")
	patcher, err := NewPatcher(repo, vanilla)
	if err != nil {
		t.Fatalf("new patcher: %v", err)
	}

	// When
	err = patcher.Apply(file)

	// Then
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	got := readTestFile(t, filepath.Join(repo, "src", file))
	if got != "vanilla\n" {
		t.Fatalf("source = %q, want %q", got, "vanilla\n")
	}
}

func TestPatcher_Apply_writesPatchedSource_whenPatchExists(t *testing.T) {
	t.Parallel()

	// Given
	repo := t.TempDir()
	vanilla := t.TempDir()
	file := filepath.Join("common", "name.txt")
	writeTestFile(t, filepath.Join(vanilla, file), "vanilla\n")
	writeTestFile(t, filepath.Join(repo, "patches", file+".patch"), strings.Join([]string{
		"diff --git a/original b/modified",
		"index 3af67b2..e45c9c2 100644",
		"--- a/original",
		"+++ b/modified",
		"@@ -1 +1 @@",
		"-vanilla",
		"+patched",
		"",
	}, "\n"))
	patcher, err := NewPatcher(repo, vanilla)
	if err != nil {
		t.Fatalf("new patcher: %v", err)
	}

	// When
	err = patcher.Apply(file)

	// Then
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	got := readTestFile(t, filepath.Join(repo, "src", file))
	if got != "patched\n" {
		t.Fatalf("source = %q, want %q", got, "patched\n")
	}
}

func TestPatcher_Generate_createsPatch_whenSourceDiffers(t *testing.T) {
	t.Parallel()

	// Given
	repo := t.TempDir()
	vanilla := t.TempDir()
	file := filepath.Join("common", "name.txt")
	writeTestFile(t, filepath.Join(vanilla, file), "vanilla\n")
	writeTestFile(t, filepath.Join(repo, "src", file), "patched\n")
	patcher, err := NewPatcher(repo, vanilla)
	if err != nil {
		t.Fatalf("new patcher: %v", err)
	}

	// When
	err = patcher.Generate(file)

	// Then
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	got := readTestFile(t, filepath.Join(repo, "patches", file+".patch"))
	for _, want := range []string{"--- a/original", "+++ b/modified", "-vanilla", "+patched"} {
		if !strings.Contains(got, want) {
			t.Fatalf("patch does not contain %q:\n%s", want, got)
		}
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
