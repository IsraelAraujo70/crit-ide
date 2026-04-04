# Layout, Windows & UI

## Layout and Windows

### Requirements

- Horizontal splits
- Vertical splits
- Resize via mouse and keyboard
- Focus switching between panes
- Tabs per workspace
- Fixed and floating panels
- Lightweight popups/modals

### Initial Pane Types

- Editor pane
- File explorer
- Terminal pane
- Diagnostics pane
- Git pane
- Command palette
- Autocomplete popup
- Hover popup

### Layout Tree (Data Model)

```go
type NodeType int

const (
    NodeLeaf NodeType = iota
    NodeSplit
)

type LayoutNode struct {
    ID         string
    Type       NodeType
    Direction  SplitDirection
    Ratio      float64
    Pane       Pane
    Children   []*LayoutNode
}
```

### V1 Default Layout

On opening the IDE, the user should see:

- Explorer on the left
- Main editor in the center
- Statusline at the bottom
- Optional terminal at the bottom
- Command palette and autocomplete as popups

## Mouse Support

### Requirements

- Position cursor by clicking
- Select text by dragging
- Click to focus a pane
- Vertical and horizontal scroll
- Resize splits by dragging borders
- Click on diagnostics to navigate
- Click on autocomplete items
- Click on files in the explorer
- Double-click and drag support

Mouse must work well in modern terminals using appropriate escape sequences.

## UX Decisions

### Non-modal but Efficient

Replace "modes" with:

- Fast actions
- Chord sequences
- Command palette
- Smart selection
- Jump commands

**Examples:**

| Shortcut | Action |
|----------|--------|
| `Ctrl+D` | Select next occurrence |
| `Alt+Up/Down` | Move line up/down |
| `Ctrl+/` | Toggle line comment |
| `Ctrl+Click` | Go to definition |
| `Alt+Enter` | Open code actions |

### Discoverability

A powerful editor without discoverability becomes a locked niche.

**Must have:**

- Command palette
- Keybinding hints next to actions
- Searchable shortcut panel
- Contextual help in popups

### Statusline as Telemetry

The statusline is not decoration — it's the user's telemetry panel.

**Must display:**

- Current branch
- Error/warning count
- Dirty file indicator
- Encoding
- Cursor position
- Active LSP server
- AI model active/inactive
- Running task status

### Empty States

When no project is open, the UI should show something useful:

- Open folder
- Recent files
- Recent projects
- Most-used commands
