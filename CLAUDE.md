# CLAUDE.md — Agent Context for crit-ide

## Project Overview

crit-ide is a terminal IDE written in Go. Non-modal, action-driven, event-loop based. Currently at Sprint 1 (core foundation).

## Build & Test

```bash
go build -o crit-ide ./cmd/ide   # Build binary
go test ./...                     # Run all tests
go vet ./...                      # Static analysis
./crit-ide <file>                 # Run (opens file in terminal editor)
```

## Architecture — Must Understand

**Event-driven, two goroutines, zero locks.**

```
Input Goroutine (tcell.PollEvent) → Event Bus → Main Loop → Action → Render
```

- **Main loop** is the ONLY place state is mutated. Never mutate state from a worker goroutine.
- **Actions** are the ONLY mechanism for state mutation. Every user operation is an Action with a string ID.
- **TextStore** is an interface. Current impl is `LineStore` ([]string). Future: rope/piece table. Never bypass the interface.
- **CursorCol is a byte offset**, not a rune index. All column math must be UTF-8 aware.
- **Bus.Send is non-blocking** — drops events when buffer is full to prevent deadlock.

## Package Dependency Rules

```
editor  → nothing (pure types)
events  → nothing (types + channel)
actions → editor, events
input   → events, tcell
render  → editor, tcell
app     → all of the above
```

**Never create circular imports.** Actions access app state through the `AppState` interface, not by importing the `app` package.

## Key Interfaces

```go
// internal/editor/textstore.go — ALL text operations go through this
type TextStore interface {
    Insert(pos Position, text string) error
    Delete(r Range) error
    Line(n int) string
    LineCount() int
    Slice(r Range) string
    Content() string
}

// internal/actions/action.go — ALL state mutations go through this
type Action interface {
    ID() string
    Run(ctx *ActionContext) error
}

// internal/actions/action.go — breaks circular dep between actions and app
type AppState interface {
    ActiveBuffer() *editor.Buffer
    ScrollY() int
    SetScrollY(y int)
    ViewportHeight() int
    Quit()
}
```

## Adding a New Action

1. Create a struct implementing `Action` in `internal/actions/`
2. Register it in `RegisterAll()` in `editor_actions.go`
3. Map a key to its ID in `internal/input/handler.go`
4. Add tests in `editor_actions_test.go`

## Adding a New Feature / Service

1. Create package under `internal/<feature>/` with zero deps on `app`
2. Define interfaces the feature exposes
3. Wire it in `internal/app/app.go`
4. Extend `AppState` interface if actions need access
5. Heavy work goes in worker goroutines → results come back as events on the bus

## Code Conventions

- All docs and code comments in English
- Column positions are byte offsets (UTF-8), not rune indices
- Use `unicode/utf8` for rune-aware cursor operations
- Tests go in `_test.go` files in the same package
- No `init()` functions
- No global mutable state

## Documentation Structure

```
docs/
├── prd.md              # Product vision, scope, roadmap
├── progress.json       # Sprint tracker (done/planned)
├── packages/           # How implemented code works (update when code changes)
└── spec/               # Feature specs for future sprints (reference when implementing)
```

## Current State (Sprint 1)

- Single buffer editing
- 14 registered actions (cursor, edit, file, scroll, quit)
- Hardcoded keymap (configurable keymap in Sprint 3)
- Full redraw rendering (diff rendering in Sprint 2+)
- No undo/redo, no splits, no mouse click, no syntax highlighting, no LSP, no Git, no AI

## What's Next (Sprint 2)

BufferManager, multiple views per buffer, LayoutTree with splits, mouse click/scroll, enhanced statusline. See `docs/progress.json` for full roadmap.
