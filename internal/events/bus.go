package events

// EventType classifies the kind of event flowing through the bus.
type EventType int

const (
	EventAction EventType = iota // An action should be executed.
	EventQuit                    // The application should terminate.
	EventResize                  // The terminal was resized.
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
