package events

// EventType classifies the kind of event flowing through the bus.
type EventType int

const (
	EventAction         EventType = iota // An action should be executed.
	EventQuit                            // The application should terminate.
	EventResize                          // The terminal was resized.
	EventLSPDiagnostics                  // LSP diagnostics received. Payload: *lsp.DiagnosticsPayload.
	EventLSPDefinition                   // LSP go-to-definition result. Payload: *lsp.DefinitionPayload.
	EventLSPHover                        // LSP hover result. Payload: *lsp.HoverPayload.
	EventLSPFormat                       // LSP format result. Payload: *lsp.FormatPayload.
	EventLSPServerState                  // LSP server state change. Payload: *lsp.ServerStatePayload.
)

// Event is the message passed through the event bus.
type Event struct {
	Type     EventType
	ActionID string // Populated for EventAction.
	Payload  any    // Optional data (e.g., rune for insert.char).
}

// Bus is a simple buffered channel-based event bus.
// It decouples producers (input goroutine) from the consumer (main event loop).
type Bus struct {
	ch chan Event
}

// NewBus creates a new event bus with the given buffer capacity.
func NewBus(size int) *Bus {
	return &Bus{ch: make(chan Event, size)}
}

// Send enqueues an event. Non-blocking: if the buffer is full, the event
// is dropped to prevent deadlocking the input goroutine.
func (b *Bus) Send(e Event) {
	select {
	case b.ch <- e:
	default:
		// Buffer full — drop event rather than deadlock.
	}
}

// Recv returns the receive-only channel for consuming events in the main loop.
func (b *Bus) Recv() <-chan Event {
	return b.ch
}

// MouseClickPayload carries raw screen coordinates from a mouse click event.
type MouseClickPayload struct {
	ScreenX int
	ScreenY int
}

// MouseScrollPayload carries scroll direction and screen position.
type MouseScrollPayload struct {
	Direction int // Negative = up, positive = down (number of lines).
	ScreenX   int // Reserved for future multi-pane scroll targeting.
	ScreenY   int // Reserved for future multi-pane scroll targeting.
}

// MouseDragPayload carries anchor (drag start) and current positions.
type MouseDragPayload struct {
	AnchorX  int
	AnchorY  int
	CurrentX int
	CurrentY int
}
