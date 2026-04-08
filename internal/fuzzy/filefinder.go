package fuzzy

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// maxResults is the maximum number of results returned by the file finder.
const maxResults = 20

// FileFinder caches a list of project files and performs fuzzy filtering.
type FileFinder struct {
	rootPath string
	files    []string // Relative paths from rootPath.
	mu       sync.RWMutex
}

// NewFileFinder creates a FileFinder and immediately scans the project directory.
func NewFileFinder(rootPath string) *FileFinder {
	ff := &FileFinder{rootPath: rootPath}
	ff.Rebuild()
	return ff
}

// Rebuild rescans the project directory and rebuilds the file cache.
// Safe to call from any goroutine.
func (ff *FileFinder) Rebuild() {
	files := walkFiles(ff.rootPath)
	ff.mu.Lock()
	ff.files = files
	ff.mu.Unlock()
}

// Filter returns the top matching files for the given pattern.
// Results are sorted by match score (best first), limited to maxResults.
func (ff *FileFinder) Filter(pattern string) []FileResult {
	ff.mu.RLock()
	files := ff.files
	ff.mu.RUnlock()

	if pattern == "" {
		// Return first N files when no pattern.
		n := len(files)
		if n > maxResults {
			n = maxResults
		}
		results := make([]FileResult, n)
		for i := 0; i < n; i++ {
			results[i] = FileResult{
				RelPath: files[i],
				AbsPath: filepath.Join(ff.rootPath, files[i]),
				Score:   1,
			}
		}
		return results
	}

	type scored struct {
		idx   int
		match MatchResult
	}

	var matches []scored
	for i, f := range files {
		r := Match(pattern, f)
		if r.Score > 0 {
			matches = append(matches, scored{i, r})
		}
	}

	// Sort by score descending.
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].match.Score > matches[j].match.Score
	})

	n := len(matches)
	if n > maxResults {
		n = maxResults
	}

	results := make([]FileResult, n)
	for i := 0; i < n; i++ {
		m := matches[i]
		results[i] = FileResult{
			RelPath:  files[m.idx],
			AbsPath:  filepath.Join(ff.rootPath, files[m.idx]),
			Score:    m.match.Score,
			Matches:  m.match.Matches,
		}
	}
	return results
}

// FileCount returns the total number of cached files.
func (ff *FileFinder) FileCount() int {
	ff.mu.RLock()
	defer ff.mu.RUnlock()
	return len(ff.files)
}

// FileResult holds a single file finder result.
type FileResult struct {
	RelPath string // Path relative to project root.
	AbsPath string // Absolute filesystem path.
	Score   int    // Fuzzy match score.
	Matches []int  // Character indices that matched in RelPath.
}

// walkFiles collects all non-hidden files under root, respecting common ignore patterns.
func walkFiles(root string) []string {
	var files []string

	// Read .gitignore patterns (simple support).
	ignorePatterns := readGitignore(root)

	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip unreadable entries.
		}

		name := d.Name()

		// Skip hidden directories and files.
		if strings.HasPrefix(name, ".") && path != root {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip common noise directories.
		if d.IsDir() {
			switch name {
			case "node_modules", "vendor", "__pycache__", "dist", "build",
				"target", ".git", ".svn", ".hg":
				return filepath.SkipDir
			}
		}

		if d.IsDir() {
			return nil
		}

		// Get relative path.
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		// Check gitignore patterns.
		if matchesGitignore(rel, ignorePatterns) {
			return nil
		}

		// Skip binary-like files.
		ext := strings.ToLower(filepath.Ext(name))
		switch ext {
		case ".exe", ".bin", ".so", ".dylib", ".dll", ".o", ".a",
			".png", ".jpg", ".jpeg", ".gif", ".ico", ".bmp",
			".zip", ".tar", ".gz", ".bz2", ".xz", ".7z",
			".pdf", ".wasm":
			return nil
		}

		files = append(files, rel)
		return nil
	})

	// Sort alphabetically for consistent results.
	sort.Strings(files)
	return files
}

// readGitignore reads simple .gitignore patterns from the project root.
func readGitignore(root string) []string {
	data, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		return nil
	}

	var patterns []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}

// matchesGitignore checks if a relative path matches any gitignore pattern.
// Simple implementation: supports directory patterns (ending with /)
// and basic glob-style patterns.
func matchesGitignore(relPath string, patterns []string) bool {
	for _, p := range patterns {
		// Directory pattern.
		if strings.HasSuffix(p, "/") {
			dir := strings.TrimSuffix(p, "/")
			if strings.HasPrefix(relPath, dir+"/") || relPath == dir {
				return true
			}
			// Check if any path component matches.
			parts := strings.Split(relPath, "/")
			for _, part := range parts {
				if part == dir {
					return true
				}
			}
			continue
		}

		// Glob pattern.
		if strings.ContainsAny(p, "*?") {
			matched, _ := filepath.Match(p, filepath.Base(relPath))
			if matched {
				return true
			}
			continue
		}

		// Exact name match (in any directory).
		base := filepath.Base(relPath)
		if base == p {
			return true
		}

		// Path prefix match.
		if strings.HasPrefix(relPath, p+"/") || relPath == p {
			return true
		}
	}
	return false
}
