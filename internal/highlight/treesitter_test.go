package highlight

import (
	"strings"
	"testing"
)

func newTestHighlighter(langID string, source string) *TreeSitterHighlighter {
	reg := DefaultTSRegistry()
	h := NewTreeSitterHighlighter(reg)
	h.SetLanguage(langID)
	h.SetSource(source)
	return h
}

func hasTokenType(tokens []Token, tt TokenType) bool {
	for _, tok := range tokens {
		if tok.Type == tt {
			return true
		}
	}
	return false
}

func hasTokenAt(tokens []Token, tt TokenType, start, end int) bool {
	for _, tok := range tokens {
		if tok.Type == tt && tok.Start == start && tok.End == end {
			return true
		}
	}
	return false
}

func TestGoKeywords(t *testing.T) {
	src := "package main\n\nfunc main() {\n}\n"
	h := newTestHighlighter("go", src)

	// Line 2: "func main() {"
	tokens := h.HighlightLine(2, "func main() {")
	if !hasTokenAt(tokens, TokenKeyword, 0, 4) {
		t.Errorf("expected keyword token for 'func', got tokens: %v", tokens)
	}
}

func TestGoString(t *testing.T) {
	src := "package main\n\nvar x = \"hello world\"\n"
	h := newTestHighlighter("go", src)

	tokens := h.HighlightLine(2, `var x = "hello world"`)
	if !hasTokenType(tokens, TokenString) {
		t.Errorf("expected string token, got tokens: %v", tokens)
	}
}

func TestGoComment(t *testing.T) {
	src := "package main\n\nx := 1 // comment\n"
	h := newTestHighlighter("go", src)

	tokens := h.HighlightLine(2, "x := 1 // comment")
	if !hasTokenType(tokens, TokenComment) {
		t.Errorf("expected comment token, got tokens: %v", tokens)
	}
}

func TestGoBlockComment(t *testing.T) {
	src := "package main\n\nx := 1 /* start\nstill in comment\nend */ x := 2\n"
	h := newTestHighlighter("go", src)

	// Line 3: "still in comment" should contain a comment token.
	tokens1 := h.HighlightLine(3, "still in comment")
	if !hasTokenType(tokens1, TokenComment) {
		t.Errorf("expected comment token on continuation line, got: %v", tokens1)
	}

	// Line 4: "end */ x := 2" should have comment and non-comment tokens.
	tokens2 := h.HighlightLine(4, "end */ x := 2")
	foundComment := false
	foundOther := false
	for _, tok := range tokens2 {
		if tok.Type == TokenComment {
			foundComment = true
		} else {
			foundOther = true
		}
	}
	if !foundComment {
		t.Error("expected comment token on closing line")
	}
	if !foundOther {
		t.Error("expected non-comment tokens after block comment close")
	}
}

func TestGoNumber(t *testing.T) {
	src := "package main\n\nvar x = 42\n"
	h := newTestHighlighter("go", src)

	tokens := h.HighlightLine(2, "var x = 42")
	if !hasTokenType(tokens, TokenNumber) {
		t.Errorf("expected number token, got: %v", tokens)
	}
}

func TestGoFunction(t *testing.T) {
	src := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(x)\n}\n"
	h := newTestHighlighter("go", src)

	// Line 5: "\tfmt.Println(x)"
	tokens := h.HighlightLine(5, "\tfmt.Println(x)")
	if !hasTokenType(tokens, TokenFunction) {
		t.Errorf("expected function token, got: %v", tokens)
	}
}

func TestGoTypeIdentifier(t *testing.T) {
	src := "package main\n\ntype Foo struct {\n\tBar string\n}\n"
	h := newTestHighlighter("go", src)

	// Line 2: "type Foo struct {"
	tokens := h.HighlightLine(2, "type Foo struct {")
	if !hasTokenType(tokens, TokenTypeName) {
		t.Errorf("expected type name token for 'Foo', got: %v", tokens)
	}
}

