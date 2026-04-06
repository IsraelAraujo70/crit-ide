package highlight

import "regexp"

// LangMarkdown returns the language definition for Markdown.
func LangMarkdown() *LanguageDef {
	return &LanguageDef{
		ID:         "markdown",
		Extensions: []string{".md", ".markdown", ".mkd", ".mdx"},
		Patterns: []PatternRule{
			{TokenString, regexp.MustCompile("```.*")},
			{TokenHeading, regexp.MustCompile(`^#{1,6}\s+.*`)},
			{TokenBold, regexp.MustCompile(`\*\*[^*]+\*\*|__[^_]+__`)},
			{TokenItalic, regexp.MustCompile(`\*[^*]+\*|_[^_]+_`)},
			{TokenString, regexp.MustCompile("`[^`]+`")},
			{TokenLink, regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)},
			{TokenLink, regexp.MustCompile(`https?://\S+`)},
			{TokenOperator, regexp.MustCompile(`^(?:\s*[-*+]|\s*\d+\.)\s`)},
			{TokenComment, regexp.MustCompile(`^>\s+.*`)},
			{TokenProperty, regexp.MustCompile(`^---\s*$`)},
		},
	}
}
