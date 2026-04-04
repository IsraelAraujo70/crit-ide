# crit-ide — Product Requirements Document

## Vision

crit-ide is a complete terminal IDE written in Go. It aims to deliver a professional development environment that runs entirely in the terminal, combining the power of keyboard-driven workflows with the accessibility of modern IDEs.

It is **not** a Vim/Neovim/Emacs clone. It inherits the best ideas from terminal editors — configurability, composability, speed — while rejecting their worst — modal editing as a requirement, hostile mouse UX, fragmented configuration.

## Core Principles

1. **Terminal-first** — Runs entirely in a modern TTY. No GUI, no Electron.
2. **Non-modal by default** — Direct editing like VS Code. No normal/insert mode dichotomy.
3. **Keyboard-first, mouse-enabled** — Everything works via keyboard. Mouse is a first-class citizen.
4. **Everything is an Action** — Keybinds, clicks, commands, and plugins all dispatch actions. Actions are the single mechanism for state mutation.
5. **Performance matters** — Fast startup, smooth scrolling, non-blocking UI. AI inference and LSP responses never freeze the editor.
6. **AI as a native feature** — Local AI autocomplete and assistance are part of the architecture, not bolted-on plugins.
7. **Simple configuration** — TOML-based config. No Lua scripts required for basic customization.

## Target Users

### Primary

Developers who:
- Work extensively in the terminal
- Want more integration than raw Vim but less bloat than VS Code
- Want local AI with privacy (no code sent to the cloud by default)
- Want high-level customization without configuration hell

### Secondary

- VS Code users who want to migrate to the terminal
- Neovim users tired of scattered plugin configuration
- JetBrains users who want a lightweight terminal-native option

## Product Scope — V1

### In Scope

| Area | Features |
|------|----------|
| **Editor** | Multi-buffer editing, splits (H/V), tabs/workspaces, undo/redo |
| **Navigation** | File tree, fuzzy finder, goto symbol, recent files |
| **Search** | Buffer search, project-wide search/replace, regex support |
| **Language** | Syntax highlighting, LSP (diagnostics, hover, definition, rename, completion, formatting, code actions) |
| **Autocomplete** | Hybrid: buffer words + LSP + local AI, with ranking |
| **Git** | Status, stage/unstage, commit, diff viewer (side-by-side + inline), hunk actions |
| **Terminal** | Embedded terminal pane, multiple sessions |
| **Input** | Configurable keymaps with context awareness, chord/leader key support, mouse events |
| **UI** | Statusline, command palette, popups, diagnostics panel |
| **Config** | TOML files (global + per-project), hot reload |
| **Plugins** | External process plugins via simple RPC |
| **Themes** | Configurable color schemes |

### Out of Scope (V1)

- Full DAP debugger
- Real-time collaboration
- GUI outside the terminal
- Cross-project semantic search at scale
- Advanced refactoring without LSP

### Post-V1 Roadmap

- DAP/debugging integration
- Profiler integration
- Remote development
- Pair programming
- Semantic search across repositories
- AI agent workflows
- Notebook-like panes

## Architecture Overview

### Layered Model

```
Layer 4 — External Integrations
  git binary, language servers, local AI (Ollama), shells, rg/fd

Layer 3 — UI (TUI)
  panes, widgets, menus, popups, statusline, mouse handling

Layer 2 — Services
  LSP, Git, AI, Search, Tasks, Config, PluginHost

Layer 1 — Core
  AppState, Buffers, Layout, Actions, Input Router, Render Scheduler
```

### Application State

```
AppState
 ├── Config
 ├── WorkspaceManager
 ├── BufferManager
 ├── LayoutTree
 ├── ActionRegistry
 ├── CommandRegistry
 ├── LSPManager
 ├── GitService
 ├── AIService
 ├── PluginHost
 ├── TaskRunner
 └── UIState
```

### Event-Driven Flow

```
User Input → Event Bus → Main Loop → Action → State Mutation → Render
                              ↑
              Async workers (LSP, Git, AI) produce events
```

All state mutations are serialized in the main event loop. Heavy work (LSP requests, AI inference, Git operations, grep) runs in async workers that produce events consumed by the main loop.

### Concurrency Model

- **Main goroutine**: event loop — consumes events, runs actions, triggers render
- **Input goroutine**: polls terminal events, translates to actions via keymap
- **Worker goroutines**: LSP, Git, AI, Search — produce events, never mutate state directly

No locks on core state. Workers send results as events. The main loop is the single writer.

## Technology Stack

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Language | Go | Fast compilation, single binary, good concurrency |
| Terminal UI | tcell | Low-level control needed for IDE-grade rendering |
| Config | TOML | Cleaner than JSON, more predictable than YAML |
| Text storage | Piece table / rope (behind interface) | Efficient for large files, natural undo/redo |
| LSP client | Custom (stdio JSON-RPC) | Minimal, no heavy framework |
| Git | Shell out to `git` binary | Pragmatic, covers all features |
| AI | Ollama API (local) | Privacy-first, low latency |
| Plugins (V1) | External process + JSON-RPC | Isolation, language-agnostic |
| Parsing | Tree-sitter (Go bindings) | Incremental, accurate highlighting |

## Key Differentiators

1. **A real terminal IDE**, not just an editor with plugins
2. **Non-modal by default** with fully configurable keymaps
3. **Native local AI** for autocomplete and code assistance
4. **Native Git + diff viewer**, not a plugin afterthought
5. **Well-defined plugin system** from day one
6. **Mouse support that actually works** in the terminal
7. **Action-oriented architecture** — composable, extensible, debuggable

## Success Criteria (V1)

- Opens real-world projects without freezing
- Comfortable editing experience
- Solid mouse support
- Useful LSP integration (diagnostics, completion, navigation)
- Functional Git diff viewing with hunk operations
- Working hybrid autocomplete (LSP + AI)
- Simple keymap configuration
- Architecture ready for community plugins

## Development Phases

| Phase | Focus | Key Deliverables |
|-------|-------|-----------------|
| **Sprint 1** | Foundation | Event loop, buffer, cursor, rendering, file I/O |
| **Sprint 2** | Multi-buffer | BufferManager, splits, statusline, mouse |
| **Sprint 3** | Commands | Command registry, keymap engine, command palette, fuzzy open |
| **Sprint 4** | Visual | Syntax highlighting, file tree, project search |
| **Sprint 5** | Language | LSP (definition, hover, diagnostics, completion) |
| **Sprint 6** | Git | Status panel, diff viewer, stage/unstage |
| **Sprint 7** | Terminal | Terminal pane, tasks, problem matcher |
| **Sprint 8** | AI | Inline completion, explain selection, config reload |
| **Post-V1** | Plugins | Plugin API, external process plugins |
