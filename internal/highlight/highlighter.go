package highlight

// Highlighter produces highlight tokens for a document.
// V1 is regex-based; V2 can swap in tree-sitter behind this interface.
type Highlighter interface {
	// HighlightLine returns tokens for a single line.
	// lineIndex is the zero-based line number in the document.
	HighlightLine(lineIndex int, line string) []Token

	// SetLanguage switches the highlighter to the given language ID.
	// Returns false if the language is not supported.
	SetLanguage(langID string) bool

	// InvalidateFrom marks that lines from lineIndex onward may need
	// re-highlighting (e.g., after an edit that could change block comment state).
	InvalidateFrom(lineIndex int)
}
