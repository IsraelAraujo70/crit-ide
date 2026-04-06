package actions

import (
	"github.com/israelcorrea/crit-ide/internal/lsp"
)

// RegisterLSPActions registers all LSP-related actions.
func RegisterLSPActions(r *Registry) {
	r.Register(&lspDefinition{})
	r.Register(&lspHover{})
	r.Register(&lspFormat{})
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
