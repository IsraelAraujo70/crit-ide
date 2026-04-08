package editor

import "unicode/utf8"

// ProjectSearchState holds the state for the project-wide search panel.
type ProjectSearchState struct {
	Query       string // Search query typed by the user.
	CursorPos   int    // Byte offset cursor within Query.
	SelectedIdx int    // Currently selected entry index in the flat list (0-based).
	ScrollY     int    // Scroll offset in the results list.

	// Entries are populated externally by the action that runs the search.
	Entries    []ProjectSearchEntry
	TotalFiles int // Number of files with matches.
	TotalHits  int // Total number of matches.
	Searching  bool // True while search is in progress.
}

// ProjectSearchEntry represents a single line in the project search results panel.
type ProjectSearchEntry struct {
	IsHeader  bool   // True = file header, False = result line.
	Text      string // Display text.
	Path      string // Absolute path for navigation.
	Line      int    // 1-based line number (0 for headers).
	Col       int    // 1-based column number (0 for headers).
}

// NewProjectSearchState creates a new empty project search state.
func NewProjectSearchState() *ProjectSearchState {
	return &ProjectSearchState{}
}

// InsertChar inserts a character at the cursor position.
func (ps *ProjectSearchState) InsertChar(ch rune) {
	c := string(ch)
	ps.Query = ps.Query[:ps.CursorPos] + c + ps.Query[ps.CursorPos:]
	ps.CursorPos += len(c)
}

// DeleteBackward removes the character before the cursor.
func (ps *ProjectSearchState) DeleteBackward() {
	if ps.CursorPos > 0 {
		prev := ps.CursorPos - 1
		for prev > 0 && ps.Query[prev]>>6 == 2 {
			prev--
		}
		ps.Query = ps.Query[:prev] + ps.Query[ps.CursorPos:]
		ps.CursorPos = prev
	}
}

// DeleteForward removes the character at the cursor.
func (ps *ProjectSearchState) DeleteForward() {
	if ps.CursorPos < len(ps.Query) {
		next := ps.CursorPos + 1
		for next < len(ps.Query) && ps.Query[next]>>6 == 2 {
			next++
		}
		ps.Query = ps.Query[:ps.CursorPos] + ps.Query[next:]
	}
}

// MoveLeft moves the cursor left one character.
func (ps *ProjectSearchState) MoveLeft() {
	if ps.CursorPos > 0 {
		_, size := utf8.DecodeLastRuneInString(ps.Query[:ps.CursorPos])
		ps.CursorPos -= size
	}
}

// MoveRight moves the cursor right one character.
func (ps *ProjectSearchState) MoveRight() {
	if ps.CursorPos < len(ps.Query) {
		_, size := utf8.DecodeRuneInString(ps.Query[ps.CursorPos:])
		ps.CursorPos += size
	}
}

// MoveHome moves the cursor to the start.
func (ps *ProjectSearchState) MoveHome() {
	ps.CursorPos = 0
}

// MoveEnd moves the cursor to the end.
func (ps *ProjectSearchState) MoveEnd() {
	ps.CursorPos = len(ps.Query)
}

// MoveUp moves the selection up by one (skipping no entries).
func (ps *ProjectSearchState) MoveUp() {
	if ps.SelectedIdx > 0 {
		ps.SelectedIdx--
		if ps.SelectedIdx < ps.ScrollY {
			ps.ScrollY = ps.SelectedIdx
		}
	}
}

// MoveDown moves the selection down by one.
func (ps *ProjectSearchState) MoveDown(maxVisible int) {
	if ps.SelectedIdx < len(ps.Entries)-1 {
		ps.SelectedIdx++
		if ps.SelectedIdx >= ps.ScrollY+maxVisible {
			ps.ScrollY = ps.SelectedIdx - maxVisible + 1
		}
	}
}

// SelectedEntry returns the currently selected entry, or nil if none.
func (ps *ProjectSearchState) SelectedEntry() *ProjectSearchEntry {
	if ps.SelectedIdx >= 0 && ps.SelectedIdx < len(ps.Entries) {
		return &ps.Entries[ps.SelectedIdx]
	}
	return nil
}

// EntryCount returns the number of entries.
func (ps *ProjectSearchState) EntryCount() int {
	return len(ps.Entries)
}

// HasResults returns true if there are search results.
func (ps *ProjectSearchState) HasResults() bool {
	return len(ps.Entries) > 0
}
