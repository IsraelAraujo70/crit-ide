package render

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/highlight"
	"github.com/israelcorrea/crit-ide/internal/theme"
)

// TabInfo holds the display information for a single tab.
type TabInfo struct {
	Name   string
	Dirty  bool
	Active bool
}

// TreeNode holds the display information for a single tree entry.
type TreeNode struct {
	Name     string
	IsDir    bool
	Expanded bool
	Depth    int
	Path     string
}

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
	Buffer  *editor.Buffer
	ScrollY int
	Width   int
	Height  int              // Total screen height (including statusline).
	Popup   *editor.MenuState // Non-nil when a popup menu is active.

	// Tab bar.
	Tabs          []TabInfo
	ActiveTabIdx  int

	// File tree.
	TreeVisible   bool
	TreeWidth     int
	TreeNodes     []TreeNode // Flattened visible nodes.
	TreeCursor    int        // Cursor index in TreeNodes.
	TreeScrollY   int        // Scroll offset for tree.
	TreeFocused   bool       // Whether the tree panel has focus.

	// Input prompt.
	Prompt        *editor.PromptState // Non-nil when prompt is active.

	// Syntax highlighting.
	Highlighter  highlight.Highlighter
	Theme        *theme.Theme

	// LSP diagnostics.
	Diagnostics  []DiagnosticRange
	StatusMsg    string // Optional message to show in statusline.
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

