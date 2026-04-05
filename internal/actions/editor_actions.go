package actions

import (
	"github.com/israelcorrea/crit-ide/internal/editor"
)

// --- Cursor movement actions ---

type cursorUp struct{}

func (a *cursorUp) ID() string { return "cursor.up" }
func (a *cursorUp) Run(ctx *ActionContext) error {
	ctx.App.ActiveBuffer().MoveCursor(editor.DirUp)
	return nil
}

type cursorDown struct{}

func (a *cursorDown) ID() string { return "cursor.down" }
func (a *cursorDown) Run(ctx *ActionContext) error {
	ctx.App.ActiveBuffer().MoveCursor(editor.DirDown)
	return nil
}

type cursorLeft struct{}

func (a *cursorLeft) ID() string { return "cursor.left" }
func (a *cursorLeft) Run(ctx *ActionContext) error {
	ctx.App.ActiveBuffer().MoveCursor(editor.DirLeft)
	return nil
}

type cursorRight struct{}

func (a *cursorRight) ID() string { return "cursor.right" }
func (a *cursorRight) Run(ctx *ActionContext) error {
	ctx.App.ActiveBuffer().MoveCursor(editor.DirRight)
	return nil
}

type cursorHome struct{}

func (a *cursorHome) ID() string { return "cursor.home" }
func (a *cursorHome) Run(ctx *ActionContext) error {
	ctx.App.ActiveBuffer().CursorHome()
	return nil
}

type cursorEnd struct{}

func (a *cursorEnd) ID() string { return "cursor.end" }
func (a *cursorEnd) Run(ctx *ActionContext) error {
	ctx.App.ActiveBuffer().CursorEnd()
	return nil
}

// --- Text editing actions ---

type insertChar struct{}

func (a *insertChar) ID() string { return "insert.char" }
func (a *insertChar) Run(ctx *ActionContext) error {
	ch, ok := ctx.Event.Payload.(rune)
	if !ok {
		return nil
	}
	ctx.App.ActiveBuffer().InsertChar(ch)
	return nil
}

type insertNewline struct{}

func (a *insertNewline) ID() string { return "insert.newline" }
func (a *insertNewline) Run(ctx *ActionContext) error {
	ctx.App.ActiveBuffer().InsertNewline()
	return nil
}

type deleteBackward struct{}

func (a *deleteBackward) ID() string { return "delete.backward" }
func (a *deleteBackward) Run(ctx *ActionContext) error {
	ctx.App.ActiveBuffer().DeleteBackward()
	return nil
}

type deleteForward struct{}

func (a *deleteForward) ID() string { return "delete.forward" }
func (a *deleteForward) Run(ctx *ActionContext) error {
	ctx.App.ActiveBuffer().DeleteForward()
	return nil
}

// --- File actions ---

type fileSave struct{}

func (a *fileSave) ID() string { return "file.save" }
func (a *fileSave) Run(ctx *ActionContext) error {
	return ctx.App.ActiveBuffer().SaveFile()
}

// --- Application actions ---

type appQuit struct{}

func (a *appQuit) ID() string { return "app.quit" }
func (a *appQuit) Run(ctx *ActionContext) error {
	ctx.App.Quit()
	return nil
}

// --- Scroll actions ---

type scrollUp struct{}

func (a *scrollUp) ID() string { return "scroll.up" }
func (a *scrollUp) Run(ctx *ActionContext) error {
	h := ctx.App.ViewportHeight()
	if h < 1 {
		h = 1
	}
	newY := ctx.App.ScrollY() - h
	if newY < 0 {
		newY = 0
	}
	ctx.App.SetScrollY(newY)
	// Also move cursor to keep it visible.
	buf := ctx.App.ActiveBuffer()
	if buf.CursorRow >= newY+h {
		buf.CursorRow = newY + h - 1
	}
	if buf.CursorRow < newY {
		buf.CursorRow = newY
	}
	buf.ClampCursor()
	return nil
}

type scrollDown struct{}

func (a *scrollDown) ID() string { return "scroll.down" }
func (a *scrollDown) Run(ctx *ActionContext) error {
	h := ctx.App.ViewportHeight()
	if h < 1 {
		h = 1
	}
	buf := ctx.App.ActiveBuffer()
	maxScroll := buf.Text.LineCount() - h
	if maxScroll < 0 {
		maxScroll = 0
	}
	newY := ctx.App.ScrollY() + h
	if newY > maxScroll {
		newY = maxScroll
	}
	if newY < 0 {
		newY = 0
	}
	ctx.App.SetScrollY(newY)
	// Also move cursor to keep it visible.
	if buf.CursorRow < newY {
		buf.CursorRow = newY
	}
	if buf.CursorRow >= newY+h {
		buf.CursorRow = newY + h - 1
	}
	buf.ClampCursor()
	return nil
}

// RegisterAll registers all actions in the given registry.
func RegisterAll(r *Registry) {
	// Cursor movement.
	r.Register(&cursorUp{})
	r.Register(&cursorDown{})
	r.Register(&cursorLeft{})
	r.Register(&cursorRight{})
	r.Register(&cursorHome{})
	r.Register(&cursorEnd{})

	// Text editing.
	r.Register(&insertChar{})
	r.Register(&insertNewline{})
	r.Register(&deleteBackward{})
	r.Register(&deleteForward{})

	// File operations.
	r.Register(&fileSave{})

	// Application.
	r.Register(&appQuit{})

	// Keyboard scroll.
	r.Register(&scrollUp{})
	r.Register(&scrollDown{})

	// Mouse.
	r.Register(&mouseClick{})
	r.Register(&mouseScroll{})
	r.Register(&mouseDrag{})

	// Selection.
	r.Register(&selectAll{})
	r.Register(&inputEscape{})

	// Clipboard.
	r.Register(&clipboardCopy{})
	r.Register(&clipboardCut{})
	r.Register(&clipboardPaste{})

	// Context menu.
	r.Register(&menuOpen{})
	r.Register(&menuClose{})
	r.Register(&menuUp{})
	r.Register(&menuDown{})
	r.Register(&menuExecute{})
	r.Register(&menuClick{})
}
