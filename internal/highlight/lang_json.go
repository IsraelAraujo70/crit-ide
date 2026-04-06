package highlight

import "regexp"

// LangJSON returns the language definition for JSON.
func LangJSON() *LanguageDef {
	return &LanguageDef{
		ID:         "json",
		Extensions: []string{".json", ".jsonc", ".json5"},
		Patterns: []PatternRule{
			{TokenComment, regexp.MustCompile(`//.*`)},
			{TokenProperty, regexp.MustCompile(`"(?:[^"\\]|\\.)*"\s*:`)},
			{TokenString, regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)},
			{TokenConstant, regexp.MustCompile(`\b(?:true|false|null)\b`)},
			{TokenNumber, regexp.MustCompile(`-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?`)},
		},
	}
}
