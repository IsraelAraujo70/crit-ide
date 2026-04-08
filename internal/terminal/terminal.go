package terminal

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
)

const (
	// maxScrollback is the maximum number of lines kept in the terminal buffer.
	maxScrollback = 10000
)

// Session represents a single terminal PTY session.
type Session struct {
	ID   int
	Name string

	cmd  *exec.Cmd
	ptmx *os.File

	mu      sync.Mutex
	lines   []string // Ring buffer of output lines.
	current string   // Current (incomplete) line being built.
	closed  bool
	cols    int
	rows    int
}

// OutputCallback is called when new output is available from the terminal.
type OutputCallback func(sessionID int)

// NewSession creates a new terminal session with a PTY.
func NewSession(id int, shell string, cols, rows int, workDir string, onOutput OutputCallback) (*Session, error) {
	if shell == "" {
		shell = os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/bash"
		}
	}

	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	if workDir != "" {
		cmd.Dir = workDir
	}

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
	if err != nil {
		return nil, err
	}

	s := &Session{
		ID:   id,
		Name: shell,
		cmd:  cmd,
		ptmx: ptmx,
		lines: make([]string, 0, 256),
		cols:  cols,
		rows:  rows,
	}

	// Start reader goroutine.
	go s.readLoop(onOutput)

	return s, nil
}

// readLoop reads output from the PTY and buffers it into lines.
func (s *Session) readLoop(onOutput OutputCallback) {
	buf := make([]byte, 4096)
	for {
		n, err := s.ptmx.Read(buf)
		if n > 0 {
			s.processOutput(string(buf[:n]))
			if onOutput != nil {
				onOutput(s.ID)
			}
		}
		if err != nil {
			if err != io.EOF {
				// PTY closed or error.
			}
			s.mu.Lock()
			s.closed = true
			// Flush any remaining current line.
			if s.current != "" {
				s.lines = append(s.lines, s.current)
				s.current = ""
				s.trimLines()
			}
			s.mu.Unlock()
			if onOutput != nil {
				onOutput(s.ID)
			}
			return
		}
	}
}

// processOutput processes raw PTY output, handling \r\n, \r, and \n.
func (s *Session) processOutput(data string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := 0; i < len(data); i++ {
		ch := data[i]
		switch ch {
		case '\n':
			s.lines = append(s.lines, s.current)
			s.current = ""
			s.trimLines()
		case '\r':
			// Carriage return: reset to beginning of current line.
			// If followed by \n, treat as newline pair.
			if i+1 < len(data) && data[i+1] == '\n' {
				s.lines = append(s.lines, s.current)
				s.current = ""
				s.trimLines()
				i++ // skip the \n
			} else {
				// Just CR: overwrite current line from the beginning.
				s.current = ""
			}
		case '\x07': // BEL - ignore
		case '\x08': // Backspace
			stripped := StripANSI(s.current)
			if len(stripped) > 0 {
				// Simple backspace: remove last visible character.
				// This is a simplification; full terminal emulation would be more complex.
			}
		default:
			s.current += string(ch)
		}
	}
}

// trimLines ensures we don't exceed maxScrollback.
func (s *Session) trimLines() {
	if len(s.lines) > maxScrollback {
		excess := len(s.lines) - maxScrollback
		s.lines = s.lines[excess:]
	}
}

// Write sends input to the PTY.
func (s *Session) Write(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return io.ErrClosedPipe
	}
	_, err := s.ptmx.Write(data)
	return err
}

// WriteString sends a string to the PTY.
func (s *Session) WriteString(str string) error {
	return s.Write([]byte(str))
}

// Lines returns a copy of the terminal output lines (including the current incomplete line).
func (s *Session) Lines() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]string, len(s.lines), len(s.lines)+1)
	copy(result, s.lines)
	if s.current != "" {
		result = append(result, s.current)
	}
	return result
}

// LineCount returns the total number of lines (including current).
func (s *Session) LineCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := len(s.lines)
	if s.current != "" {
		n++
	}
	return n
}

// IsClosed returns true if the PTY session has ended.
func (s *Session) IsClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

// Resize changes the PTY window size.
func (s *Session) Resize(cols, rows int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.cols = cols
	s.rows = rows
	// Use TIOCSWINSZ ioctl to set window size.
	ws := struct {
		Rows uint16
		Cols uint16
		X    uint16
		Y    uint16
	}{
		Rows: uint16(rows),
		Cols: uint16(cols),
	}
	syscall.Syscall(syscall.SYS_IOCTL,
		s.ptmx.Fd(),
		uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&ws)),
	)
}

// SendSignal sends an OS signal to the terminal process.
func (s *Session) SendSignal(sig os.Signal) error {
	if s.cmd.Process == nil {
		return nil
	}
	return s.cmd.Process.Signal(sig)
}

// Close terminates the session and cleans up resources.
func (s *Session) Close() {
	s.mu.Lock()
	alreadyClosed := s.closed
	s.closed = true
	s.mu.Unlock()

	if !alreadyClosed {
		// Send SIGHUP then SIGKILL if process is still alive.
		if s.cmd.Process != nil {
			_ = s.cmd.Process.Signal(syscall.SIGHUP)
		}
	}
	if s.ptmx != nil {
		_ = s.ptmx.Close()
	}
	// Wait for process to exit.
	if s.cmd.Process != nil {
		_ = s.cmd.Wait()
	}
}

// DisplayName returns the session display name.
func (s *Session) DisplayName() string {
	if s.Name != "" {
		parts := strings.Split(s.Name, "/")
		return parts[len(parts)-1]
	}
	return "terminal"
}
