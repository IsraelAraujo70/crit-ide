package actions

import (
	"fmt"

	"github.com/israelcorrea/crit-ide/internal/editor"
)

// finderMaxVisible is the maximum number of visible results in the popup.
const finderMaxVisible = 15

// --- finder.open: open the fuzzy file finder (Ctrl+P) ---

type finderOpen struct{}

func (a *finderOpen) ID() string { return "finder.open" }

func (a *finderOpen) Run(ctx *ActionContext) error {
	// Rebuild cache to ensure fresh file list.
	ctx.App.FinderRebuildCache()
	fs := editor.NewFinderState()
	fs.TotalFiles = ctx.App.FinderFileCount()
	// Populate initial results (no filter).
	results := ctx.App.FinderFilter("")
	fs.Results = results
	ctx.App.SetFinderState(fs)
	ctx.App.SetInputMode(ModeFileFinder)
	return nil
}

// --- finder.close: close the file finder (Escape) ---

type finderClose struct{}

func (a *finderClose) ID() string { return "finder.close" }

func (a *finderClose) Run(ctx *ActionContext) error {
	ctx.App.SetFinderState(nil)
	ctx.App.SetInputMode(ModeNormal)
	return nil
}

// --- finder.char: type a character in the finder input ---

type finderChar struct{}

func (a *finderChar) ID() string { return "finder.char" }

func (a *finderChar) Run(ctx *ActionContext) error {
	fs := ctx.App.FinderState()
	if fs == nil {
		return nil
	}
	ch, ok := ctx.Event.Payload.(rune)
	if !ok {
		return nil
	}
	fs.InsertChar(ch)
	// Re-filter.
	fs.Results = ctx.App.FinderFilter(fs.Query)
	return nil
}

// --- finder.backspace: delete character backward ---

type finderBackspace struct{}

func (a *finderBackspace) ID() string { return "finder.backspace" }

func (a *finderBackspace) Run(ctx *ActionContext) error {
	fs := ctx.App.FinderState()
	if fs == nil {
		return nil
	}
	fs.DeleteBackward()
	fs.Results = ctx.App.FinderFilter(fs.Query)
	return nil
}

// --- finder.delete: delete character forward ---

type finderDelete struct{}

func (a *finderDelete) ID() string { return "finder.delete" }

func (a *finderDelete) Run(ctx *ActionContext) error {
	fs := ctx.App.FinderState()
	if fs == nil {
		return nil
	}
	fs.DeleteForward()
	fs.Results = ctx.App.FinderFilter(fs.Query)
	return nil
}

// --- finder.left / finder.right / finder.home / finder.end ---

type finderLeft struct{}

func (a *finderLeft) ID() string { return "finder.left" }
func (a *finderLeft) Run(ctx *ActionContext) error {
	if fs := ctx.App.FinderState(); fs != nil {
		fs.MoveLeft()
	}
	return nil
}

type finderRight struct{}

func (a *finderRight) ID() string { return "finder.right" }
func (a *finderRight) Run(ctx *ActionContext) error {
	if fs := ctx.App.FinderState(); fs != nil {
		fs.MoveRight()
	}
	return nil
}

type finderHome struct{}

func (a *finderHome) ID() string { return "finder.home" }
func (a *finderHome) Run(ctx *ActionContext) error {
	if fs := ctx.App.FinderState(); fs != nil {
		fs.MoveHome()
	}
	return nil
}

type finderEnd struct{}

func (a *finderEnd) ID() string { return "finder.end" }
func (a *finderEnd) Run(ctx *ActionContext) error {
	if fs := ctx.App.FinderState(); fs != nil {
		fs.MoveEnd()
	}
	return nil
}

// --- finder.up / finder.down: navigate the result list ---

type finderUp struct{}

func (a *finderUp) ID() string { return "finder.up" }
func (a *finderUp) Run(ctx *ActionContext) error {
	if fs := ctx.App.FinderState(); fs != nil {
		fs.MoveUp()
	}
	return nil
}

type finderDown struct{}

func (a *finderDown) ID() string { return "finder.down" }
func (a *finderDown) Run(ctx *ActionContext) error {
	if fs := ctx.App.FinderState(); fs != nil {
		fs.MoveDown(finderMaxVisible)
	}
	return nil
}

// --- finder.confirm: open the selected file (Enter) ---

type finderConfirm struct{}

func (a *finderConfirm) ID() string { return "finder.confirm" }

func (a *finderConfirm) Run(ctx *ActionContext) error {
	fs := ctx.App.FinderState()
	if fs == nil {
		return nil
	}

	path := fs.SelectedPath()
	if path == "" {
		return nil
	}

	// Close finder first.
	ctx.App.SetFinderState(nil)
	ctx.App.SetInputMode(ModeNormal)

	// Open the file.
	if err := ctx.App.OpenFile(path); err != nil {
		ctx.App.SetStatusMessage(fmt.Sprintf("Cannot open: %s", err))
		return nil
	}
	return nil
}

// RegisterFinderActions registers all file finder actions.
func RegisterFinderActions(r *Registry) {
	r.Register(&finderOpen{})
	r.Register(&finderClose{})
	r.Register(&finderChar{})
	r.Register(&finderBackspace{})
	r.Register(&finderDelete{})
	r.Register(&finderLeft{})
	r.Register(&finderRight{})
	r.Register(&finderHome{})
	r.Register(&finderEnd{})
	r.Register(&finderUp{})
	r.Register(&finderDown{})
	r.Register(&finderConfirm{})
}
