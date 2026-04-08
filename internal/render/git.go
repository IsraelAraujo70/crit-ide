package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/israelcorrea/crit-ide/internal/editor"
)

// buildGitGutterMap creates a line→status lookup for git gutter indicators.
func (r *Renderer) buildGitGutterMap(gutter []GutterDiffInfo) map[int]int {
	m := make(map[int]int, len(gutter))
	for _, g := range gutter {
		m[g.Line] = g.Status
	}
	return m
}

// renderGitStatus draws the Git status panel as a centered popup overlay.
func (r *Renderer) renderGitStatus(state *editor.GitStatusState, width, height int) {
	popupWidth := width * 2 / 3
	if popupWidth < 50 {
		popupWidth = 50
	}
	if popupWidth > width-4 {
		popupWidth = width - 4
	}
	popupHeight := height * 2 / 3
	if popupHeight < 10 {
		popupHeight = 10
	}
	if popupHeight > height-4 {
		popupHeight = height - 4
	}

	px := (width - popupWidth) / 2
	py := (height - popupHeight) / 2

	borderStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(80, 140, 255)).Background(tcell.NewRGBColor(20, 20, 35))
	bgStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(200, 200, 200)).Background(tcell.NewRGBColor(20, 20, 35))
	headerStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(120, 200, 255)).Background(tcell.NewRGBColor(20, 20, 35)).Bold(true)

	// Draw border.
	r.drawBoxBorder(px, py, popupWidth, popupHeight, borderStyle)

	// Draw title.
	title := fmt.Sprintf(" Git Status (%s) ", state.Branch)
	r.drawString(px+2, py, title, headerStyle)

	// Draw hint line.
	hintRow := py + 1
	hints := " s:stage/unstage  d:diff  c:commit  Enter:open  Esc:close"
	for x := 1; x < popupWidth-1; x++ {
		r.screen.SetContent(px+x, hintRow, ' ', nil, bgStyle)
	}
	r.drawStringClipped(px+1, hintRow, hints, popupWidth-2, tcell.StyleDefault.Foreground(tcell.NewRGBColor(120, 120, 150)).Background(tcell.NewRGBColor(20, 20, 35)))

	// Separator.
	sepRow := py + 2
	sepStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(50, 50, 70)).Background(tcell.NewRGBColor(20, 20, 35))
	for x := 1; x < popupWidth-1; x++ {
		r.screen.SetContent(px+x, sepRow, '─', nil, sepStyle)
	}

	// Draw entries.
	contentHeight := popupHeight - 4 // border(1) + hint(1) + sep(1) + bottom border(1)
	startRow := py + 3

	// Adjust scrollY to keep cursor visible.
	if state.SelectedIdx < state.ScrollY {
		state.ScrollY = state.SelectedIdx
	}
	if state.SelectedIdx >= state.ScrollY+contentHeight {
		state.ScrollY = state.SelectedIdx - contentHeight + 1
	}

	for i := 0; i < contentHeight; i++ {
		screenRow := startRow + i
		entryIdx := state.ScrollY + i

		// Clear row.
		for x := 1; x < popupWidth-1; x++ {
			r.screen.SetContent(px+x, screenRow, ' ', nil, bgStyle)
		}

		if entryIdx >= len(state.Entries) {
			continue
		}
		entry := state.Entries[entryIdx]

		// Determine style based on status.
		var statusStyle tcell.Style
		statusChar := string(entry.Status)
		if entry.Staged {
			statusStyle = tcell.StyleDefault.Foreground(tcell.NewRGBColor(80, 200, 80)).Background(tcell.NewRGBColor(20, 20, 35))
			statusChar = "S:" + statusChar
		} else {
			switch entry.Status {
			case "M":
				statusStyle = tcell.StyleDefault.Foreground(tcell.NewRGBColor(220, 180, 50)).Background(tcell.NewRGBColor(20, 20, 35))
			case "D":
				statusStyle = tcell.StyleDefault.Foreground(tcell.NewRGBColor(220, 60, 60)).Background(tcell.NewRGBColor(20, 20, 35))
			case "??":
				statusStyle = tcell.StyleDefault.Foreground(tcell.NewRGBColor(120, 120, 140)).Background(tcell.NewRGBColor(20, 20, 35))
			default:
				statusStyle = bgStyle
			}
		}

		// Highlight selected row.
		rowStyle := bgStyle
		if entryIdx == state.SelectedIdx {
			rowStyle = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(40, 50, 80))
			statusStyle = statusStyle.Background(tcell.NewRGBColor(40, 50, 80))
			// Fill the row with highlight.
			for x := 1; x < popupWidth-1; x++ {
				r.screen.SetContent(px+x, screenRow, ' ', nil, rowStyle)
			}
		}

		// Draw status indicator.
		r.drawString(px+2, screenRow, statusChar, statusStyle)

		// Draw file path.
		pathX := px + 2 + len(statusChar) + 1
		r.drawStringClipped(pathX, screenRow, entry.Path, popupWidth-len(statusChar)-4, rowStyle)
	}

	// Draw entry count at bottom.
	countStr := fmt.Sprintf(" %d files ", len(state.Entries))
	r.drawString(px+popupWidth-len(countStr)-1, py+popupHeight-1, countStr, borderStyle)

	r.screen.HideCursor()
}

