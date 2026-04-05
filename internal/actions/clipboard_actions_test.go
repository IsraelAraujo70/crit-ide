package actions

import (
	"testing"

	"github.com/israelcorrea/crit-ide/internal/editor"
)

func TestClipboardCopy_WithSelection(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello world")
	buf.SetSelection(editor.Position{Line: 0, Col: 0}, editor.Position{Line: 0, Col: 5})
	clip := &mockClipboard{}
	app := &mockApp{buffer: buf, vpHeight: 24, clipboard: clip}

	reg.Execute("clipboard.copy", newTestContext(app, "clipboard.copy", nil))

	if clip.content != "hello" {
		t.Fatalf("expected clipboard %q, got %q", "hello", clip.content)
	}
	// Selection should still be active (copy does not clear).
	if !buf.HasSelection() {
		t.Fatal("selection should remain after copy")
	}
}

func TestClipboardCopy_NoSelection(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello")
	clip := &mockClipboard{content: "old"}
	app := &mockApp{buffer: buf, vpHeight: 24, clipboard: clip}

	reg.Execute("clipboard.copy", newTestContext(app, "clipboard.copy", nil))

	if clip.content != "old" {
		t.Fatalf("expected clipboard unchanged, got %q", clip.content)
	}
}

func TestClipboardCut_WithSelection(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello world")
	buf.SetSelection(editor.Position{Line: 0, Col: 0}, editor.Position{Line: 0, Col: 5})
	clip := &mockClipboard{}
	app := &mockApp{buffer: buf, vpHeight: 24, clipboard: clip}

	reg.Execute("clipboard.cut", newTestContext(app, "clipboard.cut", nil))

	if clip.content != "hello" {
		t.Fatalf("expected clipboard %q, got %q", "hello", clip.content)
	}
	if buf.Text.Line(0) != " world" {
		t.Fatalf("expected %q, got %q", " world", buf.Text.Line(0))
	}
	if buf.HasSelection() {
		t.Fatal("selection should be cleared after cut")
	}
}

func TestClipboardPaste_NoSelection(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello")
	buf.SetCursorPos(0, 5)
	clip := &mockClipboard{content: " world"}
	app := &mockApp{buffer: buf, vpHeight: 24, clipboard: clip}

	reg.Execute("clipboard.paste", newTestContext(app, "clipboard.paste", nil))

	if buf.Text.Line(0) != "hello world" {
		t.Fatalf("expected %q, got %q", "hello world", buf.Text.Line(0))
	}
}

func TestClipboardPaste_WithSelection(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello world")
	buf.SetSelection(editor.Position{Line: 0, Col: 0}, editor.Position{Line: 0, Col: 5})
	clip := &mockClipboard{content: "hi"}
	app := &mockApp{buffer: buf, vpHeight: 24, clipboard: clip}

	reg.Execute("clipboard.paste", newTestContext(app, "clipboard.paste", nil))

	if buf.Text.Line(0) != "hi world" {
		t.Fatalf("expected %q, got %q", "hi world", buf.Text.Line(0))
	}
}

func TestClipboardPaste_EmptyClipboard(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("hello")
	clip := &mockClipboard{content: ""}
	app := &mockApp{buffer: buf, vpHeight: 24, clipboard: clip}

	reg.Execute("clipboard.paste", newTestContext(app, "clipboard.paste", nil))

	if buf.Text.Line(0) != "hello" {
		t.Fatalf("expected %q unchanged, got %q", "hello", buf.Text.Line(0))
	}
}

func TestClipboardPaste_MultiLine(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	buf := newBufferWithLines("ac")
	buf.SetCursorPos(0, 1)
	clip := &mockClipboard{content: "b\nd"}
	app := &mockApp{buffer: buf, vpHeight: 24, clipboard: clip}

	reg.Execute("clipboard.paste", newTestContext(app, "clipboard.paste", nil))

	if buf.Text.LineCount() != 2 {
		t.Fatalf("expected 2 lines, got %d", buf.Text.LineCount())
	}
	if buf.Text.Line(0) != "ab" {
		t.Fatalf("line 0: expected %q, got %q", "ab", buf.Text.Line(0))
	}
	if buf.Text.Line(1) != "dc" {
		t.Fatalf("line 1: expected %q, got %q", "dc", buf.Text.Line(1))
	}
}
