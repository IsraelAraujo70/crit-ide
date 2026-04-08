package editor

import "testing"

func TestNewPaletteState(t *testing.T) {
	entries := []PaletteEntry{
		{ID: "file.save", Label: "Save File", Keybinding: "Ctrl+S", Category: "File"},
		{ID: "edit.undo", Label: "Undo", Keybinding: "Ctrl+Z", Category: "Edit"},
	}
	ps := NewPaletteState(entries)

	if ps == nil {
		t.Fatal("expected non-nil PaletteState")
	}
	if len(ps.AllEntries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(ps.AllEntries))
	}
	if len(ps.Filtered) != 2 {
		t.Fatalf("expected 2 filtered, got %d", len(ps.Filtered))
	}
	if ps.Query != "" {
		t.Fatalf("expected empty query, got %q", ps.Query)
	}
}

func TestPaletteInsertChar(t *testing.T) {
	ps := NewPaletteState([]PaletteEntry{
		{ID: "file.save", Label: "Save File", Category: "File"},
		{ID: "edit.undo", Label: "Undo", Category: "Edit"},
		{ID: "app.quit", Label: "Quit Application", Category: "File"},
	})

	ps.InsertChar('s')
	if ps.Query != "s" {
		t.Fatalf("expected query 's', got %q", ps.Query)
	}
	if ps.CursorPos != 1 {
		t.Fatalf("expected cursor at 1, got %d", ps.CursorPos)
	}
	// "Save File" should match, "Undo" should not.
	found := false
	for _, e := range ps.Filtered {
		if e.ID == "file.save" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected 'file.save' in filtered results")
	}
}

func TestPaletteDeleteBackward(t *testing.T) {
	ps := NewPaletteState([]PaletteEntry{
		{ID: "file.save", Label: "Save File", Category: "File"},
	})
	ps.InsertChar('x')
	ps.InsertChar('y')
	ps.DeleteBackward()
	if ps.Query != "x" {
		t.Fatalf("expected query 'x', got %q", ps.Query)
	}
	if ps.CursorPos != 1 {
		t.Fatalf("expected cursor at 1, got %d", ps.CursorPos)
	}
}

func TestPaletteDeleteForward(t *testing.T) {
	ps := NewPaletteState([]PaletteEntry{
		{ID: "file.save", Label: "Save File", Category: "File"},
	})
	ps.InsertChar('a')
	ps.InsertChar('b')
	ps.MoveHome()
	ps.DeleteForward()
	if ps.Query != "b" {
		t.Fatalf("expected query 'b', got %q", ps.Query)
	}
}

func TestPaletteMoveLeftRight(t *testing.T) {
	ps := NewPaletteState(nil)
	ps.InsertChar('a')
	ps.InsertChar('b')

	ps.MoveLeft()
	if ps.CursorPos != 1 {
		t.Fatalf("expected cursor at 1 after MoveLeft, got %d", ps.CursorPos)
	}

	ps.MoveRight()
	if ps.CursorPos != 2 {
		t.Fatalf("expected cursor at 2 after MoveRight, got %d", ps.CursorPos)
	}

	// MoveRight at end should be no-op.
	ps.MoveRight()
	if ps.CursorPos != 2 {
		t.Fatalf("expected cursor still at 2, got %d", ps.CursorPos)
	}
}

func TestPaletteMoveHomeEnd(t *testing.T) {
	ps := NewPaletteState(nil)
	ps.InsertChar('a')
	ps.InsertChar('b')
	ps.InsertChar('c')

	ps.MoveHome()
	if ps.CursorPos != 0 {
		t.Fatalf("expected cursor at 0 after MoveHome, got %d", ps.CursorPos)
	}

	ps.MoveEnd()
	if ps.CursorPos != 3 {
		t.Fatalf("expected cursor at 3 after MoveEnd, got %d", ps.CursorPos)
	}
}

func TestPaletteMoveUpDown(t *testing.T) {
	entries := []PaletteEntry{
		{ID: "a", Label: "Alpha", Category: "Cat"},
		{ID: "b", Label: "Beta", Category: "Cat"},
		{ID: "c", Label: "Charlie", Category: "Cat"},
	}
	ps := NewPaletteState(entries)

	if ps.SelectedIdx != 0 {
		t.Fatalf("expected selected 0, got %d", ps.SelectedIdx)
	}

	ps.MoveDown(15)
	if ps.SelectedIdx != 1 {
		t.Fatalf("expected selected 1, got %d", ps.SelectedIdx)
	}

	ps.MoveDown(15)
	if ps.SelectedIdx != 2 {
		t.Fatalf("expected selected 2, got %d", ps.SelectedIdx)
	}

	// MoveDown at bottom should be no-op.
	ps.MoveDown(15)
	if ps.SelectedIdx != 2 {
		t.Fatalf("expected selected still 2, got %d", ps.SelectedIdx)
	}

	ps.MoveUp()
	if ps.SelectedIdx != 1 {
		t.Fatalf("expected selected 1 after MoveUp, got %d", ps.SelectedIdx)
	}

	ps.MoveUp()
	if ps.SelectedIdx != 0 {
		t.Fatalf("expected selected 0 after MoveUp, got %d", ps.SelectedIdx)
	}

	// MoveUp at top should be no-op.
	ps.MoveUp()
	if ps.SelectedIdx != 0 {
		t.Fatalf("expected selected still 0, got %d", ps.SelectedIdx)
	}
}

