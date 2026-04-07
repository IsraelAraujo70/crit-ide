package highlight

import (
	"github.com/smacker/go-tree-sitter/rust"
)

// TSLangRust returns the tree-sitter language definition for Rust.
func TSLangRust() *TSLangDef {
	return &TSLangDef{
		ID:         "rust",
		Extensions: []string{".rs"},
		Language:   rust.GetLanguage(),
		NodeMap:    rustNodeMap(),
	}
}

func rustNodeMap() map[string]TokenType {
	return map[string]TokenType{
		// Keywords.
		"fn":              TokenKeyword,
		"let":             TokenKeyword,
		"mut":             TokenKeyword,
		"if":              TokenKeyword,
		"else":            TokenKeyword,
		"for":             TokenKeyword,
		"while":           TokenKeyword,
		"loop":            TokenKeyword,
		"match":           TokenKeyword,
		"return":          TokenKeyword,
		"break":           TokenKeyword,
		"continue":        TokenKeyword,
		"use":             TokenKeyword,
		"mod":             TokenKeyword,
		"pub":             TokenKeyword,
		"struct":          TokenKeyword,
		"enum":            TokenKeyword,
		"impl":            TokenKeyword,
		"trait":           TokenKeyword,
		"type":            TokenKeyword,
		"const":           TokenKeyword,
		"static":          TokenKeyword,
		"unsafe":          TokenKeyword,
		"async":           TokenKeyword,
		"await":           TokenKeyword,
		"move":            TokenKeyword,
		"ref":             TokenKeyword,
		"self":            TokenKeyword,
		"Self":            TokenKeyword,
		"super":           TokenKeyword,
		"crate":           TokenKeyword,
		"extern":          TokenKeyword,
		"where":           TokenKeyword,
		"as":              TokenKeyword,
		"in":              TokenKeyword,
		"dyn":             TokenKeyword,
		"mutable_specifier": TokenKeyword,

		// Types.
		"type_identifier": TokenTypeName,
		"primitive_type":  TokenTypeName,

		// Strings.
		"string_literal": TokenString,
		"char_literal":   TokenString,
		"raw_string_literal": TokenString,

		// Comments.
		"line_comment":  TokenComment,
		"block_comment": TokenComment,

		// Numbers.
		"integer_literal": TokenNumber,
		"float_literal":   TokenNumber,

		// Constants.
		"true":  TokenConstant,
		"false": TokenConstant,

		// Variables (lifetimes).
		"lifetime": TokenVariable,

		// Operators.
		"::": TokenOperator,
		"=>": TokenOperator,
		"->": TokenOperator,
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
		"=":  TokenOperator,
	}
}
