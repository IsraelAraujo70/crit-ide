package highlight

import (
	"github.com/smacker/go-tree-sitter/css"
)

// TSLangCSS returns the tree-sitter language definition for CSS.
func TSLangCSS() *TSLangDef {
	return &TSLangDef{
		ID:         "css",
		Extensions: []string{".css"},
		Language:   css.GetLanguage(),
		NodeMap:    cssNodeMap(),
	}
}

func cssNodeMap() map[string]TokenType {
	return map[string]TokenType{
		// Selectors.
		"class_name":    TokenTypeName,
		"id_name":       TokenTypeName,
		"tag_name":      TokenTag,
		"class_selector": TokenTypeName,
		"id_selector":    TokenTypeName,

		// Properties.
		"property_name": TokenProperty,
		"feature_name":  TokenProperty,

		// Values.
		"plain_value":   TokenConstant,
		"color_value":   TokenConstant,
		"integer_value": TokenNumber,
		"float_value":   TokenNumber,
		"unit":          TokenNumber,

		// Strings.
		"string_value": TokenString,

		// Comments.
		"comment": TokenComment,

		// At-rules.
		"@media":    TokenKeyword,
		"@import":   TokenKeyword,
		"@keyframes": TokenKeyword,
		"@font-face": TokenKeyword,
		"at_keyword": TokenKeyword,

		// Operators.
		":": TokenOperator,
	}
}
