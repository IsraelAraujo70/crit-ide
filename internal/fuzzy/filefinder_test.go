package fuzzy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileFinder_WalkAndFilter(t *testing.T) {
	// Create a temp directory structure.
	tmp := t.TempDir()

	dirs := []string{
		"src",
		"src/pkg",
		".git",
		"node_modules",
	}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(tmp, d), 0755)
	}

	files := []string{
		"main.go",
		"README.md",
		"src/app.go",
		"src/handler.go",
		"src/pkg/util.go",
		".git/config",
		"node_modules/pkg.js",
		".hidden_file",
	}
	for _, f := range files {
		os.WriteFile(filepath.Join(tmp, f), []byte("test"), 0644)
	}

	ff := NewFileFinder(tmp)

	// Should not include hidden files, .git, node_modules.
	count := ff.FileCount()
	if count != 5 {
		t.Errorf("expected 5 files, got %d", count)
	}

	// Filter with empty pattern.
	results := ff.Filter("")
	if len(results) == 0 {
		t.Fatal("empty pattern should return files")
	}

	// Filter with pattern.
	results = ff.Filter("app")
	found := false
	for _, r := range results {
		if r.RelPath == "src/app.go" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected src/app.go in results for pattern 'app'")
	}

	// Filter with no matches.
	results = ff.Filter("zzzzzzz")
	if len(results) != 0 {
		t.Errorf("expected 0 results for non-matching pattern, got %d", len(results))
	}
}

func TestFileFinder_Rebuild(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "a.go"), []byte(""), 0644)

	ff := NewFileFinder(tmp)
	if ff.FileCount() != 1 {
		t.Fatalf("expected 1 file, got %d", ff.FileCount())
	}

	// Add a file and rebuild.
	os.WriteFile(filepath.Join(tmp, "b.go"), []byte(""), 0644)
	ff.Rebuild()
	if ff.FileCount() != 2 {
		t.Errorf("expected 2 files after rebuild, got %d", ff.FileCount())
	}
}

func TestFileFinder_Gitignore(t *testing.T) {
	tmp := t.TempDir()

	os.WriteFile(filepath.Join(tmp, ".gitignore"), []byte("*.log\nbuild/\n"), 0644)
	os.MkdirAll(filepath.Join(tmp, "build"), 0755)
	os.WriteFile(filepath.Join(tmp, "build", "out.js"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmp, "app.log"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmp, "main.go"), []byte(""), 0644)

	ff := NewFileFinder(tmp)

	if ff.FileCount() != 1 {
		t.Errorf("expected 1 file (gitignore should filter build/ and *.log), got %d", ff.FileCount())
	}

	results := ff.Filter("")
	if len(results) != 1 || results[0].RelPath != "main.go" {
		t.Errorf("expected only main.go, got %v", results)
	}
}

func TestFileFinder_MaxResults(t *testing.T) {
	tmp := t.TempDir()

	// Create more files than maxResults.
	for i := 0; i < 30; i++ {
		name := filepath.Join(tmp, "file_"+string(rune('a'+i%26))+string(rune('0'+i/26))+".go")
		os.WriteFile(name, []byte(""), 0644)
	}

	ff := NewFileFinder(tmp)
	results := ff.Filter("file")
	if len(results) > maxResults {
		t.Errorf("expected at most %d results, got %d", maxResults, len(results))
	}
}

func TestMatchesGitignore(t *testing.T) {
	patterns := []string{"*.log", "build/", "tmp"}

	tests := []struct {
		path    string
		ignored bool
	}{
		{"app.log", true},
		{"src/debug.log", true},
		{"build/out.js", true},
		{"tmp", true},
		{"tmp/cache", true},
		{"main.go", false},
		{"src/app.go", false},
	}

	for _, tc := range tests {
		got := matchesGitignore(tc.path, patterns)
		if got != tc.ignored {
			t.Errorf("matchesGitignore(%q) = %v, want %v", tc.path, got, tc.ignored)
		}
	}
}
