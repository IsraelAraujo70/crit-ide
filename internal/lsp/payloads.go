package lsp

// DiagnosticsPayload is the event payload for LSP diagnostics.
type DiagnosticsPayload struct {
	URI         DocumentURI
	Diagnostics []Diagnostic
}

// DefinitionPayload is the event payload for go-to-definition results.
type DefinitionPayload struct {
	Locations []Location
}

// HoverPayload is the event payload for hover results.
type HoverPayload struct {
	Contents MarkupContent
	Range    *Range
}

// FormatPayload is the event payload for formatting results.
type FormatPayload struct {
	URI   DocumentURI
	Edits []TextEdit
}

// CompletionPayload is the event payload for completion results.
type CompletionPayload struct {
	Items        []CompletionItem
	IsIncomplete bool
}

// ServerStatePayload is the event payload for server state changes.
type ServerStatePayload struct {
	LangID string
	State  ServerState
	Error  string // Non-empty if state is Crashed or Degraded.
}

// ServerState represents the lifecycle state of a language server.
type ServerState int

const (
	StateStopped  ServerState = iota
	StateStarting
	StateReady
	StateDegraded
	StateCrashed
)
