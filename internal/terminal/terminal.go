package terminal

import (
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
)

const (
	maxScrollback = 5000
)

// cell represents one character in the terminal grid.
type cell struct {
	ch    rune
	style int32 // packed ANSI style state index
}

// Session represents a single terminal PTY session with a proper VT100 grid.
type Session struct {
	ID   int
	Name string

	cmd  *exec.Cmd
	ptmx *os.File

	mu     sync.Mutex
	cols   int
	rows   int
	closed bool

	// Screen grid: rows x cols cells.
	grid [][]cell
	// Scrollback buffer of completed lines (styled strings for rendering).
	scrollback []string

	// Cursor state.
	curRow int
	curCol int

	// ANSI parser state.
	ansiState  ansiParseState
	ansiBuf    []byte
	styleStack string // current ANSI SGR prefix to prepend on rendered lines

	// Saved cursor position (ESC 7 / ESC 8).
	savedRow int
	savedCol int
}

type ansiParseState int

const (
	stateNormal ansiParseState = iota
	stateESC                   // got ESC
	stateCSI                   // got ESC [
	stateOSC                   // got ESC ]
)

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
		cols: cols,
		rows: rows,
	}
	s.initGrid()

	go s.readLoop(onOutput)

	return s, nil
}

func (s *Session) initGrid() {
	s.grid = make([][]cell, s.rows)
	for i := range s.grid {
		s.grid[i] = make([]cell, s.cols)
		for j := range s.grid[i] {
			s.grid[i][j] = cell{ch: ' '}
		}
	}
	s.scrollback = make([]string, 0, 256)
}

// readLoop reads output from the PTY and processes it.
func (s *Session) readLoop(onOutput OutputCallback) {
	buf := make([]byte, 8192)
	for {
		n, err := s.ptmx.Read(buf)
		if n > 0 {
			s.mu.Lock()
			s.processBytes(buf[:n])
			s.mu.Unlock()
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
			s.mu.Unlock()
			if onOutput != nil {
				onOutput(s.ID)
			}
			return
		}
	}
}

// processBytes runs a VT100 state machine over raw PTY output.
func (s *Session) processBytes(data []byte) {
	for _, b := range data {
		switch s.ansiState {
		case stateNormal:
			s.processNormal(b)
		case stateESC:
			s.processESC(b)
		case stateCSI:
			s.processCSI(b)
		case stateOSC:
			// Consume until BEL or ST (ESC \).
			if b == '\x07' || b == '\\' {
				s.ansiState = stateNormal
				s.ansiBuf = s.ansiBuf[:0]
			} else {
				s.ansiBuf = append(s.ansiBuf, b)
			}
		}
	}
}

func (s *Session) processNormal(b byte) {
	switch b {
	case '\x1b':
		s.ansiState = stateESC
		s.ansiBuf = s.ansiBuf[:0]
	case '\n':
		s.linefeed()
	case '\r':
		s.curCol = 0
	case '\x08': // Backspace
		if s.curCol > 0 {
			s.curCol--
		}
	case '\x07': // BEL - ignore
	case '\t':
		// Tab: advance to next 8-col boundary.
		next := ((s.curCol / 8) + 1) * 8
		if next >= s.cols {
			next = s.cols - 1
		}
		s.curCol = next
	default:
		if b >= 0x20 {
			s.putChar(rune(b))
		}
		// Ignore other C0 control chars.
	}
}

