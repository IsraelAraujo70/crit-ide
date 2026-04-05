package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/israelcorrea/crit-ide/internal/actions"
	"github.com/israelcorrea/crit-ide/internal/clipboard"
	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/events"
	"github.com/israelcorrea/crit-ide/internal/filetree"
	"github.com/israelcorrea/crit-ide/internal/input"
	"github.com/israelcorrea/crit-ide/internal/render"
)

const defaultTreeWidth = 30

// App is the top-level application state and event loop.
type App struct {
	screen        tcell.Screen
	bus           *events.Bus
	registry      *actions.Registry
	renderer      *render.Renderer
	filePath      string
	quit          bool
	clip          actions.ClipboardProvider
	inputMode     actions.InputMode
	contextMenu   *editor.MenuState
	pendingAction string

	// Multi-buffer / tabs.
	buffers      []*editor.Buffer
	activeIdx    int
	scrollYs     map[editor.BufferID]int // Per-buffer scroll position.
	nextBufferID int                     // Monotonic counter for unique buffer IDs.

	// File tree.
	tree        *filetree.FileTree
	treeVisible bool
	treeWidth   int

	// Focus.
	focusArea actions.FocusArea

	// Input prompt.
	prompt *editor.PromptState
}

// New creates a new App. If filePath is non-empty, that file will be opened.
func New(filePath string) *App {
	return &App{
		filePath:  filePath,
		bus:       events.NewBus(256),
		registry:  actions.NewRegistry(),
		scrollYs:  make(map[editor.BufferID]int),
		treeWidth: defaultTreeWidth,
	}
}

// Run initializes the terminal, registers actions, and enters the main event loop.
func (a *App) Run() error {
	// Initialize tcell screen.
	screen, err := tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("failed to create screen: %w", err)
	}
	if err := screen.Init(); err != nil {
		return fmt.Errorf("failed to init screen: %w", err)
	}
	a.screen = screen
	a.screen.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorDefault))
	a.screen.EnableMouse()
	a.screen.Clear()

	// Cleanup on exit.
	defer a.screen.Fini()

	// Create renderer.
	a.renderer = render.NewRenderer(a.screen)

	// Initialize clipboard.
	a.clip = &clipboard.SystemClipboard{}

	// Register all actions.
	actions.RegisterAll(a.registry)

	// Load file or create scratch buffer.
	if a.filePath != "" {
		buf, err := editor.LoadFile("main", a.filePath)
		if err != nil {
			buf = editor.NewBuffer("main")
			buf.Path = a.filePath
			buf.Kind = editor.BufferKindFile
		}
		a.buffers = append(a.buffers, buf)
	} else {
		a.buffers = append(a.buffers, editor.NewBuffer("scratch"))
	}

	// Initialize file tree from current working directory or file's directory.
	a.initFileTree()

	// Launch input goroutine.
	inputHandler := input.NewHandler(a.screen, a.bus)
	go inputHandler.Run()

	// Initial render.
	a.render()

	// Main event loop — the only place state is mutated.
	for !a.quit {
		ev := <-a.bus.Recv()

		switch ev.Type {
		case events.EventAction:
			ctx := &actions.ActionContext{
				App:   a,
				Event: &ev,
			}

			switch a.inputMode {
			case actions.ModeNormal:
				a.handleNormalAction(ev.ActionID, ctx)
			case actions.ModeContextMenu:
				a.handleContextMenuAction(ev.ActionID, ctx)
			case actions.ModePrompt:
				a.handlePromptAction(ev.ActionID, ctx)
			}

			// Execute any pending action (from menu item execution).
			for a.pendingAction != "" {
				pending := a.pendingAction
				a.pendingAction = ""
				pctx := &actions.ActionContext{
					App:   a,
					Event: &ev,
				}
				_ = a.registry.Execute(pending, pctx)
			}

			a.ensureCursorVisible()

		case events.EventResize:
			a.screen.Sync()
			a.ensureCursorVisible()

		case events.EventQuit:
			a.quit = true
			continue
		}

		a.render()
	}

	return nil
}

// initFileTree initializes the file tree from the project root.
func (a *App) initFileTree() {
	rootPath := "."
	if a.filePath != "" {
		absPath, err := filepath.Abs(a.filePath)
		if err == nil {
			rootPath = filepath.Dir(absPath)
		}
	} else {
		cwd, err := os.Getwd()
		if err == nil {
			rootPath = cwd
		}
	}

	tree, err := filetree.New(rootPath)
	if err == nil {
		a.tree = tree
		a.treeVisible = true
	}
}

