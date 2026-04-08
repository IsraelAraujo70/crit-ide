package actions

import (
	"github.com/israelcorrea/crit-ide/internal/editor"
)

// paletteMaxVisible is the maximum number of visible commands in the palette popup.
const paletteMaxVisible = 15

// DefaultPaletteEntries returns the full list of command palette entries
// corresponding to all registered actions and their keybindings.
func DefaultPaletteEntries() []editor.PaletteEntry {
	return []editor.PaletteEntry{
		// File
		{ID: "file.save", Label: "Save File", Keybinding: "Ctrl+S", Category: "File"},
		{ID: "finder.open", Label: "Open File (Fuzzy Finder)", Keybinding: "Ctrl+P", Category: "File"},
		{ID: "tree.toggle", Label: "Toggle File Tree", Keybinding: "Ctrl+B", Category: "File"},
		{ID: "tree.refresh", Label: "Refresh File Tree", Keybinding: "", Category: "File"},
		{ID: "app.quit", Label: "Quit Application", Keybinding: "Ctrl+Q", Category: "File"},

		// Edit
		{ID: "edit.undo", Label: "Undo", Keybinding: "Ctrl+Z", Category: "Edit"},
		{ID: "edit.redo", Label: "Redo", Keybinding: "Ctrl+Y", Category: "Edit"},
		{ID: "clipboard.copy", Label: "Copy", Keybinding: "Ctrl+C", Category: "Edit"},
		{ID: "clipboard.cut", Label: "Cut", Keybinding: "Ctrl+X", Category: "Edit"},
		{ID: "clipboard.paste", Label: "Paste", Keybinding: "Ctrl+V", Category: "Edit"},
		{ID: "select.all", Label: "Select All", Keybinding: "Ctrl+A", Category: "Edit"},
		{ID: "edit.duplicate_line", Label: "Duplicate Line", Keybinding: "Ctrl+D", Category: "Edit"},
		{ID: "edit.indent", Label: "Indent", Keybinding: "Tab", Category: "Edit"},
		{ID: "edit.dedent", Label: "Dedent", Keybinding: "Shift+Tab", Category: "Edit"},

		// View
		{ID: "scroll.up", Label: "Page Up", Keybinding: "PageUp", Category: "View"},
		{ID: "scroll.down", Label: "Page Down", Keybinding: "PageDown", Category: "View"},
		{ID: "tab.next", Label: "Next Tab", Keybinding: "Ctrl+PgDn", Category: "View"},
		{ID: "tab.prev", Label: "Previous Tab", Keybinding: "Ctrl+PgUp", Category: "View"},
		{ID: "tab.close", Label: "Close Tab", Keybinding: "Ctrl+W", Category: "View"},
		{ID: "minimap.toggle", Label: "Toggle Minimap", Keybinding: "Ctrl+Shift+M", Category: "View"},

		// Search
		{ID: "search.open", Label: "Find / Replace", Keybinding: "Ctrl+F", Category: "Search"},
		{ID: "search.next", Label: "Find Next", Keybinding: "F3", Category: "Search"},
		{ID: "search.prev", Label: "Find Previous", Keybinding: "Shift+F3", Category: "Search"},

		// LSP
		{ID: "lsp.hover", Label: "Show Hover Info", Keybinding: "Ctrl+K", Category: "LSP"},
		{ID: "lsp.definition", Label: "Go to Definition", Keybinding: "Ctrl+G / F12", Category: "LSP"},
		{ID: "lsp.format", Label: "Format Document", Keybinding: "Ctrl+L", Category: "LSP"},
		{ID: "completion.trigger", Label: "Trigger Autocomplete", Keybinding: "Ctrl+Space", Category: "LSP"},

		// Navigate
		{ID: "cursor.home", Label: "Go to Line Start", Keybinding: "Home", Category: "Navigate"},
		{ID: "cursor.end", Label: "Go to Line End", Keybinding: "End", Category: "Navigate"},
		{ID: "cursor.word_left", Label: "Move Word Left", Keybinding: "Ctrl+Left", Category: "Navigate"},
		{ID: "cursor.word_right", Label: "Move Word Right", Keybinding: "Ctrl+Right", Category: "Navigate"},
	}
}

// --- palette.open: open the command palette ---

type paletteOpen struct{}

func (a *paletteOpen) ID() string { return "palette.open" }

func (a *paletteOpen) Run(ctx *ActionContext) error {
	entries := DefaultPaletteEntries()
	ps := editor.NewPaletteState(entries)
	ctx.App.SetPaletteState(ps)
	ctx.App.SetInputMode(ModeCommandPalette)
	return nil
}

