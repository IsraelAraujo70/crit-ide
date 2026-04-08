package editor

import (
	"strings"
	"unicode/utf8"
)

// PaletteEntry represents a single command in the command palette.
type PaletteEntry struct {
	ID         string // Action ID (e.g., "file.save").
	Label      string // Human-readable label (e.g., "Save File").
	Keybinding string // Keybinding hint (e.g., "Ctrl+S").
	Category   string // Category for grouping (e.g., "File", "Edit").
}

// PaletteState holds the state for the command palette popup.
type PaletteState struct {
	Query       string // Search query typed by the user.
	CursorPos   int    // Byte offset cursor within Query.
	SelectedIdx int    // Currently selected result index (0-based, across filtered list).
	ScrollY     int    // Scroll offset in the results list.

	// All registered commands and filtered results.
	AllEntries []PaletteEntry
	Filtered   []PaletteEntry
}

// NewPaletteState creates a new command palette state with the given entries.
func NewPaletteState(entries []PaletteEntry) *PaletteState {
	ps := &PaletteState{
		AllEntries: entries,
	}
	ps.Filtered = entries // Initially show all.
	return ps
}

// InsertChar inserts a character at the cursor position and re-filters.
func (p *PaletteState) InsertChar(ch rune) {
	c := string(ch)
	p.Query = p.Query[:p.CursorPos] + c + p.Query[p.CursorPos:]
	p.CursorPos += len(c)
	p.refilter()
}

// DeleteBackward removes the character before the cursor.
func (p *PaletteState) DeleteBackward() {
	if p.CursorPos > 0 {
		prev := p.CursorPos - 1
		for prev > 0 && p.Query[prev]>>6 == 2 {
			prev--
		}
		p.Query = p.Query[:prev] + p.Query[p.CursorPos:]
		p.CursorPos = prev
		p.refilter()
	}
}

// DeleteForward removes the character at the cursor.
func (p *PaletteState) DeleteForward() {
	if p.CursorPos < len(p.Query) {
		next := p.CursorPos + 1
		for next < len(p.Query) && p.Query[next]>>6 == 2 {
			next++
		}
		p.Query = p.Query[:p.CursorPos] + p.Query[next:]
		p.refilter()
	}
}

// MoveLeft moves the cursor left one character.
func (p *PaletteState) MoveLeft() {
	if p.CursorPos > 0 {
		_, size := utf8.DecodeLastRuneInString(p.Query[:p.CursorPos])
		p.CursorPos -= size
	}
}

// MoveRight moves the cursor right one character.
func (p *PaletteState) MoveRight() {
	if p.CursorPos < len(p.Query) {
		_, size := utf8.DecodeRuneInString(p.Query[p.CursorPos:])
		p.CursorPos += size
	}
}

// MoveHome moves the cursor to the start.
func (p *PaletteState) MoveHome() {
	p.CursorPos = 0
}

// MoveEnd moves the cursor to the end.
func (p *PaletteState) MoveEnd() {
	p.CursorPos = len(p.Query)
}

// MoveUp moves the selection up by one.
func (p *PaletteState) MoveUp() {
	if p.SelectedIdx > 0 {
		p.SelectedIdx--
		if p.SelectedIdx < p.ScrollY {
			p.ScrollY = p.SelectedIdx
		}
	}
}

// MoveDown moves the selection down by one.
func (p *PaletteState) MoveDown(maxVisible int) {
	if p.SelectedIdx < len(p.Filtered)-1 {
		p.SelectedIdx++
		if p.SelectedIdx >= p.ScrollY+maxVisible {
			p.ScrollY = p.SelectedIdx - maxVisible + 1
		}
	}
}

// SelectedEntry returns the currently selected entry, or nil if none.
func (p *PaletteState) SelectedEntry() *PaletteEntry {
	if p.SelectedIdx >= 0 && p.SelectedIdx < len(p.Filtered) {
		return &p.Filtered[p.SelectedIdx]
	}
	return nil
}

// ResultCount returns the number of filtered results.
func (p *PaletteState) ResultCount() int {
	return len(p.Filtered)
}

// refilter applies fuzzy filtering on the query and resets selection.
func (p *PaletteState) refilter() {
	p.SelectedIdx = 0
	p.ScrollY = 0

	if p.Query == "" {
		p.Filtered = p.AllEntries
		return
	}

	query := strings.ToLower(p.Query)
	type scored struct {
		entry PaletteEntry
		score int
	}

	var results []scored
	for _, e := range p.AllEntries {
		label := strings.ToLower(e.Label)
		id := strings.ToLower(e.ID)
		cat := strings.ToLower(e.Category)

		// Try matching against label, then ID, then category+label.
		s := fuzzyScore(query, label)
		if s2 := fuzzyScore(query, id); s2 > s {
			s = s2
		}
		if s2 := fuzzyScore(query, cat+" "+label); s2 > s {
			s = s2
		}
		if s > 0 {
			results = append(results, scored{entry: e, score: s})
		}
	}

	// Sort by score descending.
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].score > results[j-1].score; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}

	p.Filtered = make([]PaletteEntry, len(results))
	for i, r := range results {
		p.Filtered[i] = r.entry
	}
}

// fuzzyScore returns a score > 0 if pattern fuzzy-matches candidate, 0 otherwise.
func fuzzyScore(pattern, candidate string) int {
	if pattern == "" {
		return 1
	}
	pi := 0
	patRunes := []rune(pattern)
	score := 0
	consecutive := 0
	lastMatch := -1

	for ci, ch := range candidate {
		if pi < len(patRunes) && ch == patRunes[pi] {
			score += 10
			// Bonus for consecutive.
			if lastMatch == ci-1 {
				consecutive++
				score += consecutive * 5
			} else {
				consecutive = 0
			}
			// Bonus for start of word.
			if ci == 0 || candidate[ci-1] == ' ' || candidate[ci-1] == '.' || candidate[ci-1] == '_' {
				score += 15
			}
			lastMatch = ci
			pi++
		}
	}

	if pi < len(patRunes) {
		return 0 // Not all pattern chars matched.
	}

	// Bonus for exact prefix.
	if strings.HasPrefix(candidate, pattern) {
		score += 30
	}

	return score
}
