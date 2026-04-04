# Package: `internal/input`

The input package handles raw terminal events and translates them into action events on the bus. It runs as a dedicated goroutine.

## Files

| File | Purpose |
|------|---------|
| `handler.go` | `Handler` — input goroutine with keymap translation |

## How It Works

```
tcell.PollEvent() → Handler.handleKey() → bus.Send(Event{ActionID: "..."})
```

The `Handler` sits in a tight loop calling `screen.PollEvent()`. When a key event arrives, it translates it to an action ID using the keymap and sends it to the event bus. The main loop consumes these events and executes the corresponding actions.

### Event Types Handled

| tcell Event | Translation |
|-------------|------------|
| `EventKey` | Translated to an action via the keymap |
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

### Ctrl Key Handling

tcell represents Ctrl+letter combinations as special key constants (`KeyCtrlS`, `KeyCtrlQ`). The handler checks for `ModCtrl` modifier first, then falls through to special keys and rune events. This ordering ensures Ctrl combinations take priority over rune insertion.
