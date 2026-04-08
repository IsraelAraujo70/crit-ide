package actions

import (
	"fmt"

	"github.com/israelcorrea/crit-ide/internal/editor"
)

// --- search.open: open the search bar (Ctrl+F) ---

type searchOpen struct{}

func (a *searchOpen) ID() string { return "search.open" }

func (a *searchOpen) Run(ctx *ActionContext) error {
	ss := ctx.App.SearchState()
	if ss == nil {
		ss = editor.NewSearchState()
	}
	// If there's a selection, use it as the initial query.
	buf := ctx.App.ActiveBuffer()
	if buf.HasSelection() {
		sel := buf.SelectedText()
		ss.Query = sel
		ss.QueryCursor = len(sel)
		buf.ClearSelection()
	}
	ss.ActiveField = editor.FieldFind
	ctx.App.SetSearchState(ss)
	ctx.App.SetInputMode(ModeSearch)
	// Perform initial search.
	ss.FindAll(buf.Text)
	if len(ss.Matches) > 0 {
		ss.FindNearest(buf.CursorRow, buf.CursorCol)
	}
	return nil
}

// --- search.close: close the search bar (Escape) ---

type searchClose struct{}

func (a *searchClose) ID() string { return "search.close" }

func (a *searchClose) Run(ctx *ActionContext) error {
	ctx.App.SetSearchState(nil)
	ctx.App.SetInputMode(ModeNormal)
	return nil
}

// --- search.char: insert character into the active search field ---

type searchChar struct{}

func (a *searchChar) ID() string { return "search.char" }

func (a *searchChar) Run(ctx *ActionContext) error {
	ss := ctx.App.SearchState()
	if ss == nil {
		return nil
	}
	ch, ok := ctx.Event.Payload.(rune)
	if !ok {
		return nil
	}
	ss.InsertChar(ch)
	// Re-run search when query changes.
	if ss.ActiveField == editor.FieldFind {
		buf := ctx.App.ActiveBuffer()
		ss.FindAll(buf.Text)
		if len(ss.Matches) > 0 {
			ss.FindNearest(buf.CursorRow, buf.CursorCol)
		}
	}
	return nil
}

// --- search.backspace: delete character backward in active field ---

type searchBackspace struct{}

func (a *searchBackspace) ID() string { return "search.backspace" }

func (a *searchBackspace) Run(ctx *ActionContext) error {
	ss := ctx.App.SearchState()
	if ss == nil {
		return nil
	}
	ss.DeleteBackward()
	if ss.ActiveField == editor.FieldFind {
		buf := ctx.App.ActiveBuffer()
		ss.FindAll(buf.Text)
		if len(ss.Matches) > 0 {
			ss.FindNearest(buf.CursorRow, buf.CursorCol)
		}
	}
	return nil
}

// --- search.delete: delete character forward in active field ---

type searchDelete struct{}

func (a *searchDelete) ID() string { return "search.delete" }

func (a *searchDelete) Run(ctx *ActionContext) error {
	ss := ctx.App.SearchState()
	if ss == nil {
		return nil
	}
	ss.DeleteForward()
	if ss.ActiveField == editor.FieldFind {
		buf := ctx.App.ActiveBuffer()
		ss.FindAll(buf.Text)
		if len(ss.Matches) > 0 {
			ss.FindNearest(buf.CursorRow, buf.CursorCol)
		}
	}
	return nil
}

// --- search.left / search.right / search.home / search.end ---

type searchLeft struct{}

func (a *searchLeft) ID() string { return "search.left" }
func (a *searchLeft) Run(ctx *ActionContext) error {
	if ss := ctx.App.SearchState(); ss != nil {
		ss.MoveLeft()
	}
	return nil
}

type searchRight struct{}

func (a *searchRight) ID() string { return "search.right" }
func (a *searchRight) Run(ctx *ActionContext) error {
	if ss := ctx.App.SearchState(); ss != nil {
		ss.MoveRight()
	}
	return nil
}

type searchHome struct{}

func (a *searchHome) ID() string { return "search.home" }
func (a *searchHome) Run(ctx *ActionContext) error {
	if ss := ctx.App.SearchState(); ss != nil {
		ss.MoveHome()
	}
	return nil
}

type searchEnd struct{}

func (a *searchEnd) ID() string { return "search.end" }
func (a *searchEnd) Run(ctx *ActionContext) error {
	if ss := ctx.App.SearchState(); ss != nil {
		ss.MoveEnd()
	}
	return nil
}

// --- search.next: find next match (Enter / F3) ---

type searchNext struct{}

func (a *searchNext) ID() string { return "search.next" }

