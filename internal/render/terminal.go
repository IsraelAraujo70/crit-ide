package render

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/terminal"
)

// renderTerminal draws the terminal panel at the bottom of the screen.
func (r *Renderer) renderTerminal(ts *editor.TerminalState, width, height int) {
	if !ts.Visible || ts.Height < 3 {
		return
	}

	panelHeight := ts.Height
	if panelHeight > height-4 {
		panelHeight = height - 4 // Leave room for tab bar + border + statusline.
	}

	startY := height - 1 - panelHeight // -1 for statusline.

	// Colors.
	borderColor := tcell.NewRGBColor(50, 50, 50)
	if ts.Focused {
		borderColor = tcell.NewRGBColor(80, 140, 255)
	}
	borderStyle := tcell.StyleDefault.Foreground(borderColor).Background(tcell.ColorDefault)
	headerStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(180, 180, 180)).Background(tcell.NewRGBColor(30, 30, 30))
	tabActiveStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(50, 50, 50)).Bold(true)
	tabInactiveStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(120, 120, 120)).Background(tcell.NewRGBColor(30, 30, 30))
	bodyBg := tcell.StyleDefault.Background(tcell.NewRGBColor(15, 15, 15))

	// --- Draw top border ---
	for x := 0; x < width; x++ {
		r.screen.SetContent(x, startY, tcell.RuneHLine, nil, borderStyle)
	}
	startY++
	panelHeight--

	// --- Draw tab bar / header ---
	// Fill header background.
	for x := 0; x < width; x++ {
		r.screen.SetContent(x, startY, ' ', nil, headerStyle)
	}

	// Draw "TERMINAL" label.
	label := " TERMINAL "
	for i, ch := range label {
		if i < width {
			r.screen.SetContent(i, startY, ch, nil, headerStyle)
		}
	}

	// Draw session tabs.
	x := len(label) + 1
	for i, name := range ts.TabNames {
		style := tabInactiveStyle
		if i == ts.ActiveTab {
			style = tabActiveStyle
		}

		// Tab separator.
		if x < width {
			r.screen.SetContent(x, startY, '|', nil, headerStyle)
			x++
		}

		// Tab label with index.
		tabLabel := fmt.Sprintf(" %d:%s ", i+1, name)
		if ts.TabClosed != nil && i < len(ts.TabClosed) && ts.TabClosed[i] {
			tabLabel = fmt.Sprintf(" %d:%s[x] ", i+1, name)
		}
		for _, ch := range tabLabel {
			if x < width {
				r.screen.SetContent(x, startY, ch, nil, style)
				x++
			}
		}
	}

	startY++
	panelHeight--

	// --- Draw terminal output ---
	contentHeight := panelHeight
	if contentHeight < 1 {
		return
	}

	// Clear the terminal area.
	for row := 0; row < contentHeight; row++ {
		for col := 0; col < width; col++ {
			r.screen.SetContent(col, startY+row, ' ', nil, bodyBg)
		}
	}

	lines := ts.Lines
	totalLines := len(lines)

	// Calculate which lines to display.
	// ScrollY is from the bottom: 0 = show the latest lines.
	endLine := totalLines - ts.ScrollY
	if endLine < 0 {
		endLine = 0
	}
	startLine := endLine - contentHeight
	if startLine < 0 {
		startLine = 0
	}

	// Normalize selection for rendering.
	selActive := ts.HasSelection
	selStartLine, selStartCol := ts.SelStartLine, ts.SelStartCol
	selEndLine, selEndCol := ts.SelEndLine, ts.SelEndCol
	if selActive && (selStartLine > selEndLine || (selStartLine == selEndLine && selStartCol > selEndCol)) {
		selStartLine, selStartCol, selEndLine, selEndCol = selEndLine, selEndCol, selStartLine, selStartCol
	}
	selStyle := tcell.StyleDefault.Background(tcell.NewRGBColor(60, 90, 160)).Foreground(tcell.ColorWhite)

	for row := 0; row < contentHeight; row++ {
		lineIdx := startLine + row
		if lineIdx >= endLine || lineIdx >= totalLines {
			break
		}

		line := lines[lineIdx]
		styled := terminal.ParseANSILine(line)

		col := 0
		for _, sc := range styled {
			if col >= width {
				break
			}
			style := sc.Style.Background(tcell.NewRGBColor(15, 15, 15))
			// Apply selection highlight.
			if selActive && lineIdx >= selStartLine && lineIdx <= selEndLine {
				inSel := false
				if selStartLine == selEndLine {
					inSel = col >= selStartCol && col < selEndCol
				} else if lineIdx == selStartLine {
					inSel = col >= selStartCol
				} else if lineIdx == selEndLine {
					inSel = col < selEndCol
				} else {
					inSel = true
				}
				if inSel {
					style = selStyle
				}
			}
			r.screen.SetContent(col, startY+row, sc.Rune, nil, style)
			col++
		}
	}

	// Draw cursor at its real position from the terminal grid.
	if ts.Focused && ts.ScrollY == 0 && ts.GridRows > 0 {
		// The cursor row is relative to the grid. The grid rows are the last
		// ts.GridRows entries in lines[]. Map cursor grid row to display row.
		gridStartIdx := totalLines - ts.GridRows
		if gridStartIdx < 0 {
			gridStartIdx = 0
		}
		cursorLineIdx := gridStartIdx + ts.CursorRow
		displayRow := cursorLineIdx - startLine
		if displayRow >= 0 && displayRow < contentHeight && ts.CursorCol < width {
			cursorStyle := tcell.StyleDefault.Background(tcell.NewRGBColor(200, 200, 200)).Foreground(tcell.NewRGBColor(15, 15, 15))
			r.screen.SetContent(ts.CursorCol, startY+displayRow, ' ', nil, cursorStyle)
		}
	}
}
