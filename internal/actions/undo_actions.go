package actions

// --- Undo/Redo actions ---

type undoAction struct{}

func (a *undoAction) ID() string { return "edit.undo" }
func (a *undoAction) Run(ctx *ActionContext) error {
	ctx.App.ActiveBuffer().UndoEdit()
	return nil
}

type redoAction struct{}

func (a *redoAction) ID() string { return "edit.redo" }
func (a *redoAction) Run(ctx *ActionContext) error {
	ctx.App.ActiveBuffer().RedoEdit()
	return nil
}

// --- Word movement actions ---

type cursorWordLeft struct{}

func (a *cursorWordLeft) ID() string { return "cursor.word_left" }
func (a *cursorWordLeft) Run(ctx *ActionContext) error {
	ctx.App.ActiveBuffer().WordLeft()
	return nil
}

type cursorWordRight struct{}

func (a *cursorWordRight) ID() string { return "cursor.word_right" }
func (a *cursorWordRight) Run(ctx *ActionContext) error {
	ctx.App.ActiveBuffer().WordRight()
	return nil
}

// --- Line operations ---

type duplicateLine struct{}

func (a *duplicateLine) ID() string { return "edit.duplicate_line" }
func (a *duplicateLine) Run(ctx *ActionContext) error {
	ctx.App.ActiveBuffer().DuplicateLine()
	return nil
}

// --- Double-click word select ---

type selectWord struct{}

func (a *selectWord) ID() string { return "select.word" }
func (a *selectWord) Run(ctx *ActionContext) error {
	ctx.App.ActiveBuffer().SelectWord()
	return nil
}
