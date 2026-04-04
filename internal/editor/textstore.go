package editor

import "fmt"

// Position represents a zero-based line and column in a text document.
type Position struct {
	Line int
	Col  int
}

// Range represents a span of text between two positions (inclusive start, exclusive end).
type Range struct {
	Start Position
	End   Position
}

// TextStore is the abstract interface for all text storage backends.
// Sprint 1 uses LineStore ([]string). Future sprints can swap to rope or piece table
// without touching any consumer code.
//
// Column positions (Col in Position) are byte offsets within a line's UTF-8 string.
// This is consistent with Go's string indexing and slice operations.
type TextStore interface {
	// Insert inserts text at the given position.
	// If text contains newlines, it splits lines accordingly.
	Insert(pos Position, text string) error

	// Delete removes text in the given range.
	Delete(r Range) error

	// Line returns the content of line n (zero-based). Returns "" for out of range.
	Line(n int) string

	// LineCount returns the total number of lines.
	LineCount() int

	// Slice returns the text within the given range.
	Slice(r Range) string

	// Content returns the full document as a single string.
	Content() string
}

// ErrOutOfRange is returned when a position or range is outside the document bounds.
var ErrOutOfRange = fmt.Errorf("position out of range")
