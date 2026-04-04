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
3. Register all actions
4. Load file (or create scratch buffer)
5. Launch input goroutine
6. Initial render
7. Event loop:
    for !quit {
        event ← bus.Recv()
        switch event.Type:
            Action → registry.Execute(actionID, ctx) → ensureCursorVisible()
            Resize → screen.Sync() → ensureCursorVisible()
            Quit   → set quit flag
        render()
    }
8. Cleanup tcell screen
```

### Why This Design Matters

The event loop is the **single point of state mutation**. This eliminates race conditions by design:

- The input goroutine only sends events — it never reads or writes application state
- Actions run synchronously within the loop — they can freely mutate `Buffer`, `ScrollY`, etc.
- Future async workers (LSP, Git, AI) will send results as events, consumed by this same loop
- Rendering always sees a consistent state snapshot

### `ensureCursorVisible()`

After every action and resize, this method adjusts `ScrollY` to keep the cursor within the viewport:

```go
if cursorRow < scrollY          → scrollY = cursorRow
if cursorRow >= scrollY + height → scrollY = cursorRow - height + 1
```

## `App` Struct

```go
type App struct {
    screen   tcell.Screen
    bus      *events.Bus
    registry *actions.Registry
    renderer *render.Renderer
    buffer   *editor.Buffer
    scrollY  int
    quit     bool
    filePath string
}
```

Sprint 1 has a single buffer. Sprint 2 will replace `buffer` with a `BufferManager` and add a `LayoutTree`.

## `AppState` Interface Implementation

`App` implements `actions.AppState`, which is the interface actions use:

| Method | Description |
|--------|-------------|
| `ActiveBuffer()` | Returns the currently focused buffer |
| `ScrollY()` | Current vertical scroll offset |
| `SetScrollY(y)` | Set scroll offset |
| `ViewportHeight()` | Visible editor rows (screen height minus statusline) |
| `Quit()` | Sets the quit flag |

This interface will grow as features are added:
- Sprint 2: `BufferManager()`, `Layout()`
- Sprint 5: `LSPManager()`
- Sprint 6: `GitService()`

## Startup Flow

```
New(filePath) → App struct with bus and registry
Run():
  ├─ tcell.NewScreen() + Init()
  ├─ EnableMouse()
  ├─ NewRenderer(screen)
  ├─ RegisterAll(registry)      ← registers all 14 actions
  ├─ LoadFile(filePath) or NewBuffer("scratch")
  ├─ go inputHandler.Run()      ← starts input goroutine
  ├─ render()                   ← initial frame
  └─ event loop                 ← blocks until quit
```

## Concurrency Model (Sprint 1)

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
- Action execution errors are silently ignored in Sprint 1 (logged in future sprints)
- tcell initialization failures are fatal (returned from `Run`)
