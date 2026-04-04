package app

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/israelcorrea/crit-ide/internal/actions"
	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/events"
	"github.com/israelcorrea/crit-ide/internal/input"
	"github.com/israelcorrea/crit-ide/internal/render"
)

// App is the top-level application state and event loop.
type App struct {
	screen   tcell.Screen
	bus      *events.Bus
	registry *actions.Registry
	renderer *render.Renderer
	buffer   *editor.Buffer
	scrollY  int
	quit     bool
	filePath string
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
			_ = a.registry.Execute(ev.ActionID, ctx)
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
	a.renderer.Render(&render.ViewState{
		Buffer:  a.buffer,
		ScrollY: a.scrollY,
		Width:   w,
		Height:  h,
	})
}

// ensureCursorVisible adjusts scrollY so the cursor is within the viewport.
func (a *App) ensureCursorVisible() {
	_, h := a.screen.Size()
	editorHeight := h - 1 // Statusline takes 1 row.
	if editorHeight < 1 {
		editorHeight = 1
	}

	if a.buffer.CursorRow < a.scrollY {
		a.scrollY = a.buffer.CursorRow
	}
	if a.buffer.CursorRow >= a.scrollY+editorHeight {
		a.scrollY = a.buffer.CursorRow - editorHeight + 1
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

// Quit signals the application to exit.
func (a *App) Quit() {
	a.quit = true
}
