package editor

// PromptKind indicates what operation the prompt is for.
type PromptKind int

const (
	PromptNewFile  PromptKind = iota // Create a new file or directory.
	PromptRename                     // Rename a file or directory.
	PromptDelete                     // Confirm deletion (y/n).
	PromptGotoLine                   // Go to a specific line number.
	PromptGitCommit                  // Git commit message.
)

// PromptState holds the state of an interactive input prompt.
type PromptState struct {
	Kind      PromptKind
	Label     string // Display label (e.g., "New file: ").
	Input     string // Current user input.
	CursorPos int    // Byte offset cursor within Input.
	Context   string // Extra context (e.g., path being renamed/deleted).
}

// InsertChar inserts a character at the cursor position.
func (p *PromptState) InsertChar(ch rune) {
	s := string(ch)
	p.Input = p.Input[:p.CursorPos] + s + p.Input[p.CursorPos:]
	p.CursorPos += len(s)
}

// DeleteBackward removes the character before the cursor.
func (p *PromptState) DeleteBackward() {
	if p.CursorPos > 0 {
		// Find previous rune boundary.
		prev := p.CursorPos - 1
		for prev > 0 && p.Input[prev]>>6 == 2 {
			prev-- // Skip UTF-8 continuation bytes.
		}
		p.Input = p.Input[:prev] + p.Input[p.CursorPos:]
		p.CursorPos = prev
	}
}

// DeleteForward removes the character at the cursor.
func (p *PromptState) DeleteForward() {
	if p.CursorPos < len(p.Input) {
		next := p.CursorPos + 1
		for next < len(p.Input) && p.Input[next]>>6 == 2 {
			next++
		}
		p.Input = p.Input[:p.CursorPos] + p.Input[next:]
	}
}

// MoveLeft moves the cursor left one character.
func (p *PromptState) MoveLeft() {
	if p.CursorPos > 0 {
		p.CursorPos--
		for p.CursorPos > 0 && p.Input[p.CursorPos]>>6 == 2 {
			p.CursorPos--
		}
	}
}

// MoveRight moves the cursor right one character.
func (p *PromptState) MoveRight() {
	if p.CursorPos < len(p.Input) {
		p.CursorPos++
		for p.CursorPos < len(p.Input) && p.Input[p.CursorPos]>>6 == 2 {
			p.CursorPos++
		}
	}
}

// MoveHome moves the cursor to the start.
func (p *PromptState) MoveHome() {
	p.CursorPos = 0
}

// MoveEnd moves the cursor to the end.
func (p *PromptState) MoveEnd() {
	p.CursorPos = len(p.Input)
}
