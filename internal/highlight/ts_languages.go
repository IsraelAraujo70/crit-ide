package highlight

import (
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// TSLangDef defines a tree-sitter language with its node-to-token mapping.
type TSLangDef struct {
	ID        string
	Extensions []string
	FileNames  []string
	Language   *sitter.Language
	NodeMap    map[string]TokenType
}

// TSLangRegistry maps file extensions and names to tree-sitter language definitions.
type TSLangRegistry struct {
	byExtension map[string]*TSLangDef
	byFileName  map[string]*TSLangDef
	byID        map[string]*TSLangDef
}

// NewTSLangRegistry creates an empty tree-sitter language registry.
func NewTSLangRegistry() *TSLangRegistry {
	return &TSLangRegistry{
		byExtension: make(map[string]*TSLangDef),
		byFileName:  make(map[string]*TSLangDef),
		byID:        make(map[string]*TSLangDef),
	}
}

// Register adds a language definition to the registry.
func (r *TSLangRegistry) Register(def *TSLangDef) {
	r.byID[def.ID] = def
	for _, ext := range def.Extensions {
		r.byExtension[ext] = def
	}
	for _, name := range def.FileNames {
		r.byFileName[name] = def
	}
}

// DetectLanguage returns the language definition for a file path, or nil.
func (r *TSLangRegistry) DetectLanguage(filename string) *TSLangDef {
	base := filepath.Base(filename)
	if def, ok := r.byFileName[base]; ok {
		return def
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if def, ok := r.byExtension[ext]; ok {
		return def
	}
	return nil
}

// ByID returns the language definition with the given ID, or nil.
func (r *TSLangRegistry) ByID(id string) *TSLangDef {
	return r.byID[id]
}

// DefaultTSRegistry returns a TSLangRegistry pre-loaded with all built-in languages.
func DefaultTSRegistry() *TSLangRegistry {
	r := NewTSLangRegistry()
	r.Register(TSLangGo())
	r.Register(TSLangPython())
	r.Register(TSLangJavaScript())
	r.Register(TSLangTypeScript())
	r.Register(TSLangRust())
	r.Register(TSLangC())
	r.Register(TSLangCPP())
	r.Register(TSLangHTML())
	r.Register(TSLangCSS())
	r.Register(TSLangBash())
	r.Register(TSLangJSON())
	r.Register(TSLangMarkdown())
	r.Register(TSLangTOML())
	r.Register(TSLangYAML())
	return r
}
