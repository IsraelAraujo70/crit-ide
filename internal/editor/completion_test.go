package editor

import (
	"testing"
)

func makeItems() []CompletionItem {
	return []CompletionItem{
		{Label: "Println", Kind: CKFunction, Detail: "func(a ...any)", InsertText: "Println"},
		{Label: "Printf", Kind: CKFunction, Detail: "func(format string, a ...any)", InsertText: "Printf"},
		{Label: "Print", Kind: CKFunction, Detail: "func(a ...any)", InsertText: "Print"},
		{Label: "Sprintf", Kind: CKFunction, Detail: "func(format string, a ...any) string", InsertText: "Sprintf"},
		{Label: "Sprint", Kind: CKFunction, Detail: "func(a ...any) string", InsertText: "Sprint"},
		{Label: "Fprintln", Kind: CKFunction, Detail: "func(w io.Writer, a ...any)", InsertText: "Fprintln"},
		{Label: "Fprintf", Kind: CKFunction, Detail: "func(w io.Writer, format string, a ...any)", InsertText: "Fprintf"},
		{Label: "Fprint", Kind: CKFunction, Detail: "func(w io.Writer, a ...any)", InsertText: "Fprint"},
		{Label: "Errorf", Kind: CKFunction, Detail: "func(format string, a ...any) error", InsertText: "Errorf"},
		{Label: "Stringer", Kind: CKInterface, Detail: "interface", InsertText: "Stringer"},
		{Label: "Formatter", Kind: CKInterface, Detail: "interface", InsertText: "Formatter"},
		{Label: "GoStringer", Kind: CKInterface, Detail: "interface", InsertText: "GoStringer"},
	}
}

func TestNewCompletionState(t *testing.T) {
	items := makeItems()
	cs := NewCompletionState(items, 5, 10, "")

	if len(cs.AllItems) != 12 {
		t.Errorf("AllItems: got %d, want 12", len(cs.AllItems))
	}
	if len(cs.Filtered) != 12 {
		t.Errorf("Filtered: got %d, want 12 (empty prefix matches all)", len(cs.Filtered))
	}
	if cs.SelectedIdx != 0 {
		t.Errorf("SelectedIdx: got %d, want 0", cs.SelectedIdx)
	}
	if cs.AnchorRow != 5 {
		t.Errorf("AnchorRow: got %d, want 5", cs.AnchorRow)
	}
	if cs.AnchorCol != 10 {
		t.Errorf("AnchorCol: got %d, want 10", cs.AnchorCol)
	}
}

func TestCompletionPrefixFilter(t *testing.T) {
	items := makeItems()
	cs := NewCompletionState(items, 0, 0, "Pr")

	// "Pr" should match: Println, Printf, Print
	if len(cs.Filtered) != 3 {
		t.Errorf("Filtered 'Pr': got %d, want 3", len(cs.Filtered))
	}
	for _, item := range cs.Filtered {
		if item.Label != "Println" && item.Label != "Printf" && item.Label != "Print" {
			t.Errorf("unexpected item in filtered: %s", item.Label)
		}
	}
}

func TestCompletionPrefixFilterCaseInsensitive(t *testing.T) {
	items := makeItems()
	cs := NewCompletionState(items, 0, 0, "pr")

	// Case-insensitive: "pr" should match same as "Pr"
	if len(cs.Filtered) != 3 {
		t.Errorf("Filtered 'pr': got %d, want 3", len(cs.Filtered))
	}
}

func TestCompletionUpdatePrefix(t *testing.T) {
	items := makeItems()
	cs := NewCompletionState(items, 0, 0, "")

	cs.UpdatePrefix("S")
	if len(cs.Filtered) != 3 {
		t.Errorf("Filtered 'S': got %d, want 3 (Sprintf, Sprint, Stringer)", len(cs.Filtered))
	}

	cs.UpdatePrefix("Sp")
	if len(cs.Filtered) != 2 {
		t.Errorf("Filtered 'Sp': got %d, want 2 (Sprintf, Sprint)", len(cs.Filtered))
	}

	cs.UpdatePrefix("Spr")
	if len(cs.Filtered) != 2 {
		t.Errorf("Filtered 'Spr': got %d, want 2", len(cs.Filtered))
	}

	cs.UpdatePrefix("Sprint")
	// "Sprint" prefix matches both "Sprintf" and "Sprint".
	if len(cs.Filtered) != 2 {
		t.Errorf("Filtered 'Sprint': got %d, want 2 (Sprintf, Sprint)", len(cs.Filtered))
	}

	cs.UpdatePrefix("Sprintl")
	// No item starts with "Sprintl".
	if len(cs.Filtered) != 0 {
		t.Errorf("Filtered 'Sprintl': got %d, want 0", len(cs.Filtered))
	}
}

func TestCompletionNoMatch(t *testing.T) {
	items := makeItems()
	cs := NewCompletionState(items, 0, 0, "xyz")

	if !cs.IsEmpty() {
		t.Errorf("expected empty for prefix 'xyz', got %d items", len(cs.Filtered))
	}
}

func TestCompletionMoveDown(t *testing.T) {
	items := makeItems()
	cs := NewCompletionState(items, 0, 0, "")

	if cs.SelectedIdx != 0 {
		t.Fatalf("initial SelectedIdx: got %d, want 0", cs.SelectedIdx)
	}

	cs.MoveDown()
	if cs.SelectedIdx != 1 {
		t.Errorf("after MoveDown: got %d, want 1", cs.SelectedIdx)
	}

	// Move to the last item and then wrap around.
	for i := 0; i < len(cs.Filtered)-1; i++ {
		cs.MoveDown()
	}
	if cs.SelectedIdx != 0 {
		t.Errorf("after wrapping down: got %d, want 0", cs.SelectedIdx)
	}
}

