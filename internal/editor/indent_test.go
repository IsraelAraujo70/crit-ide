package editor

import (
	"strings"
	"testing"
)

// helper to create a buffer with exact multi-line content (no auto-indent).
func bufWithLines(lines ...string) *Buffer {
	b := NewBuffer("test")
	content := strings.Join(lines, "\n")
	b.Text = NewLineStore(content)
	b.Undo = NewUndoStack(1000)
	return b
}

// --- IndentSelection tests ---

func TestIndentSelection_MultiLine(t *testing.T) {
	b := bufWithLines("aaa", "bbb", "ccc")
	// Select lines 0-2 fully.
	b.SetSelection(Position{0, 0}, Position{2, 3})
	b.CursorRow = 2
	b.CursorCol = 3

	b.IndentSelection()

	if b.Text.Line(0) != "\taaa" {
		t.Fatalf("line 0: expected %q, got %q", "\taaa", b.Text.Line(0))
	}
	if b.Text.Line(1) != "\tbbb" {
		t.Fatalf("line 1: expected %q, got %q", "\tbbb", b.Text.Line(1))
	}
	if b.Text.Line(2) != "\tccc" {
		t.Fatalf("line 2: expected %q, got %q", "\tccc", b.Text.Line(2))
	}
	if !b.Dirty {
		t.Fatal("buffer should be dirty")
	}
}

func TestIndentSelection_PreservesSelection(t *testing.T) {
	b := bufWithLines("aaa", "bbb", "ccc")
	b.SetSelection(Position{0, 1}, Position{2, 2})
	b.CursorRow = 2
	b.CursorCol = 2

	b.IndentSelection()

	if !b.HasSelection() {
		t.Fatal("selection should be preserved")
	}
	sel := b.Selection
	// Anchor was at (0,1) → (0,2) after tab prepended.
	if sel.Anchor.Line != 0 || sel.Anchor.Col != 2 {
		t.Fatalf("anchor: expected (0,2), got (%d,%d)", sel.Anchor.Line, sel.Anchor.Col)
	}
	// Cursor was at (2,2) → (2,3).
	if sel.Cursor.Line != 2 || sel.Cursor.Col != 3 {
		t.Fatalf("cursor: expected (2,3), got (%d,%d)", sel.Cursor.Line, sel.Cursor.Col)
	}
}

func TestIndentSelection_EndAtCol0_SkipsLastLine(t *testing.T) {
	b := bufWithLines("aaa", "bbb", "ccc")
	// Selection from line 0 to beginning of line 2 (col 0) — should only indent lines 0-1.
	b.SetSelection(Position{0, 0}, Position{2, 0})
	b.CursorRow = 2
	b.CursorCol = 0

	b.IndentSelection()

	if b.Text.Line(0) != "\taaa" {
		t.Fatalf("line 0: expected %q, got %q", "\taaa", b.Text.Line(0))
	}
	if b.Text.Line(1) != "\tbbb" {
		t.Fatalf("line 1: expected %q, got %q", "\tbbb", b.Text.Line(1))
	}
	if b.Text.Line(2) != "ccc" {
		t.Fatalf("line 2: expected %q (unchanged), got %q", "ccc", b.Text.Line(2))
	}
}

func TestIndentSelection_SingleLine(t *testing.T) {
	b := bufWithLines("hello", "world")
	b.SetSelection(Position{0, 0}, Position{0, 5})
	b.CursorRow = 0
	b.CursorCol = 5

	b.IndentSelection()

	if b.Text.Line(0) != "\thello" {
		t.Fatalf("line 0: expected %q, got %q", "\thello", b.Text.Line(0))
	}
	if b.Text.Line(1) != "world" {
		t.Fatalf("line 1: expected %q (unchanged), got %q", "world", b.Text.Line(1))
	}
}

func TestIndentSelection_NoSelection(t *testing.T) {
	b := bufWithLines("hello")
	// No selection — should be a no-op.
	b.IndentSelection()

	if b.Text.Line(0) != "hello" {
		t.Fatalf("expected no change, got %q", b.Text.Line(0))
	}
}