func (s *Session) processESC(b byte) {
	switch b {
	case '[':
		s.ansiState = stateCSI
		s.ansiBuf = s.ansiBuf[:0]
	case ']':
		s.ansiState = stateOSC
		s.ansiBuf = s.ansiBuf[:0]
	case '7': // Save cursor.
		s.savedRow = s.curRow
		s.savedCol = s.curCol
		s.ansiState = stateNormal
	case '8': // Restore cursor.
		s.curRow = s.savedRow
		s.curCol = s.savedCol
		s.ansiState = stateNormal
	case 'M': // Reverse index (scroll down).
		if s.curRow == 0 {
			s.scrollDown()
		} else {
			s.curRow--
		}
		s.ansiState = stateNormal
	case 'D': // Index (same as linefeed).
		s.linefeed()
		s.ansiState = stateNormal
	case 'E': // Next line.
		s.curCol = 0
		s.linefeed()
		s.ansiState = stateNormal
	case '(', ')': // Character set selection — eat next byte.
		s.ansiState = stateNormal
		// We ignore character set selection entirely.
	default:
		s.ansiState = stateNormal
	}
}

func (s *Session) processCSI(b byte) {
	// Accumulate parameter bytes.
	if (b >= '0' && b <= '9') || b == ';' || b == '?' {
		s.ansiBuf = append(s.ansiBuf, b)
		return
	}

	// b is the command character.
	params := string(s.ansiBuf)
	s.ansiState = stateNormal

	switch b {
	case 'm': // SGR - Select Graphic Rendition.
		s.styleStack = "\x1b[" + params + "m"
	case 'A': // Cursor Up.
		n := parseParam(params, 1)
		s.curRow -= n
		if s.curRow < 0 {
			s.curRow = 0
		}
	case 'B': // Cursor Down.
		n := parseParam(params, 1)
		s.curRow += n
		if s.curRow >= s.rows {
			s.curRow = s.rows - 1
		}
	case 'C': // Cursor Forward.
		n := parseParam(params, 1)
		s.curCol += n
		if s.curCol >= s.cols {
			s.curCol = s.cols - 1
		}
	case 'D': // Cursor Backward.
		n := parseParam(params, 1)
		s.curCol -= n
		if s.curCol < 0 {
			s.curCol = 0
		}
	case 'H', 'f': // Cursor Position.
		row, col := parseTwoParams(params, 1, 1)
		s.curRow = row - 1
		s.curCol = col - 1
		s.clampCursor()
	case 'J': // Erase in Display.
		n := parseParam(params, 0)
		switch n {
		case 0: // Erase from cursor to end.
			s.clearRow(s.curRow, s.curCol, s.cols)
			for r := s.curRow + 1; r < s.rows; r++ {
				s.clearRow(r, 0, s.cols)
			}
		case 1: // Erase from start to cursor.
			for r := 0; r < s.curRow; r++ {
				s.clearRow(r, 0, s.cols)
			}
			s.clearRow(s.curRow, 0, s.curCol+1)
		case 2, 3: // Erase entire display.
			for r := 0; r < s.rows; r++ {
				s.clearRow(r, 0, s.cols)
			}
		}
	case 'K': // Erase in Line.
		n := parseParam(params, 0)
		switch n {
		case 0: // Erase from cursor to end of line.
			s.clearRow(s.curRow, s.curCol, s.cols)
		case 1: // Erase from start of line to cursor.
			s.clearRow(s.curRow, 0, s.curCol+1)
		case 2: // Erase entire line.
			s.clearRow(s.curRow, 0, s.cols)
		}
	case 'L': // Insert lines.
		n := parseParam(params, 1)
		s.insertLines(n)
	case 'M': // Delete lines.
		n := parseParam(params, 1)
		s.deleteLines(n)
	case 'P': // Delete characters.
		n := parseParam(params, 1)
		s.deleteChars(n)
	case '@': // Insert characters.
		n := parseParam(params, 1)
		s.insertChars(n)
	case 'G': // Cursor Horizontal Absolute.
		n := parseParam(params, 1)
		s.curCol = n - 1
		if s.curCol < 0 {
			s.curCol = 0
		}
		if s.curCol >= s.cols {
			s.curCol = s.cols - 1
		}
	case 'd': // Cursor Vertical Absolute.
		n := parseParam(params, 1)
		s.curRow = n - 1
		s.clampCursor()
	case 'r': // Set scroll region (ignored, assume full screen).
	case 'h', 'l': // Set/Reset mode (ignored).
	case 'n': // Device Status Report.
		if params == "6" {
			// Report cursor position.
			resp := "\x1b[" + strconv.Itoa(s.curRow+1) + ";" + strconv.Itoa(s.curCol+1) + "R"
			s.ptmx.Write([]byte(resp))
		}
	case 'X': // Erase characters.
		n := parseParam(params, 1)
		for i := 0; i < n && s.curCol+i < s.cols; i++ {
			s.grid[s.curRow][s.curCol+i] = cell{ch: ' '}
		}
	case 'S': // Scroll Up.
		n := parseParam(params, 1)
		for i := 0; i < n; i++ {
			s.scrollUp()
		}
	case 'T': // Scroll Down.
		n := parseParam(params, 1)
		for i := 0; i < n; i++ {
			s.scrollDown()
		}
	}
}

