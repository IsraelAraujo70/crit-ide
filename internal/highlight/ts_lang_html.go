package highlight

import (
	"github.com/smacker/go-tree-sitter/html"
)

// TSLangHTML returns the tree-sitter language definition for HTML.
func TSLangHTML() *TSLangDef {
	return &TSLangDef{
		ID:         "html",
		Extensions: []string{".html", ".htm"},
		Language:   html.GetLanguage(),
		NodeMap:    htmlNodeMap(),
	}
}

func htmlNodeMap() map[string]TokenType {
	return map[string]TokenType{
		// Tags.
		"tag_name": TokenTag,
		"doctype":  TokenTag,

		// Attributes.
		"attribute_name":  TokenAttribute,
		"attribute_value": TokenString,
		"quoted_attribute_value": TokenString,

		// Comments.
		"comment": TokenComment,

		// Special.
		"<!":  TokenTag,
		"</":  TokenTag,
		"<":   TokenTag,
		">":   TokenTag,
		"/>":  TokenTag,
	}
}
