# Syntax Highlighting and Parsing

## Requirements

- Per-language highlighting
- Code folding (future)
- Identification of comments, strings, keywords
- Matching bracket highlighting
- Incremental parsing ideally

## Strategy

### Recommendation

Use Tree-sitter via stable Go bindings if feasible.

If Tree-sitter is too complex initially:

- **V1**: Simple regex-based highlighting
- **V2**: Real incremental parser via Tree-sitter

### Theme Integration

Highlighting tokens map to theme colors. See [config-and-themes.md](config-and-themes.md) for the theme system.

### Typing Flow

When the user types code:

1. Buffer updates
2. Parser/highlighter reacts
3. LSP receives incremental sync
4. Completion pipeline prepares suggestions
5. AI is triggered when appropriate
6. Ghost text or popup appears
