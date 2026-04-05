package render

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/highlight"
	"github.com/israelcorrea/crit-ide/internal/theme"
)

// DiagnosticRange represents a diagnostic marker on a specific line.
type DiagnosticRange struct {
	Line     int // Zero-based line index.
	StartCol int // Start byte offset within line.
	EndCol   int // End byte offset within line.
	Severity int // 1=Error, 2=Warning, 3=Info, 4=Hint.
}

// ViewState contains everything the renderer needs to draw a frame.
// It decouples rendering from the app package.
type ViewState struct {
	Buffer       *editor.Buffer
	ScrollY      int
	Width        int
	Height       int // Total screen height (including statusline).
	Highlighter  highlight.Highlighter
	Theme        *theme.Theme
	Diagnostics  []DiagnosticRange
	StatusMsg    string // Optional message to show in statusline (replaces cursor pos).
	DiagErrors   int    // Error count for statusline.
	DiagWarnings int    // Warning count for statusline.
}

// Renderer draws the editor state to a tcell screen.
type Renderer struct {
	screen tcell.Screen
}

// NewRenderer creates a renderer attached to the given screen.
func NewRenderer(screen tcell.Screen) *Renderer {
	return &Renderer{screen: screen}
}

// Render draws the full editor frame: line numbers, text content, cursor, and statusline.
func (r *Renderer) Render(vs *ViewState) {
	r.screen.Clear()

	th := vs.Theme
	if th == nil {
		th = theme.DefaultTheme()
	}

	editorHeight := vs.Height - 1 // Reserve 1 line for statusline.
	if editorHeight < 1 {
		editorHeight = 1
	}

	gutterWidth := r.gutterWidth(vs.Buffer.Text.LineCount())
	textWidth := vs.Width - gutterWidth
	if textWidth < 1 {
		textWidth = 1
	}

	// Build a quick-lookup map for diagnostics on visible lines.
	diagMap := r.buildDiagMap(vs.Diagnostics, vs.ScrollY, vs.ScrollY+editorHeight)

	for row := 0; row < editorHeight; row++ {
		lineIdx := vs.ScrollY + row
		if lineIdx >= vs.Buffer.Text.LineCount() {
			// Draw tilde for lines beyond the document.
			r.drawString(0, row, "~", th.Gutter)
			continue
		}

		// Draw gutter (line number).
		lineNum := fmt.Sprintf("%*d ", gutterWidth-1, lineIdx+1)
		gs := th.Gutter
		if lineIdx == vs.Buffer.CursorRow {
			gs = th.GutterActive
		}
		r.drawString(0, row, lineNum, gs)

		// Get highlight tokens for this line.
		line := vs.Buffer.Text.Line(lineIdx)
		var tokens []highlight.Token
		if vs.Highlighter != nil {
			tokens = vs.Highlighter.HighlightLine(lineIdx, line)
		}

		// Get diagnostics for this line.
		lineDiags := diagMap[lineIdx]

		// Draw line content with highlighting.
		r.drawHighlightedLine(gutterWidth, row, line, textWidth, tokens, lineDiags, th)
	}

	// Draw statusline.
	r.drawStatusline(vs, editorHeight, th)

	// Position the terminal cursor.
	cursorScreenRow := vs.Buffer.CursorRow - vs.ScrollY
	cursorScreenCol := r.screenCol(vs.Buffer, gutterWidth)
	if cursorScreenRow >= 0 && cursorScreenRow < editorHeight {
		r.screen.ShowCursor(cursorScreenCol, cursorScreenRow)
	} else {
		r.screen.HideCursor()
	}

	r.screen.Show()
}

