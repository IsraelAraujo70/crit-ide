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
	Prompt        *editor.PromptState  // Non-nil when prompt is active.

	// Search state.
	Search        *editor.SearchState  // Non-nil when find/replace is active.

	// File finder.
	Finder        *editor.FinderState  // Non-nil when fuzzy file finder is active.

	// Completion popup.
	Completion    *editor.CompletionState // Non-nil when autocomplete popup is active.

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

	// Build search match map for visible lines.
	searchMatchStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.NewRGBColor(200, 180, 80))
	searchCurrentStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.NewRGBColor(255, 140, 50)).Bold(true)
	searchMatchMap := r.buildSearchMatchMap(vs.Search, vs.ScrollY, vs.ScrollY+editorHeight)

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

			// Search match highlighting (overrides syntax/diagnostics).
			if matches, ok := searchMatchMap[lineIdx]; ok {
				for _, sm := range matches {
					if byteOff >= sm.startCol && byteOff < sm.endCol {
						if sm.isCurrent {
							style = searchCurrentStyle
						} else {
							style = searchMatchStyle
						}
						break
					}
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

	// --- Draw statusline, prompt, or search bar ---
	statuslineRow := contentStartY + editorHeight
	if vs.Search != nil {
		r.drawSearchBar(vs.Search, statuslineRow, vs.Width)
	} else if vs.Prompt != nil {
		r.drawPrompt(vs.Prompt, statuslineRow, vs.Width)
	} else {
		r.drawStatusline(vs, statuslineRow, gutterWidth, th)
	}

	// --- Position the terminal cursor ---
	if vs.Finder != nil {
		// Cursor is drawn inside the finder popup.
		r.screen.HideCursor()
	} else if vs.Search != nil {
		// Show cursor inside the search bar's active field.
		var cursorX int
		if vs.Search.ActiveField == editor.FieldFind {
			cursorX = len("Find: ") + vs.Search.QueryCursor
		} else {
			cursorX = len("Replace: ") + vs.Search.ReplaceCursor
		}
		if cursorX < vs.Width {
			r.screen.ShowCursor(cursorX, statuslineRow)
		}
	} else if vs.Prompt != nil {
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

	// --- Draw file finder popup if active ---
	if vs.Finder != nil {
		r.renderFinder(vs.Finder, vs.Width, vs.Height)
	}

	// --- Draw completion popup if active ---
	if vs.Completion != nil && !vs.Completion.IsEmpty() {
		gutterW := r.gutterWidth(vs.Buffer.Text.LineCount())
		r.renderCompletion(vs, gutterW, contentStartY, editorHeight)
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

// searchMatchEntry represents a single search match on a line for rendering.
type searchMatchEntry struct {
	startCol  int
	endCol    int
	isCurrent bool
}

// buildSearchMatchMap builds a per-line map of search matches for the visible range.
func (r *Renderer) buildSearchMatchMap(search *editor.SearchState, minLine, maxLine int) map[int][]searchMatchEntry {
	if search == nil || len(search.Matches) == 0 {
		return nil
	}
	m := make(map[int][]searchMatchEntry)
	for i, match := range search.Matches {
		if match.Start.Line >= minLine && match.Start.Line < maxLine {
			m[match.Start.Line] = append(m[match.Start.Line], searchMatchEntry{
				startCol:  match.Start.Col,
				endCol:    match.End.Col,
				isCurrent: i == search.CurrentIdx,
			})
		}
	}
	return m
}

// renderFinder draws the centered fuzzy file finder popup.
func (r *Renderer) renderFinder(fs *editor.FinderState, screenW, screenH int) {
	// Popup dimensions.
	popupWidth := screenW * 2 / 3
	if popupWidth < 40 {
		popupWidth = 40
	}
	if popupWidth > screenW-4 {
		popupWidth = screenW - 4
	}

	maxVisible := 15
	popupHeight := maxVisible + 4 // borders (2) + input row (1) + header/separator (1)
	if popupHeight > screenH-2 {
		popupHeight = screenH - 2
		maxVisible = popupHeight - 4
		if maxVisible < 1 {
			maxVisible = 1
		}
	}

	// Center the popup.
	px := (screenW - popupWidth) / 2
	py := (screenH - popupHeight) / 4 // Slight bias toward top.
	if py < 1 {
		py = 1
	}

	// Styles.
	borderStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(80, 140, 255)).
		Background(tcell.NewRGBColor(20, 20, 30))
	headerStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(120, 160, 220)).
		Background(tcell.NewRGBColor(20, 20, 30))
	inputBg := tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.NewRGBColor(30, 30, 45))
	inputLabelStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(80, 180, 255)).
		Background(tcell.NewRGBColor(30, 30, 45)).
		Bold(true)
	resultStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(200, 200, 200)).
		Background(tcell.NewRGBColor(20, 20, 30))
	selectedStyle := tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.NewRGBColor(40, 60, 100)).
		Bold(true)
	matchCharStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(255, 200, 80)).
		Background(tcell.NewRGBColor(20, 20, 30)).
		Bold(true)
	matchCharSelectedStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(255, 220, 100)).
		Background(tcell.NewRGBColor(40, 60, 100)).
		Bold(true)
	emptyStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(100, 100, 100)).
		Background(tcell.NewRGBColor(20, 20, 30))

	// Draw top border.
	r.screen.SetContent(px, py, tcell.RuneULCorner, nil, borderStyle)
	title := " Open File "
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

	// Draw search icon and input.
	label := " > "
	cx := px + 1
	for _, ch := range label {
		r.screen.SetContent(cx, inputRow, ch, nil, inputLabelStyle)
		cx++
	}

	// Draw query text.
	cursorScreenX := cx
	qRunes := []rune(fs.Query)
	qBytePos := 0
	for _, ch := range qRunes {
		if qBytePos == fs.CursorPos {
			cursorScreenX = cx
		}
		if cx >= px+popupWidth-1 {
			break
		}
		r.screen.SetContent(cx, inputRow, ch, nil, inputBg)
		cx++
		qBytePos += len(string(ch))
	}
	if qBytePos == fs.CursorPos {
		cursorScreenX = cx
	}

	// Show cursor inside finder input.
	r.screen.ShowCursor(cursorScreenX, inputRow)

	// Draw result count on the right of the input row.
	countStr := fmt.Sprintf("%d/%d ", fs.ResultCount(), fs.TotalFiles)
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
		resultIdx := fs.ScrollY + row

		r.screen.SetContent(px, screenRow, tcell.RuneVLine, nil, borderStyle)

		if resultIdx >= len(fs.Results) {
			// Empty row.
			for x := 0; x < contentWidth; x++ {
				r.screen.SetContent(px+1+x, screenRow, ' ', nil, resultStyle)
			}
		} else {
			result := fs.Results[resultIdx]
			isSelected := resultIdx == fs.SelectedIdx

			baseStyle := resultStyle
			highlightStyle := matchCharStyle
			if isSelected {
				baseStyle = selectedStyle
				highlightStyle = matchCharSelectedStyle
			}

			// Build a set of matched character indices for quick lookup.
			matchSet := make(map[int]bool, len(result.Matches))
			for _, m := range result.Matches {
				matchSet[m] = true
			}

			// Draw the path with match highlighting.
			runes := []rune(result.RelPath)
			x := 0

			// Leading space.
			r.screen.SetContent(px+1, screenRow, ' ', nil, baseStyle)
			x++

			for ci, ch := range runes {
				if x >= contentWidth {
					break
				}
				style := baseStyle
				if matchSet[ci] {
					style = highlightStyle
				}
				r.screen.SetContent(px+1+x, screenRow, ch, nil, style)
				x++
			}

			// Fill remaining space.
			for x < contentWidth {
				r.screen.SetContent(px+1+x, screenRow, ' ', nil, baseStyle)
				x++
			}
		}

		r.screen.SetContent(px+popupWidth-1, screenRow, tcell.RuneVLine, nil, borderStyle)
	}

	// Draw empty state message if no results.
	if len(fs.Results) == 0 && fs.Query != "" {
		msgRow := py + 3
		msg := "No matching files"
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
	hint := " Enter:Open  Esc:Close  Up/Down:Navigate "
	hintStart := (popupWidth - len(hint)) / 2
	if hintStart > 1 {
		for i, ch := range hint {
			r.screen.SetContent(px+hintStart+i, bottomRow, ch, nil, tcell.StyleDefault.
				Foreground(tcell.NewRGBColor(100, 120, 160)).
				Background(tcell.NewRGBColor(20, 20, 30)))
		}
	}
}

