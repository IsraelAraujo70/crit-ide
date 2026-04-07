package highlight

import (
	"github.com/smacker/go-tree-sitter/golang"
)

// TSLangGo returns the tree-sitter language definition for Go.
func TSLangGo() *TSLangDef {
	return &TSLangDef{
		ID:        "go",
		Extensions: []string{".go"},
		Language:   golang.GetLanguage(),
		NodeMap:    goNodeMap(),
	}
}

func goNodeMap() map[string]TokenType {
	return map[string]TokenType{
		// Keywords.
		"break":       TokenKeyword,
		"case":        TokenKeyword,
		"chan":         TokenKeyword,
		"const":       TokenKeyword,
		"continue":    TokenKeyword,
		"default":     TokenKeyword,
		"defer":       TokenKeyword,
		"else":        TokenKeyword,
		"fallthrough": TokenKeyword,
		"for":         TokenKeyword,
		"func":        TokenKeyword,
		"go":          TokenKeyword,
		"goto":        TokenKeyword,
		"if":          TokenKeyword,
		"import":      TokenKeyword,
		"interface":   TokenKeyword,
		"map":         TokenKeyword,
		"package":     TokenKeyword,
		"range":       TokenKeyword,
		"return":      TokenKeyword,
		"select":      TokenKeyword,
		"struct":      TokenKeyword,
		"switch":      TokenKeyword,
		"type":        TokenKeyword,
		"var":         TokenKeyword,

		// Types.
		"type_identifier": TokenTypeName,

		// Strings.
		"interpreted_string_literal": TokenString,
		"raw_string_literal":         TokenString,
		"rune_literal":               TokenString,

		// Comments.
		"comment": TokenComment,

		// Numbers.
		"int_literal":       TokenNumber,
		"float_literal":     TokenNumber,
		"imaginary_literal": TokenNumber,

		// Operators.
		":=": TokenOperator,
		"<-": TokenOperator,
		"&&": TokenOperator,
		"||": TokenOperator,
		"==": TokenOperator,
		"!=": TokenOperator,
		"<=": TokenOperator,
		">=": TokenOperator,
		"<":  TokenOperator,
		">":  TokenOperator,
		"+":  TokenOperator,
		"-":  TokenOperator,
		"*":  TokenOperator,
		"/":  TokenOperator,
		"%":  TokenOperator,
		"&":  TokenOperator,
		"|":  TokenOperator,
		"^":  TokenOperator,
		"!":  TokenOperator,

		// Constants.
		"true":  TokenConstant,
		"false": TokenConstant,
		"nil":   TokenConstant,
		"iota":  TokenConstant,
	}
}

// goBuiltins returns the set of Go builtin function names.
var goBuiltins = map[string]bool{
	"append": true, "cap": true, "clear": true, "close": true,
	"complex": true, "copy": true, "delete": true, "imag": true,
	"len": true, "make": true, "max": true, "min": true,
	"new": true, "panic": true, "print": true, "println": true,
	"real": true, "recover": true,
}