// handleNormalAction routes actions based on focus area.
func (a *App) handleNormalAction(actionID string, ctx *actions.ActionContext) {
	// Global actions work regardless of focus.
	switch actionID {
	case "tree.toggle", "tab.next", "tab.prev", "tab.close", "app.quit",
		"file.save", "tree.refresh":
		_ = a.registry.Execute(actionID, ctx)
		return
	}

	// Mouse click routing — determine focus based on click position.
	if actionID == "mouse.click" {
		a.routeMouseClick(ctx)
		return
	}

	// Mouse scroll routing.
	if actionID == "mouse.scroll" {
		a.routeMouseScroll(ctx)
		return
	}

	// Mouse drag routing.
	if actionID == "mouse.drag" {
		// Drags only make sense in the editor area.
		if a.focusArea == actions.FocusEditor {
			_ = a.registry.Execute(actionID, ctx)
		}
		return
	}

	// When file tree has focus, remap cursor keys to tree navigation.
	if a.focusArea == actions.FocusFileTree {
		remap := map[string]string{
			"cursor.up":       "tree.up",
			"cursor.down":     "tree.down",
			"cursor.right":    "tree.expand",
			"cursor.left":     "tree.collapse",
			"insert.newline":  "tree.enter",
			"input.escape":    "tree.focus.editor",
			"delete.backward": "tree.delete",
			"delete.forward":  "tree.delete",
		}
		if mapped, ok := remap[actionID]; ok {
			_ = a.registry.Execute(mapped, ctx)
			return
		}

		// Shortcut keys when tree is focused: a=new, d=delete, r=rename.
		if actionID == "insert.char" {
			if ch, ok := ctx.Event.Payload.(rune); ok {
				switch ch {
				case 'a':
					_ = a.registry.Execute("tree.new", ctx)
					return
				case 'd':
					_ = a.registry.Execute("tree.delete", ctx)
					return
				case 'r':
					_ = a.registry.Execute("tree.rename", ctx)
					return
				}
			}
		}

		// Ignore other actions when tree is focused.
		return
	}

	// Default: execute in editor context.
	_ = a.registry.Execute(actionID, ctx)
}

// routeMouseClick determines which panel was clicked and dispatches accordingly.
func (a *App) routeMouseClick(ctx *actions.ActionContext) {
	payload, ok := ctx.Event.Payload.(events.MouseClickPayload)
	if !ok {
		return
	}

	w, _ := a.screen.Size()

	// Tab bar click (row 0).
	if payload.ScreenY == 0 {
		_ = a.registry.Execute("tab.click", ctx)
		return
	}

	// File tree area click (right side).
	treeStartX := a.editorWidth(w)
	if a.treeVisible && payload.ScreenX >= treeStartX {
		_ = a.registry.Execute("tree.click", ctx)
		return
	}

	// Editor area click.
	a.focusArea = actions.FocusEditor
	_ = a.registry.Execute("mouse.click", ctx)
}

// routeMouseScroll routes scroll events to the correct panel.
func (a *App) routeMouseScroll(ctx *actions.ActionContext) {
	payload, ok := ctx.Event.Payload.(events.MouseScrollPayload)
	if !ok {
		return
	}

	w, _ := a.screen.Size()
	treeStartX := a.editorWidth(w)

	if a.treeVisible && payload.ScreenX >= treeStartX {
		// Scroll inside the tree — adjust tree scroll.
		if a.tree != nil {
			a.tree.ScrollY += payload.Direction
			if a.tree.ScrollY < 0 {
				a.tree.ScrollY = 0
			}
			maxScroll := a.tree.VisibleCount() - a.treeViewportHeight()
			if maxScroll < 0 {
				maxScroll = 0
			}
			if a.tree.ScrollY > maxScroll {
				a.tree.ScrollY = maxScroll
			}
		}
		return
	}

	// Scroll in the editor.
	_ = a.registry.Execute("mouse.scroll", ctx)
}

// editorWidth returns the editor area width (total width minus tree width if visible).
func (a *App) editorWidth(totalWidth int) int {
	if a.treeVisible {
		w := totalWidth - a.treeWidth
		if w < 10 {
			w = 10
		}
		return w
	}
	return totalWidth
}

// treeViewportHeight returns the height available for tree content (minus tab bar, header, statusline).
func (a *App) treeViewportHeight() int {
	_, h := a.screen.Size()
	tvh := h - 3 // tab bar (1) + header (1) + statusline (1)
	if tvh < 1 {
		tvh = 1
	}
	return tvh
}

