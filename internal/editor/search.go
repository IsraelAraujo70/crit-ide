package editor

import "strings"

// SearchField indicates which input field is active in the search bar.
type SearchField int

const (
	FieldFind    SearchField = iota // Find query field.
	FieldReplace                    // Replace text field.
)

// SearchState holds the state for Find/Replace operations.
type SearchState struct {
	Query         string      // Search query.
	ReplaceText   string      // Replacement text.
	Matches       []Range     // All matches in the buffer.
	CurrentIdx    int         // Index into Matches (-1 if no matches).
	ShowReplace   bool        // Whether the replace field is visible.
	ActiveField   SearchField // Which input field is active.
	QueryCursor   int         // Byte offset cursor within Query.
	ReplaceCursor int         // Byte offset cursor within ReplaceText.
}

// NewSearchState creates a new empty search state.
func NewSearchState() *SearchState {
	return &SearchState{CurrentIdx: -1}
}

// FindAll scans the entire TextStore for occurrences of Query and populates Matches.
// Handles multi-line content by searching line-by-line (query is assumed single-line).
func (s *SearchState) FindAll(store TextStore) {
	s.Matches = nil
	s.CurrentIdx = -1
	if s.Query == "" {
		return
	}

	for i := 0; i < store.LineCount(); i++ {
		line := store.Line(i)
		offset := 0
		for {
			idx := strings.Index(line[offset:], s.Query)
			if idx < 0 {
				break
			}
			start := offset + idx
			end := start + len(s.Query)
			s.Matches = append(s.Matches, Range{
				Start: Position{i, start},
				End:   Position{i, end},
			})
			offset = end
			if offset >= len(line) {
				break
			}
		}
	}

	if len(s.Matches) > 0 {
		s.CurrentIdx = 0
	}
}

// FindNext advances to the next match after the given cursor position.
// Wraps around to the first match if no later match exists.
func (s *SearchState) FindNext(cursorRow, cursorCol int) (Position, bool) {
	if len(s.Matches) == 0 {
		return Position{}, false
	}

	for i, m := range s.Matches {
		if m.Start.Line > cursorRow || (m.Start.Line == cursorRow && m.Start.Col > cursorCol) {
			s.CurrentIdx = i
			return m.Start, true
		}
	}

	// Wrap around to first match.
	s.CurrentIdx = 0
	return s.Matches[0].Start, true
}

// FindPrev goes to the previous match before the given cursor position.
// Wraps around to the last match if no earlier match exists.
func (s *SearchState) FindPrev(cursorRow, cursorCol int) (Position, bool) {
	if len(s.Matches) == 0 {
		return Position{}, false
	}

	for i := len(s.Matches) - 1; i >= 0; i-- {
		m := s.Matches[i]
		if m.Start.Line < cursorRow || (m.Start.Line == cursorRow && m.Start.Col < cursorCol) {
			s.CurrentIdx = i
			return m.Start, true
		}
	}

	// Wrap around to last match.
	s.CurrentIdx = len(s.Matches) - 1
	return s.Matches[s.CurrentIdx].Start, true
}

// FindNearest sets CurrentIdx to the match closest to the given position.
func (s *SearchState) FindNearest(cursorRow, cursorCol int) {
	if len(s.Matches) == 0 {
		s.CurrentIdx = -1
		return
	}

	for i, m := range s.Matches {
		if m.Start.Line > cursorRow || (m.Start.Line == cursorRow && m.Start.Col >= cursorCol) {
			s.CurrentIdx = i
			return
		}
	}
	s.CurrentIdx = 0
}

// CurrentMatch returns the current match range, if any.
func (s *SearchState) CurrentMatch() (Range, bool) {
	if s.CurrentIdx < 0 || s.CurrentIdx >= len(s.Matches) {
		return Range{}, false
	}
	return s.Matches[s.CurrentIdx], true
}

// MatchCount returns the total number of matches.
func (s *SearchState) MatchCount() int {
	return len(s.Matches)
}

// CurrentMatchNumber returns the 1-based number of the current match (0 if none).
func (s *SearchState) CurrentMatchNumber() int {
	if s.CurrentIdx < 0 || s.CurrentIdx >= len(s.Matches) {
		return 0
	}
	return s.CurrentIdx + 1
}

