package highlight

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// TreeSitterHighlighter implements Highlighter using tree-sitter parsing.
// It parses the full document and extracts tokens per line on demand.
type TreeSitterHighlighter struct {
	parser   *sitter.Parser
	tree     *sitter.Tree
	source   []byte
	langID   string
	langDef  *TSLangDef
	registry *TSLangRegistry
	dirty    bool

	// lineOffsets caches byte offset of each line start for fast lookup.
	lineOffsets []int
}

// NewTreeSitterHighlighter creates a highlighter with no language set.
func NewTreeSitterHighlighter(registry *TSLangRegistry) *TreeSitterHighlighter {
	return &TreeSitterHighlighter{
		parser:   sitter.NewParser(),
		registry: registry,
	}
}

// SetLanguage switches to the given language ID.
// Returns false if the language is not in the registry.
// Languages with no tree-sitter grammar (Language == nil) are recognized
// but produce no highlighting tokens.
func (h *TreeSitterHighlighter) SetLanguage(langID string) bool {
	if langID == "" {
		h.langDef = nil
		h.langID = ""
		h.tree = nil
		return false
	}
	def := h.registry.ByID(langID)
	if def == nil {
		h.langDef = nil
		h.langID = ""
		h.tree = nil
		return false
	}
	if h.langID == langID {
		return true
	}
	h.langID = langID
	h.langDef = def
	h.tree = nil
	h.dirty = true
	if def.Language != nil {
		h.parser.SetLanguage(def.Language)
	}
	return true
}

// SetSource provides the full document source for parsing.
// Must be called before HighlightLine, and after any edits.
// Forces a fresh parse since we don't have edit deltas for incremental parsing.
func (h *TreeSitterHighlighter) SetSource(source string) {
	h.source = []byte(source)
	h.buildLineOffsets()
	h.tree = nil // Force fresh parse (incremental requires tree.Edit with precise byte offsets).
	h.dirty = true
}

// InvalidateFrom marks that lines from lineIndex onward need re-highlighting.
func (h *TreeSitterHighlighter) InvalidateFrom(lineIndex int) {
	h.dirty = true
}

// HighlightLine returns tokens for a single line.
func (h *TreeSitterHighlighter) HighlightLine(lineIndex int, line string) []Token {
	if h.langDef == nil || h.langDef.Language == nil || h.source == nil {
		return nil
	}

	// Parse or reparse if needed.
	if h.dirty || h.tree == nil {
		h.parse()
	}

	if h.tree == nil {
		return nil
	}

	// Calculate byte range for this line.
	lineStart, lineEnd := h.lineByteRange(lineIndex)
	if lineStart < 0 {
		return nil
	}

	// Walk tree and collect tokens that intersect this line.
	root := h.tree.RootNode()
	var tokens []Token
	h.collectTokens(root, uint32(lineStart), uint32(lineEnd), lineStart, &tokens)

	sortTokens(tokens)
	return tokens
}

// parse runs tree-sitter on the full source.
func (h *TreeSitterHighlighter) parse() {
	if h.langDef == nil {
		return
	}
	tree, err := h.parser.ParseCtx(context.Background(), h.tree, h.source)
	if err != nil {
		return
	}
	h.tree = tree
	h.dirty = false
}

// buildLineOffsets computes byte offsets for the start of each line.
func (h *TreeSitterHighlighter) buildLineOffsets() {
	h.lineOffsets = h.lineOffsets[:0]
	h.lineOffsets = append(h.lineOffsets, 0)
	for i, b := range h.source {
		if b == '\n' {
			h.lineOffsets = append(h.lineOffsets, i+1)
		}
	}
}

// lineByteRange returns the start and end byte offsets for the given line.
// Returns (-1, -1) if out of range.
func (h *TreeSitterHighlighter) lineByteRange(lineIndex int) (int, int) {
	if lineIndex < 0 || lineIndex >= len(h.lineOffsets) {
		return -1, -1
	}
	start := h.lineOffsets[lineIndex]
	end := len(h.source)
	if lineIndex+1 < len(h.lineOffsets) {
		end = h.lineOffsets[lineIndex+1]
	}
	// Exclude trailing newline.
	if end > start && end <= len(h.source) && h.source[end-1] == '\n' {
		end--
	}
	return start, end
}

