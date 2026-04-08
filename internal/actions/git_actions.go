package actions

import (
	"github.com/israelcorrea/crit-ide/internal/editor"
)

// GitState is the interface that git actions use to interact with Git.
// It avoids importing the git package directly from actions.
type GitState interface {
	GitStatusEntries() []editor.GitStatusEntry
	GitCurrentBranch() string
	GitStage(path string) error
	GitUnstage(path string) error
	GitCommit(msg string) error
	GitDiff(path string) string
	GitDiffStaged(path string) string
	GitGraphLines() []editor.GitGraphLine
	GitRefreshStatus()
}

// --- Git Status Panel Actions ---

// gitStatusOpen opens the Git status panel.
type gitStatusOpen struct{}

func (a *gitStatusOpen) ID() string { return "git.status" }
func (a *gitStatusOpen) Run(ctx *ActionContext) error {
	gs, ok := ctx.App.(GitState)
	if !ok {
		return nil
	}
	gs.GitRefreshStatus()

	entries := gs.GitStatusEntries()
	branch := gs.GitCurrentBranch()

	state := &editor.GitStatusState{
		Entries: entries,
		Branch:  branch,
	}

	ctx.App.SetGitStatusState(state)
	ctx.App.SetInputMode(ModeGitStatus)
	ctx.App.SetFocusArea(FocusGitPanel)
	return nil
}

// gitStatusClose closes the Git status panel.
type gitStatusClose struct{}

func (a *gitStatusClose) ID() string { return "git.status.close" }
func (a *gitStatusClose) Run(ctx *ActionContext) error {
	ctx.App.SetGitStatusState(nil)
	ctx.App.SetInputMode(ModeNormal)
	ctx.App.SetFocusArea(FocusEditor)
	return nil
}

// gitStatusUp moves the cursor up in the status panel.
type gitStatusUp struct{}

func (a *gitStatusUp) ID() string { return "git.status.up" }
func (a *gitStatusUp) Run(ctx *ActionContext) error {
	state := ctx.App.GitStatusState()
	if state == nil || len(state.Entries) == 0 {
		return nil
	}
	if state.SelectedIdx > 0 {
		state.SelectedIdx--
	}
	return nil
}

// gitStatusDown moves the cursor down in the status panel.
type gitStatusDown struct{}

func (a *gitStatusDown) ID() string { return "git.status.down" }
func (a *gitStatusDown) Run(ctx *ActionContext) error {
	state := ctx.App.GitStatusState()
	if state == nil || len(state.Entries) == 0 {
		return nil
	}
	if state.SelectedIdx < len(state.Entries)-1 {
		state.SelectedIdx++
	}
	return nil
}

// gitStage stages the file under cursor.
type gitStage struct{}

func (a *gitStage) ID() string { return "git.stage" }
func (a *gitStage) Run(ctx *ActionContext) error {
	gs, ok := ctx.App.(GitState)
	if !ok {
		return nil
	}
	state := ctx.App.GitStatusState()
	if state == nil || len(state.Entries) == 0 {
		return nil
	}
	entry := state.Entries[state.SelectedIdx]
	if entry.Staged {
		// Already staged — unstage it.
		if err := gs.GitUnstage(entry.Path); err != nil {
			ctx.App.SetStatusMessage("Unstage failed: " + err.Error())
			return nil
		}
		ctx.App.SetStatusMessage("Unstaged: " + entry.Path)
	} else {
		if err := gs.GitStage(entry.Path); err != nil {
			ctx.App.SetStatusMessage("Stage failed: " + err.Error())
			return nil
		}
		ctx.App.SetStatusMessage("Staged: " + entry.Path)
	}

	// Refresh status.
	gs.GitRefreshStatus()
	entries := gs.GitStatusEntries()
	state.Entries = entries
	if state.SelectedIdx >= len(entries) && len(entries) > 0 {
		state.SelectedIdx = len(entries) - 1
	}
	return nil
}

// gitUnstage explicitly unstages the file under cursor.
type gitUnstage struct{}

