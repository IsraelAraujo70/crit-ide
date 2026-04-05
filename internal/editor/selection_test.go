package editor

import "testing"

func TestSelection_Normalized_Forward(t *testing.T) {
	s := Selection{
		Anchor: Position{0, 2},
		Cursor: Position{1, 3},
	}
	r := s.Normalized()
	if r.Start != s.Anchor || r.End != s.Cursor {
		t.Fatalf("forward selection: expected (%v,%v), got (%v,%v)",
			s.Anchor, s.Cursor, r.Start, r.End)
	}
}

func TestSelection_Normalized_Backward(t *testing.T) {
	s := Selection{
		Anchor: Position{1, 3},
		Cursor: Position{0, 2},
	}
	r := s.Normalized()
	if r.Start != s.Cursor || r.End != s.Anchor {
		t.Fatalf("backward selection: expected (%v,%v), got (%v,%v)",
			s.Cursor, s.Anchor, r.Start, r.End)
	}
}

func TestSelection_Normalized_SameLine(t *testing.T) {
	s := Selection{
		Anchor: Position{0, 5},
		Cursor: Position{0, 2},
	}
	r := s.Normalized()
	if r.Start.Col != 2 || r.End.Col != 5 {
		t.Fatalf("expected cols (2,5), got (%d,%d)", r.Start.Col, r.End.Col)
	}
}

func TestSelection_IsEmpty(t *testing.T) {
	s := Selection{Anchor: Position{1, 2}, Cursor: Position{1, 2}}
	if !s.IsEmpty() {
		t.Fatal("expected empty selection")
	}
	s2 := Selection{Anchor: Position{0, 0}, Cursor: Position{0, 1}}
	if s2.IsEmpty() {
		t.Fatal("expected non-empty selection")
	}
}

func TestBuffer_SetSelection_ClearSelection(t *testing.T) {
	b := NewBuffer("t")
	for _, ch := range "hello" {
		b.InsertChar(ch)
	}

	b.SetSelection(Position{0, 0}, Position{0, 3})
	if !b.HasSelection() {
		t.Fatal("expected selection to be active")
	}

	b.ClearSelection()
	if b.HasSelection() {
		t.Fatal("expected selection to be cleared")
	}
}

func TestBuffer_HasSelection_NilAndEmpty(t *testing.T) {
	b := NewBuffer("t")
	b.InsertChar('a')

	// nil selection
	if b.HasSelection() {
		t.Fatal("nil selection should return false")
	}

	// empty selection (anchor == cursor)
	b.SetSelection(Position{0, 0}, Position{0, 0})
	if b.HasSelection() {
		t.Fatal("empty selection should return false")
	}
}

func TestBuffer_SelectedText_SingleLine(t *testing.T) {
	b := NewBuffer("t")
	for _, ch := range "hello world" {
		b.InsertChar(ch)
	}

	b.SetSelection(Position{0, 0}, Position{0, 5})
	got := b.SelectedText()
	if got != "hello" {
		t.Fatalf("expected %q, got %q", "hello", got)
	}
}

func TestBuffer_SelectedText_MultiLine(t *testing.T) {
	b := NewBuffer("t")
	for _, ch := range "hello" {
		b.InsertChar(ch)
	}
	b.InsertNewline()
	for _, ch := range "world" {
		b.InsertChar(ch)
	}

	b.SetSelection(Position{0, 3}, Position{1, 2})
	got := b.SelectedText()
	if got != "lo\nwo" {
		t.Fatalf("expected %q, got %q", "lo\nwo", got)
	}
}

func TestBuffer_DeleteSelection(t *testing.T) {
	b := NewBuffer("t")
	for _, ch := range "hello world" {
		b.InsertChar(ch)
	}

	b.SetSelection(Position{0, 5}, Position{0, 11})
	b.DeleteSelection()

	if b.Text.Line(0) != "hello" {
		t.Fatalf("expected %q, got %q", "hello", b.Text.Line(0))
	}
	if b.CursorRow != 0 || b.CursorCol != 5 {
		t.Fatalf("expected cursor at (0,5), got (%d,%d)", b.CursorRow, b.CursorCol)
	}
	if b.HasSelection() {
		t.Fatal("selection should be cleared after delete")
	}
	if !b.Dirty {
		t.Fatal("buffer should be dirty after delete selection")
	}
}

