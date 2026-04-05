# Package: `internal/app`

The app package is the top-level orchestrator. It initializes all subsystems, owns the main event loop, and implements the `AppState` interface that actions use to interact with application state.

## Files

| File | Purpose |
|------|---------|
| `app.go` | `App` struct, `Run()` method, `AppState` implementation |

## The Main Event Loop

This is the architectural heart of crit-ide. The `Run()` method:

```
1. Initialize tcell screen
2. Create renderer
3. Initialize clipboard
4. Register all actions
5. Load file (or create scratch buffer)
6. Launch input goroutine
7. Initial render
8. Event loop:
    for !quit {
        event ← bus.Recv()
        switch event.Type:
            Action → route by InputMode → registry.Execute() → pendingActions → ensureCursorVisible()
            Resize → screen.Sync() → ensureCursorVisible()
            Quit   → set quit flag
        render()
    }
9. Cleanup tcell screen
```

### Why This Design Matters

The event loop is the **single point of state mutation**. This eliminates race conditions by design:

- The input goroutine only sends events — it never reads or writes application state
- Actions run synchronously within the loop — they can freely mutate `Buffer`, `ScrollY`, etc.
- Future async workers (LSP, Git, AI) will send results as events, consumed by this same loop
- Rendering always sees a consistent state snapshot

### Input Mode Routing

The main loop routes events based on `InputMode`:

- **ModeNormal**: All actions are executed normally.
- **ModeContextMenu**: Only `menu.*` actions are allowed. Keyboard actions are remapped (arrows → menu navigation, Enter → execute, Escape → close). Mouse clicks are forwarded to `menu.click` which decides if the click is inside or outside the menu.

### Pending Actions

Menu execution uses a trampoline: `menu.execute` posts an action ID via `PostAction()`, then the main loop picks it up and executes it after the menu closes. This avoids actions needing access to the registry.

### `ensureCursorVisible()`

After every action and resize, this method adjusts `ScrollY` to keep the cursor within the viewport:

```go
if cursorRow < scrollY          → scrollY = cursorRow
if cursorRow >= scrollY + height → scrollY = cursorRow - height + 1
```

## `App` Struct

```go
type App struct {
    screen        tcell.Screen
    bus           *events.Bus
    registry      *actions.Registry
    renderer      *render.Renderer
    buffer        *editor.Buffer
    scrollY       int
    quit          bool
    filePath      string
    clip          ClipboardProvider
    inputMode     InputMode
    contextMenu   *editor.MenuState
    pendingAction string
}
```

Currently has a single buffer. Will be replaced with a `BufferManager` and `LayoutTree` when multi-buffer support is added.

## `AppState` Interface Implementation

`App` implements `actions.AppState`, which is the interface actions use:

| Method | Description |
|--------|-------------|
| `ActiveBuffer()` | Returns the currently focused buffer |
| `ScrollY()` / `SetScrollY()` | Vertical scroll offset |
| `ViewportHeight()` | Visible editor rows (screen height minus statusline) |
| `ScreenWidth()` | Terminal width in columns |
| `Quit()` | Sets the quit flag |
| `Clipboard()` | Returns the clipboard provider |
| `InputMode()` / `SetInputMode()` | Current input routing mode |
| `ContextMenu()` / `SetContextMenu()` | Active context menu state |
| `PostAction()` | Queue an action for the trampoline |

This interface will grow as features are added (BufferManager, LSP, Git, etc.).

## Startup Flow

```
New(filePath) → App struct with bus and registry
Run():
  ├─ tcell.NewScreen() + Init()
  ├─ EnableMouse()
  ├─ NewRenderer(screen)
  ├─ SystemClipboard init
  ├─ RegisterAll(registry)      ← registers all 28 actions
  ├─ LoadFile(filePath) or NewBuffer("scratch")
  ├─ go inputHandler.Run()      ← starts input goroutine
  ├─ render()                   ← initial frame
  └─ event loop                 ← blocks until quit
```

## Concurrency Model

```
┌──────────────┐     Event Bus      ┌──────────────────┐
│ Input        │ ──── Send ────────→ │ Main Loop        │
│ Goroutine    │                     │ (state + render) │
│              │ ←── PollEvent ───── │                  │
│ (tcell)      │                     │ (single writer)  │
└──────────────┘                     └──────────────────┘
```

Two goroutines, zero locks. The bus channel is the only synchronization primitive.

## Error Handling

- If the file doesn't exist at startup, a new file buffer is created with the given path (so the user can write and save)
- Action execution errors are silently ignored (logging planned)
- tcell initialization failures are fatal (returned from `Run`)
