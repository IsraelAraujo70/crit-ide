package highlight

import "regexp"

// LangHTML returns the language definition for HTML/XML.
func LangHTML() *LanguageDef {
	return &LanguageDef{
		ID:                "html",
		Extensions:        []string{".html", ".htm", ".xhtml", ".xml", ".svg"},
		BlockCommentOpen:  "<!--",
		BlockCommentClose: "-->",
		Patterns: []PatternRule{
			{TokenComment, regexp.MustCompile(`<!--.*?-->`)},
			{TokenString, regexp.MustCompile(`"[^"]*"|'[^']*'`)},
			{TokenTag, regexp.MustCompile(`</?\w[^>]*>`)},
			{TokenAttribute, regexp.MustCompile(`\b([a-zA-Z_:][a-zA-Z0-9_:.-]*)\s*=`)},
			{TokenConstant, regexp.MustCompile(`&\w+;|&#\d+;|&#x[0-9a-fA-F]+;`)},
		},
	}
}