// render draws the current state to the screen.
func (a *App) render() {
	w, h := a.screen.Size()
	buf := a.ActiveBuffer()

	vs := &render.ViewState{
		Buffer:       buf,
		ScrollY:      a.scrollY(),
		Width:        w,
		Height:       h,
		ActiveTabIdx: a.activeIdx,
		TreeVisible:  a.treeVisible,
		TreeWidth:    a.treeWidth,
		TreeFocused:  a.focusArea == actions.FocusFileTree,
	}

	if a.contextMenu != nil {
		vs.Popup = a.contextMenu
	}
	if a.prompt != nil {
		vs.Prompt = a.prompt
	}

	// Build tab info.
	for i, b := range a.buffers {
		vs.Tabs = append(vs.Tabs, render.TabInfo{
			Name:   b.FileName(),
			Dirty:  b.Dirty,
			Active: i == a.activeIdx,
		})
	}

	// Build tree node info.
	if a.tree != nil && a.treeVisible {
		vs.TreeScrollY = a.tree.ScrollY
		vs.TreeCursor = a.tree.Cursor
		for _, node := range a.tree.Visible {
			vs.TreeNodes = append(vs.TreeNodes, render.TreeNode{
				Name:     node.Name,
				IsDir:    node.IsDir,
				Expanded: node.Expanded,
				Depth:    node.Depth,
				Path:     node.Path,
			})
		}
	}

	a.renderer.Render(vs)
}

// ensureCursorVisible adjusts scrollY so the cursor is within the viewport.
func (a *App) ensureCursorVisible() {
	_, h := a.screen.Size()
	tabBarHeight := 1
	editorHeight := h - tabBarHeight - 1 // Statusline takes 1 row.
	if editorHeight < 1 {
		editorHeight = 1
	}

	buf := a.ActiveBuffer()
	sy := a.scrollY()

	if buf.CursorRow < sy {
		a.setScrollY(buf.CursorRow)
	}
	if buf.CursorRow >= sy+editorHeight {
		a.setScrollY(buf.CursorRow - editorHeight + 1)
	}

	// Also ensure tree cursor is visible.
	if a.tree != nil && a.treeVisible {
		tvh := a.treeViewportHeight()
		a.tree.EnsureCursorVisible(tvh)
	}
}

// handleContextMenuAction routes actions while the context menu is open.
func (a *App) handleContextMenuAction(actionID string, ctx *actions.ActionContext) {
	remap := map[string]string{
		"cursor.up":      "menu.up",
		"cursor.down":    "menu.down",
		"insert.newline": "menu.execute",
		"input.escape":   "menu.close",
	}
	if mapped, ok := remap[actionID]; ok {
		actionID = mapped
	}

	switch {
	case strings.HasPrefix(actionID, "menu."):
		_ = a.registry.Execute(actionID, ctx)
	case actionID == "mouse.click":
		ctx.Event.ActionID = "menu.click"
		_ = a.registry.Execute("menu.click", ctx)
	default:
		// Ignore all other actions while menu is open.
	}
}

// handlePromptAction routes actions while the input prompt is active.
func (a *App) handlePromptAction(actionID string, ctx *actions.ActionContext) {
	remap := map[string]string{
		"insert.char":     "prompt.char",
		"delete.backward": "prompt.backspace",
		"delete.forward":  "prompt.delete",
		"cursor.left":     "prompt.left",
		"cursor.right":    "prompt.right",
		"cursor.home":     "prompt.home",
		"cursor.end":      "prompt.end",
		"input.escape":    "prompt.cancel",
		"insert.newline":  "prompt.confirm",
	}
	if mapped, ok := remap[actionID]; ok {
		_ = a.registry.Execute(mapped, ctx)
		return
	}
	// Ignore all other actions while prompt is open.
}

// scrollY returns the scroll offset for the active buffer.
func (a *App) scrollY() int {
	buf := a.ActiveBuffer()
	return a.scrollYs[buf.ID]
}

// setScrollY sets the scroll offset for the active buffer.
func (a *App) setScrollY(y int) {
	buf := a.ActiveBuffer()
	a.scrollYs[buf.ID] = y
}

// --- AppState interface implementation ---

// ActiveBuffer returns the currently focused buffer.
// Panics if buffers slice is empty (invariant: at least one buffer must exist).
func (a *App) ActiveBuffer() *editor.Buffer {
	if len(a.buffers) == 0 {
		panic("crit-ide: no buffers open (invariant violated)")
	}
	if a.activeIdx >= 0 && a.activeIdx < len(a.buffers) {
		return a.buffers[a.activeIdx]
	}
	return a.buffers[0]
}

// ScrollY returns the current vertical scroll offset.
func (a *App) ScrollY() int {
	return a.scrollY()
}

// SetScrollY sets the vertical scroll offset.
func (a *App) SetScrollY(y int) {
	a.setScrollY(y)
}

