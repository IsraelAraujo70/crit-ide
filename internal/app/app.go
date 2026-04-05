package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/israelcorrea/crit-ide/internal/actions"
	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/events"
	"github.com/israelcorrea/crit-ide/internal/highlight"
	"github.com/israelcorrea/crit-ide/internal/input"
	"github.com/israelcorrea/crit-ide/internal/lsp"
	"github.com/israelcorrea/crit-ide/internal/render"
	"github.com/israelcorrea/crit-ide/internal/theme"
)

// App is the top-level application state and event loop.
type App struct {
	screen      tcell.Screen
	bus         *events.Bus
	registry    *actions.Registry
	renderer    *render.Renderer
	buffer      *editor.Buffer
	scrollY     int
	quit        bool
	filePath    string
	highlighter *highlight.RegexHighlighter
	langReg     *highlight.LangRegistry
	theme       *theme.Theme
	statusMsg   string // Temporary status message (e.g., hover result).

	// LSP.
	lspManager  *lsp.Manager
	diagStore   *lsp.DiagnosticsStore
	lastContent string // Tracks buffer content for change detection.
}

// New creates a new App. If filePath is non-empty, that file will be opened.
func New(filePath string) *App {
	return &App{
		filePath:    filePath,
		bus:         events.NewBus(256),
		registry:    actions.NewRegistry(),
		highlighter: highlight.NewRegexHighlighter(),
		langReg:     highlight.DefaultRegistry(),
		theme:       theme.DefaultTheme(),
		diagStore:   lsp.NewDiagnosticsStore(),
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
	actions.RegisterLSPActions(a.registry)

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

	// Detect language and configure highlighter.
	a.detectLanguage()

	// Initialize LSP manager using project root.
	rootPath := a.projectRoot()
	a.lspManager = lsp.NewManager(a.bus, rootPath)
	defer a.lspManager.StopAll()

	// Try to start the LSP server for the detected language.
	a.startLSPForBuffer()

	// Launch input goroutine.
	inputHandler := input.NewHandler(a.screen, a.bus)
	go inputHandler.Run()

	// Track initial content for change detection.
	a.lastContent = a.buffer.Text.Content()

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
			a.notifyLSPIfChanged()

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

		case events.EventLSPServerState:
			// Could show server state in statusline; for now just log.
		}

		a.render()
	}

	return nil
}

// detectLanguage sets the buffer's LanguageID and configures the highlighter.
func (a *App) detectLanguage() {
	if a.buffer.Path == "" {
		return
	}
	def := a.langReg.DetectLanguage(a.buffer.Path)
	if def != nil {
		a.buffer.LanguageID = def.ID
		a.highlighter.SetLanguageDef(def)
	}
}

// projectRoot returns the project root (directory of the open file, or cwd).
func (a *App) projectRoot() string {
	if a.buffer.Path != "" {
		return filepath.Dir(a.buffer.Path)
	}
	return "."
}

// startLSPForBuffer tries to start the LSP server for the buffer's language.
func (a *App) startLSPForBuffer() {
	if a.buffer.LanguageID == "" {
		return
	}
	srv, err := a.lspManager.EnsureServer(a.buffer.LanguageID)
	if err != nil {
		// Server not available — that's fine, editor works without LSP.
		return
	}
	// Notify the server about the open document.
	uri := lsp.URIFromPath(a.buffer.Path)
	srv.DidOpen(uri, a.buffer.LanguageID, a.buffer.Text.Content())
}

// notifyLSPIfChanged sends didChange to the LSP server if buffer content changed.
func (a *App) notifyLSPIfChanged() {
	if a.buffer.LanguageID == "" {
		return
	}
	content := a.buffer.Text.Content()
	if content == a.lastContent {
		return
	}
	a.lastContent = content

	srv := a.lspManager.ServerFor(a.buffer.LanguageID)
	if srv == nil {
		return
	}
	uri := lsp.URIFromPath(a.buffer.Path)
	srv.DidChange(uri, content)
}