func (a *gitUnstage) ID() string { return "git.unstage" }
func (a *gitUnstage) Run(ctx *ActionContext) error {
	gs, ok := ctx.App.(GitState)
	if !ok {
		return nil
	}
	state := ctx.App.GitStatusState()
	if state == nil || len(state.Entries) == 0 {
		return nil
	}
	entry := state.Entries[state.SelectedIdx]
	if !entry.Staged {
		ctx.App.SetStatusMessage("File is not staged")
		return nil
	}
	if err := gs.GitUnstage(entry.Path); err != nil {
		ctx.App.SetStatusMessage("Unstage failed: " + err.Error())
		return nil
	}
	ctx.App.SetStatusMessage("Unstaged: " + entry.Path)

	// Refresh status.
	gs.GitRefreshStatus()
	entries := gs.GitStatusEntries()
	state.Entries = entries
	if state.SelectedIdx >= len(entries) && len(entries) > 0 {
		state.SelectedIdx = len(entries) - 1
	}
	return nil
}

// gitDiff shows the diff of the file under cursor.
type gitDiff struct{}

func (a *gitDiff) ID() string { return "git.diff" }
func (a *gitDiff) Run(ctx *ActionContext) error {
	gs, ok := ctx.App.(GitState)
	if !ok {
		return nil
	}
	state := ctx.App.GitStatusState()
	if state == nil || len(state.Entries) == 0 {
		return nil
	}
	entry := state.Entries[state.SelectedIdx]

	var diff string
	var title string
	if entry.Staged {
		diff = gs.GitDiffStaged(entry.Path)
		title = "Staged: " + entry.Path
	} else {
		diff = gs.GitDiff(entry.Path)
		title = "Modified: " + entry.Path
	}

	if diff == "" {
		ctx.App.SetStatusMessage("No diff available")
		return nil
	}

	ctx.App.SetGitDiffState(&editor.GitDiffState{
		Lines: splitLines(diff),
		Path:  entry.Path,
		Title: title,
	})
	ctx.App.SetInputMode(ModeGitDiff)
	return nil
}

// gitDiffClose closes the diff view and returns to status.
type gitDiffClose struct{}

func (a *gitDiffClose) ID() string { return "git.diff.close" }
func (a *gitDiffClose) Run(ctx *ActionContext) error {
	ctx.App.SetGitDiffState(nil)
	if ctx.App.GitStatusState() != nil {
		ctx.App.SetInputMode(ModeGitStatus)
	} else {
		ctx.App.SetInputMode(ModeNormal)
		ctx.App.SetFocusArea(FocusEditor)
	}
	return nil
}

// gitDiffScrollUp scrolls the diff view up.
type gitDiffScrollUp struct{}

func (a *gitDiffScrollUp) ID() string { return "git.diff.up" }
func (a *gitDiffScrollUp) Run(ctx *ActionContext) error {
	state := ctx.App.GitDiffState()
	if state == nil {
		return nil
	}
	if state.ScrollY > 0 {
		state.ScrollY--
	}
	return nil
}

// gitDiffScrollDown scrolls the diff view down.
type gitDiffScrollDown struct{}

func (a *gitDiffScrollDown) ID() string { return "git.diff.down" }
func (a *gitDiffScrollDown) Run(ctx *ActionContext) error {
	state := ctx.App.GitDiffState()
	if state == nil {
		return nil
	}
	if state.ScrollY < len(state.Lines)-1 {
		state.ScrollY++
	}
	return nil
}

// gitCommitOpen opens the commit prompt.
type gitCommitOpen struct{}

func (a *gitCommitOpen) ID() string { return "git.commit" }
func (a *gitCommitOpen) Run(ctx *ActionContext) error {
	ctx.App.SetPrompt(&editor.PromptState{
		Kind:  editor.PromptGitCommit,
		Label: "Commit message: ",
	})
	ctx.App.SetInputMode(ModePrompt)
	return nil
}

// gitCommitExecute performs the actual commit.
type gitCommitExecute struct{}

