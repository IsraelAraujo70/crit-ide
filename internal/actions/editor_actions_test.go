package actions

import (
	"testing"

	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/events"
)

// mockClipboard implements ClipboardProvider for testing.
type mockClipboard struct {
	content string
}

func (c *mockClipboard) Read() (string, error)  { return c.content, nil }
func (c *mockClipboard) Write(text string) error { c.content = text; return nil }

// mockApp implements AppState for testing.
type mockApp struct {
	buffer        *editor.Buffer
	scrollY       int
	vpHeight      int
	scrWidth      int
	didQuit       bool
	clipboard     *mockClipboard
	inputMode     InputMode
	contextMenu   *editor.MenuState
	pendingAction string
	buffers       []*editor.Buffer
	activeIdx     int
	treeVisible   bool
	focus         FocusArea
	prompt        *editor.PromptState
	search        *editor.SearchState
}

func (m *mockApp) ActiveBuffer() *editor.Buffer          { return m.buffer }
func (m *mockApp) ScrollY() int                          { return m.scrollY }
func (m *mockApp) SetScrollY(y int)                      { m.scrollY = y }
func (m *mockApp) ViewportHeight() int                   { return m.vpHeight }
func (m *mockApp) ScreenWidth() int                      { return m.scrWidth }
func (m *mockApp) Quit()                                 { m.didQuit = true }
func (m *mockApp) Clipboard() ClipboardProvider          { return m.clipboard }
func (m *mockApp) InputMode() InputMode                  { return m.inputMode }
func (m *mockApp) SetInputMode(mode InputMode)           { m.inputMode = mode }
func (m *mockApp) ContextMenu() *editor.MenuState        { return m.contextMenu }
func (m *mockApp) SetContextMenu(menu *editor.MenuState) { m.contextMenu = menu }
func (m *mockApp) PostAction(actionID string)            { m.pendingAction = actionID }

// Tab / multi-buffer stubs.
func (m *mockApp) Buffers() []*editor.Buffer    { return m.buffers }
func (m *mockApp) ActiveBufferIndex() int        { return m.activeIdx }
func (m *mockApp) OpenFile(path string) error    { return nil }
func (m *mockApp) CloseBuffer(idx int)           {}
func (m *mockApp) SwitchBuffer(idx int)          { m.activeIdx = idx }

// File tree stubs.
func (m *mockApp) FileTree() FileTreeState       { return nil }
func (m *mockApp) FileTreeVisible() bool         { return m.treeVisible }
func (m *mockApp) SetFileTreeVisible(v bool)     { m.treeVisible = v }
func (m *mockApp) ToggleFileTree()               { m.treeVisible = !m.treeVisible }
func (m *mockApp) FileTreeWidth() int            { return 30 }
func (m *mockApp) TreeViewportHeight() int       { return m.vpHeight - 3 }

// Focus area stubs.
func (m *mockApp) FocusArea() FocusArea          { return m.focus }
func (m *mockApp) SetFocusArea(area FocusArea)   { m.focus = area }

// Prompt stubs.
func (m *mockApp) Prompt() *editor.PromptState         { return m.prompt }
func (m *mockApp) SetPrompt(p *editor.PromptState)     { m.prompt = p }

// LSP stubs.
func (m *mockApp) LSPServer(langID string) any                   { return nil }
func (m *mockApp) SetStatusMessage(msg string)                   {}
func (m *mockApp) NavigateToPosition(path string, line, col int) {}

// Search stubs.
func (m *mockApp) SearchState() *editor.SearchState      { return m.search }
func (m *mockApp) SetSearchState(s *editor.SearchState)   { m.search = s }

// File finder stubs.
func (m *mockApp) FinderState() *editor.FinderState                  { return nil }
func (m *mockApp) SetFinderState(f *editor.FinderState)              {}
func (m *mockApp) FinderFilter(pattern string) []editor.FinderResult { return nil }
func (m *mockApp) FinderRebuildCache()                               {}
func (m *mockApp) FinderFileCount() int                              { return 0 }

// Completion stubs.
func (m *mockApp) CompletionState() *editor.CompletionState    { return nil }
func (m *mockApp) SetCompletionState(c *editor.CompletionState) {}
func (m *mockApp) TriggerCompletion(triggerChar string)         {}

// Command palette stubs.
func (m *mockApp) PaletteState() *editor.PaletteState    { return nil }
func (m *mockApp) SetPaletteState(p *editor.PaletteState) {}

// Project search stubs.
func (m *mockApp) ProjectSearchState() *editor.ProjectSearchState      { return nil }
func (m *mockApp) SetProjectSearchState(ps *editor.ProjectSearchState) {}
func (m *mockApp) RunProjectSearch(query string)                       {}
func (m *mockApp) ProjectRoot() string                                 { return "/tmp" }

// Git stubs.
func (m *mockApp) GitStatusState() *editor.GitStatusState      { return nil }
func (m *mockApp) SetGitStatusState(gs *editor.GitStatusState)  {}
func (m *mockApp) GitGraphState() *editor.GitGraphState        { return nil }
func (m *mockApp) SetGitGraphState(gg *editor.GitGraphState)    {}
func (m *mockApp) GitDiffState() *editor.GitDiffState          { return nil }
func (m *mockApp) SetGitDiffState(gd *editor.GitDiffState)      {}

