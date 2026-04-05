package actions

import (
	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/events"
)

// screenToBufferPos converts screen coordinates to a buffer position.
// Returns the buffer row and byte-offset column, and whether the position
// is valid (i.e., not on gutter or statusline).
func screenToBufferPos(buf *editor.Buffer, scrollY, vpHeight, screenX, screenY int) (row, col int, ok bool) {
	if screenY >= vpHeight {
		return 0, 0, false // Statusline.
	}
	gutterWidth := editor.GutterWidth(buf.Text.LineCount())
	if screenX < gutterWidth {
		return 0, 0, false // Gutter.
	}
	bufferRow := scrollY + screenY
	maxRow := buf.Text.LineCount() - 1
	if maxRow < 0 {
		maxRow = 0
	}
	if bufferRow > maxRow {
		bufferRow = maxRow
	}
	visualCol := screenX - gutterWidth
	line := buf.Text.Line(bufferRow)
	byteOffset := editor.VisualColToByteOffset(line, visualCol)
	return bufferRow, byteOffset, true
}

// --- Mouse drag action ---

// mouseDrag sets or extends the text selection based on a drag event.
type mouseDrag struct{}

func (a *mouseDrag) ID() string { return "mouse.drag" }

func (a *mouseDrag) Run(ctx *ActionContext) error {
	payload, ok := ctx.Event.Payload.(events.MouseDragPayload)
	if !ok {
		return nil
	}

	buf := ctx.App.ActiveBuffer()
	scrollY := ctx.App.ScrollY()
	vpHeight := ctx.App.ViewportHeight()

	anchorRow, anchorCol, anchorOK := screenToBufferPos(buf, scrollY, vpHeight, payload.AnchorX, payload.AnchorY)
	curRow, curCol, curOK := screenToBufferPos(buf, scrollY, vpHeight, payload.CurrentX, payload.CurrentY)

	if !anchorOK || !curOK {
		return nil
	}

	buf.SetSelection(
		editor.Position{Line: anchorRow, Col: anchorCol},
		editor.Position{Line: curRow, Col: curCol},
	)
	buf.SetCursorPos(curRow, curCol)
	return nil
}

// --- Select all action ---

type selectAll struct{}

func (a *selectAll) ID() string { return "select.all" }

func (a *selectAll) Run(ctx *ActionContext) error {
	ctx.App.ActiveBuffer().SelectAll()
	return nil
}

// --- Input escape action ---

// inputEscape handles the Escape key. In normal mode, it clears the selection.
type inputEscape struct{}

func (a *inputEscape) ID() string { return "input.escape" }

func (a *inputEscape) Run(ctx *ActionContext) error {
	ctx.App.ActiveBuffer().ClearSelection()
	return nil
}