func TestPythonKeywords(t *testing.T) {
	src := "def foo(x):\n    pass\n"
	h := newTestHighlighter("python", src)

	tokens := h.HighlightLine(0, "def foo(x):")
	if !hasTokenAt(tokens, TokenKeyword, 0, 3) {
		t.Errorf("expected keyword 'def', got: %v", tokens)
	}
}

func TestPythonComment(t *testing.T) {
	src := "x = 1 # comment\n"
	h := newTestHighlighter("python", src)

	tokens := h.HighlightLine(0, "x = 1 # comment")
	if !hasTokenType(tokens, TokenComment) {
		t.Errorf("expected comment token, got: %v", tokens)
	}
}

func TestJSTemplateLiteral(t *testing.T) {
	src := "const x = `hello`\n"
	h := newTestHighlighter("javascript", src)

	tokens := h.HighlightLine(0, "const x = `hello`")
	if !hasTokenType(tokens, TokenString) {
		t.Errorf("expected string token for template literal, got: %v", tokens)
	}
}

func TestRustKeywords(t *testing.T) {
	src := "fn main() {\n    let mut x = 5;\n}\n"
	h := newTestHighlighter("rust", src)

	tokens := h.HighlightLine(1, "    let mut x = 5;")
	foundLet := false
	foundMut := false
	for _, tok := range tokens {
		if tok.Type == TokenKeyword {
			switch {
			case tok.Start == 4 && tok.End == 7:
				foundLet = true
			case tok.Start == 8 && tok.End == 11:
				foundMut = true
			}
		}
	}
	if !foundLet || !foundMut {
		t.Errorf("expected 'let' and 'mut' keywords, got: %v", tokens)
	}
}

func TestNoLanguage(t *testing.T) {
	reg := DefaultTSRegistry()
	h := NewTreeSitterHighlighter(reg)
	tokens := h.HighlightLine(0, "hello world")
	if tokens != nil {
		t.Errorf("expected nil tokens with no language, got: %v", tokens)
	}
}

func TestInvalidateFrom(t *testing.T) {
	src := "package main\n\nfunc main() {\n\tx := 1\n}\n"
	h := newTestHighlighter("go", src)

	// Force a parse by reading a line.
	h.HighlightLine(0, "package main")

	// Invalidate should mark tree as dirty.
	h.InvalidateFrom(1)
	if !h.dirty {
		t.Error("expected dirty=true after InvalidateFrom")
	}
}

func TestLangRegistryDetection(t *testing.T) {
	r := DefaultTSRegistry()

	tests := []struct {
		filename string
		langID   string
	}{
		{"main.go", "go"},
		{"app.py", "python"},
		{"index.js", "javascript"},
		{"app.tsx", "typescript"},
		{"lib.rs", "rust"},
		{"main.c", "c"},
		{"index.html", "html"},
		{"style.css", "css"},
		{"run.sh", "shell"},
		{"config.json", "json"},
		{"README.md", "markdown"},
		{"config.toml", "toml"},
		{"docker-compose.yml", "yaml"},
		{"unknown.xyz", ""},
	}

	for _, tt := range tests {
		def := r.DetectLanguage(tt.filename)
		if tt.langID == "" {
			if def != nil {
				t.Errorf("DetectLanguage(%q) = %q, want nil", tt.filename, def.ID)
			}
		} else if def == nil {
			t.Errorf("DetectLanguage(%q) = nil, want %q", tt.filename, tt.langID)
		} else if def.ID != tt.langID {
			t.Errorf("DetectLanguage(%q) = %q, want %q", tt.filename, def.ID, tt.langID)
		}
	}
}

func TestTokensSortedByStart(t *testing.T) {
	src := "package main\n\nfunc foo(x int) string { return \"hello\" }\n"
	h := newTestHighlighter("go", src)

	tokens := h.HighlightLine(2, `func foo(x int) string { return "hello" }`)
	for i := 1; i < len(tokens); i++ {
		if tokens[i].Start < tokens[i-1].Start {
			t.Errorf("tokens not sorted: [%d].Start=%d < [%d].Start=%d",
				i, tokens[i].Start, i-1, tokens[i-1].Start)
		}
	}
}

