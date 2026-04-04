package term

import "fmt"

// Color represents an RGB color or the terminal default.
type Color struct {
	R, G, B byte
	IsSet   bool // false means terminal default
}

// NoColor is the terminal default color.
var NoColor = Color{}

// RGB creates an RGB color.
func RGB(r, g, b int) Color {
	return Color{R: byte(r), G: byte(g), B: byte(b), IsSet: true}
}

// fgANSI returns the ANSI escape for this color as foreground.
func (c Color) fgANSI() string {
	if !c.IsSet {
		return "39"
	}
	return fmt.Sprintf("38;2;%d;%d;%d", c.R, c.G, c.B)
}

// bgANSI returns the ANSI escape for this color as background.
func (c Color) bgANSI() string {
	if !c.IsSet {
		return "49"
	}
	return fmt.Sprintf("48;2;%d;%d;%d", c.R, c.G, c.B)
}

// Style defines text appearance: foreground, background, and attributes.
type Style struct {
	Fg        Color
	Bg        Color
	Bold      bool
	Dim       bool
	Italic    bool
	Underline bool
	Reverse   bool
}

// DefaultStyle uses terminal defaults for everything.
var DefaultStyle = Style{}

// NewStyle creates a style with fg and bg colors.
func NewStyle(fg, bg Color) Style {
	return Style{Fg: fg, Bg: bg}
}

// WithFg returns a copy with a new foreground color.
func (s Style) WithFg(c Color) Style { s.Fg = c; return s }

// WithBg returns a copy with a new background color.
func (s Style) WithBg(c Color) Style { s.Bg = c; return s }

// WithBold returns a copy with bold set.
func (s Style) WithBold(v bool) Style { s.Bold = v; return s }

// WithDim returns a copy with dim set.
func (s Style) WithDim(v bool) Style { s.Dim = v; return s }

// WithItalic returns a copy with italic set.
func (s Style) WithItalic(v bool) Style { s.Italic = v; return s }

// WithUnderline returns a copy with underline set.
func (s Style) WithUnderline(v bool) Style { s.Underline = v; return s }

// WithReverse returns a copy with reverse set.
func (s Style) WithReverse(v bool) Style { s.Reverse = v; return s }

// ansiSeq builds the SGR escape sequence for this style.
func (s Style) ansiSeq() string {
	// Build SGR params
	seq := "\x1b[0"
	if s.Bold {
		seq += ";1"
	}
	if s.Dim {
		seq += ";2"
	}
	if s.Italic {
		seq += ";3"
	}
	if s.Underline {
		seq += ";4"
	}
	if s.Reverse {
		seq += ";7"
	}
	seq += ";" + s.Fg.fgANSI()
	seq += ";" + s.Bg.bgANSI()
	seq += "m"
	return seq
}
