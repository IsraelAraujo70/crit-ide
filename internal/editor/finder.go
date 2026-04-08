package editor

import "unicode/utf8"

// FinderState holds the state for the fuzzy file finder popup.
type FinderState struct {
	Query       string // Search query typed by the user.
	CursorPos   int    // Byte offset cursor within Query.
	SelectedIdx int    // Currently selected result index (0-based).
	ScrollY     int    // Scroll offset in the results list.

	// Results are populated externally by the action that manages filtering.
	Results    []FinderResult
	TotalFiles int // Total number of indexed files (shown in header).
}

// FinderResult represents a single entry in the finder results list.
type FinderResult struct {
	RelPath string // Relative path to display.
	AbsPath string // Absolute path for opening.
	Matches []int  // Character indices that matched (for highlighting).
}

// NewFinderState creates a new empty finder state.
func NewFinderState() *FinderState {
	return &FinderState{}
}

// InsertChar inserts a character at the cursor position.
func (f *FinderState) InsertChar(ch rune) {
	c := string(ch)
	f.Query = f.Query[:f.CursorPos] + c + f.Query[f.CursorPos:]
	f.CursorPos += len(c)
	// Reset selection when query changes.
	f.SelectedIdx = 0
	f.ScrollY = 0
}

// DeleteBackward removes the character before the cursor.
func (f *FinderState) DeleteBackward() {
	if f.CursorPos > 0 {
		prev := f.CursorPos - 1
		for prev > 0 && f.Query[prev]>>6 == 2 {
			prev--
		}
		f.Query = f.Query[:prev] + f.Query[f.CursorPos:]
		f.CursorPos = prev
		f.SelectedIdx = 0
		f.ScrollY = 0
	}
}

// DeleteForward removes the character at the cursor.
func (f *FinderState) DeleteForward() {
	if f.CursorPos < len(f.Query) {
		next := f.CursorPos + 1
		for next < len(f.Query) && f.Query[next]>>6 == 2 {
			next++
		}
		f.Query = f.Query[:f.CursorPos] + f.Query[next:]
		f.SelectedIdx = 0
		f.ScrollY = 0
	}
}

// MoveLeft moves the cursor left one character.
func (f *FinderState) MoveLeft() {
	if f.CursorPos > 0 {
		_, size := utf8.DecodeLastRuneInString(f.Query[:f.CursorPos])
		f.CursorPos -= size
	}
}

// MoveRight moves the cursor right one character.
func (f *FinderState) MoveRight() {
	if f.CursorPos < len(f.Query) {
		_, size := utf8.DecodeRuneInString(f.Query[f.CursorPos:])
		f.CursorPos += size
	}
}

// MoveHome moves the cursor to the start.
func (f *FinderState) MoveHome() {
	f.CursorPos = 0
}

// MoveEnd moves the cursor to the end.
func (f *FinderState) MoveEnd() {
	f.CursorPos = len(f.Query)
}

// MoveUp moves the selection up by one.
func (f *FinderState) MoveUp() {
	if f.SelectedIdx > 0 {
		f.SelectedIdx--
		if f.SelectedIdx < f.ScrollY {
			f.ScrollY = f.SelectedIdx
		}
	}
}

// MoveDown moves the selection down by one.
func (f *FinderState) MoveDown(maxVisible int) {
	if f.SelectedIdx < len(f.Results)-1 {
		f.SelectedIdx++
		if f.SelectedIdx >= f.ScrollY+maxVisible {
			f.ScrollY = f.SelectedIdx - maxVisible + 1
		}
	}
}

// SelectedPath returns the absolute path of the currently selected result.
// Returns "" if no results are available.
func (f *FinderState) SelectedPath() string {
	if f.SelectedIdx >= 0 && f.SelectedIdx < len(f.Results) {
		return f.Results[f.SelectedIdx].AbsPath
	}
	return ""
}

// ResultCount returns the number of results.
func (f *FinderState) ResultCount() int {
	return len(f.Results)
}
