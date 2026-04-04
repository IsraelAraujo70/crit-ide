package editor

import (
	"testing"
)

func TestNewLineStore_Empty(t *testing.T) {
	ls := NewLineStore("")
	if ls.LineCount() != 1 {
		t.Fatalf("expected 1 line, got %d", ls.LineCount())
	}
	if ls.Line(0) != "" {
		t.Fatalf("expected empty line, got %q", ls.Line(0))
	}
}

func TestNewLineStore_MultiLine(t *testing.T) {
	ls := NewLineStore("hello\nworld\nfoo")
	if ls.LineCount() != 3 {
		t.Fatalf("expected 3 lines, got %d", ls.LineCount())
	}
	if ls.Line(0) != "hello" {
		t.Fatalf("line 0: expected %q, got %q", "hello", ls.Line(0))
	}
	if ls.Line(1) != "world" {
		t.Fatalf("line 1: expected %q, got %q", "world", ls.Line(1))
	}
	if ls.Line(2) != "foo" {
		t.Fatalf("line 2: expected %q, got %q", "foo", ls.Line(2))
	}
}

func TestLineStore_LineOutOfRange(t *testing.T) {
	ls := NewLineStore("hello")
	if ls.Line(-1) != "" {
		t.Fatal("expected empty for negative index")
	}
	if ls.Line(5) != "" {
		t.Fatal("expected empty for out of range index")
	}
}

func TestLineStore_InsertSingleLine(t *testing.T) {
	ls := NewLineStore("hello world")
	err := ls.Insert(Position{0, 5}, " beautiful")
	if err != nil {
		t.Fatal(err)
	}
	if ls.Line(0) != "hello beautiful world" {
		t.Fatalf("expected %q, got %q", "hello beautiful world", ls.Line(0))
	}
}

func TestLineStore_InsertAtBeginning(t *testing.T) {
	ls := NewLineStore("world")
	err := ls.Insert(Position{0, 0}, "hello ")
	if err != nil {
		t.Fatal(err)
	}
	if ls.Line(0) != "hello world" {
		t.Fatalf("expected %q, got %q", "hello world", ls.Line(0))
	}
}

func TestLineStore_InsertAtEnd(t *testing.T) {
	ls := NewLineStore("hello")
	err := ls.Insert(Position{0, 5}, " world")
	if err != nil {
		t.Fatal(err)
	}
	if ls.Line(0) != "hello world" {
		t.Fatalf("expected %q, got %q", "hello world", ls.Line(0))
	}
}

func TestLineStore_InsertNewline(t *testing.T) {
	ls := NewLineStore("helloworld")
	err := ls.Insert(Position{0, 5}, "\n")
	if err != nil {
		t.Fatal(err)
	}
	if ls.LineCount() != 2 {
		t.Fatalf("expected 2 lines, got %d", ls.LineCount())
	}
	if ls.Line(0) != "hello" {
		t.Fatalf("line 0: expected %q, got %q", "hello", ls.Line(0))
	}
	if ls.Line(1) != "world" {
		t.Fatalf("line 1: expected %q, got %q", "world", ls.Line(1))
	}
}

func TestLineStore_InsertMultiLine(t *testing.T) {
	ls := NewLineStore("AD")
	err := ls.Insert(Position{0, 1}, "B\nC\n")
	if err != nil {
		t.Fatal(err)
	}
	if ls.LineCount() != 3 {
		t.Fatalf("expected 3 lines, got %d", ls.LineCount())
	}
	if ls.Line(0) != "AB" {
		t.Fatalf("line 0: expected %q, got %q", "AB", ls.Line(0))
	}
	if ls.Line(1) != "C" {
		t.Fatalf("line 1: expected %q, got %q", "C", ls.Line(1))
	}
	if ls.Line(2) != "D" {
		t.Fatalf("line 2: expected %q, got %q", "D", ls.Line(2))
	}
}

func TestLineStore_InsertOutOfRange(t *testing.T) {
	ls := NewLineStore("hello")
	err := ls.Insert(Position{5, 0}, "x")
	if err != ErrOutOfRange {
		t.Fatalf("expected ErrOutOfRange, got %v", err)
	}
	err = ls.Insert(Position{0, 100}, "x")
	if err != ErrOutOfRange {
		t.Fatalf("expected ErrOutOfRange, got %v", err)
	}
}

