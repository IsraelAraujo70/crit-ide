package highlight

// TSLangJSON returns a language definition for JSON.
// No tree-sitter grammar is bundled; this entry exists so that the registry
// can detect .json files and populate LanguageID for LSP integration.
func TSLangJSON() *TSLangDef {
	return &TSLangDef{
		ID:         "json",
		Extensions: []string{".json", ".jsonc"},
		FileNames:  []string{".prettierrc", "tsconfig.json"},
		Language:   nil,
		NodeMap:    nil,
	}
}
