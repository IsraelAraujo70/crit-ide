package filetree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CreateFile creates a new file at the given path relative to the cursor's
// directory. If name ends with "/", a directory is created instead.
// Returns the absolute path of the created entry.
func (ft *FileTree) CreateFile(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty name")
	}

	parent := ft.cursorDir()
	targetPath := filepath.Join(parent, name)

	// Ensure all parent directories exist.
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create directories: %w", err)
	}

	if strings.HasSuffix(name, "/") {
		// Create directory.
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return "", fmt.Errorf("create directory: %w", err)
		}
	} else {
		// Create file (fail if exists).
		if _, err := os.Stat(targetPath); err == nil {
			return "", fmt.Errorf("file already exists: %s", targetPath)
		}
		f, err := os.Create(targetPath)
		if err != nil {
			return "", fmt.Errorf("create file: %w", err)
		}
		f.Close()
	}

	ft.Refresh()
	ft.selectPath(targetPath)
	return targetPath, nil
}

// Rename renames the node at the cursor to the new name.
// Returns the new absolute path.
func (ft *FileTree) Rename(newName string) (string, error) {
	node := ft.CursorNode()
	if node == nil {
		return "", fmt.Errorf("no node selected")
	}
	if node == ft.Root {
		return "", fmt.Errorf("cannot rename root")
	}
	if newName == "" {
		return "", fmt.Errorf("empty name")
	}

	oldPath := node.Path
	newPath := filepath.Join(filepath.Dir(oldPath), newName)

	if oldPath == newPath {
		return oldPath, nil // No change.
	}

	// Check for conflicts.
	if _, err := os.Stat(newPath); err == nil {
		return "", fmt.Errorf("already exists: %s", filepath.Base(newPath))
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return "", fmt.Errorf("rename: %w", err)
	}

	ft.Refresh()
	ft.selectPath(newPath)
	return newPath, nil
}

// Delete removes the file or empty directory at the cursor.
// Non-empty directories are removed recursively.
func (ft *FileTree) Delete() error {
	node := ft.CursorNode()
	if node == nil {
		return fmt.Errorf("no node selected")
	}
	if node == ft.Root {
		return fmt.Errorf("cannot delete root")
	}

	if err := os.RemoveAll(node.Path); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	ft.Refresh()
	return nil
}

// CursorNodeName returns the name of the node at the cursor, or "".
func (ft *FileTree) CursorNodeName() string {
	node := ft.CursorNode()
	if node == nil {
		return ""
	}
	return node.Name
}

// CursorNodePath returns the path of the node at the cursor, or "".
func (ft *FileTree) CursorNodePath() string {
	node := ft.CursorNode()
	if node == nil {
		return ""
	}
	return node.Path
}

// CursorIsRoot returns true if the cursor is on the root node.
func (ft *FileTree) CursorIsRoot() bool {
	node := ft.CursorNode()
	return node != nil && node == ft.Root
}

// cursorDir returns the directory path where new files should be created.
// If cursor is on a directory, that directory is used.
// If cursor is on a file, its parent directory is used.
func (ft *FileTree) cursorDir() string {
	node := ft.CursorNode()
	if node == nil {
		return ft.RootPath
	}
	if node.IsDir {
		// Make sure it's expanded so the new file appears visible.
		if !node.Expanded {
			node.Expanded = true
			_ = loadChildren(node)
			ft.rebuildVisible()
		}
		return node.Path
	}
	return filepath.Dir(node.Path)
}

// selectPath moves the cursor to the visible node with the given path.
func (ft *FileTree) selectPath(path string) {
	for i, n := range ft.Visible {
		if n.Path == path {
			ft.Cursor = i
			return
		}
	}
}