// Code actions stubs.
func (m *mockApp) CodeActionsState() *editor.CodeActionsState        { return nil }
func (m *mockApp) SetCodeActionsState(ca *editor.CodeActionsState)   {}
func (m *mockApp) ApplyCodeAction(idx int)                           {}

// Signature help stubs.
func (m *mockApp) SignatureHelpState() *editor.SignatureHelpState     { return nil }
func (m *mockApp) SetSignatureHelpState(sh *editor.SignatureHelpState) {}

// Minimap stubs.
func (m *mockApp) MinimapVisible() bool    { return false }
func (m *mockApp) SetMinimapVisible(v bool) {}
func (m *mockApp) ToggleMinimap()          {}

func newTestContext(app *mockApp, actionID string, payload any) *ActionContext {
	return &ActionContext{
		App: app,
		Event: &events.Event{
			Type:     events.EventAction,
			ActionID: actionID,
			Payload:  payload,
		},
	}
}

func TestCursorActions(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := editor.NewBuffer("t")
	buf.InsertChar('a')
	buf.InsertChar('b')
	buf.InsertNewline()
	buf.InsertChar('c')
	// Buffer: "ab\nc", cursor at (1,1).

	app := &mockApp{buffer: buf, vpHeight: 24}

	// cursor.up
	ctx := newTestContext(app, "cursor.up", nil)
	reg.Execute("cursor.up", ctx)
	if buf.CursorRow != 0 {
		t.Fatalf("after cursor.up: expected row 0, got %d", buf.CursorRow)
	}

	// cursor.down
	reg.Execute("cursor.down", newTestContext(app, "cursor.down", nil))
	if buf.CursorRow != 1 {
		t.Fatalf("after cursor.down: expected row 1, got %d", buf.CursorRow)
	}

	// cursor.home
	reg.Execute("cursor.home", newTestContext(app, "cursor.home", nil))
	if buf.CursorCol != 0 {
		t.Fatalf("after cursor.home: expected col 0, got %d", buf.CursorCol)
	}

	// cursor.end
	reg.Execute("cursor.end", newTestContext(app, "cursor.end", nil))
	if buf.CursorCol != 1 {
		t.Fatalf("after cursor.end: expected col 1, got %d", buf.CursorCol)
	}
}

func TestInsertCharAction(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := editor.NewBuffer("t")
	app := &mockApp{buffer: buf, vpHeight: 24}

	ctx := newTestContext(app, "insert.char", 'X')
	reg.Execute("insert.char", ctx)

	if buf.Text.Line(0) != "X" {
		t.Fatalf("expected %q, got %q", "X", buf.Text.Line(0))
	}
}

func TestInsertNewlineAction(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := editor.NewBuffer("t")
	buf.InsertChar('a')
	app := &mockApp{buffer: buf, vpHeight: 24}

	reg.Execute("insert.newline", newTestContext(app, "insert.newline", nil))

	if buf.Text.LineCount() != 2 {
		t.Fatalf("expected 2 lines, got %d", buf.Text.LineCount())
	}
}

func TestDeleteActions(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := editor.NewBuffer("t")
	buf.InsertChar('a')
	buf.InsertChar('b')
	buf.InsertChar('c')
	app := &mockApp{buffer: buf, vpHeight: 24}

	// delete.backward removes 'c'
	reg.Execute("delete.backward", newTestContext(app, "delete.backward", nil))
	if buf.Text.Line(0) != "ab" {
		t.Fatalf("after delete.backward: expected %q, got %q", "ab", buf.Text.Line(0))
	}

	// Move to start, delete.forward removes 'a'
	buf.CursorHome()
	reg.Execute("delete.forward", newTestContext(app, "delete.forward", nil))
	if buf.Text.Line(0) != "b" {
		t.Fatalf("after delete.forward: expected %q, got %q", "b", buf.Text.Line(0))
	}
}

func TestAppQuitAction(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := editor.NewBuffer("t")
	app := &mockApp{buffer: buf, vpHeight: 24}

	reg.Execute("app.quit", newTestContext(app, "app.quit", nil))
	if !app.didQuit {
		t.Fatal("expected app to quit")
	}
}

func TestScrollActions(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	// Create a buffer with 50 lines.
	buf := editor.NewBuffer("t")
	for i := 0; i < 49; i++ {
		buf.InsertChar(rune('a' + (i % 26)))
		buf.InsertNewline()
	}
	buf.InsertChar('z')
	// Cursor at last line. Move to top.
	buf.CursorRow = 0
	buf.CursorCol = 0

	app := &mockApp{buffer: buf, vpHeight: 10, scrollY: 0}

	// scroll.down
	reg.Execute("scroll.down", newTestContext(app, "scroll.down", nil))
	if app.scrollY != 10 {
		t.Fatalf("after scroll.down: expected scrollY 10, got %d", app.scrollY)
	}

	// scroll.up
	reg.Execute("scroll.up", newTestContext(app, "scroll.up", nil))
	if app.scrollY != 0 {
		t.Fatalf("after scroll.up: expected scrollY 0, got %d", app.scrollY)
	}
}

func TestUnknownAction(t *testing.T) {
	reg := NewRegistry()
	err := reg.Execute("does.not.exist", &ActionContext{})
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
}
