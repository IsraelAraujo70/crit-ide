package actions

// TerminalState is the interface actions use to interact with the terminal.
// It avoids importing the terminal package directly from actions.
type TerminalState interface {
	TerminalToggle()
	TerminalNew()
	TerminalClose()
	TerminalNext()
	TerminalSendInput(data []byte)
	TerminalVisible() bool
	TerminalFocused() bool
}

// ModeTerminal is the input mode for the terminal panel.
// Added here to extend the InputMode enum.
const ModeTerminal InputMode = 12

// FocusTerminal is the focus area for the terminal panel.
const FocusTerminal FocusArea = 3

// --- Terminal toggle action ---

type terminalToggle struct{}

func (a *terminalToggle) ID() string { return "terminal.toggle" }
func (a *terminalToggle) Run(ctx *ActionContext) error {
	ts, ok := ctx.App.(TerminalState)
	if !ok {
		return nil
	}
	ts.TerminalToggle()
	return nil
}

// --- Terminal new session action ---

type terminalNew struct{}

func (a *terminalNew) ID() string { return "terminal.new" }
func (a *terminalNew) Run(ctx *ActionContext) error {
	ts, ok := ctx.App.(TerminalState)
	if !ok {
		return nil
	}
	ts.TerminalNew()
	return nil
}

// --- Terminal close session action ---

type terminalClose struct{}

func (a *terminalClose) ID() string { return "terminal.close" }
func (a *terminalClose) Run(ctx *ActionContext) error {
	ts, ok := ctx.App.(TerminalState)
	if !ok {
		return nil
	}
	ts.TerminalClose()
	return nil
}

// --- Terminal next session action ---

type terminalNext struct{}

func (a *terminalNext) ID() string { return "terminal.next" }
func (a *terminalNext) Run(ctx *ActionContext) error {
	ts, ok := ctx.App.(TerminalState)
	if !ok {
		return nil
	}
	ts.TerminalNext()
	return nil
}

// RegisterTerminalActions registers all terminal-related actions.
func RegisterTerminalActions(r *Registry) {
	r.Register(&terminalToggle{})
	r.Register(&terminalNew{})
	r.Register(&terminalClose{})
	r.Register(&terminalNext{})
}
