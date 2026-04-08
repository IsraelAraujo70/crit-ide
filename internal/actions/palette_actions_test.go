package actions

import (
	"testing"

	"github.com/israelcorrea/crit-ide/internal/editor"
)

func TestDefaultPaletteEntries(t *testing.T) {
	entries := DefaultPaletteEntries()
	if len(entries) == 0 {
		t.Fatal("expected non-empty palette entries")
	}

	// Verify all entries have required fields.
	for i, e := range entries {
		if e.ID == "" {
			t.Fatalf("entry %d has empty ID", i)
		}
		if e.Label == "" {
			t.Fatalf("entry %d (%s) has empty Label", i, e.ID)
		}
		if e.Category == "" {
			t.Fatalf("entry %d (%s) has empty Category", i, e.ID)
		}
	}
}

func TestDefaultPaletteEntriesHasExpectedCategories(t *testing.T) {
	entries := DefaultPaletteEntries()
	categories := make(map[string]bool)
	for _, e := range entries {
		categories[e.Category] = true
	}

	expected := []string{"File", "Edit", "View", "Search", "LSP", "Navigate"}
	for _, cat := range expected {
		if !categories[cat] {
			t.Errorf("expected category %q in palette entries", cat)
		}
	}
}

func TestDefaultPaletteEntriesHasKeyActions(t *testing.T) {
	entries := DefaultPaletteEntries()
	ids := make(map[string]bool)
	for _, e := range entries {
		ids[e.ID] = true
	}

	// Check key actions are present.
	required := []string{"file.save", "app.quit", "edit.undo", "edit.redo",
		"clipboard.copy", "clipboard.paste", "search.open", "lsp.definition"}
	for _, id := range required {
		if !ids[id] {
			t.Errorf("expected action %q in palette entries", id)
		}
	}
}

func TestRegisterPaletteActionsRegistered(t *testing.T) {
	r := NewRegistry()
	RegisterPaletteActions(r)

	// Verify all expected action IDs are registered.
	expectedIDs := []string{
		"palette.open", "palette.close", "palette.char", "palette.backspace",
		"palette.delete", "palette.left", "palette.right", "palette.home",
		"palette.end", "palette.up", "palette.down", "palette.execute",
	}
	for _, id := range expectedIDs {
		if _, ok := r.actions[id]; !ok {
			t.Errorf("action %q not registered", id)
		}
	}

	// Verify correct count.
	if len(r.actions) != len(expectedIDs) {
		t.Errorf("expected %d actions, got %d", len(expectedIDs), len(r.actions))
	}
}

func TestPaletteMaxVisible(t *testing.T) {
	if paletteMaxVisible != 15 {
		t.Fatalf("expected paletteMaxVisible == 15, got %d", paletteMaxVisible)
	}
}

func TestNewPaletteStateIntegration(t *testing.T) {
	entries := DefaultPaletteEntries()
	ps := editor.NewPaletteState(entries)

	if ps.ResultCount() != len(entries) {
		t.Fatalf("expected %d results, got %d", len(entries), ps.ResultCount())
	}

	// Filter by "save".
	for _, ch := range "save" {
		ps.InsertChar(ch)
	}

	if ps.ResultCount() == 0 {
		t.Fatal("expected results for 'save' filter")
	}

	entry := ps.SelectedEntry()
	if entry == nil {
		t.Fatal("expected non-nil selected entry")
	}
	if entry.ID != "file.save" {
		t.Fatalf("expected top result 'file.save', got %q", entry.ID)
	}
}