// putChar writes a character at the current cursor position.
func (s *Session) putChar(ch rune) {
	if s.curCol >= s.cols {
		// Auto-wrap.
		s.curCol = 0
		s.linefeed()
	}
	if s.curRow >= 0 && s.curRow < s.rows && s.curCol >= 0 && s.curCol < s.cols {
		s.grid[s.curRow][s.curCol] = cell{ch: ch}
	}
	s.curCol++
}

// linefeed moves cursor down; if at bottom, scroll the grid up.
func (s *Session) linefeed() {
	if s.curRow >= s.rows-1 {
		s.scrollUp()
	} else {
		s.curRow++
	}
}

// scrollUp scrolls the grid up by one line, pushing line 0 to scrollback.
func (s *Session) scrollUp() {
	// Save top line to scrollback.
	s.scrollback = append(s.scrollback, s.renderGridRow(0))
	if len(s.scrollback) > maxScrollback {
		s.scrollback = s.scrollback[1:]
	}
	// Shift grid up.
	for r := 0; r < s.rows-1; r++ {
		s.grid[r] = s.grid[r+1]
	}
	// Clear bottom row.
	s.grid[s.rows-1] = make([]cell, s.cols)
	for j := range s.grid[s.rows-1] {
		s.grid[s.rows-1][j] = cell{ch: ' '}
	}
}

// scrollDown scrolls the grid down by one line (inserts blank at top).
func (s *Session) scrollDown() {
	for r := s.rows - 1; r > 0; r-- {
		s.grid[r] = s.grid[r-1]
	}
	s.grid[0] = make([]cell, s.cols)
	for j := range s.grid[0] {
		s.grid[0][j] = cell{ch: ' '}
	}
}

func (s *Session) clearRow(row, fromCol, toCol int) {
	if row < 0 || row >= s.rows {
		return
	}
	for c := fromCol; c < toCol && c < s.cols; c++ {
		s.grid[row][c] = cell{ch: ' '}
	}
}

func (s *Session) insertLines(n int) {
	for i := 0; i < n; i++ {
		for r := s.rows - 1; r > s.curRow; r-- {
			s.grid[r] = s.grid[r-1]
		}
		s.grid[s.curRow] = make([]cell, s.cols)
		for j := range s.grid[s.curRow] {
			s.grid[s.curRow][j] = cell{ch: ' '}
		}
	}
}

func (s *Session) deleteLines(n int) {
	for i := 0; i < n; i++ {
		for r := s.curRow; r < s.rows-1; r++ {
			s.grid[r] = s.grid[r+1]
		}
		s.grid[s.rows-1] = make([]cell, s.cols)
		for j := range s.grid[s.rows-1] {
			s.grid[s.rows-1][j] = cell{ch: ' '}
		}
	}
}

func (s *Session) deleteChars(n int) {
	row := s.grid[s.curRow]
	for i := s.curCol; i < s.cols; i++ {
		if i+n < s.cols {
			row[i] = row[i+n]
		} else {
			row[i] = cell{ch: ' '}
		}
	}
}

