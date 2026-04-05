package editor

// MenuItem represents a single item in a context menu.
type MenuItem struct {
	Label       string
	ActionID    string
	IsSeparator bool
}

// MenuState holds the state of an open context menu.
type MenuState struct {
	ScreenX     int
	ScreenY     int
	Items       []MenuItem
	SelectedIdx int
}
