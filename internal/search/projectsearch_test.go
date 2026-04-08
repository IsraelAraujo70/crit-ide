package search

import (
	"testing"
)

func TestSplitRgLine(t *testing.T) {
	tests := []struct {
		line   string
		expect []string
	}{
		{
			line:   "/home/user/file.go:10:5:matched text here",
			expect: []string{"/home/user/file.go", "10", "5", "matched text here"},
		},
		{
			line:   "src/main.rs:1:1:fn main() {",
			expect: []string{"src/main.rs", "1", "1", "fn main() {"},
		},
		{
			line: "invalid line",
		},
		{
			line: "only:one:colon",
		},
	}

	for _, tt := range tests {
		parts := splitRgLine(tt.line)
		if tt.expect == nil {
			if parts != nil {
				t.Errorf("splitRgLine(%q) = %v, expected nil", tt.line, parts)
			}
			continue
		}
		if parts == nil {
			t.Errorf("splitRgLine(%q) = nil, expected %v", tt.line, tt.expect)
			continue
		}
		for i, p := range tt.expect {
			if parts[i] != p {
				t.Errorf("splitRgLine(%q)[%d] = %q, expected %q", tt.line, i, parts[i], p)
			}
		}
	}
}

func TestSplitGrepLine(t *testing.T) {
	tests := []struct {
		line   string
		expect []string
	}{
		{
			line:   "/home/user/file.go:10:matched text here",
			expect: []string{"/home/user/file.go", "10", "matched text here"},
		},
		{
			line: "invalid",
		},
		{
			line: "only:one",
		},
	}

	for _, tt := range tests {
		parts := splitGrepLine(tt.line)
		if tt.expect == nil {
			if parts != nil {
				t.Errorf("splitGrepLine(%q) = %v, expected nil", tt.line, parts)
			}
			continue
		}
		if parts == nil {
			t.Errorf("splitGrepLine(%q) = nil, expected %v", tt.line, tt.expect)
			continue
		}
		for i, p := range tt.expect {
			if parts[i] != p {
				t.Errorf("splitGrepLine(%q)[%d] = %q, expected %q", tt.line, i, parts[i], p)
			}
		}
	}
}

func TestGroupByFile(t *testing.T) {
	results := []SearchResult{
		{Path: "/root/a.go", Line: 1, Col: 1, MatchText: "line 1"},
		{Path: "/root/a.go", Line: 5, Col: 3, MatchText: "line 5"},
		{Path: "/root/b.go", Line: 10, Col: 1, MatchText: "line 10"},
		{Path: "/root/a.go", Line: 20, Col: 1, MatchText: "line 20"},
	}

	groups := groupByFile(results, "/root")

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	if groups[0].RelPath != "a.go" {
		t.Errorf("group 0 relPath: expected %q, got %q", "a.go", groups[0].RelPath)
	}
	if len(groups[0].Results) != 3 {
		t.Errorf("group 0 results: expected 3, got %d", len(groups[0].Results))
	}

	if groups[1].RelPath != "b.go" {
		t.Errorf("group 1 relPath: expected %q, got %q", "b.go", groups[1].RelPath)
	}
	if len(groups[1].Results) != 1 {
		t.Errorf("group 1 results: expected 1, got %d", len(groups[1].Results))
	}
}

func TestGroupByFile_Empty(t *testing.T) {
	groups := groupByFile(nil, "/root")
	if groups != nil {
		t.Errorf("expected nil for empty results, got %v", groups)
	}
}

func TestFlatten(t *testing.T) {
	groups := []FileGroup{
		{
			Path:    "/root/a.go",
			RelPath: "a.go",
			Results: []SearchResult{
				{Path: "/root/a.go", Line: 1, Col: 1, MatchText: "hello world"},
				{Path: "/root/a.go", Line: 5, Col: 3, MatchText: "hello again"},
			},
		},
		{
			Path:    "/root/b.go",
			RelPath: "b.go",
			Results: []SearchResult{
				{Path: "/root/b.go", Line: 10, Col: 1, MatchText: "goodbye"},
			},
		},
	}

	entries := Flatten(groups)

	// Expected: header + 2 results + header + 1 result = 5
	if len(entries) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(entries))
	}

	// First entry is a header.
	if !entries[0].IsHeader {
		t.Error("entry 0 should be header")
	}
	if entries[0].Path != "/root/a.go" {
		t.Errorf("entry 0 path: expected /root/a.go, got %s", entries[0].Path)
	}

	// Second entry is a result.
	if entries[1].IsHeader {
		t.Error("entry 1 should not be header")
	}
	if entries[1].Line != 1 {
		t.Errorf("entry 1 line: expected 1, got %d", entries[1].Line)
	}

	// Fourth entry is a header for b.go.
	if !entries[3].IsHeader {
		t.Error("entry 3 should be header")
	}

	// Fifth entry is a result from b.go.
	if entries[4].Line != 10 {
		t.Errorf("entry 4 line: expected 10, got %d", entries[4].Line)
	}
}

func TestFlatten_Empty(t *testing.T) {
	entries := Flatten(nil)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for nil groups, got %d", len(entries))
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	groups, count := Search("", "/tmp")
	if groups != nil || count != 0 {
		t.Error("expected nil groups and 0 count for empty query")
	}
}

func TestSearchRegex_InvalidPattern(t *testing.T) {
	groups, count := SearchRegex("[invalid", "/tmp")
	if groups != nil || count != 0 {
		t.Error("expected nil groups and 0 count for invalid regex")
	}
}
