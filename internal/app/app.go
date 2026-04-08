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
	"github.com/israelcorrea/crit-ide/internal/fuzzy"
	"github.com/israelcorrea/crit-ide/internal/highlight"
	"github.com/israelcorrea/crit-ide/internal/input"
	"github.com/israelcorrea/crit-ide/internal/logger"
	"github.com/israelcorrea/crit-ide/internal/lsp"
	"github.com/israelcorrea/crit-ide/internal/render"
	"github.com/israelcorrea/crit-ide/internal/theme"
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

	// Search state.
	search *editor.SearchState

	// File finder.
	finder     *editor.FinderState
	fileFinder *fuzzy.FileFinder

	// Completion.
	completion *editor.CompletionState

	// Syntax highlighting.
	highlighter          *highlight.TreeSitterHighlighter
	langReg              *highlight.TSLangRegistry
	theme                *theme.Theme
	lastHighlightContent string // Tracks content to avoid redundant reparses.

	// LSP.
	lspManager  *lsp.Manager
	diagStore   *lsp.DiagnosticsStore
	lastContent map[editor.BufferID]string // Tracks buffer content for change detection.
	statusMsg   string                     // Temporary status message.
}

// New creates a new App. If filePath is non-empty, that file will be opened.
func New(filePath string) *App {
	reg := highlight.DefaultTSRegistry()
	return &App{
		filePath:    filePath,
		bus:         events.NewBus(256),
		registry:    actions.NewRegistry(),
		scrollYs:    make(map[editor.BufferID]int),
		treeWidth:   defaultTreeWidth,
		highlighter: highlight.NewTreeSitterHighlighter(reg),
		langReg:     reg,
		theme:       theme.DefaultTheme(),
		diagStore:   lsp.NewDiagnosticsStore(),
		lastContent: make(map[editor.BufferID]string),
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
	actions.RegisterLSPActions(a.registry)

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

	// Detect language and configure highlighter for initial buffer.
	a.detectLanguage(a.ActiveBuffer())

	// Initialize file tree from current working directory or file's directory.
	a.initFileTree()

	// Initialize fuzzy file finder cache.
	a.fileFinder = fuzzy.NewFileFinder(a.projectRoot())

	// Initialize LSP manager.
	rootPath := a.projectRoot()
	a.lspManager = lsp.NewManager(a.bus, rootPath)
	defer a.lspManager.StopAll()

	// Start LSP server for initial buffer.
	a.startLSPForBuffer(a.ActiveBuffer())
	a.lastContent[a.ActiveBuffer().ID] = a.ActiveBuffer().Text.Content()

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
			case actions.ModeSearch:
				a.handleSearchAction(ev.ActionID, ctx)
			case actions.ModeFileFinder:
				a.handleFinderAction(ev.ActionID, ctx)
			case actions.ModeCompletion:
				a.handleCompletionAction(ev.ActionID, ctx)
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
			a.notifyLSPIfChanged()

			// Update highlighter source only when content actually changed.
			buf := a.ActiveBuffer()
			if content := buf.Text.Content(); content != a.lastHighlightContent {
				a.highlighter.SetSource(content)
				a.lastHighlightContent = content
			}

			// If save action, notify LSP.
			if ev.ActionID == "file.save" {
				a.notifyLSPSave()
			}

		case events.EventResize:
			a.screen.Sync()
			a.ensureCursorVisible()

		case events.EventQuit:
			a.quit = true
			continue

		case events.EventLSPDiagnostics:
			if p, ok := ev.Payload.(*lsp.DiagnosticsPayload); ok {
				a.diagStore.Update(p.URI, p.Diagnostics)
			}

		case events.EventLSPDefinition:
			if p, ok := ev.Payload.(*lsp.DefinitionPayload); ok {
				a.handleDefinition(p)
			}

		case events.EventLSPHover:
			if p, ok := ev.Payload.(*lsp.HoverPayload); ok {
				a.handleHover(p)
			}

		case events.EventLSPFormat:
			if p, ok := ev.Payload.(*lsp.FormatPayload); ok {
				a.handleFormat(p)
			}

		case events.EventLSPCompletion:
			if p, ok := ev.Payload.(*lsp.CompletionPayload); ok {
				a.handleCompletion(p)
			}

		case events.EventLSPServerState:
			// Server state change — could show in statusline in the future.
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
		"file.save", "tree.refresh", "edit.undo", "edit.redo",
		"search.open", "finder.open", "completion.trigger":
		_ = a.registry.Execute(actionID, ctx)
		return
	}

	// Select word is a global editor action.
	if actionID == "select.word" {
		if a.focusArea == actions.FocusEditor {
			_ = a.registry.Execute(actionID, ctx)
		}
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

	// After inserting a character in normal mode, check for auto-trigger.
	if actionID == "insert.char" && a.focusArea == actions.FocusEditor {
		if ch, ok := ctx.Event.Payload.(rune); ok {
			a.maybeAutoTriggerCompletion(string(ch))
		}
	}
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

// treeViewportHeight returns the height available for tree content.
func (a *App) treeViewportHeight() int {
	_, h := a.screen.Size()
	// tab bar (1) + top border (1) + tree header (1) + statusline (1) = 4
	tvh := h - 4
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
		Highlighter:  a.highlighter,
		Theme:        a.theme,
		StatusMsg:    a.statusMsg,
	}

	if a.contextMenu != nil {
		vs.Popup = a.contextMenu
	}
	if a.prompt != nil {
		vs.Prompt = a.prompt
	}
	if a.search != nil {
		vs.Search = a.search
	}
	if a.finder != nil {
		vs.Finder = a.finder
	}
	if a.completion != nil {
		vs.Completion = a.completion
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

	// Build diagnostic ranges for the renderer.
	if buf.Path != "" {
		uri := lsp.URIFromPath(buf.Path)
		diags := a.diagStore.ForURI(uri)
		for _, d := range diags {
			startLine := d.Range.Start.Line
			startContent := ""
			if startLine < buf.Text.LineCount() {
				startContent = buf.Text.Line(startLine)
			}
			endLine := d.Range.End.Line
			endContent := ""
			if endLine < buf.Text.LineCount() {
				endContent = buf.Text.Line(endLine)
			}

			_, startCol := lsp.LSPToEditorPosition(d.Range.Start, startContent)
			_, endCol := lsp.LSPToEditorPosition(d.Range.End, endContent)

			vs.Diagnostics = append(vs.Diagnostics, render.DiagnosticRange{
				Line:     startLine,
				StartCol: startCol,
				EndCol:   endCol,
				Severity: int(d.Severity),
			})
		}
		vs.DiagErrors, vs.DiagWarnings = a.diagStore.CountsByURI(uri)
	}

	a.renderer.Render(vs)

	// Clear status message after displaying once.
	a.statusMsg = ""
}

// ensureCursorVisible adjusts scrollY so the cursor is within the viewport.
func (a *App) ensureCursorVisible() {
	editorHeight := a.ViewportHeight()

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

// handleSearchAction routes actions while the Find/Replace bar is active.
func (a *App) handleSearchAction(actionID string, ctx *actions.ActionContext) {
	// Determine if we're in the replace field.
	inReplace := a.search != nil && a.search.ActiveField == editor.FieldReplace

	// When in replace field, Enter triggers replace-one, not find-next.
	if inReplace && actionID == "insert.newline" {
		_ = a.registry.Execute("search.replace", ctx)
		return
	}

	// Ctrl+A in search mode → replace all (when replace is visible).
	if actionID == "select.all" && a.search != nil && a.search.ShowReplace {
		_ = a.registry.Execute("search.replace_all", ctx)
		return
	}

	remap := map[string]string{
		"insert.char":     "search.char",
		"delete.backward": "search.backspace",
		"delete.forward":  "search.delete",
		"cursor.left":     "search.left",
		"cursor.right":    "search.right",
		"cursor.home":     "search.home",
		"cursor.end":      "search.end",
		"input.escape":    "search.close",
		"insert.newline":  "search.next",
	}
	if mapped, ok := remap[actionID]; ok {
		_ = a.registry.Execute(mapped, ctx)
		return
	}

	// Allow direct search actions through (F3, Shift+F3).
	switch actionID {
	case "search.next", "search.prev", "search.replace", "search.replace_all":
		_ = a.registry.Execute(actionID, ctx)
		return
	}

	// Tab toggles between find/replace fields.
	if actionID == "insert.char" {
		if ch, ok := ctx.Event.Payload.(rune); ok && ch == '\t' {
			_ = a.registry.Execute("search.toggle_replace", ctx)
			return
		}
	}

	// Ignore all other actions while search is open.
}

// handleFinderAction routes actions while the fuzzy file finder is active.
func (a *App) handleFinderAction(actionID string, ctx *actions.ActionContext) {
	remap := map[string]string{
		"insert.char":     "finder.char",
		"delete.backward": "finder.backspace",
		"delete.forward":  "finder.delete",
		"cursor.left":     "finder.left",
		"cursor.right":    "finder.right",
		"cursor.home":     "finder.home",
		"cursor.end":      "finder.end",
		"cursor.up":       "finder.up",
		"cursor.down":     "finder.down",
		"input.escape":    "finder.close",
		"insert.newline":  "finder.confirm",
	}
	if mapped, ok := remap[actionID]; ok {
		_ = a.registry.Execute(mapped, ctx)
		return
	}
	// Ignore all other actions while finder is open.
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
	// tab bar (1) + focus border (1) + statusline (1) = 3 reserved rows
	eh := h - 3
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

	// Detect language and set up highlighting.
	a.detectLanguage(buf)

	// Start LSP server and notify didOpen.
	a.startLSPForBuffer(buf)
	a.lastContent[buf.ID] = buf.Text.Content()

	return nil
}

// CloseBuffer closes the buffer at the given index.
func (a *App) CloseBuffer(idx int) {
	if idx < 0 || idx >= len(a.buffers) || len(a.buffers) <= 1 {
		return
	}

	// Notify LSP server about buffer close.
	closingBuf := a.buffers[idx]
	if closingBuf.Path != "" && closingBuf.LanguageID != "" && a.lspManager != nil {
		if srv := a.lspManager.ServerFor(closingBuf.LanguageID); srv != nil {
			uri := lsp.URIFromPath(closingBuf.Path)
			srv.DidClose(uri)
		}
	}
	delete(a.lastContent, closingBuf.ID)

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
		// Switch highlighter to the new buffer's language and source.
		buf := a.buffers[idx]
		if buf.LanguageID != "" {
			a.highlighter.SetLanguage(buf.LanguageID)
			content := buf.Text.Content()
			a.highlighter.SetSource(content)
			a.lastHighlightContent = content
		} else {
			a.highlighter.SetLanguage("")
			a.lastHighlightContent = ""
		}
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

// ToggleFileTree toggles focus between editor and file tree.
// If the tree is hidden, it shows it and focuses it.
// If the tree is visible and editor has focus, focus moves to tree.
// If the tree is visible and tree has focus, focus moves to editor.
func (a *App) ToggleFileTree() {
	if !a.treeVisible {
		// Tree hidden → show and focus.
		a.treeVisible = true
		if a.tree == nil {
			a.initFileTree()
		}
		a.focusArea = actions.FocusFileTree
		return
	}
	// Tree visible → toggle focus.
	if a.focusArea == actions.FocusFileTree {
		a.focusArea = actions.FocusEditor
	} else {
		a.focusArea = actions.FocusFileTree
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

// --- Search state interface ---

// SearchState returns the active search state, or nil.
func (a *App) SearchState() *editor.SearchState {
	return a.search
}

// SetSearchState sets or clears the search state.
func (a *App) SetSearchState(s *editor.SearchState) {
	a.search = s
}

// --- File finder interface ---

// FinderState returns the active finder state, or nil.
func (a *App) FinderState() *editor.FinderState {
	return a.finder
}

// SetFinderState sets or clears the finder state.
func (a *App) SetFinderState(f *editor.FinderState) {
	a.finder = f
}

// FinderFilter returns fuzzy-filtered file results for the given pattern.
func (a *App) FinderFilter(pattern string) []editor.FinderResult {
	if a.fileFinder == nil {
		return nil
	}
	fResults := a.fileFinder.Filter(pattern)
	results := make([]editor.FinderResult, len(fResults))
	for i, fr := range fResults {
		results[i] = editor.FinderResult{
			RelPath: fr.RelPath,
			AbsPath: fr.AbsPath,
			Matches: fr.Matches,
		}
	}
	return results
}

// FinderRebuildCache rebuilds the file finder cache (called on tree.refresh).
func (a *App) FinderRebuildCache() {
	if a.fileFinder != nil {
		a.fileFinder.Rebuild()
	}
}

// FinderFileCount returns the total number of indexed files.
func (a *App) FinderFileCount() int {
	if a.fileFinder != nil {
		return a.fileFinder.FileCount()
	}
	return 0
}

// --- Completion interface ---

// CompletionState returns the active completion state, or nil.
func (a *App) CompletionState() *editor.CompletionState {
	return a.completion
}

// SetCompletionState sets or clears the completion state.
func (a *App) SetCompletionState(c *editor.CompletionState) {
	a.completion = c
}

// TriggerCompletion initiates an LSP completion request.
func (a *App) TriggerCompletion(triggerChar string) {
	buf := a.ActiveBuffer()
	if buf.LanguageID == "" {
		return
	}
	srvAny := a.LSPServer(buf.LanguageID)
	if srvAny == nil {
		return
	}
	srv, ok := srvAny.(*lsp.Server)
	if !ok || srv.TriggerCharacters() == nil && triggerChar != "" {
		// If server has no completion support at all, skip auto-trigger.
		// Manual trigger (triggerChar == "") still proceeds.
	}

	uri := lsp.URIFromPath(buf.Path)
	lineContent := buf.Text.Line(buf.CursorRow)
	pos := lsp.EditorToLSPPosition(buf.CursorRow, buf.CursorCol, lineContent)

	triggerKind := lsp.CompletionTriggerInvoked
	if triggerChar != "" {
		triggerKind = lsp.CompletionTriggerTriggerCharacter
	}
	srv.Completion(uri, pos, triggerKind, triggerChar)
}

// handleCompletion processes completion results from LSP.
func (a *App) handleCompletion(p *lsp.CompletionPayload) {
	if len(p.Items) == 0 {
		return
	}

	buf := a.ActiveBuffer()

	// Convert LSP items to editor items.
	items := make([]editor.CompletionItem, len(p.Items))
	for i, lspItem := range p.Items {
		items[i] = editor.CompletionItem{
			Label:      lspItem.Label,
			Kind:       editor.CompletionItemKind(lspItem.Kind),
			Detail:     lspItem.Detail,
			InsertText: lspItem.InsertText,
			FilterText: lspItem.FilterText,
			SortText:   lspItem.SortText,
		}
	}

	// Calculate anchor position: find the start of the word being typed.
	anchorRow := buf.CursorRow
	anchorCol := buf.CursorCol
	line := buf.Text.Line(anchorRow)

	// Walk backwards to find the start of the identifier.
	for anchorCol > 0 {
		if anchorCol > len(line) {
			anchorCol = len(line)
		}
		ch := line[anchorCol-1]
		if isIdentChar(ch) {
			anchorCol--
		} else {
			break
		}
	}

	prefix := ""
	if anchorCol < buf.CursorCol && anchorCol < len(line) {
		prefix = line[anchorCol:buf.CursorCol]
	}

	a.completion = editor.NewCompletionState(items, anchorRow, anchorCol, prefix)

	// If no items match after filtering, don't show popup.
	if a.completion.IsEmpty() {
		a.completion = nil
		return
	}

	a.inputMode = actions.ModeCompletion
}

// handleCompletionAction routes actions while the completion popup is active.
func (a *App) handleCompletionAction(actionID string, ctx *actions.ActionContext) {
	switch actionID {
	case "completion.trigger":
		// Re-trigger while already completing — just request fresh items.
		a.TriggerCompletion("")
		return

	case "cursor.up":
		_ = a.registry.Execute("completion.up", ctx)
		return

	case "cursor.down":
		_ = a.registry.Execute("completion.down", ctx)
		return

	case "insert.newline", "edit.indent":
		// Accept completion on Enter or Tab.
		_ = a.registry.Execute("completion.accept", ctx)
		return

	case "input.escape":
		_ = a.registry.Execute("completion.dismiss", ctx)
		return

	case "insert.char":
		// First insert the character normally.
		_ = a.registry.Execute(actionID, ctx)

		// Then refilter the completion list.
		if a.completion != nil {
			buf := a.ActiveBuffer()
			line := buf.Text.Line(buf.CursorRow)
			// Recalculate prefix from anchor to current cursor.
			if buf.CursorRow == a.completion.AnchorRow && buf.CursorCol >= a.completion.AnchorCol {
				newPrefix := ""
				if a.completion.AnchorCol < len(line) && buf.CursorCol <= len(line) {
					newPrefix = line[a.completion.AnchorCol:buf.CursorCol]
				}
				a.completion.UpdatePrefix(newPrefix)
				if a.completion.IsEmpty() {
					a.completion = nil
					a.inputMode = actions.ModeNormal
				}
			} else {
				// Cursor moved to different line — dismiss.
				a.completion = nil
				a.inputMode = actions.ModeNormal
			}
		}

		// Check if the typed character is a trigger character for auto-completion.
		if a.completion == nil {
			if ch, ok := ctx.Event.Payload.(rune); ok {
				a.maybeAutoTriggerCompletion(string(ch))
			}
		}
		return

	case "delete.backward":
		_ = a.registry.Execute(actionID, ctx)

		// Refilter after backspace.
		if a.completion != nil {
			buf := a.ActiveBuffer()
			if buf.CursorRow == a.completion.AnchorRow && buf.CursorCol >= a.completion.AnchorCol {
				line := buf.Text.Line(buf.CursorRow)
				newPrefix := ""
				if a.completion.AnchorCol < len(line) && buf.CursorCol <= len(line) {
					newPrefix = line[a.completion.AnchorCol:buf.CursorCol]
				}
				a.completion.UpdatePrefix(newPrefix)
				if a.completion.IsEmpty() {
					a.completion = nil
					a.inputMode = actions.ModeNormal
				}
			} else {
				// Backspaced before anchor — dismiss.
				a.completion = nil
				a.inputMode = actions.ModeNormal
			}
		}
		return

	default:
		// Any other action: dismiss completion and process normally.
		a.completion = nil
		a.inputMode = actions.ModeNormal
		a.handleNormalAction(actionID, ctx)
		return
	}
}

// maybeAutoTriggerCompletion checks if the typed character is a trigger character
// and initiates completion if so.
func (a *App) maybeAutoTriggerCompletion(ch string) {
	buf := a.ActiveBuffer()
	if buf.LanguageID == "" {
		return
	}
	srvAny := a.LSPServer(buf.LanguageID)
	if srvAny == nil {
		return
	}
	srv, ok := srvAny.(*lsp.Server)
	if !ok {
		return
	}
	triggers := srv.TriggerCharacters()
	for _, tc := range triggers {
		if tc == ch {
			a.TriggerCompletion(ch)
			return
		}
	}
}

// isIdentChar returns true if the byte is a valid identifier character.
func isIdentChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

// --- LSP support interface ---

// LSPServer returns the running LSP server for the given language, or nil.
func (a *App) LSPServer(langID string) any {
	if a.lspManager == nil {
		return nil
	}
	return a.lspManager.ServerFor(langID)
}

// SetStatusMessage sets a temporary message to display in the statusline.
func (a *App) SetStatusMessage(msg string) {
	a.statusMsg = msg
}

// NavigateToPosition moves the cursor to the given position.
func (a *App) NavigateToPosition(path string, line, col int) {
	buf := a.ActiveBuffer()
	if path == buf.Path {
		buf.CursorRow = line
		buf.CursorCol = col
		buf.ClampCursor()
		a.ensureCursorVisible()
	} else {
		a.statusMsg = fmt.Sprintf("-> %s:%d", filepath.Base(path), line+1)
	}
}

// --- Syntax highlighting helpers ---

// detectLanguage sets the buffer's LanguageID and configures the highlighter.
func (a *App) detectLanguage(buf *editor.Buffer) {
	if buf.Path == "" {
		return
	}
	def := a.langReg.DetectLanguage(buf.Path)
	if def != nil {
		buf.LanguageID = def.ID
		a.highlighter.SetLanguage(def.ID)
		content := buf.Text.Content()
		a.highlighter.SetSource(content)
		a.lastHighlightContent = content
		logger.Info("highlight: detected language %q for %s", def.ID, buf.FileName())
	} else {
		// Reset highlighter so stale tokens from the previous buffer
		// are not applied to an unsupported file.
		buf.LanguageID = ""
		a.highlighter.SetLanguage("")
		a.lastHighlightContent = ""
	}
}

// --- LSP helpers ---

// projectRoot returns the project root directory.
func (a *App) projectRoot() string {
	if a.filePath != "" {
		absPath, _ := filepath.Abs(a.filePath)
		return filepath.Dir(absPath)
	}
	cwd, _ := os.Getwd()
	return cwd
}

// startLSPForBuffer starts an LSP server for the buffer's language.
func (a *App) startLSPForBuffer(buf *editor.Buffer) {
	if buf.LanguageID == "" || a.lspManager == nil {
		return
	}
	srv, err := a.lspManager.EnsureServer(buf.LanguageID)
	if err != nil {
		return // Server not available — editor works without LSP.
	}
	if buf.Path != "" {
		uri := lsp.URIFromPath(buf.Path)
		srv.DidOpen(uri, buf.LanguageID, buf.Text.Content())
	}
}

// notifyLSPIfChanged sends didChange to the LSP server if buffer content changed.
func (a *App) notifyLSPIfChanged() {
	buf := a.ActiveBuffer()
	if buf.LanguageID == "" || a.lspManager == nil {
		return
	}
	content := buf.Text.Content()
	if content == a.lastContent[buf.ID] {
		return
	}
	a.lastContent[buf.ID] = content

	srv := a.lspManager.ServerFor(buf.LanguageID)
	if srv == nil {
		return
	}
	uri := lsp.URIFromPath(buf.Path)
	srv.DidChange(uri, content)
}

// notifyLSPSave sends didSave to the LSP server.
func (a *App) notifyLSPSave() {
	buf := a.ActiveBuffer()
	if buf.LanguageID == "" || a.lspManager == nil {
		return
	}
	srv := a.lspManager.ServerFor(buf.LanguageID)
	if srv == nil {
		return
	}
	uri := lsp.URIFromPath(buf.Path)
	srv.DidSave(uri)
}

// handleDefinition processes a go-to-definition response.
func (a *App) handleDefinition(p *lsp.DefinitionPayload) {
	if len(p.Locations) == 0 {
		a.statusMsg = "No definition found"
		return
	}
	loc := p.Locations[0]
	path, err := lsp.PathFromURI(loc.URI)
	if err != nil {
		return
	}
	buf := a.ActiveBuffer()
	if path == buf.Path {
		lineContent := buf.Text.Line(loc.Range.Start.Line)
		_, byteCol := lsp.LSPToEditorPosition(loc.Range.Start, lineContent)
		buf.CursorRow = loc.Range.Start.Line
		buf.CursorCol = byteCol
		a.ensureCursorVisible()
	} else {
		a.statusMsg = fmt.Sprintf("-> %s:%d", filepath.Base(path), loc.Range.Start.Line+1)
	}
}

// handleHover processes a hover response.
func (a *App) handleHover(p *lsp.HoverPayload) {
	msg := p.Contents.Value
	if idx := strings.Index(msg, "\n"); idx >= 0 {
		msg = msg[:idx]
	}
	if len(msg) > 120 {
		msg = msg[:117] + "..."
	}
	a.statusMsg = msg
}

// handleFormat applies formatting edits from the LSP server.
func (a *App) handleFormat(p *lsp.FormatPayload) {
	buf := a.ActiveBuffer()
	edits := p.Edits
	for i := len(edits) - 1; i >= 0; i-- {
		edit := edits[i]
		startLine := edit.Range.Start.Line
		startContent := ""
		if startLine < buf.Text.LineCount() {
			startContent = buf.Text.Line(startLine)
		}
		endLine := edit.Range.End.Line
		endContent := ""
		if endLine < buf.Text.LineCount() {
			endContent = buf.Text.Line(endLine)
		}

		_, startCol := lsp.LSPToEditorPosition(edit.Range.Start, startContent)
		_, endCol := lsp.LSPToEditorPosition(edit.Range.End, endContent)

		_ = buf.Text.Delete(editor.Range{
			Start: editor.Position{Line: startLine, Col: startCol},
			End:   editor.Position{Line: endLine, Col: endCol},
		})
		if edit.NewText != "" {
			_ = buf.Text.Insert(
				editor.Position{Line: startLine, Col: startCol},
				edit.NewText,
			)
		}
	}
	buf.Dirty = true
	buf.ClampCursor()
	a.statusMsg = "Formatted"
}
