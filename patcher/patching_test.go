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

func TestPatcher_ApplyAll_skipsIgnoredFiles(t *testing.T) {
	t.Parallel()

	// Given
	repo := t.TempDir()
	vanilla := t.TempDir()
	writeTestFile(t, filepath.Join(repo, ".spignore"), "common/ignored.txt\n")
	if err := os.MkdirAll(filepath.Join(repo, "extras"), 0o755); err != nil {
		t.Fatalf("mkdir extras: %v", err)
	}
	writeTestFile(t, filepath.Join(vanilla, "common", "kept.txt"), "vanilla\n")
	writeTestFile(t, filepath.Join(vanilla, "common", "ignored.txt"), "vanilla\n")
	writeTestFile(t, filepath.Join(repo, "patches", "common", "kept.txt.patch"), strings.Join([]string{
		"diff --git a/original b/modified",
		"index 3af67b2..e45c9c2 100644",
		"--- a/original",
		"+++ b/modified",
		"@@ -1 +1 @@",
		"-vanilla",
		"+patched",
		"",
	}, "\n"))
	writeTestFile(t, filepath.Join(repo, "patches", "common", "ignored.txt.patch"), strings.Join([]string{
		"diff --git a/original b/modified",
		"index 3af67b2..e45c9c2 100644",
		"--- a/original",
		"+++ b/modified",
		"@@ -1 +1 @@",
		"-vanilla",
		"+ignored",
		"",
	}, "\n"))
	patcher, err := NewPatcher(repo, vanilla)
	if err != nil {
		t.Fatalf("new patcher: %v", err)
	}

	// When
	err = patcher.ApplyAll()

	// Then
	if err != nil {
		t.Fatalf("apply all: %v", err)
	}
	if got := readTestFile(t, filepath.Join(repo, "src", "common", "kept.txt")); got != "patched\n" {
		t.Fatalf("kept source = %q, want %q", got, "patched\n")
	}
	if _, err := os.Stat(filepath.Join(repo, "src", "common", "ignored.txt")); !os.IsNotExist(err) {
		t.Fatalf("ignored file was applied, stat err = %v", err)
	}
}

func TestPatcher_ignored_matchesGitignore(t *testing.T) {
	t.Parallel()

	// Given
	ign := LoadIgnoreRules(strings.NewReader("common/traits/\n*.log\n"))
	patcher, err := NewPatcher(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatalf("new patcher: %v", err)
	}
	patcher.ign = ign

	// When / Then
	cases := []struct {
		rel  string
		dir  bool
		want bool
	}{
		{"common/traits/00_traits.txt", false, true},
		{"common/traits", true, true},
		{"common/name.txt", false, false},
		{"logs/error.log", false, true},
		{"logs/error.log", true, true},
	}
	for _, c := range cases {
		if got := patcher.ignored(c.rel, c.dir); got != c.want {
			t.Errorf("ignored(%q, dir=%v) = %v, want %v", c.rel, c.dir, got, c.want)
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
