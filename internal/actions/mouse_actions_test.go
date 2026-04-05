package actions

import (
	"fmt"
	"testing"

	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/events"
)

// mockApp and newTestContext are defined in editor_actions_test.go.

func TestMouseClick_TextArea(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello", "world")
	app := &mockApp{buffer: buf, vpHeight: 24}

	// gutterWidth for 2 lines = 4 (min 3 digits + 1 space).
	// Click at screenX=5, screenY=1 (row 0 is tab bar) → visualCol = 5-4 = 1 → byte offset 1.
	ctx := newTestContext(app, "mouse.click", events.MouseClickPayload{ScreenX: 5, ScreenY: 1})
	reg.Execute("mouse.click", ctx)

	if buf.CursorRow != 0 || buf.CursorCol != 1 {
		t.Fatalf("expected (0,1), got (%d,%d)", buf.CursorRow, buf.CursorCol)
	}
}

func TestMouseClick_SecondLine(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello", "world")
	app := &mockApp{buffer: buf, vpHeight: 24}

	// Click at screenY=2 (tab bar at 0) → buffer row 1.
	ctx := newTestContext(app, "mouse.click", events.MouseClickPayload{ScreenX: 6, ScreenY: 2})
	reg.Execute("mouse.click", ctx)

	if buf.CursorRow != 1 || buf.CursorCol != 2 {
		t.Fatalf("expected (1,2), got (%d,%d)", buf.CursorRow, buf.CursorCol)
	}
}

func TestMouseClick_WithScrollOffset(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("line0", "line1", "line2", "line3", "line4", "line5", "line6", "line7", "line8", "line9")
	app := &mockApp{buffer: buf, vpHeight: 24, scrollY: 5}

	// scrollY=5, screenY=3 (tab bar at 0) → editorScreenY=2 → bufferRow = 5+2 = 7.
	ctx := newTestContext(app, "mouse.click", events.MouseClickPayload{ScreenX: 4, ScreenY: 3})
	reg.Execute("mouse.click", ctx)

	if buf.CursorRow != 7 {
		t.Fatalf("expected row 7, got %d", buf.CursorRow)
	}
}

func TestMouseClick_Gutter(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello", "world")
	buf.SetCursorPos(0, 3)
	app := &mockApp{buffer: buf, vpHeight: 24}

	// Click on gutter (screenX=2 < gutterWidth=4), screenY=1 (after tab bar).
	ctx := newTestContext(app, "mouse.click", events.MouseClickPayload{ScreenX: 2, ScreenY: 1})
	reg.Execute("mouse.click", ctx)

	// Cursor should not move.
	if buf.CursorRow != 0 || buf.CursorCol != 3 {
		t.Fatalf("gutter click should not move cursor, got (%d,%d)", buf.CursorRow, buf.CursorCol)
	}
}

func TestMouseClick_Statusline(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello")
	buf.SetCursorPos(0, 3)
	app := &mockApp{buffer: buf, vpHeight: 24}

	// Click on statusline (screenY=25 = tabBarHeight + vpHeight).
	ctx := newTestContext(app, "mouse.click", events.MouseClickPayload{ScreenX: 5, ScreenY: 25})
	reg.Execute("mouse.click", ctx)

	// Cursor should not move.
	if buf.CursorRow != 0 || buf.CursorCol != 3 {
		t.Fatalf("statusline click should not move cursor, got (%d,%d)", buf.CursorRow, buf.CursorCol)
	}
}

func TestMouseClick_TabBar(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello")
	buf.SetCursorPos(0, 3)
	app := &mockApp{buffer: buf, vpHeight: 24}

	// Click on tab bar (screenY=0).
	ctx := newTestContext(app, "mouse.click", events.MouseClickPayload{ScreenX: 5, ScreenY: 0})
	reg.Execute("mouse.click", ctx)

	// Cursor should not move (tab bar is not editor area).
	if buf.CursorRow != 0 || buf.CursorCol != 3 {
		t.Fatalf("tab bar click should not move cursor, got (%d,%d)", buf.CursorRow, buf.CursorCol)
	}
}

func TestMouseClick_BeyondEOF(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello", "world")
	app := &mockApp{buffer: buf, vpHeight: 24}

	// Click at screenY=11 (tab bar offset), but buffer only has 2 lines. Should clamp to last line.
	ctx := newTestContext(app, "mouse.click", events.MouseClickPayload{ScreenX: 4, ScreenY: 11})
	reg.Execute("mouse.click", ctx)

	if buf.CursorRow != 1 {
		t.Fatalf("expected row clamped to 1, got %d", buf.CursorRow)
	}
}

