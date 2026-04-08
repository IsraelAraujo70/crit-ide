package editor

// GitStatusEntry represents a single file entry in the git status panel.
type GitStatusEntry struct {
	Path   string
	Status string // "M", "A", "D", "R", "U", "??"
	Staged bool
}

// GitStatusState holds the state of the git status panel.
type GitStatusState struct {
	Entries     []GitStatusEntry
	SelectedIdx int
	ScrollY     int
	Branch      string
}

// GitGraphLine represents a single line in the git graph display.
type GitGraphLine struct {
	Text    string
	Hash    string // Short hash if this line has a commit.
	Refs    string // Branch/tag decorations.
	IsGraph bool   // True if this is a graph line (not a commit).
}

// GitGraphState holds the state of the git graph panel.
type GitGraphState struct {
	Lines       []GitGraphLine
	SelectedIdx int
	ScrollY     int
}

// GitDiffState holds the state for viewing a diff.
type GitDiffState struct {
	Lines []string
	Path  string // File path being diffed.
	Title string // Display title.
	ScrollY int
}
