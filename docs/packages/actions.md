# Package: `internal/actions`

The actions package defines the Action system — the **single mechanism** for mutating application state. Every user-visible operation (keystroke, mouse click, command, plugin call) is an Action.

## Files

| File | Purpose |
|------|---------|
| `action.go` | `Action` interface, `ActionContext`, `AppState` interface, `Registry`, `InputMode`, `ClipboardProvider` |
| `editor_actions.go` | Cursor, text editing, file, scroll actions + `RegisterAll()` |
| `mouse_actions.go` | Mouse click and scroll actions |
| `selection_actions.go` | Drag selection, select all, escape, shared `screenToBufferPos()` helper |
| `clipboard_actions.go` | Copy, cut, paste actions |
| `menu_actions.go` | Context menu open/close/navigate/execute/click actions |

## Key Types

### `Action` (interface)

```go
type Action interface {
    ID() string
    Run(ctx *ActionContext) error
}
```

Every action has a unique string ID (e.g., `"cursor.up"`, `"file.save"`) and a `Run` method that receives the full application context.

### `ActionContext`

```go
type ActionContext struct {
    App   AppState       // Access to application state
    Event *events.Event  // The triggering event (contains Payload)
}
```

### `AppState` (interface)

This interface breaks the circular dependency between `actions` and `app`. Actions interact with application state through this interface, not by importing the `app` package directly.

```go
type AppState interface {
    ActiveBuffer() *editor.Buffer
    ScrollY() int
    SetScrollY(y int)
    ViewportHeight() int
    Quit()
}
```

As the application grows, this interface expands (e.g., `BufferManager()`, `LayoutTree()`, `LSPManager()`).

### `Registry`

Maps action IDs to implementations. Actions are registered at startup and looked up at runtime.

```go
reg := actions.NewRegistry()
actions.RegisterAll(reg)
reg.Execute("cursor.up", ctx)  // Looks up and runs the action
```

## Registered Actions

| ID | Description | Category |
|----|-------------|----------|
| `cursor.up` | Move cursor one line up | Navigation |
| `cursor.down` | Move cursor one line down | Navigation |
| `cursor.left` | Move cursor one character left (wraps to prev line) | Navigation |
| `cursor.right` | Move cursor one character right (wraps to next line) | Navigation |
| `cursor.home` | Move cursor to start of line | Navigation |
| `cursor.end` | Move cursor to end of line | Navigation |
| `insert.char` | Insert character from `Event.Payload` (rune) | Editing |
| `insert.newline` | Insert newline at cursor | Editing |
| `delete.backward` | Backspace — delete character before cursor | Editing |
| `delete.forward` | Delete — delete character at cursor | Editing |
| `file.save` | Save current buffer to disk | File |
| `app.quit` | Exit the application | App |
| `scroll.up` | Scroll viewport one page up | Scroll |
| `scroll.down` | Scroll viewport one page down | Scroll |

## Design Decisions

### Why Actions Instead of Direct Method Calls?

1. **Keybindings**: A keymap maps key combos to action IDs. The input handler doesn't need to know what each action does — it just sends the ID.
2. **Command palette**: The command palette lists all registered actions by ID.
3. **Plugins**: External plugins can trigger actions by ID via RPC.
4. **Logging/debugging**: Every state mutation flows through a single `Registry.Execute` call, making it easy to log, trace, or replay.
5. **Composability**: Future actions can compose existing ones (e.g., "save and format" = `file.save` + `lsp.format`).

### Scroll Actions Move the Cursor

`scroll.up` and `scroll.down` adjust `ScrollY` by one viewport height **and** move the cursor to stay visible. This matches user expectations — after PageDown, the cursor should be in the visible area, not off-screen.

The `scroll.down` action limits scroll to `lineCount - viewportHeight` to prevent scrolling past the end of the document into empty space.
