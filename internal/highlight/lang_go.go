package highlight

import "regexp"

// LangGo returns the language definition for Go.
func LangGo() *LanguageDef {
	return &LanguageDef{
		ID:                "go",
		Extensions:        []string{".go"},
		LineComment:       "//",
		BlockCommentOpen:  "/*",
		BlockCommentClose: "*/",
		Patterns: []PatternRule{
			{TokenComment, regexp.MustCompile(`//.*`)},
			{TokenString, regexp.MustCompile("\"(?:[^\"\\\\]|\\\\.)*\"")},
			{TokenString, regexp.MustCompile("`[^`]*`")},
			{TokenString, regexp.MustCompile("'(?:[^'\\\\]|\\\\.)*'")},
			{TokenKeyword, regexp.MustCompile(`\b(?:break|case|chan|const|continue|default|defer|else|fallthrough|for|func|go|goto|if|import|interface|map|package|range|return|select|struct|switch|type|var)\b`)},
			{TokenTypeName, regexp.MustCompile(`\b(?:bool|byte|complex64|complex128|error|float32|float64|int|int8|int16|int32|int64|rune|string|uint|uint8|uint16|uint32|uint64|uintptr|any)\b`)},
			{TokenBuiltin, regexp.MustCompile(`\b(?:append|cap|clear|close|complex|copy|delete|imag|len|make|max|min|new|panic|print|println|real|recover)\b`)},
			{TokenConstant, regexp.MustCompile(`\b(?:true|false|nil|iota)\b`)},
			{TokenNumber, regexp.MustCompile(`\b(?:0[xX][0-9a-fA-F_]+|0[oO]?[0-7_]+|0[bB][01_]+|[0-9][0-9_]*(?:\.[0-9_]+)?(?:[eE][+-]?[0-9_]+)?i?)\b`)},
			{TokenFunction, regexp.MustCompile(`\b([a-zA-Z_]\w*)\s*\(`)},
			{TokenOperator, regexp.MustCompile(`(?::=|<-|&&|\|\||[+\-*/%&|^<>=!]=?|\.{3})`)},
		},
	}
}
