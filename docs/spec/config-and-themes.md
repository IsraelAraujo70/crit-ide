# Configuration and Themes

## Configuration

### Requirements

- Global user config file
- Per-project config file
- Local overrides
- Reload without restarting
- Validatable schema

### Configurable Areas

- Theme
- Keymaps
- Fonts
- Editor behavior
- LSP servers
- AI settings
- Git settings
- Plugins
- Tasks

### Format

**TOML** — because:

- Cleaner than JSON
- More predictable than YAML
- Great for configuration files

### File Locations

- Global: `~/.config/crit-ide/config.toml`
- Project: `<project_root>/.crit-ide.toml`

## Theme System

In the terminal, a bad theme kills the experience.

### Configurable Elements

| Category | Tokens |
|----------|--------|
| Base | foreground, background |
| Syntax | keyword, string, comment, function, type |
| Statusline | normal, warning, error |
| Diff | added, removed, header |
| Diagnostics | error, warning, info, hint |
| UI | popup borders, selection, cursor line |

### Example Theme

```toml
[theme]
name = "midnight"
background = "#0b1020"
foreground = "#d6deeb"

[theme.syntax]
keyword = "#c792ea"
string = "#ecc48d"
comment = "#637777"
function = "#82aaff"
```

## Observability

### Requirements

- Internal logs
- Log/debug panel
- Optional tracing
- Redraw latency measurement
- Autocomplete latency measurement
- Plugin crash capture

## Session Persistence (Future)

Even if session restore isn't delivered in V1, the architecture should leave room for it.

### What to Persist

- Open files
- Layout state
- Recent buffers
- Cursor position per file
- Folds (future)
- Command history

### Useful Local Cache

- Recent file index
- Autocomplete acceptance history
- Explorer expanded state