func TestBuffer_DeleteSelection_MultiLine(t *testing.T) {
	b := NewBuffer("t")
	for _, ch := range "hello" {
		b.InsertChar(ch)
	}
	b.InsertNewline()
	for _, ch := range "world" {
		b.InsertChar(ch)
	}

	b.SetSelection(Position{0, 3}, Position{1, 2})
	b.DeleteSelection()

	if b.Text.LineCount() != 1 {
		t.Fatalf("expected 1 line, got %d", b.Text.LineCount())
	}
	if b.Text.Line(0) != "helrld" {
		t.Fatalf("expected %q, got %q", "helrld", b.Text.Line(0))
	}
	if b.CursorRow != 0 || b.CursorCol != 3 {
		t.Fatalf("expected cursor at (0,3), got (%d,%d)", b.CursorRow, b.CursorCol)
	}
}

func TestBuffer_ReplaceSelection(t *testing.T) {
	b := NewBuffer("t")
	for _, ch := range "hello world" {
		b.InsertChar(ch)
	}

	b.SetSelection(Position{0, 0}, Position{0, 5})
	b.ReplaceSelection("hi")

	if b.Text.Line(0) != "hi world" {
		t.Fatalf("expected %q, got %q", "hi world", b.Text.Line(0))
	}
	if b.CursorCol != 2 {
		t.Fatalf("expected cursor col 2, got %d", b.CursorCol)
	}
}

func TestBuffer_InsertCharWithSelection(t *testing.T) {
	b := NewBuffer("t")
	for _, ch := range "abcd" {
		b.InsertChar(ch)
	}

	b.SetSelection(Position{0, 1}, Position{0, 3})
	b.InsertChar('X')

	if b.Text.Line(0) != "aXd" {
		t.Fatalf("expected %q, got %q", "aXd", b.Text.Line(0))
	}
}

func TestBuffer_InsertNewlineWithSelection(t *testing.T) {
	b := NewBuffer("t")
	for _, ch := range "abcd" {
		b.InsertChar(ch)
	}

	b.SetSelection(Position{0, 1}, Position{0, 3})
	b.InsertNewline()

	if b.Text.LineCount() != 2 {
		t.Fatalf("expected 2 lines, got %d", b.Text.LineCount())
	}
	if b.Text.Line(0) != "a" {
		t.Fatalf("line 0: expected %q, got %q", "a", b.Text.Line(0))
	}
	if b.Text.Line(1) != "d" {
		t.Fatalf("line 1: expected %q, got %q", "d", b.Text.Line(1))
	}
}

func TestBuffer_DeleteBackwardWithSelection(t *testing.T) {
	b := NewBuffer("t")
	for _, ch := range "abcd" {
		b.InsertChar(ch)
	}

	b.SetSelection(Position{0, 1}, Position{0, 3})
	b.DeleteBackward()

	if b.Text.Line(0) != "ad" {
		t.Fatalf("expected %q, got %q", "ad", b.Text.Line(0))
	}
	if b.CursorCol != 1 {
		t.Fatalf("expected col 1, got %d", b.CursorCol)
	}
}

func TestBuffer_DeleteForwardWithSelection(t *testing.T) {
	b := NewBuffer("t")
	for _, ch := range "abcd" {
		b.InsertChar(ch)
	}

	b.SetSelection(Position{0, 1}, Position{0, 3})
	b.DeleteForward()

	if b.Text.Line(0) != "ad" {
		t.Fatalf("expected %q, got %q", "ad", b.Text.Line(0))
	}
	if b.CursorCol != 1 {
		t.Fatalf("expected col 1, got %d", b.CursorCol)
	}
}

func TestBuffer_SelectAll(t *testing.T) {
	b := NewBuffer("t")
	for _, ch := range "hello" {
		b.InsertChar(ch)
	}
	b.InsertNewline()
	for _, ch := range "world" {
		b.InsertChar(ch)
	}

	b.SelectAll()

	if !b.HasSelection() {
		t.Fatal("expected selection after SelectAll")
	}
	r := b.Selection.Normalized()
	if r.Start.Line != 0 || r.Start.Col != 0 {
		t.Fatalf("expected start (0,0), got (%d,%d)", r.Start.Line, r.Start.Col)
	}
	if r.End.Line != 1 || r.End.Col != 5 {
		t.Fatalf("expected end (1,5), got (%d,%d)", r.End.Line, r.End.Col)
	}

	got := b.SelectedText()
	if got != "hello\nworld" {
		t.Fatalf("expected %q, got %q", "hello\nworld", got)
	}
}

func TestBuffer_DeleteSelection_NoSelection(t *testing.T) {
	b := NewBuffer("t")
	for _, ch := range "hello" {
		b.InsertChar(ch)
	}
	b.Dirty = false // Reset after insertions.

	b.DeleteSelection() // no-op

	if b.Text.Line(0) != "hello" {
		t.Fatalf("expected %q, got %q", "hello", b.Text.Line(0))
	}
	if b.Dirty {
		t.Fatal("dirty flag should not be set when no selection to delete")
	}
}