// ViewportHeight returns the number of visible editor rows (excluding tab bar and statusline).
func (a *App) ViewportHeight() int {
	_, h := a.screen.Size()
	tabBarHeight := 1
	eh := h - tabBarHeight - 1 // statusline
	if eh < 1 {
		return 1
	}
	return eh
}

// ScreenWidth returns the terminal width in columns.
func (a *App) ScreenWidth() int {
	w, _ := a.screen.Size()
	return w
}

// Quit signals the application to exit.
func (a *App) Quit() {
	a.quit = true
}

// Clipboard returns the clipboard provider.
func (a *App) Clipboard() actions.ClipboardProvider {
	return a.clip
}

// InputMode returns the current input routing mode.
func (a *App) InputMode() actions.InputMode {
	return a.inputMode
}

// SetInputMode sets the input routing mode.
func (a *App) SetInputMode(mode actions.InputMode) {
	a.inputMode = mode
}

// ContextMenu returns the active context menu state, or nil.
func (a *App) ContextMenu() *editor.MenuState {
	return a.contextMenu
}

// SetContextMenu sets or clears the context menu state.
func (a *App) SetContextMenu(menu *editor.MenuState) {
	a.contextMenu = menu
}

// PostAction queues an action to execute after the current action completes.
func (a *App) PostAction(actionID string) {
	a.pendingAction = actionID
}

// --- Tab / multi-buffer interface ---

// Buffers returns all open buffers.
func (a *App) Buffers() []*editor.Buffer {
	return a.buffers
}

// ActiveBufferIndex returns the index of the active buffer.
func (a *App) ActiveBufferIndex() int {
	return a.activeIdx
}

// OpenFile opens a file in a new tab, or switches to the existing tab.
func (a *App) OpenFile(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Check if already open.
	for i, buf := range a.buffers {
		if buf.Path == absPath {
			a.activeIdx = i
			return nil
		}
	}

	// Generate a unique buffer ID using a monotonic counter.
	a.nextBufferID++
	id := editor.BufferID(fmt.Sprintf("buf-%d", a.nextBufferID))
	buf, err := editor.LoadFile(id, absPath)
	if err != nil {
		return err
	}

	a.buffers = append(a.buffers, buf)
	a.activeIdx = len(a.buffers) - 1
	return nil
}

// CloseBuffer closes the buffer at the given index.
func (a *App) CloseBuffer(idx int) {
	if idx < 0 || idx >= len(a.buffers) || len(a.buffers) <= 1 {
		return
	}

	// Remove scroll state for this buffer.
	delete(a.scrollYs, a.buffers[idx].ID)

	// Remove from slice.
	a.buffers = append(a.buffers[:idx], a.buffers[idx+1:]...)

	// Adjust active index.
	if a.activeIdx >= len(a.buffers) {
		a.activeIdx = len(a.buffers) - 1
	}
	if a.activeIdx < 0 {
		a.activeIdx = 0
	}
}

// SwitchBuffer switches to the buffer at the given index.
func (a *App) SwitchBuffer(idx int) {
	if idx >= 0 && idx < len(a.buffers) {
		a.activeIdx = idx
	}
}

// --- File tree interface ---

// FileTree returns the file tree state as a FileTreeState interface.
func (a *App) FileTree() actions.FileTreeState {
	if a.tree == nil {
		return nil
	}
	return a.tree
}

// FileTreeVisible returns whether the file tree is visible.
func (a *App) FileTreeVisible() bool {
	return a.treeVisible
}

// SetFileTreeVisible sets file tree visibility.
func (a *App) SetFileTreeVisible(v bool) {
	a.treeVisible = v
}

// ToggleFileTree toggles the file tree visibility.
func (a *App) ToggleFileTree() {
	a.treeVisible = !a.treeVisible
	if a.treeVisible && a.tree == nil {
		a.initFileTree()
	}
	if !a.treeVisible && a.focusArea == actions.FocusFileTree {
		a.focusArea = actions.FocusEditor
	}
}

// FileTreeWidth returns the file tree panel width.
func (a *App) FileTreeWidth() int {
	return a.treeWidth
}

// TreeViewportHeight returns the height available for tree node content.
func (a *App) TreeViewportHeight() int {
	return a.treeViewportHeight()
}

// --- Focus area interface ---

// FocusArea returns the current focus area.
func (a *App) FocusArea() actions.FocusArea {
	return a.focusArea
}

// SetFocusArea sets the current focus area.
func (a *App) SetFocusArea(area actions.FocusArea) {
	a.focusArea = area
}

// --- Input prompt interface ---

// Prompt returns the active prompt state, or nil.
func (a *App) Prompt() *editor.PromptState {
	return a.prompt
}

// SetPrompt sets or clears the prompt state.
func (a *App) SetPrompt(p *editor.PromptState) {
	a.prompt = p
}