func (a *searchNext) Run(ctx *ActionContext) error {
	ss := ctx.App.SearchState()
	if ss == nil {
		return nil
	}
	buf := ctx.App.ActiveBuffer()
	pos, found := ss.FindNext(buf.CursorRow, buf.CursorCol)
	if found {
		buf.SetCursorPos(pos.Line, pos.Col)
		ctx.App.SetStatusMessage(fmt.Sprintf("%d/%d matches", ss.CurrentMatchNumber(), ss.MatchCount()))
	} else {
		ctx.App.SetStatusMessage("No matches")
	}
	return nil
}

// --- search.prev: find previous match (Shift+F3) ---

type searchPrev struct{}

func (a *searchPrev) ID() string { return "search.prev" }

func (a *searchPrev) Run(ctx *ActionContext) error {
	ss := ctx.App.SearchState()
	if ss == nil {
		return nil
	}
	buf := ctx.App.ActiveBuffer()
	pos, found := ss.FindPrev(buf.CursorRow, buf.CursorCol)
	if found {
		buf.SetCursorPos(pos.Line, pos.Col)
		ctx.App.SetStatusMessage(fmt.Sprintf("%d/%d matches", ss.CurrentMatchNumber(), ss.MatchCount()))
	} else {
		ctx.App.SetStatusMessage("No matches")
	}
	return nil
}

// --- search.toggle_replace: toggle the replace field (Tab in search mode) ---

type searchToggleReplace struct{}

func (a *searchToggleReplace) ID() string { return "search.toggle_replace" }

func (a *searchToggleReplace) Run(ctx *ActionContext) error {
	if ss := ctx.App.SearchState(); ss != nil {
		ss.ToggleField()
	}
	return nil
}

// --- search.replace: replace the current match ---

type searchReplace struct{}

func (a *searchReplace) ID() string { return "search.replace" }

func (a *searchReplace) Run(ctx *ActionContext) error {
	ss := ctx.App.SearchState()
	if ss == nil {
		return nil
	}
	m, ok := ss.CurrentMatch()
	if !ok {
		return nil
	}

	buf := ctx.App.ActiveBuffer()

	// Record undo for the replacement.
	deleted := buf.Text.Slice(m)
	buf.Undo.Push(editor.UndoEntry{
		Kind:      editor.EditDelete,
		Pos:       m.Start,
		Text:      deleted,
		CursorRow: buf.CursorRow,
		CursorCol: buf.CursorCol,
	})
	_ = buf.Text.Delete(m)

	if ss.ReplaceText != "" {
		buf.Undo.Push(editor.UndoEntry{
			Kind:      editor.EditInsert,
			Pos:       m.Start,
			Text:      ss.ReplaceText,
			CursorRow: m.Start.Line,
			CursorCol: m.Start.Col,
		})
		_ = buf.Text.Insert(m.Start, ss.ReplaceText)
	}

	buf.Dirty = true
	buf.SetCursorPos(m.Start.Line, m.Start.Col+len(ss.ReplaceText))

	// Re-run search to update matches.
	ss.FindAll(buf.Text)
	if len(ss.Matches) > 0 {
		ss.FindNearest(buf.CursorRow, buf.CursorCol)
	}

	ctx.App.SetStatusMessage(fmt.Sprintf("Replaced — %d matches remaining", ss.MatchCount()))
	return nil
}

// --- search.replace_all: replace all matches ---

type searchReplaceAll struct{}

func (a *searchReplaceAll) ID() string { return "search.replace_all" }

func (a *searchReplaceAll) Run(ctx *ActionContext) error {
	ss := ctx.App.SearchState()
	if ss == nil || len(ss.Matches) == 0 {
		return nil
	}

	buf := ctx.App.ActiveBuffer()
	count := len(ss.Matches)

	// Replace in reverse order to preserve positions.
	for i := len(ss.Matches) - 1; i >= 0; i-- {
		m := ss.Matches[i]
		deleted := buf.Text.Slice(m)
		buf.Undo.Push(editor.UndoEntry{
			Kind:      editor.EditDelete,
			Pos:       m.Start,
			Text:      deleted,
			CursorRow: buf.CursorRow,
			CursorCol: buf.CursorCol,
		})
		_ = buf.Text.Delete(m)

		if ss.ReplaceText != "" {
			buf.Undo.Push(editor.UndoEntry{
				Kind:      editor.EditInsert,
				Pos:       m.Start,
				Text:      ss.ReplaceText,
				CursorRow: m.Start.Line,
				CursorCol: m.Start.Col,
			})
			_ = buf.Text.Insert(m.Start, ss.ReplaceText)
		}
	}

	buf.Dirty = true

	// Re-run search (should find 0 matches now).
	ss.FindAll(buf.Text)

	ctx.App.SetStatusMessage(fmt.Sprintf("Replaced %d occurrences", count))
	return nil
}
