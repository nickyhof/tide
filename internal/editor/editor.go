package editor

import (
	"fmt"
	"path/filepath"

	"github.com/nickyhof/tide/internal/buffer"
	"github.com/nickyhof/tide/internal/highlight"
	"github.com/nickyhof/tide/internal/term"
)

const gutterWidth = 5

// Theme holds the styles used by the editor for rendering.
type Theme struct {
	Foreground term.Style
	LineNumber term.Style
}

// Editor manages the editing state for a single buffer.
type Editor struct {
	Buffer      *buffer.Buffer
	CursorRow   int
	CursorCol   int
	ScrollRow   int
	ScrollCol   int
	highlighter *highlight.Highlighter
}

// NewEditor creates an editor for the given buffer.
func NewEditor(buf *buffer.Buffer) *Editor {
	e := &Editor{Buffer: buf}
	e.setupHighlighter()
	return e
}

func (e *Editor) setupHighlighter() {
	if e.Buffer.FilePath != "" {
		bg := term.RGB(30, 30, 30)
		e.highlighter = highlight.ForFile(e.Buffer.FilePath, bg)
	}
}

// Title returns a display title for the tab.
func (e *Editor) Title() string {
	if e.Buffer.FilePath != "" {
		name := filepath.Base(e.Buffer.FilePath)
		if e.Buffer.Modified {
			return "● " + name
		}
		return name
	}
	if e.Buffer.Modified {
		return "● Untitled"
	}
	return "Untitled"
}

func (e *Editor) MoveCursor(dRow, dCol int) {
	e.CursorRow += dRow
	e.CursorCol += dCol
	e.clampCursor()
}

func (e *Editor) SetCursor(row, col int) {
	e.CursorRow = row
	e.CursorCol = col
	e.clampCursor()
}

func (e *Editor) InsertChar(ch rune) {
	offset := e.cursorOffset()
	e.Buffer.Insert(offset, string(ch))
	e.CursorCol++
}

func (e *Editor) InsertNewline() {
	offset := e.cursorOffset()
	e.Buffer.Insert(offset, "\n")
	e.CursorRow++
	e.CursorCol = 0
}

func (e *Editor) Backspace() {
	if e.CursorCol == 0 && e.CursorRow == 0 {
		return
	}
	if e.CursorCol > 0 {
		offset := e.cursorOffset()
		e.Buffer.Delete(offset-1, 1)
		e.CursorCol--
	} else {
		prevLen := e.Buffer.LineLength(e.CursorRow - 1)
		offset := e.cursorOffset()
		e.Buffer.Delete(offset-1, 1)
		e.CursorRow--
		e.CursorCol = prevLen
	}
}

func (e *Editor) DeleteChar() {
	offset := e.cursorOffset()
	if offset < e.Buffer.Length() {
		e.Buffer.Delete(offset, 1)
	}
}

func (e *Editor) InsertTab() {
	offset := e.cursorOffset()
	e.Buffer.Insert(offset, "    ")
	e.CursorCol += 4
}

func (e *Editor) ScrollIntoView(viewHeight, viewWidth int) {
	if e.CursorRow < e.ScrollRow {
		e.ScrollRow = e.CursorRow
	}
	if e.CursorRow >= e.ScrollRow+viewHeight {
		e.ScrollRow = e.CursorRow - viewHeight + 1
	}
	if e.CursorCol < e.ScrollCol {
		e.ScrollCol = e.CursorCol
	}
	if e.CursorCol >= e.ScrollCol+viewWidth {
		e.ScrollCol = e.CursorCol - viewWidth + 1
	}
}

