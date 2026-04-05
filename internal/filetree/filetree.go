package filetree

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Node represents a single entry in the file tree (file or directory).
type Node struct {
	Name     string
	Path     string
	IsDir    bool
	Expanded bool
	Children []*Node
	Depth    int
}

// NodePath returns the filesystem path of this node (implements FileTreeNode).
func (n *Node) NodePath() string { return n.Path }

// NodeIsDir returns whether this node is a directory (implements FileTreeNode).
func (n *Node) NodeIsDir() bool { return n.IsDir }

// FileTree holds the state of the file explorer panel.
type FileTree struct {
	Root      *Node
	RootPath  string
	Cursor    int // Index into flattened visible nodes.
	ScrollY   int // Scroll offset for the tree viewport.
	Visible   []*Node
	maxScroll int
}

// New creates a new FileTree rooted at the given directory.
// It reads the top-level entries and expands the root.
func New(rootPath string) (*FileTree, error) {
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	root := &Node{
		Name:     filepath.Base(absPath),
		Path:     absPath,
		IsDir:    true,
		Expanded: true,
		Depth:    0,
	}

	if err := loadChildren(root); err != nil {
		return nil, err
	}

	ft := &FileTree{
		Root:     root,
		RootPath: absPath,
	}
	ft.rebuildVisible()
	return ft, nil
}

