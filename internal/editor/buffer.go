package editor

import (
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// BufferID uniquely identifies a buffer.
type BufferID string

// BufferKind classifies the buffer type.
type BufferKind int

const (
	BufferKindFile    BufferKind = iota // Backed by a file on disk.
	BufferKindScratch                   // Ephemeral, not backed by a file.
)

// CursorDir represents a direction for cursor movement.
type CursorDir int

const (
	DirUp CursorDir = iota
	DirDown
	DirLeft
	DirRight
)

// Buffer represents an open document with cursor state.
// CursorCol is a byte offset within the current line (consistent with TextStore).
type Buffer struct {
	ID         BufferID
	Path       string
	Kind       BufferKind
	Text       TextStore
	Dirty      bool
	ReadOnly   bool
	CursorRow  int
	CursorCol  int
	Selection  *Selection // Active text selection, nil when no selection.
	LanguageID string     // Language identifier for syntax highlighting and LSP.
	desiredCol int        // Sticky column for Up/Down movement (byte offset).
}

// NewBuffer creates a new empty scratch buffer.
func NewBuffer(id BufferID) *Buffer {
	return &Buffer{
		ID:   id,
		Kind: BufferKindScratch,
		Text: NewLineStore(""),
	}
}

// LoadFile reads a file from disk into a new buffer.
func LoadFile(id BufferID, path string) (*Buffer, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	content := string(data)
	// Normalize line endings to \n for internal representation.
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	// Remove trailing newline to avoid a phantom empty last line.
	content = strings.TrimSuffix(content, "\n")

	return &Buffer{
		ID:   id,
		Path: absPath,
		Kind: BufferKindFile,
		Text: NewLineStore(content),
	}, nil
}

// SaveFile writes the buffer content to its file path.
func (b *Buffer) SaveFile() error {
	if b.Path == "" {
		return nil // Scratch buffers can't be saved without a path.
	}
	content := b.Text.Content() + "\n" // Ensure trailing newline on disk.
	err := os.WriteFile(b.Path, []byte(content), 0644)
	if err != nil {
		return err
	}
	b.Dirty = false
	return nil
}

// InsertChar inserts a single character at the cursor position.
// If text is selected, it replaces the selection.
// CursorCol advances by the UTF-8 byte length of the rune.
func (b *Buffer) InsertChar(ch rune) {
	if b.ReadOnly {
		return
	}
	if b.HasSelection() {
		b.ReplaceSelection(string(ch))
		return
	}
	s := string(ch)
	err := b.Text.Insert(Position{b.CursorRow, b.CursorCol}, s)
	if err != nil {
		return
	}
	b.CursorCol += len(s)
	b.desiredCol = b.CursorCol
	b.Dirty = true
}

// InsertNewline splits the current line at the cursor position.
// If text is selected, it replaces the selection.
func (b *Buffer) InsertNewline() {
	if b.ReadOnly {
		return
	}
	if b.HasSelection() {
		b.ReplaceSelection("\n")
		return
	}
	err := b.Text.Insert(Position{b.CursorRow, b.CursorCol}, "\n")
	if err != nil {
		return
	}
	b.CursorRow++
	b.CursorCol = 0
	b.desiredCol = 0
	b.Dirty = true
}

// DeleteBackward removes the character before the cursor (backspace).
// If text is selected, it deletes the selection instead.
func (b *Buffer) DeleteBackward() {
	if b.ReadOnly {
		return
	}
	if b.HasSelection() {
		b.DeleteSelection()
		return
	}
	if b.CursorCol > 0 {
		// Find the start of the previous rune.
		line := b.Text.Line(b.CursorRow)
		_, size := utf8.DecodeLastRuneInString(line[:b.CursorCol])
		if size == 0 {
			size = 1
		}
		err := b.Text.Delete(Range{
			Start: Position{b.CursorRow, b.CursorCol - size},
			End:   Position{b.CursorRow, b.CursorCol},
		})
		if err != nil {
			return
		}
		b.CursorCol -= size
		b.desiredCol = b.CursorCol
		b.Dirty = true
	} else if b.CursorRow > 0 {
		// At the start of a line: merge with the previous line.
		prevLineLen := len(b.Text.Line(b.CursorRow - 1))
		err := b.Text.Delete(Range{
			Start: Position{b.CursorRow - 1, prevLineLen},
			End:   Position{b.CursorRow, 0},
		})
		if err != nil {
			return
		}
		b.CursorRow--
		b.CursorCol = prevLineLen
		b.desiredCol = b.CursorCol
		b.Dirty = true
	}
}

// DeleteForward removes the character at the cursor position (delete key).
// If text is selected, it deletes the selection instead.
func (b *Buffer) DeleteForward() {
	if b.ReadOnly {
		return
	}
	if b.HasSelection() {
		b.DeleteSelection()
		return
	}
	line := b.Text.Line(b.CursorRow)
	if b.CursorCol < len(line) {
		// Find the size of the rune at cursor.
		_, size := utf8.DecodeRuneInString(line[b.CursorCol:])
		if size == 0 {
			size = 1
		}
		err := b.Text.Delete(Range{
			Start: Position{b.CursorRow, b.CursorCol},
			End:   Position{b.CursorRow, b.CursorCol + size},
		})
		if err != nil {
			return
		}
		b.Dirty = true
	} else if b.CursorRow < b.Text.LineCount()-1 {
		// At end of line: merge with the next line.
		err := b.Text.Delete(Range{
			Start: Position{b.CursorRow, len(line)},
			End:   Position{b.CursorRow + 1, 0},
		})
		if err != nil {
			return
		}
		b.Dirty = true
	}
}

// MoveCursor moves the cursor in the given direction.
func (b *Buffer) MoveCursor(dir CursorDir) {
	switch dir {
	case DirUp:
		if b.CursorRow > 0 {
			b.CursorRow--
			b.CursorCol = b.desiredCol
			b.ClampCursor()
		}
	case DirDown:
		if b.CursorRow < b.Text.LineCount()-1 {
			b.CursorRow++
			b.CursorCol = b.desiredCol
			b.ClampCursor()
		}
	case DirLeft:
		if b.CursorCol > 0 {
			// Move back one rune.
			line := b.Text.Line(b.CursorRow)
			_, size := utf8.DecodeLastRuneInString(line[:b.CursorCol])
			if size == 0 {
				size = 1
			}
			b.CursorCol -= size
		} else if b.CursorRow > 0 {
			b.CursorRow--
			b.CursorCol = len(b.Text.Line(b.CursorRow))
		}
		b.desiredCol = b.CursorCol
	case DirRight:
		line := b.Text.Line(b.CursorRow)
		if b.CursorCol < len(line) {
			// Move forward one rune.
			_, size := utf8.DecodeRuneInString(line[b.CursorCol:])
			if size == 0 {
				size = 1
			}
			b.CursorCol += size
		} else if b.CursorRow < b.Text.LineCount()-1 {
			b.CursorRow++
			b.CursorCol = 0
		}
		b.desiredCol = b.CursorCol
	}
}

// CursorHome moves the cursor to the start of the current line.
func (b *Buffer) CursorHome() {
	b.CursorCol = 0
	b.desiredCol = 0
}

// CursorEnd moves the cursor to the end of the current line.
func (b *Buffer) CursorEnd() {
	b.CursorCol = len(b.Text.Line(b.CursorRow))
	b.desiredCol = b.CursorCol
}

// ClampCursor ensures the cursor is within valid bounds.
func (b *Buffer) ClampCursor() {
	if b.CursorRow < 0 {
		b.CursorRow = 0
	}
	maxRow := b.Text.LineCount() - 1
	if maxRow < 0 {
		maxRow = 0
	}
	if b.CursorRow > maxRow {
		b.CursorRow = maxRow
	}
	lineLen := len(b.Text.Line(b.CursorRow))
	if b.CursorCol < 0 {
		b.CursorCol = 0
	}
	if b.CursorCol > lineLen {
		b.CursorCol = lineLen
	}
}

// SetSelection creates or updates the text selection.
func (b *Buffer) SetSelection(anchor, cursor Position) {
	b.Selection = &Selection{Anchor: anchor, Cursor: cursor}
}

// ClearSelection removes the active selection.
func (b *Buffer) ClearSelection() {
	b.Selection = nil
}

// HasSelection returns true if there is a non-empty selection.
func (b *Buffer) HasSelection() bool {
	return b.Selection != nil && !b.Selection.IsEmpty()
}

// SelectedText returns the text within the current selection, or "".
func (b *Buffer) SelectedText() string {
	if !b.HasSelection() {
		return ""
	}
	return b.Text.Slice(b.Selection.Normalized())
}

// DeleteSelection deletes the selected text, moves the cursor to the
// start of the deleted range, clears the selection, and marks dirty.
func (b *Buffer) DeleteSelection() {
	if !b.HasSelection() {
		return
	}
	r := b.Selection.Normalized()
	_ = b.Text.Delete(r)
	b.CursorRow = r.Start.Line
	b.CursorCol = r.Start.Col
	b.desiredCol = b.CursorCol
	b.ClearSelection()
	b.ClampCursor()
	b.Dirty = true
}

// ReplaceSelection deletes the selection and inserts text at the cursor.
func (b *Buffer) ReplaceSelection(text string) {
	b.DeleteSelection()
	_ = b.Text.Insert(Position{b.CursorRow, b.CursorCol}, text)
	// Advance cursor past inserted text.
	for _, ch := range text {
		if ch == '\n' {
			b.CursorRow++
			b.CursorCol = 0
		} else {
			b.CursorCol += len(string(ch))
		}
	}
	b.desiredCol = b.CursorCol
	b.Dirty = true
}

// SelectAll selects the entire buffer content.
func (b *Buffer) SelectAll() {
	lastLine := b.Text.LineCount() - 1
	if lastLine < 0 {
		lastLine = 0
	}
	lastCol := len(b.Text.Line(lastLine))
	b.SetSelection(
		Position{0, 0},
		Position{lastLine, lastCol},
	)
}

// SetCursorPos moves the cursor to the given row and byte-offset column,
// clamping to valid bounds. It also updates the sticky desiredCol.
func (b *Buffer) SetCursorPos(row, col int) {
	b.CursorRow = row
	b.CursorCol = col
	b.ClampCursor()
	b.desiredCol = b.CursorCol
}

// GutterWidth calculates the number of columns needed for line numbers.
// The formula is: max(3, digits(lineCount)) + 1 (for the space separator).
func GutterWidth(lineCount int) int {
	digits := 1
	n := lineCount
	for n >= 10 {
		digits++
		n /= 10
	}
	if digits < 3 {
		digits = 3
	}
	return digits + 1
}

// VisualColToByteOffset converts a visual column (accounting for tab expansion
// with 4-space tab stops) to a byte offset within the given line string.
// If visualCol falls beyond the end of the line, it returns len(line).
// If visualCol lands in the middle of a tab expansion, it snaps to the
// byte offset of that tab character.
func VisualColToByteOffset(line string, visualCol int) int {
	vcol := 0
	for i, ch := range line {
		if vcol >= visualCol {
			return i
		}
		if ch == '\t' {
			tabWidth := 4 - (vcol % 4)
			if vcol+tabWidth > visualCol {
				return i // Click lands inside tab expansion — snap to the tab.
			}
			vcol += tabWidth
		} else {
			vcol++
		}
	}
	return len(line)
}

// FileName returns the base name of the file, or "[scratch]" for scratch buffers.
func (b *Buffer) FileName() string {
	if b.Path == "" {
		return "[scratch]"
	}
	return filepath.Base(b.Path)
}
