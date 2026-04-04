package ui

import (
	"github.com/nickyhof/tide/internal/term"
)

// PaletteItem represents a selectable entry in the command palette.
type PaletteItem struct {
	Name     string
	Desc     string
	Shortcut string
}

// PaletteTheme holds styles for the palette.
type PaletteTheme struct {
	Border      term.Style
	Input       term.Style
	Item        term.Style
	Selected    term.Style
	Shortcut    term.Style
	ShortcutSel term.Style
}

// Palette is a VSCode-style command palette overlay.
type Palette struct {
	Open     bool
	Query    string
	Items    []PaletteItem
	AllItems []PaletteItem
	Cursor   int
	Scroll   int
	OnSelect func(name string)
	OnClose  func()
}

// NewPalette creates a command palette.
func NewPalette() *Palette {
	return &Palette{}
}

// Show opens the palette with the given items.
func (p *Palette) Show(items []PaletteItem) {
	p.Open = true
	p.Query = ""
	p.AllItems = items
	p.Items = items
	p.Cursor = 0
	p.Scroll = 0
}

// Close dismisses the palette.
func (p *Palette) Close() {
	p.Open = false
	p.Query = ""
	if p.OnClose != nil {
		p.OnClose()
	}
}

// HandleKey processes input.
func (p *Palette) HandleKey(ev term.KeyEvent) bool {
	if !p.Open {
		return false
	}
	switch ev.Key {
	case term.KeyEscape:
		p.Close()
	case term.KeyEnter:
		if p.Cursor < len(p.Items) {
			name := p.Items[p.Cursor].Name
			p.Open = false
			p.Query = ""
			if p.OnSelect != nil {
				p.OnSelect(name)
			}
		}
	case term.KeyUp:
		if p.Cursor > 0 {
			p.Cursor--
		}
	case term.KeyDown:
		if p.Cursor < len(p.Items)-1 {
			p.Cursor++
		}
	case term.KeyBackspace:
		if len(p.Query) > 0 {
			p.Query = p.Query[:len(p.Query)-1]
			p.filter()
		}
	case term.KeyRune:
		p.Query += string(ev.Rune)
		p.filter()
	default:
		return true
	}
	return true
}

func (p *Palette) filter() {
	if p.Query == "" {
		p.Items = p.AllItems
	} else {
		q := toLower(p.Query)
		p.Items = nil
		for _, item := range p.AllItems {
			if fuzzyMatch(toLower(item.Name), q) || fuzzyMatch(toLower(item.Desc), q) {
				p.Items = append(p.Items, item)
			}
		}
	}
	p.Cursor = 0
	p.Scroll = 0
}

// Draw renders the palette.
func (p *Palette) Draw(scr *term.Screen, theme PaletteTheme) {
	if !p.Open {
		return
	}

	paletteW := 50
	if paletteW > scr.Width-4 {
		paletteW = scr.Width - 4
	}
	maxVisible := 10
	if maxVisible > len(p.Items) {
		maxVisible = len(p.Items)
	}
	paletteH := maxVisible + 2

	x := (scr.Width - paletteW) / 2
	y := 2

	scr.FillRect(x-1, y-1, paletteW+2, paletteH+2, theme.Border, ' ')
	scr.FillRect(x, y, paletteW, paletteH, theme.Border, ' ')

	// Input
	scr.FillRect(x, y, paletteW, 1, theme.Input, ' ')
	prompt := "> " + p.Query
	scr.DrawText(x+1, y, theme.Input, prompt, paletteW-2)

	// Scroll
	if p.Cursor < p.Scroll {
		p.Scroll = p.Cursor
	}
	if p.Cursor >= p.Scroll+maxVisible {
		p.Scroll = p.Cursor - maxVisible + 1
	}

	// Items
	for i := 0; i < maxVisible; i++ {
		idx := p.Scroll + i
		if idx >= len(p.Items) {
			break
		}
		item := p.Items[idx]
		row := y + 1 + i

		style := theme.Item
		scStyle := theme.Shortcut
		if idx == p.Cursor {
			style = theme.Selected
			scStyle = theme.ShortcutSel
		}

		scr.FillRect(x, row, paletteW, 1, style, ' ')
		scr.DrawText(x+2, row, style, item.Desc, paletteW-12)

		if item.Shortcut != "" {
			scX := x + paletteW - len(item.Shortcut) - 2
			if scX > x+2 {
				scr.DrawText(scX, row, scStyle, item.Shortcut, len(item.Shortcut))
			}
		}
	}
}

func fuzzyMatch(str, pattern string) bool {
	pi := 0
	for i := 0; i < len(str) && pi < len(pattern); i++ {
		if str[i] == pattern[pi] {
			pi++
		}
	}
	return pi == len(pattern)
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