func (a *gitCommitExecute) ID() string { return "git.commit.execute" }
func (a *gitCommitExecute) Run(ctx *ActionContext) error {
	gs, ok := ctx.App.(GitState)
	if !ok {
		return nil
	}
	prompt := ctx.App.Prompt()
	if prompt == nil || prompt.Input == "" {
		ctx.App.SetStatusMessage("Commit cancelled: empty message")
		return nil
	}

	msg := prompt.Input
	ctx.App.SetPrompt(nil)

	if err := gs.GitCommit(msg); err != nil {
		ctx.App.SetStatusMessage("Commit failed: " + err.Error())
		ctx.App.SetInputMode(ModeGitStatus)
		return nil
	}

	ctx.App.SetStatusMessage("Committed: " + msg)

	// Refresh status.
	gs.GitRefreshStatus()
	state := ctx.App.GitStatusState()
	if state != nil {
		entries := gs.GitStatusEntries()
		state.Entries = entries
		state.SelectedIdx = 0
	}
	ctx.App.SetInputMode(ModeGitStatus)
	return nil
}

// --- Git Graph Actions ---

// gitGraphOpen opens the Git graph panel.
type gitGraphOpen struct{}

func (a *gitGraphOpen) ID() string { return "git.graph" }
func (a *gitGraphOpen) Run(ctx *ActionContext) error {
	gs, ok := ctx.App.(GitState)
	if !ok {
		return nil
	}

	lines := gs.GitGraphLines()
	state := &editor.GitGraphState{
		Lines: lines,
	}
	ctx.App.SetGitGraphState(state)
	ctx.App.SetInputMode(ModeGitGraph)
	ctx.App.SetFocusArea(FocusGitPanel)
	return nil
}

// gitGraphClose closes the Git graph panel.
type gitGraphClose struct{}

func (a *gitGraphClose) ID() string { return "git.graph.close" }
func (a *gitGraphClose) Run(ctx *ActionContext) error {
	ctx.App.SetGitGraphState(nil)
	ctx.App.SetInputMode(ModeNormal)
	ctx.App.SetFocusArea(FocusEditor)
	return nil
}

// gitGraphUp moves the cursor up in the graph.
type gitGraphUp struct{}

func (a *gitGraphUp) ID() string { return "git.graph.up" }
func (a *gitGraphUp) Run(ctx *ActionContext) error {
	state := ctx.App.GitGraphState()
	if state == nil || len(state.Lines) == 0 {
		return nil
	}
	if state.SelectedIdx > 0 {
		state.SelectedIdx--
	}
	return nil
}

// gitGraphDown moves the cursor down in the graph.
type gitGraphDown struct{}

func (a *gitGraphDown) ID() string { return "git.graph.down" }
func (a *gitGraphDown) Run(ctx *ActionContext) error {
	state := ctx.App.GitGraphState()
	if state == nil || len(state.Lines) == 0 {
		return nil
	}
	if state.SelectedIdx < len(state.Lines)-1 {
		state.SelectedIdx++
	}
	return nil
}

// gitStatusEnter opens the file under cursor in the editor.
type gitStatusEnter struct{}

func (a *gitStatusEnter) ID() string { return "git.status.enter" }
func (a *gitStatusEnter) Run(ctx *ActionContext) error {
	state := ctx.App.GitStatusState()
	if state == nil || len(state.Entries) == 0 {
		return nil
	}
	entry := state.Entries[state.SelectedIdx]

	// Close the git panel and open the file.
	ctx.App.SetGitStatusState(nil)
	ctx.App.SetInputMode(ModeNormal)
	ctx.App.SetFocusArea(FocusEditor)

	if err := ctx.App.OpenFile(entry.Path); err != nil {
		ctx.App.SetStatusMessage("Cannot open: " + err.Error())
	}
	return nil
}

// splitLines splits a string into lines.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// RegisterGitActions registers all git-related actions.
func RegisterGitActions(r *Registry) {
	r.Register(&gitStatusOpen{})
	r.Register(&gitStatusClose{})
	r.Register(&gitStatusUp{})
	r.Register(&gitStatusDown{})
	r.Register(&gitStage{})
	r.Register(&gitUnstage{})
	r.Register(&gitDiff{})
	r.Register(&gitDiffClose{})
	r.Register(&gitDiffScrollUp{})
	r.Register(&gitDiffScrollDown{})
	r.Register(&gitCommitOpen{})
	r.Register(&gitCommitExecute{})
	r.Register(&gitGraphOpen{})
	r.Register(&gitGraphClose{})
	r.Register(&gitGraphUp{})
	r.Register(&gitGraphDown{})
	r.Register(&gitStatusEnter{})
}
