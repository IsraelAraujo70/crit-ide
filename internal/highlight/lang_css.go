package highlight

import "regexp"

// LangCSS returns the language definition for CSS.
func LangCSS() *LanguageDef {
	return &LanguageDef{
		ID:                "css",
		Extensions:        []string{".css", ".scss", ".sass", ".less"},
		BlockCommentOpen:  "/*",
		BlockCommentClose: "*/",
		LineComment:       "//",
		Patterns: []PatternRule{
			{TokenComment, regexp.MustCompile(`//.*`)},
			{TokenString, regexp.MustCompile(`"(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'`)},
			{TokenProperty, regexp.MustCompile(`[\w-]+\s*:`)},
			{TokenVariable, regexp.MustCompile(`--[\w-]+|@[\w-]+|\$[\w-]+`)},
			{TokenConstant, regexp.MustCompile(`#[0-9a-fA-F]{3,8}\b`)},
			{TokenNumber, regexp.MustCompile(`-?(?:\d+\.?\d*|\.\d+)(?:px|em|rem|%|vh|vw|vmin|vmax|ch|ex|cm|mm|in|pt|pc|s|ms|deg|rad|grad|turn|fr)?`)},
			{TokenKeyword, regexp.MustCompile(`@(?:media|import|keyframes|font-face|supports|charset|namespace|layer|container|property|scope)\b`)},
			{TokenTag, regexp.MustCompile(`\b(?:html|body|div|span|a|p|h[1-6]|ul|ol|li|table|tr|td|th|form|input|button|img|section|article|nav|header|footer|main|aside)\b`)},
			{TokenBuiltin, regexp.MustCompile(`\b(?:inherit|initial|unset|revert|none|auto|normal|bold|italic|block|inline|flex|grid|absolute|relative|fixed|sticky)\b`)},
			{TokenFunction, regexp.MustCompile(`\b([a-zA-Z_-][\w-]*)\s*\(`)},
		},
	}
}