// --- palette.close: close the command palette ---

type paletteClose struct{}

func (a *paletteClose) ID() string { return "palette.close" }

func (a *paletteClose) Run(ctx *ActionContext) error {
	ctx.App.SetPaletteState(nil)
	ctx.App.SetInputMode(ModeNormal)
	return nil
}

// --- palette.char: type a character in the palette input ---

type paletteChar struct{}

func (a *paletteChar) ID() string { return "palette.char" }

func (a *paletteChar) Run(ctx *ActionContext) error {
	ps := ctx.App.PaletteState()
	if ps == nil {
		return nil
	}
	ch, ok := ctx.Event.Payload.(rune)
	if !ok {
		return nil
	}
	ps.InsertChar(ch)
	return nil
}

// --- palette.backspace: delete character backward ---

type paletteBackspace struct{}

func (a *paletteBackspace) ID() string { return "palette.backspace" }

func (a *paletteBackspace) Run(ctx *ActionContext) error {
	ps := ctx.App.PaletteState()
	if ps == nil {
		return nil
	}
	ps.DeleteBackward()
	return nil
}

// --- palette.delete: delete character forward ---

type paletteDeleteFwd struct{}

func (a *paletteDeleteFwd) ID() string { return "palette.delete" }

func (a *paletteDeleteFwd) Run(ctx *ActionContext) error {
	ps := ctx.App.PaletteState()
	if ps == nil {
		return nil
	}
	ps.DeleteForward()
	return nil
}

// --- palette.left / palette.right / palette.home / palette.end ---

type paletteLeft struct{}

func (a *paletteLeft) ID() string { return "palette.left" }
func (a *paletteLeft) Run(ctx *ActionContext) error {
	if ps := ctx.App.PaletteState(); ps != nil {
		ps.MoveLeft()
	}
	return nil
}

type paletteRight struct{}

func (a *paletteRight) ID() string { return "palette.right" }
func (a *paletteRight) Run(ctx *ActionContext) error {
	if ps := ctx.App.PaletteState(); ps != nil {
		ps.MoveRight()
	}
	return nil
}

type paletteHome struct{}

func (a *paletteHome) ID() string { return "palette.home" }
func (a *paletteHome) Run(ctx *ActionContext) error {
	if ps := ctx.App.PaletteState(); ps != nil {
		ps.MoveHome()
	}
	return nil
}

type paletteEnd struct{}

func (a *paletteEnd) ID() string { return "palette.end" }
func (a *paletteEnd) Run(ctx *ActionContext) error {
	if ps := ctx.App.PaletteState(); ps != nil {
		ps.MoveEnd()
	}
	return nil
}

// --- palette.up / palette.down: navigate the command list ---

type paletteUp struct{}

func (a *paletteUp) ID() string { return "palette.up" }
func (a *paletteUp) Run(ctx *ActionContext) error {
	if ps := ctx.App.PaletteState(); ps != nil {
		ps.MoveUp()
	}
	return nil
}

type paletteDown struct{}

func (a *paletteDown) ID() string { return "palette.down" }
func (a *paletteDown) Run(ctx *ActionContext) error {
	if ps := ctx.App.PaletteState(); ps != nil {
		ps.MoveDown(paletteMaxVisible)
	}
	return nil
}

// --- palette.execute: execute the selected command ---

type paletteExecute struct{}

func (a *paletteExecute) ID() string { return "palette.execute" }

func (a *paletteExecute) Run(ctx *ActionContext) error {
	ps := ctx.App.PaletteState()
	if ps == nil {
		return nil
	}

	entry := ps.SelectedEntry()
	if entry == nil {
		return nil
	}

	actionID := entry.ID

	// Close palette first.
	ctx.App.SetPaletteState(nil)
	ctx.App.SetInputMode(ModeNormal)

	// Execute the selected command via PostAction (trampoline).
	ctx.App.PostAction(actionID)
	return nil
}

// RegisterPaletteActions registers all command palette actions.
func RegisterPaletteActions(r *Registry) {
	r.Register(&paletteOpen{})
	r.Register(&paletteClose{})
	r.Register(&paletteChar{})
	r.Register(&paletteBackspace{})
	r.Register(&paletteDeleteFwd{})
	r.Register(&paletteLeft{})
	r.Register(&paletteRight{})
	r.Register(&paletteHome{})
	r.Register(&paletteEnd{})
	r.Register(&paletteUp{})
	r.Register(&paletteDown{})
	r.Register(&paletteExecute{})
}
