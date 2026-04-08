// Package git provides Git repository operations for the IDE.
// It shells out to the git CLI and parses its output.
package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// FileStatusCode represents the status of a file in the working tree.
type FileStatusCode string

const (
	StatusModified  FileStatusCode = "M"
	StatusAdded     FileStatusCode = "A"
	StatusDeleted   FileStatusCode = "D"
	StatusRenamed   FileStatusCode = "R"
	StatusUnmerged  FileStatusCode = "U"
	StatusUntracked FileStatusCode = "??"
)

// FileStatus describes the status of a single file.
type FileStatus struct {
	Path   string
	Status FileStatusCode
	Staged bool
}

// CommitInfo holds parsed information about a single commit.
type CommitInfo struct {
	Hash    string
	Author  string
	Date    time.Time
	Message string
	Refs    string   // Decorated refs (e.g., "HEAD -> main, origin/main").
	Parents []string // Parent commit hashes.
}

// BranchInfo describes a Git branch.
type BranchInfo struct {
	Name    string
	Current bool
	Remote  string
}

// DiffLine represents a single line from a unified diff.
type DiffLine struct {
	Type    DiffLineType
	Content string
	OldNum  int // Line number in old file (0 if not applicable).
	NewNum  int // Line number in new file (0 if not applicable).
}

// DiffLineType categorizes a diff line.
type DiffLineType int

const (
	DiffContext  DiffLineType = iota // Unchanged context line.
	DiffAdded                       // Added line (+).
	DiffRemoved                     // Removed line (-).
	DiffHeader                      // Diff header / hunk header.
)

// LineDiffInfo contains per-line diff information for gutter indicators.
type LineDiffInfo struct {
	Line   int          // Zero-based line number in current file.
	Status DiffLineType // DiffAdded or DiffRemoved (marker) or DiffContext.
}

// Repo represents a Git repository rooted at a directory.
type Repo struct {
	root string
}

// NewRepo creates a Repo for the given directory. Returns nil if the
// directory is not inside a Git repository.
func NewRepo(dir string) *Repo {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil
	}
	out, err := runGit(absDir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil
	}
	root := strings.TrimSpace(out)
	if root == "" {
		return nil
	}
	return &Repo{root: root}
}

// Root returns the repository root directory.
func (r *Repo) Root() string {
	return r.root
}

// CurrentBranch returns the name of the current branch, or "HEAD" if detached.
func (r *Repo) CurrentBranch() string {
	out, err := r.run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return ""
	}
	branch := strings.TrimSpace(out)
	if branch == "" {
		return "HEAD"
	}
	return branch
}

// Status returns the list of files with changes (staged and unstaged).
func (r *Repo) Status() []FileStatus {
	out, err := r.run("status", "--porcelain=v1", "-uall")
	if err != nil {
		return nil
	}
	return parseStatus(out)
}

// Log returns the most recent commits up to the given limit.
func (r *Repo) Log(limit int) []CommitInfo {
	out, err := r.run("log",
		fmt.Sprintf("--max-count=%d", limit),
		"--format=%H%x00%an%x00%aI%x00%s%x00%D%x00%P",
		"--all",
	)
	if err != nil {
		return nil
	}
	return parseLog(out)
}

// GraphLog returns a visual git graph as raw lines.
func (r *Repo) GraphLog(limit int) []string {
	out, err := r.run("log",
		"--graph", "--oneline", "--all", "--decorate",
		fmt.Sprintf("--max-count=%d", limit),
	)
	if err != nil {
		return nil
	}
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

// Diff returns the unstaged diff for a specific file, or for all files if path is empty.
func (r *Repo) Diff(path string) string {
	args := []string{"diff", "--no-color"}
	if path != "" {
		args = append(args, "--", path)
	}
	out, err := r.run(args...)
	if err != nil {
		return ""
	}
	return out
}

// DiffStaged returns the staged diff for a specific file, or for all files if path is empty.
func (r *Repo) DiffStaged(path string) string {
	args := []string{"diff", "--cached", "--no-color"}
	if path != "" {
		args = append(args, "--", path)
	}
	out, err := r.run(args...)
	if err != nil {
		return ""
	}
	return out
}

// DiffForGutter returns per-line diff information for the given file,
// suitable for rendering gutter indicators.
func (r *Repo) DiffForGutter(path string) []LineDiffInfo {
	// Use diff against HEAD for combined staged+unstaged changes.
	out, err := r.run("diff", "HEAD", "--no-color", "--unified=0", "--", path)
	if err != nil {
		return nil
	}
	return parseGutterDiff(out)
}

// Stage adds a file to the staging area.
func (r *Repo) Stage(path string) error {
	_, err := r.run("add", "--", path)
	return err
}

// Unstage removes a file from the staging area.
func (r *Repo) Unstage(path string) error {
	_, err := r.run("reset", "HEAD", "--", path)
	return err
}

// Commit creates a new commit with the given message.
func (r *Repo) Commit(msg string) error {
	_, err := r.run("commit", "-m", msg)
	return err
}

// BranchList returns all local branches.
func (r *Repo) BranchList() []BranchInfo {
	out, err := r.run("branch", "--format=%(HEAD)%(refname:short)%00%(upstream:short)")
	if err != nil {
		return nil
	}
	return parseBranches(out)
}

// run executes a git command in the repo root and returns stdout.
func (r *Repo) run(args ...string) (string, error) {
	return runGit(r.root, args...)
}

// runGit executes a git command in the given directory.
func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.String(), nil
}

