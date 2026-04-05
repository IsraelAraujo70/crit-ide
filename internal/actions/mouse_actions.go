package actions

import (
	"github.com/israelcorrea/crit-ide/internal/events"
)

// --- Mouse click action ---

// mouseClick positions the cursor at the screen location of a mouse click
// and clears any active selection.
type mouseClick struct{}

func (a *mouseClick) ID() string { return "mouse.click" }

func (a *mouseClick) Run(ctx *ActionContext) error {
	payload, ok := ctx.Event.Payload.(events.MouseClickPayload)
	if !ok {
		return nil
	}

	buf := ctx.App.ActiveBuffer()
	scrollY := ctx.App.ScrollY()
	vpHeight := ctx.App.ViewportHeight()

	row, col, valid := screenToBufferPos(buf, scrollY, vpHeight, payload.ScreenX, payload.ScreenY)
	if !valid {
		return nil
	}

	buf.ClearSelection()
	buf.SetCursorPos(row, col)
	return nil
}

// --- Mouse scroll action ---

// mouseScroll scrolls the viewport by the given number of lines.
type mouseScroll struct{}

func (a *mouseScroll) ID() string { return "mouse.scroll" }

func (a *mouseScroll) Run(ctx *ActionContext) error {
	payload, ok := ctx.Event.Payload.(events.MouseScrollPayload)
	if !ok {
		return nil
	}

	buf := ctx.App.ActiveBuffer()
	vpHeight := ctx.App.ViewportHeight()

	newScrollY := ctx.App.ScrollY() + payload.Direction

	// Clamp scroll to valid range.
	if newScrollY < 0 {
		newScrollY = 0
	}
	maxScroll := buf.Text.LineCount() - vpHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if newScrollY > maxScroll {
		newScrollY = maxScroll
	}

	ctx.App.SetScrollY(newScrollY)

	// Cursor visibility is handled by ensureCursorVisible() in the main loop
	// after every action, so no adjustment is needed here.

	return nil
}
