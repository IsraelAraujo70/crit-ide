# Package: filetree

## Purpose

Manages the state of the file explorer panel (NeoTree-style tree). Reads the filesystem, maintains a tree of nodes, and provides navigation/manipulation operations.

## Dependencies

- `os`, `path/filepath`, `sort`, `strings` (stdlib only)
- Zero dependencies on other crit-ide packages

## Key Types

### Node

```go
type Node struct {
    Name     string    // Base filename
    Path     string    // Absolute filesystem path
    IsDir    bool
    Expanded bool      // Only meaningful for directories
    Children []*Node   // Populated when expanded
    Depth    int       // Nesting level (root=0)
}
```

Implements the `actions.FileTreeNode` interface via `NodePath()` and `NodeIsDir()` methods.

### FileTree

```go
type FileTree struct {
    Root      *Node
    RootPath  string
    Cursor    int       // Index into flattened visible nodes
    ScrollY   int       // Scroll offset for tree viewport
    Visible   []*Node   // Flattened visible nodes (rebuilt on expand/collapse)
}
```

## Core Operations

| Method | Description |
|--------|-------------|
| `New(rootPath)` | Create tree, read top-level, expand root |
| `MoveUp() / MoveDown()` | Navigate cursor within visible nodes |
| `Toggle()` | Expand/collapse dir or return file path |
| `Expand() / Collapse()` | Explicit expand/collapse for arrow keys |
| `SetCursorToScreenRow(row)` | Convert screen row to cursor index |
| `EnsureCursorVisible(vpH)` | Adjust ScrollY to keep cursor in viewport |
| `Refresh()` | Reload from disk preserving expanded state |

## Filesystem Behavior

- **Hidden files** (names starting with `.`) are filtered out
- **Sort order**: directories first, then alphabetical (case-insensitive)
- **Lazy loading**: children are loaded when a directory is first expanded
- **Refresh** preserves the expanded/collapsed state of all directories

## Visible Node Flattening

The tree is flattened into a `Visible` slice via depth-first traversal. Only nodes whose parents are expanded appear in the list. This slice is rebuilt whenever a node is expanded or collapsed.

## Integration with App

The `App` creates a `FileTree` at startup using the file's directory (or CWD). The `FileTree` satisfies the `actions.FileTreeState` interface, which allows tree actions to interact with it without importing the `filetree` package.

## Rendering

The `render` package receives tree data via `ViewState.TreeNodes` (a slice of `render.TreeNode` structs built by the app from `filetree.Visible`). The renderer draws:
- Vertical border separating editor and tree
- "EXPLORER" header
- Indented nodes with directory arrows (▸/▾) and file type icons
- Cursor highlight when tree has focus
