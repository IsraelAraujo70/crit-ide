package terminal

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestParseANSILine_Plain(t *testing.T) {
	line := "hello world"
	styled := ParseANSILine(line)
	if len(styled) != 11 {
		t.Fatalf("expected 11 chars, got %d", len(styled))
	}
	for i, ch := range "hello world" {
		if styled[i].Rune != ch {
			t.Errorf("char %d: expected %c, got %c", i, ch, styled[i].Rune)
		}
	}
}

func TestParseANSILine_Red(t *testing.T) {
	// ESC[31m = red foreground, ESC[0m = reset
	line := "\x1b[31mERROR\x1b[0m: something"
	styled := ParseANSILine(line)

	// "ERROR: something" = 17 chars
	text := ""
	for _, sc := range styled {
		text += string(sc.Rune)
	}
	if text != "ERROR: something" {
		t.Fatalf("expected 'ERROR: something', got %q", text)
	}

	// First 5 chars should be red.
	for i := 0; i < 5; i++ {
		fg, _, _ := styled[i].Style.Decompose()
		if fg != tcell.ColorMaroon {
			t.Errorf("char %d: expected red fg, got %v", i, fg)
		}
	}

	// Chars after reset should have default style.
	for i := 5; i < len(styled); i++ {
		fg, _, _ := styled[i].Style.Decompose()
		if fg != tcell.ColorDefault {
			t.Errorf("char %d: expected default fg after reset, got %v", i, fg)
		}
	}
}

func TestParseANSILine_Bold(t *testing.T) {
	line := "\x1b[1mBOLD\x1b[0m"
	styled := ParseANSILine(line)

	text := ""
	for _, sc := range styled {
		text += string(sc.Rune)
	}
	if text != "BOLD" {
		t.Fatalf("expected 'BOLD', got %q", text)
	}

	// Check bold attribute.
	_, _, attrs := styled[0].Style.Decompose()
	if attrs&tcell.AttrBold == 0 {
		t.Error("expected bold attribute on first char")
	}
}

func TestParseANSILine_MultipleColors(t *testing.T) {
	line := "\x1b[32mgreen\x1b[34mblue\x1b[0mnormal"
	styled := ParseANSILine(line)

	text := ""
	for _, sc := range styled {
		text += string(sc.Rune)
	}
	if text != "greenbluenormal" {
		t.Fatalf("expected 'greenbluenormal', got %q", text)
	}

	// Green chars: 0-4
	for i := 0; i < 5; i++ {
		fg, _, _ := styled[i].Style.Decompose()
		if fg != tcell.ColorGreen {
			t.Errorf("char %d: expected green, got %v", i, fg)
		}
	}

	// Blue chars: 5-8
	for i := 5; i < 9; i++ {
		fg, _, _ := styled[i].Style.Decompose()
		if fg != tcell.ColorNavy {
			t.Errorf("char %d: expected blue (navy), got %v", i, fg)
		}
	}
}

func TestParseANSILine_BrightColors(t *testing.T) {
	line := "\x1b[91mred\x1b[0m"
	styled := ParseANSILine(line)

	if len(styled) != 3 {
		t.Fatalf("expected 3 chars, got %d", len(styled))
	}

	fg, _, _ := styled[0].Style.Decompose()
	if fg != tcell.ColorRed {
		t.Errorf("expected bright red, got %v", fg)
	}
}

func TestParseANSILine_256Color(t *testing.T) {
	// 38;5;196 = 256-color red
	line := "\x1b[38;5;196mhello\x1b[0m"
	styled := ParseANSILine(line)

	text := ""
	for _, sc := range styled {
		text += string(sc.Rune)
	}
	if text != "hello" {
		t.Fatalf("expected 'hello', got %q", text)
	}

	fg, _, _ := styled[0].Style.Decompose()
	expected := tcell.PaletteColor(196)
	if fg != expected {
		t.Errorf("expected palette color 196, got %v", fg)
	}
}

func TestParseANSILine_Tab(t *testing.T) {
	line := "a\tb"
	styled := ParseANSILine(line)

	// 'a' + 7 spaces (tab to 8) + 'b' = 9
	if len(styled) != 9 {
		t.Fatalf("expected 9 chars (tab expansion), got %d", len(styled))
	}
	if styled[0].Rune != 'a' {
		t.Error("first char should be 'a'")
	}
	if styled[8].Rune != 'b' {
		t.Error("last char should be 'b'")
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"\x1b[31mred\x1b[0m text", "red text"},
		{"\x1b[1;32;44mcomplex\x1b[0m", "complex"},
		{"", ""},
	}

	for _, tc := range tests {
		got := StripANSI(tc.input)
		if got != tc.expected {
			t.Errorf("StripANSI(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestParseANSILine_Empty(t *testing.T) {
	styled := ParseANSILine("")
	if len(styled) != 0 {
		t.Fatalf("expected 0 chars for empty input, got %d", len(styled))
	}
}

func TestParseANSILine_OnlyEscape(t *testing.T) {
	line := "\x1b[0m"
	styled := ParseANSILine(line)
	if len(styled) != 0 {
		t.Fatalf("expected 0 chars for reset-only, got %d", len(styled))
	}
}