// drawSearchBar renders the Find/Replace bar (replaces statusline when active).
func (r *Renderer) drawSearchBar(s *editor.SearchState, y, width int) {
	labelStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(100, 200, 100)).
		Background(tcell.NewRGBColor(25, 35, 25)).
		Bold(true)
	inputStyle := tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.NewRGBColor(25, 35, 25))
	inactiveInputStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(150, 150, 150)).
		Background(tcell.NewRGBColor(30, 30, 30))
	countStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(180, 180, 100)).
		Background(tcell.NewRGBColor(25, 35, 25))

	// Clear the row.
	for x := 0; x < width; x++ {
		r.screen.SetContent(x, y, ' ', nil, inputStyle)
	}

	x := 0

	if s.ActiveField == editor.FieldFind || !s.ShowReplace {
		// Draw find field.
		findLabel := "Find: "
		for _, ch := range findLabel {
			if x >= width {
				break
			}
			r.screen.SetContent(x, y, ch, nil, labelStyle)
			x++
		}
		for _, ch := range s.Query {
			if x >= width {
				break
			}
			r.screen.SetContent(x, y, ch, nil, inputStyle)
			x++
		}

		// Draw match count on the right.
		var countStr string
		if s.Query == "" {
			countStr = ""
		} else if len(s.Matches) == 0 {
			countStr = " No matches"
		} else {
			countStr = fmt.Sprintf(" %d/%d", s.CurrentMatchNumber(), s.MatchCount())
		}
		if s.ShowReplace {
			countStr += " [Tab: Replace]"
		} else {
			countStr += " [Tab: +Replace]"
		}
		right := countStr + " "
		rx := width - len(right)
		if rx > x+1 {
			r.drawString(rx, y, right, countStyle)
		}
	} else {
		// Draw replace field.
		replaceLabel := "Replace: "
		for _, ch := range replaceLabel {
			if x >= width {
				break
			}
			r.screen.SetContent(x, y, ch, nil, labelStyle)
			x++
		}
		for _, ch := range s.ReplaceText {
			if x >= width {
				break
			}
			r.screen.SetContent(x, y, ch, nil, inputStyle)
			x++
		}

		// Draw hints on the right.
		hints := " [Tab: Find] [Enter: Next] [Ctrl+R: Replace] [Ctrl+A: All] "
		_ = inactiveInputStyle // suppress unused
		rx := width - len(hints)
		if rx > x+1 {
			r.drawString(rx, y, hints, countStyle)
		}
	}
}

