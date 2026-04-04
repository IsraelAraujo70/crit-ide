# Package: `internal/render`

The render package draws the editor state to the terminal screen using tcell. It is a pure consumer of state — it reads `ViewState` and produces screen output, never mutating application state.

## Files

| File | Purpose |
|------|---------|
| `renderer.go` | `Renderer` — draws buffer content, line numbers, cursor, statusline |

## Key Types

### `ViewState`

```go
type ViewState struct {
    Buffer  *editor.Buffer
    ScrollY int
    Width   int    // Terminal width in columns
    Height  int    // Terminal height in rows (including statusline)
}
```

`ViewState` decouples the renderer from the `app` package. The app constructs a `ViewState` snapshot and passes it to `Render()`. This makes the renderer testable and independent.

### `Renderer`

```go
renderer := render.NewRenderer(screen)
renderer.Render(&viewState)
```

## Rendering Pipeline

Each `Render()` call performs a full redraw:

1. **Clear screen** — `screen.Clear()`
2. **Draw editor area** — for each visible row:
   - Calculate which document line maps to this screen row (`ScrollY + row`)
   - Draw line number in the gutter (right-aligned, dimmed)
   - Highlight current line's gutter number (bright white)
   - Draw line content, expanding tabs to 4-space stops
   - Draw `~` for rows beyond the document
3. **Draw statusline** — bottom row with:
   - Left: filename + dirty flag (`[+]`)
   - Right: cursor position (`Ln X, Col Y`)
4. **Position cursor** — `screen.ShowCursor(x, y)` at the correct screen position
5. **Flush** — `screen.Show()`

## Gutter (Line Numbers)

The gutter width adapts to the document size:
- Files with < 100 lines: 4 columns (min 3 digits + 1 space)
- Files with 1000+ lines: 5 columns
- Formula: `max(3, digits(lineCount)) + 1`

## Tab Rendering

Tabs are expanded to spaces using 4-column tab stops:
```
spaces = 4 - (currentCol % 4)
```

This means a tab at column 0 produces 4 spaces, at column 1 produces 3 spaces, etc.

## Cursor Positioning

The cursor's screen position accounts for:
- **Gutter offset**: cursor X = gutterWidth + visual column
- **Tab expansion**: each tab before the cursor expands to variable spaces
- **Scroll offset**: cursor Y = cursorRow - scrollY

The `screenCol` method iterates through the line's runes up to `CursorCol` (byte offset), accumulating visual columns with tab expansion. This correctly handles the byte-offset cursor model.

If the cursor is outside the visible viewport (shouldn't happen if `ensureCursorVisible` works), the cursor is hidden.

## Design Decisions

### Full Redraw (Sprint 1)

Sprint 1 does a full `Clear()` + redraw on every frame. This is simple and correct. For Sprint 2+, diff rendering (only redrawing changed cells) will improve performance, especially during fast typing. The tcell library supports cell-level diffing natively via `Show()` which only flushes changed cells to the terminal.

### Statusline

The statusline occupies the last row of the terminal. `editorHeight = Height - 1`. This is a fixed layout for Sprint 1. Sprint 2 will introduce a flexible panel system.

### Style Constants

Colors are defined inline in Sprint 1:
- Default text: white on default background
- Gutter: gray on default background
- Current line gutter: bright white
- Statusline: black on white (inverted)

Sprint 4 will introduce a theme system with TOML-configurable color schemes.
