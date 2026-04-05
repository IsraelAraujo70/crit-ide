package filetree

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestDir creates a temporary directory with test files and folders.
func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create directories.
	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.MkdirAll(filepath.Join(dir, "docs"), 0755)

	// Create files.
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
	os.WriteFile(filepath.Join(dir, "src", "app.go"), []byte("package src"), 0644)
	os.WriteFile(filepath.Join(dir, "src", "utils.go"), []byte("package src"), 0644)
	os.WriteFile(filepath.Join(dir, "docs", "readme.md"), []byte("# docs"), 0644)

	// Create hidden file (should be filtered out).
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.o"), 0644)

	return dir
}

func TestNew(t *testing.T) {
	dir := setupTestDir(t)
	ft, err := New(dir)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if ft.Root == nil {
		t.Fatal("expected root node")
	}
	if !ft.Root.IsDir {
		t.Fatal("expected root to be a directory")
	}
	if !ft.Root.Expanded {
		t.Fatal("expected root to be expanded")
	}
	if ft.VisibleCount() < 3 {
		t.Fatalf("expected at least 3 visible nodes (root + dirs + files), got %d", ft.VisibleCount())
	}
}

func TestHiddenFilesFiltered(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	for _, node := range ft.Visible {
		if node.Name == ".gitignore" {
			t.Fatal("hidden files should be filtered out")
		}
	}
}

func TestDirectoriesFirst(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	// Root's children should have dirs first.
	children := ft.Root.Children
	foundFile := false
	for _, child := range children {
		if !child.IsDir && !foundFile {
			foundFile = true
		}
		if child.IsDir && foundFile {
			t.Fatal("directories should come before files")
		}
	}
}

func TestMoveUpDown(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	if ft.Cursor != 0 {
		t.Fatalf("expected initial cursor 0, got %d", ft.Cursor)
	}

	ft.MoveDown()
	if ft.Cursor != 1 {
		t.Fatalf("expected cursor 1 after MoveDown, got %d", ft.Cursor)
	}

	ft.MoveUp()
	if ft.Cursor != 0 {
		t.Fatalf("expected cursor 0 after MoveUp, got %d", ft.Cursor)
	}

	// MoveUp at top should stay at 0.
	ft.MoveUp()
	if ft.Cursor != 0 {
		t.Fatalf("expected cursor to stay at 0, got %d", ft.Cursor)
	}
}

func TestMoveDownClamp(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	total := ft.VisibleCount()
	// Move to the last node.
	for i := 0; i < total+5; i++ {
		ft.MoveDown()
	}
	if ft.Cursor != total-1 {
		t.Fatalf("expected cursor clamped to %d, got %d", total-1, ft.Cursor)
	}
}

func TestToggleDirectory(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	// Move cursor to a directory child (first child of root).
	ft.MoveDown() // Should be a directory (docs or src, sorted alphabetically).
	node := ft.CursorNode()
	if node == nil || !node.IsDir {
		t.Fatal("expected cursor on a directory")
	}

	initialExpanded := node.Expanded
	result := ft.Toggle()

	if result != "" {
		t.Fatal("toggling a directory should return empty string")
	}
	if node.Expanded == initialExpanded {
		t.Fatal("toggle should change expanded state")
	}
}

func TestToggleFile(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	// Find a file node.
	for i, node := range ft.Visible {
		if !node.IsDir {
			ft.Cursor = i
			path := ft.Toggle()
			if path == "" {
				t.Fatal("toggling a file should return its path")
			}
			if path != node.Path {
				t.Fatalf("expected path %q, got %q", node.Path, path)
			}
			return
		}
	}
	t.Fatal("no file found in test tree")
}

func TestCollapseExpandDirectory(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	// Find an expanded directory.
	for i, node := range ft.Visible {
		if node.IsDir && node.Expanded {
			ft.Cursor = i
			countBefore := ft.VisibleCount()

			ft.Collapse()
			if node.Expanded {
				t.Fatal("expected directory to be collapsed")
			}

			countAfter := ft.VisibleCount()
			if countAfter >= countBefore {
				t.Fatalf("collapsing should reduce visible count: before=%d, after=%d", countBefore, countAfter)
			}

			ft.Expand()
			if !node.Expanded {
				t.Fatal("expected directory to be expanded")
			}
			return
		}
	}
}

func TestSetCursorToScreenRow(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	ft.ScrollY = 0
	ft.SetCursorToScreenRow(2)
	if ft.Cursor != 2 {
		t.Fatalf("expected cursor 2, got %d", ft.Cursor)
	}

	ft.ScrollY = 1
	ft.SetCursorToScreenRow(2)
	if ft.Cursor != 3 {
		t.Fatalf("expected cursor 3 (scrollY=1 + row=2), got %d", ft.Cursor)
	}
}

func TestEnsureCursorVisible(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	// Set cursor beyond viewport.
	ft.Cursor = 5
	ft.ScrollY = 0
	ft.EnsureCursorVisible(3) // viewport of 3 rows

	if ft.ScrollY < 3 {
		t.Fatalf("expected scrollY >= 3 to make cursor 5 visible, got %d", ft.ScrollY)
	}
}

func TestRefresh(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	countBefore := ft.VisibleCount()

	// Add a new file.
	os.WriteFile(filepath.Join(dir, "new_file.txt"), []byte("new"), 0644)

	ft.Refresh()

	countAfter := ft.VisibleCount()
	if countAfter <= countBefore {
		t.Fatalf("expected more visible nodes after refresh: before=%d, after=%d", countBefore, countAfter)
	}
}

func TestNodeInterface(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	node := ft.CursorNode()
	if node == nil {
		t.Fatal("expected a cursor node")
	}

	// Test the NodePath/NodeIsDir interface methods.
	if node.NodePath() != node.Path {
		t.Fatalf("NodePath() should return Path, got %q vs %q", node.NodePath(), node.Path)
	}
	if node.NodeIsDir() != node.IsDir {
		t.Fatalf("NodeIsDir() should return IsDir, got %v vs %v", node.NodeIsDir(), node.IsDir)
	}
}
