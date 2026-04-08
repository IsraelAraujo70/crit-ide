package editor

import (
	"testing"
)

func newTestStore(content string) TextStore {
	return NewLineStore(content)
}

func TestSearchState_FindAll_Basic(t *testing.T) {
	store := newTestStore("hello world\nhello go\ngoodbye world")
	s := NewSearchState()
	s.Query = "hello"
	s.FindAll(store)

	if len(s.Matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(s.Matches))
	}
	if s.Matches[0].Start.Line != 0 || s.Matches[0].Start.Col != 0 {
		t.Errorf("match 0: expected (0,0), got (%d,%d)", s.Matches[0].Start.Line, s.Matches[0].Start.Col)
	}
	if s.Matches[1].Start.Line != 1 || s.Matches[1].Start.Col != 0 {
		t.Errorf("match 1: expected (1,0), got (%d,%d)", s.Matches[1].Start.Line, s.Matches[1].Start.Col)
	}
	if s.CurrentIdx != 0 {
		t.Errorf("expected currentIdx 0, got %d", s.CurrentIdx)
	}
}

func TestSearchState_FindAll_EmptyQuery(t *testing.T) {
	store := newTestStore("hello world")
	s := NewSearchState()
	s.Query = ""
	s.FindAll(store)

	if len(s.Matches) != 0 {
		t.Fatalf("expected 0 matches for empty query, got %d", len(s.Matches))
	}
	if s.CurrentIdx != -1 {
		t.Errorf("expected currentIdx -1, got %d", s.CurrentIdx)
	}
}

func TestSearchState_FindAll_NoMatch(t *testing.T) {
	store := newTestStore("hello world")
	s := NewSearchState()
	s.Query = "xyz"
	s.FindAll(store)

	if len(s.Matches) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(s.Matches))
	}
}

func TestSearchState_FindAll_MultipleOnSameLine(t *testing.T) {
	store := newTestStore("abcabcabc")
	s := NewSearchState()
	s.Query = "abc"
	s.FindAll(store)

	if len(s.Matches) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(s.Matches))
	}
	if s.Matches[0].Start.Col != 0 {
		t.Errorf("match 0 col: expected 0, got %d", s.Matches[0].Start.Col)
	}
	if s.Matches[1].Start.Col != 3 {
		t.Errorf("match 1 col: expected 3, got %d", s.Matches[1].Start.Col)
	}
	if s.Matches[2].Start.Col != 6 {
		t.Errorf("match 2 col: expected 6, got %d", s.Matches[2].Start.Col)
	}
}

func TestSearchState_FindNext_WrapAround(t *testing.T) {
	store := newTestStore("aaa\nbbb\naaa")
	s := NewSearchState()
	s.Query = "aaa"
	s.FindAll(store)

	if len(s.Matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(s.Matches))
	}

	// From after second match, should wrap to first.
	pos, found := s.FindNext(2, 5)
	if !found {
		t.Fatal("expected to find next match")
	}
	if pos.Line != 0 || pos.Col != 0 {
		t.Errorf("expected wrap to (0,0), got (%d,%d)", pos.Line, pos.Col)
	}
	if s.CurrentIdx != 0 {
		t.Errorf("expected currentIdx 0, got %d", s.CurrentIdx)
	}
}

func TestSearchState_FindNext_Sequential(t *testing.T) {
	store := newTestStore("hello hello hello")
	s := NewSearchState()
	s.Query = "hello"
	s.FindAll(store)

	// From beginning.
	pos, _ := s.FindNext(0, 0)
	if pos.Col != 6 {
		t.Errorf("expected col 6, got %d", pos.Col)
	}

	pos, _ = s.FindNext(0, 6)
	if pos.Col != 12 {
		t.Errorf("expected col 12, got %d", pos.Col)
	}
}

func TestSearchState_FindPrev_WrapAround(t *testing.T) {
	store := newTestStore("aaa\nbbb\naaa")
	s := NewSearchState()
	s.Query = "aaa"
	s.FindAll(store)

	// From before first match, should wrap to last.
	pos, found := s.FindPrev(0, 0)
	if !found {
		t.Fatal("expected to find prev match")
	}
	if pos.Line != 2 || pos.Col != 0 {
		t.Errorf("expected wrap to (2,0), got (%d,%d)", pos.Line, pos.Col)
	}
}

func TestSearchState_FindNearest(t *testing.T) {
	store := newTestStore("xxx\nabc\nxxx\nabc")
	s := NewSearchState()
	s.Query = "abc"
	s.FindAll(store)

	s.FindNearest(2, 0) // After first match, before second.
	if s.CurrentIdx != 1 {
		t.Errorf("expected currentIdx 1, got %d", s.CurrentIdx)
	}
}

func TestSearchState_CurrentMatch(t *testing.T) {
	s := NewSearchState()

	// No matches.
	_, ok := s.CurrentMatch()
	if ok {
		t.Error("expected no current match")
	}

	store := newTestStore("hello")
	s.Query = "hello"
	s.FindAll(store)

	m, ok := s.CurrentMatch()
	if !ok {
		t.Fatal("expected current match")
	}
	if m.Start.Col != 0 || m.End.Col != 5 {
		t.Errorf("unexpected match range: %v", m)
	}
}

