package highlight

import "regexp"

// LangPython returns the language definition for Python.
func LangPython() *LanguageDef {
	return &LanguageDef{
		ID:           "python",
		Extensions:   []string{".py", ".pyi", ".pyw"},
		LineComment:  "#",
		Patterns: []PatternRule{
			{TokenComment, regexp.MustCompile(`#.*`)},
			{TokenString, regexp.MustCompile(`"""[\s\S]*?"""`)},
			{TokenString, regexp.MustCompile(`'''[\s\S]*?'''`)},
			{TokenString, regexp.MustCompile(`[fFrRbBuU]?"(?:[^"\\]|\\.)*"`)},
			{TokenString, regexp.MustCompile(`[fFrRbBuU]?'(?:[^'\\]|\\.)*'`)},
			{TokenKeyword, regexp.MustCompile(`\b(?:and|as|assert|async|await|break|class|continue|def|del|elif|else|except|finally|for|from|global|if|import|in|is|lambda|nonlocal|not|or|pass|raise|return|try|while|with|yield)\b`)},
			{TokenBuiltin, regexp.MustCompile(`\b(?:abs|all|any|bin|bool|bytes|callable|chr|classmethod|compile|complex|delattr|dict|dir|divmod|enumerate|eval|exec|filter|float|format|frozenset|getattr|globals|hasattr|hash|help|hex|id|input|int|isinstance|issubclass|iter|len|list|locals|map|max|memoryview|min|next|object|oct|open|ord|pow|print|property|range|repr|reversed|round|set|setattr|slice|sorted|staticmethod|str|sum|super|tuple|type|vars|zip)\b`)},
			{TokenConstant, regexp.MustCompile(`\b(?:True|False|None)\b`)},
			{TokenTypeName, regexp.MustCompile(`\b(?:int|float|str|bool|bytes|list|dict|set|tuple|complex|frozenset|bytearray|memoryview|type|object)\b`)},
			{TokenNumber, regexp.MustCompile(`\b(?:0[xX][0-9a-fA-F_]+|0[oO][0-7_]+|0[bB][01_]+|[0-9][0-9_]*(?:\.[0-9_]+)?(?:[eE][+-]?[0-9_]+)?j?)\b`)},
			{TokenFunction, regexp.MustCompile(`\b([a-zA-Z_]\w*)\s*\(`)},
			{TokenPreproc, regexp.MustCompile(`@\w+`)},
		},
	}
}
