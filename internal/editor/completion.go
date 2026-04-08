package editor

import (
	"strings"
)

// CompletionItemKind mirrors LSP CompletionItemKind for use in the editor package.
type CompletionItemKind int

const (
	CKText          CompletionItemKind = 1
	CKMethod        CompletionItemKind = 2
	CKFunction      CompletionItemKind = 3
	CKConstructor   CompletionItemKind = 4
	CKField         CompletionItemKind = 5
	CKVariable      CompletionItemKind = 6
	CKClass         CompletionItemKind = 7
	CKInterface     CompletionItemKind = 8
	CKModule        CompletionItemKind = 9
	CKProperty      CompletionItemKind = 10
	CKUnit          CompletionItemKind = 11
	CKValue         CompletionItemKind = 12
	CKEnum          CompletionItemKind = 13
	CKKeyword       CompletionItemKind = 14
	CKSnippet       CompletionItemKind = 15
	CKConstant      CompletionItemKind = 21
	CKStruct        CompletionItemKind = 22
	CKTypeParameter CompletionItemKind = 25
)

// CompletionItem represents a single completion suggestion in the editor.
type CompletionItem struct {
	Label      string
	Kind       CompletionItemKind
	Detail     string
	InsertText string // Text to insert; falls back to Label if empty.
	FilterText string // Text to match against; falls back to Label if empty.
	SortText   string // Text to sort by; falls back to Label if empty.
}

// InsertString returns the text that should be inserted for this item.
func (ci *CompletionItem) InsertString() string {
	if ci.InsertText != "" {
		return ci.InsertText
	}
	return ci.Label
}

// FilterString returns the text to use for prefix matching.
func (ci *CompletionItem) FilterString() string {
	if ci.FilterText != "" {
		return ci.FilterText
	}
	return ci.Label
}

// KindIcon returns a short icon/label for the completion kind.
func (ci *CompletionItem) KindIcon() string {
	switch ci.Kind {
	case CKFunction, CKMethod:
		return "fn"
	case CKVariable, CKField, CKProperty:
		return "vr"
	case CKClass, CKStruct:
		return "st"
	case CKInterface:
		return "if"
	case CKConstant:
		return "cn"
	case CKEnum:
		return "en"
	case CKModule:
		return "pk"
	case CKKeyword:
		return "kw"
	case CKSnippet:
		return "sn"
	case CKConstructor:
		return "co"
	case CKValue, CKUnit:
		return "vl"
	case CKTypeParameter:
		return "tp"
	default:
		return "  "
	}
}

// CompletionMaxVisible is the maximum number of items shown in the popup.
const CompletionMaxVisible = 10

// CompletionState holds the state of the autocomplete popup.
type CompletionState struct {
	AllItems    []CompletionItem // Full unfiltered list from LSP.
	Filtered    []CompletionItem // Items after prefix filtering.
	SelectedIdx int              // Currently highlighted item in Filtered.
	ScrollY     int              // Scroll offset in the filtered list.
	AnchorRow   int              // Editor row where completion was triggered.
	AnchorCol   int              // Editor byte-offset col where completion was triggered.
	Prefix      string           // Typed prefix since anchor.
}

// NewCompletionState creates a new completion state from LSP items.
func NewCompletionState(items []CompletionItem, anchorRow, anchorCol int, prefix string) *CompletionState {
	cs := &CompletionState{
		AllItems:  items,
		AnchorRow: anchorRow,
		AnchorCol: anchorCol,
		Prefix:    prefix,
	}
	cs.Refilter()
	return cs
}

// Refilter filters AllItems based on the current Prefix (case-insensitive prefix match).
func (cs *CompletionState) Refilter() {
	cs.Filtered = cs.Filtered[:0]
	lowerPrefix := strings.ToLower(cs.Prefix)
	for _, item := range cs.AllItems {
		filterText := strings.ToLower(item.FilterString())
		if strings.HasPrefix(filterText, lowerPrefix) {
			cs.Filtered = append(cs.Filtered, item)
		}
	}
	// Clamp selected index.
	if cs.SelectedIdx >= len(cs.Filtered) {
		cs.SelectedIdx = len(cs.Filtered) - 1
	}
	if cs.SelectedIdx < 0 {
		cs.SelectedIdx = 0
	}
	cs.ensureSelectedVisible()
}

// UpdatePrefix sets a new prefix and refilters.
func (cs *CompletionState) UpdatePrefix(prefix string) {
	cs.Prefix = prefix
	cs.Refilter()
}

// MoveUp moves the selection up by one.
func (cs *CompletionState) MoveUp() {
	if len(cs.Filtered) == 0 {
		return
	}
	cs.SelectedIdx--
	if cs.SelectedIdx < 0 {
		cs.SelectedIdx = len(cs.Filtered) - 1
		// Scroll to show the last item.
		cs.ScrollY = len(cs.Filtered) - CompletionMaxVisible
		if cs.ScrollY < 0 {
			cs.ScrollY = 0
		}
		return
	}
	cs.ensureSelectedVisible()
}

// MoveDown moves the selection down by one.
func (cs *CompletionState) MoveDown() {
	if len(cs.Filtered) == 0 {
		return
	}
	cs.SelectedIdx++
	if cs.SelectedIdx >= len(cs.Filtered) {
		cs.SelectedIdx = 0
		cs.ScrollY = 0
		return
	}
	cs.ensureSelectedVisible()
}

// SelectedItem returns the currently selected item, or nil if none.
func (cs *CompletionState) SelectedItem() *CompletionItem {
	if len(cs.Filtered) == 0 || cs.SelectedIdx < 0 || cs.SelectedIdx >= len(cs.Filtered) {
		return nil
	}
	return &cs.Filtered[cs.SelectedIdx]
}

// VisibleItems returns the slice of items currently visible in the popup.
func (cs *CompletionState) VisibleItems() []CompletionItem {
	if len(cs.Filtered) == 0 {
		return nil
	}
	start := cs.ScrollY
	end := start + CompletionMaxVisible
	if end > len(cs.Filtered) {
		end = len(cs.Filtered)
	}
	if start >= end {
		return nil
	}
	return cs.Filtered[start:end]
}

// VisibleSelectedIdx returns the index of the selected item within the visible items.
func (cs *CompletionState) VisibleSelectedIdx() int {
	return cs.SelectedIdx - cs.ScrollY
}

// IsEmpty returns true if there are no filtered items.
func (cs *CompletionState) IsEmpty() bool {
	return len(cs.Filtered) == 0
}

// ensureSelectedVisible adjusts ScrollY so the selected item is visible.
func (cs *CompletionState) ensureSelectedVisible() {
	if cs.SelectedIdx < cs.ScrollY {
		cs.ScrollY = cs.SelectedIdx
	}
	if cs.SelectedIdx >= cs.ScrollY+CompletionMaxVisible {
		cs.ScrollY = cs.SelectedIdx - CompletionMaxVisible + 1
	}
	if cs.ScrollY < 0 {
		cs.ScrollY = 0
	}
}
