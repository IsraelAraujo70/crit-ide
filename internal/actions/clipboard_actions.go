package actions

// --- Clipboard copy action ---

type clipboardCopy struct{}

func (a *clipboardCopy) ID() string { return "clipboard.copy" }

func (a *clipboardCopy) Run(ctx *ActionContext) error {
	buf := ctx.App.ActiveBuffer()
	if !buf.HasSelection() {
		return nil
	}
	text := buf.SelectedText()
	return ctx.App.Clipboard().Write(text)
}

// --- Clipboard cut action ---

type clipboardCut struct{}

func (a *clipboardCut) ID() string { return "clipboard.cut" }

func (a *clipboardCut) Run(ctx *ActionContext) error {
	buf := ctx.App.ActiveBuffer()
	if !buf.HasSelection() {
		return nil
	}
	text := buf.SelectedText()
	if err := ctx.App.Clipboard().Write(text); err != nil {
		return err
	}
	buf.DeleteSelection()
	return nil
}

// --- Clipboard paste action ---

type clipboardPaste struct{}

func (a *clipboardPaste) ID() string { return "clipboard.paste" }

func (a *clipboardPaste) Run(ctx *ActionContext) error {
	text, err := ctx.App.Clipboard().Read()
	if err != nil || text == "" {
		return err
	}
	buf := ctx.App.ActiveBuffer()
	// ReplaceSelection handles both cases: if there's a selection it deletes
	// it first, then inserts at cursor. If no selection, it just inserts.
	buf.ReplaceSelection(text)
	return nil
}
