# Package: `internal/events`

The events package provides the communication backbone between the input goroutine and the main event loop. It has **zero internal dependencies**.

## Files

| File | Purpose |
|------|---------|
| `bus.go` | `Event` type, `EventType` constants, `Bus` struct |

## Key Types

### `EventType`

```go
const (
    EventAction  // An action should be executed
    EventQuit    // The application should terminate
    EventResize  // The terminal was resized
)
```

### `Event`

```go
type Event struct {
    Type     EventType
    ActionID string    // Which action to run (for EventAction)
    Payload  any       // Optional data (e.g., rune for insert.char)
}
```

### `Bus`

A thin wrapper around a buffered Go channel.

```go
bus := events.NewBus(256)  // 256-event buffer
bus.Send(event)            // Non-blocking — drops event if buffer full
<-bus.Recv()               // Blocking receive in the main loop
```

## Design Decisions

### Non-blocking Send

`Send` uses a `select` with `default` to avoid deadlocking the input goroutine when the buffer is full:

```go
func (b *Bus) Send(e Event) {
    select {
    case b.ch <- e:
    default:
        // Drop rather than deadlock
    }
}
```

This is critical because the input goroutine calls `Send` synchronously from `PollEvent`. If the main loop is slow (e.g., during a large file save), a blocking send would freeze the input goroutine, which in turn blocks tcell's event processing, creating a deadlock.

### Why Not a More Complex Event System?

A channel is the simplest correct solution for the current two-goroutine model. As more async workers (LSP, Git, AI) are added, the bus may evolve to support:
- Event priorities (resize > action)
- Event coalescing (multiple rapid resize events → single render)
- Typed event channels per subsystem

But those are future concerns. The current `Bus` interface is narrow enough to swap implementations without changing producers or consumers.