// Render draws the full editor frame: tab bar, line numbers, text content,
// file tree, cursor, and statusline.
func (r *Renderer) Render(vs *ViewState) {
	r.screen.Clear()

	th := vs.Theme
	if th == nil {
		th = theme.DefaultTheme()
	}

	tabBarHeight := 1 // Always show tab bar.
	statuslineHeight := 1

	editorHeight := vs.Height - tabBarHeight - statuslineHeight
	if editorHeight < 1 {
		editorHeight = 1
	}

	// Calculate editor width (accounting for file tree on the right).
	treeWidth := 0
	if vs.TreeVisible {
		treeWidth = vs.TreeWidth
		if treeWidth < 1 {
			treeWidth = 30
		}
	}
	editorWidth := vs.Width - treeWidth
	if editorWidth < 10 {
		editorWidth = 10
	}

	// Focus border color.
	focusBorderColor := tcell.NewRGBColor(80, 140, 255)
	dimBorderColor := tcell.NewRGBColor(50, 50, 50)

	// --- Draw tab bar (row 0) ---
	r.drawTabBar(vs, editorWidth, treeWidth)

	// --- Draw editor top border (row between tab bar and content) ---
	editorBorderColor := dimBorderColor
	if !vs.TreeFocused {
		editorBorderColor = focusBorderColor
	}
	editorBorderStyle := tcell.StyleDefault.Foreground(editorBorderColor).Background(tcell.ColorDefault)
	for x := 0; x < editorWidth; x++ {
		r.screen.SetContent(x, tabBarHeight, tcell.RuneHLine, nil, editorBorderStyle)
	}

	// Adjust: content starts after the border line.
	contentStartY := tabBarHeight + 1
	editorHeight = vs.Height - tabBarHeight - 1 - statuslineHeight // -1 for border
	if editorHeight < 1 {
		editorHeight = 1
	}

	// --- Draw editor area ---
	gutterWidth := r.gutterWidth(vs.Buffer.Text.LineCount())
	textWidth := editorWidth - gutterWidth
	if textWidth < 1 {
		textWidth = 1
	}

	defaultStyle := th.Default
	selectionStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorLightGray)

	// Precompute selection range for efficient per-character checks.
	var selStart, selEnd editor.Position
	hasSel := vs.Buffer.HasSelection()
	if hasSel {
		selRange := vs.Buffer.Selection.Normalized()
		selStart = selRange.Start
		selEnd = selRange.End
	}

	// Build a quick-lookup map for diagnostics on visible lines.
	diagMap := r.buildDiagMap(vs.Diagnostics, vs.ScrollY, vs.ScrollY+editorHeight)

	for row := 0; row < editorHeight; row++ {
		screenRow := contentStartY + row
		lineIdx := vs.ScrollY + row
		if lineIdx >= vs.Buffer.Text.LineCount() {
			// Draw tilde for lines beyond the document.
			r.drawString(0, screenRow, "~", th.Gutter)
			continue
		}

		// Draw gutter (line number).
		lineNum := fmt.Sprintf("%*d ", gutterWidth-1, lineIdx+1)
		gs := th.Gutter
		if lineIdx == vs.Buffer.CursorRow {
			gs = th.GutterActive
		}
		r.drawString(0, screenRow, lineNum, gs)

		// Compute selection byte range for this line.
		lineSelStart := -1
		lineSelEnd := -1
		if hasSel {
			if lineIdx > selStart.Line && lineIdx < selEnd.Line {
				lineSelStart = 0
				lineSelEnd = len(vs.Buffer.Text.Line(lineIdx)) + 1
			} else if lineIdx == selStart.Line && lineIdx == selEnd.Line {
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

		// Get highlight tokens for this line.
		line := vs.Buffer.Text.Line(lineIdx)
		var tokens []highlight.Token
		if vs.Highlighter != nil {
			tokens = vs.Highlighter.HighlightLine(lineIdx, line)
		}

		// Get diagnostics for this line.
		lineDiags := diagMap[lineIdx]

		// Draw line content with highlighting, selection, and diagnostics.
		col := 0
		tokenIdx := 0
		byteOff := 0
		for _, ch := range line {
			if col >= textWidth {
				break
			}

			runeLen := len(string(ch))

			// Determine style from highlight tokens.
			style := defaultStyle
			for tokenIdx < len(tokens) && tokens[tokenIdx].End <= byteOff {
				tokenIdx++
			}
			if tokenIdx < len(tokens) && tokens[tokenIdx].Start <= byteOff && byteOff < tokens[tokenIdx].End {
				style = th.StyleFor(tokens[tokenIdx].Type)
			}

			// Apply diagnostic underline if applicable.
			for _, d := range lineDiags {
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

			// Selection overrides everything.
			if lineSelStart >= 0 && byteOff >= lineSelStart && byteOff < lineSelEnd {
				style = selectionStyle
			}

			if ch == '\t' {
				spaces := 4 - (col % 4)
				for s := 0; s < spaces && col < textWidth; s++ {
					r.screen.SetContent(gutterWidth+col, screenRow, ' ', nil, style)
					col++
				}
			} else {
				r.screen.SetContent(gutterWidth+col, screenRow, ch, nil, style)
				col++
			}

			byteOff += runeLen
		}
	}

	// --- Draw file tree (right panel) ---
	if vs.TreeVisible {
		treeBorderColor := dimBorderColor
		if vs.TreeFocused {
			treeBorderColor = focusBorderColor
		}
		r.drawFileTree(vs, editorWidth, treeWidth, tabBarHeight, editorHeight+1, treeBorderColor)
	}

	// --- Draw statusline or prompt ---
	statuslineRow := contentStartY + editorHeight
	if vs.Prompt != nil {
		r.drawPrompt(vs.Prompt, statuslineRow, vs.Width)
	} else {
		r.drawStatusline(vs, statuslineRow, gutterWidth, th)
	}

	// --- Position the terminal cursor ---
	if vs.Prompt != nil {
		// Show cursor inside the prompt input.
		promptCursorX := len(vs.Prompt.Label) + vs.Prompt.CursorPos
		if promptCursorX < vs.Width {
			r.screen.ShowCursor(promptCursorX, statuslineRow)
		}
	} else {
		cursorScreenRow := vs.Buffer.CursorRow - vs.ScrollY + contentStartY
		cursorScreenCol := r.screenCol(vs.Buffer, gutterWidth)
		if !vs.TreeFocused && cursorScreenRow >= contentStartY && cursorScreenRow < contentStartY+editorHeight {
			r.screen.ShowCursor(cursorScreenCol, cursorScreenRow)
		} else {
			r.screen.HideCursor()
		}
	}

	// --- Draw popup menu on top if active ---
	if vs.Popup != nil {
		r.renderPopup(vs.Popup, vs.Width, vs.Height)
	}

	r.screen.Show()
}

// drawTabBar renders the tab bar at row 0.
func (r *Renderer) drawTabBar(vs *ViewState, editorWidth, treeWidth int) {
	tabBg := tcell.StyleDefault.
		Foreground(tcell.ColorDimGray).
		Background(tcell.ColorBlack)
	tabActive := tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.NewRGBColor(45, 45, 45)).
		Bold(true)
	tabInactive := tcell.StyleDefault.
		Foreground(tcell.ColorDarkGray).
		Background(tcell.ColorBlack)

	// Fill entire tab bar row with background.
	for x := 0; x < vs.Width; x++ {
		r.screen.SetContent(x, 0, ' ', nil, tabBg)
	}

	x := 0
	for i, tab := range vs.Tabs {
		style := tabInactive
		if i == vs.ActiveTabIdx {
			style = tabActive
		}

		// Build tab label: " name [+] x "
		label := " " + tab.Name
		if tab.Dirty {
			label += " +"
		}
		label += " x "

		// Draw tab.
		for _, ch := range label {
			if x >= vs.Width {
				break
			}
			r.screen.SetContent(x, 0, ch, nil, style)
			x++
		}

		// Draw separator.
		if x < vs.Width {
			r.screen.SetContent(x, 0, '|', nil, tabBg)
			x++
		}
	}
}

// drawFileTree renders the file tree panel on the right side.
func (r *Renderer) drawFileTree(vs *ViewState, startX, treeWidth, startY, height int, borderColor tcell.Color) {
	treeBg := tcell.NewRGBColor(24, 24, 24)
	borderStyle := tcell.StyleDefault.
		Foreground(borderColor).
		Background(treeBg)
	headerStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(180, 180, 180)).
		Background(tcell.NewRGBColor(35, 35, 35)).
		Bold(true)
	dirStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(220, 180, 90)).
		Background(treeBg)
	fileStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(200, 200, 200)).
		Background(treeBg)
	cursorStyle := tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.NewRGBColor(55, 55, 75)).
		Bold(true)

	// Draw top border of the tree panel.
	topBorderStyle := tcell.StyleDefault.Foreground(borderColor).Background(tcell.ColorDefault)
	r.screen.SetContent(startX, startY, tcell.RuneTTee, nil, topBorderStyle)
	for x := startX + 1; x < startX+treeWidth; x++ {
		r.screen.SetContent(x, startY, tcell.RuneHLine, nil, topBorderStyle)
	}

	// Draw vertical border (separator between editor and tree).
	for row := 1; row < height; row++ {
		r.screen.SetContent(startX, startY+row, tcell.RuneVLine, nil, borderStyle)
	}

	contentX := startX + 1 // After the border.
	contentW := treeWidth - 1
	if contentW < 1 {
		return
	}

	// Draw header.
	headerRow := startY + 1
	header := " EXPLORER"
	for j := 0; j < contentW; j++ {
		ch := ' '
		if j < len(header) {
			ch = rune(header[j])
		}
		r.screen.SetContent(contentX+j, headerRow, ch, nil, headerStyle)
	}

	// Draw tree nodes.
	treeContentStart := startY + 2 // After top border + header.
	treeContentHeight := height - 2
	if treeContentHeight < 1 {
		return
	}

	for row := 0; row < treeContentHeight; row++ {
		nodeIdx := vs.TreeScrollY + row
		screenRow := treeContentStart + row

		if nodeIdx >= len(vs.TreeNodes) {
			// Clear remaining rows.
			for j := 0; j < contentW; j++ {
				r.screen.SetContent(contentX+j, screenRow, ' ', nil, fileStyle)
			}
			continue
		}

		node := vs.TreeNodes[nodeIdx]

		// Choose style.
		style := fileStyle
		if node.IsDir {
			style = dirStyle
		}
		if nodeIdx == vs.TreeCursor {
			if vs.TreeFocused {
				style = cursorStyle
			} else {
				// Dim cursor when tree is not focused.
				style = tcell.StyleDefault.
					Foreground(tcell.NewRGBColor(200, 200, 200)).
					Background(tcell.NewRGBColor(40, 40, 40))
			}
		}

		// Build display line: indent + icon + name.
		indent := strings.Repeat("  ", node.Depth)
		icon := "  "
		if node.IsDir {
			if node.Expanded {
				icon = "▾ "
			} else {
				icon = "▸ "
			}
		}

		// File icon based on extension.
		if !node.IsDir {
			ext := strings.ToLower(filepath.Ext(node.Name))
			switch ext {
			case ".go":
				icon = " "
			case ".js", ".ts":
				icon = " "
			case ".json":
				icon = " "
			case ".md":
				icon = " "
			case ".toml", ".yaml", ".yml":
				icon = " "
			default:
				icon = " "
			}
		}

		line := indent + icon + node.Name

		// Convert to runes for correct Unicode rendering.
		runes := []rune(line)
		for j := 0; j < contentW; j++ {
			ch := ' '
			if j < len(runes) {
				ch = runes[j]
			}
			r.screen.SetContent(contentX+j, screenRow, ch, nil, style)
		}
	}
}

