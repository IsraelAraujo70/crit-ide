package editor

import "testing"

func TestUndoInsertChar(t *testing.T) {
	buf := NewBuffer("test")
	buf.InsertChar('H')
	buf.InsertChar('i')

	if buf.Text.Content() != "Hi" {
		t.Fatalf("expected 'Hi', got %q", buf.Text.Content())
	}

	buf.UndoEdit() // Undo 'i'
	if buf.Text.Content() != "H" {
		t.Fatalf("after first undo expected 'H', got %q", buf.Text.Content())
	}

	buf.UndoEdit() // Undo 'H'
	if buf.Text.Content() != "" {
		t.Fatalf("after second undo expected '', got %q", buf.Text.Content())
	}
}

func TestRedoInsertChar(t *testing.T) {
	buf := NewBuffer("test")
	buf.InsertChar('A')
	buf.InsertChar('B')
	buf.UndoEdit() // Undo 'B'

	if buf.Text.Content() != "A" {
		t.Fatalf("expected 'A', got %q", buf.Text.Content())
	}

	buf.RedoEdit() // Redo 'B'
	if buf.Text.Content() != "AB" {
		t.Fatalf("after redo expected 'AB', got %q", buf.Text.Content())
	}
}

func TestUndoDeleteBackward(t *testing.T) {
	buf := NewBuffer("test")
	buf.InsertChar('A')
	buf.InsertChar('B')
	buf.InsertChar('C')
	buf.DeleteBackward() // Delete 'C'

	if buf.Text.Content() != "AB" {
		t.Fatalf("expected 'AB', got %q", buf.Text.Content())
	}

	buf.UndoEdit() // Undo the delete → restore 'C'
	if buf.Text.Content() != "ABC" {
		t.Fatalf("after undo expected 'ABC', got %q", buf.Text.Content())
	}
}

func TestUndoNewline(t *testing.T) {
	buf := NewBuffer("test")
	buf.InsertChar('A')
	buf.InsertNewline()
	buf.InsertChar('B')

	if buf.Text.LineCount() != 2 {
		t.Fatalf("expected 2 lines, got %d", buf.Text.LineCount())
	}

	buf.UndoEdit() // Undo 'B'
	buf.UndoEdit() // Undo newline
	if buf.Text.LineCount() != 1 {
		t.Fatalf("after undo expected 1 line, got %d", buf.Text.LineCount())
	}
	if buf.Text.Content() != "A" {
		t.Fatalf("after undo expected 'A', got %q", buf.Text.Content())
	}
}

func TestUndoNewEditClearsRedo(t *testing.T) {
	buf := NewBuffer("test")
	buf.InsertChar('A')
	buf.InsertChar('B')
	buf.UndoEdit()

	// Now insert a different character — should clear redo.
	buf.InsertChar('C')
	if !buf.Undo.CanUndo() {
		t.Fatal("should be able to undo")
	}
	if buf.Undo.CanRedo() {
		t.Fatal("redo should be cleared after new edit")
	}
}

func TestWordLeftRight(t *testing.T) {
	buf := NewBuffer("test")
	// Set content to "hello world"
	for _, ch := range "hello world" {
		buf.InsertChar(ch)
	}
	// Cursor is at end: col 11.
	if buf.CursorCol != 11 {
		t.Fatalf("expected cursor at 11, got %d", buf.CursorCol)
	}

	buf.WordLeft() // Should jump to start of "world" (col 6).
	if buf.CursorCol != 6 {
		t.Fatalf("after WordLeft expected col 6, got %d", buf.CursorCol)
	}

	buf.WordLeft() // Should jump to start of "hello" (col 0).
	if buf.CursorCol != 0 {
		t.Fatalf("after second WordLeft expected col 0, got %d", buf.CursorCol)
	}

	buf.WordRight() // Should jump to start of "world" (col 6).
	if buf.CursorCol != 6 {
		t.Fatalf("after WordRight expected col 6, got %d", buf.CursorCol)
	}
}

func TestDuplicateLine(t *testing.T) {
	buf := NewBuffer("test")
	for _, ch := range "hello" {
		buf.InsertChar(ch)
	}
	buf.DuplicateLine()

	if buf.Text.LineCount() != 2 {
		t.Fatalf("expected 2 lines, got %d", buf.Text.LineCount())
	}
	if buf.Text.Line(0) != "hello" || buf.Text.Line(1) != "hello" {
		t.Fatalf("expected both lines 'hello', got %q and %q", buf.Text.Line(0), buf.Text.Line(1))
	}
	if buf.CursorRow != 1 {
		t.Fatalf("expected cursor on row 1, got %d", buf.CursorRow)
	}
}

func TestSelectWord(t *testing.T) {
	buf := NewBuffer("test")
	for _, ch := range "hello world" {
		buf.InsertChar(ch)
	}
	buf.CursorCol = 2 // Inside "hello"
	buf.SelectWord()

	if !buf.HasSelection() {
		t.Fatal("expected selection after SelectWord")
	}
	if buf.SelectedText() != "hello" {
		t.Fatalf("expected 'hello' selected, got %q", buf.SelectedText())
	}
}

func TestAutoIndent(t *testing.T) {
	buf := NewBuffer("test")
	// Type "    hello" (4 spaces + hello)
	for _, ch := range "    hello" {
		buf.InsertChar(ch)
	}
	buf.InsertNewline()

	// New line should have the same 4-space indent.
	if buf.CursorRow != 1 {
		t.Fatalf("expected cursor on row 1, got %d", buf.CursorRow)
	}
	if buf.CursorCol != 4 {
		t.Fatalf("expected cursor at col 4 (indent), got %d", buf.CursorCol)
	}
	if buf.Text.Line(1) != "    " {
		t.Fatalf("expected line 1 to have indent, got %q", buf.Text.Line(1))
	}
}
