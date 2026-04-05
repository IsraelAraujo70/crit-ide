package actions

import (
	"github.com/israelcorrea/crit-ide/internal/events"
)

// --- Tab next (Ctrl+Tab / Ctrl+PageDown) ---

type tabNext struct{}

func (a *tabNext) ID() string { return "tab.next" }

func (a *tabNext) Run(ctx *ActionContext) error {
	bufs := ctx.App.Buffers()
	if len(bufs) <= 1 {
		return nil
	}
	next := ctx.App.ActiveBufferIndex() + 1
	if next >= len(bufs) {
		next = 0
	}
	ctx.App.SwitchBuffer(next)
	return nil
}

// --- Tab previous (Ctrl+Shift+Tab / Ctrl+PageUp) ---

type tabPrev struct{}

func (a *tabPrev) ID() string { return "tab.prev" }

func (a *tabPrev) Run(ctx *ActionContext) error {
	bufs := ctx.App.Buffers()
	if len(bufs) <= 1 {
		return nil
	}
	prev := ctx.App.ActiveBufferIndex() - 1
	if prev < 0 {
		prev = len(bufs) - 1
	}
	ctx.App.SwitchBuffer(prev)
	return nil
}

// --- Tab close (Ctrl+W) ---

type tabClose struct{}

func (a *tabClose) ID() string { return "tab.close" }

func (a *tabClose) Run(ctx *ActionContext) error {
	bufs := ctx.App.Buffers()
	if len(bufs) <= 1 {
		// Don't close the last tab — just quit or leave it.
		return nil
	}
	ctx.App.CloseBuffer(ctx.App.ActiveBufferIndex())
	return nil
}

// --- Tab click (mouse click on tab bar) ---

type tabClick struct{}

func (a *tabClick) ID() string { return "tab.click" }

func (a *tabClick) Run(ctx *ActionContext) error {
	payload, ok := ctx.Event.Payload.(events.MouseClickPayload)
	if !ok {
		return nil
	}

	// Tab bar is at row 0. Determine which tab was clicked.
	if payload.ScreenY != 0 {
		return nil
	}

	bufs := ctx.App.Buffers()
	if len(bufs) == 0 {
		return nil
	}

	// Calculate tab positions matching renderer output.
	// Renderer draws: " " + name + dirty(" +") + " x " + "|"
	// Example: " main.go x |" = 1 + 7 + 3 + 1 = 12 chars
	x := 0
	for i, buf := range bufs {
		name := buf.FileName()
		label := " " + name
		if buf.Dirty {
			label += " +"
		}
		label += " x "
		tabWidth := len(label) + 1 // +1 for the "|" separator

		if payload.ScreenX >= x && payload.ScreenX < x+tabWidth {
			// Check if click is on the close button ("x" area).
			closeStart := x + len(label) - 2 // Position of "x" in " x "
			if payload.ScreenX >= closeStart && payload.ScreenX < closeStart+1 {
				// Close button clicked.
				if len(bufs) > 1 {
					ctx.App.CloseBuffer(i)
				}
				return nil
			}
			ctx.App.SwitchBuffer(i)
			ctx.App.SetFocusArea(FocusEditor)
			return nil
		}
		x += tabWidth
	}

	return nil
}