func TestIndentSelection_ReadOnly(t *testing.T) {
	b := bufWithLines("hello")
	b.ReadOnly = true
	b.SetSelection(Position{0, 0}, Position{0, 5})

	b.IndentSelection()

	if b.Text.Line(0) != "hello" {
		t.Fatalf("readonly: expected no change, got %q", b.Text.Line(0))
	}
}

func TestIndentSelection_Undo(t *testing.T) {
	b := bufWithLines("aaa", "bbb", "ccc")
	b.SetSelection(Position{0, 0}, Position{2, 3})
	b.CursorRow = 2
	b.CursorCol = 3

	b.IndentSelection()

	// Verify indented.
	if b.Text.Line(0) != "\taaa" {
		t.Fatalf("after indent, line 0: expected %q, got %q", "\taaa", b.Text.Line(0))
	}

	// Undo: should revert insert first, then re-insert old text.
	b.UndoEdit() // undo the insert of new text
	b.UndoEdit() // undo the delete of old text

	if b.Text.Line(0) != "aaa" {
		t.Fatalf("after undo, line 0: expected %q, got %q", "aaa", b.Text.Line(0))
	}
	if b.Text.Line(1) != "bbb" {
		t.Fatalf("after undo, line 1: expected %q, got %q", "bbb", b.Text.Line(1))
	}
	if b.Text.Line(2) != "ccc" {
		t.Fatalf("after undo, line 2: expected %q, got %q", "ccc", b.Text.Line(2))
	}
}

// --- DedentSelection tests ---

func TestDedentSelection_TabIndent(t *testing.T) {
	b := bufWithLines("\taaa", "\tbbb", "\tccc")
	b.SetSelection(Position{0, 0}, Position{2, 4})
	b.CursorRow = 2
	b.CursorCol = 4

	b.DedentSelection()

	if b.Text.Line(0) != "aaa" {
		t.Fatalf("line 0: expected %q, got %q", "aaa", b.Text.Line(0))
	}
	if b.Text.Line(1) != "bbb" {
		t.Fatalf("line 1: expected %q, got %q", "bbb", b.Text.Line(1))
	}
	if b.Text.Line(2) != "ccc" {
		t.Fatalf("line 2: expected %q, got %q", "ccc", b.Text.Line(2))
	}
}

func TestDedentSelection_SpaceIndent(t *testing.T) {
	b := bufWithLines("    aaa", "    bbb")
	b.SetSelection(Position{0, 0}, Position{1, 7})
	b.CursorRow = 1
	b.CursorCol = 7

	b.DedentSelection()

	if b.Text.Line(0) != "aaa" {
		t.Fatalf("line 0: expected %q, got %q", "aaa", b.Text.Line(0))
	}
	if b.Text.Line(1) != "bbb" {
		t.Fatalf("line 1: expected %q, got %q", "bbb", b.Text.Line(1))
	}
}

func TestDedentSelection_PartialSpaceIndent(t *testing.T) {
	b := bufWithLines("  aaa", "      bbb")
	b.SetSelection(Position{0, 0}, Position{1, 9})
	b.CursorRow = 1
	b.CursorCol = 9

	b.DedentSelection()

	// Line 0 had 2 spaces → removed 2.
	if b.Text.Line(0) != "aaa" {
		t.Fatalf("line 0: expected %q, got %q", "aaa", b.Text.Line(0))
	}
	// Line 1 had 6 spaces → removed 4.
	if b.Text.Line(1) != "  bbb" {
		t.Fatalf("line 1: expected %q, got %q", "  bbb", b.Text.Line(1))
	}
}

func TestDedentSelection_MixedWhitespace(t *testing.T) {
	b := bufWithLines("\taaa", "    bbb", "  ccc")
	b.SetSelection(Position{0, 0}, Position{2, 5})
	b.CursorRow = 2
	b.CursorCol = 5

	b.DedentSelection()

	if b.Text.Line(0) != "aaa" {
		t.Fatalf("line 0: expected %q, got %q", "aaa", b.Text.Line(0))
	}
	if b.Text.Line(1) != "bbb" {
		t.Fatalf("line 1: expected %q, got %q", "bbb", b.Text.Line(1))
	}
	if b.Text.Line(2) != "ccc" {
		t.Fatalf("line 2: expected %q, got %q", "ccc", b.Text.Line(2))
	}
}