func TestCompletionMoveUp(t *testing.T) {
	items := makeItems()
	cs := NewCompletionState(items, 0, 0, "")

	cs.MoveUp()
	// Should wrap to last item.
	if cs.SelectedIdx != len(cs.Filtered)-1 {
		t.Errorf("after MoveUp from 0: got %d, want %d", cs.SelectedIdx, len(cs.Filtered)-1)
	}

	cs.MoveUp()
	if cs.SelectedIdx != len(cs.Filtered)-2 {
		t.Errorf("after second MoveUp: got %d, want %d", cs.SelectedIdx, len(cs.Filtered)-2)
	}
}

func TestCompletionSelectedItem(t *testing.T) {
	items := makeItems()
	cs := NewCompletionState(items, 0, 0, "")

	item := cs.SelectedItem()
	if item == nil {
		t.Fatal("SelectedItem should not be nil")
	}
	if item.Label != "Println" {
		t.Errorf("SelectedItem: got %s, want Println", item.Label)
	}

	cs.MoveDown()
	cs.MoveDown()
	item = cs.SelectedItem()
	if item.Label != "Print" {
		t.Errorf("SelectedItem after 2 downs: got %s, want Print", item.Label)
	}
}

func TestCompletionSelectedItemEmpty(t *testing.T) {
	cs := NewCompletionState(nil, 0, 0, "")
	item := cs.SelectedItem()
	if item != nil {
		t.Error("SelectedItem should be nil for empty state")
	}
}

func TestCompletionVisibleItems(t *testing.T) {
	items := makeItems()
	cs := NewCompletionState(items, 0, 0, "")

	visible := cs.VisibleItems()
	expected := CompletionMaxVisible
	if len(items) < expected {
		expected = len(items)
	}
	if len(visible) != expected {
		t.Errorf("VisibleItems: got %d, want %d", len(visible), expected)
	}
}

func TestCompletionScrolling(t *testing.T) {
	items := makeItems()
	cs := NewCompletionState(items, 0, 0, "")

	// With 12 items and max visible 10, scrolling should work.
	// Move to item 11 (index 10).
	for i := 0; i < 10; i++ {
		cs.MoveDown()
	}

	if cs.ScrollY == 0 {
		t.Error("ScrollY should have advanced after moving past visible area")
	}

	visible := cs.VisibleItems()
	if len(visible) != CompletionMaxVisible {
		t.Errorf("visible items after scroll: got %d, want %d", len(visible), CompletionMaxVisible)
	}
}

func TestCompletionInsertString(t *testing.T) {
	item := CompletionItem{Label: "Println", InsertText: "Println("}
	if item.InsertString() != "Println(" {
		t.Errorf("InsertString: got %s, want Println(", item.InsertString())
	}

	// Falls back to Label when InsertText is empty.
	item2 := CompletionItem{Label: "Println"}
	if item2.InsertString() != "Println" {
		t.Errorf("InsertString fallback: got %s, want Println", item2.InsertString())
	}
}

func TestCompletionFilterString(t *testing.T) {
	item := CompletionItem{Label: "Println", FilterText: "printline"}
	if item.FilterString() != "printline" {
		t.Errorf("FilterString: got %s, want printline", item.FilterString())
	}

	item2 := CompletionItem{Label: "Println"}
	if item2.FilterString() != "Println" {
		t.Errorf("FilterString fallback: got %s, want Println", item2.FilterString())
	}
}

func TestCompletionKindIcon(t *testing.T) {
	tests := []struct {
		kind CompletionItemKind
		icon string
	}{
		{CKFunction, "fn"},
		{CKMethod, "fn"},
		{CKVariable, "vr"},
		{CKField, "vr"},
		{CKStruct, "st"},
		{CKClass, "st"},
		{CKInterface, "if"},
		{CKConstant, "cn"},
		{CKModule, "pk"},
		{CKKeyword, "kw"},
		{CKSnippet, "sn"},
		{CKText, "  "},
	}

	for _, tt := range tests {
		item := CompletionItem{Kind: tt.kind}
		if got := item.KindIcon(); got != tt.icon {
			t.Errorf("KindIcon(%d): got %q, want %q", tt.kind, got, tt.icon)
		}
	}
}

func TestCompletionRefilterClampsSelectedIdx(t *testing.T) {
	items := makeItems()
	cs := NewCompletionState(items, 0, 0, "")

	// Move to item 5.
	for i := 0; i < 5; i++ {
		cs.MoveDown()
	}
	if cs.SelectedIdx != 5 {
		t.Fatalf("SelectedIdx: got %d, want 5", cs.SelectedIdx)
	}

	// Now filter to only 2 items — selectedIdx should clamp.
	cs.UpdatePrefix("Sp")
	if cs.SelectedIdx >= len(cs.Filtered) {
		t.Errorf("SelectedIdx after filter: got %d, want < %d", cs.SelectedIdx, len(cs.Filtered))
	}
}

func TestCompletionVisibleSelectedIdx(t *testing.T) {
	items := makeItems()
	cs := NewCompletionState(items, 0, 0, "")

	// Initially, visible selected should be 0.
	if cs.VisibleSelectedIdx() != 0 {
		t.Errorf("VisibleSelectedIdx: got %d, want 0", cs.VisibleSelectedIdx())
	}

	// Move down a few.
	cs.MoveDown()
	cs.MoveDown()
	if cs.VisibleSelectedIdx() != 2 {
		t.Errorf("VisibleSelectedIdx after 2 downs: got %d, want 2", cs.VisibleSelectedIdx())
	}
}
