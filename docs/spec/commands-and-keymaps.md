# Command System and Keymaps

## Command System

### Requirements

- Command palette UI
- Internal commands registered by ID
- Commands exposed to plugins
- Optional arguments
- Commands invocable by keybind or menu

### Example Command IDs

- `file.open`
- `buffer.next`
- `lsp.definition`
- `git.diff.open`
- `ai.explain.selection`

## Input and Keymaps

### Philosophy

Everything the user does is an **Action**. Keybinds trigger Actions.

### Requirements

- Customizable keymaps
- Context-based namespaces
- Chord mappings
- Optional leader key
- Binding by focus context, not modal editing
- Configurable mouse events
- Fallback default keymap

### Input Contexts

- `editor`
- `file_tree`
- `terminal`
- `popup`
- `diff_viewer`
- `global`

### Configuration Format

```toml
[keymap.global]
"ctrl+p" = "file.find"
"ctrl+shift+p" = "command.palette"

[keymap.editor]
"f12" = "lsp.definition"
"f2" = "lsp.rename"
"leader g d" = "git.diff.open"
```

### Input Pipeline (Full)

1. Terminal event arrives
2. Parser normalizes the event
3. Input context resolver determines the active context
4. Keymap matcher resolves chord/binding
5. Action dispatcher executes the action
6. State changes
7. Render scheduler triggers partial repaint

### Example Flow

```
Ctrl+P
→ KeyEvent{Ctrl:true, Key:"p"}
→ Context: editor
→ Resolve binding: file.find
→ Dispatch action
→ Open fuzzy finder popup
→ Re-render popup + statusline
```
