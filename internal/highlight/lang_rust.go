package highlight

import "regexp"

// LangRust returns the language definition for Rust.
func LangRust() *LanguageDef {
	return &LanguageDef{
		ID:                "rust",
		Extensions:        []string{".rs"},
		LineComment:       "//",
		BlockCommentOpen:  "/*",
		BlockCommentClose: "*/",
		Patterns: []PatternRule{
			{TokenComment, regexp.MustCompile(`//.*`)},
			{TokenString, regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)},
			{TokenString, regexp.MustCompile(`r#*"[\s\S]*?"#*`)},
			{TokenString, regexp.MustCompile(`'[^'\\]'|'\\.'`)},
			{TokenPreproc, regexp.MustCompile(`#!\[.*?\]|#\[.*?\]`)},
			{TokenKeyword, regexp.MustCompile(`\b(?:as|async|await|break|const|continue|crate|dyn|else|enum|extern|fn|for|if|impl|in|let|loop|match|mod|move|mut|pub|ref|return|self|Self|static|struct|super|trait|type|union|unsafe|use|where|while|yield)\b`)},
			{TokenTypeName, regexp.MustCompile(`\b(?:bool|char|f32|f64|i8|i16|i32|i64|i128|isize|str|u8|u16|u32|u64|u128|usize|String|Vec|Box|Rc|Arc|Option|Result|HashMap|HashSet|BTreeMap|BTreeSet)\b`)},
			{TokenBuiltin, regexp.MustCompile(`\b(?:println|print|eprintln|eprint|format|vec|todo|unimplemented|unreachable|panic|assert|assert_eq|assert_ne|dbg|cfg|include|include_str|include_bytes|env|concat|stringify|line|column|file|module_path)\b!?`)},
			{TokenConstant, regexp.MustCompile(`\b(?:true|false|None|Some|Ok|Err)\b`)},
			{TokenNumber, regexp.MustCompile(`\b(?:0[xX][0-9a-fA-F_]+|0[oO][0-7_]+|0[bB][01_]+|[0-9][0-9_]*(?:\.[0-9_]+)?(?:[eE][+-]?[0-9_]+)?(?:f32|f64|i8|i16|i32|i64|i128|isize|u8|u16|u32|u64|u128|usize)?)\b`)},
			{TokenVariable, regexp.MustCompile(`'[a-zA-Z_]\w*`)},
			{TokenFunction, regexp.MustCompile(`\b([a-zA-Z_]\w*)\s*[(<]`)},
			{TokenOperator, regexp.MustCompile(`(?:=>|->|&&|\|\||::|\.\.=?|[+\-*/%&|^<>=!]=?)`)},
		},
	}
}
