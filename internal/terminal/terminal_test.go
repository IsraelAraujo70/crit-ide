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

func TestSession_LineCount(t *testing.T) {
	s := &Session{
		lines:   []string{"line1", "line2", "line3"},
		current: "partial",
	}
	if s.LineCount() != 4 {
		t.Errorf("expected 4 lines (3 + current), got %d", s.LineCount())
	}

	s2 := &Session{
		lines:   []string{"line1"},
		current: "",
	}
	if s2.LineCount() != 1 {
		t.Errorf("expected 1 line, got %d", s2.LineCount())
	}
}

func TestProcessOutput(t *testing.T) {
	s := &Session{
		lines: make([]string, 0),
	}

	// Process output with newlines.
	s.processOutput("hello\nworld\n")

	if len(s.lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(s.lines), s.lines)
	}
	if s.lines[0] != "hello" {
		t.Errorf("expected first line 'hello', got %q", s.lines[0])
	}
	if s.lines[1] != "world" {
		t.Errorf("expected second line 'world', got %q", s.lines[1])
	}
	if s.current != "" {
		t.Errorf("expected empty current line, got %q", s.current)
	}
}

func TestProcessOutput_CRLF(t *testing.T) {
	s := &Session{
		lines: make([]string, 0),
	}

	s.processOutput("hello\r\nworld\r\n")

	if len(s.lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(s.lines), s.lines)
	}
	if s.lines[0] != "hello" {
		t.Errorf("expected first line 'hello', got %q", s.lines[0])
	}
}

func TestProcessOutput_CarriageReturn(t *testing.T) {
	s := &Session{
		lines: make([]string, 0),
	}

	// CR without LF: overwrite current line.
	s.processOutput("loading...\rDone!     ")

	if s.current != "Done!     " {
		t.Errorf("expected current line 'Done!     ', got %q", s.current)
	}
}
