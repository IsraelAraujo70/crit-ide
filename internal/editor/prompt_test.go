package editor

import "testing"

func TestPromptInsertChar(t *testing.T) {
	p := &PromptState{}

	p.InsertChar('h')
	p.InsertChar('i')

	if p.Input != "hi" {
		t.Fatalf("expected %q, got %q", "hi", p.Input)
	}
	if p.CursorPos != 2 {
		t.Fatalf("expected cursor at 2, got %d", p.CursorPos)
	}
}

func TestPromptInsertCharMiddle(t *testing.T) {
	p := &PromptState{Input: "hllo", CursorPos: 1}

	p.InsertChar('e')

	if p.Input != "hello" {
		t.Fatalf("expected %q, got %q", "hello", p.Input)
	}
	if p.CursorPos != 2 {
		t.Fatalf("expected cursor at 2, got %d", p.CursorPos)
	}
}

func TestPromptDeleteBackward(t *testing.T) {
	p := &PromptState{Input: "hello", CursorPos: 5}

	p.DeleteBackward()

	if p.Input != "hell" {
		t.Fatalf("expected %q, got %q", "hell", p.Input)
	}
	if p.CursorPos != 4 {
		t.Fatalf("expected cursor at 4, got %d", p.CursorPos)
	}
}

func TestPromptDeleteBackwardAtStart(t *testing.T) {
	p := &PromptState{Input: "hello", CursorPos: 0}

	p.DeleteBackward()

	if p.Input != "hello" {
		t.Fatalf("expected %q unchanged, got %q", "hello", p.Input)
	}
}

func TestPromptDeleteForward(t *testing.T) {
	p := &PromptState{Input: "hello", CursorPos: 0}

	p.DeleteForward()

	if p.Input != "ello" {
		t.Fatalf("expected %q, got %q", "ello", p.Input)
	}
}

func TestPromptDeleteForwardAtEnd(t *testing.T) {
	p := &PromptState{Input: "hello", CursorPos: 5}

	p.DeleteForward()

	if p.Input != "hello" {
		t.Fatalf("expected %q unchanged, got %q", "hello", p.Input)
	}
}

func TestPromptMoveLeftRight(t *testing.T) {
	p := &PromptState{Input: "abc", CursorPos: 3}

	p.MoveLeft()
	if p.CursorPos != 2 {
		t.Fatalf("expected cursor at 2, got %d", p.CursorPos)
	}

	p.MoveRight()
	if p.CursorPos != 3 {
		t.Fatalf("expected cursor at 3, got %d", p.CursorPos)
	}
}

func TestPromptMoveLeftAtStart(t *testing.T) {
	p := &PromptState{Input: "abc", CursorPos: 0}

	p.MoveLeft()
	if p.CursorPos != 0 {
		t.Fatalf("expected cursor at 0, got %d", p.CursorPos)
	}
}

func TestPromptMoveRightAtEnd(t *testing.T) {
	p := &PromptState{Input: "abc", CursorPos: 3}

	p.MoveRight()
	if p.CursorPos != 3 {
		t.Fatalf("expected cursor at 3, got %d", p.CursorPos)
	}
}

func TestPromptHomeEnd(t *testing.T) {
	p := &PromptState{Input: "hello", CursorPos: 3}

	p.MoveHome()
	if p.CursorPos != 0 {
		t.Fatalf("expected cursor at 0, got %d", p.CursorPos)
	}

	p.MoveEnd()
	if p.CursorPos != 5 {
		t.Fatalf("expected cursor at 5, got %d", p.CursorPos)
	}
}
