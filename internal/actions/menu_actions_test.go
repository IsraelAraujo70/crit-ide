package actions

import (
	"testing"

	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/events"
)

func TestMenuOpen(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello")
	app := &mockApp{buffer: buf, vpHeight: 24, scrWidth: 80, clipboard: &mockClipboard{}}

	ctx := newTestContext(app, "menu.open", events.MouseClickPayload{ScreenX: 10, ScreenY: 5})
	reg.Execute("menu.open", ctx)

	if app.inputMode != ModeContextMenu {
		t.Fatal("expected ModeContextMenu")
	}
	if app.contextMenu == nil {
		t.Fatal("expected context menu to be set")
	}
	if app.contextMenu.ScreenX != 10 || app.contextMenu.ScreenY != 5 {
		t.Fatalf("expected menu at (10,5), got (%d,%d)",
			app.contextMenu.ScreenX, app.contextMenu.ScreenY)
	}
	if len(app.contextMenu.Items) != 5 { // Copy, Cut, Paste, separator, Select All
		t.Fatalf("expected 5 menu items, got %d", len(app.contextMenu.Items))
	}
}

func TestMenuClose(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello")
	app := &mockApp{
		buffer:      buf,
		vpHeight:    24,
		clipboard:   &mockClipboard{},
		inputMode:   ModeContextMenu,
		contextMenu: &editor.MenuState{Items: defaultContextMenuItems()},
	}

	reg.Execute("menu.close", newTestContext(app, "menu.close", nil))

	if app.inputMode != ModeNormal {
		t.Fatal("expected ModeNormal after close")
	}
	if app.contextMenu != nil {
		t.Fatal("expected context menu to be nil after close")
	}
}

func TestMenuUpDown(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	items := defaultContextMenuItems()
	menu := &editor.MenuState{Items: items, SelectedIdx: 0}
	buf := newBufferWithLines("hello")
	app := &mockApp{
		buffer:      buf,
		vpHeight:    24,
		clipboard:   &mockClipboard{},
		inputMode:   ModeContextMenu,
		contextMenu: menu,
	}

	// Down from 0 → should go to 1 (Cut).
	reg.Execute("menu.down", newTestContext(app, "menu.down", nil))
	if menu.SelectedIdx != 1 {
		t.Fatalf("expected idx 1 after down, got %d", menu.SelectedIdx)
	}

	// Down from 1 → should go to 2 (Paste).
	reg.Execute("menu.down", newTestContext(app, "menu.down", nil))
	if menu.SelectedIdx != 2 {
		t.Fatalf("expected idx 2, got %d", menu.SelectedIdx)
	}

	// Down from 2 → should skip separator (3) and go to 4 (Select All).
	reg.Execute("menu.down", newTestContext(app, "menu.down", nil))
	if menu.SelectedIdx != 4 {
		t.Fatalf("expected idx 4 (skip separator), got %d", menu.SelectedIdx)
	}

	// Down from 4 → should wrap to 0 (Copy).
	reg.Execute("menu.down", newTestContext(app, "menu.down", nil))
	if menu.SelectedIdx != 0 {
		t.Fatalf("expected idx 0 (wrap), got %d", menu.SelectedIdx)
	}

	// Up from 0 → should wrap to 4 (Select All), skipping separator.
	reg.Execute("menu.up", newTestContext(app, "menu.up", nil))
	if menu.SelectedIdx != 4 {
		t.Fatalf("expected idx 4 after up-wrap, got %d", menu.SelectedIdx)
	}

	// Up from 4 → should skip separator and go to 2 (Paste).
	reg.Execute("menu.up", newTestContext(app, "menu.up", nil))
	if menu.SelectedIdx != 2 {
		t.Fatalf("expected idx 2, got %d", menu.SelectedIdx)
	}
}

func TestMenuExecute(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	items := defaultContextMenuItems()
	menu := &editor.MenuState{Items: items, SelectedIdx: 0} // Copy selected
	buf := newBufferWithLines("hello")
	buf.SetSelection(editor.Position{Line: 0, Col: 0}, editor.Position{Line: 0, Col: 5})
	app := &mockApp{
		buffer:      buf,
		vpHeight:    24,
		clipboard:   &mockClipboard{},
		inputMode:   ModeContextMenu,
		contextMenu: menu,
	}

	reg.Execute("menu.execute", newTestContext(app, "menu.execute", nil))

	if app.inputMode != ModeNormal {
		t.Fatal("expected ModeNormal after execute")
	}
	if app.contextMenu != nil {
		t.Fatal("expected menu closed after execute")
	}
	if app.pendingAction != "clipboard.copy" {
		t.Fatalf("expected pending action %q, got %q", "clipboard.copy", app.pendingAction)
	}
}

func TestMenuClick_Inside(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	// Menu at (10, 5), items: Copy(0), Cut(1), Paste(2), sep(3), SelectAll(4)
	// maxLabel = "Select All" = 10, popupWidth = 10+4=14, popupHeight = 5+2=7
	// Top border at y=5, items start at y=6
	// Item 0 (Copy) at y=6, Item 1 (Cut) at y=7, etc.
	// Left border at x=10, content starts at x=11
	items := defaultContextMenuItems()
	menu := &editor.MenuState{ScreenX: 10, ScreenY: 5, Items: items, SelectedIdx: 0}
	buf := newBufferWithLines("hello")
	app := &mockApp{
		buffer:      buf,
		vpHeight:    24,
		scrWidth:    80,
		clipboard:   &mockClipboard{},
		inputMode:   ModeContextMenu,
		contextMenu: menu,
	}

	// Click on Paste (item idx 2, y = 5+1+2 = 8), x inside content area.
	ctx := newTestContext(app, "menu.click", events.MouseClickPayload{ScreenX: 15, ScreenY: 8})
	reg.Execute("menu.click", ctx)

	if app.inputMode != ModeNormal {
		t.Fatal("expected ModeNormal after click inside")
	}
	if app.pendingAction != "clipboard.paste" {
		t.Fatalf("expected pending action %q, got %q", "clipboard.paste", app.pendingAction)
	}
}

func TestMenuClick_Outside(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	items := defaultContextMenuItems()
	menu := &editor.MenuState{ScreenX: 10, ScreenY: 5, Items: items, SelectedIdx: 0}
	buf := newBufferWithLines("hello")
	app := &mockApp{
		buffer:      buf,
		vpHeight:    24,
		scrWidth:    80,
		clipboard:   &mockClipboard{},
		inputMode:   ModeContextMenu,
		contextMenu: menu,
	}

	// Click outside the menu bounds.
	ctx := newTestContext(app, "menu.click", events.MouseClickPayload{ScreenX: 0, ScreenY: 0})
	reg.Execute("menu.click", ctx)

	if app.inputMode != ModeNormal {
		t.Fatal("expected ModeNormal after outside click")
	}
	if app.contextMenu != nil {
		t.Fatal("expected menu closed after outside click")
	}
}
