package editor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewBuffer(t *testing.T) {
	b := NewBuffer("test")
	if b.ID != "test" {
		t.Fatalf("expected id %q, got %q", "test", b.ID)
	}
	if b.Kind != BufferKindScratch {
		t.Fatal("expected scratch buffer")
	}
	if b.Text.LineCount() != 1 {
		t.Fatalf("expected 1 line, got %d", b.Text.LineCount())
	}
}

func TestLoadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	err := os.WriteFile(path, []byte("hello\nworld\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	b, err := LoadFile("f1", path)
	if err != nil {
		t.Fatal(err)
	}
	if b.Text.LineCount() != 2 {
		t.Fatalf("expected 2 lines, got %d", b.Text.LineCount())
	}
	if b.Text.Line(0) != "hello" {
		t.Fatalf("line 0: expected %q, got %q", "hello", b.Text.Line(0))
	}
	if b.Text.Line(1) != "world" {
		t.Fatalf("line 1: expected %q, got %q", "world", b.Text.Line(1))
	}
	if b.Kind != BufferKindFile {
		t.Fatal("expected file buffer")
	}
	if b.Dirty {
		t.Fatal("buffer should not be dirty after load")
	}
}

func TestBuffer_SaveFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")

	b := NewBuffer("s1")
	b.Path = path
	b.Kind = BufferKindFile
	b.InsertChar('H')
	b.InsertChar('i')

	err := b.SaveFile()
	if err != nil {
		t.Fatal(err)
	}
	if b.Dirty {
		t.Fatal("buffer should not be dirty after save")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "Hi\n" {
		t.Fatalf("expected %q, got %q", "Hi\n", string(data))
	}
}

func TestBuffer_InsertChar(t *testing.T) {
	b := NewBuffer("t")
	b.InsertChar('a')
	b.InsertChar('b')
	b.InsertChar('c')

	if b.Text.Line(0) != "abc" {
		t.Fatalf("expected %q, got %q", "abc", b.Text.Line(0))
	}
	if b.CursorCol != 3 {
		t.Fatalf("expected cursor col 3, got %d", b.CursorCol)
	}
	if !b.Dirty {
		t.Fatal("buffer should be dirty after insert")
	}
}

func TestBuffer_InsertNewline(t *testing.T) {
	b := NewBuffer("t")
	b.InsertChar('a')
	b.InsertChar('b')
	b.InsertNewline()
	b.InsertChar('c')

	if b.Text.LineCount() != 2 {
		t.Fatalf("expected 2 lines, got %d", b.Text.LineCount())
	}
	if b.Text.Line(0) != "ab" {
		t.Fatalf("line 0: expected %q, got %q", "ab", b.Text.Line(0))
	}
	if b.Text.Line(1) != "c" {
		t.Fatalf("line 1: expected %q, got %q", "c", b.Text.Line(1))
	}
	if b.CursorRow != 1 || b.CursorCol != 1 {
		t.Fatalf("expected cursor at (1,1), got (%d,%d)", b.CursorRow, b.CursorCol)
	}
}

func TestBuffer_DeleteBackward_SameLine(t *testing.T) {
	b := NewBuffer("t")
	b.InsertChar('a')
	b.InsertChar('b')
	b.InsertChar('c')
	b.DeleteBackward()

	if b.Text.Line(0) != "ab" {
		t.Fatalf("expected %q, got %q", "ab", b.Text.Line(0))
	}
	if b.CursorCol != 2 {
		t.Fatalf("expected cursor col 2, got %d", b.CursorCol)
	}
}

func TestBuffer_DeleteBackward_MergeLines(t *testing.T) {
	b := NewBuffer("t")
	b.InsertChar('a')
	b.InsertNewline()
	b.InsertChar('b')
	// Cursor at (1,1). Move to start of line 1.
	b.CursorHome()
	b.DeleteBackward()

	if b.Text.LineCount() != 1 {
		t.Fatalf("expected 1 line, got %d", b.Text.LineCount())
	}
	if b.Text.Line(0) != "ab" {
		t.Fatalf("expected %q, got %q", "ab", b.Text.Line(0))
	}
	if b.CursorRow != 0 || b.CursorCol != 1 {
		t.Fatalf("expected cursor at (0,1), got (%d,%d)", b.CursorRow, b.CursorCol)
	}
}

func TestBuffer_DeleteBackward_AtStart(t *testing.T) {
	b := NewBuffer("t")
	b.InsertChar('a')
	b.CursorHome()
	b.DeleteBackward() // Should do nothing.

	if b.Text.Line(0) != "a" {
		t.Fatalf("expected %q, got %q", "a", b.Text.Line(0))
	}
}

func TestBuffer_DeleteForward_SameLine(t *testing.T) {
	b := NewBuffer("t")
	b.InsertChar('a')
	b.InsertChar('b')
	b.CursorHome()
	b.DeleteForward()

	if b.Text.Line(0) != "b" {
		t.Fatalf("expected %q, got %q", "b", b.Text.Line(0))
	}
	if b.CursorCol != 0 {
		t.Fatalf("expected cursor col 0, got %d", b.CursorCol)
	}
}

func TestBuffer_DeleteForward_MergeLines(t *testing.T) {
	b := NewBuffer("t")
	b.InsertChar('a')
	b.InsertNewline()
	b.InsertChar('b')
	// Move cursor to end of line 0.
	b.CursorRow = 0
	b.CursorEnd()
	b.DeleteForward()

	if b.Text.LineCount() != 1 {
		t.Fatalf("expected 1 line, got %d", b.Text.LineCount())
	}
	if b.Text.Line(0) != "ab" {
		t.Fatalf("expected %q, got %q", "ab", b.Text.Line(0))
	}
}

func TestBuffer_MoveCursor(t *testing.T) {
	b := NewBuffer("t")
	b.InsertChar('a')
	b.InsertChar('b')
	b.InsertNewline()
	b.InsertChar('c')
	b.InsertChar('d')
	// Buffer: "ab\ncd", cursor at (1,2).

	b.MoveCursor(DirUp)
	if b.CursorRow != 0 || b.CursorCol != 2 {
		t.Fatalf("after up: expected (0,2), got (%d,%d)", b.CursorRow, b.CursorCol)
	}

	b.MoveCursor(DirLeft)
	if b.CursorCol != 1 {
		t.Fatalf("after left: expected col 1, got %d", b.CursorCol)
	}

	b.MoveCursor(DirDown)
	if b.CursorRow != 1 {
		t.Fatalf("after down: expected row 1, got %d", b.CursorRow)
	}

	b.MoveCursor(DirRight)
	if b.CursorCol != 2 {
		t.Fatalf("after right: expected col 2, got %d", b.CursorCol)
	}
}

func TestBuffer_MoveCursor_WrapLeft(t *testing.T) {
	b := NewBuffer("t")
	b.InsertChar('a')
	b.InsertNewline()
	b.InsertChar('b')
	// Cursor at (1,1). Move to start of line 1.
	b.CursorHome()
	// Left should wrap to end of line 0.
	b.MoveCursor(DirLeft)
	if b.CursorRow != 0 || b.CursorCol != 1 {
		t.Fatalf("expected (0,1), got (%d,%d)", b.CursorRow, b.CursorCol)
	}
}

func TestBuffer_MoveCursor_WrapRight(t *testing.T) {
	b := NewBuffer("t")
	b.InsertChar('a')
	b.InsertNewline()
	b.InsertChar('b')
	// Move to end of line 0.
	b.CursorRow = 0
	b.CursorEnd()
	// Right should wrap to start of line 1.
	b.MoveCursor(DirRight)
	if b.CursorRow != 1 || b.CursorCol != 0 {
		t.Fatalf("expected (1,0), got (%d,%d)", b.CursorRow, b.CursorCol)
	}
}

func TestBuffer_CursorHomeEnd(t *testing.T) {
	b := NewBuffer("t")
	b.InsertChar('a')
	b.InsertChar('b')
	b.InsertChar('c')

	b.CursorHome()
	if b.CursorCol != 0 {
		t.Fatalf("expected col 0 after Home, got %d", b.CursorCol)
	}

	b.CursorEnd()
	if b.CursorCol != 3 {
		t.Fatalf("expected col 3 after End, got %d", b.CursorCol)
	}
}

func TestBuffer_ClampCursor(t *testing.T) {
	b := NewBuffer("t")
	b.InsertChar('a')
	b.InsertChar('b')
	b.CursorRow = 100
	b.CursorCol = 100
	b.ClampCursor()

	if b.CursorRow != 0 {
		t.Fatalf("expected row 0, got %d", b.CursorRow)
	}
	if b.CursorCol != 2 {
		t.Fatalf("expected col 2, got %d", b.CursorCol)
	}
}

func TestBuffer_ReadOnly(t *testing.T) {
	b := NewBuffer("t")
	b.ReadOnly = true
	b.InsertChar('a')

	if b.Text.Line(0) != "" {
		t.Fatal("readonly buffer should not accept inserts")
	}
}

func TestBuffer_FileName(t *testing.T) {
	b := NewBuffer("t")
	if b.FileName() != "[scratch]" {
		t.Fatalf("expected [scratch], got %q", b.FileName())
	}

	b.Path = "/tmp/foo/bar.go"
	if b.FileName() != "bar.go" {
		t.Fatalf("expected bar.go, got %q", b.FileName())
	}
}

func TestBuffer_InsertChar_Unicode(t *testing.T) {
	b := NewBuffer("t")
	b.InsertChar('a')       // 1 byte
	b.InsertChar('\u00f1')  // ñ = 2 bytes
	b.InsertChar('\u4e2d')  // 中 = 3 bytes

	want := "a\u00f1\u4e2d"
	if b.Text.Line(0) != want {
		t.Fatalf("expected %q, got %q", want, b.Text.Line(0))
	}
	// CursorCol should be 1 + 2 + 3 = 6 (byte offset).
	if b.CursorCol != 6 {
		t.Fatalf("expected cursor col 6, got %d", b.CursorCol)
	}
}

func TestBuffer_DeleteBackward_Unicode(t *testing.T) {
	b := NewBuffer("t")
	b.InsertChar('a')
	b.InsertChar('\u00f1') // ñ = 2 bytes
	// CursorCol = 3. Backspace should remove the full rune.
	b.DeleteBackward()
	if b.Text.Line(0) != "a" {
		t.Fatalf("expected %q, got %q", "a", b.Text.Line(0))
	}
	if b.CursorCol != 1 {
		t.Fatalf("expected cursor col 1, got %d", b.CursorCol)
	}
}

func TestBuffer_DeleteForward_Unicode(t *testing.T) {
	b := NewBuffer("t")
	b.InsertChar('\u4e2d') // 中 = 3 bytes
	b.InsertChar('a')
	b.CursorHome()
	b.DeleteForward()
	if b.Text.Line(0) != "a" {
		t.Fatalf("expected %q, got %q", "a", b.Text.Line(0))
	}
}

func TestBuffer_StickyColumn(t *testing.T) {
	b := NewBuffer("t")
	// Line 0: "abcdef" (6 chars)
	for _, ch := range "abcdef" {
		b.InsertChar(ch)
	}
	b.InsertNewline()
	// Line 1: "xy" (2 chars)
	b.InsertChar('x')
	b.InsertChar('y')
	b.InsertNewline()
	// Line 2: "abcdef" (6 chars)
	for _, ch := range "abcdef" {
		b.InsertChar(ch)
	}

	// Move cursor to end of line 2 (col 6).
	// Already there.
	if b.CursorCol != 6 {
		t.Fatalf("expected col 6, got %d", b.CursorCol)
	}

	// Move up to line 1 — should clamp to col 2 (line is "xy").
	b.MoveCursor(DirUp)
	if b.CursorRow != 1 || b.CursorCol != 2 {
		t.Fatalf("after up to short line: expected (1,2), got (%d,%d)", b.CursorRow, b.CursorCol)
	}

	// Move up to line 0 — should restore to col 6 (sticky column).
	b.MoveCursor(DirUp)
	if b.CursorRow != 0 || b.CursorCol != 6 {
		t.Fatalf("after up to long line: expected (0,6), got (%d,%d)", b.CursorRow, b.CursorCol)
	}
}

func TestBuffer_DeleteForward_AtEnd(t *testing.T) {
	b := NewBuffer("t")
	b.InsertChar('a')
	b.CursorEnd()
	// Should do nothing — single line, at end.
	b.DeleteForward()
	if b.Text.Line(0) != "a" {
		t.Fatalf("expected %q, got %q", "a", b.Text.Line(0))
	}
}
