package terminal

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// StyledChar represents a single character with ANSI styling.
type StyledChar struct {
	Rune  rune
	Style tcell.Style
}

// StyledLine is a line of styled characters ready for rendering.
type StyledLine []StyledChar

// ansiColorMap maps ANSI color codes to tcell colors.
var ansiColorMap = map[int]tcell.Color{
	30: tcell.ColorBlack,
	31: tcell.ColorMaroon,
	32: tcell.ColorGreen,
	33: tcell.ColorOlive,
	34: tcell.ColorNavy,
	35: tcell.ColorPurple,
	36: tcell.ColorTeal,
	37: tcell.ColorSilver,
	// Bright colors.
	90: tcell.ColorGray,
	91: tcell.ColorRed,
	92: tcell.ColorLime,
	93: tcell.ColorYellow,
	94: tcell.ColorBlue,
	95: tcell.ColorFuchsia,
	96: tcell.ColorAqua,
	97: tcell.ColorWhite,
}

// ansiBgColorMap maps ANSI background color codes to tcell colors.
var ansiBgColorMap = map[int]tcell.Color{
	40:  tcell.ColorBlack,
	41:  tcell.ColorMaroon,
	42:  tcell.ColorGreen,
	43:  tcell.ColorOlive,
	44:  tcell.ColorNavy,
	45:  tcell.ColorPurple,
	46:  tcell.ColorTeal,
	47:  tcell.ColorSilver,
	100: tcell.ColorGray,
	101: tcell.ColorRed,
	102: tcell.ColorLime,
	103: tcell.ColorYellow,
	104: tcell.ColorBlue,
	105: tcell.ColorFuchsia,
	106: tcell.ColorAqua,
	107: tcell.ColorWhite,
}

// ansiEscapeRe matches ANSI escape sequences.
var ansiEscapeRe = regexp.MustCompile(`\x1b\[([0-9;]*)([A-Za-z])`)

// ParseANSILine parses a single line containing ANSI escape codes into styled characters.
func ParseANSILine(line string) StyledLine {
	result := make(StyledLine, 0, len(line))
	currentStyle := tcell.StyleDefault

	i := 0
	for i < len(line) {
		// Check for ESC character.
		if line[i] == '\x1b' && i+1 < len(line) && line[i+1] == '[' {
			// Find the end of the escape sequence.
			j := i + 2
			for j < len(line) && ((line[j] >= '0' && line[j] <= '9') || line[j] == ';') {
				j++
			}
			if j < len(line) {
				cmd := line[j]
				params := line[i+2 : j]
				if cmd == 'm' {
					currentStyle = applyANSIParams(currentStyle, params)
				}
				// Skip cursor movement and other CSI sequences.
				i = j + 1
				continue
			}
		}

		// Regular character (handle multi-byte UTF-8).
		r, size := decodeRune(line[i:])
		if r != '\t' {
			result = append(result, StyledChar{Rune: r, Style: currentStyle})
		} else {
			// Expand tab to spaces.
			spaces := 8 - (len(result) % 8)
			for s := 0; s < spaces; s++ {
				result = append(result, StyledChar{Rune: ' ', Style: currentStyle})
			}
		}
		i += size
	}

	return result
}

// decodeRune decodes a UTF-8 rune from a string slice.
func decodeRune(s string) (rune, int) {
	if len(s) == 0 {
		return 0, 0
	}
	b := s[0]
	if b < 0x80 {
		return rune(b), 1
	}
	// Multi-byte sequence.
	var r rune
	var size int
	switch {
	case b&0xE0 == 0xC0:
		r, size = rune(b&0x1F), 2
	case b&0xF0 == 0xE0:
		r, size = rune(b&0x0F), 3
	case b&0xF8 == 0xF0:
		r, size = rune(b&0x07), 4
	default:
		return '?', 1 // Invalid UTF-8.
	}
	if len(s) < size {
		return '?', 1
	}
	for i := 1; i < size; i++ {
		r = (r << 6) | rune(s[i]&0x3F)
	}
	return r, size
}

// applyANSIParams applies SGR (Select Graphic Rendition) parameters to a style.
func applyANSIParams(style tcell.Style, params string) tcell.Style {
	if params == "" || params == "0" {
		return tcell.StyleDefault
	}

	parts := strings.Split(params, ";")
	for i := 0; i < len(parts); i++ {
		code, err := strconv.Atoi(parts[i])
		if err != nil {
			continue
		}

		switch {
		case code == 0:
			style = tcell.StyleDefault
		case code == 1:
			style = style.Bold(true)
		case code == 2:
			style = style.Dim(true)
		case code == 3:
			style = style.Italic(true)
		case code == 4:
			style = style.Underline(true)
		case code == 7:
			style = style.Reverse(true)
		case code == 22:
			style = style.Bold(false).Dim(false)
		case code == 23:
			style = style.Italic(false)
		case code == 24:
			style = style.Underline(false)
		case code == 27:
			style = style.Reverse(false)
		case code >= 30 && code <= 37:
			if c, ok := ansiColorMap[code]; ok {
				style = style.Foreground(c)
			}
		case code == 38:
			// Extended foreground color: 38;5;N (256-color) or 38;2;R;G;B (truecolor).
			if i+1 < len(parts) {
				next, _ := strconv.Atoi(parts[i+1])
				if next == 5 && i+2 < len(parts) {
					idx, _ := strconv.Atoi(parts[i+2])
					style = style.Foreground(tcell.PaletteColor(idx))
					i += 2
				} else if next == 2 && i+4 < len(parts) {
					r, _ := strconv.Atoi(parts[i+2])
					g, _ := strconv.Atoi(parts[i+3])
					b, _ := strconv.Atoi(parts[i+4])
					style = style.Foreground(tcell.NewRGBColor(int32(r), int32(g), int32(b)))
					i += 4
				}
			}
		case code == 39:
			style = style.Foreground(tcell.ColorDefault)
		case code >= 40 && code <= 47:
			if c, ok := ansiBgColorMap[code]; ok {
				style = style.Background(c)
			}
		case code == 48:
			// Extended background color.
			if i+1 < len(parts) {
				next, _ := strconv.Atoi(parts[i+1])
				if next == 5 && i+2 < len(parts) {
					idx, _ := strconv.Atoi(parts[i+2])
					style = style.Background(tcell.PaletteColor(idx))
					i += 2
				} else if next == 2 && i+4 < len(parts) {
					r, _ := strconv.Atoi(parts[i+2])
					g, _ := strconv.Atoi(parts[i+3])
					b, _ := strconv.Atoi(parts[i+4])
					style = style.Background(tcell.NewRGBColor(int32(r), int32(g), int32(b)))
					i += 4
				}
			}
		case code == 49:
			style = style.Background(tcell.ColorDefault)
		case code >= 90 && code <= 97:
			if c, ok := ansiColorMap[code]; ok {
				style = style.Foreground(c)
			}
		case code >= 100 && code <= 107:
			if c, ok := ansiBgColorMap[code]; ok {
				style = style.Background(c)
			}
		}
	}

	return style
}

// StripANSI removes all ANSI escape sequences from a string.
func StripANSI(s string) string {
	return ansiEscapeRe.ReplaceAllString(s, "")
}
