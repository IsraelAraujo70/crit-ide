package render

import (
	"fmt"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/israelcorrea/crit-ide/internal/editor"
)

// renderPalette draws the centered command palette popup.
func (r *Renderer) renderPalette(ps *editor.PaletteState, screenW, screenH int) {
	// Popup dimensions.
	popupWidth := screenW * 2 / 3
	if popupWidth < 50 {
		popupWidth = 50
	}
	if popupWidth > screenW-4 {
		popupWidth = screenW - 4
	}

	maxVisible := 15
	popupHeight := maxVisible + 4 // borders (2) + input row (1) + separator (1)
	if popupHeight > screenH-2 {
		popupHeight = screenH - 2
		maxVisible = popupHeight - 4
		if maxVisible < 1 {
			maxVisible = 1
		}
	}

	// Center horizontally, bias toward top vertically.
	px := (screenW - popupWidth) / 2
	py := (screenH - popupHeight) / 4
	if py < 1 {
		py = 1
	}

	// Styles.
	borderStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(180, 120, 255)).
		Background(tcell.NewRGBColor(20, 20, 30))
	headerStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(200, 160, 255)).
		Background(tcell.NewRGBColor(20, 20, 30))
	inputBg := tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.NewRGBColor(30, 30, 45))
	inputLabelStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(180, 120, 255)).
		Background(tcell.NewRGBColor(30, 30, 45)).
		Bold(true)
	resultStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(200, 200, 200)).
		Background(tcell.NewRGBColor(20, 20, 30))
	selectedStyle := tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.NewRGBColor(60, 40, 100)).
		Bold(true)
	categoryStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(120, 160, 220)).
		Background(tcell.NewRGBColor(20, 20, 30))
	categorySelectedStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(150, 180, 255)).
		Background(tcell.NewRGBColor(60, 40, 100))
	keybindStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(130, 130, 160)).
		Background(tcell.NewRGBColor(20, 20, 30))
	keybindSelectedStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(170, 170, 210)).
		Background(tcell.NewRGBColor(60, 40, 100))
	emptyStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(100, 100, 100)).
		Background(tcell.NewRGBColor(20, 20, 30))

	// Draw top border.
	r.screen.SetContent(px, py, tcell.RuneULCorner, nil, borderStyle)
	title := " Command Palette "
	titleStart := (popupWidth - len(title)) / 2
	for x := 1; x < popupWidth-1; x++ {
		ch := tcell.RuneHLine
		style := borderStyle
		if x >= titleStart && x < titleStart+len(title) {
			ch = rune(title[x-titleStart])
			style = headerStyle
		}
		r.screen.SetContent(px+x, py, ch, nil, style)
	}
	r.screen.SetContent(px+popupWidth-1, py, tcell.RuneURCorner, nil, borderStyle)

	// Draw input row (py+1).
	inputRow := py + 1
	r.screen.SetContent(px, inputRow, tcell.RuneVLine, nil, borderStyle)

	// Fill input row background.
	for x := 1; x < popupWidth-1; x++ {
		r.screen.SetContent(px+x, inputRow, ' ', nil, inputBg)
	}

	// Draw prompt prefix and input.
	label := " > "
	cx := px + 1
	for _, ch := range label {
		r.screen.SetContent(cx, inputRow, ch, nil, inputLabelStyle)
		cx++
	}

	// Draw query text.
	cursorScreenX := cx
	qRunes := []rune(ps.Query)
	qBytePos := 0
	for _, ch := range qRunes {
		if qBytePos == ps.CursorPos {
			cursorScreenX = cx
		}
		if cx >= px+popupWidth-1 {
			break
		}
		r.screen.SetContent(cx, inputRow, ch, nil, inputBg)
		cx++
		qBytePos += utf8.RuneLen(ch)
	}
	if qBytePos == ps.CursorPos {
		cursorScreenX = cx
	}

	// Show cursor.
	r.screen.ShowCursor(cursorScreenX, inputRow)

	// Show result count.
	countStr := fmt.Sprintf("%d/%d ", ps.ResultCount(), len(ps.AllEntries))
	countX := px + popupWidth - 1 - len(countStr)
	if countX > cx+1 {
		for _, ch := range countStr {
			r.screen.SetContent(countX, inputRow, ch, nil, tcell.StyleDefault.
				Foreground(tcell.NewRGBColor(100, 130, 170)).
				Background(tcell.NewRGBColor(30, 30, 45)))
			countX++
		}
	}

	r.screen.SetContent(px+popupWidth-1, inputRow, tcell.RuneVLine, nil, borderStyle)

	// Draw separator between input and results (py+2).
	sepRow := py + 2
	r.screen.SetContent(px, sepRow, tcell.RuneLTee, nil, borderStyle)
	for x := 1; x < popupWidth-1; x++ {
		r.screen.SetContent(px+x, sepRow, tcell.RuneHLine, nil, borderStyle)
	}
	r.screen.SetContent(px+popupWidth-1, sepRow, tcell.RuneRTee, nil, borderStyle)

	// Draw results.
	contentWidth := popupWidth - 2
	for row := 0; row < maxVisible; row++ {
		screenRow := py + 3 + row
		resultIdx := ps.ScrollY + row

		r.screen.SetContent(px, screenRow, tcell.RuneVLine, nil, borderStyle)

		if resultIdx >= len(ps.Filtered) {
			// Empty row.
			for x := 0; x < contentWidth; x++ {
				r.screen.SetContent(px+1+x, screenRow, ' ', nil, resultStyle)
			}
		} else {
			entry := ps.Filtered[resultIdx]
			isSelected := resultIdx == ps.SelectedIdx

			baseStyle := resultStyle
			catStyle := categoryStyle
			kbStyle := keybindStyle
			if isSelected {
				baseStyle = selectedStyle
				catStyle = categorySelectedStyle
				kbStyle = keybindSelectedStyle
			}

			// Layout: " [Category]  Label                  Keybinding "
			// Category tag.
			catTag := "[" + entry.Category + "]"
			kbText := entry.Keybinding

			// Reserve space for keybinding on the right.
			kbWidth := 0
			if kbText != "" {
				kbWidth = len(kbText) + 2 // padding
			}

			x := 0

			// Leading space.
			r.screen.SetContent(px+1, screenRow, ' ', nil, baseStyle)
			x++

			// Draw category tag.
			for _, ch := range catTag {
				if x >= contentWidth {
					break
				}
				r.screen.SetContent(px+1+x, screenRow, ch, nil, catStyle)
				x++
			}

			// Space after category.
			if x < contentWidth {
				r.screen.SetContent(px+1+x, screenRow, ' ', nil, baseStyle)
				x++
			}

			// Draw label.
			labelEnd := contentWidth - kbWidth
			for _, ch := range entry.Label {
				if x >= labelEnd {
					break
				}
				r.screen.SetContent(px+1+x, screenRow, ch, nil, baseStyle)
				x++
			}

			// Fill gap between label and keybinding.
			for x < contentWidth-kbWidth {
				r.screen.SetContent(px+1+x, screenRow, ' ', nil, baseStyle)
				x++
			}

			// Draw keybinding on the right.
			if kbText != "" {
				r.screen.SetContent(px+1+x, screenRow, ' ', nil, kbStyle)
				x++
				for _, ch := range kbText {
					if x >= contentWidth {
						break
					}
					r.screen.SetContent(px+1+x, screenRow, ch, nil, kbStyle)
					x++
				}
			}

			// Fill remaining.
			for x < contentWidth {
				r.screen.SetContent(px+1+x, screenRow, ' ', nil, baseStyle)
				x++
			}
		}

		r.screen.SetContent(px+popupWidth-1, screenRow, tcell.RuneVLine, nil, borderStyle)
	}

	// Draw empty state message if no results.
	if len(ps.Filtered) == 0 && ps.Query != "" {
		msgRow := py + 3
		msg := "No matching commands"
		msgStart := (contentWidth - len(msg)) / 2
		if msgStart < 1 {
			msgStart = 1
		}
		for i, ch := range msg {
			if msgStart+i < contentWidth {
				r.screen.SetContent(px+1+msgStart+i, msgRow, ch, nil, emptyStyle)
			}
		}
	}

	// Draw bottom border.
	bottomRow := py + 3 + maxVisible
	r.screen.SetContent(px, bottomRow, tcell.RuneLLCorner, nil, borderStyle)
	for x := 1; x < popupWidth-1; x++ {
		r.screen.SetContent(px+x, bottomRow, tcell.RuneHLine, nil, borderStyle)
	}
	r.screen.SetContent(px+popupWidth-1, bottomRow, tcell.RuneLRCorner, nil, borderStyle)

	// Hint line inside the bottom border.
	hint := " Enter:Execute  Esc:Close  Up/Down:Navigate "
	hintStart := (popupWidth - len(hint)) / 2
	if hintStart > 1 {
		for i, ch := range hint {
			r.screen.SetContent(px+hintStart+i, bottomRow, ch, nil, tcell.StyleDefault.
				Foreground(tcell.NewRGBColor(100, 120, 160)).
				Background(tcell.NewRGBColor(20, 20, 30)))
		}
	}
}