func TestDedentSelection_NoIndent(t *testing.T) {
	b := bufWithLines("aaa", "bbb")
	b.SetSelection(Position{0, 0}, Position{1, 3})
	b.CursorRow = 1
	b.CursorCol = 3

	b.DedentSelection()

	// No change expected.
	if b.Text.Line(0) != "aaa" {
		t.Fatalf("line 0: expected %q, got %q", "aaa", b.Text.Line(0))
	}
	if b.Text.Line(1) != "bbb" {
		t.Fatalf("line 1: expected %q, got %q", "bbb", b.Text.Line(1))
	}
}

func TestDedentSelection_PreservesSelection(t *testing.T) {
	b := bufWithLines("\taaa", "\tbbb")
	b.SetSelection(Position{0, 2}, Position{1, 3})
	b.CursorRow = 1
	b.CursorCol = 3

	b.DedentSelection()

	if !b.HasSelection() {
		t.Fatal("selection should be preserved")
	}
	sel := b.Selection
	// Anchor was at (0,2) → (0,1) after tab removed.
	if sel.Anchor.Line != 0 || sel.Anchor.Col != 1 {
		t.Fatalf("anchor: expected (0,1), got (%d,%d)", sel.Anchor.Line, sel.Anchor.Col)
	}
	// Cursor was at (1,3) → (1,2).
	if sel.Cursor.Line != 1 || sel.Cursor.Col != 2 {
		t.Fatalf("cursor: expected (1,2), got (%d,%d)", sel.Cursor.Line, sel.Cursor.Col)
	}
}

func TestDedentSelection_Undo(t *testing.T) {
	b := bufWithLines("\taaa", "\tbbb")
	b.SetSelection(Position{0, 0}, Position{1, 4})
	b.CursorRow = 1
	b.CursorCol = 4

	b.DedentSelection()

	if b.Text.Line(0) != "aaa" {
		t.Fatalf("after dedent, line 0: expected %q, got %q", "aaa", b.Text.Line(0))
	}

	b.UndoEdit() // undo insert of new text
	b.UndoEdit() // undo delete of old text

	if b.Text.Line(0) != "\taaa" {
		t.Fatalf("after undo, line 0: expected %q, got %q", "\taaa", b.Text.Line(0))
	}
	if b.Text.Line(1) != "\tbbb" {
		t.Fatalf("after undo, line 1: expected %q, got %q", "\tbbb", b.Text.Line(1))
	}
}

func TestDedentSelection_EndAtCol0_SkipsLastLine(t *testing.T) {
	b := bufWithLines("\taaa", "\tbbb", "\tccc")
	b.SetSelection(Position{0, 0}, Position{2, 0})
	b.CursorRow = 2
	b.CursorCol = 0

	b.DedentSelection()

	if b.Text.Line(0) != "aaa" {
		t.Fatalf("line 0: expected %q, got %q", "aaa", b.Text.Line(0))
	}
	if b.Text.Line(1) != "bbb" {
		t.Fatalf("line 1: expected %q, got %q", "bbb", b.Text.Line(1))
	}
	// Line 2 should be unchanged.
	if b.Text.Line(2) != "\tccc" {
		t.Fatalf("line 2: expected %q, got %q", "\tccc", b.Text.Line(2))
	}
}

func TestIndentDedent_Roundtrip(t *testing.T) {
	b := bufWithLines("aaa", "bbb", "ccc")
	b.SetSelection(Position{0, 0}, Position{2, 3})
	b.CursorRow = 2
	b.CursorCol = 3

	b.IndentSelection()
	b.DedentSelection()

	if b.Text.Line(0) != "aaa" {
		t.Fatalf("line 0: expected %q, got %q", "aaa", b.Text.Line(0))
	}
	if b.Text.Line(1) != "bbb" {
		t.Fatalf("line 1: expected %q, got %q", "bbb", b.Text.Line(1))
	}
	if b.Text.Line(2) != "ccc" {
		t.Fatalf("line 2: expected %q, got %q", "ccc", b.Text.Line(2))
	}
}