// renderGitGraph draws the Git graph panel as a centered popup overlay.
func (r *Renderer) renderGitGraph(state *editor.GitGraphState, width, height int) {
	popupWidth := width * 3 / 4
	if popupWidth < 60 {
		popupWidth = 60
	}
	if popupWidth > width-4 {
		popupWidth = width - 4
	}
	popupHeight := height * 3 / 4
	if popupHeight < 15 {
		popupHeight = 15
	}
	if popupHeight > height-4 {
		popupHeight = height - 4
	}

	px := (width - popupWidth) / 2
	py := (height - popupHeight) / 2

	borderStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(80, 140, 255)).Background(tcell.NewRGBColor(20, 20, 35))
	bgStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(200, 200, 200)).Background(tcell.NewRGBColor(20, 20, 35))
	headerStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(120, 200, 255)).Background(tcell.NewRGBColor(20, 20, 35)).Bold(true)

	// Draw border.
	r.drawBoxBorder(px, py, popupWidth, popupHeight, borderStyle)

	// Draw title.
	r.drawString(px+2, py, " Git Graph ", headerStyle)

	// Hints.
	hintRow := py + 1
	for x := 1; x < popupWidth-1; x++ {
		r.screen.SetContent(px+x, hintRow, ' ', nil, bgStyle)
	}
	r.drawStringClipped(px+1, hintRow, " Up/Down:navigate  Esc:close", popupWidth-2, tcell.StyleDefault.Foreground(tcell.NewRGBColor(120, 120, 150)).Background(tcell.NewRGBColor(20, 20, 35)))

	// Content.
	contentHeight := popupHeight - 3
	startRow := py + 2

	// Adjust scrollY.
	if state.SelectedIdx < state.ScrollY {
		state.ScrollY = state.SelectedIdx
	}
	if state.SelectedIdx >= state.ScrollY+contentHeight {
		state.ScrollY = state.SelectedIdx - contentHeight + 1
	}

	// Graph line colors by branch level.
	graphColors := []tcell.Color{
		tcell.NewRGBColor(80, 200, 80),   // Green.
		tcell.NewRGBColor(80, 140, 255),  // Blue.
		tcell.NewRGBColor(255, 140, 50),  // Orange.
		tcell.NewRGBColor(200, 80, 200),  // Purple.
		tcell.NewRGBColor(80, 200, 200),  // Cyan.
		tcell.NewRGBColor(220, 180, 50),  // Yellow.
	}

	for i := 0; i < contentHeight; i++ {
		screenRow := startRow + i
		lineIdx := state.ScrollY + i

		// Clear row.
		for x := 1; x < popupWidth-1; x++ {
			r.screen.SetContent(px+x, screenRow, ' ', nil, bgStyle)
		}

		if lineIdx >= len(state.Lines) {
			continue
		}
		line := state.Lines[lineIdx]

		// Highlight selected row.
		rowStyle := bgStyle
		if lineIdx == state.SelectedIdx {
			rowStyle = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(40, 50, 80))
			for x := 1; x < popupWidth-1; x++ {
				r.screen.SetContent(px+x, screenRow, ' ', nil, rowStyle)
			}
		}

		// Render graph line with colors.
		text := line.Text
		col := px + 1
		graphPart := true
		colorIdx := 0
		for _, ch := range text {
			if col >= px+popupWidth-1 {
				break
			}
			style := rowStyle
			if graphPart {
				// Graph characters get colored.
				switch ch {
				case '│', '├', '─', '┤', '┘', '┐', '┌', '└', '╱', '╲', '|', '\\', '/', '*':
					style = style.Foreground(graphColors[colorIdx%len(graphColors)])
				case ' ':
					// Space in graph portion.
				default:
					// Non-graph character — switch to text mode.
					graphPart = false
				}
				if ch == '|' || ch == '│' || ch == '\\' || ch == '/' || ch == '*' {
					colorIdx++
				}
			}
			// Draw refs in special color.
			if !graphPart && line.Refs != "" && strings.Contains(text, line.Refs) {
				// Keep default style for commit text.
			}
			r.screen.SetContent(col, screenRow, ch, nil, style)
			col++
		}
	}

	// Line count.
	countStr := fmt.Sprintf(" %d commits ", len(state.Lines))
	r.drawString(px+popupWidth-len(countStr)-1, py+popupHeight-1, countStr, borderStyle)

	r.screen.HideCursor()
}

