# crit-ide

A complete terminal IDE written in Go.

**Non-modal. Action-driven. AI-native.**

crit-ide aims to be a professional development environment that runs entirely in the terminal — combining the power of keyboard-driven workflows with the accessibility of modern IDEs. It is not a Vim clone. It inherits the best ideas from terminal editors (configurability, composability, speed) while rejecting their worst (forced modal editing, hostile mouse UX, fragmented configuration).

## Status

Core editing with mouse support is working. See [progress](docs/progress.json) for details.

## Quick Start

```bash
go build -o crit-ide ./cmd/ide
./crit-ide path/to/file.go
```

### Keybindings

| Key | Action |
|-----|--------|
| Arrow keys | Move cursor |
| Home / End | Start / end of line |
| PageUp / PageDown | Scroll viewport |
| Enter | New line |
| Backspace / Delete | Delete character |
| Tab | Insert tab |
| Ctrl+S | Save file |
| Ctrl+Q | Quit |
| Ctrl+C / Ctrl+X / Ctrl+V | Copy / Cut / Paste |
| Ctrl+A | Select all |
| Escape | Clear selection |

### Mouse

| Action | Behavior |
|--------|----------|
| Left click | Position cursor |
| Left drag | Select text |
| Right click | Context menu (Copy, Cut, Paste, Select All) |
| Wheel up/down | Scroll viewport |

## Architecture

crit-ide is built on an **event-driven, action-oriented architecture**:

```
Input Goroutine → Event Bus → Main Event Loop → Action → State Mutation → Render
```

- **Everything is an Action** — keybinds, clicks, commands, and plugins all dispatch actions
- **Serialized state** — all mutations happen in the main event loop (no locks)
- **TextStore interface** — text storage is abstracted for future swap to rope/piece table
- **Non-modal** — direct editing, no normal/insert mode

### Package Structure

```
cmd/ide/            Entry point
internal/
├── app/            Main event loop, AppState
├── editor/         Buffer, TextStore, Selection, cursor, file I/O
├── events/         Event bus (non-blocking channel)
├── actions/        Action interface, registry, 28 concrete actions
├── input/          Input goroutine, keymap translation, mouse/drag handling
├── render/         tcell renderer, line numbers, statusline, popup menus
└── clipboard/      System clipboard abstraction
```

## Roadmap

| Phase | Focus | Status |
|-------|-------|--------|
| Foundation | Event loop, buffer, cursor, rendering, file I/O | Done |
| Mouse & Selection | Click, drag select, scroll, right-click menu, clipboard | Done |
| Multi-buffer | BufferManager, splits, tabbed views | Planned |
| Commands | Command registry, configurable keymaps, command palette, fuzzy open | Planned |
| Visual | Syntax highlighting, file tree, project search | Planned |
| Language | LSP — diagnostics, hover, definition, completion | Planned |
| Git | Status panel, diff viewer, stage/unstage | Planned |
| Terminal | Terminal pane, tasks, problem matcher | Planned |
| AI | Inline completion, explain selection | Planned |

## Tech Stack

| Component | Choice |
|-----------|--------|
| Language | Go |
| Terminal UI | [tcell](https://github.com/gdamore/tcell) |
| Config | TOML |
| LSP | Custom stdio JSON-RPC client |
| Git | Shell out to `git` |
| AI | Ollama (local) |
| Text storage | LineStore now, rope/piece table later (behind interface) |
| Clipboard | [atotto/clipboard](https://github.com/atotto/clipboard) |

## Documentation

- **[Product Requirements](docs/prd.md)** — vision, scope, architecture overview
- **[Progress Tracker](docs/progress.json)** — feature status and deliverables
- **[Package Docs](docs/packages/)** — how each implemented package works
- **[Feature Specs](docs/spec/)** — detailed specifications for future features

## Design Principles

1. **Terminal-first** — runs entirely in a modern TTY
2. **Non-modal by default** — direct editing, no mode switching
3. **Keyboard-first, mouse-enabled** — everything works via keyboard, mouse is first-class
4. **AI as a native feature** — not a bolted-on plugin
5. **Simple configuration** — TOML config, no Lua required for basics
6. **Performance** — fast startup, smooth scrolling, non-blocking UI

## License

MIT
