package input

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/israelcorrea/crit-ide/internal/events"
)

// Handler reads terminal events from tcell and translates them into
// application events on the bus. It runs in its own goroutine.
type Handler struct {
	screen tcell.Screen
	bus    *events.Bus
	// Drag tracking state.
	btn1Down bool // Is Button1 currently held?
	anchorX  int  // Screen X where Button1 was first pressed.
	anchorY  int  // Screen Y where Button1 was first pressed.
	// Double-click detection.
	lastClickTime int64 // Unix milliseconds of last click.
	lastClickX    int
	lastClickY    int
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
		case *tcell.EventMouse:
			h.handleMouse(tev)
		case *tcell.EventResize:
			h.bus.Send(events.Event{Type: events.EventResize})
		}
	}
}

// handleMouse translates a tcell mouse event into action events.
// Supports: left click, drag selection, right click, and wheel scroll.
func (h *Handler) handleMouse(ev *tcell.EventMouse) {
	x, y := ev.Position()
	btn := ev.Buttons()

	// Handle wheel events first (they can coexist with button state).
	switch {
	case btn&tcell.WheelUp != 0:
		h.bus.Send(events.Event{
			Type:     events.EventAction,
			ActionID: "mouse.scroll",
			Payload:  events.MouseScrollPayload{Direction: -3, ScreenX: x, ScreenY: y},
		})
		return
	case btn&tcell.WheelDown != 0:
		h.bus.Send(events.Event{
			Type:     events.EventAction,
			ActionID: "mouse.scroll",
			Payload:  events.MouseScrollPayload{Direction: 3, ScreenX: x, ScreenY: y},
		})
		return
	}

	// Right click (Button2 = secondary button in tcell).
	if btn&tcell.Button2 != 0 {
		h.bus.Send(events.Event{
			Type:     events.EventAction,
			ActionID: "menu.open",
			Payload:  events.MouseClickPayload{ScreenX: x, ScreenY: y},
		})
		return
	}

	// Left button drag tracking.
	if btn&tcell.Button1 != 0 {
		if !h.btn1Down {
			// Button just pressed — record anchor.
			h.btn1Down = true
			h.anchorX = x
			h.anchorY = y
			// Don't send click yet; wait for release to distinguish click vs drag.
		} else if x != h.anchorX || y != h.anchorY {
			// Button held and position changed — this is a drag.
			h.bus.Send(events.Event{
				Type:     events.EventAction,
				ActionID: "mouse.drag",
				Payload: events.MouseDragPayload{
					AnchorX:  h.anchorX,
					AnchorY:  h.anchorY,
					CurrentX: x,
					CurrentY: y,
				},
			})
		}
		return
	}

	// Button released (ButtonNone).
	if h.btn1Down {
		h.btn1Down = false
		if x == h.anchorX && y == h.anchorY {
			now := time.Now().UnixMilli()
			// Detect double-click (within 400ms and same position).
			if now-h.lastClickTime < 400 && x == h.lastClickX && y == h.lastClickY {
				// Double-click: first send click to position cursor, then select word.
				h.bus.Send(events.Event{
					Type:     events.EventAction,
					ActionID: "mouse.click",
					Payload:  events.MouseClickPayload{ScreenX: x, ScreenY: y},
				})
				h.bus.Send(events.Event{
					Type:     events.EventAction,
					ActionID: "select.word",
				})
				h.lastClickTime = 0 // Reset to prevent triple-click triggering another double.
			} else {
				// Single click.
				h.bus.Send(events.Event{
					Type:     events.EventAction,
					ActionID: "mouse.click",
					Payload:  events.MouseClickPayload{ScreenX: x, ScreenY: y},
				})
				h.lastClickTime = now
				h.lastClickX = x
				h.lastClickY = y
			}
		}
		// If position differs, the drag events already handled it.
	}
}

