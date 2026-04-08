# CLAUDE.md — Agent Context for crit-ide

## Project Overview

crit-ide is a terminal IDE written in Go. Non-modal, action-driven, event-loop based.

## Build & Test

```bash
go build -o crit-ide ./cmd/ide   # Build binary
go test ./...                     # Run all tests
go vet ./...                      # Static analysis
./crit-ide <file>                 # Run (opens file in terminal editor)
```

## Architecture — Must Understand

**Event-driven, two goroutines, zero locks.**

```
Input Goroutine (tcell.PollEvent) → Event Bus → Main Loop → Action → Render
```

- **Main loop** is the ONLY place state is mutated. Never mutate state from a worker goroutine.
- **Actions** are the ONLY mechanism for state mutation. Every user operation is an Action with a string ID.
- **TextStore** is an interface. Current impl is `LineStore` ([]string). Future: rope/piece table. Never bypass the interface.
- **CursorCol is a byte offset**, not a rune index. All column math must be UTF-8 aware.
- **Bus.Send is non-blocking** — drops events when buffer is full to prevent deadlock.

## Package Dependency Rules

```
editor    → nothing (pure types)
events    → nothing (types + channel)
clipboard → atotto/clipboard (external only)
filetree  → nothing (os, path/filepath, sort, strings)
actions   → editor, events
input     → events, tcell
render    → editor, tcell
app       → all of the above (including filetree)
```

**Never create circular imports.** Actions access app state through the `AppState` interface, not by importing the `app` package.

## Key Interfaces

```go
// internal/editor/textstore.go — ALL text operations go through this
type TextStore interface {
    Insert(pos Position, text string) error
    Delete(r Range) error
    Line(n int) string
    LineCount() int
    Slice(r Range) string
    Content() string
}

// internal/actions/action.go — ALL state mutations go through this
type Action interface {
    ID() string
    Run(ctx *ActionContext) error
}

// internal/actions/action.go — breaks circular dep between actions and app
type AppState interface {
    ActiveBuffer() *editor.Buffer
    ScrollY() int
    SetScrollY(y int)
    ViewportHeight() int
    ScreenWidth() int
    Quit()
    Clipboard() ClipboardProvider
    InputMode() InputMode
    SetInputMode(mode InputMode)
    ContextMenu() *editor.MenuState
    SetContextMenu(menu *editor.MenuState)
    PostAction(actionID string)
    // Tab / multi-buffer
    Buffers() []*editor.Buffer
    ActiveBufferIndex() int
    OpenFile(path string) error
    CloseBuffer(idx int)
    SwitchBuffer(idx int)
    // File tree
    FileTree() FileTreeState
    FileTreeVisible() bool
    SetFileTreeVisible(v bool)
    ToggleFileTree()
    FileTreeWidth() int
    // Focus
    FocusArea() FocusArea
    SetFocusArea(area FocusArea)
}
```

## Adding a New Action

1. Create a struct implementing `Action` in `internal/actions/`
2. Register it in `RegisterAll()` in `editor_actions.go`
3. Map a key to its ID in `internal/input/handler.go`
4. Add tests in the appropriate `_test.go` file

## Adding a New Feature / Service

1. Create package under `internal/<feature>/` with zero deps on `app`
2. Define interfaces the feature exposes
3. Wire it in `internal/app/app.go`
4. Extend `AppState` interface if actions need access
5. Heavy work goes in worker goroutines → results come back as events on the bus

## Code Conventions

- All docs and code comments in English
- Column positions are byte offsets (UTF-8), not rune indices
- Use `unicode/utf8` for rune-aware cursor operations
- Tests go in `_test.go` files in the same package
- No `init()` functions
- No global mutable state

## Documentation Structure

```
docs/
├── prd.md              # Product vision, scope, roadmap
├── progress.json       # Progress tracker (done/planned)
├── packages/           # How implemented code works (update when code changes)
└── spec/               # Feature specs (reference when implementing)
```

## Current State

- Multi-buffer editing with tab bar and text selection
- **Undo/redo** (Ctrl+Z / Ctrl+Y) with full edit history
- **Word movement** (Ctrl+Left / Ctrl+Right) for fast navigation
- **Duplicate line** (Ctrl+D)
- **Auto-indent** on Enter (copies leading whitespace)
- **Double-click** to select word
- **Syntax highlighting** via tree-sitter (14 languages: Go, Python, JS, TS, Rust, C, CSS, HTML, JSON, MD, TOML, YAML, Bash)
- **LSP integration** — diagnostics (inline underline), hover (Ctrl+K), go-to-definition (Ctrl+G/F12), format (Ctrl+L)
- File tree panel (NeoTree-style) on the right side with toggle (Ctrl+B)
- 47 registered actions (cursor, edit, file, scroll, mouse, clipboard, selection, context menu, file tree, tabs, undo, word-movement, LSP)
- Mouse: click, double-click, drag-select, wheel scroll, right-click context menu, tab click, tree click
- Clipboard: Ctrl+C/X/V, system clipboard via atotto/clipboard
- Tab management: Ctrl+W close, Ctrl+PgDn/PgUp switch, mouse click tabs
- File tree: keyboard navigation (arrows, Enter), expand/collapse dirs, click to open, create/rename/delete
- Focus routing: editor vs file tree, with smart action remapping
- Hardcoded keymap (configurable keymap engine planned)
- Full redraw rendering (diff rendering planned)
- No splits, no Git integration, no AI yet

## Roadmap

See `docs/progress.json` for full feature tracker and `docs/prd.md` for the development phases.