func TestLineStore_DeleteWithinLine(t *testing.T) {
	ls := NewLineStore("hello beautiful world")
	err := ls.Delete(Range{Position{0, 5}, Position{0, 15}})
	if err != nil {
		t.Fatal(err)
	}
	if ls.Line(0) != "hello world" {
		t.Fatalf("expected %q, got %q", "hello world", ls.Line(0))
	}
}

func TestLineStore_DeleteMergingLines(t *testing.T) {
	ls := NewLineStore("hello\nworld")
	// Delete from end of line 0 to start of line 1 — merges them.
	err := ls.Delete(Range{Position{0, 5}, Position{1, 0}})
	if err != nil {
		t.Fatal(err)
	}
	if ls.LineCount() != 1 {
		t.Fatalf("expected 1 line, got %d", ls.LineCount())
	}
	if ls.Line(0) != "helloworld" {
		t.Fatalf("expected %q, got %q", "helloworld", ls.Line(0))
	}
}

func TestLineStore_DeleteMultipleLines(t *testing.T) {
	ls := NewLineStore("aaa\nbbb\nccc\nddd")
	// Delete from middle of line 0 to middle of line 2.
	err := ls.Delete(Range{Position{0, 1}, Position{2, 2}})
	if err != nil {
		t.Fatal(err)
	}
	if ls.LineCount() != 2 {
		t.Fatalf("expected 2 lines, got %d", ls.LineCount())
	}
	if ls.Line(0) != "ac" {
		t.Fatalf("line 0: expected %q, got %q", "ac", ls.Line(0))
	}
	if ls.Line(1) != "ddd" {
		t.Fatalf("line 1: expected %q, got %q", "ddd", ls.Line(1))
	}
}

func TestLineStore_DeleteOutOfRange(t *testing.T) {
	ls := NewLineStore("hello")
	err := ls.Delete(Range{Position{5, 0}, Position{5, 1}})
	if err != ErrOutOfRange {
		t.Fatalf("expected ErrOutOfRange, got %v", err)
	}
}

func TestLineStore_SliceSingleLine(t *testing.T) {
	ls := NewLineStore("hello world")
	s := ls.Slice(Range{Position{0, 0}, Position{0, 5}})
	if s != "hello" {
		t.Fatalf("expected %q, got %q", "hello", s)
	}
}

func TestLineStore_SliceMultiLine(t *testing.T) {
	ls := NewLineStore("aaa\nbbb\nccc")
	s := ls.Slice(Range{Position{0, 1}, Position{2, 2}})
	if s != "aa\nbbb\ncc" {
		t.Fatalf("expected %q, got %q", "aa\nbbb\ncc", s)
	}
}

func TestLineStore_Content(t *testing.T) {
	ls := NewLineStore("hello\nworld")
	if ls.Content() != "hello\nworld" {
		t.Fatalf("expected %q, got %q", "hello\nworld", ls.Content())
	}
}

func TestLineStore_DeleteReversedRange(t *testing.T) {
	ls := NewLineStore("hello\nworld")
	// Start > End should return error.
	err := ls.Delete(Range{Position{1, 0}, Position{0, 0}})
	if err != ErrOutOfRange {
		t.Fatalf("expected ErrOutOfRange for reversed range, got %v", err)
	}
	// Same line, reversed cols.
	err = ls.Delete(Range{Position{0, 3}, Position{0, 1}})
	if err != ErrOutOfRange {
		t.Fatalf("expected ErrOutOfRange for reversed col range, got %v", err)
	}
}

func TestLineStore_SliceNegativeEnd(t *testing.T) {
	ls := NewLineStore("hello")
	s := ls.Slice(Range{Position{0, 0}, Position{-1, 0}})
	if s != "" {
		t.Fatalf("expected empty for negative End.Line, got %q", s)
	}
}

func TestLineStore_SliceReversedRange(t *testing.T) {
	ls := NewLineStore("hello\nworld")
	s := ls.Slice(Range{Position{1, 0}, Position{0, 0}})
	if s != "" {
		t.Fatalf("expected empty for reversed range, got %q", s)
	}
}

func TestLineStore_InsertUnicode(t *testing.T) {
	ls := NewLineStore("ac")
	// Insert a multi-byte rune at position 1 (byte offset).
	err := ls.Insert(Position{0, 1}, "\u00f1") // ñ = 2 bytes
	if err != nil {
		t.Fatal(err)
	}
	if ls.Line(0) != "a\u00f1c" {
		t.Fatalf("expected %q, got %q", "a\u00f1c", ls.Line(0))
	}
}
