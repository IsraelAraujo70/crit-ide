package actions

import (
	"testing"

	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/events"
)

func TestLSPRenameOpensPrompt(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)
	RegisterLSPActions(reg)

	buf := editor.NewBuffer("test.go")
	// Type "myVar" so we have a word under cursor.
	for _, ch := range "myVar" {
		buf.InsertChar(ch)
	}
	buf.CursorCol = 2 // Inside "myVar"
	buf.LanguageID = "go"

	app := &mockApp{buffer: buf, vpHeight: 24}
	// No LSP server → should show status message.
	ctx := newTestContext(app, "lsp.rename", nil)
	err := reg.Execute("lsp.rename", ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Without LSP server, prompt should not open.
	if app.prompt != nil {
		t.Fatal("expected no prompt without LSP server")
	}
}

func TestWordUnderCursor(t *testing.T) {
	buf := editor.NewBuffer("test.go")
	for _, ch := range "hello world" {
		buf.InsertChar(ch)
	}
	buf.CursorRow = 0

	tests := []struct {
		col  int
		want string
	}{
		{0, "hello"},
		{2, "hello"},
		{4, "hello"},
		{5, "hello"}, // right after "hello", word scan finds it
		{6, "world"},
		{10, "world"},
	}

	for _, tt := range tests {
		buf.CursorCol = tt.col
		got := wordUnderCursor(buf)
		if got != tt.want {
			t.Errorf("col=%d: got %q, want %q", tt.col, got, tt.want)
		}
	}
}

func TestWordUnderCursorEmpty(t *testing.T) {
	buf := editor.NewBuffer("test.go")
	// Empty buffer.
	got := wordUnderCursor(buf)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestCodeActionsState(t *testing.T) {
	state := &editor.CodeActionsState{
		Items: []editor.CodeActionItem{
			{Title: "Organize Imports", Kind: "source.organizeImports", Index: 0},
			{Title: "Extract function", Kind: "refactor.extract", Index: 1},
			{Title: "Add missing import", Kind: "quickfix", Index: 2},
		},
		SelectedIdx: 0,
	}

	// Test navigation.
	state.MoveDown()
	if state.SelectedIdx != 1 {
		t.Fatalf("expected 1, got %d", state.SelectedIdx)
	}
	state.MoveDown()
	if state.SelectedIdx != 2 {
		t.Fatalf("expected 2, got %d", state.SelectedIdx)
	}
	state.MoveDown() // Should clamp.
	if state.SelectedIdx != 2 {
		t.Fatalf("expected 2 (clamped), got %d", state.SelectedIdx)
	}
	state.MoveUp()
	if state.SelectedIdx != 1 {
		t.Fatalf("expected 1, got %d", state.SelectedIdx)
	}
	state.MoveUp()
	state.MoveUp() // Should clamp.
	if state.SelectedIdx != 0 {
		t.Fatalf("expected 0 (clamped), got %d", state.SelectedIdx)
	}

	// Test selected item.
	item := state.SelectedItem()
	if item == nil {
		t.Fatal("expected non-nil item")
	}
	if item.Title != "Organize Imports" {
		t.Fatalf("expected 'Organize Imports', got %q", item.Title)
	}
}

func TestCodeActionsStateEmpty(t *testing.T) {
	state := &editor.CodeActionsState{
		Items: []editor.CodeActionItem{},
	}
	item := state.SelectedItem()
	if item != nil {
		t.Fatal("expected nil item for empty state")
	}
}

func TestSignatureHelpState(t *testing.T) {
	state := &editor.SignatureHelpState{
		Label: "func Foo(a int, b string) error",
		Parameters: []editor.SignatureParam{
			{Label: "a int", Start: 9, End: 14},
			{Label: "b string", Start: 16, End: 24},
		},
		ActiveParameter: 0,
	}

	if state.Label != "func Foo(a int, b string) error" {
		t.Fatalf("unexpected label: %s", state.Label)
	}
	if len(state.Parameters) != 2 {
		t.Fatalf("expected 2 params, got %d", len(state.Parameters))
	}
	if state.Parameters[0].Label != "a int" {
		t.Fatalf("expected 'a int', got %q", state.Parameters[0].Label)
	}
	if state.ActiveParameter != 0 {
		t.Fatalf("expected active param 0, got %d", state.ActiveParameter)
	}
}

func TestCodeActionDismiss(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)
	RegisterLSPActions(reg)

	buf := editor.NewBuffer("test.go")
	app := &mockAppWithCodeActions{
		mockApp: mockApp{buffer: buf, vpHeight: 24, inputMode: ModeCodeActions},
		codeActions: &editor.CodeActionsState{
			Items: []editor.CodeActionItem{
				{Title: "Test Action", Index: 0},
			},
		},
	}

	ctx := newTestContextCA(app, "code_action.dismiss", nil)
	err := reg.Execute("code_action.dismiss", ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.codeActions != nil {
		t.Fatal("expected code actions to be nil after dismiss")
	}
	if app.inputMode != ModeNormal {
		t.Fatalf("expected ModeNormal, got %d", app.inputMode)
	}
}

// mockAppWithCodeActions extends mockApp to implement CodeActions interface.
type mockAppWithCodeActions struct {
	mockApp
	codeActions *editor.CodeActionsState
	sigHelp     *editor.SignatureHelpState
	appliedIdx  int
}

func (m *mockAppWithCodeActions) CodeActionsState() *editor.CodeActionsState        { return m.codeActions }
func (m *mockAppWithCodeActions) SetCodeActionsState(ca *editor.CodeActionsState)   { m.codeActions = ca }
func (m *mockAppWithCodeActions) ApplyCodeAction(idx int)                           { m.appliedIdx = idx }
func (m *mockAppWithCodeActions) SignatureHelpState() *editor.SignatureHelpState     { return m.sigHelp }
func (m *mockAppWithCodeActions) SetSignatureHelpState(sh *editor.SignatureHelpState) { m.sigHelp = sh }

func newTestContextCA(app *mockAppWithCodeActions, actionID string, payload any) *ActionContext {
	return &ActionContext{
		App: app,
		Event: &events.Event{
			Type:     events.EventAction,
			ActionID: actionID,
			Payload:  payload,
		},
	}
}
