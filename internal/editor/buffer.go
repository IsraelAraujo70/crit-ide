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
	desiredCol int // Sticky column for Up/Down movement (byte offset).
	LanguageID string // Language identifier for highlighting and LSP (e.g., "go", "python").
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
// CursorCol advances by the UTF-8 byte length of the rune.
func (b *Buffer) InsertChar(ch rune) {
	if b.ReadOnly {
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
func (b *Buffer) InsertNewline() {
	if b.ReadOnly {
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
func (b *Buffer) DeleteBackward() {
	if b.ReadOnly {
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
func (b *Buffer) DeleteForward() {
	if b.ReadOnly {
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

// FileName returns the base name of the file, or "[scratch]" for scratch buffers.
func (b *Buffer) FileName() string {
	if b.Path == "" {
		return "[scratch]"
	}
	return filepath.Base(b.Path)
}
