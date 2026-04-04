# Plugin System

## Goal

Allow extension without turning the core into chaos.

## Requirements

- Plugins can register commands
- Plugins can register default keymaps
- Plugins can add panels
- Plugins can listen to events
- Plugins can contribute completion sources
- Plugins can contribute language features

## Architecture Options Evaluated

### Option A — External Process Plugins (Recommended for V1)

**Pros:**
- Isolation — plugin crash doesn't take down the IDE
- Language-agnostic — plugins can be written in any language
- Simpler security model

**Cons:**
- IPC is more complex
- Higher latency

### Option B — Compiled Go Plugins

**Pros:**
- Strong integration
- Fast

**Cons:**
- Delicate compatibility (Go plugin ABI)
- Worse deployment story

### Option C — Embedded Runtime (Lua, JS, Starlark)

**Pros:**
- Scriptable configuration
- Community tends to like it

**Cons:**
- Larger technical surface area

### Decision

- **V1**: Config + commands + extensions via external process with lightweight RPC
- **V2**: Add embedded scripting

## Minimum Plugin API

- Register command
- Register event handler
- Register completion source
- Register panel
- Register settings schema

## Important Events

- `on_startup`
- `on_buffer_open`
- `on_buffer_save`
- `on_diagnostics_update`
- `on_git_state_change`
- `on_command_executed`

## Plugin Manifest Example

```toml
name = "git-tools"
version = "0.1.0"
entry = "./git-tools-plugin"

[contributes.commands]
"git.openHistory" = "Open Git History"

[contributes.keymap.editor]
"leader g h" = "git.openHistory"
```

## RPC Protocol

### Lifecycle

1. IDE reads manifest
2. Spawns plugin process
3. Capabilities handshake
4. Registers commands/events
5. Exchanges JSON-RPC messages

### Minimum Messages

| Message | Direction | Purpose |
|---------|-----------|---------|
| `initialize` | IDE → Plugin | Start plugin with config |
| `shutdown` | IDE → Plugin | Graceful shutdown |
| `registerCommands` | Plugin → IDE | Declare available commands |
| `executeCommand` | IDE → Plugin | Run a command |
| `eventNotification` | IDE → Plugin | Notify of editor events |
| `requestCompletions` | IDE → Plugin | Ask for completion items |
| `providePanelData` | Plugin → IDE | Send panel content |

### Pragmatic Security (V1)

- Plugin is trusted by explicit installation
- Crash logging
- Request timeout
- Kill/restart process if it hangs