// renderCompletion draws the autocomplete popup below (or above) the cursor.
func (r *Renderer) renderCompletion(vs *ViewState, gutterWidth, contentStartY, editorHeight int) {
	cs := vs.Completion
	visible := cs.VisibleItems()
	if len(visible) == 0 {
		return
	}

	// Styles.
	borderStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(80, 100, 140)).
		Background(tcell.NewRGBColor(25, 25, 35))
	itemStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(200, 200, 200)).
		Background(tcell.NewRGBColor(30, 30, 45))
	selectedStyle := tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.NewRGBColor(50, 60, 90)).
		Bold(true)
	kindStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(130, 170, 230)).
		Background(tcell.NewRGBColor(30, 30, 45))
	kindSelectedStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(160, 200, 255)).
		Background(tcell.NewRGBColor(50, 60, 90)).
		Bold(true)
	detailStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(130, 130, 150)).
		Background(tcell.NewRGBColor(30, 30, 45))
	detailSelectedStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(160, 160, 180)).
		Background(tcell.NewRGBColor(50, 60, 90))

	// Calculate popup dimensions.
	maxLabelLen := 0
	maxDetailLen := 0
	for _, item := range visible {
		if len(item.Label) > maxLabelLen {
			maxLabelLen = len(item.Label)
		}
		if len(item.Detail) > maxDetailLen {
			maxDetailLen = len(item.Detail)
		}
	}

	// Popup width: " kk label   detail "
	// kindWidth=2, padding=5 (spaces around), label, optional detail.
	kindWidth := 2
	popupWidth := kindWidth + 1 + maxLabelLen + 2 // " kk label "
	if maxDetailLen > 0 {
		detailWidth := maxDetailLen
		if detailWidth > 30 {
			detailWidth = 30 // Limit detail width.
		}
		popupWidth += detailWidth + 1
	}
	if popupWidth < 20 {
		popupWidth = 20
	}
	if popupWidth > vs.Width-gutterWidth-2 {
		popupWidth = vs.Width - gutterWidth - 2
	}

	popupHeight := len(visible)

	// Position: below the cursor line, aligned to anchor column.
	cursorScreenRow := cs.AnchorRow - vs.ScrollY + contentStartY
	cursorScreenCol := gutterWidth + r.visualCol(vs.Buffer, cs.AnchorCol)

	// Try below cursor.
	py := cursorScreenRow + 1
	if py+popupHeight > contentStartY+editorHeight {
		// Try above cursor.
		py = cursorScreenRow - popupHeight
		if py < contentStartY {
			py = contentStartY
		}
	}

	px := cursorScreenCol
	if px+popupWidth > vs.Width {
		px = vs.Width - popupWidth
	}
	if px < 0 {
		px = 0
	}

	selIdx := cs.VisibleSelectedIdx()

	// Draw each item.
	for row, item := range visible {
		screenRow := py + row
		isSelected := row == selIdx

		baseStyle := itemStyle
		kStyle := kindStyle
		dStyle := detailStyle
		if isSelected {
			baseStyle = selectedStyle
			kStyle = kindSelectedStyle
			dStyle = detailSelectedStyle
		}

		x := px

		// Draw kind icon.
		icon := item.KindIcon()
		for _, ch := range icon {
			if x < vs.Width {
				r.screen.SetContent(x, screenRow, ch, nil, kStyle)
				x++
			}
		}

		// Space after kind.
		if x < vs.Width {
			r.screen.SetContent(x, screenRow, ' ', nil, baseStyle)
			x++
		}

		// Draw label.
		labelRunes := []rune(item.Label)
		for _, ch := range labelRunes {
			if x >= px+popupWidth {
				break
			}
			r.screen.SetContent(x, screenRow, ch, nil, baseStyle)
			x++
		}

		// Pad after label.
		labelEnd := px + kindWidth + 1 + maxLabelLen + 1
		for x < labelEnd && x < px+popupWidth {
			r.screen.SetContent(x, screenRow, ' ', nil, baseStyle)
			x++
		}

		// Draw detail if space allows.
		if item.Detail != "" && x < px+popupWidth-1 {
			detailRunes := []rune(item.Detail)
			for _, ch := range detailRunes {
				if x >= px+popupWidth {
					break
				}
				r.screen.SetContent(x, screenRow, ch, nil, dStyle)
				x++
			}
		}

		// Fill remaining width.
		for x < px+popupWidth {
			r.screen.SetContent(x, screenRow, ' ', nil, baseStyle)
			x++
		}
	}

	// Draw scroll indicator if there are more items.
	if len(cs.Filtered) > editor.CompletionMaxVisible {
		total := len(cs.Filtered)
		scrollFraction := fmt.Sprintf("%d/%d", cs.SelectedIdx+1, total)
		indicatorX := px + popupWidth - len(scrollFraction) - 1
		if indicatorX > px {
			for _, ch := range scrollFraction {
				r.screen.SetContent(indicatorX, py, ch, nil, borderStyle)
				indicatorX++
			}
		}
	}
}

// visualCol computes the visual column (accounting for tabs) for a given byte offset.
func (r *Renderer) visualCol(buf *editor.Buffer, byteCol int) int {
	line := buf.Text.Line(buf.CursorRow)
	col := 0
	for i, ch := range line {
		if i >= byteCol {
			break
		}
		if ch == '\t' {
			col += 4 - (col % 4)
		} else {
			col++
		}
	}
	return col
}
