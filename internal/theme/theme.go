package theme

import (
	"github.com/gdamore/tcell/v2"
	"github.com/israelcorrea/crit-ide/internal/highlight"
)

// Theme maps semantic tokens and UI elements to tcell styles.
type Theme struct {
	Name   string
	Syntax map[highlight.TokenType]tcell.Style
	// UI element styles.
	Default      tcell.Style
	Gutter       tcell.Style
	GutterActive tcell.Style
	StatusLine   tcell.Style
	DiagError    tcell.Style
	DiagWarn     tcell.Style
	DiagInfo     tcell.Style
	DiagHint     tcell.Style
}

// StyleFor returns the style for a token type, falling back to Default.
func (t *Theme) StyleFor(tt highlight.TokenType) tcell.Style {
	if s, ok := t.Syntax[tt]; ok {
		return s
	}
	return t.Default
}

// DefaultTheme returns a built-in dark theme with sensible colors.
func DefaultTheme() *Theme {
	def := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault)

	return &Theme{
		Name: "default-dark",
		Syntax: map[highlight.TokenType]tcell.Style{
			highlight.TokenKeyword:   def.Foreground(tcell.ColorMediumPurple),
			highlight.TokenString:    def.Foreground(tcell.ColorSandyBrown),
			highlight.TokenComment:   def.Foreground(tcell.ColorGray).Italic(true),
			highlight.TokenFunction:  def.Foreground(tcell.ColorDodgerBlue),
			highlight.TokenTypeName:  def.Foreground(tcell.ColorTeal),
			highlight.TokenNumber:    def.Foreground(tcell.ColorLightGreen),
			highlight.TokenOperator:  def.Foreground(tcell.ColorLightCyan),
			highlight.TokenBuiltin:   def.Foreground(tcell.ColorGold),
			highlight.TokenConstant:  def.Foreground(tcell.ColorOrangeRed),
			highlight.TokenTag:       def.Foreground(tcell.ColorDodgerBlue),
			highlight.TokenAttribute: def.Foreground(tcell.ColorLightGreen),
			highlight.TokenHeading:   def.Foreground(tcell.ColorMediumPurple).Bold(true),
			highlight.TokenBold:      def.Bold(true),
			highlight.TokenItalic:    def.Italic(true),
			highlight.TokenLink:      def.Foreground(tcell.ColorDodgerBlue).Underline(true),
			highlight.TokenProperty:  def.Foreground(tcell.ColorLightCoral),
			highlight.TokenVariable:  def.Foreground(tcell.ColorLightSkyBlue),
			highlight.TokenPreproc:   def.Foreground(tcell.ColorMediumPurple),
		},
		Default:      def,
		Gutter:       def.Foreground(tcell.ColorGray),
		GutterActive: def.Foreground(tcell.ColorWhite),
		StatusLine:   tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite),
		DiagError:    def.Foreground(tcell.ColorRed).Underline(true),
		DiagWarn:     def.Foreground(tcell.ColorYellow).Underline(true),
		DiagInfo:     def.Foreground(tcell.ColorBlue).Underline(true),
		DiagHint:     def.Foreground(tcell.ColorGray).Underline(true),
	}
}
