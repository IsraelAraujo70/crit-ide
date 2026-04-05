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
	Height  int              // Total screen height (including statusline).
	Popup   *editor.MenuState // Non-nil when a popup menu is active.
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
	selectionStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorLightGray)
	gutterStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDefault)
	cursorLineGutterStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault)

	// Precompute selection range for efficient per-character checks.
	var selStart, selEnd editor.Position
	hasSel := vs.Buffer.HasSelection()
	if hasSel {
		selRange := vs.Buffer.Selection.Normalized()
		selStart = selRange.Start
		selEnd = selRange.End
	}

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

		// Compute selection byte range for this line.
		lineSelStart := -1
		lineSelEnd := -1
		if hasSel {
			if lineIdx > selStart.Line && lineIdx < selEnd.Line {
				// Entire line is selected.
				lineSelStart = 0
				lineSelEnd = len(vs.Buffer.Text.Line(lineIdx)) + 1 // +1 to include past-end
			} else if lineIdx == selStart.Line && lineIdx == selEnd.Line {
				// Single-line selection.
				lineSelStart = selStart.Col
				lineSelEnd = selEnd.Col
			} else if lineIdx == selStart.Line {
				lineSelStart = selStart.Col
				lineSelEnd = len(vs.Buffer.Text.Line(lineIdx)) + 1
			} else if lineIdx == selEnd.Line {
				lineSelStart = 0
				lineSelEnd = selEnd.Col
			}
		}

		// Draw line content.
		line := vs.Buffer.Text.Line(lineIdx)
		col := 0
		for i, ch := range line {
			if col >= textWidth {
				break
			}
			// Choose style based on whether this byte offset is selected.
			style := defaultStyle
			if lineSelStart >= 0 && i >= lineSelStart && i < lineSelEnd {
				style = selectionStyle
			}

			if ch == '\t' {
				// Render tab as spaces (4-space tab stops).
				spaces := 4 - (col % 4)
				for s := 0; s < spaces && col < textWidth; s++ {
					r.screen.SetContent(gutterWidth+col, row, ' ', nil, style)
					col++
				}
			} else {
				r.screen.SetContent(gutterWidth+col, row, ch, nil, style)
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

	// Draw popup menu on top of editor content if active.
	if vs.Popup != nil {
		r.renderPopup(vs.Popup, vs.Width, vs.Height)
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

// gutterWidth delegates to editor.GutterWidth for a single source of truth.
func (r *Renderer) gutterWidth(lineCount int) int {
	return editor.GutterWidth(lineCount)
}

// renderPopup draws a context menu popup on top of existing content.
func (r *Renderer) renderPopup(menu *editor.MenuState, screenW, screenH int) {
	menuBg := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)
	menuSel := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
	menuSep := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDarkBlue)

	// Compute popup width from longest label.
	maxLabel := 0
	for _, item := range menu.Items {
		if !item.IsSeparator && len(item.Label) > maxLabel {
			maxLabel = len(item.Label)
		}
	}
	popupWidth := maxLabel + 4 // 2 chars padding on each side.
	if popupWidth < 12 {
		popupWidth = 12
	}
	popupHeight := len(menu.Items) + 2 // +2 for top/bottom border.

	// Clamp position so popup stays on screen.
	px := menu.ScreenX
	py := menu.ScreenY
	if px+popupWidth > screenW {
		px = screenW - popupWidth
	}
	if px < 0 {
		px = 0
	}
	if py+popupHeight > screenH {
		py = screenH - popupHeight
	}
	if py < 0 {
		py = 0
	}

	// Draw top border.
	r.screen.SetContent(px, py, tcell.RuneULCorner, nil, menuBg)
	for x := 1; x < popupWidth-1; x++ {
		r.screen.SetContent(px+x, py, tcell.RuneHLine, nil, menuBg)
	}
	r.screen.SetContent(px+popupWidth-1, py, tcell.RuneURCorner, nil, menuBg)

	// Draw items.
	for i, item := range menu.Items {
		row := py + 1 + i
		if item.IsSeparator {
			// Separator line.
			r.screen.SetContent(px, row, tcell.RuneLTee, nil, menuSep)
			for x := 1; x < popupWidth-1; x++ {
				r.screen.SetContent(px+x, row, tcell.RuneHLine, nil, menuSep)
			}
			r.screen.SetContent(px+popupWidth-1, row, tcell.RuneRTee, nil, menuSep)
		} else {
			style := menuBg
			if i == menu.SelectedIdx {
				style = menuSel
			}
			// Left border.
			r.screen.SetContent(px, row, tcell.RuneVLine, nil, menuBg)
			// Label with padding.
			label := fmt.Sprintf(" %-*s ", maxLabel, item.Label)
			for j, ch := range label {
				if j < popupWidth-2 {
					r.screen.SetContent(px+1+j, row, ch, nil, style)
				}
			}
			// Fill remaining space.
			for j := len(label); j < popupWidth-2; j++ {
				r.screen.SetContent(px+1+j, row, ' ', nil, style)
			}
			// Right border.
			r.screen.SetContent(px+popupWidth-1, row, tcell.RuneVLine, nil, menuBg)
		}
	}

	// Draw bottom border.
	bottomRow := py + popupHeight - 1
	r.screen.SetContent(px, bottomRow, tcell.RuneLLCorner, nil, menuBg)
	for x := 1; x < popupWidth-1; x++ {
		r.screen.SetContent(px+x, bottomRow, tcell.RuneHLine, nil, menuBg)
	}
	r.screen.SetContent(px+popupWidth-1, bottomRow, tcell.RuneLRCorner, nil, menuBg)
}

// drawString draws a string at the given position with the given style.
func (r *Renderer) drawString(x, y int, s string, style tcell.Style) {
	for _, ch := range s {
		r.screen.SetContent(x, y, ch, nil, style)
		x++
	}
}
