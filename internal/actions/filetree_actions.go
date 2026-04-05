package actions

import (
	"github.com/israelcorrea/crit-ide/internal/events"
)

// --- File tree toggle action ---

type treeToggleVisible struct{}

func (a *treeToggleVisible) ID() string { return "tree.toggle" }

func (a *treeToggleVisible) Run(ctx *ActionContext) error {
	ctx.App.ToggleFileTree()
	return nil
}

// --- File tree cursor up ---

type treeUp struct{}

func (a *treeUp) ID() string { return "tree.up" }

func (a *treeUp) Run(ctx *ActionContext) error {
	ft := ctx.App.FileTree()
	if ft == nil {
		return nil
	}
	ft.MoveUp()
	ft.EnsureCursorVisible(ctx.App.TreeViewportHeight())
	return nil
}

// --- File tree cursor down ---

type treeDown struct{}

func (a *treeDown) ID() string { return "tree.down" }

func (a *treeDown) Run(ctx *ActionContext) error {
	ft := ctx.App.FileTree()
	if ft == nil {
		return nil
	}
	ft.MoveDown()
	ft.EnsureCursorVisible(ctx.App.TreeViewportHeight())
	return nil
}

// --- File tree enter / toggle ---

type treeEnter struct{}

func (a *treeEnter) ID() string { return "tree.enter" }

func (a *treeEnter) Run(ctx *ActionContext) error {
	ft := ctx.App.FileTree()
	if ft == nil {
		return nil
	}

	filePath := ft.Toggle()
	if filePath != "" {
		// Open the file in a new tab (or switch to existing tab).
		_ = ctx.App.OpenFile(filePath)
		ctx.App.SetFocusArea(FocusEditor)
	}
	return nil
}

// --- File tree expand (right arrow) ---

type treeExpand struct{}

func (a *treeExpand) ID() string { return "tree.expand" }

func (a *treeExpand) Run(ctx *ActionContext) error {
	ft := ctx.App.FileTree()
	if ft == nil {
		return nil
	}
	ft.Expand()
	return nil
}

// --- File tree collapse (left arrow) ---

type treeCollapse struct{}

func (a *treeCollapse) ID() string { return "tree.collapse" }

func (a *treeCollapse) Run(ctx *ActionContext) error {
	ft := ctx.App.FileTree()
	if ft == nil {
		return nil
	}
	ft.Collapse()
	return nil
}

// --- File tree click ---

type treeClick struct{}

func (a *treeClick) ID() string { return "tree.click" }

func (a *treeClick) Run(ctx *ActionContext) error {
	payload, ok := ctx.Event.Payload.(events.MouseClickPayload)
	if !ok {
		return nil
	}

	if !ctx.App.FileTreeVisible() {
		return nil
	}

	ft := ctx.App.FileTree()
	if ft == nil {
		return nil
	}

	// Calculate the row within the tree viewport.
	// Tab bar (1) + focus border (1) + EXPLORER header (1) = 3 rows before tree content.
	treeRow := payload.ScreenY - 3
	if treeRow < 0 {
		return nil
	}

	ft.SetCursorToScreenRow(treeRow)
	ctx.App.SetFocusArea(FocusFileTree)

	// Toggle the clicked node.
	filePath := ft.Toggle()
	if filePath != "" {
		_ = ctx.App.OpenFile(filePath)
		ctx.App.SetFocusArea(FocusEditor)
	}

	return nil
}

// --- File tree focus (escape returns to editor) ---

type treeFocusEditor struct{}

func (a *treeFocusEditor) ID() string { return "tree.focus.editor" }

func (a *treeFocusEditor) Run(ctx *ActionContext) error {
	ctx.App.SetFocusArea(FocusEditor)
	return nil
}

// --- File tree refresh ---

type treeRefresh struct{}

func (a *treeRefresh) ID() string { return "tree.refresh" }

func (a *treeRefresh) Run(ctx *ActionContext) error {
	ft := ctx.App.FileTree()
	if ft == nil {
		return nil
	}
	ft.Refresh()
	return nil
}
