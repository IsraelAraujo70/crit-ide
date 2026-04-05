package highlight

import "regexp"

// LangJavaScript returns the language definition for JavaScript.
func LangJavaScript() *LanguageDef {
	return &LanguageDef{
		ID:                "javascript",
		Extensions:        []string{".js", ".jsx", ".mjs", ".cjs"},
		LineComment:       "//",
		BlockCommentOpen:  "/*",
		BlockCommentClose: "*/",
		Patterns: []PatternRule{
			{TokenComment, regexp.MustCompile(`//.*`)},
			{TokenString, regexp.MustCompile("`(?:[^`\\\\]|\\\\.)*`")},
			{TokenString, regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)},
			{TokenString, regexp.MustCompile(`'(?:[^'\\]|\\.)*'`)},
			{TokenKeyword, regexp.MustCompile(`\b(?:async|await|break|case|catch|class|const|continue|debugger|default|delete|do|else|export|extends|finally|for|from|function|if|import|in|instanceof|let|new|of|return|static|super|switch|this|throw|try|typeof|var|void|while|with|yield)\b`)},
			{TokenConstant, regexp.MustCompile(`\b(?:true|false|null|undefined|NaN|Infinity)\b`)},
			{TokenBuiltin, regexp.MustCompile(`\b(?:console|Math|JSON|Object|Array|String|Number|Boolean|Date|RegExp|Error|Map|Set|WeakMap|WeakSet|Promise|Symbol|Proxy|Reflect|parseInt|parseFloat|isNaN|isFinite|setTimeout|setInterval|clearTimeout|clearInterval|fetch|require)\b`)},
			{TokenNumber, regexp.MustCompile(`\b(?:0[xX][0-9a-fA-F_]+|0[oO][0-7_]+|0[bB][01_]+|[0-9][0-9_]*(?:\.[0-9_]+)?(?:[eE][+-]?[0-9_]+)?n?)\b`)},
			{TokenFunction, regexp.MustCompile(`\b([a-zA-Z_$]\w*)\s*\(`)},
			{TokenOperator, regexp.MustCompile(`(?:===|!==|=>|&&|\|\||\?\?|\?\.|\.\.\.|[+\-*/%&|^<>=!]=?)`)},
		},
	}
}

// LangTypeScript returns the language definition for TypeScript.
func LangTypeScript() *LanguageDef {
	ts := LangJavaScript()
	ts.ID = "typescript"
	ts.Extensions = []string{".ts", ".tsx", ".mts", ".cts"}
	// Add TS-specific keywords before the JS keywords.
	tsKeywords := PatternRule{
		TokenKeyword,
		regexp.MustCompile(`\b(?:abstract|as|declare|enum|implements|interface|keyof|module|namespace|never|override|readonly|satisfies|type|unknown)\b`),
	}
	tsTypes := PatternRule{
		TokenTypeName,
		regexp.MustCompile(`\b(?:string|number|boolean|void|any|never|unknown|object|symbol|bigint|undefined)\b`),
	}
	ts.Patterns = append([]PatternRule{tsKeywords, tsTypes}, ts.Patterns...)
	return ts
}