// --- Parsing functions ---

// parseStatus parses `git status --porcelain=v1` output.
func parseStatus(output string) []FileStatus {
	var result []FileStatus
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 4 {
			continue
		}
		x := line[0] // Staged status.
		y := line[1] // Unstaged status.
		path := line[3:]

		// Handle rename: "R  old -> new"
		if idx := strings.Index(path, " -> "); idx >= 0 {
			path = path[idx+4:]
		}

		if x == '?' && y == '?' {
			result = append(result, FileStatus{
				Path:   path,
				Status: StatusUntracked,
				Staged: false,
			})
			continue
		}

		// Staged change.
		if x != ' ' && x != '?' {
			result = append(result, FileStatus{
				Path:   path,
				Status: charToStatus(x),
				Staged: true,
			})
		}

		// Unstaged change.
		if y != ' ' && y != '?' {
			result = append(result, FileStatus{
				Path:   path,
				Status: charToStatus(y),
				Staged: false,
			})
		}
	}
	return result
}

func charToStatus(c byte) FileStatusCode {
	switch c {
	case 'M':
		return StatusModified
	case 'A':
		return StatusAdded
	case 'D':
		return StatusDeleted
	case 'R':
		return StatusRenamed
	case 'U':
		return StatusUnmerged
	default:
		return StatusModified
	}
}

// parseLog parses the custom formatted git log output.
func parseLog(output string) []CommitInfo {
	var commits []CommitInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\x00", 6)
		if len(parts) < 6 {
			continue
		}
		date, _ := time.Parse(time.RFC3339, parts[2])
		var parents []string
		if parts[5] != "" {
			parents = strings.Split(parts[5], " ")
		}
		commits = append(commits, CommitInfo{
			Hash:    parts[0],
			Author:  parts[1],
			Date:    date,
			Message: parts[3],
			Refs:    parts[4],
			Parents: parents,
		})
	}
	return commits
}

// parseBranches parses the custom formatted git branch output.
func parseBranches(output string) []BranchInfo {
	var branches []BranchInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 2 {
			continue
		}
		current := line[0] == '*'
		rest := line[1:]
		parts := strings.SplitN(rest, "\x00", 2)
		name := parts[0]
		remote := ""
		if len(parts) > 1 {
			remote = parts[1]
		}
		branches = append(branches, BranchInfo{
			Name:    name,
			Current: current,
			Remote:  remote,
		})
	}
	return branches
}

// parseGutterDiff parses unified diff output (with --unified=0) and extracts
// per-line change information for gutter indicators.
func parseGutterDiff(output string) []LineDiffInfo {
	var result []LineDiffInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "@@") {
			continue
		}
		// Parse hunk header: @@ -oldStart,oldCount +newStart,newCount @@
		newStart, newCount, oldCount := parseHunkHeader(line)
		if newCount > 0 {
			diffType := DiffAdded
			if oldCount > 0 {
				diffType = DiffContext // Modified (both added and removed).
			}
			for i := 0; i < newCount; i++ {
				result = append(result, LineDiffInfo{
					Line:   newStart + i - 1, // Convert to zero-based.
					Status: diffType,
				})
			}
		}
		if oldCount > 0 && newCount == 0 {
			// Pure deletion — mark the line after the deletion point.
			result = append(result, LineDiffInfo{
				Line:   newStart - 1, // Zero-based; the line where deletion happened.
				Status: DiffRemoved,
			})
		}
	}
	return result
}

// parseHunkHeader extracts newStart, newCount, and oldCount from a hunk header line.
func parseHunkHeader(line string) (newStart, newCount, oldCount int) {
	// Format: @@ -oldStart[,oldCount] +newStart[,newCount] @@
	parts := strings.SplitN(line, "@@", 3)
	if len(parts) < 2 {
		return 0, 0, 0
	}
	ranges := strings.TrimSpace(parts[1])
	rangeParts := strings.Split(ranges, " ")
	if len(rangeParts) < 2 {
		return 0, 0, 0
	}

	// Parse old range.
	oldRange := strings.TrimPrefix(rangeParts[0], "-")
	oldParts := strings.Split(oldRange, ",")
	if len(oldParts) >= 2 {
		oldCount, _ = strconv.Atoi(oldParts[1])
	} else {
		oldCount = 1
	}

	// Parse new range.
	newRange := strings.TrimPrefix(rangeParts[1], "+")
	newParts := strings.Split(newRange, ",")
	newStart, _ = strconv.Atoi(newParts[0])
	if len(newParts) >= 2 {
		newCount, _ = strconv.Atoi(newParts[1])
	} else {
		newCount = 1
	}

	return newStart, newCount, oldCount
}
