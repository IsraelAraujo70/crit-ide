package render

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/israelcorrea/crit-ide/internal/editor"
)

// ViewState contains everything the renderer needs to draw a frame.
// It decouples rendering from the app package.
type ViewState struct {
	Buffer  *editor.Buffer
	ScrollY int
	Width   int
	Height  int // Total screen height (including statusline).
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

	editorHeight := vs.Height - 1 // Reserve 1 line for statusline.
	if editorHeight < 1 {
		editorHeight = 1
	}

	gutterWidth := r.gutterWidth(vs.Buffer.Text.LineCount())
	textWidth := vs.Width - gutterWidth
	if textWidth < 1 {
		textWidth = 1
	}

	defaultStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault)
	gutterStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDefault)
	cursorLineGutterStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault)

	for row := 0; row < editorHeight; row++ {
		lineIdx := vs.ScrollY + row
		if lineIdx >= vs.Buffer.Text.LineCount() {
			// Draw tilde for lines beyond the document.
			r.drawString(0, row, "~", gutterStyle)
			continue
		}

		// Draw gutter (line number).
		lineNum := fmt.Sprintf("%*d ", gutterWidth-1, lineIdx+1)
		gs := gutterStyle
		if lineIdx == vs.Buffer.CursorRow {
			gs = cursorLineGutterStyle
		}
		r.drawString(0, row, lineNum, gs)

		// Draw line content.
		line := vs.Buffer.Text.Line(lineIdx)
		col := 0
		for _, ch := range line {
			if col >= textWidth {
				break
			}
			if ch == '\t' {
				// Render tab as spaces (4-space tab stops).
				spaces := 4 - (col % 4)
				for s := 0; s < spaces && col < textWidth; s++ {
					r.screen.SetContent(gutterWidth+col, row, ' ', nil, defaultStyle)
					col++
				}
			} else {
				r.screen.SetContent(gutterWidth+col, row, ch, nil, defaultStyle)
				col++
			}
		}
	}

	// Draw statusline.
	r.drawStatusline(vs, editorHeight, gutterWidth)

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

// drawStatusline renders the bottom status bar.
func (r *Renderer) drawStatusline(vs *ViewState, y int, gutterWidth int) {
	statusStyle := tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite)

	// Clear the statusline.
	for x := 0; x < vs.Width; x++ {
		r.screen.SetContent(x, y, ' ', nil, statusStyle)
	}

	// Left: file name + dirty flag.
	name := vs.Buffer.FileName()
	dirty := ""
	if vs.Buffer.Dirty {
		dirty = " [+]"
	}
	left := fmt.Sprintf(" %s%s", name, dirty)
	r.drawString(0, y, left, statusStyle)

	// Right: cursor position.
	right := fmt.Sprintf("Ln %d, Col %d ", vs.Buffer.CursorRow+1, vs.Buffer.CursorCol+1)
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
