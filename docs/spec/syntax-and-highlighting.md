# Syntax Highlighting and Parsing

## Requirements

- Per-language highlighting
- Code folding (future)
- Identification of comments, strings, keywords
- Matching bracket highlighting
- Incremental parsing ideally

## Current Implementation (V1)

Regex-based per-line tokenizer using `internal/highlight/`. Ships with 14 built-in languages: Go, Rust, JavaScript, TypeScript, Python, C, C++, Markdown, JSON, HTML, CSS, Shell, TOML, YAML.

Each language is a `LanguageDef` struct with regex patterns and comment delimiters. The `RegexHighlighter` applies patterns per-line with block comment state tracking across lines.

Theme lives in `internal/theme/` and maps `TokenType` to `tcell.Style`.

## Roadmap

### V2 — Tree-sitter

Replace regex tokenizer with Tree-sitter via stable Go bindings for accurate, incremental parsing. The `Highlighter` interface already abstracts the backend, so swapping implementations requires no renderer changes.

### V3 — Extensible Language Support via Plugins

The highlight system is designed for future extensibility. Users and third-party extensions will be able to install new language definitions without modifying the core:

- **TOML/JSON language definition files**: A language can be described declaratively (patterns, keywords, comment delimiters) and loaded at runtime from `~/.config/crit-ide/languages/` or a plugin directory.
- **Plugin-contributed languages**: The [plugin system](plugins.md) will allow plugins to register new `LanguageDef` entries via the plugin API (`registerLanguage`), following the same pattern as commands and keymaps.
- **Community language packs**: Installable packages (e.g., `crit-ide install lang-haskell`) that drop a language definition into the registry.
- **Override built-in languages**: Users can override a built-in language definition by providing their own file with the same extensions, enabling customization of keyword lists, token patterns, etc.

The `LangRegistry` already supports dynamic registration — adding a new language at runtime is a single `Register()` call with a `LanguageDef`. The extension layer will build on this foundation.

### Theme Integration

Highlighting tokens map to theme colors. See [config-and-themes.md](config-and-themes.md) for the theme system.

### Typing Flow

When the user types code:

1. Buffer updates
2. Highlighter invalidates changed lines and re-tokenizes visible range
3. LSP receives incremental sync (didChange)
4. Completion pipeline prepares suggestions
5. AI is triggered when appropriate
6. Ghost text or popup appears
