package highlight

import (
	"github.com/smacker/go-tree-sitter/yaml"
)

// TSLangYAML returns the tree-sitter language definition for YAML.
func TSLangYAML() *TSLangDef {
	return &TSLangDef{
		ID:         "yaml",
		Extensions: []string{".yml", ".yaml"},
		Language:   yaml.GetLanguage(),
		NodeMap:    yamlNodeMap(),
	}
}

func yamlNodeMap() map[string]TokenType {
	return map[string]TokenType{
		// Keys.
		"block_mapping_pair": TokenProperty,
		"flow_pair":          TokenProperty,

		// Strings.
		"double_quote_scalar": TokenString,
		"single_quote_scalar": TokenString,
		"block_scalar":        TokenString,
		"string_scalar":       TokenString,

		// Comments.
		"comment": TokenComment,

		// Numbers.
		"integer_scalar": TokenNumber,
		"float_scalar":   TokenNumber,

		// Constants.
		"boolean_scalar": TokenConstant,
		"null_scalar":    TokenConstant,

		// Tags/anchors.
		"tag":    TokenTag,
		"anchor": TokenVariable,
		"alias":  TokenVariable,

		// Operators.
		":": TokenOperator,
		"-": TokenOperator,
	}
}
