# Architectural Decisions, Risks and Guardrails

## Architectural Decisions

### Decision 1 — Non-modal by Default

No normal/insert mode as the foundation. Direct editing like modern IDEs.

### Decision 2 — Everything is an Action

Keybinds, clicks, commands, and plugins all dispatch actions. Actions are the single mechanism for state mutation.

### Decision 3 — External Services are Isolated

LSP, Git, AI, and plugins sit behind interfaces. They never directly mutate core state.

### Decision 4 — Core First, Plugins Later

Don't open a public API too early. First stabilize actions, commands, and events.

### Decision 5 — Git and AI are Native Features

Don't treat them as afterthoughts or bolt-on plugins. They are part of the architecture from day one.

## Correct Build Order

The fatal mistake would be starting with "pretty visuals" or the plugin system before stabilizing the core.

**The right order is:**

1. Editor + buffer + input + layout
2. Actions / commands / keymaps
3. LSP
4. Git / diff
5. Local AI
6. Plugins

Invert this and the project never ships.

## Technical Risks

### Risk 1 — Terminal Rendering

Building rich, fast, stable UI in the terminal is hard.

**Mitigation:**
- Keep the render loop simple
- Use diff rendering (only redraw changed cells)
- Test across multiple terminals (kitty, alacritty, ghostty, tmux)

### Risk 2 — Tree-sitter in Go

Bindings and integration may be troublesome.

**Mitigation:**
- Start with simple regex highlighting
- Isolate the parser behind an interface
- Swap to Tree-sitter when bindings are stable

### Risk 3 — Premature Plugin System

A poorly defined plugin system ruins the project.

**Mitigation:**
- First stabilize actions, commands, and events
- Only then open the public API
- Start with external process plugins (isolation)

### Risk 4 — Local AI Latency

AI inference can be too heavy for real-time completion.

**Mitigation:**
- Complete only when context makes sense (don't spam)
- Aggressive cancellation of stale requests
- Small models by default
- Streaming responses

### Risk 5 — Cross-Platform Terminal Support

Windows tends to be painful.

**Mitigation:**
- Linux and macOS first
- Windows support after core is stable
- tcell handles most cross-platform concerns

## Non-Functional Requirements

### Performance

- Fast startup
- Smooth scrolling
- Low relative memory consumption
- Reasonable support for large files
- Autocomplete must never freeze the UI

### Reliability

- Plugin crash must not bring down the IDE
- Optional autosave
- Session recovery (future)

### Portability

- **Linux**: primary platform
- **macOS**: supported
- **Windows**: deferred, if terminal architecture allows

### Security

- External plugins with clear permissions (future)
- Local AI does not send code externally by default

## Things NOT to Do Early

- Complex debugger (DAP)
- Custom language parser
- Giant public plugin API
- Overly sophisticated theme system
- Integration with 10 AI backends at once
- Excessive abstraction before the real flow exists

## V1 Success Criteria

V1 will be considered successful if:

- Opens real projects without freezing
- Editing code is comfortable
- Mouse support works well
- LSP is useful (diagnostics, completion, navigation)
- Git diffs are solid
- Hybrid autocomplete is functional
- Keymap configuration is simple
- Architecture is ready for plugins