// renderGitDiff draws the Git diff viewer as a full-screen overlay.
func (r *Renderer) renderGitDiff(state *editor.GitDiffState, width, height int) {
	popupWidth := width * 3 / 4
	if popupWidth < 60 {
		popupWidth = 60
	}
	if popupWidth > width-4 {
		popupWidth = width - 4
	}
	popupHeight := height * 3 / 4
	if popupHeight < 15 {
		popupHeight = 15
	}
	if popupHeight > height-4 {
		popupHeight = height - 4
	}

	px := (width - popupWidth) / 2
	py := (height - popupHeight) / 2

	borderStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(80, 140, 255)).Background(tcell.NewRGBColor(20, 20, 35))
	bgStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(200, 200, 200)).Background(tcell.NewRGBColor(20, 20, 35))
	headerStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(120, 200, 255)).Background(tcell.NewRGBColor(20, 20, 35)).Bold(true)
	addedStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(80, 200, 80)).Background(tcell.NewRGBColor(20, 30, 20))
	removedStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(220, 80, 80)).Background(tcell.NewRGBColor(30, 20, 20))
	hunkStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(100, 160, 220)).Background(tcell.NewRGBColor(20, 20, 35))

	// Draw border.
	r.drawBoxBorder(px, py, popupWidth, popupHeight, borderStyle)

	// Draw title.
	title := fmt.Sprintf(" %s ", state.Title)
	r.drawString(px+2, py, title, headerStyle)

	// Hints.
	hintRow := py + 1
	for x := 1; x < popupWidth-1; x++ {
		r.screen.SetContent(px+x, hintRow, ' ', nil, bgStyle)
	}
	r.drawStringClipped(px+1, hintRow, " Up/Down:scroll  Esc:close", popupWidth-2, tcell.StyleDefault.Foreground(tcell.NewRGBColor(120, 120, 150)).Background(tcell.NewRGBColor(20, 20, 35)))

	// Content.
	contentHeight := popupHeight - 3
	startRow := py + 2

	for i := 0; i < contentHeight; i++ {
		screenRow := startRow + i
		lineIdx := state.ScrollY + i

		// Clear row.
		for x := 1; x < popupWidth-1; x++ {
			r.screen.SetContent(px+x, screenRow, ' ', nil, bgStyle)
		}

		if lineIdx >= len(state.Lines) {
			continue
		}
		line := state.Lines[lineIdx]

		// Determine style based on line content.
		style := bgStyle
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			style = addedStyle
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			style = removedStyle
		} else if strings.HasPrefix(line, "@@") {
			style = hunkStyle
		} else if strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "index ") ||
			strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
			style = hunkStyle
		}

		// Draw line.
		col := px + 1
		for _, ch := range line {
			if col >= px+popupWidth-1 {
				break
			}
			r.screen.SetContent(col, screenRow, ch, nil, style)
			col++
		}
	}

	r.screen.HideCursor()
}

// drawBoxBorder draws a box border at the given position.
func (r *Renderer) drawBoxBorder(x, y, w, h int, style tcell.Style) {
	// Top border.
	r.screen.SetContent(x, y, tcell.RuneULCorner, nil, style)
	for i := 1; i < w-1; i++ {
		r.screen.SetContent(x+i, y, tcell.RuneHLine, nil, style)
	}
	r.screen.SetContent(x+w-1, y, tcell.RuneURCorner, nil, style)

	// Side borders.
	for i := 1; i < h-1; i++ {
		r.screen.SetContent(x, y+i, tcell.RuneVLine, nil, style)
		r.screen.SetContent(x+w-1, y+i, tcell.RuneVLine, nil, style)
	}

	// Bottom border.
	r.screen.SetContent(x, y+h-1, tcell.RuneLLCorner, nil, style)
	for i := 1; i < w-1; i++ {
		r.screen.SetContent(x+i, y+h-1, tcell.RuneHLine, nil, style)
	}
	r.screen.SetContent(x+w-1, y+h-1, tcell.RuneLRCorner, nil, style)
}

// drawStringClipped draws a string clipped to maxWidth characters.
func (r *Renderer) drawStringClipped(x, y int, s string, maxWidth int, style tcell.Style) {
	col := 0
	for _, ch := range s {
		if col >= maxWidth {
			break
		}
		r.screen.SetContent(x+col, y, ch, nil, style)
		col++
	}
}
