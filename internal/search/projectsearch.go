package search

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// SearchResult represents a single match from the project-wide search.
type SearchResult struct {
	Path        string // Absolute file path.
	Line        int    // 1-based line number.
	Col         int    // 1-based column number.
	MatchText   string // The full text of the matching line.
	ContextBefore []string // Lines before the match (for future context display).
	ContextAfter  []string // Lines after the match (for future context display).
}

// FileGroup groups search results by file path for display.
type FileGroup struct {
	Path    string         // Absolute file path.
	RelPath string         // Relative path for display.
	Results []SearchResult // Matches in this file.
}

// maxResults limits the total number of results to prevent UI overload.
const maxResults = 500

// Search performs a project-wide text search using ripgrep (rg) if available,
// falling back to grep -rn. Returns results grouped by file.
func Search(query, rootPath string) ([]FileGroup, int) {
	if query == "" {
		return nil, 0
	}

	results := searchWithRg(query, rootPath)
	if results == nil {
		results = searchWithGrep(query, rootPath)
	}

	return groupByFile(results, rootPath), len(results)
}

// SearchRegex performs a project-wide regex search.
func SearchRegex(pattern, rootPath string) ([]FileGroup, int) {
	if pattern == "" {
		return nil, 0
	}

	// Validate regex.
	if _, err := regexp.Compile(pattern); err != nil {
		return nil, 0
	}

	results := searchRegexWithRg(pattern, rootPath)
	if results == nil {
		results = searchRegexWithGrep(pattern, rootPath)
	}

	return groupByFile(results, rootPath), len(results)
}

// searchWithRg uses ripgrep for fast searching.
func searchWithRg(query, rootPath string) []SearchResult {
	rgPath, err := exec.LookPath("rg")
	if err != nil {
		return nil // rg not found, signal fallback.
	}

	cmd := exec.Command(rgPath, "--no-heading", "--line-number", "--column",
		"--color=never", "--max-count=100", "--fixed-strings",
		"--max-filesize=1M", "-g", "!.git",
		query, rootPath)

	return runSearchCommand(cmd, rootPath)
}

// searchRegexWithRg uses ripgrep for regex searching.
func searchRegexWithRg(pattern, rootPath string) []SearchResult {
	rgPath, err := exec.LookPath("rg")
	if err != nil {
		return nil
	}

	cmd := exec.Command(rgPath, "--no-heading", "--line-number", "--column",
		"--color=never", "--max-count=100",
		"--max-filesize=1M", "-g", "!.git",
		pattern, rootPath)

	return runSearchCommand(cmd, rootPath)
}

// searchWithGrep falls back to grep -rn.
func searchWithGrep(query, rootPath string) []SearchResult {
	grepPath, err := exec.LookPath("grep")
	if err != nil {
		return nil
	}

	cmd := exec.Command(grepPath, "-rn", "--include=*",
		"--exclude-dir=.git", "--exclude-dir=node_modules",
		"--exclude-dir=vendor", "--exclude-dir=__pycache__",
		"-F", query, rootPath)

	return runGrepCommand(cmd, rootPath)
}

// searchRegexWithGrep uses grep with regex.
func searchRegexWithGrep(pattern, rootPath string) []SearchResult {
	grepPath, err := exec.LookPath("grep")
	if err != nil {
		return nil
	}

	cmd := exec.Command(grepPath, "-rn", "--include=*",
		"--exclude-dir=.git", "--exclude-dir=node_modules",
		"--exclude-dir=vendor", "--exclude-dir=__pycache__",
		"-E", pattern, rootPath)

	return runGrepCommand(cmd, rootPath)
}

// runSearchCommand runs rg and parses output in the format: path:line:col:text
func runSearchCommand(cmd *exec.Cmd, rootPath string) []SearchResult {
	out, err := cmd.Output()
	if err != nil {
		// Exit code 1 means no matches (not an error).
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []SearchResult{}
		}
		return []SearchResult{}
	}

	var results []SearchResult
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() && len(results) < maxResults {
		line := scanner.Text()
		parts := splitRgLine(line)
		if parts == nil {
			continue
		}

		lineNum, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		colNum, err := strconv.Atoi(parts[2])
		if err != nil {
			colNum = 1
		}

		results = append(results, SearchResult{
			Path:      parts[0],
			Line:      lineNum,
			Col:       colNum,
			MatchText: parts[3],
		})
	}

	return results
}

