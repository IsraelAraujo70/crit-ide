package filetree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateFile(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	// Move cursor to root (a directory).
	ft.Cursor = 0

	path, err := ft.CreateFile("newfile.txt")
	if err != nil {
		t.Fatalf("CreateFile error: %v", err)
	}
	if path != filepath.Join(dir, "newfile.txt") {
		t.Fatalf("expected path %q, got %q", filepath.Join(dir, "newfile.txt"), path)
	}

	// Verify file exists.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("file should exist on disk")
	}

	// Verify cursor moved to the new file.
	node := ft.CursorNode()
	if node == nil || node.Name != "newfile.txt" {
		t.Fatalf("cursor should be on newfile.txt, got %v", node)
	}
}

func TestCreateDirectory(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	ft.Cursor = 0

	path, err := ft.CreateFile("newdir/")
	if err != nil {
		t.Fatalf("CreateFile (dir) error: %v", err)
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Fatal("directory should exist")
	}
	if !info.IsDir() {
		t.Fatal("should be a directory")
	}
}

func TestCreateFileInSubdir(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	// Find and move to "src" directory.
	for i, node := range ft.Visible {
		if node.Name == "src" && node.IsDir {
			ft.Cursor = i
			break
		}
	}

	path, err := ft.CreateFile("handler.go")
	if err != nil {
		t.Fatalf("CreateFile error: %v", err)
	}

	expected := filepath.Join(dir, "src", "handler.go")
	if path != expected {
		t.Fatalf("expected path %q, got %q", expected, path)
	}
}

func TestCreateNestedPath(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	ft.Cursor = 0

	// Create with nested path — intermediate dirs should be created.
	path, err := ft.CreateFile("pkg/models/user.go")
	if err != nil {
		t.Fatalf("CreateFile nested error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("nested file should exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "pkg", "models")); os.IsNotExist(err) {
		t.Fatal("intermediate directories should exist")
	}
}

func TestCreateFileDuplicate(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	ft.Cursor = 0

	_, err := ft.CreateFile("main.go")
	if err == nil {
		t.Fatal("expected error for duplicate file")
	}
}

func TestRename(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	// Find main.go.
	for i, node := range ft.Visible {
		if node.Name == "main.go" {
			ft.Cursor = i
			break
		}
	}

	newPath, err := ft.Rename("app.go")
	if err != nil {
		t.Fatalf("Rename error: %v", err)
	}

	expected := filepath.Join(dir, "app.go")
	if newPath != expected {
		t.Fatalf("expected %q, got %q", expected, newPath)
	}

	// Old file should not exist.
	if _, err := os.Stat(filepath.Join(dir, "main.go")); !os.IsNotExist(err) {
		t.Fatal("old file should not exist")
	}
	// New file should exist.
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Fatal("new file should exist")
	}
}

func TestRenameRoot(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	ft.Cursor = 0 // Root node.
	_, err := ft.Rename("something")
	if err == nil {
		t.Fatal("expected error when renaming root")
	}
}

func TestRenameDuplicate(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	// Find go.mod.
	for i, node := range ft.Visible {
		if node.Name == "go.mod" {
			ft.Cursor = i
			break
		}
	}

	// Try to rename to an existing file.
	_, err := ft.Rename("main.go")
	if err == nil {
		t.Fatal("expected error for rename to existing file")
	}
}

func TestDelete(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	// Find main.go.
	for i, node := range ft.Visible {
		if node.Name == "main.go" {
			ft.Cursor = i
			break
		}
	}

	err := ft.Delete()
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "main.go")); !os.IsNotExist(err) {
		t.Fatal("file should be deleted")
	}
}

func TestDeleteDirectory(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	// Find docs directory.
	for i, node := range ft.Visible {
		if node.Name == "docs" && node.IsDir {
			ft.Cursor = i
			break
		}
	}

	err := ft.Delete()
	if err != nil {
		t.Fatalf("Delete dir error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "docs")); !os.IsNotExist(err) {
		t.Fatal("directory should be deleted")
	}
}

func TestDeleteRoot(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	ft.Cursor = 0
	err := ft.Delete()
	if err == nil {
		t.Fatal("expected error when deleting root")
	}
}

func TestCursorNodeHelpers(t *testing.T) {
	dir := setupTestDir(t)
	ft, _ := New(dir)

	// Root.
	ft.Cursor = 0
	if !ft.CursorIsRoot() {
		t.Fatal("expected cursor on root")
	}
	if ft.CursorNodeName() == "" {
		t.Fatal("expected non-empty name")
	}
	if ft.CursorNodePath() == "" {
		t.Fatal("expected non-empty path")
	}

	// Not root.
	ft.Cursor = 1
	if ft.CursorIsRoot() {
		t.Fatal("cursor should not be on root")
	}
}
