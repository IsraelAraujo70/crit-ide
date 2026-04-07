package highlight

import (
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
)

// TSLangC returns the tree-sitter language definition for C.
func TSLangC() *TSLangDef {
	return &TSLangDef{
		ID:         "c",
		Extensions: []string{".c", ".h"},
		Language:   c.GetLanguage(),
		NodeMap:    cNodeMap(),
	}
}

// TSLangCPP returns the tree-sitter language definition for C++.
func TSLangCPP() *TSLangDef {
	m := cNodeMap()
	// Add C++ specific keywords.
	m["class"] = TokenKeyword
	m["namespace"] = TokenKeyword
	m["template"] = TokenKeyword
	m["virtual"] = TokenKeyword
	m["override"] = TokenKeyword
	m["public"] = TokenKeyword
	m["private"] = TokenKeyword
	m["protected"] = TokenKeyword
	m["new"] = TokenKeyword
	m["delete"] = TokenKeyword
	m["try"] = TokenKeyword
	m["catch"] = TokenKeyword
	m["throw"] = TokenKeyword
	m["using"] = TokenKeyword
	m["auto"] = TokenKeyword
	m["constexpr"] = TokenKeyword
	m["nullptr"] = TokenConstant

	return &TSLangDef{
		ID:         "cpp",
		Extensions: []string{".cpp", ".cc", ".cxx", ".hpp", ".hh", ".hxx"},
		Language:   cpp.GetLanguage(),
		NodeMap:    m,
	}
}

func cNodeMap() map[string]TokenType {
	return map[string]TokenType{
		// Keywords.
		"if":       TokenKeyword,
		"else":     TokenKeyword,
		"for":      TokenKeyword,
		"while":    TokenKeyword,
		"do":       TokenKeyword,
		"switch":   TokenKeyword,
		"case":     TokenKeyword,
		"default":  TokenKeyword,
		"break":    TokenKeyword,
		"continue": TokenKeyword,
		"return":   TokenKeyword,
		"struct":   TokenKeyword,
		"enum":     TokenKeyword,
		"union":    TokenKeyword,
		"typedef":  TokenKeyword,
		"static":   TokenKeyword,
		"extern":   TokenKeyword,
		"const":    TokenKeyword,
		"volatile": TokenKeyword,
		"sizeof":   TokenKeyword,
		"void":     TokenKeyword,
		"inline":   TokenKeyword,
		"register": TokenKeyword,
		"goto":     TokenKeyword,

		// Types.
		"type_identifier":  TokenTypeName,
		"primitive_type":   TokenTypeName,
		"sized_type_specifier": TokenTypeName,

		// Strings.
		"string_literal":    TokenString,
		"char_literal":      TokenString,
		"system_lib_string": TokenString,

		// Comments.
		"comment": TokenComment,

		// Numbers.
		"number_literal": TokenNumber,

		// Constants.
		"true":  TokenConstant,
		"false": TokenConstant,
		"null":  TokenConstant,
		"NULL":  TokenConstant,

		// Preprocessor.
		"preproc_include":    TokenPreproc,
		"preproc_def":        TokenPreproc,
		"preproc_ifdef":      TokenPreproc,
		"preproc_ifndef":     TokenPreproc,
		"preproc_if":         TokenPreproc,
		"preproc_else":       TokenPreproc,
		"preproc_elif":       TokenPreproc,
		"preproc_directive":  TokenPreproc,
		"#include":           TokenPreproc,
		"#define":            TokenPreproc,
		"#ifdef":             TokenPreproc,
		"#ifndef":            TokenPreproc,
		"#if":                TokenPreproc,
		"#else":              TokenPreproc,
		"#endif":             TokenPreproc,

		// Operators.
		"->": TokenOperator,
		"&&": TokenOperator,
		"||": TokenOperator,
		"==": TokenOperator,
		"!=": TokenOperator,
		"<=": TokenOperator,
		">=": TokenOperator,
		"<<": TokenOperator,
		">>": TokenOperator,
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
		"~":  TokenOperator,
		"!":  TokenOperator,
		"=":  TokenOperator,
	}
}
