package actions

import (
	"github.com/israelcorrea/crit-ide/internal/editor"
)

// --- Go to Line: open prompt ---

type gotoLine struct{}

func (a *gotoLine) ID() string { return "goto.line" }

func (a *gotoLine) Run(ctx *ActionContext) error {
	ctx.App.SetPrompt(&editor.PromptState{
		Kind:  editor.PromptGotoLine,
		Label: "Go to line: ",
	})
	ctx.App.SetInputMode(ModePrompt)
	return nil
}
