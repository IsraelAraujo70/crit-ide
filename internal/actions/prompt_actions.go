package actions

import (
	"strconv"
	"strings"

	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/lsp"
)

// --- Prompt: insert character ---

type promptChar struct{}

func (a *promptChar) ID() string { return "prompt.char" }

func (a *promptChar) Run(ctx *ActionContext) error {
	p := ctx.App.Prompt()
	if p == nil {
		return nil
	}
	ch, ok := ctx.Event.Payload.(rune)
	if !ok {
		return nil
	}
	p.InsertChar(ch)
	return nil
}

// --- Prompt: backspace ---

type promptBackspace struct{}

func (a *promptBackspace) ID() string { return "prompt.backspace" }

func (a *promptBackspace) Run(ctx *ActionContext) error {
	p := ctx.App.Prompt()
	if p == nil {
		return nil
	}
	p.DeleteBackward()
	return nil
}

// --- Prompt: delete forward ---

type promptDelete struct{}

func (a *promptDelete) ID() string { return "prompt.delete" }

func (a *promptDelete) Run(ctx *ActionContext) error {
	p := ctx.App.Prompt()
	if p == nil {
		return nil
	}
	p.DeleteForward()
	return nil
}

// --- Prompt: cursor left ---

type promptLeft struct{}

func (a *promptLeft) ID() string { return "prompt.left" }

func (a *promptLeft) Run(ctx *ActionContext) error {
	p := ctx.App.Prompt()
	if p == nil {
		return nil
	}
	p.MoveLeft()
	return nil
}

// --- Prompt: cursor right ---

type promptRight struct{}

func (a *promptRight) ID() string { return "prompt.right" }

func (a *promptRight) Run(ctx *ActionContext) error {
	p := ctx.App.Prompt()
	if p == nil {
		return nil
	}
	p.MoveRight()
	return nil
}

// --- Prompt: home ---

type promptHome struct{}

func (a *promptHome) ID() string { return "prompt.home" }

func (a *promptHome) Run(ctx *ActionContext) error {
	p := ctx.App.Prompt()
	if p == nil {
		return nil
	}
	p.MoveHome()
	return nil
}

// --- Prompt: end ---

type promptEnd struct{}

func (a *promptEnd) ID() string { return "prompt.end" }

func (a *promptEnd) Run(ctx *ActionContext) error {
	p := ctx.App.Prompt()
	if p == nil {
		return nil
	}
	p.MoveEnd()
	return nil
}

// --- Prompt: cancel (Escape) ---

type promptCancel struct{}

func (a *promptCancel) ID() string { return "prompt.cancel" }

func (a *promptCancel) Run(ctx *ActionContext) error {
	p := ctx.App.Prompt()
	ctx.App.SetPrompt(nil)
	// If we were in a git commit prompt, return to git status.
	if p != nil && p.Kind == editor.PromptGitCommit && ctx.App.GitStatusState() != nil {
		ctx.App.SetInputMode(ModeGitStatus)
		return nil
	}
	ctx.App.SetInputMode(ModeNormal)
	return nil
}

// --- Prompt: confirm (Enter) ---

type promptConfirm struct{}

func (a *promptConfirm) ID() string { return "prompt.confirm" }

