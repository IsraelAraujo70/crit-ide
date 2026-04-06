package highlight

import (
	"strings"
)

// lineState tracks cross-line state (block comments) entering a given line.
type lineState struct {
	inBlockComment bool
}

// RegexHighlighter implements Highlighter using compiled regexp patterns.
type RegexHighlighter struct {
	lang       *LanguageDef
	stateCache []lineState
}

// NewRegexHighlighter creates a highlighter with no language set.
func NewRegexHighlighter() *RegexHighlighter {
	return &RegexHighlighter{}
}

// SetLanguage switches to the given language ID using the provided registry.
// Returns false if the language is not found.
func (h *RegexHighlighter) SetLanguage(langID string) bool {
	// This method cannot look up the registry by itself; use SetLanguageDef instead.
	return h.lang != nil && h.lang.ID == langID
}

// SetLanguageDef directly sets the language definition.
func (h *RegexHighlighter) SetLanguageDef(def *LanguageDef) {
	h.lang = def
	h.stateCache = nil
}

// InvalidateFrom clears cached state from the given line onward.
func (h *RegexHighlighter) InvalidateFrom(lineIndex int) {
	if lineIndex < len(h.stateCache) {
		h.stateCache = h.stateCache[:lineIndex]
	}
}

// HighlightLine returns tokens for a single line.
func (h *RegexHighlighter) HighlightLine(lineIndex int, line string) []Token {
	if h.lang == nil {
		return nil
	}

	// Determine entering state for this line.
	state := h.getState(lineIndex)

	var tokens []Token

	// Handle block comments that span lines.
	if state.inBlockComment && h.lang.BlockCommentClose != "" {
		closeIdx := strings.Index(line, h.lang.BlockCommentClose)
		if closeIdx == -1 {
			// Entire line is still in a block comment.
			tokens = append(tokens, Token{Start: 0, End: len(line), Type: TokenComment})
			h.setState(lineIndex+1, lineState{inBlockComment: true})
			return tokens
		}
		// Block comment ends on this line.
		end := closeIdx + len(h.lang.BlockCommentClose)
		tokens = append(tokens, Token{Start: 0, End: end, Type: TokenComment})
		// Continue highlighting the rest of the line.
		line = line[end:]
		if len(line) > 0 {
			rest := h.highlightSegment(end, line)
			tokens = append(tokens, rest...)
		}
		h.setState(lineIndex+1, lineState{inBlockComment: false})
		return tokens
	}

	tokens = h.highlightSegment(0, line)

	// Check if a block comment was opened but not closed on this line.
	nextState := lineState{inBlockComment: false}
	if h.lang.BlockCommentOpen != "" {
		nextState.inBlockComment = h.endsInBlockComment(line)
	}
	h.setState(lineIndex+1, nextState)

	return tokens
}

// highlightSegment applies regex patterns to a segment of text.
// baseOffset is the byte offset of the segment within the full line.
func (h *RegexHighlighter) highlightSegment(baseOffset int, text string) []Token {
	// occupied tracks which byte positions are already claimed.
	occupied := make([]bool, len(text))
	var tokens []Token

	for _, rule := range h.lang.Patterns {
		matches := rule.Pattern.FindAllStringIndex(text, -1)
		for _, m := range matches {
			start, end := m[0], m[1]
			// Check if any position in this range is already occupied.
			overlap := false
			for i := start; i < end; i++ {
				if occupied[i] {
					overlap = true
					break
				}
			}
			if overlap {
				continue
			}
			// Claim the range.
			for i := start; i < end; i++ {
				occupied[i] = true
			}
			tokens = append(tokens, Token{
				Start: baseOffset + start,
				End:   baseOffset + end,
				Type:  rule.Type,
			})
		}
	}

	// Sort tokens by start position for efficient rendering.
	sortTokens(tokens)
	return tokens
}

// endsInBlockComment checks if the line ends inside an unclosed block comment.
func (h *RegexHighlighter) endsInBlockComment(line string) bool {
	open := h.lang.BlockCommentOpen
	close := h.lang.BlockCommentClose
	if open == "" || close == "" {
		return false
	}

	depth := 0
	i := 0
	for i < len(line) {
		// Skip strings to avoid false block comment detection.
		if line[i] == '"' || line[i] == '\'' || line[i] == '`' {
			quote := line[i]
			i++
			for i < len(line) {
				if line[i] == '\\' {
					i += 2
					continue
				}
				if line[i] == quote {
					i++
					break
				}
				i++
			}
			continue
		}

		// Check for line comment (skip rest of line).
		if h.lang.LineComment != "" && strings.HasPrefix(line[i:], h.lang.LineComment) && depth == 0 {
			return false
		}

		if strings.HasPrefix(line[i:], open) {
			depth++
			i += len(open)
			continue
		}
		if depth > 0 && strings.HasPrefix(line[i:], close) {
			depth--
			i += len(close)
			continue
		}
		i++
	}
	return depth > 0
}

// getState returns the entering state for the given line.
func (h *RegexHighlighter) getState(lineIndex int) lineState {
	if lineIndex == 0 {
		return lineState{}
	}
	if lineIndex <= len(h.stateCache) {
		return h.stateCache[lineIndex-1]
	}
	// State not cached; assume not in block comment.
	return lineState{}
}

// setState caches the entering state for the given line.
func (h *RegexHighlighter) setState(lineIndex int, state lineState) {
	for len(h.stateCache) < lineIndex {
		h.stateCache = append(h.stateCache, lineState{})
	}
	if lineIndex > 0 && lineIndex <= len(h.stateCache) {
		h.stateCache[lineIndex-1] = state
	}
}

// sortTokens sorts tokens by start offset using insertion sort (small slices).
func sortTokens(tokens []Token) {
	for i := 1; i < len(tokens); i++ {
		key := tokens[i]
		j := i - 1
		for j >= 0 && tokens[j].Start > key.Start {
			tokens[j+1] = tokens[j]
			j--
		}
		tokens[j+1] = key
	}
}
