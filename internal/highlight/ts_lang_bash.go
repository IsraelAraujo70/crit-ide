package highlight

import (
	"github.com/smacker/go-tree-sitter/bash"
)

// TSLangBash returns the tree-sitter language definition for Bash/Shell.
func TSLangBash() *TSLangDef {
	return &TSLangDef{
		ID:         "shell",
		Extensions: []string{".sh", ".bash", ".zsh"},
		FileNames:  []string{"Makefile", ".bashrc", ".zshrc", ".profile"},
		Language:   bash.GetLanguage(),
		NodeMap:    bashNodeMap(),
	}
}

func bashNodeMap() map[string]TokenType {
	return map[string]TokenType{
		// Keywords.
		"if":       TokenKeyword,
		"then":     TokenKeyword,
		"else":     TokenKeyword,
		"elif":     TokenKeyword,
		"fi":       TokenKeyword,
		"for":      TokenKeyword,
		"while":    TokenKeyword,
		"do":       TokenKeyword,
		"done":     TokenKeyword,
		"in":       TokenKeyword,
		"case":     TokenKeyword,
		"esac":     TokenKeyword,
		"function": TokenKeyword,

		// Strings.
		"string":         TokenString,
		"raw_string":     TokenString,
		"heredoc_body":   TokenString,
		"heredoc_start":  TokenString,

		// Comments.
		"comment": TokenComment,

		// Variables.
		"variable_name":    TokenVariable,
		"simple_expansion": TokenVariable,
		"special_variable_name": TokenVariable,
		"$":                TokenVariable,

		// Numbers.
		"number": TokenNumber,

		// Operators.
		"test_operator": TokenOperator,
		"|":             TokenOperator,
		"||":            TokenOperator,
		"&&":            TokenOperator,
		">":             TokenOperator,
		"<":             TokenOperator,
		">>":            TokenOperator,
		"=":             TokenOperator,
	}
}