func (a *promptConfirm) Run(ctx *ActionContext) error {
	p := ctx.App.Prompt()
	if p == nil {
		return nil
	}

	input := strings.TrimSpace(p.Input)

	// Handle git commit prompt.
	if p.Kind == editor.PromptGitCommit {
		if input != "" {
			// Keep prompt with the input so git.commit.execute can read it.
			ctx.App.PostAction("git.commit.execute")
		} else {
			ctx.App.SetPrompt(nil)
			ctx.App.SetStatusMessage("Commit cancelled: empty message")
			if ctx.App.GitStatusState() != nil {
				ctx.App.SetInputMode(ModeGitStatus)
			} else {
				ctx.App.SetInputMode(ModeNormal)
			}
		}
		return nil
	}

	// Handle LSP rename prompt.
	if p.Kind == editor.PromptLSPRename {
		if input != "" {
			buf := ctx.App.ActiveBuffer()
			if buf.LanguageID != "" {
				srvAny := ctx.App.LSPServer(buf.LanguageID)
				if srvAny != nil {
					if srv, ok := srvAny.(*lsp.Server); ok {
						uri := lsp.URIFromPath(buf.Path)
						lineContent := buf.Text.Line(buf.CursorRow)
						pos := lsp.EditorToLSPPosition(buf.CursorRow, buf.CursorCol, lineContent)
						srv.RequestRename(uri, pos, input)
					}
				}
			}
		}
		ctx.App.SetPrompt(nil)
		ctx.App.SetInputMode(ModeNormal)
		return nil
	}

	// Handle goto-line prompt (does not require file tree).
	if p.Kind == editor.PromptGotoLine {
		if input != "" {
			lineNum, err := strconv.Atoi(input)
			if err == nil {
				buf := ctx.App.ActiveBuffer()
				// Convert 1-indexed user input to 0-indexed internal row.
				row := lineNum - 1
				if row < 0 {
					row = 0
				}
				maxRow := buf.Text.LineCount() - 1
				if maxRow < 0 {
					maxRow = 0
				}
				if row > maxRow {
					row = maxRow
				}
				buf.CursorRow = row
				buf.CursorCol = 0
				buf.ClampCursor()
				ctx.App.NavigateToPosition(buf.Path, row, 0)
			}
		}
		ctx.App.SetPrompt(nil)
		ctx.App.SetInputMode(ModeNormal)
		return nil
	}

	ft := ctx.App.FileTree()
	if ft == nil {
		ctx.App.SetPrompt(nil)
		ctx.App.SetInputMode(ModeNormal)
		return nil
	}

	switch p.Kind {
	case editor.PromptNewFile:
		if input != "" {
			path, err := ft.CreateFile(input)
			if err == nil && path != "" && !strings.HasSuffix(input, "/") {
				// Open the newly created file.
				_ = ctx.App.OpenFile(path)
				ctx.App.SetFocusArea(FocusEditor)
			}
		}

	case editor.PromptRename:
		if input != "" {
			newPath, err := ft.Rename(input)
			// If a file was renamed and it's currently open, we could
			// update the buffer path. For now, the tree just refreshes.
			_ = newPath
			_ = err
		}

	case editor.PromptDelete:
		if strings.ToLower(input) == "y" {
			_ = ft.Delete()
		}
	}

	ctx.App.SetPrompt(nil)
	ctx.App.SetInputMode(ModeNormal)
	return nil
}

// --- Tree file operations: open prompts ---

// tree.new — opens prompt for creating a new file/directory.
type treeNew struct{}

func (a *treeNew) ID() string { return "tree.new" }

func (a *treeNew) Run(ctx *ActionContext) error {
	ft := ctx.App.FileTree()
	if ft == nil {
		return nil
	}

	ctx.App.SetPrompt(&editor.PromptState{
		Kind:  editor.PromptNewFile,
		Label: "New file (end with / for dir): ",
	})
	ctx.App.SetInputMode(ModePrompt)
	return nil
}

// tree.rename — opens prompt for renaming the cursor node.
type treeRename struct{}

func (a *treeRename) ID() string { return "tree.rename" }

func (a *treeRename) Run(ctx *ActionContext) error {
	ft := ctx.App.FileTree()
	if ft == nil {
		return nil
	}
	if ft.CursorIsRoot() {
		return nil
	}

	name := ft.CursorNodeName()
	ctx.App.SetPrompt(&editor.PromptState{
		Kind:      editor.PromptRename,
		Label:     "Rename: ",
		Input:     name,
		CursorPos: len(name),
		Context:   ft.CursorNodePath(),
	})
	ctx.App.SetInputMode(ModePrompt)
	return nil
}

// tree.delete — opens confirmation prompt for deletion.
type treeDelete struct{}

func (a *treeDelete) ID() string { return "tree.delete" }

func (a *treeDelete) Run(ctx *ActionContext) error {
	ft := ctx.App.FileTree()
	if ft == nil {
		return nil
	}
	if ft.CursorIsRoot() {
		return nil
	}

	name := ft.CursorNodeName()
	ctx.App.SetPrompt(&editor.PromptState{
		Kind:    editor.PromptDelete,
		Label:   "Delete '" + name + "'? (y/n): ",
		Context: ft.CursorNodePath(),
	})
	ctx.App.SetInputMode(ModePrompt)
	return nil
}