func TestSetSourceUpdatesTree(t *testing.T) {
	src1 := "package main\n\nvar x = 1\n"
	h := newTestHighlighter("go", src1)

	tokens1 := h.HighlightLine(2, "var x = 1")
	if !hasTokenType(tokens1, TokenNumber) {
		t.Errorf("expected number token in initial source, got: %v", tokens1)
	}

	// Update source with a string literal.
	src2 := "package main\n\nvar x = \"hello\"\n"
	h.SetSource(src2)

	tokens2 := h.HighlightLine(2, `var x = "hello"`)
	if !hasTokenType(tokens2, TokenString) {
		t.Errorf("expected string token after source update, got: %v", tokens2)
	}
}

func TestMultipleLanguages(t *testing.T) {
	reg := DefaultTSRegistry()
	h := NewTreeSitterHighlighter(reg)

	// Start with Go.
	h.SetLanguage("go")
	h.SetSource("package main\n\nfunc main() {}\n")
	tokens := h.HighlightLine(2, "func main() {}")
	if !hasTokenType(tokens, TokenKeyword) {
		t.Error("expected keyword in Go source")
	}

	// Switch to Python.
	h.SetLanguage("python")
	h.SetSource("def foo():\n    pass\n")
	tokens = h.HighlightLine(0, "def foo():")
	if !hasTokenType(tokens, TokenKeyword) {
		t.Error("expected keyword in Python source")
	}
}

func TestNilLanguageNoTokens(t *testing.T) {
	// Languages without a tree-sitter grammar (JSON, Markdown) should be
	// recognized by SetLanguage but produce no tokens.
	reg := DefaultTSRegistry()
	h := NewTreeSitterHighlighter(reg)

	if !h.SetLanguage("json") {
		t.Fatal("SetLanguage('json') should return true for registered language")
	}
	h.SetSource(`{"key": "value"}`)
	tokens := h.HighlightLine(0, `{"key": "value"}`)
	if tokens != nil {
		t.Errorf("expected nil tokens for language without grammar, got: %v", tokens)
	}

	if !h.SetLanguage("markdown") {
		t.Fatal("SetLanguage('markdown') should return true for registered language")
	}
	h.SetSource("# Hello\n\nWorld\n")
	tokens = h.HighlightLine(0, "# Hello")
	if tokens != nil {
		t.Errorf("expected nil tokens for markdown without grammar, got: %v", tokens)
	}
}

func TestAllLanguagesLoad(t *testing.T) {
	// Verify all registered languages can parse without panicking.
	langs := []struct {
		id  string
		src string
	}{
		{"go", "package main\nfunc main() {}\n"},
		{"python", "def foo():\n    pass\n"},
		{"javascript", "const x = 1;\n"},
		{"typescript", "const x: number = 1;\n"},
		{"rust", "fn main() {}\n"},
		{"c", "int main() { return 0; }\n"},
		{"cpp", "int main() { return 0; }\n"},
		{"html", "<html><body>Hello</body></html>\n"},
		{"css", ".foo { color: red; }\n"},
		{"shell", "echo hello\n"},
		{"toml", "[section]\nkey = \"value\"\n"},
		{"yaml", "key: value\n"},
	}

	reg := DefaultTSRegistry()
	for _, lang := range langs {
		h := NewTreeSitterHighlighter(reg)
		if !h.SetLanguage(lang.id) {
			t.Errorf("SetLanguage(%q) returned false", lang.id)
			continue
		}
		h.SetSource(lang.src)
		lines := strings.Split(lang.src, "\n")
		for i, line := range lines {
			tokens := h.HighlightLine(i, line)
			_ = tokens // Just verify no panic.
		}
	}
}
