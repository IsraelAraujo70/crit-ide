package highlight

import (
	"testing"
)

func TestGoKeywords(t *testing.T) {
	h := NewRegexHighlighter()
	h.SetLanguageDef(LangGo())

	tokens := h.HighlightLine(0, "func main() {")
	found := false
	for _, tok := range tokens {
		if tok.Type == TokenKeyword && tok.Start == 0 && tok.End == 4 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected keyword token for 'func', got tokens: %v", tokens)
	}
}

func TestGoString(t *testing.T) {
	h := NewRegexHighlighter()
	h.SetLanguageDef(LangGo())

	tokens := h.HighlightLine(0, `x := "hello world"`)
	found := false
	for _, tok := range tokens {
		if tok.Type == TokenString {
			found = true
		}
	}
	if !found {
		t.Errorf("expected string token, got tokens: %v", tokens)
	}
}

func TestGoComment(t *testing.T) {
	h := NewRegexHighlighter()
	h.SetLanguageDef(LangGo())

	tokens := h.HighlightLine(0, "x := 1 // comment")
	found := false
	for _, tok := range tokens {
		if tok.Type == TokenComment {
			found = true
		}
	}
	if !found {
		t.Errorf("expected comment token, got tokens: %v", tokens)
	}
}

func TestGoBlockComment(t *testing.T) {
	h := NewRegexHighlighter()
	h.SetLanguageDef(LangGo())

	// Line 0 opens a block comment.
	tokens0 := h.HighlightLine(0, "x := 1 /* start")
	_ = tokens0

	// Line 1 should be entirely a comment.
	tokens1 := h.HighlightLine(1, "still in comment")
	if len(tokens1) != 1 || tokens1[0].Type != TokenComment {
		t.Errorf("expected full line comment, got: %v", tokens1)
	}

	// Line 2 closes the block comment.
	tokens2 := h.HighlightLine(2, "end */ x := 2")
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
	h := NewRegexHighlighter()
	h.SetLanguageDef(LangGo())

	tokens := h.HighlightLine(0, "x := 42")
	found := false
	for _, tok := range tokens {
		if tok.Type == TokenNumber {
			found = true
		}
	}
	if !found {
		t.Errorf("expected number token, got: %v", tokens)
	}
}

func TestGoFunction(t *testing.T) {
	h := NewRegexHighlighter()
	h.SetLanguageDef(LangGo())

	tokens := h.HighlightLine(0, "fmt.Println(x)")
	found := false
	for _, tok := range tokens {
		if tok.Type == TokenFunction {
			found = true
		}
	}
	if !found {
		t.Errorf("expected function token, got: %v", tokens)
	}
}

func TestPythonKeywords(t *testing.T) {
	h := NewRegexHighlighter()
	h.SetLanguageDef(LangPython())

	tokens := h.HighlightLine(0, "def foo(x):")
	found := false
	for _, tok := range tokens {
		if tok.Type == TokenKeyword && tok.Start == 0 && tok.End == 3 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected keyword 'def', got: %v", tokens)
	}
}

func TestPythonComment(t *testing.T) {
	h := NewRegexHighlighter()
	h.SetLanguageDef(LangPython())

	tokens := h.HighlightLine(0, "x = 1 # comment")
	found := false
	for _, tok := range tokens {
		if tok.Type == TokenComment {
			found = true
		}
	}
	if !found {
		t.Errorf("expected comment token, got: %v", tokens)
	}
}

func TestJSTemplateLiteral(t *testing.T) {
	h := NewRegexHighlighter()
	h.SetLanguageDef(LangJavaScript())

	tokens := h.HighlightLine(0, "const x = `hello`")
	found := false
	for _, tok := range tokens {
		if tok.Type == TokenString {
			found = true
		}
	}
	if !found {
		t.Errorf("expected string token for template literal, got: %v", tokens)
	}
}

func TestRustLifetime(t *testing.T) {
	h := NewRegexHighlighter()
	h.SetLanguageDef(LangRust())

	tokens := h.HighlightLine(0, "fn foo<'a>(x: &'a str) {")
	foundLifetime := false
	for _, tok := range tokens {
		if tok.Type == TokenVariable {
			foundLifetime = true
		}
	}
	if !foundLifetime {
		t.Errorf("expected lifetime variable token, got: %v", tokens)
	}
}

func TestRustKeywords(t *testing.T) {
	h := NewRegexHighlighter()
	h.SetLanguageDef(LangRust())

	tokens := h.HighlightLine(0, "let mut x = 5;")
	foundLet := false
	foundMut := false
	for _, tok := range tokens {
		if tok.Type == TokenKeyword {
			switch {
			case tok.Start == 0 && tok.End == 3:
				foundLet = true
			case tok.Start == 4 && tok.End == 7:
				foundMut = true
			}
		}
	}
	if !foundLet || !foundMut {
		t.Errorf("expected 'let' and 'mut' keywords, got: %v", tokens)
	}
}

func TestNoLanguage(t *testing.T) {
	h := NewRegexHighlighter()
	tokens := h.HighlightLine(0, "hello world")
	if tokens != nil {
		t.Errorf("expected nil tokens with no language, got: %v", tokens)
	}
}

func TestInvalidateFrom(t *testing.T) {
	h := NewRegexHighlighter()
	h.SetLanguageDef(LangGo())

	// Build up state cache.
	h.HighlightLine(0, "func main() {")
	h.HighlightLine(1, "x := 1")
	h.HighlightLine(2, "}")

	// Invalidate from line 1.
	h.InvalidateFrom(1)
	if len(h.stateCache) > 1 {
		t.Errorf("expected stateCache truncated to 1, got len %d", len(h.stateCache))
	}
}

func TestLangRegistryDetection(t *testing.T) {
	r := DefaultRegistry()

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
		{"README.md", "markdown"},
		{"config.json", "json"},
		{"index.html", "html"},
		{"style.css", "css"},
		{"run.sh", "shell"},
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
	h := NewRegexHighlighter()
	h.SetLanguageDef(LangGo())

	tokens := h.HighlightLine(0, `func foo(x int) string { return "hello" }`)
	for i := 1; i < len(tokens); i++ {
		if tokens[i].Start < tokens[i-1].Start {
			t.Errorf("tokens not sorted: [%d].Start=%d < [%d].Start=%d",
				i, tokens[i].Start, i-1, tokens[i-1].Start)
		}
	}
}
