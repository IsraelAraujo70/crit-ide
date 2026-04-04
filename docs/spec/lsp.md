# Language Server Protocol (LSP)

## Requirements

- Start/stop language servers
- Real-time diagnostics
- Hover information
- Go to definition
- Go to references
- Rename symbol
- Completion
- Signature help
- Formatting
- Code actions
- Document symbols
- Workspace symbols
- Semantic tokens (if supported)

## Architecture

### Components

- `LSPManager` — manages server lifecycle per language/workspace
- `LSPClient` — JSON-RPC communication over stdio
- `LanguageRegistry` — maps file types to language server commands
- `DiagnosticsStore` — aggregates diagnostics from all servers

### Server Lifecycle Flow

1. Buffer opens
2. Language is detected (by file extension / config)
3. Corresponding server starts
4. Editing events generate incremental document sync
5. Responses update diagnostics/completion/navigation

## Client Implementation Details

### Recommendation

Build a small, custom client. Don't over-engineer a framework early.

### Client Responsibilities

- Process spawn management
- stdin/stdout JSON-RPC communication
- Request ID tracking
- Pending requests map
- Async notification handling
- Capabilities negotiation

### Server States

```
stopped → starting → ready → degraded → crashed
                       ↑                    │
                       └────── restart ─────┘
```

### Workspace Strategy

- One server per workspace/language when it makes sense
- Or multiplex buffers through a single server instance

### Edge Cases to Handle from Day One

| Edge Case | Solution |
|-----------|----------|
| Server hangs | Timeout + kill/restart |
| Out-of-order responses | Match by request ID, validate version |
| Request cancellation | Send `$/cancelRequest` |
| Unsupported capability | Check capabilities before calling |
| Stale diagnostics | Validate buffer version on response |

### Response Validation

Every response must check:

- Does the buffer still exist?
- Is the version still current?
- Is the pane still visible?

If not, discard the response.

## Planned Action IDs

- `lsp.definition`
- `lsp.references`
- `lsp.rename`
- `lsp.hover`
- `lsp.format`
- `lsp.codeAction`