// Draw renders the editor content to the screen.
func (e *Editor) Draw(scr *term.Screen, x, y, w, h int, theme Theme) {
	editX := x + gutterWidth
	editW := w - gutterWidth
	if editW < 1 {
		return
	}

	e.ScrollIntoView(h, editW)

	// Track block comment state
	inBlock := false
	if e.highlighter != nil {
		for lineIdx := 0; lineIdx < e.ScrollRow && lineIdx < e.Buffer.LineCount; lineIdx++ {
			line := e.Buffer.Line(lineIdx)
			_, inBlock = e.highlighter.Highlight(line, inBlock)
		}
	}

	for row := 0; row < h; row++ {
		lineIdx := e.ScrollRow + row
		screenY := y + row

		if lineIdx < e.Buffer.LineCount {
			numStr := fmt.Sprintf("%4d ", lineIdx+1)
			scr.DrawText(x, screenY, theme.LineNumber, numStr, gutterWidth)
		} else {
			scr.DrawText(x, screenY, theme.LineNumber, "   ~ ", gutterWidth)
		}

		if lineIdx < e.Buffer.LineCount {
			line := e.Buffer.Line(lineIdx)
			if len(line) > 0 && line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
			}

			if e.highlighter != nil {
				tokens, newInBlock := e.highlighter.Highlight(line, inBlock)
				inBlock = newInBlock
				e.drawHighlightedLine(scr, editX, screenY, editW, line, tokens, theme)
			} else {
				if e.ScrollCol < len(line) {
					visible := line[e.ScrollCol:]
					scr.DrawText(editX, screenY, theme.Foreground, visible, editW)
				}
			}
		}
	}
}

func (e *Editor) drawHighlightedLine(scr *term.Screen, x, y, maxW int, line string, tokens []highlight.Token, theme Theme) {
	if len(line) == 0 {
		return
	}

	styles := make([]term.Style, len(line))
	for i := range styles {
		styles[i] = theme.Foreground
	}
	for _, tok := range tokens {
		style := e.highlighter.Palette.Style(tok.Kind)
		for i := tok.Start; i < tok.End && i < len(line); i++ {
			styles[i] = style
		}
	}

	col := 0
	for i, ch := range line {
		if i < e.ScrollCol {
			continue
		}
		if col >= maxW {
			break
		}
		scr.SetContent(x+col, y, ch, styles[i])
		col++
	}
}

// CursorScreenPos returns the screen position of the cursor.
func (e *Editor) CursorScreenPos(x, y int) (int, int) {
	return x + gutterWidth + e.CursorCol - e.ScrollCol, y + e.CursorRow - e.ScrollRow
}

// HandleKey processes a key event and returns true if handled.
func (e *Editor) HandleKey(ev term.KeyEvent) bool {
	switch ev.Key {
	case term.KeyUp:
		e.MoveCursor(-1, 0)
	case term.KeyDown:
		e.MoveCursor(1, 0)
	case term.KeyLeft:
		e.MoveCursor(0, -1)
	case term.KeyRight:
		e.MoveCursor(0, 1)
	case term.KeyHome:
		e.CursorCol = 0
	case term.KeyEnd:
		e.CursorCol = e.Buffer.LineLength(e.CursorRow)
	case term.KeyPgUp:
		e.MoveCursor(-20, 0)
	case term.KeyPgDn:
		e.MoveCursor(20, 0)
	case term.KeyBackspace:
		e.Backspace()
	case term.KeyDelete:
		e.DeleteChar()
	case term.KeyEnter:
		e.InsertNewline()
	case term.KeyTab:
		e.InsertTab()
	case term.KeyRune:
		e.InsertChar(ev.Rune)
	default:
		return false
	}
	return true
}

func (e *Editor) cursorOffset() int {
	return e.Buffer.LineStart(e.CursorRow) + e.CursorCol
}

func (e *Editor) clampCursor() {
	maxLine := e.Buffer.LineCount - 1
	if e.CursorRow < 0 {
		e.CursorRow = 0
	}
	if e.CursorRow > maxLine {
		e.CursorRow = maxLine
	}
	maxCol := e.Buffer.LineLength(e.CursorRow)
	if e.CursorCol < 0 {
		e.CursorCol = 0
	}
	if e.CursorCol > maxCol {
		e.CursorCol = maxCol
	}
}
