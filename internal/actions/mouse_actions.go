package actions

import (
	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/events"
)

// --- Mouse click action ---

// mouseClick positions the cursor at the screen location of a mouse click.
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

	screenX := payload.ScreenX
	screenY := payload.ScreenY

	// Ignore clicks on the statusline (last row).
	if screenY >= vpHeight {
		return nil
	}

	gutterWidth := editor.GutterWidth(buf.Text.LineCount())

	// Ignore clicks on the gutter.
	if screenX < gutterWidth {
		return nil
	}

	// Convert screen Y to buffer row.
	bufferRow := scrollY + screenY

	// Clamp to last line if clicking below EOF.
	maxRow := buf.Text.LineCount() - 1
	if maxRow < 0 {
		maxRow = 0
	}
	if bufferRow > maxRow {
		bufferRow = maxRow
	}

	// Convert screen X to visual column, then to byte offset.
	visualCol := screenX - gutterWidth
	line := buf.Text.Line(bufferRow)
	byteOffset := editor.VisualColToByteOffset(line, visualCol)

	buf.SetCursorPos(bufferRow, byteOffset)
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
