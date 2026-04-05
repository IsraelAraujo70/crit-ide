# Package: `internal/editor`

The editor package contains the core text editing types. It has **zero internal dependencies** — it depends only on the Go standard library. This makes it the foundation that every other package builds upon.

## Files

| File | Purpose |
|------|---------|
| `textstore.go` | `TextStore` interface, `Position`, `Range` types |
| `linestore.go` | `LineStore` — the `[]string` implementation of `TextStore` |
| `buffer.go` | `Buffer` — document model with cursor, editing, and file I/O |

## Key Types

### `Position`

A zero-based `{Line, Col}` coordinate within a document. **Col is a byte offset** within the line's UTF-8 string, consistent with Go's string indexing.

### `Range`

A span between two `Position` values. Start is inclusive, end is exclusive.

### `TextStore` (interface)

The central abstraction that protects the entire codebase from the underlying text storage implementation.

```go
type TextStore interface {
    Insert(pos Position, text string) error
    Delete(r Range) error
    Line(n int) string
    LineCount() int
    Slice(r Range) string
    Content() string
}
```

**Why this matters**: The current `TextStore` is a simple `[]string` (one string per line). When performance requires it, we can swap to a rope or piece table by implementing a new struct that satisfies this interface — without changing a single line in `buffer.go`, `actions/`, `render/`, or `app/`.

### `LineStore`

The concrete `TextStore` implementation. Stores lines as `[]string`.

- `Insert` handles both single-line and multi-line (newline-containing) inserts
- `Delete` validates range ordering (Start must be <= End) before operating
- `Slice` returns text within a range, joining lines with `\n`
- `Content()` joins all lines with `\n`

**Trade-offs**: Simple and correct. O(n) for insertions in the middle of the line slice. Good enough for files under ~100K lines. For larger files, the `TextStore` interface allows swapping to a rope/piece table.

### `Buffer`

Represents an open document with cursor and selection state. One buffer can be displayed in multiple views (planned).

```go
type Buffer struct {
    ID         BufferID
    Path       string
    Kind       BufferKind    // File or Scratch
    Text       TextStore
    Dirty      bool
    ReadOnly   bool
    CursorRow  int
    CursorCol  int           // Byte offset within the line
    desiredCol int           // Sticky column for Up/Down movement
}
```

**Editing methods** — all are Unicode-aware:
- `InsertChar(ch rune)` — advances `CursorCol` by `len(string(ch))` bytes
- `InsertNewline()` — splits line, moves cursor to next line col 0
- `DeleteBackward()` — uses `utf8.DecodeLastRuneInString` to find rune boundary
- `DeleteForward()` — uses `utf8.DecodeRuneInString` to find rune boundary

**Cursor movement**:
- `MoveCursor(dir)` — Up/Down use sticky column (`desiredCol`), Left/Right wrap across lines
- `CursorHome()` / `CursorEnd()` — beginning/end of current line
- `ClampCursor()` — ensures cursor stays within document bounds

**Sticky column**: When moving vertically through lines of different lengths, the cursor remembers its desired column. Moving up from col 10 through a 3-character line and back down restores col 10. Horizontal movements and edits update the desired column.

**File I/O**:
- `LoadFile(id, path)` — reads file, normalizes `\r\n` to `\n`, strips trailing newline
- `SaveFile()` — writes content + trailing newline to disk, clears `Dirty` flag

## Design Decisions

1. **Byte-offset columns**: `CursorCol` is a byte offset, not a rune index. This avoids `[]rune` conversion on every operation and stays consistent with Go's string slicing. The renderer handles the byte-to-visual-column conversion.

2. **No undo/redo yet**: The `Buffer` struct is designed to accommodate an `UndoManager` field. The `TextStore` interface's `Insert`/`Delete` operations are the natural points to capture undo entries.

3. **Single selection**: The `Selection` struct tracks anchor (drag start) and cursor (drag end) positions. Selection-aware mutations (`InsertChar`, `InsertNewline`, `DeleteBackward`, `DeleteForward`) replace or delete selected text. Multi-cursor is planned for later.
