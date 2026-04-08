package actions

import (
	"github.com/israelcorrea/crit-ide/internal/editor"
)

// RegisterCompletionActions registers all completion-related actions.
func RegisterCompletionActions(r *Registry) {
	r.Register(&completionTrigger{})
	r.Register(&completionAccept{})
	r.Register(&completionDismiss{})
	r.Register(&completionUp{})
	r.Register(&completionDown{})
}

// --- completion.trigger ---

type completionTrigger struct{}

func (a *completionTrigger) ID() string { return "completion.trigger" }

func (a *completionTrigger) Run(ctx *ActionContext) error {
	ctx.App.TriggerCompletion("")
	return nil
}

// --- completion.accept ---

type completionAccept struct{}

func (a *completionAccept) ID() string { return "completion.accept" }

func (a *completionAccept) Run(ctx *ActionContext) error {
	cs := ctx.App.CompletionState()
	if cs == nil {
		return nil
	}
	item := cs.SelectedItem()
	if item == nil {
		ctx.App.SetCompletionState(nil)
		ctx.App.SetInputMode(ModeNormal)
		return nil
	}

	buf := ctx.App.ActiveBuffer()
	insertText := item.InsertString()

	// Delete typed prefix (from anchor to current cursor).
	if buf.CursorRow == cs.AnchorRow && buf.CursorCol > cs.AnchorCol {
		delStart := editor.Position{Line: cs.AnchorRow, Col: cs.AnchorCol}
		delEnd := editor.Position{Line: buf.CursorRow, Col: buf.CursorCol}
		delRange := editor.Range{Start: delStart, End: delEnd}
		deleted := buf.Text.Slice(delRange)

		buf.Undo.Push(editor.UndoEntry{
			Kind:      editor.EditDelete,
			Pos:       delStart,
			Text:      deleted,
			CursorRow: buf.CursorRow,
			CursorCol: buf.CursorCol,
		})
		_ = buf.Text.Delete(delRange)
		buf.CursorRow = cs.AnchorRow
		buf.CursorCol = cs.AnchorCol
	}

	// Insert the completion text.
	insertPos := editor.Position{Line: buf.CursorRow, Col: buf.CursorCol}
	buf.Undo.Push(editor.UndoEntry{
		Kind:      editor.EditInsert,
		Pos:       insertPos,
		Text:      insertText,
		CursorRow: buf.CursorRow,
		CursorCol: buf.CursorCol,
	})
	_ = buf.Text.Insert(insertPos, insertText)

	// Advance cursor past the inserted text.
	for _, ch := range insertText {
		if ch == '\n' {
			buf.CursorRow++
			buf.CursorCol = 0
		} else {
			buf.CursorCol += len(string(ch))
		}
	}
	buf.Dirty = true

	// Close completion.
	ctx.App.SetCompletionState(nil)
	ctx.App.SetInputMode(ModeNormal)
	return nil
}

// --- completion.dismiss ---

type completionDismiss struct{}

func (a *completionDismiss) ID() string { return "completion.dismiss" }

func (a *completionDismiss) Run(ctx *ActionContext) error {
	ctx.App.SetCompletionState(nil)
	ctx.App.SetInputMode(ModeNormal)
	return nil
}

// --- completion.up ---

type completionUp struct{}

func (a *completionUp) ID() string { return "completion.up" }

func (a *completionUp) Run(ctx *ActionContext) error {
	cs := ctx.App.CompletionState()
	if cs == nil {
		return nil
	}
	cs.MoveUp()
	return nil
}

// --- completion.down ---

type completionDown struct{}

func (a *completionDown) ID() string { return "completion.down" }

func (a *completionDown) Run(ctx *ActionContext) error {
	cs := ctx.App.CompletionState()
	if cs == nil {
		return nil
	}
	cs.MoveDown()
	return nil
}
