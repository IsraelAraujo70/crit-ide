package highlight

// TokenType identifies a semantic highlight category.
type TokenType int

const (
	TokenNone     TokenType = iota
	TokenKeyword            // Language keywords (func, if, return, etc.)
	TokenString             // String literals.
	TokenComment            // Line and block comments.
	TokenFunction           // Function/method names.
	TokenTypeName           // Type names.
	TokenNumber             // Numeric literals.
	TokenOperator           // Operators (+, -, =, etc.)
	TokenBuiltin            // Built-in functions/identifiers.
	TokenTag                // HTML/XML tags.
	TokenAttribute          // HTML/XML attributes.
	TokenHeading            // Markdown headings.
	TokenBold               // Markdown bold text.
	TokenItalic             // Markdown italic text.
	TokenLink               // Markdown/HTML links.
	TokenProperty           // Object keys, CSS properties, TOML keys.
	TokenVariable           // Variables, shell $vars.
	TokenConstant           // Constants, booleans, null.
	TokenPreproc            // Preprocessor directives (#include, #define).
)

// Token represents a highlighted span within a single line.
// Start and End are byte offsets within the line string.
// Start is inclusive, End is exclusive.
type Token struct {
	Start int
	End   int
	Type  TokenType
}
