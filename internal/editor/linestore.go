package editor

import (
	"strings"
)

// LineStore is a simple TextStore backed by a slice of strings (one per line).
// It is correct and easy to reason about. For large files, it can be replaced
// by a rope or piece table behind the same TextStore interface.
type LineStore struct {
	lines []string
}

// NewLineStore creates a LineStore from initial content.
// An empty string produces a single empty line (the minimum valid document).
func NewLineStore(content string) *LineStore {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	return &LineStore{lines: lines}
}

func (ls *LineStore) LineCount() int {
	return len(ls.lines)
}

func (ls *LineStore) Line(n int) string {
	if n < 0 || n >= len(ls.lines) {
		return ""
	}
	return ls.lines[n]
}

func (ls *LineStore) Insert(pos Position, text string) error {
	if pos.Line < 0 || pos.Line >= len(ls.lines) {
		return ErrOutOfRange
	}
	line := ls.lines[pos.Line]
	if pos.Col < 0 || pos.Col > len(line) {
		return ErrOutOfRange
	}

	before := line[:pos.Col]
	after := line[pos.Col:]

	if !strings.Contains(text, "\n") {
		// Simple case: insert within a single line.
		ls.lines[pos.Line] = before + text + after
		return nil
	}

	// Multi-line insert: split the inserted text into lines.
	parts := strings.Split(text, "\n")
	// First part joins with 'before', last part joins with 'after'.
	newLines := make([]string, 0, len(ls.lines)+len(parts)-1)
	newLines = append(newLines, ls.lines[:pos.Line]...)
	newLines = append(newLines, before+parts[0])
	for _, mid := range parts[1 : len(parts)-1] {
		newLines = append(newLines, mid)
	}
	newLines = append(newLines, parts[len(parts)-1]+after)
	newLines = append(newLines, ls.lines[pos.Line+1:]...)
	ls.lines = newLines
	return nil
}

func (ls *LineStore) Delete(r Range) error {
	// Validate range ordering.
	if r.Start.Line > r.End.Line || (r.Start.Line == r.End.Line && r.Start.Col > r.End.Col) {
		return ErrOutOfRange
	}
	if r.Start.Line < 0 || r.Start.Line >= len(ls.lines) {
		return ErrOutOfRange
	}
	if r.End.Line < 0 || r.End.Line >= len(ls.lines) {
		return ErrOutOfRange
	}

	startLine := ls.lines[r.Start.Line]
	endLine := ls.lines[r.End.Line]

	if r.Start.Col < 0 || r.Start.Col > len(startLine) {
		return ErrOutOfRange
	}
	if r.End.Col < 0 || r.End.Col > len(endLine) {
		return ErrOutOfRange
	}

	before := startLine[:r.Start.Col]
	after := endLine[r.End.Col:]
	merged := before + after

	newLines := make([]string, 0, len(ls.lines)-(r.End.Line-r.Start.Line))
	newLines = append(newLines, ls.lines[:r.Start.Line]...)
	newLines = append(newLines, merged)
	newLines = append(newLines, ls.lines[r.End.Line+1:]...)
	ls.lines = newLines
	return nil
}

func (ls *LineStore) Slice(r Range) string {
	if r.Start.Line < 0 || r.End.Line < 0 || r.End.Line >= len(ls.lines) {
		return ""
	}
	if r.Start.Line > r.End.Line || (r.Start.Line == r.End.Line && r.Start.Col > r.End.Col) {
		return ""
	}
	if r.Start.Line == r.End.Line {
		line := ls.lines[r.Start.Line]
		start := clampInt(r.Start.Col, 0, len(line))
		end := clampInt(r.End.Col, 0, len(line))
		return line[start:end]
	}

	var sb strings.Builder
	// First line: from start col to end of line.
	first := ls.lines[r.Start.Line]
	start := clampInt(r.Start.Col, 0, len(first))
	sb.WriteString(first[start:])

	// Middle lines: full lines.
	for i := r.Start.Line + 1; i < r.End.Line; i++ {
		sb.WriteByte('\n')
		sb.WriteString(ls.lines[i])
	}

	// Last line: from start of line to end col.
	sb.WriteByte('\n')
	last := ls.lines[r.End.Line]
	end := clampInt(r.End.Col, 0, len(last))
	sb.WriteString(last[:end])

	return sb.String()
}

// Content returns the full document as a single string with newline separators.
func (ls *LineStore) Content() string {
	return strings.Join(ls.lines, "\n")
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
