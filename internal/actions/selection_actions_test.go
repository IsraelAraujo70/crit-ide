package actions

import (
	"testing"

	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/events"
)

func TestMouseDrag_BasicSelection(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello", "world")
	app := &mockApp{buffer: buf, vpHeight: 24, clipboard: &mockClipboard{}}

	// gutterWidth=4 for 2 lines.
	// Drag from (4,0) to (8,0) → visualCol 0 to visualCol 4 → byte 0 to byte 4.
	ctx := newTestContext(app, "mouse.drag", events.MouseDragPayload{
		AnchorX: 4, AnchorY: 0, CurrentX: 8, CurrentY: 0,
	})
	reg.Execute("mouse.drag", ctx)

	if !buf.HasSelection() {
		t.Fatal("expected selection to be set")
	}
	r := buf.Selection.Normalized()
	if r.Start.Col != 0 || r.End.Col != 4 {
		t.Fatalf("expected cols (0,4), got (%d,%d)", r.Start.Col, r.End.Col)
	}
	// Cursor should be at the drag end.
	if buf.CursorRow != 0 || buf.CursorCol != 4 {
		t.Fatalf("expected cursor at (0,4), got (%d,%d)", buf.CursorRow, buf.CursorCol)
	}
}

func TestMouseDrag_MultiLine(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello", "world")
	app := &mockApp{buffer: buf, vpHeight: 24, clipboard: &mockClipboard{}}

	// Drag from line 0 col 2 to line 1 col 3.
	ctx := newTestContext(app, "mouse.drag", events.MouseDragPayload{
		AnchorX: 6, AnchorY: 0, CurrentX: 7, CurrentY: 1,
	})
	reg.Execute("mouse.drag", ctx)

	if !buf.HasSelection() {
		t.Fatal("expected selection")
	}
	got := buf.SelectedText()
	if got != "llo\nwor" {
		t.Fatalf("expected %q, got %q", "llo\nwor", got)
	}
}

func TestMouseDrag_ReverseSelection(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello")
	app := &mockApp{buffer: buf, vpHeight: 24, clipboard: &mockClipboard{}}

	// Drag right-to-left.
	ctx := newTestContext(app, "mouse.drag", events.MouseDragPayload{
		AnchorX: 8, AnchorY: 0, CurrentX: 5, CurrentY: 0,
	})
	reg.Execute("mouse.drag", ctx)

	got := buf.SelectedText()
	if got != "ell" {
		t.Fatalf("expected %q, got %q", "ell", got)
	}
}

func TestMouseClick_ClearsSelection(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello", "world")
	buf.SetSelection(editor.Position{Line: 0, Col: 0}, editor.Position{Line: 0, Col: 5})
	app := &mockApp{buffer: buf, vpHeight: 24, clipboard: &mockClipboard{}}

	ctx := newTestContext(app, "mouse.click", events.MouseClickPayload{ScreenX: 5, ScreenY: 0})
	reg.Execute("mouse.click", ctx)

	if buf.HasSelection() {
		t.Fatal("expected selection to be cleared after click")
	}
}

func TestSelectAll(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello", "world")
	app := &mockApp{buffer: buf, vpHeight: 24, clipboard: &mockClipboard{}}

	reg.Execute("select.all", newTestContext(app, "select.all", nil))

	if !buf.HasSelection() {
		t.Fatal("expected selection after select.all")
	}
	got := buf.SelectedText()
	if got != "hello\nworld" {
		t.Fatalf("expected %q, got %q", "hello\nworld", got)
	}
}

func TestInputEscape_ClearsSelection(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello")
	buf.SetSelection(editor.Position{Line: 0, Col: 0}, editor.Position{Line: 0, Col: 3})
	app := &mockApp{buffer: buf, vpHeight: 24, clipboard: &mockClipboard{}}

	reg.Execute("input.escape", newTestContext(app, "input.escape", nil))

	if buf.HasSelection() {
		t.Fatal("expected selection cleared after escape")
	}
}