// runGrepCommand runs grep and parses output in the format: path:line:text
func runGrepCommand(cmd *exec.Cmd, rootPath string) []SearchResult {
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []SearchResult{}
		}
		return []SearchResult{}
	}

	var results []SearchResult
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() && len(results) < maxResults {
		line := scanner.Text()
		parts := splitGrepLine(line)
		if parts == nil {
			continue
		}

		lineNum, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		results = append(results, SearchResult{
			Path:      parts[0],
			Line:      lineNum,
			Col:       1,
			MatchText: parts[2],
		})
	}

	return results
}

// splitRgLine splits a ripgrep output line (path:line:col:text).
func splitRgLine(line string) []string {
	// Find first colon (end of path).
	idx1 := strings.Index(line, ":")
	if idx1 < 0 {
		return nil
	}
	// Find second colon (end of line number).
	idx2 := strings.Index(line[idx1+1:], ":")
	if idx2 < 0 {
		return nil
	}
	idx2 += idx1 + 1
	// Find third colon (end of column number).
	idx3 := strings.Index(line[idx2+1:], ":")
	if idx3 < 0 {
		return nil
	}
	idx3 += idx2 + 1

	return []string{
		line[:idx1],
		line[idx1+1 : idx2],
		line[idx2+1 : idx3],
		line[idx3+1:],
	}
}

// splitGrepLine splits a grep output line (path:line:text).
func splitGrepLine(line string) []string {
	idx1 := strings.Index(line, ":")
	if idx1 < 0 {
		return nil
	}
	idx2 := strings.Index(line[idx1+1:], ":")
	if idx2 < 0 {
		return nil
	}
	idx2 += idx1 + 1

	return []string{
		line[:idx1],
		line[idx1+1 : idx2],
		line[idx2+1:],
	}
}

// groupByFile groups flat results into FileGroup slices, ordered by first appearance.
func groupByFile(results []SearchResult, rootPath string) []FileGroup {
	if len(results) == 0 {
		return nil
	}

	groupMap := make(map[string]int) // path -> index in groups
	var groups []FileGroup

	for _, r := range results {
		idx, ok := groupMap[r.Path]
		if !ok {
			relPath, err := filepath.Rel(rootPath, r.Path)
			if err != nil {
				relPath = r.Path
			}
			idx = len(groups)
			groupMap[r.Path] = idx
			groups = append(groups, FileGroup{
				Path:    r.Path,
				RelPath: relPath,
			})
		}
		groups[idx].Results = append(groups[idx].Results, r)
	}

	return groups
}

// FlattenResults creates a flat list of display entries (headers + results).
type DisplayEntry struct {
	IsHeader  bool         // True = file header, False = result line.
	GroupIdx  int          // Index into the groups slice.
	ResultIdx int          // Index within the group's Results (-1 for headers).
	Text      string       // Display text.
	Path      string       // Absolute path (for navigation).
	Line      int          // Line number (for navigation).
	Col       int          // Column number (for navigation).
}

// Flatten converts FileGroups into a flat list for rendering.
func Flatten(groups []FileGroup) []DisplayEntry {
	var entries []DisplayEntry
	for gi, g := range groups {
		// File header.
		entries = append(entries, DisplayEntry{
			IsHeader:  true,
			GroupIdx:  gi,
			ResultIdx: -1,
			Text:      fmt.Sprintf("  %s (%d)", g.RelPath, len(g.Results)),
			Path:      g.Path,
		})

		// Results.
		for ri, r := range g.Results {
			text := fmt.Sprintf("    %d: %s", r.Line, strings.TrimSpace(r.MatchText))
			entries = append(entries, DisplayEntry{
				IsHeader:  false,
				GroupIdx:  gi,
				ResultIdx: ri,
				Text:      text,
				Path:      r.Path,
				Line:      r.Line,
				Col:       r.Col,
			})
		}
	}
	return entries
}
