package actions

import (
	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/lsp"
)

// RegisterLSPActions registers all LSP-related actions.
func RegisterLSPActions(r *Registry) {
	r.Register(&lspDefinition{})
	r.Register(&lspHover{})
	r.Register(&lspFormat{})
	r.Register(&lspRename{})
	r.Register(&lspCodeAction{})
	r.Register(&lspSignatureHelp{})
	r.Register(&codeActionUp{})
	r.Register(&codeActionDown{})
	r.Register(&codeActionExecute{})
	r.Register(&codeActionDismiss{})
}

// --- lsp.definition ---

type lspDefinition struct{}

func (a *lspDefinition) ID() string { return "lsp.definition" }

func (a *lspDefinition) Run(ctx *ActionContext) error {
	buf := ctx.App.ActiveBuffer()
	if buf.LanguageID == "" {
		return nil
	}
	srvAny := ctx.App.LSPServer(buf.LanguageID)
	if srvAny == nil {
		ctx.App.SetStatusMessage("LSP: no server running")
		return nil
	}
	srv, ok := srvAny.(*lsp.Server)
	if !ok {
		return nil
	}
	uri := lsp.URIFromPath(buf.Path)
	lineContent := buf.Text.Line(buf.CursorRow)
	pos := lsp.EditorToLSPPosition(buf.CursorRow, buf.CursorCol, lineContent)
	srv.Definition(uri, pos)
	return nil
}

// --- lsp.hover ---

type lspHover struct{}

func (a *lspHover) ID() string { return "lsp.hover" }

func (a *lspHover) Run(ctx *ActionContext) error {
	buf := ctx.App.ActiveBuffer()
	if buf.LanguageID == "" {
		return nil
	}
	srvAny := ctx.App.LSPServer(buf.LanguageID)
	if srvAny == nil {
		ctx.App.SetStatusMessage("LSP: no server running")
		return nil
	}
	srv, ok := srvAny.(*lsp.Server)
	if !ok {
		return nil
	}
	uri := lsp.URIFromPath(buf.Path)
	lineContent := buf.Text.Line(buf.CursorRow)
	pos := lsp.EditorToLSPPosition(buf.CursorRow, buf.CursorCol, lineContent)
	srv.HoverRequest(uri, pos)
	return nil
}

// --- lsp.format ---

type lspFormat struct{}

func (a *lspFormat) ID() string { return "lsp.format" }

func (a *lspFormat) Run(ctx *ActionContext) error {
	buf := ctx.App.ActiveBuffer()
	if buf.LanguageID == "" {
		return nil
	}
	srvAny := ctx.App.LSPServer(buf.LanguageID)
	if srvAny == nil {
		ctx.App.SetStatusMessage("LSP: no server running")
		return nil
	}
	srv, ok := srvAny.(*lsp.Server)
	if !ok {
		return nil
	}
	uri := lsp.URIFromPath(buf.Path)
	srv.Format(uri)
	return nil
}

// --- lsp.rename (F2) ---

type lspRename struct{}

func (a *lspRename) ID() string { return "lsp.rename" }

func (a *lspRename) Run(ctx *ActionContext) error {
	buf := ctx.App.ActiveBuffer()
	if buf.LanguageID == "" {
		return nil
	}
	srvAny := ctx.App.LSPServer(buf.LanguageID)
	if srvAny == nil {
		ctx.App.SetStatusMessage("LSP: no server running")
		return nil
	}
	srv, ok := srvAny.(*lsp.Server)
	if !ok {
		return nil
	}
	if !srv.HasRenameProvider() {
		ctx.App.SetStatusMessage("LSP: rename not supported")
		return nil
	}
	// Get current word under cursor for pre-fill.
	word := wordUnderCursor(buf)
	ctx.App.SetPrompt(&editor.PromptState{
		Kind:      editor.PromptLSPRename,
		Label:     "Rename: ",
		Input:     word,
		CursorPos: len(word),
	})
	ctx.App.SetInputMode(ModePrompt)
	return nil
}