// drawStatusline renders the bottom status bar.
func (r *Renderer) drawStatusline(vs *ViewState, y int, gutterWidth int, th *theme.Theme) {
	statusStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(200, 200, 200)).
		Background(tcell.NewRGBColor(30, 30, 50))
	errorStyle := tcell.StyleDefault.
		Foreground(tcell.ColorRed).
		Background(tcell.NewRGBColor(30, 30, 50))
	warningStyle := tcell.StyleDefault.
		Foreground(tcell.ColorYellow).
		Background(tcell.NewRGBColor(30, 30, 50))

	// Clear the statusline.
	for x := 0; x < vs.Width; x++ {
		r.screen.SetContent(x, y, ' ', nil, statusStyle)
	}

	// Left: file name + dirty flag + language.
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

	// Diagnostic counts after filename.
	diagX := len(left) + 2
	if vs.DiagErrors > 0 {
		errStr := fmt.Sprintf("E:%d", vs.DiagErrors)
		r.drawString(diagX, y, errStr, errorStyle)
		diagX += len(errStr) + 1
	}
	if vs.DiagWarnings > 0 {
		warnStr := fmt.Sprintf("W:%d", vs.DiagWarnings)
		r.drawString(diagX, y, warnStr, warningStyle)
	}

	// Right: status message or cursor position.
	var right string
	if vs.StatusMsg != "" {
		right = vs.StatusMsg + " "
	} else {
		right = fmt.Sprintf("Ln %d, Col %d ", vs.Buffer.CursorRow+1, vs.Buffer.CursorCol+1)
	}
	r.drawString(vs.Width-len(right), y, right, statusStyle)
}