func TestPaletteSelectedEntry(t *testing.T) {
	entries := []PaletteEntry{
		{ID: "a", Label: "Alpha", Category: "Cat"},
		{ID: "b", Label: "Beta", Category: "Cat"},
	}
	ps := NewPaletteState(entries)

	entry := ps.SelectedEntry()
	if entry == nil || entry.ID != "a" {
		t.Fatalf("expected entry 'a', got %v", entry)
	}

	ps.MoveDown(15)
	entry = ps.SelectedEntry()
	if entry == nil || entry.ID != "b" {
		t.Fatalf("expected entry 'b', got %v", entry)
	}
}

func TestPaletteEmptySelectedEntry(t *testing.T) {
	ps := NewPaletteState(nil)
	entry := ps.SelectedEntry()
	if entry != nil {
		t.Fatal("expected nil entry from empty palette")
	}
}

func TestPaletteFuzzyFilter(t *testing.T) {
	entries := []PaletteEntry{
		{ID: "file.save", Label: "Save File", Category: "File"},
		{ID: "edit.undo", Label: "Undo", Category: "Edit"},
		{ID: "app.quit", Label: "Quit Application", Category: "File"},
		{ID: "search.open", Label: "Find / Replace", Category: "Search"},
	}
	ps := NewPaletteState(entries)

	// Type "save" — should filter to Save File.
	for _, ch := range "save" {
		ps.InsertChar(ch)
	}
	if ps.ResultCount() == 0 {
		t.Fatal("expected at least 1 result for 'save'")
	}
	// "Save File" should be the top result.
	if ps.Filtered[0].ID != "file.save" {
		t.Fatalf("expected top result 'file.save', got %q", ps.Filtered[0].ID)
	}
}

func TestPaletteFuzzyFilterNoMatch(t *testing.T) {
	entries := []PaletteEntry{
		{ID: "file.save", Label: "Save File", Category: "File"},
	}
	ps := NewPaletteState(entries)

	for _, ch := range "zzzzz" {
		ps.InsertChar(ch)
	}
	if ps.ResultCount() != 0 {
		t.Fatalf("expected 0 results for 'zzzzz', got %d", ps.ResultCount())
	}
}

func TestPaletteFuzzyFilterByCategory(t *testing.T) {
	entries := []PaletteEntry{
		{ID: "file.save", Label: "Save", Category: "File"},
		{ID: "edit.undo", Label: "Undo", Category: "Edit"},
	}
	ps := NewPaletteState(entries)

	for _, ch := range "edit" {
		ps.InsertChar(ch)
	}
	found := false
	for _, e := range ps.Filtered {
		if e.ID == "edit.undo" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected 'edit.undo' in results when filtering by 'edit'")
	}
}

func TestPaletteResultCount(t *testing.T) {
	ps := NewPaletteState([]PaletteEntry{
		{ID: "a", Label: "Alpha", Category: "X"},
		{ID: "b", Label: "Beta", Category: "X"},
		{ID: "c", Label: "Charlie", Category: "X"},
	})

	if ps.ResultCount() != 3 {
		t.Fatalf("expected 3 results, got %d", ps.ResultCount())
	}
}

func TestPaletteScrollOnMoveDown(t *testing.T) {
	entries := make([]PaletteEntry, 20)
	for i := range entries {
		entries[i] = PaletteEntry{ID: string(rune('a' + i)), Label: string(rune('A' + i)), Category: "X"}
	}
	ps := NewPaletteState(entries)

	// Move down past maxVisible (3 for this test).
	maxVisible := 3
	for i := 0; i < 5; i++ {
		ps.MoveDown(maxVisible)
	}

	if ps.SelectedIdx != 5 {
		t.Fatalf("expected selected 5, got %d", ps.SelectedIdx)
	}
	if ps.ScrollY < 3 {
		t.Fatalf("expected scrollY >= 3, got %d", ps.ScrollY)
	}
}

func TestFuzzyScore(t *testing.T) {
	// Exact match should score > 0.
	s := fuzzyScore("save", "save file")
	if s <= 0 {
		t.Fatal("expected positive score for matching pattern")
	}

	// No match should return 0.
	s = fuzzyScore("xyz", "save file")
	if s != 0 {
		t.Fatalf("expected 0 for non-matching pattern, got %d", s)
	}

	// Empty pattern should return 1.
	s = fuzzyScore("", "anything")
	if s != 1 {
		t.Fatalf("expected 1 for empty pattern, got %d", s)
	}
}
