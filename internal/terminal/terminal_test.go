package terminal

import (
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewSession_Creates(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("PTY tests only run on Unix-like systems")
	}

	callback := func(id int) {
		// Output received.
	}

	s, err := NewSession(1, "/bin/sh", 80, 24, os.TempDir(), callback)
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	defer s.Close()

	if s.ID != 1 {
		t.Errorf("expected ID 1, got %d", s.ID)
	}

	if s.IsClosed() {
		t.Error("session should not be closed immediately after creation")
	}
}

func TestSession_WriteAndRead(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("PTY tests only run on Unix-like systems")
	}

	done := make(chan struct{})
	var mu sync.Mutex
	outputCount := 0

	callback := func(id int) {
		mu.Lock()
		outputCount++
		if outputCount > 2 {
			select {
			case done <- struct{}{}:
			default:
			}
		}
		mu.Unlock()
	}

	s, err := NewSession(1, "/bin/sh", 80, 24, os.TempDir(), callback)
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	defer s.Close()

	// Send a simple echo command.
	err = s.WriteString("echo TESTOUTPUT123\n")
	if err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}

	// Wait for output.
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		// Timeout is OK; we'll check what we got.
	}

	lines := s.Lines()
	found := false
	for _, line := range lines {
		if strings.Contains(StripANSI(line), "TESTOUTPUT123") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected output containing 'TESTOUTPUT123', got lines: %v", lines)
	}
}

func TestSession_Resize(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("PTY tests only run on Unix-like systems")
	}

	s, err := NewSession(1, "/bin/sh", 80, 24, os.TempDir(), nil)
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	defer s.Close()

	// Resize should not panic or error.
	s.Resize(120, 40)

	if s.GridRows() != 40 {
		t.Errorf("expected 40 rows after resize, got %d", s.GridRows())
	}
}

func TestSession_Close(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("PTY tests only run on Unix-like systems")
	}

	s, err := NewSession(1, "/bin/sh", 80, 24, os.TempDir(), nil)
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}

	s.Close()

	if !s.IsClosed() {
		t.Error("session should be closed after Close()")
	}

	// Writing to a closed session should fail.
	err = s.WriteString("test")
	if err == nil {
		t.Error("expected error writing to closed session")
	}
}

func TestSession_DisplayName(t *testing.T) {
	s := &Session{Name: "/bin/bash"}
	if s.DisplayName() != "bash" {
		t.Errorf("expected 'bash', got %q", s.DisplayName())
	}

	s2 := &Session{Name: "zsh"}
	if s2.DisplayName() != "zsh" {
		t.Errorf("expected 'zsh', got %q", s2.DisplayName())
	}

	s3 := &Session{Name: ""}
	if s3.DisplayName() != "terminal" {
		t.Errorf("expected 'terminal', got %q", s3.DisplayName())
	}
}

// newTestSession creates a Session with a grid for unit tests (no PTY).
func newTestSession(cols, rows int) *Session {
	s := &Session{
		cols: cols,
		rows: rows,
	}
	s.initGrid()
	return s
}

func TestProcessOutput_Newlines(t *testing.T) {
	s := newTestSession(80, 24)

	s.processBytes([]byte("hello\nworld\n"))

	lines := s.Lines()
	// "hello" should be in scrollback or grid. Check that both appear.
	found := 0
	for i, l := range lines {
		if strings.Contains(l, "hello") || strings.Contains(l, "world") {
			found++
		} else if strings.TrimSpace(l) != "" {
			t.Logf("line[%d] = %q", i, l)
		}
	}
	if found < 2 {
		t.Errorf("expected to find 'hello' and 'world' in lines (found %d): %q", found, lines)
	}
}

func TestProcessOutput_CRLF(t *testing.T) {
	s := newTestSession(80, 24)

	s.processBytes([]byte("hello\r\nworld\r\n"))

	lines := s.Lines()
	found := 0
	for _, l := range lines {
		if strings.Contains(l, "hello") || strings.Contains(l, "world") {
			found++
		}
	}
	if found < 2 {
		t.Errorf("expected to find 'hello' and 'world' in lines (found %d): %q", found, lines)
	}
}

func TestProcessOutput_CarriageReturn(t *testing.T) {
	s := newTestSession(80, 24)

	// Write "loading..." then CR then "Done!" — should overwrite.
	s.processBytes([]byte("loading...\rDone!"))

	row := s.renderGridRow(0)
	if !strings.HasPrefix(row, "Done!") {
		t.Errorf("expected row to start with 'Done!', got %q", row)
	}
}

func TestProcessOutput_Backspace(t *testing.T) {
	s := newTestSession(80, 24)

	// Write "abc" then backspace — cursor should move back.
	s.processBytes([]byte("abc\x08"))

	_, col := s.CursorPos()
	if col != 2 {
		t.Errorf("expected cursor at col 2 after backspace, got %d", col)
	}
}

func TestProcessOutput_CursorMovement(t *testing.T) {
	s := newTestSession(80, 24)

	// Move cursor to row 5, col 10 using CSI H.
	s.processBytes([]byte("\x1b[5;10H"))

	row, col := s.CursorPos()
	if row != 4 || col != 9 { // 0-based
		t.Errorf("expected cursor at (4,9), got (%d,%d)", row, col)
	}
}

func TestProcessOutput_EraseDisplay(t *testing.T) {
	s := newTestSession(80, 24)

	// Write some text then clear screen.
	s.processBytes([]byte("hello world"))
	s.processBytes([]byte("\x1b[2J"))

	row := s.renderGridRow(0)
	if strings.TrimSpace(row) != "" {
		t.Errorf("expected empty grid after erase, got %q", row)
	}
}

func TestProcessOutput_EraseLine(t *testing.T) {
	s := newTestSession(80, 24)

	s.processBytes([]byte("hello world"))
	// Move cursor to col 5, erase from cursor to end.
	s.processBytes([]byte("\x1b[1;6H\x1b[K"))

	row := s.renderGridRow(0)
	if row != "hello" {
		t.Errorf("expected 'hello' after erase to EOL, got %q", row)
	}
}
