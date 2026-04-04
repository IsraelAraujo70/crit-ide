# Git Integration

## Requirements

- Current branch in statusline
- Modified files tracking
- Stage/unstage files
- Discard changes
- Commit
- Amend commit
- Checkout branch
- Create branch
- Pull/push
- Blame (future)
- Hunk-level actions
- Diff viewer

## Diff Viewer

Must support:

- Side-by-side view
- Inline diff view
- Hunk navigation
- Stage hunk
- Revert hunk
- Compare worktree vs index
- Compare index vs HEAD
- Compare branches/commits

## Internal Data Model

Don't treat Git as "just run a command". You need consistent derived state.

### RepoState

```go
type RepoState struct {
    RootPath        string
    Branch          string
    HeadSHA         string
    HasUncommitted  bool
    StagedFiles     []GitFileStatus
    UnstagedFiles   []GitFileStatus
    UntrackedFiles  []GitFileStatus
    Conflicts       []GitFileStatus
}

type GitFileStatus struct {
    Path        string
    X           string   // Index status
    Y           string   // Worktree status
    RenamedFrom string
}
```

### Diff Engine

Start by showing raw Git diff output. Then evolve to a structured hunk parser:

```go
type DiffFile struct {
    OldPath string
    NewPath string
    Hunks   []DiffHunk
}

type DiffHunk struct {
    Header string
    Lines  []DiffLine
}

type DiffLine struct {
    Kind    DiffLineKind  // Added, Removed, Context
    OldNo   int
    NewNo   int
    Text    string
}
```

This unlocks the side-by-side viewer and per-hunk actions.

### Useful Git Commands

| Command | Purpose |
|---------|---------|
| `git status --porcelain=v1 -b` | Status with branch info |
| `git diff -- ...` | Worktree vs index |
| `git diff --cached -- ...` | Index vs HEAD |
| `git show` | Commit details |
| `git log --oneline --decorate` | History |
| `git blame` | Line-level attribution |

## Components

- `GitService` — orchestrates git operations
- `RepoState` — derived state from git commands
- `DiffEngine` — parses unified diff into structured hunks
- `GitPanel` — UI panel for status/staging

## Implementation Strategy

Start by shelling out to the `git` binary with robust output parsing. Evaluate a Go library later if it makes sense.

## Planned Action IDs

- `git.status.open`
- `git.diff.open`
- `git.stage.hunk`
- `git.stage.file`
- `git.unstage.file`
- `git.commit`
- `git.checkout.branch`
