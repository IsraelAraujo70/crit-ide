package highlight

import (
	"github.com/smacker/go-tree-sitter/toml"
)

// TSLangTOML returns the tree-sitter language definition for TOML.
func TSLangTOML() *TSLangDef {
	return &TSLangDef{
		ID:         "toml",
		Extensions: []string{".toml"},
		FileNames:  []string{"Cargo.toml", "pyproject.toml"},
		Language:   toml.GetLanguage(),
		NodeMap:    tomlNodeMap(),
	}
}

func tomlNodeMap() map[string]TokenType {
	return map[string]TokenType{
		// Keys.
		"bare_key":   TokenProperty,
		"dotted_key": TokenProperty,
		"quoted_key": TokenProperty,

		// Strings.
		"string":          TokenString,
		"basic_string":    TokenString,
		"literal_string":  TokenString,
		"multiline_basic_string":   TokenString,
		"multiline_literal_string": TokenString,

		// Comments.
		"comment": TokenComment,

		// Numbers.
		"integer": TokenNumber,
		"float":   TokenNumber,

		// Constants.
		"true":  TokenConstant,
		"false": TokenConstant,

		// Section headers.
		"table":       TokenTag,
		"table_array": TokenTag,

		// Operators.
		"=": TokenOperator,
	}
}
