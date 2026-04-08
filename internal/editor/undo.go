package editor

// EditKind classifies the type of edit operation for undo/redo.
type EditKind int

const (
	EditInsert EditKind = iota // Text was inserted.
	EditDelete                 // Text was deleted.
)

// UndoEntry captures a single atomic edit operation that can be reversed.
type UndoEntry struct {
	Kind      EditKind
	Pos       Position // Where the edit happened (start position).
	Text      string   // Text that was inserted (EditInsert) or deleted (EditDelete).
	CursorRow int      // Cursor row before the edit.
	CursorCol int      // Cursor col before the edit.
}

// UndoStack manages undo and redo history for a buffer.
type UndoStack struct {
	undos   []UndoEntry
	redos   []UndoEntry
	maxSize int
	// Group tracking: sequential single-char inserts/deletes are grouped.
	grouping bool
}

// NewUndoStack creates a new undo stack with the given max history size.
func NewUndoStack(maxSize int) *UndoStack {
	return &UndoStack{
		maxSize: maxSize,
	}
}

// Push adds a new undo entry. Clears the redo stack (new edit branch).
func (s *UndoStack) Push(entry UndoEntry) {
	s.undos = append(s.undos, entry)
	if len(s.undos) > s.maxSize {
		// Trim oldest entries.
		s.undos = s.undos[len(s.undos)-s.maxSize:]
	}
	// New edit invalidates redo history.
	s.redos = s.redos[:0]
}

// CanUndo returns true if there are entries to undo.
func (s *UndoStack) CanUndo() bool {
	return len(s.undos) > 0
}

// CanRedo returns true if there are entries to redo.
func (s *UndoStack) CanRedo() bool {
	return len(s.redos) > 0
}

// PopUndo removes and returns the most recent undo entry.
// Returns the entry and true, or a zero entry and false if empty.
func (s *UndoStack) PopUndo() (UndoEntry, bool) {
	if len(s.undos) == 0 {
		return UndoEntry{}, false
	}
	entry := s.undos[len(s.undos)-1]
	s.undos = s.undos[:len(s.undos)-1]
	return entry, true
}

// PushRedo adds an entry to the redo stack.
func (s *UndoStack) PushRedo(entry UndoEntry) {
	s.redos = append(s.redos, entry)
	if len(s.redos) > s.maxSize {
		s.redos = s.redos[len(s.redos)-s.maxSize:]
	}
}

// PopRedo removes and returns the most recent redo entry.
func (s *UndoStack) PopRedo() (UndoEntry, bool) {
	if len(s.redos) == 0 {
		return UndoEntry{}, false
	}
	entry := s.redos[len(s.redos)-1]
	s.redos = s.redos[:len(s.redos)-1]
	return entry, true
}

// Clear empties both undo and redo stacks.
func (s *UndoStack) Clear() {
	s.undos = s.undos[:0]
	s.redos = s.redos[:0]
}
