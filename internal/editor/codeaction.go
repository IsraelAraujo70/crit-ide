package editor

// CodeActionItem represents a single code action in the editor.
type CodeActionItem struct {
	Title string
	Kind  string
	Index int // Index into the original LSP code action list.
}

// CodeActionsState holds the state of the code actions popup.
type CodeActionsState struct {
	Items       []CodeActionItem
	SelectedIdx int
	CursorRow   int // Editor row where popup was triggered.
	CursorCol   int // Editor col where popup was triggered.
}

// MoveUp moves the selection up.
func (s *CodeActionsState) MoveUp() {
	if s.SelectedIdx > 0 {
		s.SelectedIdx--
	}
}

// MoveDown moves the selection down.
func (s *CodeActionsState) MoveDown() {
	if s.SelectedIdx < len(s.Items)-1 {
		s.SelectedIdx++
	}
}

// SelectedItem returns the currently selected item, or nil if empty.
func (s *CodeActionsState) SelectedItem() *CodeActionItem {
	if len(s.Items) == 0 {
		return nil
	}
	return &s.Items[s.SelectedIdx]
}
