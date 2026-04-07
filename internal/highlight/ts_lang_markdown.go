package highlight

// TSLangMarkdown returns a language definition for Markdown.
// The markdown tree-sitter grammar uses a two-parser architecture that is
// not compatible with the single-parser TreeSitterHighlighter. This entry
// exists so that the registry can detect .md files and populate LanguageID.
func TSLangMarkdown() *TSLangDef {
	return &TSLangDef{
		ID:         "markdown",
		Extensions: []string{".md", ".mdx"},
		FileNames:  []string{"README.md"},
		Language:   nil,
		NodeMap:    nil,
	}
}
