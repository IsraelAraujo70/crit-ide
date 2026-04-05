package editor

// Selection represents a contiguous text selection in a buffer.
// Anchor is where the selection started (e.g., mouse-down position).
// Cursor is where the selection currently ends (e.g., current drag position).
// Use Normalized() to get a Range with Start <= End.
type Selection struct {
	Anchor Position
	Cursor Position
}

// Normalized returns a Range where Start <= End, suitable for
// TextStore.Slice() and TextStore.Delete().
func (s Selection) Normalized() Range {
	if s.Anchor.Line < s.Cursor.Line ||
		(s.Anchor.Line == s.Cursor.Line && s.Anchor.Col <= s.Cursor.Col) {
		return Range{Start: s.Anchor, End: s.Cursor}
	}
	return Range{Start: s.Cursor, End: s.Anchor}
}

// IsEmpty returns true if the selection has zero extent.
func (s Selection) IsEmpty() bool {
	return s.Anchor == s.Cursor
}