// drawHighlightedLine draws a single line of text with syntax highlighting and diagnostic markers.
func (r *Renderer) drawHighlightedLine(x, y int, line string, maxCols int, tokens []highlight.Token, diags []DiagnosticRange, th *theme.Theme) {
	col := 0        // Screen column.
	tokenIdx := 0   // Current token index.
	byteOff := 0    // Current byte offset in line.

	for _, ch := range line {
		if col >= maxCols {
			break
		}

		runeLen := len(string(ch))

		// Determine style from highlight tokens.
		style := th.Default
		for tokenIdx < len(tokens) && tokens[tokenIdx].End <= byteOff {
			tokenIdx++
		}
		if tokenIdx < len(tokens) && tokens[tokenIdx].Start <= byteOff && byteOff < tokens[tokenIdx].End {
			style = th.StyleFor(tokens[tokenIdx].Type)
		}

		// Apply diagnostic underline if applicable.
		for _, d := range diags {
			if byteOff >= d.StartCol && byteOff < d.EndCol {
				switch d.Severity {
				case 1:
					style = style.Underline(true).Foreground(tcell.ColorRed)
				case 2:
					style = style.Underline(true).Foreground(tcell.ColorYellow)
				case 3:
					style = style.Underline(true).Foreground(tcell.ColorBlue)
				default:
					style = style.Underline(true)
				}
				break
			}
		}

		if ch == '\t' {
			// Render tab as spaces (4-space tab stops).
			spaces := 4 - (col % 4)
			for s := 0; s < spaces && col < maxCols; s++ {
				r.screen.SetContent(x+col, y, ' ', nil, style)
				col++
			}
		} else {
			r.screen.SetContent(x+col, y, ch, nil, style)
			col++
		}

		byteOff += runeLen
	}
}

// drawStatusline renders the bottom status bar.
func (r *Renderer) drawStatusline(vs *ViewState, y int, th *theme.Theme) {
	statusStyle := th.StatusLine

	// Clear the statusline.
	for x := 0; x < vs.Width; x++ {
		r.screen.SetContent(x, y, ' ', nil, statusStyle)
	}

	// Left: file name + dirty flag + language ID.
	name := vs.Buffer.FileName()
	dirty := ""
	if vs.Buffer.Dirty {
		dirty = " [+]"
	}
	langInfo := ""
	if vs.Buffer.LanguageID != "" {
		langInfo = " [" + vs.Buffer.LanguageID + "]"
	}
	left := fmt.Sprintf(" %s%s%s", name, dirty, langInfo)
	r.drawString(0, y, left, statusStyle)

	// Right side content.
	var right string
	if vs.StatusMsg != "" {
		right = vs.StatusMsg + " "
	} else {
		right = fmt.Sprintf("Ln %d, Col %d ", vs.Buffer.CursorRow+1, vs.Buffer.CursorCol+1)
	}

	// Add diagnostic counts if present.
	if vs.DiagErrors > 0 || vs.DiagWarnings > 0 {
		diagStr := fmt.Sprintf("E:%d W:%d  ", vs.DiagErrors, vs.DiagWarnings)
		right = diagStr + right
	}

	r.drawString(vs.Width-len(right), y, right, statusStyle)
}

// screenCol calculates the screen X position of the cursor, accounting for
// the gutter width and tab expansion.
func (r *Renderer) screenCol(buf *editor.Buffer, gutterWidth int) int {
	line := buf.Text.Line(buf.CursorRow)
	col := 0
	for i, ch := range line {
		if i >= buf.CursorCol {
			break
		}
		if ch == '\t' {
			col += 4 - (col % 4)
		} else {
			col++
		}
	}
	return gutterWidth + col
}

// gutterWidth calculates how many columns the line number gutter needs.
func (r *Renderer) gutterWidth(lineCount int) int {
	digits := 1
	n := lineCount
	for n >= 10 {
		digits++
		n /= 10
	}
	if digits < 3 {
		digits = 3 // Minimum gutter width.
	}
	return digits + 1 // +1 for the space separator.
}

// drawString draws a string at the given position with the given style.
func (r *Renderer) drawString(x, y int, s string, style tcell.Style) {
	for _, ch := range s {
		r.screen.SetContent(x, y, ch, nil, style)
		x++
	}
}

// buildDiagMap groups diagnostics by line for quick lookup.
func (r *Renderer) buildDiagMap(diags []DiagnosticRange, minLine, maxLine int) map[int][]DiagnosticRange {
	if len(diags) == 0 {
		return nil
	}
	m := make(map[int][]DiagnosticRange)
	for _, d := range diags {
		if d.Line >= minLine && d.Line < maxLine {
			m[d.Line] = append(m[d.Line], d)
		}
	}
	return m
}