func TestMouseClick_BeyondLineEnd(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hi")
	app := &mockApp{buffer: buf, vpHeight: 24}

	// Click at screenX=gutterWidth+50, screenY=1 (after tab bar), but line "hi" is only 2 chars.
	ctx := newTestContext(app, "mouse.click", events.MouseClickPayload{ScreenX: 54, ScreenY: 1})
	reg.Execute("mouse.click", ctx)

	if buf.CursorCol != 2 {
		t.Fatalf("expected col clamped to 2, got %d", buf.CursorCol)
	}
}

func TestMouseClick_TabLine(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("\thello")
	app := &mockApp{buffer: buf, vpHeight: 24}

	// gutterWidth=4. Tab expands to 4 spaces (visual cols 0-3).
	// Click at screenX=4+3=7, screenY=1 (after tab bar) → visualCol=3, in the middle of tab expansion.
	// Should snap to byte offset 0 (the tab char).
	ctx := newTestContext(app, "mouse.click", events.MouseClickPayload{ScreenX: 7, ScreenY: 1})
	reg.Execute("mouse.click", ctx)

	if buf.CursorCol != 0 {
		t.Fatalf("expected col 0 (tab snap), got %d", buf.CursorCol)
	}

	// Click at screenX=4+4=8, screenY=1 → visualCol=4, which is after the tab → byte offset 1 ('h').
	ctx2 := newTestContext(app, "mouse.click", events.MouseClickPayload{ScreenX: 8, ScreenY: 1})
	reg.Execute("mouse.click", ctx2)

	if buf.CursorCol != 1 {
		t.Fatalf("expected col 1 (after tab), got %d", buf.CursorCol)
	}
}

// --- Mouse scroll tests ---

func TestMouseScroll_Down(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines(makeLines(50)...)
	app := &mockApp{buffer: buf, vpHeight: 10, scrollY: 0}

	ctx := newTestContext(app, "mouse.scroll", events.MouseScrollPayload{Direction: 3})
	reg.Execute("mouse.scroll", ctx)

	if app.scrollY != 3 {
		t.Fatalf("expected scrollY 3, got %d", app.scrollY)
	}
}

func TestMouseScroll_Up(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines(makeLines(50)...)
	app := &mockApp{buffer: buf, vpHeight: 10, scrollY: 5}

	ctx := newTestContext(app, "mouse.scroll", events.MouseScrollPayload{Direction: -3})
	reg.Execute("mouse.scroll", ctx)

	if app.scrollY != 2 {
		t.Fatalf("expected scrollY 2, got %d", app.scrollY)
	}
}

func TestMouseScroll_ClampTop(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines(makeLines(50)...)
	app := &mockApp{buffer: buf, vpHeight: 10, scrollY: 1}

	ctx := newTestContext(app, "mouse.scroll", events.MouseScrollPayload{Direction: -3})
	reg.Execute("mouse.scroll", ctx)

	if app.scrollY != 0 {
		t.Fatalf("expected scrollY clamped to 0, got %d", app.scrollY)
	}
}

func TestMouseScroll_ClampBottom(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	// 10 lines, vpHeight=8 → maxScroll = 10 - 8 = 2.
	buf := newBufferWithLines(makeLines(10)...)
	app := &mockApp{buffer: buf, vpHeight: 8, scrollY: 0}

	ctx := newTestContext(app, "mouse.scroll", events.MouseScrollPayload{Direction: 10})
	reg.Execute("mouse.scroll", ctx)

	if app.scrollY != 2 {
		t.Fatalf("expected scrollY clamped to 2, got %d", app.scrollY)
	}
}

func TestMouseScroll_DoesNotMoveCursor(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines(makeLines(50)...)
	buf.SetCursorPos(0, 0)
	app := &mockApp{buffer: buf, vpHeight: 10, scrollY: 0}

	// Scroll down 5 lines. The action only changes scrollY; cursor adjustment
	// is handled by ensureCursorVisible() in the main loop.
	ctx := newTestContext(app, "mouse.scroll", events.MouseScrollPayload{Direction: 5})
	reg.Execute("mouse.scroll", ctx)

	if app.scrollY != 5 {
		t.Fatalf("expected scrollY 5, got %d", app.scrollY)
	}
	// Cursor stays at row 0 — the action does not move it.
	if buf.CursorRow != 0 {
		t.Fatalf("expected cursor row 0 (unchanged), got %d", buf.CursorRow)
	}
}

// --- Test helpers ---

// newBufferWithLines creates a buffer with the given lines pre-populated.
func newBufferWithLines(lines ...string) *editor.Buffer {
	buf := editor.NewBuffer("test")
	for i, line := range lines {
		for _, ch := range line {
			buf.InsertChar(ch)
		}
		if i < len(lines)-1 {
			buf.InsertNewline()
		}
	}
	// Reset cursor to origin.
	buf.SetCursorPos(0, 0)
	return buf
}

// makeLines generates n lines of content ("line0", "line1", ..., "line49").
func makeLines(n int) []string {
	lines := make([]string, n)
	for i := 0; i < n; i++ {
		lines[i] = fmt.Sprintf("line%d", i)
	}
	return lines
}
