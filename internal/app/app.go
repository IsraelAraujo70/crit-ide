package app

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/israelcorrea/crit-ide/internal/actions"
	"github.com/israelcorrea/crit-ide/internal/clipboard"
	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/events"
	"github.com/israelcorrea/crit-ide/internal/input"
	"github.com/israelcorrea/crit-ide/internal/render"
)

// App is the top-level application state and event loop.
type App struct {
	screen        tcell.Screen
	bus           *events.Bus
	registry      *actions.Registry
	renderer      *render.Renderer
	buffer        *editor.Buffer
	scrollY       int
	quit          bool
	filePath      string
	clip          actions.ClipboardProvider
	inputMode     actions.InputMode
	contextMenu   *editor.MenuState
	pendingAction string
}

// New creates a new App. If filePath is non-empty, that file will be opened.
func New(filePath string) *App {
	return &App{
		filePath: filePath,
		bus:      events.NewBus(256),
		registry: actions.NewRegistry(),
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
			// If file doesn't exist, create a new file buffer.
			buf = editor.NewBuffer("main")
			buf.Path = a.filePath
			buf.Kind = editor.BufferKindFile
		}
		a.buffer = buf
	} else {
		a.buffer = editor.NewBuffer("scratch")
	}

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
				_ = a.registry.Execute(ev.ActionID, ctx)
			case actions.ModeContextMenu:
				a.handleContextMenuAction(ev.ActionID, ctx)
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

// render draws the current state to the screen.
func (a *App) render() {
	w, h := a.screen.Size()
	vs := &render.ViewState{
		Buffer:  a.buffer,
		ScrollY: a.scrollY,
		Width:   w,
		Height:  h,
	}
	if a.contextMenu != nil {
		vs.Popup = a.contextMenu
	}
	a.renderer.Render(vs)
}

// ensureCursorVisible adjusts scrollY so the cursor is within the viewport.
func (a *App) ensureCursorVisible() {
	_, h := a.screen.Size()
	editorHeight := h - 1 // Statusline takes 1 row.
	editorHeight = max(editorHeight, 1)

	if a.buffer.CursorRow < a.scrollY {
		a.scrollY = a.buffer.CursorRow
	}
	if a.buffer.CursorRow >= a.scrollY+editorHeight {
		a.scrollY = a.buffer.CursorRow - editorHeight + 1
	}
}

// handleContextMenuAction routes actions while the context menu is open.
// Only menu-related actions and clicks are processed; everything else is ignored.
func (a *App) handleContextMenuAction(actionID string, ctx *actions.ActionContext) {
	// Remap normal actions to menu actions when in menu mode.
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
		// Click while menu open — treat as menu click (action decides
		// whether click is inside or outside the menu).
		ctx.Event.ActionID = "menu.click"
		_ = a.registry.Execute("menu.click", ctx)
	default:
		// Ignore all other actions while menu is open.
	}
}

// --- AppState interface implementation ---

// ActiveBuffer returns the currently focused buffer.
func (a *App) ActiveBuffer() *editor.Buffer {
	return a.buffer
}

// ScrollY returns the current vertical scroll offset.
func (a *App) ScrollY() int {
	return a.scrollY
}

// SetScrollY sets the vertical scroll offset.
func (a *App) SetScrollY(y int) {
	a.scrollY = y
}

// ViewportHeight returns the number of visible editor rows (excluding statusline).
func (a *App) ViewportHeight() int {
	_, h := a.screen.Size()
	eh := h - 1
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
