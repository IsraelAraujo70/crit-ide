# Navigation and Search

## Fuzzy Finder

### Requirements

- Open file by fuzzy match
- Switch buffer
- Go to symbol (document)
- Go to workspace symbol
- Recent files
- Command palette
- Search in project
- Grep results navigation

### Sources

The fuzzy finder aggregates multiple sources:

- Files in the project
- Open buffers
- LSP symbols
- Registered commands
- Git branches
- Commits
- Modified files

## Search and Replace

### Requirements

- Local search within the buffer
- Replace within the buffer
- Global search across the project
- Global replace with preview
- Optional regex support
- Optional case-sensitive mode
- Optional whole-word matching

### Implementation Strategy

Use an internal search engine with future integration with `ripgrep` when available on the system.

## Key Flows

### Open Project

1. User opens a directory
2. Workspace initializes
3. Git is detected
4. Project-local config is loaded
5. File tree indexes minimal structure
6. Language services become ready on demand

### Open Diff

1. User triggers `git.diff.open`
2. GitService fetches diff
3. DiffEngine processes hunks
4. DiffView opens in a dedicated pane
5. User navigates via mouse or keybinds

## Planned Action IDs

- `file.open`
- `buffer.close`
- `buffer.next`
- `search.project`
- `command.palette`
- `pane.split.horizontal`
- `pane.split.vertical`
- `pane.focus.left`
