package actions

import (
	"fmt"

	"github.com/israelcorrea/crit-ide/internal/editor"
)

// projectSearchMaxVisible is the maximum number of visible entries in the results panel.
const projectSearchMaxVisible = 20

// --- project.search: open the project search panel (Ctrl+Shift+F / F5) ---

type projectSearchOpen struct{}

func (a *projectSearchOpen) ID() string { return "project.search" }

func (a *projectSearchOpen) Run(ctx *ActionContext) error {
	ps := ctx.App.ProjectSearchState()
	if ps == nil {
		ps = editor.NewProjectSearchState()
	}
	// If there's a selection, use it as the initial query.
	buf := ctx.App.ActiveBuffer()
	if buf.HasSelection() {
		sel := buf.SelectedText()
		ps.Query = sel
		ps.CursorPos = len(sel)
		buf.ClearSelection()
	}
	ctx.App.SetProjectSearchState(ps)
	ctx.App.SetInputMode(ModeProjectSearch)
	return nil
}

// --- project.search_close: close the project search panel (Escape) ---

type projectSearchClose struct{}

func (a *projectSearchClose) ID() string { return "project.search_close" }

func (a *projectSearchClose) Run(ctx *ActionContext) error {
	ctx.App.SetProjectSearchState(nil)
	ctx.App.SetInputMode(ModeNormal)
	return nil
}

// --- project.search_char: type a character in the search input ---

type projectSearchChar struct{}

func (a *projectSearchChar) ID() string { return "project.search_char" }

func (a *projectSearchChar) Run(ctx *ActionContext) error {
	ps := ctx.App.ProjectSearchState()
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

// --- project.search_backspace: delete backward in the search input ---

type projectSearchBackspace struct{}

func (a *projectSearchBackspace) ID() string { return "project.search_backspace" }

func (a *projectSearchBackspace) Run(ctx *ActionContext) error {
	ps := ctx.App.ProjectSearchState()
	if ps == nil {
		return nil
	}
	ps.DeleteBackward()
	return nil
}

// --- project.search_delete: delete forward in the search input ---

type projectSearchDelete struct{}

func (a *projectSearchDelete) ID() string { return "project.search_delete" }

func (a *projectSearchDelete) Run(ctx *ActionContext) error {
	ps := ctx.App.ProjectSearchState()
	if ps == nil {
		return nil
	}
	ps.DeleteForward()
	return nil
}

// --- project.search_left / right / home / end ---

type projectSearchLeft struct{}

func (a *projectSearchLeft) ID() string { return "project.search_left" }
func (a *projectSearchLeft) Run(ctx *ActionContext) error {
	if ps := ctx.App.ProjectSearchState(); ps != nil {
		ps.MoveLeft()
	}
	return nil
}

type projectSearchRight struct{}

func (a *projectSearchRight) ID() string { return "project.search_right" }
func (a *projectSearchRight) Run(ctx *ActionContext) error {
	if ps := ctx.App.ProjectSearchState(); ps != nil {
		ps.MoveRight()
	}
	return nil
}

type projectSearchHome struct{}

func (a *projectSearchHome) ID() string { return "project.search_home" }
func (a *projectSearchHome) Run(ctx *ActionContext) error {
	if ps := ctx.App.ProjectSearchState(); ps != nil {
		ps.MoveHome()
	}
	return nil
}

type projectSearchEnd struct{}

func (a *projectSearchEnd) ID() string { return "project.search_end" }
func (a *projectSearchEnd) Run(ctx *ActionContext) error {
	if ps := ctx.App.ProjectSearchState(); ps != nil {
		ps.MoveEnd()
	}
	return nil
}

// --- project.search_execute: run the search (Enter in input) ---

type projectSearchExecute struct{}

func (a *projectSearchExecute) ID() string { return "project.search_execute" }

func (a *projectSearchExecute) Run(ctx *ActionContext) error {
	ps := ctx.App.ProjectSearchState()
	if ps == nil || ps.Query == "" {
		return nil
	}

	ctx.App.RunProjectSearch(ps.Query)
	return nil
}

// --- project.search_up / project.search_down: navigate results ---

type projectSearchUp struct{}

func (a *projectSearchUp) ID() string { return "project.search_up" }
func (a *projectSearchUp) Run(ctx *ActionContext) error {
	if ps := ctx.App.ProjectSearchState(); ps != nil {
		ps.MoveUp()
	}
	return nil
}

type projectSearchDown struct{}

func (a *projectSearchDown) ID() string { return "project.search_down" }
func (a *projectSearchDown) Run(ctx *ActionContext) error {
	if ps := ctx.App.ProjectSearchState(); ps != nil {
		ps.MoveDown(projectSearchMaxVisible)
	}
	return nil
}

// --- project.search_open_result: open the selected result (Enter on result) ---

type projectSearchOpenResult struct{}

func (a *projectSearchOpenResult) ID() string { return "project.search_open_result" }

func (a *projectSearchOpenResult) Run(ctx *ActionContext) error {
	ps := ctx.App.ProjectSearchState()
	if ps == nil {
		return nil
	}

	entry := ps.SelectedEntry()
	if entry == nil {
		return nil
	}

	// If it's a header, toggle expand/collapse (future feature) or do nothing.
	if entry.IsHeader {
		return nil
	}

	// Open the file and navigate to the line.
	path := entry.Path
	line := entry.Line - 1 // Convert 1-based to 0-based.
	col := entry.Col - 1
	if col < 0 {
		col = 0
	}

	// Close project search panel.
	ctx.App.SetProjectSearchState(nil)
	ctx.App.SetInputMode(ModeNormal)

	// Open the file.
	if err := ctx.App.OpenFile(path); err != nil {
		ctx.App.SetStatusMessage(fmt.Sprintf("Cannot open: %s", err))
		return nil
	}

	// Navigate to position.
	ctx.App.NavigateToPosition(path, line, col)
	return nil
}

// --- project.search_next: jump to next result without closing panel ---

type projectSearchNext struct{}

func (a *projectSearchNext) ID() string { return "project.search_next" }

func (a *projectSearchNext) Run(ctx *ActionContext) error {
	ps := ctx.App.ProjectSearchState()
	if ps == nil || len(ps.Entries) == 0 {
		return nil
	}

	// Move to next non-header entry.
	start := ps.SelectedIdx + 1
	for i := start; i < len(ps.Entries); i++ {
		if !ps.Entries[i].IsHeader {
			ps.SelectedIdx = i
			if ps.SelectedIdx >= ps.ScrollY+projectSearchMaxVisible {
				ps.ScrollY = ps.SelectedIdx - projectSearchMaxVisible + 1
			}
			return nil
		}
	}

	// Wrap around.
	for i := 0; i < start && i < len(ps.Entries); i++ {
		if !ps.Entries[i].IsHeader {
			ps.SelectedIdx = i
			ps.ScrollY = 0
			return nil
		}
	}

	return nil
}

// RegisterProjectSearchActions registers all project search actions.
func RegisterProjectSearchActions(r *Registry) {
	r.Register(&projectSearchOpen{})
	r.Register(&projectSearchClose{})
	r.Register(&projectSearchChar{})
	r.Register(&projectSearchBackspace{})
	r.Register(&projectSearchDelete{})
	r.Register(&projectSearchLeft{})
	r.Register(&projectSearchRight{})
	r.Register(&projectSearchHome{})
	r.Register(&projectSearchEnd{})
	r.Register(&projectSearchExecute{})
	r.Register(&projectSearchUp{})
	r.Register(&projectSearchDown{})
	r.Register(&projectSearchOpenResult{})
	r.Register(&projectSearchNext{})
}
