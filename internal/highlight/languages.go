package highlight

import (
	"path/filepath"
	"regexp"
	"strings"
)

// PatternRule maps a regex pattern to a token type. Patterns are applied in
// order; earlier rules take priority when spans overlap.
type PatternRule struct {
	Type    TokenType
	Pattern *regexp.Regexp
}

// LanguageDef defines syntax patterns for a programming language.
type LanguageDef struct {
	ID                string
	Extensions        []string // e.g., [".go", ".mod"]
	FileNames         []string // Exact filenames (e.g., "Makefile")
	LineComment       string   // e.g., "//"
	BlockCommentOpen  string   // e.g., "/*"
	BlockCommentClose string   // e.g., "*/"
	Patterns          []PatternRule
}

// LangRegistry maps file extensions and names to language definitions.
type LangRegistry struct {
	byExtension map[string]*LanguageDef
	byFileName  map[string]*LanguageDef
	byID        map[string]*LanguageDef
}

// NewLangRegistry creates an empty language registry.
func NewLangRegistry() *LangRegistry {
	return &LangRegistry{
		byExtension: make(map[string]*LanguageDef),
		byFileName:  make(map[string]*LanguageDef),
		byID:        make(map[string]*LanguageDef),
	}
}

// Register adds a language definition to the registry.
func (r *LangRegistry) Register(def *LanguageDef) {
	r.byID[def.ID] = def
	for _, ext := range def.Extensions {
		r.byExtension[ext] = def
	}
	for _, name := range def.FileNames {
		r.byFileName[name] = def
	}
}

// DetectLanguage returns the language definition for a file path, or nil.
func (r *LangRegistry) DetectLanguage(filename string) *LanguageDef {
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
func (r *LangRegistry) ByID(id string) *LanguageDef {
	return r.byID[id]
}

// DefaultRegistry returns a LangRegistry pre-loaded with all built-in languages.
func DefaultRegistry() *LangRegistry {
	r := NewLangRegistry()
	r.Register(LangGo())
	r.Register(LangPython())
	r.Register(LangJavaScript())
	r.Register(LangTypeScript())
	r.Register(LangRust())
	r.Register(LangC())
	r.Register(LangCPP())
	r.Register(LangMarkdown())
	r.Register(LangJSON())
	r.Register(LangHTML())
	r.Register(LangCSS())
	r.Register(LangShell())
	r.Register(LangTOML())
	r.Register(LangYAML())
	return r
}
