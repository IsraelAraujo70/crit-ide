package actions

import (
	"testing"

	"github.com/israelcorrea/crit-ide/internal/editor"
)

func TestGotoLineOpensPrompt(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := editor.NewBuffer("test.go")
	app := &mockApp{buffer: buf, vpHeight: 24}

	ctx := newTestContext(app, "goto.line", nil)
	err := reg.Execute("goto.line", ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if app.inputMode != ModePrompt {
		t.Fatalf("expected ModePrompt, got %d", app.inputMode)
	}
	if app.prompt == nil {
		t.Fatal("expected prompt to be set")
	}
	if app.prompt.Kind != editor.PromptGotoLine {
		t.Fatalf("expected PromptGotoLine, got %d", app.prompt.Kind)
	}
	if app.prompt.Label != "Go to line: " {
		t.Fatalf("expected label %q, got %q", "Go to line: ", app.prompt.Label)
	}
}

func TestPromptConfirmGotoLine(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	// Create buffer with 10 lines.
	buf := editor.NewBuffer("test.go")
	for i := 0; i < 9; i++ {
		buf.InsertChar(rune('a' + i))
		buf.InsertNewline()
	}
	buf.InsertChar('j')
	// Buffer has 10 lines. Cursor is at last line.

	app := &mockApp{buffer: buf, vpHeight: 24}

	// Simulate: user typed "5" in the prompt.
	app.prompt = &editor.PromptState{
		Kind:      editor.PromptGotoLine,
		Label:     "Go to line: ",
		Input:     "5",
		CursorPos: 1,
	}
	app.inputMode = ModePrompt

	ctx := newTestContext(app, "prompt.confirm", nil)
	err := reg.Execute("prompt.confirm", ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Line 5 (1-indexed) = row 4 (0-indexed).
	if buf.CursorRow != 4 {
		t.Fatalf("expected CursorRow 4, got %d", buf.CursorRow)
	}
	if buf.CursorCol != 0 {
		t.Fatalf("expected CursorCol 0, got %d", buf.CursorCol)
	}
	if app.inputMode != ModeNormal {
		t.Fatalf("expected ModeNormal after confirm, got %d", app.inputMode)
	}
	if app.prompt != nil {
		t.Fatal("expected prompt to be nil after confirm")
	}
}

func TestPromptConfirmGotoLineClampHigh(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	// Buffer with 5 lines.
	buf := editor.NewBuffer("test.go")
	for i := 0; i < 4; i++ {
		buf.InsertChar(rune('a' + i))
		buf.InsertNewline()
	}
	buf.InsertChar('e')
	buf.CursorRow = 0
	buf.CursorCol = 0

	app := &mockApp{buffer: buf, vpHeight: 24}
	app.prompt = &editor.PromptState{
		Kind:  editor.PromptGotoLine,
		Label: "Go to line: ",
		Input: "999",
	}
	app.inputMode = ModePrompt

	ctx := newTestContext(app, "prompt.confirm", nil)
	_ = reg.Execute("prompt.confirm", ctx)

	// Should clamp to last line (row 4).
	if buf.CursorRow != 4 {
		t.Fatalf("expected CursorRow 4 (clamped), got %d", buf.CursorRow)
	}
}

func TestPromptConfirmGotoLineClampLow(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := editor.NewBuffer("test.go")
	buf.InsertChar('a')
	buf.InsertNewline()
	buf.InsertChar('b')

	app := &mockApp{buffer: buf, vpHeight: 24}
	app.prompt = &editor.PromptState{
		Kind:  editor.PromptGotoLine,
		Label: "Go to line: ",
		Input: "0",
	}
	app.inputMode = ModePrompt

	ctx := newTestContext(app, "prompt.confirm", nil)
	_ = reg.Execute("prompt.confirm", ctx)

	// Line 0 → clamped to row 0.
	if buf.CursorRow != 0 {
		t.Fatalf("expected CursorRow 0 (clamped), got %d", buf.CursorRow)
	}
}

func TestPromptConfirmGotoLineNonNumeric(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := editor.NewBuffer("test.go")
	buf.InsertChar('a')
	buf.CursorRow = 0
	originalRow := buf.CursorRow

	app := &mockApp{buffer: buf, vpHeight: 24}
	app.prompt = &editor.PromptState{
		Kind:  editor.PromptGotoLine,
		Label: "Go to line: ",
		Input: "abc",
	}
	app.inputMode = ModePrompt

	ctx := newTestContext(app, "prompt.confirm", nil)
	_ = reg.Execute("prompt.confirm", ctx)

	// Non-numeric input should not move cursor.
	if buf.CursorRow != originalRow {
		t.Fatalf("expected cursor to stay at row %d, got %d", originalRow, buf.CursorRow)
	}
	// But prompt should still close.
	if app.inputMode != ModeNormal {
		t.Fatalf("expected ModeNormal after non-numeric confirm, got %d", app.inputMode)
	}
	if app.prompt != nil {
		t.Fatal("expected prompt to be nil after confirm")
	}
}

func TestPromptConfirmGotoLineEmpty(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := editor.NewBuffer("test.go")
	buf.InsertChar('a')

	app := &mockApp{buffer: buf, vpHeight: 24}
	app.prompt = &editor.PromptState{
		Kind:  editor.PromptGotoLine,
		Label: "Go to line: ",
		Input: "",
	}
	app.inputMode = ModePrompt

	ctx := newTestContext(app, "prompt.confirm", nil)
	_ = reg.Execute("prompt.confirm", ctx)

	// Empty input: prompt closes, cursor stays.
	if app.inputMode != ModeNormal {
		t.Fatalf("expected ModeNormal, got %d", app.inputMode)
	}
}

func TestPromptConfirmGotoLineNegative(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := editor.NewBuffer("test.go")
	buf.InsertChar('a')
	buf.InsertNewline()
	buf.InsertChar('b')
	buf.CursorRow = 1

	app := &mockApp{buffer: buf, vpHeight: 24}
	app.prompt = &editor.PromptState{
		Kind:  editor.PromptGotoLine,
		Label: "Go to line: ",
		Input: "-5",
	}
	app.inputMode = ModePrompt

	ctx := newTestContext(app, "prompt.confirm", nil)
	_ = reg.Execute("prompt.confirm", ctx)

	// Negative → clamped to row 0.
	if buf.CursorRow != 0 {
		t.Fatalf("expected CursorRow 0 (clamped from negative), got %d", buf.CursorRow)
	}
}
