# Architecture Deep Dive

## Concurrency and Scheduling

### The Classic Go Trap

Creating goroutines for everything and losing control of state.

### Core Rule

UI and editor state must be mutated in a serialized manner.

### Strategy

- 1 main event/state loop
- Separate workers for heavy tasks
- Results come back as events to the main loop

### Tasks That Can Run in Workers

- LSP requests
- Git status refresh
- Global grep
- AI inference
- Heavy parsing

### Tasks That Must NOT Mutate State Directly

- Any parallel worker

### Recommended Pattern

```
parallel worker
→ produces Event
→ main event loop consumes
→ updates AppState
→ schedules render
```

This prevents race conditions and erratic UI behavior.

## Rendering Architecture

### Requirements

- Incremental rendering
- Screen diffing when possible
- Avoid full redraw on every keystroke
- Render scheduler decoupled from heavy processing

### Recommended Separation

- **Input loop** — polls terminal events
- **Event loop** — processes events, runs actions, mutates state
- **Render loop** — draws current state to screen

AI inference or LSP responses must **never** block rendering.

## Event System

### Event-Driven Architecture

Almost everything communicates through events and actions:

- Input generates an action
- Action changes state or calls a service
- Service publishes an event
- UI reacts to the new state

### Core Events

| Event | Trigger |
|-------|---------|
| `BufferOpened` | New buffer loaded |
| `BufferChanged` | Text edited |
| `DiagnosticsUpdated` | LSP sends new diagnostics |
| `GitStatusUpdated` | Git state refresh completes |
| `CompletionRequested` | User triggers completion |
| `CompletionReceived` | Provider returns suggestions |
| `LayoutChanged` | Splits/panes modified |
| `PluginCrashed` | Plugin process died |

This reduces coupling and enables the plugin system.

## Full Data Model (Target)

### Buffer (Full Spec)

```go
type Buffer struct {
    ID           BufferID
    Path         string
    Kind         BufferKind
    Rope         TextRope       // or PieceTable
    Version      int64
    Dirty        bool
    ReadOnly     bool
    LanguageID   string
    LineEndings  LineEnding
    Encoding     string
    CursorState  []Cursor
    Selections   []Selection
    Undo         *UndoManager
    Metadata     map[string]any
}
```

### EditorView (Full Spec)

```go
type EditorView struct {
    ID              ViewID
    BufferID        BufferID
    ScrollX         int
    ScrollY         int
    Width           int
    Height          int
    ShowLineNumbers bool
    Wrap            bool
    Focused         bool
}
```

### Text Storage

Don't use raw `string` for the entire buffer text in serious editing. It degrades fast.

**Candidate structures:**

| Structure | Pros | Cons |
|-----------|------|------|
| Piece table | Natural undo/redo, good for large files | More complex to implement |
| Rope | Balanced tree of strings, good split/join | Memory overhead |
| Gap buffer | Simple, cache-friendly | Poor for large insertions |

**Decision**: Piece table or rope, behind the `TextStore` interface. V0/V1 uses a simple line-based store, already hidden behind the interface for seamless swap.

## Buffer Types (Full Spec)

- File buffer — backed by a file on disk
- Ephemeral buffer — temporary, not saved
- Search result buffer — grep/search output
- Diff buffer — diff view content
- Log buffer — internal logs
- Terminal session buffer — embedded terminal output

## Core Editor Features (Full Spec)

- Open, create, save, save as, reload
- Multiple simultaneous buffers
- External file change detection
- Multi-step undo/redo
- Selection via keyboard and mouse
- Copy/cut/paste
- Multiple cursors (roadmap)
- Automatic indentation
- Comment/uncomment toggle
- Auto-pair delimiters
- Line numbers
- Soft wrap / hard wrap
- Highlight: selection, current word, matching brackets