// drawPrompt renders the input prompt bar (replaces statusline when active).
func (r *Renderer) drawPrompt(p *editor.PromptState, y, width int) {
	labelStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(100, 180, 255)).
		Background(tcell.NewRGBColor(25, 25, 40)).
		Bold(true)
	inputStyle := tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.NewRGBColor(25, 25, 40))

	// Clear the row.
	for x := 0; x < width; x++ {
		r.screen.SetContent(x, y, ' ', nil, inputStyle)
	}

	// Draw label.
	x := 0
	for _, ch := range p.Label {
		if x >= width {
			break
		}
		r.screen.SetContent(x, y, ch, nil, labelStyle)
		x++
	}

	// Draw input text.
	for _, ch := range p.Input {
		if x >= width {
			break
		}
		r.screen.SetContent(x, y, ch, nil, inputStyle)
		x++
	}
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
	popupWidth := maxLabel + 4
	if popupWidth < 12 {
		popupWidth = 12
	}
	popupHeight := len(menu.Items) + 2

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
			r.screen.SetContent(px, row, tcell.RuneVLine, nil, menuBg)
			label := fmt.Sprintf(" %-*s ", maxLabel, item.Label)
			for j, ch := range label {
				if j < popupWidth-2 {
					r.screen.SetContent(px+1+j, row, ch, nil, style)
				}
			}
			for j := len(label); j < popupWidth-2; j++ {
				r.screen.SetContent(px+1+j, row, ' ', nil, style)
			}
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
