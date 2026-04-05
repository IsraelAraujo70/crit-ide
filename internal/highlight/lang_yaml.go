package highlight

import "regexp"

// LangYAML returns the language definition for YAML.
func LangYAML() *LanguageDef {
	return &LanguageDef{
		ID:          "yaml",
		Extensions:  []string{".yaml", ".yml"},
		LineComment: "#",
		Patterns: []PatternRule{
			{TokenComment, regexp.MustCompile(`#.*`)},
			{TokenProperty, regexp.MustCompile(`^[\s-]*[\w.-]+\s*:`)},
			{TokenString, regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)},
			{TokenString, regexp.MustCompile(`'(?:[^'\\]|\\.)*'`)},
			{TokenVariable, regexp.MustCompile(`&\w+|\*\w+`)},
			{TokenConstant, regexp.MustCompile(`\b(?:true|false|yes|no|on|off|null|~)\b`)},
			{TokenTag, regexp.MustCompile(`!!\w+|!\w+`)},
			{TokenKeyword, regexp.MustCompile(`^---\s*$|^\.\.\.\s*$`)},
			{TokenNumber, regexp.MustCompile(`\b(?:0[xX][0-9a-fA-F]+|0[oO][0-7]+|[+-]?(?:[0-9]+(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?|\.inf|\.nan))\b`)},
		},
	}
}
