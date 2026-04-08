package git

import (
	"testing"
)

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []FileStatus
	}{
		{
			name:  "modified staged",
			input: "M  internal/app/app.go\n",
			expect: []FileStatus{
				{Path: "internal/app/app.go", Status: StatusModified, Staged: true},
			},
		},
		{
			name:  "modified unstaged",
			input: " M internal/app/app.go\n",
			expect: []FileStatus{
				{Path: "internal/app/app.go", Status: StatusModified, Staged: false},
			},
		},
		{
			name:  "untracked",
			input: "?? newfile.go\n",
			expect: []FileStatus{
				{Path: "newfile.go", Status: StatusUntracked, Staged: false},
			},
		},
		{
			name:  "added staged",
			input: "A  newfile.go\n",
			expect: []FileStatus{
				{Path: "newfile.go", Status: StatusAdded, Staged: true},
			},
		},
		{
			name:  "deleted unstaged",
			input: " D oldfile.go\n",
			expect: []FileStatus{
				{Path: "oldfile.go", Status: StatusDeleted, Staged: false},
			},
		},
		{
			name:  "renamed staged",
			input: "R  old.go -> new.go\n",
			expect: []FileStatus{
				{Path: "new.go", Status: StatusRenamed, Staged: true},
			},
		},
		{
			name:  "both staged and unstaged",
			input: "MM internal/app/app.go\n",
			expect: []FileStatus{
				{Path: "internal/app/app.go", Status: StatusModified, Staged: true},
				{Path: "internal/app/app.go", Status: StatusModified, Staged: false},
			},
		},
		{
			name: "multiple files",
			input: "M  file1.go\n" +
				" M file2.go\n" +
				"?? file3.go\n",
			expect: []FileStatus{
				{Path: "file1.go", Status: StatusModified, Staged: true},
				{Path: "file2.go", Status: StatusModified, Staged: false},
				{Path: "file3.go", Status: StatusUntracked, Staged: false},
			},
		},
		{
			name:   "empty output",
			input:  "",
			expect: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseStatus(tt.input)
			if len(result) != len(tt.expect) {
				t.Fatalf("expected %d entries, got %d: %+v", len(tt.expect), len(result), result)
			}
			for i, exp := range tt.expect {
				got := result[i]
				if got.Path != exp.Path {
					t.Errorf("[%d] path: expected %q, got %q", i, exp.Path, got.Path)
				}
				if got.Status != exp.Status {
					t.Errorf("[%d] status: expected %q, got %q", i, exp.Status, got.Status)
				}
				if got.Staged != exp.Staged {
					t.Errorf("[%d] staged: expected %v, got %v", i, exp.Staged, got.Staged)
				}
			}
		})
	}
}

func TestParseLog(t *testing.T) {
	input := "abc1234\x00John Doe\x002024-01-15T10:30:00Z\x00feat: add feature\x00HEAD -> main\x00parent1 parent2\n"
	commits := parseLog(input)
	if len(commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(commits))
	}
	c := commits[0]
	if c.Hash != "abc1234" {
		t.Errorf("hash: expected abc1234, got %s", c.Hash)
	}
	if c.Author != "John Doe" {
		t.Errorf("author: expected John Doe, got %s", c.Author)
	}
	if c.Message != "feat: add feature" {
		t.Errorf("message: expected 'feat: add feature', got %q", c.Message)
	}
	if c.Refs != "HEAD -> main" {
		t.Errorf("refs: expected 'HEAD -> main', got %q", c.Refs)
	}
	if len(c.Parents) != 2 || c.Parents[0] != "parent1" || c.Parents[1] != "parent2" {
		t.Errorf("parents: expected [parent1 parent2], got %v", c.Parents)
	}
}

func TestParseBranches(t *testing.T) {
	input := "*main\x00origin/main\n feat/test\x00origin/feat/test\n develop\x00\n"
	branches := parseBranches(input)
	if len(branches) != 3 {
		t.Fatalf("expected 3 branches, got %d", len(branches))
	}
	if !branches[0].Current || branches[0].Name != "main" || branches[0].Remote != "origin/main" {
		t.Errorf("branch 0: %+v", branches[0])
	}
	if branches[1].Current || branches[1].Name != "feat/test" || branches[1].Remote != "origin/feat/test" {
		t.Errorf("branch 1: %+v", branches[1])
	}
	if branches[2].Current || branches[2].Name != "develop" || branches[2].Remote != "" {
		t.Errorf("branch 2: %+v", branches[2])
	}
}

func TestParseHunkHeader(t *testing.T) {
	tests := []struct {
		line     string
		newStart int
		newCount int
		oldCount int
	}{
		{"@@ -10,3 +15,5 @@ func foo()", 15, 5, 3},
		{"@@ -1 +1,2 @@", 1, 2, 1},
		{"@@ -5,0 +6,3 @@", 6, 3, 0},
		{"@@ -5,2 +5,0 @@", 5, 0, 2},
	}

	for _, tt := range tests {
		newStart, newCount, oldCount := parseHunkHeader(tt.line)
		if newStart != tt.newStart || newCount != tt.newCount || oldCount != tt.oldCount {
			t.Errorf("parseHunkHeader(%q) = (%d, %d, %d), want (%d, %d, %d)",
				tt.line, newStart, newCount, oldCount, tt.newStart, tt.newCount, tt.oldCount)
		}
	}
}

func TestParseGutterDiff(t *testing.T) {
	input := `diff --git a/file.go b/file.go
index abc..def 100644
--- a/file.go
+++ b/file.go
@@ -5,0 +6,3 @@ func foo()
+line1
+line2
+line3
@@ -10,2 +14,0 @@ func bar()
-old1
-old2
@@ -15,1 +16,1 @@ func baz()
-old
+new
`
	diffs := parseGutterDiff(input)
	if len(diffs) < 3 {
		t.Fatalf("expected at least 3 diff entries, got %d: %+v", len(diffs), diffs)
	}

	// First hunk: 3 added lines starting at line 6 (zero-based: 5, 6, 7).
	if diffs[0].Line != 5 || diffs[0].Status != DiffAdded {
		t.Errorf("diff[0]: expected {Line:5, Status:DiffAdded}, got %+v", diffs[0])
	}

	// Second hunk: pure deletion at line 14 (zero-based: 13).
	found := false
	for _, d := range diffs {
		if d.Status == DiffRemoved {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected at least one DiffRemoved entry")
	}
}
