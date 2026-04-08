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
	Undo       *UndoStack // Undo/redo history.
}

// NewBuffer creates a new empty scratch buffer.
func NewBuffer(id BufferID) *Buffer {
	return &Buffer{
		ID:   id,
		Kind: BufferKindScratch,
		Text: NewLineStore(""),
		Undo: NewUndoStack(1000),
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
		Undo: NewUndoStack(1000),
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
	b.Undo.Push(UndoEntry{
		Kind:      EditInsert,
		Pos:       Position{b.CursorRow, b.CursorCol},
		Text:      s,
		CursorRow: b.CursorRow,
		CursorCol: b.CursorCol,
	})
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
// Auto-indents by copying leading whitespace from the current line.
func (b *Buffer) InsertNewline() {
	if b.ReadOnly {
		return
	}
	if b.HasSelection() {
		b.ReplaceSelection("\n")
		return
	}
	// Compute auto-indent: leading whitespace of the current line.
	currentLine := b.Text.Line(b.CursorRow)
	indent := ""
	for _, ch := range currentLine {
		if ch == ' ' || ch == '\t' {
			indent += string(ch)
		} else {
			break
		}
	}
	insertText := "\n" + indent
	b.Undo.Push(UndoEntry{
		Kind:      EditInsert,
		Pos:       Position{b.CursorRow, b.CursorCol},
		Text:      insertText,
		CursorRow: b.CursorRow,
		CursorCol: b.CursorCol,
	})
	err := b.Text.Insert(Position{b.CursorRow, b.CursorCol}, insertText)
	if err != nil {
		return
	}
	b.CursorRow++
	b.CursorCol = len(indent)
	b.desiredCol = b.CursorCol
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
		delRange := Range{
			Start: Position{b.CursorRow, b.CursorCol - size},
			End:   Position{b.CursorRow, b.CursorCol},
		}
		deleted := b.Text.Slice(delRange)
		b.Undo.Push(UndoEntry{
			Kind:      EditDelete,
			Pos:       delRange.Start,
			Text:      deleted,
			CursorRow: b.CursorRow,
			CursorCol: b.CursorCol,
		})
		err := b.Text.Delete(delRange)
		if err != nil {
			return
		}
		b.CursorCol -= size
		b.desiredCol = b.CursorCol
		b.Dirty = true
	} else if b.CursorRow > 0 {
		// At the start of a line: merge with the previous line.
		prevLineLen := len(b.Text.Line(b.CursorRow - 1))
		delRange := Range{
			Start: Position{b.CursorRow - 1, prevLineLen},
			End:   Position{b.CursorRow, 0},
		}
		deleted := b.Text.Slice(delRange)
		b.Undo.Push(UndoEntry{
			Kind:      EditDelete,
			Pos:       delRange.Start,
			Text:      deleted,
			CursorRow: b.CursorRow,
			CursorCol: b.CursorCol,
		})
		err := b.Text.Delete(delRange)
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
		delRange := Range{
			Start: Position{b.CursorRow, b.CursorCol},
			End:   Position{b.CursorRow, b.CursorCol + size},
		}
		deleted := b.Text.Slice(delRange)
		b.Undo.Push(UndoEntry{
			Kind:      EditDelete,
			Pos:       delRange.Start,
			Text:      deleted,
			CursorRow: b.CursorRow,
			CursorCol: b.CursorCol,
		})
		err := b.Text.Delete(delRange)
		if err != nil {
			return
		}
		b.Dirty = true
	} else if b.CursorRow < b.Text.LineCount()-1 {
		// At end of line: merge with the next line.
		delRange := Range{
			Start: Position{b.CursorRow, len(line)},
			End:   Position{b.CursorRow + 1, 0},
		}
		deleted := b.Text.Slice(delRange)
		b.Undo.Push(UndoEntry{
			Kind:      EditDelete,
			Pos:       delRange.Start,
			Text:      deleted,
			CursorRow: b.CursorRow,
			CursorCol: b.CursorCol,
		})
		err := b.Text.Delete(delRange)
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
	deleted := b.Text.Slice(r)
	b.Undo.Push(UndoEntry{
		Kind:      EditDelete,
		Pos:       r.Start,
		Text:      deleted,
		CursorRow: b.CursorRow,
		CursorCol: b.CursorCol,
	})
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
	b.Undo.Push(UndoEntry{
		Kind:      EditInsert,
		Pos:       Position{b.CursorRow, b.CursorCol},
		Text:      text,
		CursorRow: b.CursorRow,
		CursorCol: b.CursorCol,
	})
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

// UndoEdit reverses the most recent edit operation.
func (b *Buffer) UndoEdit() {
	entry, ok := b.Undo.PopUndo()
	if !ok {
		return
	}
	b.ClearSelection()
	switch entry.Kind {
	case EditInsert:
		// To undo an insert, delete the inserted text.
		endPos := b.positionAfterInsert(entry.Pos, entry.Text)
		_ = b.Text.Delete(Range{Start: entry.Pos, End: endPos})
	case EditDelete:
		// To undo a delete, re-insert the deleted text.
		_ = b.Text.Insert(entry.Pos, entry.Text)
	}
	// Restore cursor position.
	b.CursorRow = entry.CursorRow
	b.CursorCol = entry.CursorCol
	b.desiredCol = b.CursorCol
	b.ClampCursor()
	b.Dirty = true
	b.Undo.PushRedo(entry)
}

// RedoEdit re-applies the most recently undone edit.
func (b *Buffer) RedoEdit() {
	entry, ok := b.Undo.PopRedo()
	if !ok {
		return
	}
	b.ClearSelection()
	switch entry.Kind {
	case EditInsert:
		// Re-apply the insert.
		_ = b.Text.Insert(entry.Pos, entry.Text)
		// Move cursor to end of inserted text.
		endPos := b.positionAfterInsert(entry.Pos, entry.Text)
		b.CursorRow = endPos.Line
		b.CursorCol = endPos.Col
	case EditDelete:
		// Re-apply the delete.
		endPos := b.positionAfterInsert(entry.Pos, entry.Text)
		_ = b.Text.Delete(Range{Start: entry.Pos, End: endPos})
		b.CursorRow = entry.Pos.Line
		b.CursorCol = entry.Pos.Col
	}
	b.desiredCol = b.CursorCol
	b.ClampCursor()
	b.Dirty = true
	// Push back onto undo stack (without clearing redo — use internal push).
	b.Undo.undos = append(b.Undo.undos, entry)
}

// positionAfterInsert computes the end position after inserting text at pos.
func (b *Buffer) positionAfterInsert(pos Position, text string) Position {
	row := pos.Line
	col := pos.Col
	for _, ch := range text {
		if ch == '\n' {
			row++
			col = 0
		} else {
			col += len(string(ch))
		}
	}
	return Position{row, col}
}

// WordLeft moves the cursor to the beginning of the previous word.
func (b *Buffer) WordLeft() {
	line := b.Text.Line(b.CursorRow)
	if b.CursorCol == 0 {
		// Move to end of previous line.
		if b.CursorRow > 0 {
			b.CursorRow--
			b.CursorCol = len(b.Text.Line(b.CursorRow))
		}
		b.desiredCol = b.CursorCol
		return
	}
	// Work backwards from cursor through the line bytes.
	col := b.CursorCol
	// Skip whitespace first.
	for col > 0 {
		_, size := utf8.DecodeLastRuneInString(line[:col])
		if size == 0 {
			break
		}
		r, _ := utf8.DecodeRuneInString(line[col-size:])
		if !isWordSeparator(r) {
			break
		}
		col -= size
	}
	// Skip word characters.
	for col > 0 {
		_, size := utf8.DecodeLastRuneInString(line[:col])
		if size == 0 {
			break
		}
		r, _ := utf8.DecodeRuneInString(line[col-size:])
		if isWordSeparator(r) {
			break
		}
		col -= size
	}
	b.CursorCol = col
	b.desiredCol = b.CursorCol
}

// WordRight moves the cursor to the beginning of the next word.
func (b *Buffer) WordRight() {
	line := b.Text.Line(b.CursorRow)
	if b.CursorCol >= len(line) {
		// Move to start of next line.
		if b.CursorRow < b.Text.LineCount()-1 {
			b.CursorRow++
			b.CursorCol = 0
		}
		b.desiredCol = b.CursorCol
		return
	}
	col := b.CursorCol
	// Skip word characters first.
	for col < len(line) {
		r, size := utf8.DecodeRuneInString(line[col:])
		if isWordSeparator(r) {
			break
		}
		col += size
	}
	// Skip whitespace/separators.
	for col < len(line) {
		r, size := utf8.DecodeRuneInString(line[col:])
		if !isWordSeparator(r) {
			break
		}
		col += size
	}
	b.CursorCol = col
	b.desiredCol = b.CursorCol
}

// isWordSeparator returns true if the rune is a whitespace or punctuation character.
func isWordSeparator(r rune) bool {
	if r == '_' {
		return false // Treat underscore as part of a word (identifiers).
	}
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' ||
		r == '.' || r == ',' || r == ';' || r == ':' ||
		r == '(' || r == ')' || r == '[' || r == ']' ||
		r == '{' || r == '}' || r == '<' || r == '>' ||
		r == '"' || r == '\'' || r == '`' ||
		r == '+' || r == '-' || r == '*' || r == '/' ||
		r == '=' || r == '!' || r == '&' || r == '|' ||
		r == '~' || r == '^' || r == '%' || r == '#' ||
		r == '@' || r == '?' || r == '\\'
}

// DuplicateLine duplicates the current line below.
func (b *Buffer) DuplicateLine() {
	if b.ReadOnly {
		return
	}
	line := b.Text.Line(b.CursorRow)
	insertPos := Position{b.CursorRow, len(line)}
	insertText := "\n" + line
	b.Undo.Push(UndoEntry{
		Kind:      EditInsert,
		Pos:       insertPos,
		Text:      insertText,
		CursorRow: b.CursorRow,
		CursorCol: b.CursorCol,
	})
	_ = b.Text.Insert(insertPos, insertText)
	b.CursorRow++
	b.Dirty = true
}

// SelectWord selects the word under or near the cursor.
func (b *Buffer) SelectWord() {
	line := b.Text.Line(b.CursorRow)
	if len(line) == 0 {
		return
	}
	col := b.CursorCol
	if col >= len(line) {
		col = len(line) - 1
		if col < 0 {
			return
		}
	}
	// Find word boundaries.
	start := col
	for start > 0 {
		_, size := utf8.DecodeLastRuneInString(line[:start])
		r, _ := utf8.DecodeRuneInString(line[start-size:])
		if isWordSeparator(r) {
			break
		}
		start -= size
	}
	end := col
	for end < len(line) {
		r, size := utf8.DecodeRuneInString(line[end:])
		if isWordSeparator(r) {
			break
		}
		end += size
	}
	if start != end {
		b.SetSelection(
			Position{b.CursorRow, start},
			Position{b.CursorRow, end},
		)
		b.CursorCol = end
		b.desiredCol = b.CursorCol
	}
}

// IndentSelection adds a tab character at the beginning of each line in the
// current selection. The operation is recorded as a single undo entry (one
// delete of the original block + one insert of the indented block).
// The selection is preserved and adjusted to cover the indented text.
func (b *Buffer) IndentSelection() {
	if b.ReadOnly || !b.HasSelection() {
		return
	}
	r := b.Selection.Normalized()
	startLine := r.Start.Line
	endLine := r.End.Line
	// If selection ends at column 0 of a line, don't indent that line
	// (the cursor is at the very start, no content selected on that line).
	if r.End.Col == 0 && endLine > startLine {
		endLine--
	}

	// Capture original block text (full lines from startLine to endLine).
	oldStart := Position{startLine, 0}
	lastLineLen := len(b.Text.Line(endLine))
	oldEnd := Position{endLine, lastLineLen}
	oldText := b.Text.Slice(Range{Start: oldStart, End: oldEnd})

	// Build indented version.
	var newLines []string
	for i := startLine; i <= endLine; i++ {
		newLines = append(newLines, "\t"+b.Text.Line(i))
	}
	newText := strings.Join(newLines, "\n")

	// Record undo: delete old block, insert new block.
	// We push two entries so that undo reverses both in order.
	// Push insert first (will be undone last), then delete (undone first).
	b.Undo.Push(UndoEntry{
		Kind:      EditDelete,
		Pos:       oldStart,
		Text:      oldText,
		CursorRow: b.CursorRow,
		CursorCol: b.CursorCol,
	})

	// Replace old text with new text.
	_ = b.Text.Delete(Range{Start: oldStart, End: oldEnd})
	_ = b.Text.Insert(oldStart, newText)

	b.Undo.undos[len(b.Undo.undos)-1] = UndoEntry{
		Kind:      EditDelete,
		Pos:       oldStart,
		Text:      oldText,
		CursorRow: b.CursorRow,
		CursorCol: b.CursorCol,
	}
	// Push the insert entry on top so undo pops it first.
	b.Undo.undos = append(b.Undo.undos, UndoEntry{
		Kind:      EditInsert,
		Pos:       oldStart,
		Text:      newText,
		CursorRow: b.CursorRow,
		CursorCol: b.CursorCol,
	})
	// Clear redo since we manually appended.
	b.Undo.redos = b.Undo.redos[:0]

	// Adjust selection: each line got 1 byte ("\t") prepended.
	newAnchor := b.Selection.Anchor
	newCursor := b.Selection.Cursor
	if newAnchor.Line >= startLine && newAnchor.Line <= endLine {
		newAnchor.Col += 1
	}
	if newCursor.Line >= startLine && newCursor.Line <= endLine {
		newCursor.Col += 1
	}
	b.SetSelection(newAnchor, newCursor)
	b.CursorRow = newCursor.Line
	b.CursorCol = newCursor.Col
	b.desiredCol = b.CursorCol
	b.Dirty = true
}

// DedentSelection removes one level of indentation (one tab or up to 4 spaces)
// from the beginning of each line in the current selection.
// The operation is recorded as a single undo entry.
// The selection is preserved and adjusted.
func (b *Buffer) DedentSelection() {
	if b.ReadOnly || !b.HasSelection() {
		return
	}
	r := b.Selection.Normalized()
	startLine := r.Start.Line
	endLine := r.End.Line
	if r.End.Col == 0 && endLine > startLine {
		endLine--
	}

	// Capture original block text.
	oldStart := Position{startLine, 0}
	lastLineLen := len(b.Text.Line(endLine))
	oldEnd := Position{endLine, lastLineLen}
	oldText := b.Text.Slice(Range{Start: oldStart, End: oldEnd})

	// Build dedented version and track how many bytes removed per line.
	removed := make([]int, endLine-startLine+1)
	var newLines []string
	for i := startLine; i <= endLine; i++ {
		line := b.Text.Line(i)
		idx := i - startLine
		if len(line) > 0 && line[0] == '\t' {
			removed[idx] = 1
			newLines = append(newLines, line[1:])
		} else {
			// Remove up to 4 leading spaces.
			spaces := 0
			for spaces < 4 && spaces < len(line) && line[spaces] == ' ' {
				spaces++
			}
			removed[idx] = spaces
			newLines = append(newLines, line[spaces:])
		}
	}

	// Check if any line was actually dedented.
	anyRemoved := false
	for _, r := range removed {
		if r > 0 {
			anyRemoved = true
			break
		}
	}
	if !anyRemoved {
		return
	}

	newText := strings.Join(newLines, "\n")

	// Record undo entries (same pattern as IndentSelection).
	b.Undo.Push(UndoEntry{
		Kind:      EditDelete,
		Pos:       oldStart,
		Text:      oldText,
		CursorRow: b.CursorRow,
		CursorCol: b.CursorCol,
	})
	_ = b.Text.Delete(Range{Start: oldStart, End: oldEnd})
	_ = b.Text.Insert(oldStart, newText)

	b.Undo.undos[len(b.Undo.undos)-1] = UndoEntry{
		Kind:      EditDelete,
		Pos:       oldStart,
		Text:      oldText,
		CursorRow: b.CursorRow,
		CursorCol: b.CursorCol,
	}
	b.Undo.undos = append(b.Undo.undos, UndoEntry{
		Kind:      EditInsert,
		Pos:       oldStart,
		Text:      newText,
		CursorRow: b.CursorRow,
		CursorCol: b.CursorCol,
	})
	b.Undo.redos = b.Undo.redos[:0]

	// Adjust selection columns.
	newAnchor := b.Selection.Anchor
	newCursor := b.Selection.Cursor
	if newAnchor.Line >= startLine && newAnchor.Line <= endLine {
		rem := removed[newAnchor.Line-startLine]
		newAnchor.Col -= rem
		if newAnchor.Col < 0 {
			newAnchor.Col = 0
		}
	}
	if newCursor.Line >= startLine && newCursor.Line <= endLine {
		rem := removed[newCursor.Line-startLine]
		newCursor.Col -= rem
		if newCursor.Col < 0 {
			newCursor.Col = 0
		}
	}
	b.SetSelection(newAnchor, newCursor)
	b.CursorRow = newCursor.Line
	b.CursorCol = newCursor.Col
	b.desiredCol = b.CursorCol
	b.Dirty = true
}

// FileName returns the base name of the file, or "[scratch]" for scratch buffers.
func (b *Buffer) FileName() string {
	if b.Path == "" {
		return "[scratch]"
	}
	return filepath.Base(b.Path)
}