// wordUnderCursor extracts the identifier word under the cursor.
func wordUnderCursor(buf *editor.Buffer) string {
	line := buf.Text.Line(buf.CursorRow)
	if len(line) == 0 {
		return ""
	}
	col := buf.CursorCol
	if col > len(line) {
		col = len(line)
	}
	// Find word boundaries.
	start := col
	for start > 0 && isWordChar(line[start-1]) {
		start--
	}
	end := col
	for end < len(line) && isWordChar(line[end]) {
		end++
	}
	if start == end {
		return ""
	}
	return line[start:end]
}

// isWordChar returns true if a byte is part of an identifier word.
func isWordChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

// --- lsp.code_action (Ctrl+.) ---

type lspCodeAction struct{}

func (a *lspCodeAction) ID() string { return "lsp.code_action" }

func (a *lspCodeAction) Run(ctx *ActionContext) error {
	buf := ctx.App.ActiveBuffer()
	if buf.LanguageID == "" {
		return nil
	}
	srvAny := ctx.App.LSPServer(buf.LanguageID)
	if srvAny == nil {
		ctx.App.SetStatusMessage("LSP: no server running")
		return nil
	}
	srv, ok := srvAny.(*lsp.Server)
	if !ok {
		return nil
	}
	if !srv.HasCodeActionProvider() {
		ctx.App.SetStatusMessage("LSP: code actions not supported")
		return nil
	}
	uri := lsp.URIFromPath(buf.Path)
	lineContent := buf.Text.Line(buf.CursorRow)
	pos := lsp.EditorToLSPPosition(buf.CursorRow, buf.CursorCol, lineContent)
	// Use cursor position as both start and end of range.
	rng := lsp.Range{Start: pos, End: pos}
	srv.RequestCodeAction(uri, rng, nil)
	return nil
}

// --- lsp.signature_help ---

type lspSignatureHelp struct{}

func (a *lspSignatureHelp) ID() string { return "lsp.signature_help" }

func (a *lspSignatureHelp) Run(ctx *ActionContext) error {
	buf := ctx.App.ActiveBuffer()
	if buf.LanguageID == "" {
		return nil
	}
	srvAny := ctx.App.LSPServer(buf.LanguageID)
	if srvAny == nil {
		return nil
	}
	srv, ok := srvAny.(*lsp.Server)
	if !ok {
		return nil
	}
	uri := lsp.URIFromPath(buf.Path)
	lineContent := buf.Text.Line(buf.CursorRow)
	pos := lsp.EditorToLSPPosition(buf.CursorRow, buf.CursorCol, lineContent)
	srv.RequestSignatureHelp(uri, pos)
	return nil
}

// --- Code Actions popup actions ---

type codeActionUp struct{}

func (a *codeActionUp) ID() string { return "code_action.up" }
func (a *codeActionUp) Run(ctx *ActionContext) error {
	if ca := ctx.App.CodeActionsState(); ca != nil {
		ca.MoveUp()
	}
	return nil
}

type codeActionDown struct{}

func (a *codeActionDown) ID() string { return "code_action.down" }
func (a *codeActionDown) Run(ctx *ActionContext) error {
	if ca := ctx.App.CodeActionsState(); ca != nil {
		ca.MoveDown()
	}
	return nil
}

type codeActionExecute struct{}

func (a *codeActionExecute) ID() string { return "code_action.execute" }
func (a *codeActionExecute) Run(ctx *ActionContext) error {
	ca := ctx.App.CodeActionsState()
	if ca == nil {
		return nil
	}
	item := ca.SelectedItem()
	if item == nil {
		return nil
	}
	idx := item.Index
	ctx.App.SetCodeActionsState(nil)
	ctx.App.SetInputMode(ModeNormal)
	ctx.App.ApplyCodeAction(idx)
	return nil
}

type codeActionDismiss struct{}

func (a *codeActionDismiss) ID() string { return "code_action.dismiss" }
func (a *codeActionDismiss) Run(ctx *ActionContext) error {
	ctx.App.SetCodeActionsState(nil)
	ctx.App.SetInputMode(ModeNormal)
	return nil
}
