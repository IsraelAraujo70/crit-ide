package highlight

import (
	"github.com/smacker/go-tree-sitter/python"
)

// TSLangPython returns the tree-sitter language definition for Python.
func TSLangPython() *TSLangDef {
	return &TSLangDef{
		ID:         "python",
		Extensions: []string{".py", ".pyw"},
		FileNames:  []string{"SConstruct", "SConscript"},
		Language:   python.GetLanguage(),
		NodeMap:    pythonNodeMap(),
	}
}

func pythonNodeMap() map[string]TokenType {
	return map[string]TokenType{
		// Keywords.
		"def":      TokenKeyword,
		"class":    TokenKeyword,
		"if":       TokenKeyword,
		"elif":     TokenKeyword,
		"else":     TokenKeyword,
		"for":      TokenKeyword,
		"while":    TokenKeyword,
		"return":   TokenKeyword,
		"import":   TokenKeyword,
		"from":     TokenKeyword,
		"as":       TokenKeyword,
		"with":     TokenKeyword,
		"try":      TokenKeyword,
		"except":   TokenKeyword,
		"finally":  TokenKeyword,
		"raise":    TokenKeyword,
		"pass":     TokenKeyword,
		"break":    TokenKeyword,
		"continue": TokenKeyword,
		"yield":    TokenKeyword,
		"lambda":   TokenKeyword,
		"global":   TokenKeyword,
		"nonlocal": TokenKeyword,
		"del":      TokenKeyword,
		"assert":   TokenKeyword,
		"async":    TokenKeyword,
		"await":    TokenKeyword,
		"in":       TokenKeyword,
		"not":      TokenKeyword,
		"and":      TokenKeyword,
		"or":       TokenKeyword,
		"is":       TokenKeyword,

		// Strings.
		"string": TokenString,

		// Comments.
		"comment": TokenComment,

		// Numbers.
		"integer": TokenNumber,
		"float":   TokenNumber,

		// Types.
		"type": TokenTypeName,

		// Constants.
		"true":  TokenConstant,
		"false": TokenConstant,
		"none":  TokenConstant,
		"True":  TokenConstant,
		"False": TokenConstant,
		"None":  TokenConstant,

		// Operators.
		"+":  TokenOperator,
		"-":  TokenOperator,
		"*":  TokenOperator,
		"/":  TokenOperator,
		"//": TokenOperator,
		"%":  TokenOperator,
		"**": TokenOperator,
		"==": TokenOperator,
		"!=": TokenOperator,
		"<":  TokenOperator,
		">":  TokenOperator,
		"<=": TokenOperator,
		">=": TokenOperator,
		"=":  TokenOperator,
		"->": TokenOperator,
	}
}
