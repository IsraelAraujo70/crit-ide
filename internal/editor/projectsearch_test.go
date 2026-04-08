package editor

import "testing"

func TestProjectSearchState_InsertChar(t *testing.T) {
	ps := NewProjectSearchState()
	ps.InsertChar('h')
	ps.InsertChar('i')
	if ps.Query != "hi" {
		t.Errorf("expected query %q, got %q", "hi", ps.Query)
	}
	if ps.CursorPos != 2 {
		t.Errorf("expected cursor at 2, got %d", ps.CursorPos)
	}
}

func TestProjectSearchState_DeleteBackward(t *testing.T) {
	ps := NewProjectSearchState()
	ps.Query = "hello"
	ps.CursorPos = 5
	ps.DeleteBackward()
	if ps.Query != "hell" {
		t.Errorf("expected %q, got %q", "hell", ps.Query)
	}
	if ps.CursorPos != 4 {
		t.Errorf("expected cursor at 4, got %d", ps.CursorPos)
	}

	// Delete at beginning should be no-op.
	ps.CursorPos = 0
	ps.DeleteBackward()
	if ps.Query != "hell" {
		t.Errorf("expected %q after delete at start, got %q", "hell", ps.Query)
	}
}

func TestProjectSearchState_DeleteForward(t *testing.T) {
	ps := NewProjectSearchState()
	ps.Query = "hello"
	ps.CursorPos = 0
	ps.DeleteForward()
	if ps.Query != "ello" {
		t.Errorf("expected %q, got %q", "ello", ps.Query)
	}

	// Delete at end should be no-op.
	ps.CursorPos = len(ps.Query)
	ps.DeleteForward()
	if ps.Query != "ello" {
		t.Errorf("expected %q, got %q", "ello", ps.Query)
	}
}

func TestProjectSearchState_MoveLeftRight(t *testing.T) {
	ps := NewProjectSearchState()
	ps.Query = "abc"
	ps.CursorPos = 3

	ps.MoveLeft()
	if ps.CursorPos != 2 {
		t.Errorf("after MoveLeft: expected 2, got %d", ps.CursorPos)
	}

	ps.MoveRight()
	if ps.CursorPos != 3 {
		t.Errorf("after MoveRight: expected 3, got %d", ps.CursorPos)
	}

	// Left at 0 should stay.
	ps.CursorPos = 0
	ps.MoveLeft()
	if ps.CursorPos != 0 {
		t.Errorf("MoveLeft at 0: expected 0, got %d", ps.CursorPos)
	}

	// Right at end should stay.
	ps.CursorPos = len(ps.Query)
	ps.MoveRight()
	if ps.CursorPos != 3 {
		t.Errorf("MoveRight at end: expected 3, got %d", ps.CursorPos)
	}
}

func TestProjectSearchState_HomeEnd(t *testing.T) {
	ps := NewProjectSearchState()
	ps.Query = "hello"
	ps.CursorPos = 3

	ps.MoveHome()
	if ps.CursorPos != 0 {
		t.Errorf("after MoveHome: expected 0, got %d", ps.CursorPos)
	}

	ps.MoveEnd()
	if ps.CursorPos != 5 {
		t.Errorf("after MoveEnd: expected 5, got %d", ps.CursorPos)
	}
}

func TestProjectSearchState_MoveUpDown(t *testing.T) {
	ps := NewProjectSearchState()
	ps.Entries = []ProjectSearchEntry{
		{IsHeader: true, Text: "file.go"},
		{IsHeader: false, Text: "  1: match"},
		{IsHeader: false, Text: "  2: match"},
		{IsHeader: true, Text: "other.go"},
		{IsHeader: false, Text: "  5: match"},
	}
	ps.SelectedIdx = 0

	ps.MoveDown(20)
	if ps.SelectedIdx != 1 {
		t.Errorf("after MoveDown: expected 1, got %d", ps.SelectedIdx)
	}

	ps.MoveDown(20)
	if ps.SelectedIdx != 2 {
		t.Errorf("after MoveDown: expected 2, got %d", ps.SelectedIdx)
	}

	ps.MoveUp()
	if ps.SelectedIdx != 1 {
		t.Errorf("after MoveUp: expected 1, got %d", ps.SelectedIdx)
	}

	// MoveUp at 0 should stay.
	ps.SelectedIdx = 0
	ps.MoveUp()
	if ps.SelectedIdx != 0 {
		t.Errorf("MoveUp at 0: expected 0, got %d", ps.SelectedIdx)
	}
}

func TestProjectSearchState_SelectedEntry(t *testing.T) {
	ps := NewProjectSearchState()

	// No entries.
	if ps.SelectedEntry() != nil {
		t.Error("expected nil for empty entries")
	}

	ps.Entries = []ProjectSearchEntry{
		{IsHeader: true, Text: "header"},
		{IsHeader: false, Text: "result", Path: "/a.go", Line: 5},
	}
	ps.SelectedIdx = 1

	entry := ps.SelectedEntry()
	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if entry.Line != 5 {
		t.Errorf("expected line 5, got %d", entry.Line)
	}
}

func TestProjectSearchState_HasResults(t *testing.T) {
	ps := NewProjectSearchState()
	if ps.HasResults() {
		t.Error("expected no results for empty state")
	}

	ps.Entries = []ProjectSearchEntry{{Text: "test"}}
	if !ps.HasResults() {
		t.Error("expected HasResults true after adding entry")
	}
}

func TestProjectSearchState_ScrollVisible(t *testing.T) {
	ps := NewProjectSearchState()
	// Create 25 entries.
	for i := 0; i < 25; i++ {
		ps.Entries = append(ps.Entries, ProjectSearchEntry{Text: "entry"})
	}
	ps.SelectedIdx = 0

	// Move down with maxVisible=5 should trigger scroll.
	for i := 0; i < 6; i++ {
		ps.MoveDown(5)
	}
	if ps.ScrollY == 0 {
		t.Error("expected scroll to advance when selection goes past maxVisible")
	}
	if ps.SelectedIdx != 6 {
		t.Errorf("expected selectedIdx 6, got %d", ps.SelectedIdx)
	}
}
