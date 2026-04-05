package input

import (
	"github.com/gdamore/tcell/v2"
	"github.com/israelcorrea/crit-ide/internal/events"
)

// Handler reads terminal events from tcell and translates them into
// application events on the bus. It runs in its own goroutine.
type Handler struct {
	screen tcell.Screen
	bus    *events.Bus
}

// NewHandler creates a new input handler.
func NewHandler(screen tcell.Screen, bus *events.Bus) *Handler {
	return &Handler{screen: screen, bus: bus}
}

// Run is the input goroutine loop. It blocks on PollEvent and sends
// translated events to the bus. It returns when the screen is finalized
// or a nil event is received.
func (h *Handler) Run() {
	for {
		ev := h.screen.PollEvent()
		if ev == nil {
			return
		}
		switch tev := ev.(type) {
		case *tcell.EventKey:
			h.handleKey(tev)
		case *tcell.EventResize:
			h.bus.Send(events.Event{Type: events.EventResize})
		}
	}
}

// handleKey translates a tcell key event into an action event.
// Sprint 1: hardcoded keymap. Sprint 3 will extract this to a configurable keymap engine.
func (h *Handler) handleKey(ev *tcell.EventKey) {
	// Check for modifier+key combinations first.
	if ev.Modifiers()&tcell.ModCtrl != 0 {
		switch ev.Key() {
		case tcell.KeyCtrlS:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "file.save"})
			return
		case tcell.KeyCtrlQ:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "app.quit"})
			return
		case tcell.KeyCtrlK:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "lsp.hover"})
			return
		case tcell.KeyCtrlG:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "lsp.definition"})
			return
		case tcell.KeyCtrlL:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "lsp.format"})
			return
		}
	}

	// Function keys.
	switch ev.Key() {
	case tcell.KeyF12:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "lsp.definition"})
		return
	}

	// Special keys.
	switch ev.Key() {
	case tcell.KeyUp:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "cursor.up"})
	case tcell.KeyDown:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "cursor.down"})
	case tcell.KeyLeft:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "cursor.left"})
	case tcell.KeyRight:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "cursor.right"})
	case tcell.KeyHome:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "cursor.home"})
	case tcell.KeyEnd:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "cursor.end"})
	case tcell.KeyPgUp:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "scroll.up"})
	case tcell.KeyPgDn:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "scroll.down"})
	case tcell.KeyEnter:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "insert.newline"})
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "delete.backward"})
	case tcell.KeyDelete:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "delete.forward"})
	case tcell.KeyTab:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "insert.char", Payload: '\t'})
	case tcell.KeyRune:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "insert.char", Payload: ev.Rune()})
	}
}