func (s *Session) insertChars(n int) {
	row := s.grid[s.curRow]
	for i := s.cols - 1; i >= s.curCol+n; i-- {
		row[i] = row[i-n]
	}
	for i := s.curCol; i < s.curCol+n && i < s.cols; i++ {
		row[i] = cell{ch: ' '}
	}
}

func (s *Session) clampCursor() {
	if s.curRow < 0 {
		s.curRow = 0
	}
	if s.curRow >= s.rows {
		s.curRow = s.rows - 1
	}
	if s.curCol < 0 {
		s.curCol = 0
	}
	if s.curCol >= s.cols {
		s.curCol = s.cols - 1
	}
}

// renderGridRow converts a grid row to a string, trimming trailing spaces.
func (s *Session) renderGridRow(row int) string {
	if row < 0 || row >= s.rows {
		return ""
	}
	var sb strings.Builder
	sb.Grow(s.cols)
	lastNonSpace := -1
	for c := 0; c < s.cols; c++ {
		ch := s.grid[row][c].ch
		if ch == 0 {
			ch = ' '
		}
		if ch != ' ' {
			lastNonSpace = c
		}
	}
	for c := 0; c <= lastNonSpace; c++ {
		ch := s.grid[row][c].ch
		if ch == 0 {
			ch = ' '
		}
		sb.WriteRune(ch)
	}
	return sb.String()
}

// Lines returns the visible terminal content: scrollback + current grid rows.
func (s *Session) Lines() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]string, 0, len(s.scrollback)+s.rows)
	result = append(result, s.scrollback...)
	for r := 0; r < s.rows; r++ {
		result = append(result, s.renderGridRow(r))
	}
	// Trim trailing empty lines from grid portion.
	for len(result) > len(s.scrollback) && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}
	return result
}

// CursorPos returns the cursor row relative to the visible grid (0-based)
// and the cursor column.
func (s *Session) CursorPos() (row, col int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.curRow, s.curCol
}

// GridRows returns the number of rows in the terminal grid.
func (s *Session) GridRows() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rows
}

// LineCount returns the total number of lines (scrollback + visible grid).
func (s *Session) LineCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.scrollback) + s.rows
}

// IsClosed returns true if the PTY session has ended.
func (s *Session) IsClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
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

// Resize changes the PTY window size and resizes the grid.
func (s *Session) Resize(cols, rows int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || (cols == s.cols && rows == s.rows) {
		return
	}

	// Resize grid.
	newGrid := make([][]cell, rows)
	for r := range newGrid {
		newGrid[r] = make([]cell, cols)
		for c := range newGrid[r] {
			newGrid[r][c] = cell{ch: ' '}
		}
	}
	// Copy what fits from old grid.
	for r := 0; r < rows && r < s.rows; r++ {
		for c := 0; c < cols && c < s.cols; c++ {
			newGrid[r][c] = s.grid[r][c]
		}
	}
	s.grid = newGrid
	s.cols = cols
	s.rows = rows
	s.clampCursor()

	// Notify PTY of new size.
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
		if s.cmd.Process != nil {
			_ = s.cmd.Process.Signal(syscall.SIGHUP)
		}
	}
	if s.ptmx != nil {
		_ = s.ptmx.Close()
	}
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

// --- helpers ---

func parseParam(s string, def int) int {
	s = strings.TrimPrefix(s, "?")
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func parseTwoParams(s string, def1, def2 int) (int, int) {
	s = strings.TrimPrefix(s, "?")
	parts := strings.SplitN(s, ";", 2)
	a := def1
	b := def2
	if len(parts) >= 1 && parts[0] != "" {
		if n, err := strconv.Atoi(parts[0]); err == nil {
			a = n
		}
	}
	if len(parts) >= 2 && parts[1] != "" {
		if n, err := strconv.Atoi(parts[1]); err == nil {
			b = n
		}
	}
	return a, b
}
