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
)

// ClipboardProvider abstracts clipboard read/write for actions.
type ClipboardProvider interface {
	Read() (string, error)
	Write(text string) error
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