// InsertChar inserts a character at the cursor of the active field.
func (s *SearchState) InsertChar(ch rune) {
	c := string(ch)
	if s.ActiveField == FieldFind {
		s.Query = s.Query[:s.QueryCursor] + c + s.Query[s.QueryCursor:]
		s.QueryCursor += len(c)
	} else {
		s.ReplaceText = s.ReplaceText[:s.ReplaceCursor] + c + s.ReplaceText[s.ReplaceCursor:]
		s.ReplaceCursor += len(c)
	}
}

// DeleteBackward removes the character before the cursor in the active field.
func (s *SearchState) DeleteBackward() {
	if s.ActiveField == FieldFind {
		if s.QueryCursor > 0 {
			prev := s.QueryCursor - 1
			for prev > 0 && s.Query[prev]>>6 == 2 {
				prev--
			}
			s.Query = s.Query[:prev] + s.Query[s.QueryCursor:]
			s.QueryCursor = prev
		}
	} else {
		if s.ReplaceCursor > 0 {
			prev := s.ReplaceCursor - 1
			for prev > 0 && s.ReplaceText[prev]>>6 == 2 {
				prev--
			}
			s.ReplaceText = s.ReplaceText[:prev] + s.ReplaceText[s.ReplaceCursor:]
			s.ReplaceCursor = prev
		}
	}
}

// DeleteForward removes the character at the cursor in the active field.
func (s *SearchState) DeleteForward() {
	if s.ActiveField == FieldFind {
		if s.QueryCursor < len(s.Query) {
			next := s.QueryCursor + 1
			for next < len(s.Query) && s.Query[next]>>6 == 2 {
				next++
			}
			s.Query = s.Query[:s.QueryCursor] + s.Query[next:]
		}
	} else {
		if s.ReplaceCursor < len(s.ReplaceText) {
			next := s.ReplaceCursor + 1
			for next < len(s.ReplaceText) && s.ReplaceText[next]>>6 == 2 {
				next++
			}
			s.ReplaceText = s.ReplaceText[:s.ReplaceCursor] + s.ReplaceText[next:]
		}
	}
}

// MoveLeft moves the cursor left one character in the active field.
func (s *SearchState) MoveLeft() {
	if s.ActiveField == FieldFind {
		if s.QueryCursor > 0 {
			s.QueryCursor--
			for s.QueryCursor > 0 && s.Query[s.QueryCursor]>>6 == 2 {
				s.QueryCursor--
			}
		}
	} else {
		if s.ReplaceCursor > 0 {
			s.ReplaceCursor--
			for s.ReplaceCursor > 0 && s.ReplaceText[s.ReplaceCursor]>>6 == 2 {
				s.ReplaceCursor--
			}
		}
	}
}

// MoveRight moves the cursor right one character in the active field.
func (s *SearchState) MoveRight() {
	if s.ActiveField == FieldFind {
		if s.QueryCursor < len(s.Query) {
			s.QueryCursor++
			for s.QueryCursor < len(s.Query) && s.Query[s.QueryCursor]>>6 == 2 {
				s.QueryCursor++
			}
		}
	} else {
		if s.ReplaceCursor < len(s.ReplaceText) {
			s.ReplaceCursor++
			for s.ReplaceCursor < len(s.ReplaceText) && s.ReplaceText[s.ReplaceCursor]>>6 == 2 {
				s.ReplaceCursor++
			}
		}
	}
}

// MoveHome moves the cursor to the start of the active field.
func (s *SearchState) MoveHome() {
	if s.ActiveField == FieldFind {
		s.QueryCursor = 0
	} else {
		s.ReplaceCursor = 0
	}
}

// MoveEnd moves the cursor to the end of the active field.
func (s *SearchState) MoveEnd() {
	if s.ActiveField == FieldFind {
		s.QueryCursor = len(s.Query)
	} else {
		s.ReplaceCursor = len(s.ReplaceText)
	}
}

// ToggleField switches focus between Find and Replace fields.
// Enables the replace field if it's not yet visible.
func (s *SearchState) ToggleField() {
	if !s.ShowReplace {
		s.ShowReplace = true
		s.ActiveField = FieldReplace
		return
	}
	if s.ActiveField == FieldFind {
		s.ActiveField = FieldReplace
	} else {
		s.ActiveField = FieldFind
	}
}
