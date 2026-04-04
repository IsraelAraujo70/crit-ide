# AI Integration and Autocomplete

## Objective

Hybrid autocomplete with prioritization:

1. Buffer/project suggestions
2. LSP suggestions
3. Local AI model suggestions

## Use Cases

- Line completion
- Block completion
- Boilerplate suggestion
- Error explanation
- Test generation
- Refactor selected code

## Requirements

- Run model locally
- Low latency
- Streaming suggestions
- Inference cancellation
- Fallback when model is offline
- Current file context
- Nearby relevant file context
- Configurable token/context limit

## Two-Level Strategy

### Level 1 — Inline Completion

- Ghost text overlay
- Accept with `Tab` or configurable binding
- Cancel on continued typing

### Level 2 — Assist Panel

- Panel for quick prompts
- Contextual actions on selection
- Explain, refactor, docstring, tests

## Backend Abstraction

```go
type CompletionProvider interface {
    Complete(ctx context.Context, req CompletionRequest) (<-chan CompletionChunk, error)
}

type AIProvider interface {
    CompleteInline(ctx context.Context, req InlineCompletionRequest) (<-chan InlineSuggestion, error)
    RunAction(ctx context.Context, req AIActionRequest) (<-chan AIActionChunk, error)
}
```

Supported local engines:

- **Ollama** (first priority)
- llama.cpp server
- vLLM local (future)

## Completion Pipeline

Autocomplete is not a single source — it's an aggregator of providers.

### Providers

- Buffer words
- Snippets
- File paths
- LSP completion
- AI inline completion
- AI list completion (optional)

### Suggestion Interface

```go
type Suggestion struct {
    Label         string
    InsertText    string
    Detail        string
    Kind          string
    Score         float64
    Source        string
    IsInlineGhost bool
}

type SuggestionProvider interface {
    Name() string
    Suggestions(ctx context.Context, req SuggestionRequest) ([]Suggestion, error)
}

type CompletionEngine struct {
    Providers []SuggestionProvider
}
```

### Ranking

```
final score =
    sourceWeight
  + prefixMatchWeight
  + syntaxContextWeight
  + recencyWeight
  + userAcceptanceWeight
```

### UX Rule

Don't mix ghost text with an aggressive popup at the same time.

- Inline ghost text when confidence is high
- Popup when there are multiple good suggestions

## Integrating Local AI Without Breaking UX

This is the make-or-break point for whether the IDE feels good or sluggish.

### Mandatory Rules

- Cancelable requests
- Short debounce (100-250ms)
- Timeout
- Never block typing
- Small, relevant context
- Cache recent suggestions

### Recommended Flow

1. User pauses for 100-250ms
2. Current context is captured
3. Async request is sent
4. If user continues typing, previous request is cancelled
5. Response arrives via streaming
6. Ghost text appears if it still makes sense

### Context Builder (Minimum)

- Current line prefix
- Nearby suffix
- Current function/block
- Nearby imports
- File name
- Language ID
- LSP symbols (if available)

### Anti-patterns

- Sending the entire file every time
- Sending the entire project every time
- Waiting for a complete response before updating the UI

## Planned Action IDs

- `ai.complete.inline`
- `ai.explain.selection`
- `ai.refactor.selection`
- `ai.generate.test`
