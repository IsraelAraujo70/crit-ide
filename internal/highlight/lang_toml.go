package highlight

import "regexp"

// LangTOML returns the language definition for TOML.
func LangTOML() *LanguageDef {
	return &LanguageDef{
		ID:          "toml",
		Extensions:  []string{".toml"},
		FileNames:   []string{"Cargo.toml", "pyproject.toml"},
		LineComment: "#",
		Patterns: []PatternRule{
			{TokenComment, regexp.MustCompile(`#.*`)},
			{TokenHeading, regexp.MustCompile(`^\s*\[\[?[^\]]+\]\]?`)},
			{TokenString, regexp.MustCompile(`"""[\s\S]*?"""`)},
			{TokenString, regexp.MustCompile(`'''[\s\S]*?'''`)},
			{TokenString, regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)},
			{TokenString, regexp.MustCompile(`'[^']*'`)},
			{TokenProperty, regexp.MustCompile(`^[\w.-]+\s*=`)},
			{TokenConstant, regexp.MustCompile(`\b(?:true|false)\b`)},
			{TokenNumber, regexp.MustCompile(`(?:0[xX][0-9a-fA-F_]+|0[oO][0-7_]+|0[bB][01_]+|[+-]?(?:[0-9][0-9_]*(?:\.[0-9_]+)?(?:[eE][+-]?[0-9_]+)?|inf|nan))`)},
			{TokenTypeName, regexp.MustCompile(`\d{4}-\d{2}-\d{2}(?:[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?)?`)},
		},
	}
}