// loadChildren reads the directory contents of a node and populates its Children.
func loadChildren(node *Node) error {
	if !node.IsDir {
		return nil
	}

	entries, err := os.ReadDir(node.Path)
	if err != nil {
		return err
	}

	node.Children = nil
	for _, e := range entries {
		// Skip hidden files and common noise.
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		child := &Node{
			Name:  name,
			Path:  filepath.Join(node.Path, name),
			IsDir: e.IsDir(),
			Depth: node.Depth + 1,
		}
		node.Children = append(node.Children, child)
	}

	// Sort: directories first, then alphabetical.
	sort.Slice(node.Children, func(i, j int) bool {
		a, b := node.Children[i], node.Children[j]
		if a.IsDir != b.IsDir {
			return a.IsDir
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})

	return nil
}

// rebuildVisible flattens the tree into a list of visible nodes for rendering.
func (ft *FileTree) rebuildVisible() {
	ft.Visible = nil
	ft.flatten(ft.Root)
}

func (ft *FileTree) flatten(node *Node) {
	ft.Visible = append(ft.Visible, node)
	if node.IsDir && node.Expanded {
		for _, child := range node.Children {
			ft.flatten(child)
		}
	}
}

// VisibleCount returns the number of visible nodes.
func (ft *FileTree) VisibleCount() int {
	return len(ft.Visible)
}

// CursorNode returns the node at the current cursor position.
func (ft *FileTree) CursorNode() *Node {
	if ft.Cursor >= 0 && ft.Cursor < len(ft.Visible) {
		return ft.Visible[ft.Cursor]
	}
	return nil
}

// MoveUp moves the cursor up one visible node.
func (ft *FileTree) MoveUp() {
	if ft.Cursor > 0 {
		ft.Cursor--
	}
}

// MoveDown moves the cursor down one visible node.
func (ft *FileTree) MoveDown() {
	if ft.Cursor < len(ft.Visible)-1 {
		ft.Cursor++
	}
}

// Toggle expands or collapses the node at the cursor.
// For files, it returns the file path to be opened.
// For directories, it toggles expansion and returns "".
func (ft *FileTree) Toggle() string {
	node := ft.CursorNode()
	if node == nil {
		return ""
	}

	if node.IsDir {
		node.Expanded = !node.Expanded
		if node.Expanded && len(node.Children) == 0 {
			_ = loadChildren(node)
		}
		ft.rebuildVisible()
		return ""
	}

	// It's a file — return its path for opening.
	return node.Path
}

// Expand expands the directory at the cursor (no-op for files).
func (ft *FileTree) Expand() {
	node := ft.CursorNode()
	if node == nil || !node.IsDir {
		return
	}
	if !node.Expanded {
		node.Expanded = true
		if len(node.Children) == 0 {
			_ = loadChildren(node)
		}
		ft.rebuildVisible()
	}
}

// Collapse collapses the directory at the cursor, or moves to the parent.
func (ft *FileTree) Collapse() {
	node := ft.CursorNode()
	if node == nil {
		return
	}

	if node.IsDir && node.Expanded {
		node.Expanded = false
		ft.rebuildVisible()
		return
	}

	// Move to the immediate parent directory.
	// Search backwards from the cursor to find the closest ancestor.
	for i := ft.Cursor - 1; i >= 0; i-- {
		n := ft.Visible[i]
		if n.IsDir && n.Depth == node.Depth-1 && isDescendant(n, node) {
			ft.Cursor = i
			return
		}
	}
}

// isDescendant checks if child is a descendant of parent.
func isDescendant(parent, child *Node) bool {
	return strings.HasPrefix(child.Path, parent.Path+string(filepath.Separator))
}

// NodeAtScreenRow returns the node at a given row offset within the tree viewport.
// row is relative to the tree viewport top (0-based).
func (ft *FileTree) NodeAtScreenRow(row int) *Node {
	idx := ft.ScrollY + row
	if idx >= 0 && idx < len(ft.Visible) {
		return ft.Visible[idx]
	}
	return nil
}

// NodeAtScreenRowAsInterface returns the node as a FileTreeNode interface.
// This allows the actions package to use the node without importing filetree.
func (ft *FileTree) NodeAtScreenRowAsInterface(row int) interface{ NodePath() string; NodeIsDir() bool } {
	n := ft.NodeAtScreenRow(row)
	if n == nil {
		return nil
	}
	return n
}

// SetCursorToScreenRow sets the cursor to match a screen row click.
func (ft *FileTree) SetCursorToScreenRow(row int) {
	idx := ft.ScrollY + row
	if idx >= 0 && idx < len(ft.Visible) {
		ft.Cursor = idx
	}
}

// EnsureCursorVisible adjusts ScrollY to keep the cursor in view.
func (ft *FileTree) EnsureCursorVisible(viewportHeight int) {
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	if ft.Cursor < ft.ScrollY {
		ft.ScrollY = ft.Cursor
	}
	if ft.Cursor >= ft.ScrollY+viewportHeight {
		ft.ScrollY = ft.Cursor - viewportHeight + 1
	}
}

// Refresh reloads the tree from disk, preserving expanded states.
func (ft *FileTree) Refresh() {
	expanded := make(map[string]bool)
	ft.collectExpanded(ft.Root, expanded)

	root := &Node{
		Name:     filepath.Base(ft.RootPath),
		Path:     ft.RootPath,
		IsDir:    true,
		Expanded: true,
		Depth:    0,
	}
	_ = loadChildren(root)
	ft.restoreExpanded(root, expanded)

	ft.Root = root
	ft.rebuildVisible()

	// Clamp cursor.
	if ft.Cursor >= len(ft.Visible) {
		ft.Cursor = len(ft.Visible) - 1
	}
	if ft.Cursor < 0 {
		ft.Cursor = 0
	}
}

func (ft *FileTree) collectExpanded(node *Node, expanded map[string]bool) {
	if node.IsDir && node.Expanded {
		expanded[node.Path] = true
		for _, child := range node.Children {
			ft.collectExpanded(child, expanded)
		}
	}
}

func (ft *FileTree) restoreExpanded(node *Node, expanded map[string]bool) {
	if node.IsDir && expanded[node.Path] {
		node.Expanded = true
		if len(node.Children) == 0 {
			_ = loadChildren(node)
		}
		for _, child := range node.Children {
			ft.restoreExpanded(child, expanded)
		}
	}
}
