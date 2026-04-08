package editor

// TerminalState holds the display state for the embedded terminal panel.
type TerminalState struct {
	Visible      bool     // Whether the terminal panel is visible.
	Lines        []string // Output lines (may contain ANSI escape codes).
	ScrollY      int      // Scroll offset from bottom (0 = bottom, i.e., auto-scroll).
	Height       int      // Panel height in rows.
	ActiveTab    int      // Currently active session tab.
	TabNames     []string // Display names for session tabs.
	TabClosed    []bool   // Whether each session is closed.
	Focused      bool     // Whether the terminal panel has keyboard focus.
	CursorRow    int      // Cursor row relative to grid (0-based).
	CursorCol    int      // Cursor column (0-based).
	GridRows     int      // Number of rows in the terminal grid.
	// Selection (nil if none).
	SelStartLine int
	SelStartCol  int
	SelEndLine   int
	SelEndCol    int
	HasSelection bool
}

// ScrollUp scrolls the terminal output up by n lines.
func (t *TerminalState) ScrollUp(n int) {
	t.ScrollY += n
	maxScroll := len(t.Lines) - t.Height + 2 // +2 for header/tab bar
	if maxScroll < 0 {
		maxScroll = 0
	}
	if t.ScrollY > maxScroll {
		t.ScrollY = maxScroll
	}
}

// ScrollDown scrolls the terminal output down by n lines (toward the bottom).
func (t *TerminalState) ScrollDown(n int) {
	t.ScrollY -= n
	if t.ScrollY < 0 {
		t.ScrollY = 0
	}
}

// ScrollToBottom resets scroll to follow output.
func (t *TerminalState) ScrollToBottom() {
	t.ScrollY = 0
}
