# Package: `internal/input`

The input package handles raw terminal events and translates them into action events on the bus. It runs as a dedicated goroutine.

## Files

| File | Purpose |
|------|---------|
| `handler.go` | `Handler` — input goroutine with keymap translation |

## How It Works

```
tcell.PollEvent() → Handler.handleKey()   → bus.Send(Event{ActionID: "..."})
                  → Handler.handleMouse() → bus.Send(Event{ActionID: "...", Payload: ...})
```

The `Handler` sits in a tight loop calling `screen.PollEvent()`. When an event arrives, it routes by type: key events go to the keymap translator, mouse events go to the mouse handler, and resize events are forwarded directly. The main loop consumes these events and executes the corresponding actions.

### Event Types Handled

| tcell Event | Translation |
|-------------|------------|
| `EventKey` | Translated to an action via the keymap |
| `EventMouse` | Translated to `mouse.click` or `mouse.scroll` actions with screen coordinates |
| `EventResize` | Sent as `EventResize` (triggers screen sync + re-render) |
| `nil` | Goroutine exits (screen was finalized) |

## Sprint 1 Keymap (Hardcoded)

| Key | Action ID |
|-----|-----------|
| Arrow Up/Down/Left/Right | `cursor.up` / `cursor.down` / `cursor.left` / `cursor.right` |
| Home / End | `cursor.home` / `cursor.end` |
| PageUp / PageDown | `scroll.up` / `scroll.down` |
| Enter | `insert.newline` |
| Backspace | `delete.backward` |
| Delete | `delete.forward` |
| Tab | `insert.char` (payload: `'\t'`) |
| Ctrl+S | `file.save` |
| Ctrl+Q | `app.quit` |
| Any printable rune | `insert.char` (payload: the rune) |

## Design Decisions

### Why a Separate Goroutine?

`tcell.PollEvent()` is a blocking call. If it ran in the main loop, the application couldn't process events from other sources (LSP responses, AI completions, Git status updates) while waiting for user input. The dedicated goroutine ensures the main loop is always responsive.

### Hardcoded Keymap (Sprint 1)

The keymap is hardcoded in `handleKey()` for Sprint 1. Sprint 3 will extract this into a configurable keymap engine that:
- Loads bindings from TOML config
- Resolves context-aware bindings (editor vs. file tree vs. popup)
- Supports chord sequences and leader keys
- Falls back to default bindings

The current structure already separates "what key was pressed" from "what action to run", making the extraction straightforward.

### Mouse Handling (Sprint 2)

Mouse events (`tcell.EventMouse`) are handled by `handleMouse()`. The handler converts tcell mouse events into actions:

| Mouse Event | Action ID | Payload |
|-------------|-----------|---------|
| Left click (`Button1`) | `mouse.click` | `MouseClickPayload{ScreenX, ScreenY}` |
| Wheel up (`WheelUp`) | `mouse.scroll` | `MouseScrollPayload{Direction: -3}` |
| Wheel down (`WheelDown`) | `mouse.scroll` | `MouseScrollPayload{Direction: 3}` |
| Other (right click, drag, motion) | Ignored | — |

The handler sends raw screen coordinates. Coordinate conversion (screen → buffer position) is done in the action itself, keeping the input handler free of editor/render dependencies.

The `mouse.click` action:
1. Ignores clicks on the statusline (screenY >= viewportHeight)
2. Ignores clicks on the gutter (screenX < gutterWidth)
3. Converts screen position to buffer row/col using `editor.GutterWidth()` and `editor.VisualColToByteOffset()`
4. Handles tabs (snap to tab byte offset) and UTF-8 correctly
5. Clamps to document bounds (past EOF → last line, past line end → end of line)

The `mouse.scroll` action scrolls the viewport by 3 lines per wheel event and adjusts the cursor if it scrolls out of view.

### Ctrl Key Handling

tcell represents Ctrl+letter combinations as special key constants (`KeyCtrlS`, `KeyCtrlQ`). The handler checks for `ModCtrl` modifier first, then falls through to special keys and rune events. This ordering ensures Ctrl combinations take priority over rune insertion.
