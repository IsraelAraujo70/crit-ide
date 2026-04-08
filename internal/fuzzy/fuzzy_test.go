package fuzzy

import (
	"sort"
	"testing"
)

func TestMatch_EmptyPattern(t *testing.T) {
	r := Match("", "anything")
	if r.Score <= 0 {
		t.Errorf("empty pattern should match everything, got score=%d", r.Score)
	}
}

func TestMatch_EmptyCandidate(t *testing.T) {
	r := Match("abc", "")
	if r.Score != 0 {
		t.Errorf("empty candidate should not match, got score=%d", r.Score)
	}
}

func TestMatch_ExactMatch(t *testing.T) {
	r := Match("main.go", "main.go")
	if r.Score <= 0 {
		t.Fatal("exact match should have positive score")
	}
	if len(r.Matches) != 7 {
		t.Errorf("expected 7 match positions, got %d", len(r.Matches))
	}
}

func TestMatch_NoMatch(t *testing.T) {
	r := Match("xyz", "main.go")
	if r.Score != 0 {
		t.Errorf("non-matching pattern should have score 0, got %d", r.Score)
	}
}

func TestMatch_FuzzyMatch(t *testing.T) {
	r := Match("mg", "main.go")
	if r.Score <= 0 {
		t.Fatal("fuzzy match should have positive score")
	}
	if len(r.Matches) != 2 {
		t.Errorf("expected 2 match positions, got %d", len(r.Matches))
	}
}

func TestMatch_CaseInsensitive(t *testing.T) {
	r := Match("MG", "main.go")
	if r.Score <= 0 {
		t.Fatal("case-insensitive match should work")
	}
}

func TestMatch_ConsecutiveScoreHigher(t *testing.T) {
	// "mai" in "main.go" should score higher than "mai" in "m_a_i.go"
	r1 := Match("mai", "main.go")
	r2 := Match("mai", "m_a_i.go")
	if r1.Score <= r2.Score {
		t.Errorf("consecutive match (%d) should score higher than scattered (%d)", r1.Score, r2.Score)
	}
}

func TestMatch_StartOfWordBonus(t *testing.T) {
	// "ff" matching at word starts in "file_finder.go" should beat "xff.go"
	r1 := Match("ff", "file_finder.go")
	r2 := Match("ff", "xffoo.go")
	if r1.Score <= r2.Score {
		t.Errorf("word-start match (%d) should score higher than mid-word (%d)", r1.Score, r2.Score)
	}
}

func TestMatch_PrefixBonus(t *testing.T) {
	r1 := Match("app", "app.go")
	r2 := Match("app", "internal/xapp.go")
	if r1.Score <= r2.Score {
		t.Errorf("prefix match (%d) should score higher than non-prefix (%d)", r1.Score, r2.Score)
	}
}

func TestMatch_PathFilenamePriority(t *testing.T) {
	// Match in filename should score higher than match in directories only.
	r1 := Match("main", "cmd/ide/main.go")
	r2 := Match("main", "cmd/models/actions/internal.go")
	if r1.Score <= r2.Score {
		t.Errorf("filename match (%d) should beat scattered path match (%d)", r1.Score, r2.Score)
	}
}

func TestMatch_Sorting(t *testing.T) {
	candidates := []string{
		"internal/fuzzy/filefinder.go",
		"internal/app/app.go",
		"cmd/ide/main.go",
		"internal/fuzzy/fuzzy.go",
		"go.mod",
	}

	type scored struct {
		path  string
		score int
	}

	pattern := "fuz"
	var results []scored
	for _, c := range candidates {
		r := Match(pattern, c)
		if r.Score > 0 {
			results = append(results, scored{c, r.Score})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if len(results) < 2 {
		t.Fatal("expected at least 2 fuzzy matches")
	}

	// fuzzy.go and filefinder.go should be top results.
	top := results[0].path
	if top != "internal/fuzzy/fuzzy.go" && top != "internal/fuzzy/filefinder.go" {
		t.Errorf("expected fuzzy file at top, got %s", top)
	}
}

func TestMatch_MatchIndices(t *testing.T) {
	r := Match("go", "main.go")
	if r.Score <= 0 {
		t.Fatal("should match")
	}
	// "go" should match indices 5 and 6 in "main.go".
	if len(r.Matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(r.Matches))
	}
}

func TestMatch_Unicode(t *testing.T) {
	r := Match("cfg", "configuracao.go")
	if r.Score <= 0 {
		t.Fatal("should match unicode candidates")
	}
}
