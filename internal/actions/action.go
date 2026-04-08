package actions

import (
	"fmt"

	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/events"
)

// InputMode represents the current input routing mode.
type InputMode int

const (
	ModeNormal      InputMode = iota // Normal editing mode.
	ModeContextMenu                  // Context menu is open.
	ModePrompt                       // Input prompt is active.
	ModeSearch                       // Find/Replace bar is active.
	ModeFileFinder                   // Fuzzy file finder popup is active.
	ModeCompletion                   // Autocomplete popup is active.
	ModeCommandPalette               // Command palette popup is active.
	ModeProjectSearch                // Project-wide search panel is active.
	ModeGitStatus                    // Git status panel is active.
	ModeGitGraph                     // Git graph panel is active.
	ModeGitDiff                      // Git diff viewer is active.
	ModeCodeActions                  // Code actions popup is active.
)

// FocusArea indicates which panel currently has keyboard focus.
type FocusArea int

const (
	FocusEditor   FocusArea = iota // Editor pane has focus.
	FocusFileTree                  // File tree panel has focus.
	FocusGitPanel                  // Git status/graph panel has focus.
)

// ClipboardProvider abstracts clipboard read/write for actions.
type ClipboardProvider interface {
	Read() (string, error)
	Write(text string) error
}

// FileTreeState is the interface actions use to interact with the file tree.
// It avoids importing the filetree package directly from actions.
type FileTreeState interface {
	MoveUp()
	MoveDown()
	Toggle() string // Returns file path if a file was selected, "" otherwise.
	Expand()
	Collapse()
	Refresh()
	SetCursorToScreenRow(row int)
	EnsureCursorVisible(viewportHeight int)

	// File operations.
	CreateFile(name string) (string, error)
	Rename(newName string) (string, error)
	Delete() error
	CursorNodeName() string
	CursorNodePath() string
	CursorIsRoot() bool
}

// AppState is the interface that actions use to interact with the application.
// It breaks the circular dependency: actions don't import app, app implements this.
type AppState interface {
	ActiveBuffer() *editor.Buffer
	ScrollY() int
	SetScrollY(y int)
	ViewportHeight() int
	ScreenWidth() int
	Quit()

	// Clipboard access.
	Clipboard() ClipboardProvider

	// Input mode and context menu.
	InputMode() InputMode
	SetInputMode(mode InputMode)
	ContextMenu() *editor.MenuState
	SetContextMenu(menu *editor.MenuState)

	// Pending action trampoline for menu execution.
	PostAction(actionID string)

	// Tab / multi-buffer management.
	Buffers() []*editor.Buffer
	ActiveBufferIndex() int
	OpenFile(path string) error
	CloseBuffer(idx int)
	SwitchBuffer(idx int)

	// File tree.
	FileTree() FileTreeState
	FileTreeVisible() bool
	SetFileTreeVisible(v bool)
	ToggleFileTree()
	FileTreeWidth() int
	TreeViewportHeight() int

	// Focus area.
	FocusArea() FocusArea
	SetFocusArea(area FocusArea)

	// Input prompt.
	Prompt() *editor.PromptState
	SetPrompt(p *editor.PromptState)

	// LSP support.
	LSPServer(langID string) any                      // Returns *lsp.Server or nil (typed as any to avoid import cycle).
	SetStatusMessage(msg string)                      // Show temporary message in statusline.
	NavigateToPosition(path string, line, col int)    // Jump to position (same file) or show path (different file).

	// Search.
	SearchState() *editor.SearchState
	SetSearchState(s *editor.SearchState)

	// File finder.
	FinderState() *editor.FinderState
	SetFinderState(f *editor.FinderState)
	FinderFilter(pattern string) []editor.FinderResult
	FinderRebuildCache()
	FinderFileCount() int

	// Completion.
	CompletionState() *editor.CompletionState
	SetCompletionState(c *editor.CompletionState)
	TriggerCompletion(triggerChar string)

	// Command palette.
	PaletteState() *editor.PaletteState
	SetPaletteState(p *editor.PaletteState)

	// Project search.
	ProjectSearchState() *editor.ProjectSearchState
	SetProjectSearchState(ps *editor.ProjectSearchState)
	RunProjectSearch(query string)
	ProjectRoot() string

	// Git status.
	GitStatusState() *editor.GitStatusState
	SetGitStatusState(gs *editor.GitStatusState)

	// Git graph.
	GitGraphState() *editor.GitGraphState
	SetGitGraphState(gg *editor.GitGraphState)

	// Git diff.
	GitDiffState() *editor.GitDiffState
	SetGitDiffState(gd *editor.GitDiffState)

	// Code actions.
	CodeActionsState() *editor.CodeActionsState
	SetCodeActionsState(ca *editor.CodeActionsState)
	ApplyCodeAction(idx int) // Apply code action by index in the original LSP list.

	// Signature help.
	SignatureHelpState() *editor.SignatureHelpState
	SetSignatureHelpState(sh *editor.SignatureHelpState)
}

// ActionContext carries everything an action needs to execute.
type ActionContext struct {
	App   AppState
	Event *events.Event
}

// Action is the fundamental unit of behavior in the IDE.
// Every user-visible operation is an Action.
type Action interface {
	ID() string
	Run(ctx *ActionContext) error
}

// Registry maps action IDs to their implementations.
type Registry struct {
	actions map[string]Action
}

// NewRegistry creates an empty action registry.
func NewRegistry() *Registry {
	return &Registry{actions: make(map[string]Action)}
}

// Register adds an action to the registry. Panics on duplicate IDs.
func (r *Registry) Register(a Action) {
	if _, exists := r.actions[a.ID()]; exists {
		panic(fmt.Sprintf("action already registered: %s", a.ID()))
	}
	r.actions[a.ID()] = a
}

// Execute runs the action with the given ID. Returns an error if not found.
func (r *Registry) Execute(id string, ctx *ActionContext) error {
	a, ok := r.actions[id]
	if !ok {
		return fmt.Errorf("unknown action: %s", id)
	}
	return a.Run(ctx)
}