// handleKey translates a tcell key event into an action event.
// Hardcoded keymap. Will be replaced by a configurable keymap engine.
func (h *Handler) handleKey(ev *tcell.EventKey) {
	// Ctrl+Space (NUL) for manual completion trigger.
	if ev.Key() == tcell.KeyNUL || (ev.Key() == tcell.KeyRune && ev.Rune() == ' ' && ev.Modifiers()&tcell.ModCtrl != 0) {
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "completion.trigger"})
		return
	}

	// Check for modifier+key combinations first.
	if ev.Modifiers()&tcell.ModCtrl != 0 {
		switch ev.Key() {
		case tcell.KeyCtrlS:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "file.save"})
			return
		case tcell.KeyCtrlQ:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "app.quit"})
			return
		case tcell.KeyCtrlC:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "clipboard.copy"})
			return
		case tcell.KeyCtrlX:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "clipboard.cut"})
			return
		case tcell.KeyCtrlV:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "clipboard.paste"})
			return
		case tcell.KeyCtrlA:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "select.all"})
			return
		case tcell.KeyCtrlB:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "tree.toggle"})
			return
		case tcell.KeyCtrlW:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "tab.close"})
			return
		case tcell.KeyCtrlK:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "lsp.hover"})
			return
		case tcell.KeyCtrlG:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "goto.line"})
			return
		case tcell.KeyCtrlL:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "lsp.format"})
			return
		case tcell.KeyCtrlZ:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "edit.undo"})
			return
		case tcell.KeyCtrlY:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "edit.redo"})
			return
		case tcell.KeyCtrlD:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "edit.duplicate_line"})
			return
		case tcell.KeyCtrlF:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "search.open"})
			return
		case tcell.KeyCtrlP:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "finder.open"})
			return
		}
	}

	// Ctrl+Arrow for word movement.
	if ev.Modifiers()&tcell.ModCtrl != 0 {
		switch ev.Key() {
		case tcell.KeyLeft:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "cursor.word_left"})
			return
		case tcell.KeyRight:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "cursor.word_right"})
			return
		}
	}

	// Ctrl+Shift combinations (when terminal supports it).
	if ev.Modifiers()&tcell.ModCtrl != 0 && ev.Modifiers()&tcell.ModShift != 0 {
		if ev.Key() == tcell.KeyRune && (ev.Rune() == 'P' || ev.Rune() == 'p') {
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "palette.open"})
			return
		}
		if ev.Key() == tcell.KeyRune && (ev.Rune() == 'F' || ev.Rune() == 'f') {
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "project.search"})
			return
		}
		if ev.Key() == tcell.KeyRune && (ev.Rune() == 'G' || ev.Rune() == 'g') {
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "git.status"})
			return
		}
	}

	// Ctrl+. for code actions.
	if ev.Modifiers()&tcell.ModCtrl != 0 && ev.Key() == tcell.KeyRune && ev.Rune() == '.' {
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "lsp.code_action"})
		return
	}

	// Function keys.
	switch ev.Key() {
	case tcell.KeyF1:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "palette.open"})
		return
	case tcell.KeyF2:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "lsp.rename"})
		return
	case tcell.KeyF3:
		if ev.Modifiers()&tcell.ModShift != 0 {
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "search.prev"})
		} else {
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "search.next"})
		}
		return
	case tcell.KeyF5:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "project.search"})
		return
	case tcell.KeyF6:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "git.graph"})
		return
	case tcell.KeyF12:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "lsp.definition"})
		return
	}

	// Tab switching: Ctrl+PageDown / Ctrl+PageUp for next/prev tab.
	if ev.Modifiers()&tcell.ModCtrl != 0 {
		switch ev.Key() {
		case tcell.KeyPgDn:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "tab.next"})
			return
		case tcell.KeyPgUp:
			h.bus.Send(events.Event{Type: events.EventAction, ActionID: "tab.prev"})
			return
		}
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
	case tcell.KeyEscape:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "input.escape"})
	case tcell.KeyBacktab:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "edit.dedent"})
	case tcell.KeyTab:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "edit.indent"})
	case tcell.KeyRune:
		h.bus.Send(events.Event{Type: events.EventAction, ActionID: "insert.char", Payload: ev.Rune()})
	}
}