func TestSearchState_MatchCount(t *testing.T) {
	store := newTestStore("aaa bbb aaa")
	s := NewSearchState()
	s.Query = "aaa"
	s.FindAll(store)

	if s.MatchCount() != 2 {
		t.Errorf("expected 2, got %d", s.MatchCount())
	}
	if s.CurrentMatchNumber() != 1 {
		t.Errorf("expected 1, got %d", s.CurrentMatchNumber())
	}
}

func TestSearchState_InsertChar(t *testing.T) {
	s := NewSearchState()
	s.ActiveField = FieldFind

	s.InsertChar('a')
	s.InsertChar('b')
	s.InsertChar('c')

	if s.Query != "abc" {
		t.Errorf("expected 'abc', got %q", s.Query)
	}
	if s.QueryCursor != 3 {
		t.Errorf("expected cursor at 3, got %d", s.QueryCursor)
	}
}

func TestSearchState_InsertChar_Replace(t *testing.T) {
	s := NewSearchState()
	s.ActiveField = FieldReplace

	s.InsertChar('x')
	s.InsertChar('y')

	if s.ReplaceText != "xy" {
		t.Errorf("expected 'xy', got %q", s.ReplaceText)
	}
}

func TestSearchState_DeleteBackward(t *testing.T) {
	s := NewSearchState()
	s.ActiveField = FieldFind
	s.Query = "abc"
	s.QueryCursor = 3

	s.DeleteBackward()
	if s.Query != "ab" || s.QueryCursor != 2 {
		t.Errorf("after delete: query=%q cursor=%d", s.Query, s.QueryCursor)
	}

	// Delete at start should be a no-op.
	s.QueryCursor = 0
	s.DeleteBackward()
	if s.Query != "ab" {
		t.Errorf("delete at start should be no-op, got %q", s.Query)
	}
}

func TestSearchState_DeleteForward(t *testing.T) {
	s := NewSearchState()
	s.ActiveField = FieldFind
	s.Query = "abc"
	s.QueryCursor = 0

	s.DeleteForward()
	if s.Query != "bc" || s.QueryCursor != 0 {
		t.Errorf("after delete: query=%q cursor=%d", s.Query, s.QueryCursor)
	}
}

func TestSearchState_MoveLeftRight(t *testing.T) {
	s := NewSearchState()
	s.ActiveField = FieldFind
	s.Query = "abc"
	s.QueryCursor = 3

	s.MoveLeft()
	if s.QueryCursor != 2 {
		t.Errorf("expected cursor 2, got %d", s.QueryCursor)
	}

	s.MoveRight()
	if s.QueryCursor != 3 {
		t.Errorf("expected cursor 3, got %d", s.QueryCursor)
	}
}

func TestSearchState_MoveHomeEnd(t *testing.T) {
	s := NewSearchState()
	s.ActiveField = FieldFind
	s.Query = "hello"
	s.QueryCursor = 3

	s.MoveHome()
	if s.QueryCursor != 0 {
		t.Errorf("expected cursor 0, got %d", s.QueryCursor)
	}

	s.MoveEnd()
	if s.QueryCursor != 5 {
		t.Errorf("expected cursor 5, got %d", s.QueryCursor)
	}
}

func TestSearchState_ToggleField(t *testing.T) {
	s := NewSearchState()
	s.ActiveField = FieldFind

	// First toggle should enable replace and switch to it.
	s.ToggleField()
	if !s.ShowReplace {
		t.Error("expected ShowReplace to be true")
	}
	if s.ActiveField != FieldReplace {
		t.Error("expected FieldReplace")
	}

	// Toggle again should go back to find.
	s.ToggleField()
	if s.ActiveField != FieldFind {
		t.Error("expected FieldFind")
	}

	// Toggle once more back to replace.
	s.ToggleField()
	if s.ActiveField != FieldReplace {
		t.Error("expected FieldReplace again")
	}
}

func TestSearchState_ReplaceField_Operations(t *testing.T) {
	s := NewSearchState()
	s.ShowReplace = true
	s.ActiveField = FieldReplace

	s.InsertChar('n')
	s.InsertChar('e')
	s.InsertChar('w')
	if s.ReplaceText != "new" {
		t.Errorf("expected 'new', got %q", s.ReplaceText)
	}

	s.MoveHome()
	if s.ReplaceCursor != 0 {
		t.Errorf("expected cursor 0, got %d", s.ReplaceCursor)
	}

	s.MoveEnd()
	if s.ReplaceCursor != 3 {
		t.Errorf("expected cursor 3, got %d", s.ReplaceCursor)
	}

	s.DeleteBackward()
	if s.ReplaceText != "ne" {
		t.Errorf("expected 'ne', got %q", s.ReplaceText)
	}
}