// collectTokens recursively walks the tree and emits tokens for leaf/mapped nodes
// that intersect the byte range [lineStart, lineEnd).
func (h *TreeSitterHighlighter) collectTokens(node *sitter.Node, lineStart, lineEnd uint32, baseOffset int, tokens *[]Token) {
	if node == nil {
		return
	}

	nodeStart := node.StartByte()
	nodeEnd := node.EndByte()

	// Skip nodes entirely outside the line range.
	if nodeEnd <= lineStart || nodeStart >= lineEnd {
		return
	}

	nodeType := node.Type()

	// Check if this node type maps to a token.
	tokenType, mapped := h.langDef.NodeMap[nodeType]

	// For function detection: check parent context.
	if !mapped && nodeType == "identifier" {
		tokenType, mapped = h.resolveIdentifier(node)
	}
	if !mapped && nodeType == "field_identifier" {
		tokenType, mapped = h.resolveFieldIdentifier(node)
	}

	// If mapped and it's a leaf-ish node (or a known compound like comment/string),
	// emit the token. For compound nodes, emit the whole span.
	if mapped {
		start := int(nodeStart) - baseOffset
		end := int(nodeEnd) - baseOffset
		if start < 0 {
			start = 0
		}
		lineLen := int(lineEnd) - baseOffset
		if end > lineLen {
			end = lineLen
		}
		if start < end {
			*tokens = append(*tokens, Token{Start: start, End: end, Type: tokenType})
		}
		// For compound nodes (comment, string), don't recurse into children.
		if isCompoundNode(nodeType) {
			return
		}
	}

	// Recurse into children.
	childCount := int(node.ChildCount())
	for i := 0; i < childCount; i++ {
		child := node.Child(i)
		h.collectTokens(child, lineStart, lineEnd, baseOffset, tokens)
	}
}

// resolveIdentifier checks if an identifier is in a function call or declaration context.
func (h *TreeSitterHighlighter) resolveIdentifier(node *sitter.Node) (TokenType, bool) {
	parent := node.Parent()
	if parent == nil {
		return TokenNone, false
	}
	parentType := parent.Type()

	// Function calls: call_expression where function child is this identifier.
	if parentType == "call_expression" {
		// Check if this is the function part (first named child).
		fn := parent.ChildByFieldName("function")
		if fn != nil && fn.StartByte() == node.StartByte() && fn.EndByte() == node.EndByte() {
			return TokenFunction, true
		}
	}

	// Function declarations.
	if parentType == "function_declaration" || parentType == "method_declaration" {
		name := parent.ChildByFieldName("name")
		if name != nil && name.StartByte() == node.StartByte() {
			return TokenFunction, true
		}
	}

	return TokenNone, false
}

// resolveFieldIdentifier checks if a field_identifier is a function call (selector.method()).
func (h *TreeSitterHighlighter) resolveFieldIdentifier(node *sitter.Node) (TokenType, bool) {
	parent := node.Parent()
	if parent == nil {
		return TokenNone, false
	}

	// selector_expression -> call_expression means this is a method call.
	if parent.Type() == "selector_expression" {
		grandparent := parent.Parent()
		if grandparent != nil && grandparent.Type() == "call_expression" {
			return TokenFunction, true
		}
	}

	return TokenNone, false
}

// isCompoundNode returns true for node types whose children should not be
// individually tokenized (the whole span is one token).
func isCompoundNode(nodeType string) bool {
	switch nodeType {
	case "comment",
		"interpreted_string_literal", "raw_string_literal", "string_literal",
		"string", "template_string", "string_content",
		"char_literal", "rune_literal":
		return true
	}
	return strings.HasSuffix(nodeType, "_comment") || strings.HasSuffix(nodeType, "_string")
}
