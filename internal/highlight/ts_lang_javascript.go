package highlight

import (
	"github.com/smacker/go-tree-sitter/javascript"
	ts "github.com/smacker/go-tree-sitter/typescript/typescript"
)

// TSLangJavaScript returns the tree-sitter language definition for JavaScript.
func TSLangJavaScript() *TSLangDef {
	return &TSLangDef{
		ID:         "javascript",
		Extensions: []string{".js", ".jsx", ".mjs", ".cjs"},
		Language:   javascript.GetLanguage(),
		NodeMap:    jsNodeMap(),
	}
}

// TSLangTypeScript returns the tree-sitter language definition for TypeScript.
func TSLangTypeScript() *TSLangDef {
	m := jsNodeMap()
	// Add TypeScript-specific node types.
	m["type_identifier"] = TokenTypeName
	m["predefined_type"] = TokenTypeName
	m["type_annotation"] = TokenTypeName
	m["interface"] = TokenKeyword
	m["enum"] = TokenKeyword
	m["implements"] = TokenKeyword
	m["declare"] = TokenKeyword
	m["namespace"] = TokenKeyword
	m["abstract"] = TokenKeyword
	m["override"] = TokenKeyword
	m["readonly"] = TokenKeyword
	m["keyof"] = TokenKeyword
	m["satisfies"] = TokenKeyword
	m["as"] = TokenKeyword

	return &TSLangDef{
		ID:         "typescript",
		Extensions: []string{".ts", ".tsx", ".mts", ".cts"},
		Language:   ts.GetLanguage(),
		NodeMap:    m,
	}
}

func jsNodeMap() map[string]TokenType {
	return map[string]TokenType{
		// Keywords.
		"const":      TokenKeyword,
		"let":        TokenKeyword,
		"var":        TokenKeyword,
		"function":   TokenKeyword,
		"return":     TokenKeyword,
		"if":         TokenKeyword,
		"else":       TokenKeyword,
		"for":        TokenKeyword,
		"while":      TokenKeyword,
		"do":         TokenKeyword,
		"switch":     TokenKeyword,
		"case":       TokenKeyword,
		"default":    TokenKeyword,
		"break":      TokenKeyword,
		"continue":   TokenKeyword,
		"class":      TokenKeyword,
		"extends":    TokenKeyword,
		"new":        TokenKeyword,
		"this":       TokenKeyword,
		"super":      TokenKeyword,
		"import":     TokenKeyword,
		"export":     TokenKeyword,
		"from":       TokenKeyword,
		"async":      TokenKeyword,
		"await":      TokenKeyword,
		"yield":      TokenKeyword,
		"throw":      TokenKeyword,
		"try":        TokenKeyword,
		"catch":      TokenKeyword,
		"finally":    TokenKeyword,
		"typeof":     TokenKeyword,
		"instanceof": TokenKeyword,
		"in":         TokenKeyword,
		"of":         TokenKeyword,
		"void":       TokenKeyword,
		"delete":     TokenKeyword,
		"with":       TokenKeyword,
		"debugger":   TokenKeyword,
		"static":     TokenKeyword,

		// Strings.
		"string":          TokenString,
		"template_string": TokenString,

		// Comments.
		"comment": TokenComment,

		// Numbers.
		"number": TokenNumber,

		// Constants.
		"true":      TokenConstant,
		"false":     TokenConstant,
		"null":      TokenConstant,
		"undefined": TokenConstant,

		// Properties.
		"property_identifier": TokenProperty,

		// Operators.
		"=>": TokenOperator,
		"===": TokenOperator,
		"!==": TokenOperator,
		"==":  TokenOperator,
		"!=":  TokenOperator,
		"&&":  TokenOperator,
		"||":  TokenOperator,
		"??":  TokenOperator,
		"?.":  TokenOperator,
		"...": TokenOperator,
		"+":   TokenOperator,
		"-":   TokenOperator,
		"*":   TokenOperator,
		"/":   TokenOperator,
		"%":   TokenOperator,
		"<":   TokenOperator,
		">":   TokenOperator,
		"<=":  TokenOperator,
		">=":  TokenOperator,
		"=":   TokenOperator,
	}
}
