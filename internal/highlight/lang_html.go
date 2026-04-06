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
			{TokenTag, regexp.MustCompile(`</?\w+`)},          // Opening part: <div, </div, <a
			{TokenTag, regexp.MustCompile(`/?>`)},              // Closing part: > or />
			{TokenAttribute, regexp.MustCompile(`\b([a-zA-Z_:][a-zA-Z0-9_:.-]*)\s*=`)}, // attr name before =
			{TokenString, regexp.MustCompile(`"[^"]*"|'[^']*'`)},
			{TokenConstant, regexp.MustCompile(`&\w+;|&#\d+;|&#x[0-9a-fA-F]+;`)},
		},
	}
}