// notifyLSPSave sends didSave to the LSP server.
func (a *App) notifyLSPSave() {
	if a.buffer.LanguageID == "" {
		return
	}
	srv := a.lspManager.ServerFor(a.buffer.LanguageID)
	if srv == nil {
		return
	}
	uri := lsp.URIFromPath(a.buffer.Path)
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

	// If same file, navigate to the position.
	if path == a.buffer.Path {
		lineContent := a.buffer.Text.Line(loc.Range.Start.Line)
		_, byteCol := lsp.LSPToEditorPosition(loc.Range.Start, lineContent)
		a.buffer.CursorRow = loc.Range.Start.Line
		a.buffer.CursorCol = byteCol
		a.ensureCursorVisible()
	} else {
		// Different file — show in statusline (no multi-buffer yet).
		a.statusMsg = fmt.Sprintf("→ %s:%d", filepath.Base(path), loc.Range.Start.Line+1)
	}
}

// handleHover processes a hover response.
func (a *App) handleHover(p *lsp.HoverPayload) {
	msg := p.Contents.Value
	// Truncate for statusline display.
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
	// Apply edits in reverse order to preserve positions.
	edits := p.Edits
	for i := len(edits) - 1; i >= 0; i-- {
		edit := edits[i]
		startLine := edit.Range.Start.Line
		startContent := ""
		if startLine < a.buffer.Text.LineCount() {
			startContent = a.buffer.Text.Line(startLine)
		}
		endLine := edit.Range.End.Line
		endContent := ""
		if endLine < a.buffer.Text.LineCount() {
			endContent = a.buffer.Text.Line(endLine)
		}

		_, startCol := lsp.LSPToEditorPosition(edit.Range.Start, startContent)
		_, endCol := lsp.LSPToEditorPosition(edit.Range.End, endContent)

		// Delete the range.
		_ = a.buffer.Text.Delete(editor.Range{
			Start: editor.Position{Line: startLine, Col: startCol},
			End:   editor.Position{Line: endLine, Col: endCol},
		})
		// Insert new text.
		if edit.NewText != "" {
			_ = a.buffer.Text.Insert(
				editor.Position{Line: startLine, Col: startCol},
				edit.NewText,
			)
		}
	}
	a.buffer.Dirty = true
	a.buffer.ClampCursor()
	a.statusMsg = "Formatted"
}

// render draws the current state to the screen.
func (a *App) render() {
	w, h := a.screen.Size()

	// Build diagnostic ranges for the renderer.
	var diagRanges []render.DiagnosticRange
	var diagErrors, diagWarnings int
	if a.buffer.Path != "" {
		uri := lsp.URIFromPath(a.buffer.Path)
		diags := a.diagStore.ForURI(uri)
		for _, d := range diags {
			startLine := d.Range.Start.Line
			startContent := ""
			if startLine < a.buffer.Text.LineCount() {
				startContent = a.buffer.Text.Line(startLine)
			}
			endLine := d.Range.End.Line
			endContent := ""
			if endLine < a.buffer.Text.LineCount() {
				endContent = a.buffer.Text.Line(endLine)
			}

			_, startCol := lsp.LSPToEditorPosition(d.Range.Start, startContent)
			_, endCol := lsp.LSPToEditorPosition(d.Range.End, endContent)

			diagRanges = append(diagRanges, render.DiagnosticRange{
				Line:     startLine,
				StartCol: startCol,
				EndCol:   endCol,
				Severity: int(d.Severity),
			})
		}
		diagErrors, diagWarnings = a.diagStore.CountsByURI(uri)
	}

	a.renderer.Render(&render.ViewState{
		Buffer:       a.buffer,
		ScrollY:      a.scrollY,
		Width:        w,
		Height:       h,
		Highlighter:  a.highlighter,
		Theme:        a.theme,
		StatusMsg:    a.statusMsg,
		Diagnostics:  diagRanges,
		DiagErrors:   diagErrors,
		DiagWarnings: diagWarnings,
	})
	// Clear status message after displaying once.
	a.statusMsg = ""
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

// SetStatusMessage sets a temporary message to display in the statusline.
func (a *App) SetStatusMessage(msg string) {
	a.statusMsg = msg
}

// LSPServer returns the running LSP server for the given language, or nil.
func (a *App) LSPServer(langID string) any {
	if a.lspManager == nil {
		return nil
	}
	return a.lspManager.ServerFor(langID)
}

// NavigateToPosition moves the cursor to the given position.
// If the path differs from the current buffer, shows it in the statusline.
func (a *App) NavigateToPosition(path string, line, col int) {
	if path == a.buffer.Path {
		a.buffer.CursorRow = line
		a.buffer.CursorCol = col
		a.buffer.ClampCursor()
		a.ensureCursorVisible()
	} else {
		a.statusMsg = fmt.Sprintf("→ %s:%d", filepath.Base(path), line+1)
	}
}
